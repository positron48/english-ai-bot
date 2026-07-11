package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// SentenceCompositionWorker generates sentence-composition sets for Pro users on an hourly tick.
// Regeneration guard: while the latest set for a user/course is not completed (ready or started),
// no new set is created. After completion, the next hourly tick may create another set, even
// on the same calendar day.
type SentenceCompositionWorker struct {
	aiService       *ai.Service
	repo            *repository.SentenceCompositionRepository
	userRepo        *repository.UserRepository
	courseRepo      *repository.CourseRepository
	cbService       circuitBreakerForWorker
	cfg             config.SentenceCompositionConfig
	learning        config.LearningConfig
	grammarByBundle map[string]*repository.GrammarContentRepository
	grammarAttempts *repository.GrammarAttemptRepository
	defaultCode     string
	logger          *zap.Logger
	stopChan        chan struct{}
}

func NewSentenceCompositionWorker(
	aiService *ai.Service,
	repo *repository.SentenceCompositionRepository,
	userRepo *repository.UserRepository,
	courseRepo *repository.CourseRepository,
	cbService circuitBreakerForWorker,
	cfg config.SentenceCompositionConfig,
	learning config.LearningConfig,
	grammarByBundle map[string]*repository.GrammarContentRepository,
	grammarAttempts *repository.GrammarAttemptRepository,
	defaultCourseCode string,
	logger *zap.Logger,
) *SentenceCompositionWorker {
	return &SentenceCompositionWorker{
		aiService:       aiService,
		repo:            repo,
		userRepo:        userRepo,
		courseRepo:      courseRepo,
		cbService:       cbService,
		cfg:             cfg,
		learning:        learning,
		grammarByBundle: grammarByBundle,
		grammarAttempts: grammarAttempts,
		defaultCode:     defaultCourseCode,
		logger:          logger,
		stopChan:        make(chan struct{}),
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

	courseCodes, err := w.sentenceCourseCodesForUser(ctx, user.ID)
	if err != nil {
		return err
	}
	if len(courseCodes) == 0 {
		return nil
	}

	// Local day in the user's timezone (set keys on the user's calendar day).
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")

	for _, courseCode := range courseCodes {
		if !w.aiService.HasSentencePromptsForCourse(courseCode) {
			continue
		}

		// Regeneration guard is per user/course: skip while an incomplete set exists.
		latest, err := w.repo.LatestSet(user.ID, courseCode)
		if err != nil {
			return err
		}
		if !shouldGenerateSentenceSet(latest) {
			continue
		}

		if _, err := w.generateSet(ctx, user, courseCode, today); err != nil {
			w.logger.Warn("sentence worker: course generation failed",
				zap.Int64("user_id", user.ID), zap.String("course", courseCode), zap.Error(err))
		}
	}
	return nil
}

// shouldGenerateSentenceSet returns true when the worker may create a new set for the course.
// No prior set, or the latest set is completed — generate. Ready or started — wait.
func shouldGenerateSentenceSet(latest *models.SentenceSet) bool {
	if latest == nil {
		return true
	}
	return latest.Status == models.SentenceSetCompleted
}

func (w *SentenceCompositionWorker) sentenceCourseCodesForUser(ctx context.Context, userID int64) ([]string, error) {
	courses, err := w.courseRepo.ListCoursesForUser(ctx, userID, w.defaultCode)
	if err != nil {
		courseCode, resolveErr := w.courseRepo.ResolveCurrentCourseCode(ctx, userID, w.defaultCode)
		if resolveErr != nil || strings.TrimSpace(courseCode) == "" {
			return nil, err
		}
		return []string{strings.ToLower(strings.TrimSpace(courseCode))}, nil
	}
	out := make([]string, 0, len(courses))
	seen := map[string]bool{}
	for _, course := range courses {
		if course.UserCourseID == nil {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(course.UserStatus))
		if status != "" && status != "active" && status != "completed" {
			continue
		}
		code := strings.ToLower(strings.TrimSpace(course.Code))
		if code == "" || seen[code] {
			continue
		}
		out = append(out, code)
		seen[code] = true
	}
	return out, nil
}

// ForceGenerateForUser generates a fresh set for the user's current course.
func (w *SentenceCompositionWorker) ForceGenerateForUser(ctx context.Context, userID int64) (int64, error) {
	user, err := w.userRepo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, fmt.Errorf("user %d not found", userID)
	}
	courseCode, err := w.courseRepo.ResolveCurrentCourseCode(ctx, user.ID, w.defaultCode)
	if err != nil || courseCode == "" {
		return 0, fmt.Errorf("cannot resolve course for user %d", userID)
	}
	return w.ForceGenerateForUserCourse(ctx, userID, courseCode)
}

// ForceGenerateForUserCourse generates a fresh set for an active user course regardless
// of the daily regeneration guard (admin-triggered).
func (w *SentenceCompositionWorker) ForceGenerateForUserCourse(ctx context.Context, userID int64, courseCode string) (int64, error) {
	user, err := w.userRepo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, fmt.Errorf("user %d not found", userID)
	}
	if !models.ParseUserTier(string(user.SubscriptionTier)).AtLeast(models.TierPro) {
		return 0, fmt.Errorf("user %d is not Pro", userID)
	}
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	courses, err := w.sentenceCourseCodesForUser(ctx, user.ID)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve courses for user %d: %w", userID, err)
	}
	allowed := false
	for _, code := range courses {
		if code == courseCode {
			allowed = true
			break
		}
	}
	if !allowed {
		return 0, fmt.Errorf("course %q is not active for user %d", courseCode, userID)
	}
	if !w.aiService.HasSentencePromptsForCourse(courseCode) {
		return 0, fmt.Errorf("sentence prompts not configured for course %q", courseCode)
	}
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")
	return w.generateSet(ctx, user, courseCode, today)
}

// SentenceCourseCodesForUser returns the active courses eligible for sentence generation.
func (w *SentenceCompositionWorker) SentenceCourseCodesForUser(ctx context.Context, userID int64) ([]string, error) {
	codes, err := w.sentenceCourseCodesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		if w.aiService.HasSentencePromptsForCourse(code) {
			out = append(out, code)
		}
	}
	return out, nil
}

// generateSet builds and persists one sentence set for the user/course, bumping word-usage
// counters. It performs no regeneration-guard checks — the caller decides whether to generate.
// Returns the new set id, or 0 when there are not enough well-learned words yet.
func (w *SentenceCompositionWorker) generateSet(ctx context.Context, user *models.User, courseCode, today string) (int64, error) {
	// Select least-used well-learned words.
	candidates, err := w.repo.SelectCandidateWords(user.ID, courseCode, w.cfg.MasteringThreshold, w.wordsPerSet())
	if err != nil {
		return 0, err
	}
	minWords := w.cfg.MinWords
	if minWords <= 0 {
		minWords = 12
	}
	if len(candidates) < minWords {
		return 0, nil // not enough well-learned words yet
	}

	// Resolve grammar tenses the user has unlocked.
	scopes := w.resolveScopes(user, courseCode)
	if strings.EqualFold(learningForCourseCode(w.learning, courseCode).TargetLang, "en") && len(scopes) == 0 {
		w.logger.Info("sentence set skipped: english grammar scopes unavailable",
			zap.Int64("user_id", user.ID), zap.String("course", courseCode))
		return 0, nil
	}

	words := make([]ai.GenSentenceWord, 0, len(candidates))
	lemmaToID := make(map[string]int64, len(candidates))
	for _, c := range candidates {
		words = append(words, ai.GenSentenceWord{Lemma: c.Lemma, Translation: c.Translation})
		lemmaToID[strings.ToLower(strings.TrimSpace(c.Lemma))] = c.WordCardID
	}

	focusCount := sentenceFocusWordCount(len(words), w.sentencesPerSet())
	focusWords := words[:focusCount]
	supportWords := words[focusCount:]

	// Uses the default model (5.5-nano by config); generation is once/day, grading is the hot path.
	sentences, err := w.aiService.GenerateSentenceSetForCourse(ctx, courseCode, focusWords, supportWords, humanizeScopes(scopes), w.sentencesPerSet())
	if err != nil {
		if w.cbService != nil {
			_ = w.cbService.RecordFailure(err.Error())
		}
		return 0, err
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
			Position:        i,
			PromptRU:        s.PromptRU,
			ClarificationRU: s.ClarificationRU,
			ReferenceES:     s.ReferenceES,
			WordCardIDs:     ids,
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
	setID, err := w.repo.CreateSet(set, items, usedIDs)
	if err != nil {
		return 0, err
	}
	w.logger.Info("sentence set generated",
		zap.Int64("user_id", user.ID),
		zap.String("course", courseCode),
		zap.String("date", today),
		zap.Int("sentences", len(items)))
	return setID, nil
}

func (w *SentenceCompositionWorker) resolveScopes(user *models.User, courseCode string) []string {
	lc := learningForCourseCode(w.learning, courseCode)
	var settings models.UserSettings
	if strings.TrimSpace(user.SettingsJSON) != "" {
		_ = json.Unmarshal([]byte(user.SettingsJSON), &settings)
	}
	if strings.EqualFold(lc.TargetLang, "en") {
		return w.resolveEnglishScopes(user.ID, courseCode)
	}
	return ResolveVerbScopes(&settings, lc)
}

func (w *SentenceCompositionWorker) resolveEnglishScopes(userID int64, courseCode string) []string {
	if w.grammarAttempts == nil {
		return nil
	}
	bundleID := targetLangFromCourseCode(courseCode)
	if bundleID == "" {
		bundleID = "en"
	}
	contentRepo := w.grammarByBundle[strings.ToLower(bundleID)]
	if contentRepo == nil {
		return nil
	}
	sections, err := contentRepo.GetSections()
	if err != nil {
		w.logger.Warn("sentence english scopes: load sections failed",
			zap.Int64("user_id", userID), zap.String("course", courseCode), zap.Error(err))
		return nil
	}
	progress, err := w.grammarAttempts.GetAllChapterProgress(userID)
	if err != nil {
		w.logger.Warn("sentence english scopes: load chapter progress failed",
			zap.Int64("user_id", userID), zap.String("course", courseCode), zap.Error(err))
		return nil
	}
	placement, err := w.grammarAttempts.GetPlacementTestResult(userID)
	if err != nil {
		w.logger.Warn("sentence english scopes: load placement failed",
			zap.Int64("user_id", userID), zap.String("course", courseCode), zap.Error(err))
		return nil
	}
	return ResolveEnglishSentenceScopes(sections, placement, progress)
}

func (w *SentenceCompositionWorker) wordsPerSet() int {
	if w.cfg.WordsPerSet > 0 {
		return w.cfg.WordsPerSet
	}
	return 80
}

// sentenceFocusWordCount keeps a smaller least-used focus list and reserves the
// rest of the user's known vocabulary as natural sentence-building support.
func sentenceFocusWordCount(total, sentenceCount int) int {
	if total <= 0 {
		return 0
	}
	focus := total / 3
	if focus < 1 {
		focus = 1
	}
	if sentenceCount > 0 && focus > sentenceCount {
		focus = sentenceCount
	}
	if focus > total {
		focus = total
	}
	return focus
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

func targetLangFromCourseCode(courseCode string) string {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(courseCode)), "_", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

// humanizeScopes turns verb scope ids like "es.presente.indicativo" into readable
// "presente (indicativo)" hints for the generation prompt.
func humanizeScopes(scopes []string) []string {
	if len(scopes) > 0 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(scopes[0])), "en.") {
		return EnglishSentenceScopeLabels(scopes)
	}
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
