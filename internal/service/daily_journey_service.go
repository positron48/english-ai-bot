package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	mathrand "math/rand"
	"time"

	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
)

// DailyJourneyAdapter reuses a subject engine. Candidates/Validate must be
// read-only and must not start sessions or call providers. Apply writes only via
// tx; it must not make network calls. AI modes need a separate reservation path
// before registering an adapter; they cannot call providers inside Apply.
// Every successful answer is saved here, not via a second client completion call.
type DailyJourneyAdapter interface {
	Candidates(context.Context, database.ContextDBTX, int64, string) ([]dailyjourney.Step, error)
	Validate(context.Context, database.ContextDBTX, int64, string, *dailyjourney.Step) error
	Apply(context.Context, *sql.Tx, int64, *dailyjourney.Journey, *dailyjourney.Step, dailyjourney.Action) error
}

type DailyJourneyService struct {
	db       *sql.DB
	repo     *repository.DailyJourneyRepository
	courses  *repository.CourseRepository
	adapters map[dailyjourney.Mode]DailyJourneyAdapter
	now      func() time.Time
}

func NewDailyJourneyService(db *sql.DB, adapters map[dailyjourney.Mode]DailyJourneyAdapter) *DailyJourneyService {
	registry := map[dailyjourney.Mode]DailyJourneyAdapter{}
	for mode, adapter := range adapters {
		registry[mode] = adapter
	}
	return &DailyJourneyService{db: db, repo: repository.NewDailyJourneyRepository(db), courses: repository.NewCourseRepository(db, nil), adapters: registry, now: time.Now}
}

// Ready deliberately keeps the production entrance disabled until ALL seven
// modes have real adapters. Course-specific unavailability is handled by an
// adapter's empty Candidates result, not by pretending an adapter exists.
func (s *DailyJourneyService) Ready() bool {
	if s == nil {
		return false
	}
	for _, m := range journeyModes {
		if s.adapters[m] == nil {
			return false
		}
	}
	return true
}

var journeyModes = []dailyjourney.Mode{dailyjourney.Words, dailyjourney.Reading, dailyjourney.Grammar, dailyjourney.GrammarSRS, dailyjourney.VerbForms, dailyjourney.Picture, dailyjourney.Conversation}

func (s *DailyJourneyService) Current(ctx context.Context, userID int64, course string) (*dailyjourney.Journey, error) {
	return s.repo.Current(ctx, userID, course, s.now().Format("2006-01-02"))
}
func (s *DailyJourneyService) Get(ctx context.Context, userID int64, course, id string) (*dailyjourney.Journey, error) {
	return s.repo.Get(ctx, userID, course, id)
}
func (s *DailyJourneyService) Start(ctx context.Context, userID int64, course string) (*dailyjourney.Journey, error) {
	if _, err := s.courses.EnsureUserCourse(ctx, userID, course); err != nil {
		return nil, err
	}
	return s.repo.Start(ctx, userID, course, s.now(), func(tx *sql.Tx) ([]dailyjourney.Step, error) {
		candidates := []dailyjourney.Step{}
		for _, mode := range journeyModes {
			adapter := s.adapters[mode]
			if adapter == nil {
				continue
			}
			items, err := adapter.Candidates(ctx, tx, userID, course)
			if err != nil {
				return nil, fmt.Errorf("daily journey %s candidates: %w", mode, err)
			}
			for _, item := range items {
				if item.Mode != mode || item.TargetCount < 1 || !validJourneyLaunch(item) || !json.Valid(item.ContentRef) || item.ContentVersion == "" {
					return nil, fmt.Errorf("invalid daily journey %s candidate", mode)
				}
				// Adapters supply content and limits, not already awarded progress.
				item.ID = ""
				item.CompletedCount = 0
				item.Result = nil
				item.Feedback = nil
				item.ResultDetails = nil
				candidates = append(candidates, item)
			}
		}
		var seed [8]byte
		if _, err := rand.Read(seed[:]); err != nil {
			return nil, err
		}
		steps := dailyjourney.Select(candidates, mathrand.New(mathrand.NewSource(int64(binary.LittleEndian.Uint64(seed[:])))))
		for i := range steps {
			if steps[i].Status == "unavailable" {
				if steps[i].Mode == dailyjourney.Words {
					steps[i].Title = "Слова"
				} else {
					steps[i].Title = "Чтение"
				}
				reason := "Нет доступного содержимого"
				steps[i].Result = &reason
			}
		}
		return steps, nil
	})
}

// start is also resume: it never creates an unrelated global active session.
// The adapter initializes/retrieves its bound session in Apply using the same tx.
func (s *DailyJourneyService) Act(ctx context.Context, userID int64, course, id, stepID string, action dailyjourney.Action) (*dailyjourney.Journey, error) {
	switch action.Type {
	case "start", "skip", "add_word", "answer", "message", "ack_feedback":
	default:
		return nil, dailyjourney.ErrInvalidAction
	}
	// Selection is tentative; ReserveAI rechecks owner/revision/current step under lock.
	if action.Type == "start" || action.Type == "message" {
		j, err := s.repo.Get(ctx, userID, course, id)
		if err != nil {
			return nil, err
		}
		if step := j.Current(); step != nil && step.ID == stepID {
			if adapter, ok := s.adapters[step.Mode].(DailyJourneyAIAdapter); ok && adapter.RequiresAI(action) {
				return s.actAI(ctx, userID, course, id, stepID, action, adapter)
			}
		}
	}
	return s.repo.Apply(ctx, userID, course, id, stepID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
		if j.AlgorithmVersion != dailyjourney.AlgorithmVersion {
			return dailyjourney.ErrConflict
		}
		if action.Type == "skip" {
			if !j.CanSkip(step) {
				return dailyjourney.ErrInvalidAction
			}
			if adapter, ok := s.adapters[step.Mode].(interface {
				Skip(context.Context, *sql.Tx, int64, *dailyjourney.Journey, *dailyjourney.Step) error
			}); ok {
				if err := adapter.Skip(ctx, tx, userID, j, step); err != nil {
					return err
				}
			}
			step.Status = "skipped"
			step.Feedback = nil
			step.Result = nil
			step.ResultDetails = nil
			return nil
		}
		adapter := s.adapters[step.Mode]
		if adapter == nil {
			return fmt.Errorf("daily journey adapter %s not configured", step.Mode)
		}
		if err := adapter.Validate(ctx, tx, userID, course, step); err != nil {
			if errors.Is(err, ErrDailyJourneyContentUnavailable) {
				step.Status = "unavailable"
				reason := "Задание больше недоступно"
				step.Result = &reason
				return nil
			}
			return err
		}
		if action.Type != "start" && step.Status != "active" {
			return dailyjourney.ErrInvalidAction
		}
		if err := adapter.Apply(ctx, tx, userID, j, step, action); err != nil {
			return err
		}
		if action.Type == "start" {
			step.Status = "active"
		}
		return nil
	})
}

var ErrDailyJourneyContentUnavailable = errors.New("daily journey content no longer available")

// Route names and parameter keys are code-controlled; content cannot inject
// an arbitrary navigation URL or route. User/course/step checks still happen
// on the exercise API; a valid route is not an authorization token.
func validJourneyLaunch(step dailyjourney.Step) bool {
	if step.Launch == nil {
		return false
	}
	name := step.Launch.RouteName
	requiredParam := ""
	switch step.Mode {
	case dailyjourney.Words:
		if name != "Training" && name != "WordSets" {
			return false
		}
	case dailyjourney.Reading:
		if name != "ReadingText" {
			return false
		}
		requiredParam = "textId"
	case dailyjourney.Grammar:
		switch name {
		case "GrammarChapter", "GrammarChapterTest":
			requiredParam = "chapterId"
		case "GrammarCategoryTest":
			requiredParam = "sectionId"
		default:
			return false
		}
	case dailyjourney.GrammarSRS:
		if name != "GrammarTraining" {
			return false
		}
	case dailyjourney.VerbForms:
		if name != "VerbTraining" {
			return false
		}
	case dailyjourney.Picture:
		if name != "PictureQuestChat" {
			return false
		}
		requiredParam = "questCode"
	case dailyjourney.Conversation:
		if name != "PlaceChat" || step.Launch.Params["districtCode"] == "" {
			return false
		}
		requiredParam = "scenarioCode"
	default:
		return false
	}
	for key, value := range step.Launch.Params {
		if value == "" || (key != requiredParam && (step.Mode != dailyjourney.Conversation || key != "districtCode")) {
			return false
		}
	}
	return requiredParam == "" || step.Launch.Params[requiredParam] != ""
}

// DailyJourneyExerciseAdapter exposes a subject's existing renderer contract,
// derived from the private persisted state, never a client-provided queue.
type DailyJourneyExerciseAdapter interface {
	Exercise(context.Context, int64, *dailyjourney.Journey, *dailyjourney.Step) (any, error)
}
type DailyJourneyExerciseView struct {
	Journey *dailyjourney.Journey `json:"journey"`
	StepID  string                `json:"step_id"`
	Mode    dailyjourney.Mode     `json:"mode"`
	Data    any                   `json:"data"`
}

func (s *DailyJourneyService) Exercise(ctx context.Context, user int64, course, id, stepID string) (*DailyJourneyExerciseView, error) {
	j, err := s.repo.Get(ctx, user, course, id)
	if err != nil {
		return nil, err
	}
	step := j.Current()
	if j.Status != "active" || step == nil || step.ID != stepID || step.Status != "active" {
		return nil, dailyjourney.ErrInvalidAction
	}
	adapter := s.adapters[step.Mode]
	if adapter == nil {
		return nil, dailyjourney.ErrInvalidAction
	}
	if err := adapter.Validate(ctx, s.db, user, course, step); err != nil {
		return nil, err
	}
	renderer, ok := adapter.(DailyJourneyExerciseAdapter)
	if !ok {
		return nil, dailyjourney.ErrInvalidAction
	}
	data, err := renderer.Exercise(ctx, user, j, step)
	if err != nil {
		return nil, err
	}
	return &DailyJourneyExerciseView{Journey: j, StepID: step.ID, Mode: step.Mode, Data: data}, nil
}
