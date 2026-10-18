package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
)

type DailyJourneyReadingAdapter struct{ db *sql.DB }

func NewDailyJourneyReadingAdapter(db *sql.DB) *DailyJourneyReadingAdapter {
	return &DailyJourneyReadingAdapter{db: db}
}

type journeyReadingRef struct {
	TextID string `json:"text_id"`
}
type journeyReadingChoice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type journeyReadingQuestion struct {
	Prompt        string                 `json:"prompt"`
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Choices       []journeyReadingChoice `json:"choices,omitempty"`
	CorrectAnswer json.RawMessage        `json:"correct_answer,omitempty"`
	Explanation   json.RawMessage        `json:"explanation,omitempty"`
}
type journeyReadingState struct {
	Document  *repository.ReadingTextDocument `json:"document"`
	Questions []journeyReadingQuestion        `json:"questions"`
	Answers   map[int]string                  `json:"answers"`
	Correct   map[int]bool                    `json:"correct"`
	Index     int                             `json:"index"`
	QuizOpen  bool                            `json:"quiz_open"`
	Feedback  bool                            `json:"feedback"`
	Done      bool                            `json:"done"`
	Passed    bool                            `json:"passed"`
}

// JourneyReadingExercise uses the existing passage/quiz renderer. Correct answers
// only appear for answered questions; the persisted private snapshot stays server-side.
type JourneyReadingExercise struct {
	Document  *repository.ReadingTextDocument `json:"document"`
	Questions []journeyReadingQuestion        `json:"questions"`
	Answers   map[int]string                  `json:"answers"`
	Correct   map[int]bool                    `json:"correct"`
	Index     int                             `json:"index"`
	QuizOpen  bool                            `json:"quiz_open"`
	Feedback  bool                            `json:"feedback"`
	Done      bool                            `json:"done"`
	Passed    bool                            `json:"passed"`
}

func readingVersion(doc *repository.ReadingTextDocument) string {
	data, _ := json.Marshal(doc)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func readingQuestions(doc *repository.ReadingTextDocument) ([]journeyReadingQuestion, error) {
	data, err := json.Marshal(doc.ReadingPassage["comprehension_questions"])
	if err != nil {
		return nil, err
	}
	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(data, &rawQuestions); err != nil {
		return nil, err
	}
	valid := []journeyReadingQuestion{}
	for i, raw := range rawQuestions {
		var fields struct {
			ID          string          `json:"id"`
			Prompt      string          `json:"prompt"`
			Question    string          `json:"question"`
			Type        string          `json:"type"`
			Choices     json.RawMessage `json:"choices"`
			Correct     json.RawMessage `json:"correct_answer"`
			Explanation json.RawMessage `json:"explanation"`
		}
		if err := json.Unmarshal(raw, &fields); err != nil {
			continue
		}
		if fields.Prompt == "" {
			fields.Prompt = fields.Question
		}
		if fields.Prompt == "" {
			continue
		}
		q := journeyReadingQuestion{ID: fmt.Sprintf("reading-%d", i), Prompt: fields.Prompt, Type: fields.Type, CorrectAnswer: fields.Correct, Explanation: fields.Explanation}
		if len(fields.Choices) > 0 && string(fields.Choices) != "null" {
			if err := json.Unmarshal(fields.Choices, &q.Choices); err != nil {
				var texts []string
				if err := json.Unmarshal(fields.Choices, &texts); err != nil {
					continue
				}
				for k, text := range texts {
					id := strconv.Itoa(k)
					if len(text) > 2 && (text[1] == ')' || text[1] == '.' || text[1] == ':') {
						id = text[:1]
						text = strings.TrimSpace(text[2:])
					}
					q.Choices = append(q.Choices, journeyReadingChoice{ID: id, Text: text})
				}
			}
		}
		answer := readingAnswer(q.CorrectAnswer)
		if q.Type == "true_false" || len(q.Choices) == 0 {
			if answer != "true" && answer != "false" {
				continue
			}
			q.Type = "true_false"
		} else {
			q.Type = "mcq_single"
			found := false
			for _, choice := range q.Choices {
				if strings.ToLower(choice.ID) == answer {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		valid = append(valid, q)
	}
	return valid, nil
}
func (a *DailyJourneyReadingAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	levels, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course)
	if err != nil {
		return nil, err
	}
	language := strings.SplitN(course, "_", 2)[0]
	rows, err := q.QueryContext(ctx, `SELECT t.text_id, EXISTS(SELECT 1 FROM reading_text_progress p WHERE p.user_id=? AND p.chapter_id=t.text_id),c.level FROM reading_texts t JOIN reading_categories c ON c.category_id=t.category_id WHERE LOWER(t.target_language)=? AND c.text_ids::jsonb @> jsonb_build_array(t.text_id) ORDER BY 2, RANDOM()`, user, language)
	if err != nil {
		return nil, err
	}
	type entry struct {
		id    string
		read  bool
		level string
	}
	entries := []entry{}
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.read, &e.level); err != nil {
			rows.Close()
			return nil, err
		}
		entries = append(entries, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	// IDs are randomized within unread/read groups; load only the first usable
	// document instead of decoding and hashing the entire reading catalog.
	for _, entry := range entries {
		if !levels[strings.ToUpper(entry.level)] {
			continue
		}
		doc, exists, err := repository.ReadReadingTextDocument(ctx, q, entry.id)
		if err != nil {
			return nil, err
		}
		if !exists || !levels[strings.ToUpper(doc.Level)] || !levels[strings.ToUpper(entry.level)] {
			continue
		}
		qs, err := readingQuestions(doc)
		if err != nil {
			return nil, err
		}
		if len(qs) == 0 {
			continue
		}
		ref, _ := json.Marshal(journeyReadingRef{TextID: doc.ID})
		step := dailyjourney.Step{Mode: dailyjourney.Reading, Title: "Чтение", Goal: doc.Title, ContentRef: ref, ContentVersion: readingVersion(doc), TargetCount: min(3, len(qs)), Launch: &dailyjourney.Launch{RouteName: "ReadingText", Params: map[string]string{"textId": doc.ID}}}
		return []dailyjourney.Step{step}, nil
	}
	return nil, nil
}
func (a *DailyJourneyReadingAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	var ref journeyReadingRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return err
	}
	levels, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course)
	if err != nil {
		return err
	}
	doc, exists, err := repository.ReadReadingTextDocument(ctx, q, ref.TextID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDailyJourneyContentUnavailable
	}
	var categoryLevel string
	err = q.QueryRowContext(ctx, `SELECT level FROM reading_categories WHERE category_id=? AND text_ids::jsonb @> jsonb_build_array(CAST(? AS text))`, doc.CategoryID, doc.ID).Scan(&categoryLevel)
	if err == sql.ErrNoRows {
		return ErrDailyJourneyContentUnavailable
	}
	if err != nil {
		return err
	}
	if !levels[strings.ToUpper(categoryLevel)] {
		return ErrDailyJourneyContentUnavailable
	}
	if !exists || doc.TargetLanguage != strings.SplitN(course, "_", 2)[0] || !levels[strings.ToUpper(doc.Level)] || readingVersion(doc) != step.ContentVersion {
		return ErrDailyJourneyContentUnavailable
	}
	return nil
}
func decodeJourneyPayload(raw json.RawMessage, dest any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(dest); err != nil {
		return dailyjourney.ErrInvalidAction
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return dailyjourney.ErrInvalidAction
	}
	return nil
}
func readingAnswer(value json.RawMessage) string {
	var answer any
	if err := json.Unmarshal(value, &answer); err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(fmt.Sprint(answer)))
}
func (a *DailyJourneyReadingAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	var state journeyReadingState
	if len(step.State) > 0 && string(step.State) != "{}" {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	switch action.Type {
	case "start":
		var payload struct {
			OpenQuiz bool `json:"open_quiz"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if state.Document == nil {
			var ref journeyReadingRef
			if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
				return err
			}
			doc, exists, err := repository.ReadReadingTextDocument(ctx, tx, ref.TextID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrDailyJourneyContentUnavailable
			}
			questions, err := readingQuestions(doc)
			if err != nil {
				return err
			}
			if len(questions) < step.TargetCount {
				return ErrDailyJourneyContentUnavailable
			}
			rand.Shuffle(len(questions), func(i, k int) { questions[i], questions[k] = questions[k], questions[i] })
			state = journeyReadingState{Document: doc, Questions: questions[:step.TargetCount], Answers: map[int]string{}, Correct: map[int]bool{}}
		}
		if payload.OpenQuiz {
			state.QuizOpen = true
		}
	case "answer":
		var payload struct {
			Index  int    `json:"index"`
			Answer string `json:"answer"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if state.Document == nil || !state.QuizOpen || state.Feedback || state.Done || payload.Index != state.Index || state.Index >= len(state.Questions) {
			return dailyjourney.ErrInvalidAction
		}
		question := state.Questions[state.Index]
		answer := strings.ToLower(strings.TrimSpace(payload.Answer))
		valid := question.Type == "true_false" && (answer == "true" || answer == "false")
		for _, choice := range question.Choices {
			if answer == strings.ToLower(strings.TrimSpace(choice.ID)) {
				valid = true
			}
		}
		if !valid {
			return dailyjourney.ErrInvalidAction
		}
		state.Answers[state.Index] = answer
		state.Correct[state.Index] = answer == readingAnswer(question.CorrectAnswer)
		state.Feedback = true
		step.CompletedCount++
		feedback := "Ответ принят"
		step.Feedback = &feedback
		if step.CompletedCount == len(state.Questions) {
			correct := 0
			for _, ok := range state.Correct {
				if ok {
					correct++
				}
			}
			state.Done = true
			state.Passed = correct > len(state.Questions)/2
			step.ResultDetails, _ = json.Marshal(map[string]any{"text_id": state.Document.ID, "title": state.Document.Title, "correct": correct, "total": len(state.Questions), "passed": state.Passed})
			result := fmt.Sprintf("%s: %d из %d", state.Document.Title, correct, len(state.Questions))
			step.Result = &result
			// Preserve the ordinary >50% rule. An unsuccessful attempt still resolves
			// the stop after feedback, but never marks the text read or emits success.
			if state.Passed {
				if err := repository.NewReadingTextProgressRepository(tx).MarkRead(user, state.Document.ID); err != nil {
					return err
				}
				if _, err := repository.NewLinglowEventRepository(a.db).RecordReadingCompletedTx(ctx, tx, repository.ReadingCompletedInput{UserID: user, CourseCode: j.CourseCode, ChapterID: state.Document.ID}); err != nil {
					return err
				}
			}
		}
	case "ack_feedback":
		var empty struct{}
		if err := decodeJourneyPayload(action.Payload, &empty); err != nil {
			return err
		}
		if !state.Feedback {
			return dailyjourney.ErrInvalidAction
		}
		state.Feedback = false
		step.Feedback = nil
		if state.Done {
			step.Status = "completed"
		} else {
			state.Index++
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	var err error
	step.State, err = json.Marshal(state)
	return err
}
func (a *DailyJourneyReadingAdapter) Exercise(_ context.Context, _ int64, _ *dailyjourney.Journey, step *dailyjourney.Step) (any, error) {
	var state journeyReadingState
	if err := json.Unmarshal(step.State, &state); err != nil {
		return nil, err
	}
	if state.Document == nil {
		return nil, dailyjourney.ErrInvalidAction
	}
	// Clone before stripping answers, so no caller can mutate the pinned snapshot.
	public := JourneyReadingExercise(state)
	passage := map[string]any{}
	for k, v := range public.Document.ReadingPassage {
		if k != "comprehension_questions" {
			passage[k] = v
		}
	}
	public.Document.ReadingPassage = passage
	for i := range public.Questions {
		if _, answered := state.Answers[i]; !answered {
			public.Questions[i].CorrectAnswer = nil
			public.Questions[i].Explanation = nil
		}
	}
	return public, nil
}
