package web

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

const journeySentenceLimit = 10

type dailyJourneySentencesAdapter struct {
	router *Router
	run    func(context.Context, string, string, string, string, string, ...string) (*ai.SentenceGrade, error)
}
type journeySentenceRef struct {
	SetID int64   `json:"set_id"`
	Items []int64 `json:"items"`
}
type journeySentenceState struct {
	Index int `json:"index"`
}
type journeySentenceInput struct {
	Course string
	Item   models.SentenceItem
	Text   string
}

func (a *dailyJourneySentencesAdapter) access(ctx context.Context, q database.ContextDBTX, user int64, course string) error {
	var tier string
	err := q.QueryRowContext(ctx, `SELECT COALESCE(u.subscription_tier,'free') FROM users u JOIN user_courses uc ON uc.user_id=u.id JOIN courses c ON c.id=uc.course_id WHERE u.id=? AND c.code=? AND uc.status='active' AND c.status='active'`, user, course).Scan(&tier)
	if err == sql.ErrNoRows {
		return dailyjourney.ErrAccess
	}
	if err != nil {
		return err
	}
	if !a.router.config.SentenceComposition.Enabled || !models.TierAllowsFeature(models.ParseUserTier(tier), "sentence_composition") {
		return service.ErrDailyJourneyContentUnavailable
	}
	if a.run == nil && (a.router.aiSvc() == nil || !a.router.aiSvc().HasSentencePromptsForCourse(course)) {
		return service.ErrDailyJourneyContentUnavailable
	}
	return nil
}
func (a *dailyJourneySentencesAdapter) repo(q database.ContextDBTX) *repository.SentenceCompositionRepository {
	return a.router.sentenceRepo().WithQueries(q.(database.DBTX))
}
func journeySentenceVersion(items []models.SentenceItem) string {
	content := make([]any, 0, len(items))
	for _, it := range items {
		content = append(content, []any{it.ID, it.SetID, it.PromptRU, it.ClarificationRU, it.ReferenceES})
	}
	raw, _ := json.Marshal(content)
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}
func (a *dailyJourneySentencesAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	if err := a.access(ctx, q, user, course); err != nil {
		if errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
			return nil, nil
		}
		return nil, err
	}
	repo := a.repo(q)
	set, err := repo.LatestSet(user, course)
	if err != nil {
		return nil, err
	}
	if set == nil || set.Status == models.SentenceSetCompleted {
		return nil, nil
	}
	items, err := repo.GetItems(set.ID)
	if err != nil {
		return nil, err
	}
	ref := journeySentenceRef{SetID: set.ID}
	selected := []models.SentenceItem{}
	for _, it := range items {
		if it.AttemptedAt != nil {
			continue
		}
		ref.Items = append(ref.Items, it.ID)
		selected = append(selected, it)
		if len(selected) == journeySentenceLimit {
			break
		}
	}
	if len(selected) == 0 {
		return nil, nil
	}
	raw, _ := json.Marshal(ref)
	return []dailyjourney.Step{{Mode: dailyjourney.Sentences, Title: "Составление предложений", Goal: "Составить предложения", TargetCount: len(selected), ContentRef: raw, ContentVersion: journeySentenceVersion(selected), Launch: &dailyjourney.Launch{RouteName: "SentenceComposition"}}}, nil
}
func (a *dailyJourneySentencesAdapter) bound(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) ([]models.SentenceItem, journeySentenceState, error) {
	var ref journeySentenceRef
	var state journeySentenceState
	if err := a.access(ctx, q, user, course); err != nil {
		return nil, state, err
	}
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return nil, state, err
	}
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return nil, state, err
		}
	}
	if len(ref.Items) == 0 || len(ref.Items) > journeySentenceLimit || len(ref.Items) != step.TargetCount || state.Index < 0 || state.Index > len(ref.Items) {
		return nil, state, service.ErrDailyJourneyContentUnavailable
	}
	repo := a.repo(q)
	set, err := repo.GetSetByID(ref.SetID)
	if err != nil {
		return nil, state, err
	}
	if set == nil || set.UserID != user || set.CourseCode != course {
		return nil, state, service.ErrDailyJourneyContentUnavailable
	}
	all, err := repo.GetItems(set.ID)
	if err != nil {
		return nil, state, err
	}
	byID := map[int64]models.SentenceItem{}
	for _, it := range all {
		byID[it.ID] = it
	}
	items := make([]models.SentenceItem, 0, len(ref.Items))
	for _, id := range ref.Items {
		it, ok := byID[id]
		if !ok {
			return nil, state, service.ErrDailyJourneyContentUnavailable
		}
		items = append(items, it)
		delete(byID, id)
	}
	if journeySentenceVersion(items) != step.ContentVersion {
		return nil, state, service.ErrDailyJourneyContentUnavailable
	}
	return items, state, nil
}
func (a *dailyJourneySentencesAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	_, _, err := a.bound(ctx, q, user, course, step)
	return err
}
func (a *dailyJourneySentencesAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	items, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	switch action.Type {
	case "start":
		if err := a.repo(tx).MarkStarted(items[0].SetID); err != nil {
			return err
		}
	case "ack_feedback":
		if state.Index >= len(items) || items[state.Index].AttemptedAt == nil {
			return dailyjourney.ErrInvalidAction
		}
		state.Index++
		if state.Index == len(items) {
			step.Status = "completed"
			correct := 0
			for _, it := range items {
				if it.Outcome != nil && *it.Outcome != models.SentenceOutcomeFailed {
					correct++
				}
			}
			step.ResultDetails, _ = json.Marshal(map[string]any{"answered": len(items), "correct": correct})
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	step.CompletedCount = state.Index
	if state.Index < len(items) && items[state.Index].AttemptedAt != nil {
		step.CompletedCount++
	}
	step.State, err = json.Marshal(state)
	return err
}
func (a *dailyJourneySentencesAdapter) RequiresAI(action dailyjourney.Action) bool {
	return action.Type == "answer"
}
func (a *dailyJourneySentencesAdapter) PrepareAI(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) (json.RawMessage, error) {
	items, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return nil, err
	}
	var payload struct {
		ItemID int64  `json:"item_id"`
		Text   string `json:"user_input"`
	}
	if json.Unmarshal(action.Payload, &payload) != nil || state.Index >= len(items) || payload.ItemID != items[state.Index].ID || strings.TrimSpace(payload.Text) == "" || len(payload.Text) > 10000 {
		return nil, dailyjourney.ErrInvalidAction
	}
	return json.Marshal(journeySentenceInput{Course: j.CourseCode, Item: items[state.Index], Text: strings.TrimSpace(payload.Text)})
}
func (a *dailyJourneySentencesAdapter) RunAI(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var input journeySentenceInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, err
	}
	// Another entrance may have answered the same pinned item. Show that saved result.
	if input.Item.AttemptedAt != nil {
		return json.RawMessage(`{}`), nil
	}
	var grade *ai.SentenceGrade
	var err error
	if ai.NormalizedSentenceAnswer(input.Text) == ai.NormalizedSentenceAnswer(input.Item.ReferenceES) {
		grade = ai.NewExactSentenceGrade(input.Text)
	} else {
		run := a.run
		if run == nil {
			run = a.router.aiSvc().GradeSentenceForCourse
		}
		grade, err = run(ctx, input.Course, input.Item.PromptRU, input.Item.ClarificationRU, input.Item.ReferenceES, input.Text)
	}
	if err != nil {
		return nil, err
	}
	if grade == nil {
		return nil, errors.New("empty sentence grade")
	}
	grade.Outcome = models.OutcomeForErrorCount(grade.ErrorCount)
	return json.Marshal(grade)
}
func (a *dailyJourneySentencesAdapter) ApplyAI(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action, raw json.RawMessage) error {
	// Serialize with standalone answers before re-reading the canonical attempt.
	var ref journeySentenceRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM sentence_sets WHERE id=? FOR UPDATE`, ref.SetID).Scan(&id); err != nil {
		return err
	}
	items, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	if state.Index >= len(items) {
		return dailyjourney.ErrInvalidAction
	}
	item := items[state.Index]
	if item.AttemptedAt == nil {
		var payload struct {
			ItemID int64  `json:"item_id"`
			Text   string `json:"user_input"`
		}
		var grade ai.SentenceGrade
		if json.Unmarshal(action.Payload, &payload) != nil || payload.ItemID != item.ID || json.Unmarshal(raw, &grade) != nil || grade.CorrectedES == "" {
			return dailyjourney.ErrInvalidAction
		}
		grade.Outcome = models.OutcomeForErrorCount(grade.ErrorCount)
		grading, err := json.Marshal(grade)
		if err != nil {
			return err
		}
		if _, err := a.repo(tx).RecordAttemptTx(tx, item.ID, strings.TrimSpace(payload.Text), grade.ErrorCount, grade.Outcome, string(grading)); err != nil {
			return err
		}
		stats := repository.NewLinglowDailyStatsRepository(tx)
		uc, err := stats.ResolveUserCourseID(ctx, user, j.CourseCode)
		if err != nil {
			return err
		}
		correct := 0
		if grade.Outcome != models.SentenceOutcomeFailed {
			correct = 1
		}
		if err := stats.Bump(ctx, repository.DailyBump{UserCourseID: uc, Day: repository.LocalDayFromTime(time.Now()), Mode: "sentences", Attempts: 1, Correct: correct}); err != nil {
			return err
		}
	}
	step.CompletedCount = state.Index + 1
	return nil
}
func (a *dailyJourneySentencesAdapter) Exercise(ctx context.Context, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (any, error) {
	return a.exercise(ctx, a.router.db, user, j, step)
}
func (a *dailyJourneySentencesAdapter) ExerciseTx(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (any, error) {
	return a.exercise(ctx, tx, user, j, step)
}
func (a *dailyJourneySentencesAdapter) exercise(ctx context.Context, q database.ContextDBTX, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (any, error) {
	items, state, err := a.bound(ctx, q, user, j.CourseCode, step)
	if err != nil {
		return nil, err
	}
	summary := map[string]any{"status": "started", "total": len(items), "attempted": step.CompletedCount, "stars": 0, "passed": 0, "failed": 0}
	for _, it := range items {
		if it.Outcome == nil {
			continue
		}
		key := *it.Outcome
		if key == "star" {
			key = "stars"
		}
		if value, ok := summary[key].(int); ok {
			summary[key] = value + 1
		}
	}
	var current any
	if state.Index < len(items) {
		current = sentenceItemJSON(items[state.Index], true)
	} else {
		summary["status"] = "completed"
	}
	return map[string]any{"current": current, "set": summary, "done": state.Index == len(items)}, nil
}
