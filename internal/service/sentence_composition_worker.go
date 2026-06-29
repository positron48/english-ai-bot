package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// SentenceCompositionWorker generates one daily sentence-composition set per Pro user/course.
// It honours the regeneration guard: a previous set that was never started (status "ready")
// blocks generating a new one, so we never spend LLM budget on sets the user ignored.
type SentenceCompositionWorker struct {
	aiService   *ai.Service
	repo        *repository.SentenceCompositionRepository
	userRepo    *repository.UserRepository
	courseRepo  *repository.CourseRepository
	cbService   circuitBreakerForWorker
	cfg         config.SentenceCompositionConfig
	learning    config.LearningConfig
	defaultCode string
	logger      *zap.Logger
	stopChan    chan struct{}
}

func NewSentenceCompositionWorker(
	aiService *ai.Service,
	repo *repository.SentenceCompositionRepository,
	userRepo *repository.UserRepository,
	courseRepo *repository.CourseRepository,
	cbService circuitBreakerForWorker,
	cfg config.SentenceCompositionConfig,
	learning config.LearningConfig,
	defaultCourseCode string,
	logger *zap.Logger,
) *SentenceCompositionWorker {
	return &SentenceCompositionWorker{
		aiService:   aiService,
		repo:        repo,
		userRepo:    userRepo,
		courseRepo:  courseRepo,
		cbService:   cbService,
		cfg:         cfg,
		learning:    learning,
		defaultCode: defaultCourseCode,
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

// Start runs the generation loop until ctx is cancelled.
func (w *SentenceCompositionWorker) Start(ctx context.Context) {
	interval := ai.ParseHTTPTimeout(w.cfg.Interval)
	if interval <= 0 {
		interval = time.Hour
	}
	w.logger.Info("starting sentence composition worker", zap.Duration("interval", interval))

	go w.generateForAllUsers(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.generateForAllUsers(ctx)
		}
	}
}

// Stop stops the worker.
func (w *SentenceCompositionWorker) Stop() { close(w.stopChan) }

func (w *SentenceCompositionWorker) generateForAllUsers(ctx context.Context) {
	if w.cbService != nil {
		if isOpen, _ := w.cbService.IsOpen(); isOpen {
			w.logger.Debug("sentence worker: circuit breaker open, skipping tick")
			return
		}
	}
	users, err := w.userRepo.GetAllUsers()
	if err != nil {
		w.logger.Error("sentence worker: list users failed", zap.Error(err))
		return
	}
	for _, user := range users {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := w.generateForUser(ctx, user); err != nil {
			w.logger.Warn("sentence worker: generation failed",
				zap.Int64("user_id", user.ID), zap.Error(err))
		}
	}
}

func (w *SentenceCompositionWorker) generateForUser(ctx context.Context, user *models.User) error {
	// Pro gating.
	if !models.ParseUserTier(string(user.SubscriptionTier)).AtLeast(models.TierPro) {
		return nil
	}

	// Resolve the user's current course; only generate when the course has both prompts.
	courseCode, err := w.courseRepo.ResolveCurrentCourseCode(ctx, user.ID, w.defaultCode)
	if err != nil || courseCode == "" {
		return nil
	}
	if !w.aiService.HasSentencePromptsForCourse(courseCode) {
		return nil
	}

	// Local day in the user's timezone (set keys on the user's calendar day).
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")

	// Regeneration guard.
	latest, err := w.repo.LatestSet(user.ID, courseCode)
	if err != nil {
		return err
	}
	if latest != nil {
		if latest.GenerationDate == today {
			return nil // already generated today
		}
		if latest.Status == models.SentenceSetReady {
			return nil // previous set never consumed — don't waste budget
		}
	}

	// Select least-used well-learned words.
	candidates, err := w.repo.SelectCandidateWords(user.ID, courseCode, w.cfg.MasteringThreshold, w.wordsPerSet())
	if err != nil {
		return err
	}
	minWords := w.cfg.MinWords
	if minWords <= 0 {
		minWords = 12
	}
	if len(candidates) < minWords {
		return nil // not enough well-learned words yet
	}

	// Resolve grammar tenses the user has unlocked.
	scopes := w.resolveScopes(user, courseCode)

	words := make([]ai.GenSentenceWord, 0, len(candidates))
	lemmaToID := make(map[string]int64, len(candidates))
	for _, c := range candidates {
		words = append(words, ai.GenSentenceWord{Lemma: c.Lemma, Translation: c.Translation})
		lemmaToID[strings.ToLower(strings.TrimSpace(c.Lemma))] = c.WordCardID
	}

	// Uses the default model (5.5-nano by config); generation is once/day, grading is the hot path.
	sentences, err := w.aiService.GenerateSentenceSetForCourse(ctx, courseCode, words, humanizeScopes(scopes), w.sentencesPerSet())
	if err != nil {
		if w.cbService != nil {
			_ = w.cbService.RecordFailure(err.Error())
		}
		return err
	}
	if w.cbService != nil {
		_ = w.cbService.RecordSuccess()
	}

	items := make([]models.SentenceItem, 0, len(sentences))
	usedSet := map[int64]bool{}
	for i, s := range sentences {
		var ids []int64
		for _, lemma := range s.UsedWords {
			if id, ok := lemmaToID[strings.ToLower(strings.TrimSpace(lemma))]; ok {
				ids = append(ids, id)
				usedSet[id] = true
			}
		}
		items = append(items, models.SentenceItem{
			Position:    i,
			PromptRU:    s.PromptRU,
			ReferenceES: s.ReferenceES,
			WordCardIDs: ids,
		})
	}
	usedIDs := make([]int64, 0, len(usedSet))
	for id := range usedSet {
		usedIDs = append(usedIDs, id)
	}

	set := &models.SentenceSet{
		UserID:         user.ID,
		CourseCode:     courseCode,
		GenerationDate: today,
		Scopes:         scopes,
	}
	if _, err := w.repo.CreateSet(set, items, usedIDs); err != nil {
		return err
	}
	w.logger.Info("sentence set generated",
		zap.Int64("user_id", user.ID),
		zap.String("course", courseCode),
		zap.String("date", today),
		zap.Int("sentences", len(items)))
	return nil
}

func (w *SentenceCompositionWorker) resolveScopes(user *models.User, courseCode string) []string {
	lc := learningForCourseCode(w.learning, courseCode)
	var settings models.UserSettings
	if strings.TrimSpace(user.SettingsJSON) != "" {
		_ = json.Unmarshal([]byte(user.SettingsJSON), &settings)
	}
	return ResolveVerbScopes(&settings, lc)
}

func (w *SentenceCompositionWorker) wordsPerSet() int {
	if w.cfg.WordsPerSet > 0 {
		return w.cfg.WordsPerSet
	}
	return 40
}

func (w *SentenceCompositionWorker) sentencesPerSet() int {
	if w.cfg.SentencesPerSet > 0 {
		return w.cfg.SentencesPerSet
	}
	return 20
}

// learningForCourseCode derives a minimal learning config (target lang/pair) from a course
// code like "es_ru". Falls back to the provided config when the code is malformed.
func learningForCourseCode(fallback config.LearningConfig, courseCode string) config.LearningConfig {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(courseCode)), "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fallback
	}
	target, native := parts[0], parts[1]
	lc := fallback
	lc.TargetLang = target
	lc.NativeLang = native
	lc.Pair = native + "-" + target
	return lc
}

// humanizeScopes turns verb scope ids like "es.presente.indicativo" into readable
// "presente (indicativo)" hints for the generation prompt.
func humanizeScopes(scopes []string) []string {
	out := make([]string, 0, len(scopes))
	for _, s := range scopes {
		parts := strings.Split(s, ".")
		switch len(parts) {
		case 3:
			out = append(out, parts[1]+" ("+parts[2]+")")
		case 2:
			out = append(out, parts[1])
		default:
			out = append(out, s)
		}
	}
	return out
}
