package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
)

type DailyJourneyVerbsAdapter struct {
	db       *sql.DB
	cfg      config.TrainingConfig
	learning config.LearningConfig
	logger   *zap.Logger
}

func NewDailyJourneyVerbsAdapter(db *sql.DB, cfg config.TrainingConfig, lc config.LearningConfig, logger *zap.Logger) *DailyJourneyVerbsAdapter {
	return &DailyJourneyVerbsAdapter{db: db, cfg: cfg, learning: lc, logger: logger}
}

type journeyVerbRef struct {
	Cards    []int64  `json:"cards"`
	Versions []string `json:"versions"`
}

func journeyVerbVersion(c repository.VerbQueueCard) string {
	raw, _ := json.Marshal([]interface{}{c.TrainingCardID, c.WordCardID, c.CardType, c.PromptJSON, c.AnswerJSON, c.DistractorsJSON})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func (a *DailyJourneyVerbsAdapter) access(ctx context.Context, q database.ContextDBTX, user int64, course string) (*repository.VerbFormsRepository, []string, error) {
	if !a.cfg.SpanishVerbFormsEnabled || !strings.HasPrefix(course, "es_") {
		return nil, nil, ErrDailyJourneyContentUnavailable
	}
	if _, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course); err != nil {
		return nil, nil, err
	}
	conn, ok := q.(database.DBTX)
	if !ok {
		return nil, nil, fmt.Errorf("verb transaction query surface missing")
	}
	lc := learningForCourseCode(a.learning, course)
	lc.GrammarBundleID = "es"
	scopes, err := UnlockedVerbScopes(conn, user, lc)
	if err != nil {
		return nil, nil, err
	}
	return repository.NewVerbFormsRepository(a.db, a.logger).WithQueries(conn), scopes, nil
}
func (a *DailyJourneyVerbsAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	if !a.cfg.SpanishVerbFormsEnabled || !strings.HasPrefix(course, "es_") {
		return nil, nil
	}
	repo, scopes, err := a.access(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	queue, err := repo.GetJourneyPracticeQueue(user, scopes, min(a.cfg.VerbFormsMaxCards, 10), a.cfg.VerbFormsMaxNew, course)
	if err != nil {
		return nil, err
	}
	if len(queue) == 0 {
		return nil, nil
	}
	ref := journeyVerbRef{}
	for _, card := range queue {
		ref.Cards = append(ref.Cards, card.TrainingCardID)
		ref.Versions = append(ref.Versions, journeyVerbVersion(card))
	}
	raw, _ := json.Marshal(ref)
	sum := sha256.Sum256(raw)
	return []dailyjourney.Step{{Mode: dailyjourney.VerbForms, Title: "Формы глаголов", Goal: "Пройти короткую тренировку форм", TargetCount: len(queue), ContentRef: raw, ContentVersion: hex.EncodeToString(sum[:]), Launch: &dailyjourney.Launch{RouteName: "VerbTraining"}}}, nil
}
func (a *DailyJourneyVerbsAdapter) bound(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) (*repository.VerbFormsRepository, []string, []repository.VerbQueueCard, error) {
	repo, scopes, err := a.access(ctx, q, user, course)
	if err != nil {
		return nil, nil, nil, err
	}
	var ref journeyVerbRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return nil, nil, nil, err
	}
	if len(ref.Cards) == 0 || len(ref.Cards) != len(ref.Versions) || len(ref.Cards) != step.TargetCount {
		return nil, nil, nil, ErrDailyJourneyContentUnavailable
	}
	var saved repository.VerbPractice
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &saved); err != nil {
			return nil, nil, nil, err
		}
	}
	cards := []repository.VerbQueueCard{}
	for index, id := range ref.Cards {
		card, err := repo.GetJourneyVerbCard(user, id, course)
		if err == sql.ErrNoRows {
			return nil, nil, nil, ErrDailyJourneyContentUnavailable
		}
		if err != nil {
			return nil, nil, nil, err
		}
		if journeyVerbVersion(*card) != ref.Versions[index] || !VerbScopeAllowed(*card, scopes) {
			return nil, nil, nil, ErrDailyJourneyContentUnavailable
		}
		if saved.Queue != nil && (card.UserVerbCardID == 0 || index >= len(saved.Queue) || saved.Queue[index].UserVerbCardID != card.UserVerbCardID) {
			return nil, nil, nil, ErrDailyJourneyContentUnavailable
		}
		cards = append(cards, *card)
	}
	return repo, scopes, cards, nil
}
func (a *DailyJourneyVerbsAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	_, _, _, err := a.bound(ctx, q, user, course, step)
	return err
}
func (a *DailyJourneyVerbsAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	repo, scopes, cards, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	var state repository.VerbPractice
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	switch action.Type {
	case "start":
		if state.Queue == nil {
			state = repository.VerbPractice{Version: 2, ID: journeySessionID(j.ID + ":" + step.ID), Queue: cards, Results: []repository.VerbPracticeFeedback{}}
			rows, err := repo.GetLinkedVerbFormsForUser(user, scopes)
			if err != nil {
				return err
			}
			for index := range state.Queue {
				card := &state.Queue[index]
				if card.UserVerbCardID == 0 {
					id, err := repo.GetOrCreateUserVerbCard(user, card.TrainingCardID)
					if err != nil {
						return err
					}
					card.UserVerbCardID = id
				}
				contrastVerbOptions(card, rows)
			}
		}
		var payload struct {
			Help   bool  `json:"help"`
			Index  int   `json:"index"`
			CardID int64 `json:"card_id"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if payload.Help {
			if state.Index >= len(state.Queue) || payload.Index != state.Index || payload.CardID != state.Queue[state.Index].UserVerbCardID {
				return dailyjourney.ErrInvalidAction
			}
			if state.Feedback == nil {
				state.Assisted = true
			}
		}
	case "answer":
		if state.Queue == nil || state.Feedback != nil || state.Completed || state.Index >= len(state.Queue) {
			return dailyjourney.ErrInvalidAction
		}
		var payload struct {
			Index  int    `json:"index"`
			CardID int64  `json:"card_id"`
			Answer string `json:"answer"`
			Skip   bool   `json:"skip"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		card := state.Queue[state.Index]
		if payload.Index != state.Index || payload.CardID != card.UserVerbCardID || (!payload.Skip && strings.TrimSpace(payload.Answer) == "") {
			return dailyjourney.ErrInvalidAction
		}
		feedback := GradeVerbPractice(card, payload.Answer, payload.Skip, state.Assisted)
		srs, err := repository.ReadVerbSRSTx(tx, user, card.UserVerbCardID)
		if err != nil {
			return err
		}
		due, quality := AdvanceVerbSRS(srs, feedback.Outcome == "correct", state.Assisted)
		reviewID, err := repository.SaveVerbReviewTx(tx, user, nil, false, srs, due, quality, feedback)
		if err != nil {
			return err
		}
		if err := recordJourneyVerbEvent(ctx, tx, user, j.CourseCode, reviewID, feedback); err != nil {
			return err
		}
		state.Feedback = &feedback
		state.Results = append(state.Results, feedback)
		step.CompletedCount = len(state.Results)
	case "ack_feedback":
		if state.Feedback == nil {
			return dailyjourney.ErrInvalidAction
		}
		state.Index++
		state.Feedback = nil
		state.Assisted = false
		state.Completed = state.Index == len(state.Queue)
		if state.Completed {
			step.Status = "completed"
			correct := 0
			for _, result := range state.Results {
				if result.Outcome == "correct" && !result.Assisted {
					correct++
				}
			}
			step.ResultDetails, _ = json.Marshal(map[string]interface{}{"answered": len(state.Results), "correct": correct})
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	step.State, err = json.Marshal(state)
	return err
}
func (a *DailyJourneyVerbsAdapter) Exercise(ctx context.Context, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (interface{}, error) {
	var state repository.VerbPractice
	if err := json.Unmarshal(step.State, &state); err != nil {
		return nil, err
	}
	public := PublicVerbPractice(&state)
	if (state.Assisted || state.Feedback != nil) && state.Index < len(state.Queue) {
		var prompt map[string]interface{}
		_ = json.Unmarshal([]byte(state.Queue[state.Index].PromptJSON), &prompt)
		lemma, _ := prompt["lemma"].(string)
		forms, err := repository.NewVerbFormsRepository(a.db, a.logger).ListVerbFormsForLemma(lemma, "es")
		if err != nil {
			return nil, err
		}
		out := []map[string]interface{}{}
		for _, form := range forms {
			out = append(out, map[string]interface{}{"mood": form.Mood, "tense": form.Tense, "person": form.Person, "number": form.Number, "surface_form": form.SurfaceForm})
		}
		public["forms"] = out
	}
	return public, nil
}
func recordJourneyVerbEvent(ctx context.Context, tx *sql.Tx, user int64, course string, reviewID int64, feedback repository.VerbPracticeFeedback) error {
	var userCourseID int64
	if err := tx.QueryRowContext(ctx, `SELECT uc.id FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=?`, user, course).Scan(&userCourseID); err != nil {
		return err
	}
	raw, _ := json.Marshal(feedback)
	answer, _ := json.Marshal(feedback.Chosen)
	correct := feedback.Outcome == "correct" && !feedback.Assisted
	now := time.Now()
	key := fmt.Sprintf("%d", reviewID)
	var attemptID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO exercise_attempts(user_course_id,mode,started_at,answered_at,is_correct,answer_json,result_json,source_table,source_pk) VALUES(?,'verb_training',?,?,?,CAST(? AS jsonb),CAST(? AS jsonb),'verb_review_events',?) RETURNING id`, userCourseID, now, now, correct, string(answer), string(raw), key).Scan(&attemptID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO learning_events(user_course_id,exercise_attempt_id,event_type,event_time,mode,source_table,source_pk,event_json) VALUES(?,?,'verb_training_answered',?,'verb_training','verb_review_events',?,CAST(? AS jsonb))`, userCourseID, attemptID, now, key, string(raw)); err != nil {
		return err
	}
	n := 0
	if correct {
		n = 1
	}
	return repository.NewLinglowDailyStatsRepository(tx).Bump(ctx, repository.DailyBump{UserCourseID: userCourseID, Day: repository.LocalDayFromTime(now), Mode: "verb_training", Attempts: 1, Correct: n})
}
