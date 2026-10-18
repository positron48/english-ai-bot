package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
)

type DailyJourneyGrammarSRSAdapter struct {
	*DailyJourneyGrammarAdapter
	MirrorSRS bool
}

func NewDailyJourneyGrammarSRSAdapter(db *sql.DB, resolve func(string) *GrammarService) *DailyJourneyGrammarSRSAdapter {
	return &DailyJourneyGrammarSRSAdapter{DailyJourneyGrammarAdapter: NewDailyJourneyGrammarAdapter(db, resolve)}
}
func (a *DailyJourneyGrammarSRSAdapter) pool(ctx context.Context, q database.ContextDBTX, user int64, course string) (*GrammarService, map[string][]map[string]interface{}, string, error) {
	if _, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course); err != nil {
		return nil, nil, "", err
	}
	s, err := a.engine(q, course)
	if err != nil {
		return nil, nil, "", err
	}
	if s.TrainingPackRepo == nil {
		return s, map[string][]map[string]interface{}{}, "", nil
	}
	byBlock, err := s.TrainingPackRepo.QuestionsByTheoryBlock()
	if errors.Is(err, repository.ErrGrammarTrainingEmpty) {
		return s, map[string][]map[string]interface{}{}, "", nil
	}
	if err != nil {
		return nil, nil, "", err
	}
	allowed, err := s.allowedTrainingChapters(ctx, user)
	if err != nil {
		return nil, nil, "", err
	}
	for id := range allowed {
		c, err := s.ContentRepo.GetChapter(id)
		if err != nil {
			return nil, nil, "", err
		}
		pub, err := s.PublishRepo.IsPublished("chapter", id)
		if err != nil {
			return nil, nil, "", err
		}
		sectionPub, err := s.PublishRepo.IsPublished("section", c.SectionID)
		if err != nil {
			return nil, nil, "", err
		}
		access, err := s.CanAccessChapter(ctx, user, id)
		if err != nil {
			return nil, nil, "", err
		}
		if !pub || !sectionPub || !access {
			delete(allowed, id)
		}
	}
	filtered := s.filterBlocksByAllowedChapters(byBlock, allowed)
	// Include theory in the content version; saved help must match the bound questions.
	chapters := map[string]*repository.Chapter{}
	for id := range allowed {
		c, err := s.ContentRepo.GetChapter(id)
		if err != nil {
			return nil, nil, "", err
		}
		chapters[id] = c
	}
	raw, _ := json.Marshal([]interface{}{filtered, chapters})
	sum := sha256.Sum256(raw)
	return s, filtered, hex.EncodeToString(sum[:]), nil
}
func (a *DailyJourneyGrammarSRSAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	_, pool, version, err := a.pool(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	count := len(pool)
	if count == 0 {
		return nil, nil
	}
	if count > 5 {
		count = 5
	}
	ref, _ := json.Marshal(journeyGrammarRef{Kind: "srs"})
	return []dailyjourney.Step{{Mode: dailyjourney.GrammarSRS, Title: "Тренировка грамматики", Goal: "Повторить доступные правила", TargetCount: count, ContentRef: ref, ContentVersion: version, Launch: &dailyjourney.Launch{RouteName: "GrammarTraining"}}}, nil
}
func (a *DailyJourneyGrammarSRSAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	_, pool, version, err := a.pool(ctx, q, user, course)
	if err != nil {
		return err
	}
	if len(pool) == 0 || version != step.ContentVersion {
		return ErrDailyJourneyContentUnavailable
	}
	return nil
}
func (a *DailyJourneyGrammarSRSAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	s, pool, version, err := a.pool(ctx, tx, user, j.CourseCode)
	if err != nil {
		return err
	}
	if version != step.ContentVersion {
		return ErrDailyJourneyContentUnavailable
	}
	var state journeyGrammarState
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	switch action.Type {
	case "start":
		if state.Questions != nil {
			break
		}
		ids := []string{}
		for id := range pool {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		args := []interface{}{user, s.learning.TargetLang, s.learning.GrammarBundleID}
		for _, id := range ids {
			args = append(args, id)
		}
		rows, err := tx.QueryContext(ctx, `SELECT theory_block_id FROM grammar_theory_memory WHERE user_id=? AND language=? AND course_id=? AND next_review_at<=CURRENT_TIMESTAMP AND theory_block_id IN (`+strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")+`) ORDER BY next_review_at,theory_block_id`, args...)
		if err != nil {
			return err
		}
		blocks := []string{}
		used := map[string]bool{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			blocks = append(blocks, id)
			used[id] = true
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		rest := []string{}
		for id := range pool {
			if !used[id] {
				rest = append(rest, id)
			}
		}
		sort.Strings(rest)
		rand.Shuffle(len(rest), func(i, k int) { rest[i], rest[k] = rest[k], rest[i] })
		blocks = append(blocks, rest...)
		if len(blocks) > step.TargetCount {
			blocks = blocks[:step.TargetCount]
		}
		state.Answers = map[int]interface{}{}
		state.Results = []map[string]interface{}{}
		state.Questions = []map[string]interface{}{}
		state.Chapters = map[string]*repository.Chapter{}
		for _, block := range blocks {
			questions := pool[block]
			question := questions[rand.Intn(len(questions))]
			state.Questions = append(state.Questions, question)
			id, _ := question["chapter_id"].(string)
			chapter, err := s.ContentRepo.GetChapter(id)
			if err != nil {
				return err
			}
			clone := *chapter
			clone.QuestionBank = nil
			clone.ChapterTest = nil
			clone.Blocks = []interface{}{}
			for _, item := range chapter.Blocks {
				if b, ok := item.(map[string]interface{}); ok && b["type"] == "theory" {
					clone.Blocks = append(clone.Blocks, b)
				}
			}
			state.Chapters[id] = &clone
		}
		if len(state.Questions) == 0 {
			return ErrDailyJourneyContentUnavailable
		}
	case "answer":
		if state.Questions == nil || state.Feedback || state.Index >= len(state.Questions) {
			return dailyjourney.ErrInvalidAction
		}
		var payload struct {
			Index  int         `json:"index"`
			Answer interface{} `json:"answer"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if payload.Index != state.Index || payload.Answer == nil {
			return dailyjourney.ErrInvalidAction
		}
		question := state.Questions[state.Index]
		correct := s.compareQuestionAnswer(payload.Answer, question)
		if question["type"] == "true_false" {
			u, okU := normalizeTrueFalseValue(payload.Answer)
			c, okC := normalizeTrueFalseValue(question["correct_answer"])
			correct = okU && okC && u == c
		}
		chapterID, _ := question["chapter_id"].(string)
		block, _ := question["theory_block_id"].(string)
		concept, _ := question["concept_id"].(string)
		qid, _ := question["id"].(string)
		key := fmt.Sprintf("daily-journey:%s:%s:%d", j.ID, step.ID, state.Index)
		now := time.Now()
		attempt, err := s.SRSRepo.RecordAnswerTx(tx, user, s.learning.TargetLang, s.learning.GrammarBundleID, chapterID, block, concept, qid, payload.Answer, question["correct_answer"], correct, key, now)
		if err != nil {
			return err
		}
		if a.MirrorSRS {
			if err := repository.NewLinglowGrammarSRSBackfillRepository(tx).SyncTheoryBlockForUser(ctx, s.learning, user, chapterID, block); err != nil {
				return err
			}
		}
		answerJSON, _ := json.Marshal(payload.Answer)
		correctJSON, _ := json.Marshal(question["correct_answer"])
		if _, err := repository.NewLinglowEventRepository(a.db).RecordGrammarTrainingAttemptTx(ctx, tx, s.learning, repository.GrammarTrainingEventInput{UserID: user, AttemptID: attempt, ChapterID: chapterID, TheoryBlockID: block, ConceptID: concept, QuestionID: qid, IsCorrect: correct, AnswerJSON: string(answerJSON), CorrectJSON: string(correctJSON), ClientAttemptID: key, AnsweredAt: now}); err != nil {
			return err
		}
		state.Answers[state.Index] = payload.Answer
		resultCorrect := question["correct_answer"]
		if question["type"] == "true_false" {
			if normalized, ok := normalizeTrueFalseValue(resultCorrect); ok {
				resultCorrect = normalized
			}
		}
		state.Results = append(state.Results, map[string]interface{}{"correct": correct, "correct_answer": resultCorrect, "explanation": question["explanation"], "user_answer": payload.Answer})
		state.Feedback = true
		step.CompletedCount = len(state.Answers)
	case "ack_feedback":
		if !state.Feedback {
			return dailyjourney.ErrInvalidAction
		}
		state.Index++
		state.Feedback = false
		if state.Index == len(state.Questions) {
			correct := 0
			for _, result := range state.Results {
				if result["correct"] == true {
					correct++
				}
			}
			step.Status = "completed"
			step.ResultDetails, _ = json.Marshal(map[string]interface{}{"answered": len(state.Answers), "correct": correct})
			state.Result = map[string]interface{}{"correct": correct, "total": len(state.Answers)}
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	step.State, err = json.Marshal(state)
	return err
}
