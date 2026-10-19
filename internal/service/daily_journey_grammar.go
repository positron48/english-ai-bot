package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
)

// Resolve is course-specific and must never fall back to another language.
type DailyJourneyGrammarAdapter struct {
	db        *sql.DB
	Resolve   func(string) *GrammarService
	MirrorSRS bool
}

func NewDailyJourneyGrammarAdapter(db *sql.DB, resolve func(string) *GrammarService) *DailyJourneyGrammarAdapter {
	return &DailyJourneyGrammarAdapter{db: db, Resolve: resolve}
}

type journeyGrammarRef struct {
	Kind      string `json:"kind"`
	ChapterID string `json:"chapter_id,omitempty"`
	SectionID string `json:"section_id"`
	BlockID   string `json:"block_id,omitempty"`
}
type journeyGrammarState struct {
	Phase       string                         `json:"phase,omitempty"`
	TestAttempt int                            `json:"test_attempt,omitempty"`
	Chapters    map[string]*repository.Chapter `json:"chapters,omitempty"`
	Chapter     *repository.Chapter            `json:"chapter,omitempty"`
	Questions   []map[string]interface{}       `json:"questions"`
	Answers     map[int]interface{}            `json:"answers"`
	Results     []map[string]interface{}       `json:"results"`
	Index       int                            `json:"index"`
	Feedback    bool                           `json:"feedback"`
	Result      map[string]interface{}         `json:"result,omitempty"`
}

func grammarJourneyVersion(chapters ...*repository.Chapter) string {
	raw, _ := json.Marshal(chapters)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func (a *DailyJourneyGrammarAdapter) engine(q database.ContextDBTX, course string) (*GrammarService, error) {
	if a.Resolve == nil {
		return nil, fmt.Errorf("grammar resolver is not configured")
	}
	base := a.Resolve(course)
	if base == nil || base.SRSRepo == nil || base.AttemptRepo.CourseCode() != course {
		return nil, ErrDailyJourneyContentUnavailable
	}
	conn, ok := q.(interface {
		database.DBTX
		database.ContextDBTX
	})
	if !ok {
		return nil, fmt.Errorf("grammar connection has no transaction query surface")
	}
	clone := *base
	clone.ContentRepo = base.ContentRepo.WithQueries(conn)
	clone.PublishRepo = base.PublishRepo.WithQueries(conn)
	clone.AttemptRepo = base.AttemptRepo.WithQueries(conn)
	if base.TrainingPackRepo != nil {
		clone.TrainingPackRepo = base.TrainingPackRepo.WithQueries(conn)
	}
	if base.SRSRepo != nil {
		clone.SRSRepo = base.SRSRepo.WithQueries(conn)
	}
	return &clone, nil
}
func grammarQuestionBank(c *repository.Chapter) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	bank, _ := c.QuestionBank["questions"].([]interface{})
	for _, item := range bank {
		q, ok := item.(map[string]interface{})
		if !ok || !repository.GrammarQuestionAvailable(q) {
			continue
		}
		id, _ := q["id"].(string)
		if id != "" {
			out[id] = q
		}
	}
	return out
}

// Inline questions are bound by explicit theory_block_id and quiz question IDs.
func grammarJourneyBlock(c *repository.Chapter, id string) (*repository.Chapter, []map[string]interface{}, bool) {
	bank := grammarQuestionBank(c)
	var theory interface{}
	questions := []map[string]interface{}{}
	seen := map[string]bool{}
	for _, item := range c.Blocks {
		b, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if b["id"] == id && b["type"] == "theory" {
			theory = b
		}
		quiz, _ := b["quiz_inline"].(map[string]interface{})
		ids, _ := quiz["question_ids"].([]interface{})
		for _, qid := range ids {
			key, _ := qid.(string)
			q := bank[key]
			if q != nil && q["theory_block_id"] == id && !seen[key] {
				clone := map[string]interface{}{}
				for k, v := range q {
					clone[k] = v
				}
				clone["chapter_id"] = c.ID
				questions = append(questions, clone)
				seen[key] = true
			}
		}
	}
	if theory == nil {
		return nil, nil, false
	}
	clone := *c
	clone.Blocks = []interface{}{theory}
	clone.QuestionBank = map[string]interface{}{"questions": []interface{}{}}
	clone.ChapterTest = nil
	return &clone, questions, true
}
func (a *DailyJourneyGrammarAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	if _, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course); err != nil {
		return nil, err
	}
	s, err := a.engine(q, course)
	if err != nil {
		return nil, err
	}
	sections, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, err
	}
	resume, err := s.GetContinueChapter(ctx, user)
	if err != nil {
		return nil, err
	}
	for _, section := range sections.Sections {
		pub, err := s.PublishRepo.IsPublished("section", section.SectionID)
		if err != nil {
			return nil, err
		}
		if !pub {
			continue
		}
		access, err := s.CanAccessSection(ctx, user, section.SectionID)
		if err != nil {
			return nil, err
		}
		if !access {
			continue
		}
		chapters := []*repository.Chapter{}
		allPassed := true
		for _, id := range section.ChapterIDs {
			published, err := s.PublishRepo.IsPublished("chapter", id)
			if err != nil {
				return nil, err
			}
			if !published {
				continue
			}
			c, err := s.ContentRepo.GetChapter(id)
			if err != nil {
				return nil, err
			}
			chapters = append(chapters, c)
			p, err := s.AttemptRepo.GetChapterProgress(user, id)
			if err != nil {
				return nil, err
			}
			if !p.Passed {
				allPassed = false
			}
		}
		if len(chapters) == 0 {
			continue
		}
		if allPassed {
			passed, err := s.AttemptRepo.GetCategoryTestProgress(user, section.SectionID)
			if err != nil {
				return nil, err
			}
			if !passed {
				return a.candidate(s, ctx, journeyGrammarRef{Kind: "category", SectionID: section.SectionID}, chapters)
			}
			continue
		}
		for _, c := range chapters {
			if resume == nil || c.ID != resume.ChapterID {
				continue
			}
			p, err := s.AttemptRepo.GetChapterProgress(user, c.ID)
			if err != nil {
				return nil, err
			}
			if p.Passed {
				continue
			}
			access, err := s.CanAccessChapter(ctx, user, c.ID)
			if err != nil {
				return nil, err
			}
			if !access {
				continue
			}
			return a.candidate(s, ctx, journeyGrammarRef{Kind: "chapter", ChapterID: c.ID, SectionID: c.SectionID}, []*repository.Chapter{c})
		}
	}
	return nil, nil
}
func (a *DailyJourneyGrammarAdapter) candidate(s *GrammarService, ctx context.Context, ref journeyGrammarRef, chapters []*repository.Chapter) ([]dailyjourney.Step, error) {
	count := 1
	route := "GrammarChapter"
	params := map[string]string{"chapterId": ref.ChapterID}
	title := "Грамматика"
	goal := "Изучить главу и пройти её тест"
	if ref.Kind == "block" {
		_, qs, ok := grammarJourneyBlock(chapters[0], ref.BlockID)
		if !ok {
			return nil, ErrDailyJourneyContentUnavailable
		}
		if len(qs) > 0 {
			count = len(qs)
		}
	} else {
		var test *TestQuestions
		var err error
		if ref.Kind == "category" {
			test, err = s.GenerateCategoryTest(ctx, ref.SectionID)
			route = "GrammarCategoryTest"
			params = map[string]string{"sectionId": ref.SectionID}
			goal = "Пройти полный тест категории"
		} else {
			test, err = s.GenerateChapterTest(ctx, ref.ChapterID)
			route = "GrammarChapter"
			goal = "Изучить главу и пройти её тест"
		}
		if err != nil {
			return nil, err
		}
		count = test.Total
		if count == 0 {
			return nil, nil
		}
	}
	raw, _ := json.Marshal(ref)
	return []dailyjourney.Step{{Mode: dailyjourney.Grammar, Title: title, Goal: goal, TargetCount: count, ContentRef: raw, ContentVersion: grammarJourneyVersion(chapters...), Launch: &dailyjourney.Launch{RouteName: route, Params: params}}}, nil
}
func (a *DailyJourneyGrammarAdapter) live(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) (*GrammarService, journeyGrammarRef, []*repository.Chapter, error) {
	var ref journeyGrammarRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return nil, ref, nil, err
	}
	if _, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course); err != nil {
		return nil, ref, nil, err
	}
	s, err := a.engine(q, course)
	if err != nil {
		return nil, ref, nil, err
	}
	published, err := s.PublishRepo.IsPublished("section", ref.SectionID)
	if err != nil {
		return nil, ref, nil, err
	}
	if !published {
		return nil, ref, nil, ErrDailyJourneyContentUnavailable
	}
	access, err := s.CanAccessSection(ctx, user, ref.SectionID)
	if err != nil {
		return nil, ref, nil, err
	}
	if !access {
		return nil, ref, nil, ErrDailyJourneyContentUnavailable
	}
	sec, err := s.GetSectionBySectionID(ctx, ref.SectionID)
	if err != nil {
		return nil, ref, nil, err
	}
	chapters := []*repository.Chapter{}
	for _, id := range sec.ChapterIDs {
		if ref.Kind != "category" && id != ref.ChapterID {
			continue
		}
		pub, err := s.PublishRepo.IsPublished("chapter", id)
		if err != nil {
			return nil, ref, nil, err
		}
		if !pub {
			continue
		}
		c, err := s.ContentRepo.GetChapter(id)
		if err != nil {
			return nil, ref, nil, err
		}
		chapters = append(chapters, c)
		if ref.Kind == "category" {
			p, err := s.AttemptRepo.GetChapterProgress(user, id)
			if err != nil {
				return nil, ref, nil, err
			}
			if !p.Passed {
				return nil, ref, nil, ErrDailyJourneyContentUnavailable
			}
		}
	}
	if len(chapters) == 0 || grammarJourneyVersion(chapters...) != step.ContentVersion {
		return nil, ref, nil, ErrDailyJourneyContentUnavailable
	}
	if ref.Kind != "category" {
		access, err := s.CanAccessChapter(ctx, user, ref.ChapterID)
		if err != nil {
			return nil, ref, nil, err
		}
		if !access {
			return nil, ref, nil, ErrDailyJourneyContentUnavailable
		}
	}
	return s, ref, chapters, nil
}
func (a *DailyJourneyGrammarAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	_, _, _, err := a.live(ctx, q, user, course, step)
	return err
}
func (a *DailyJourneyGrammarAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	s, ref, chapters, err := a.live(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	state := journeyGrammarState{}
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	if state.Answers == nil {
		state.Answers = map[int]interface{}{}
	}
	switch action.Type {
	case "start":
		var payload struct {
			BeginTest bool `json:"begin_test"`
			RetryTest bool `json:"retry_test"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		// Upgrade an unfinished legacy block to the full chapter on its test button.
		if ref.Kind == "block" && payload.BeginTest {
			ref.Kind = "chapter"
			ref.BlockID = ""
			step.ContentRef, _ = json.Marshal(ref)
			state = journeyGrammarState{}
			test, err := s.GenerateChapterTest(ctx, ref.ChapterID)
			if err != nil {
				return err
			}
			step.TargetCount = test.Total
			step.CompletedCount = 0
		}
		if payload.RetryTest {
			if state.Result == nil || state.Result["passed"] == true {
				return dailyjourney.ErrInvalidAction
			}
			state = journeyGrammarState{TestAttempt: state.TestAttempt + 1, Phase: "test"}
			step.CompletedCount = 0
			step.ResultDetails = nil
		}
		if state.Answers == nil {
			state.Answers = map[int]interface{}{}
		}
		if ref.Kind == "chapter" {
			state.Chapter = chapters[0]
			if payload.BeginTest || payload.RetryTest || state.Phase == "test" {
				state.Phase = "test"
				step.Launch = &dailyjourney.Launch{RouteName: "GrammarChapterTest", Params: map[string]string{"chapterId": ref.ChapterID}}
			} else {
				state.Phase = "read"
				step.Launch = &dailyjourney.Launch{RouteName: "GrammarChapter", Params: map[string]string{"chapterId": ref.ChapterID}}
			}
		}
		if state.Questions == nil {
			state.Questions = []map[string]interface{}{}
			state.Results = []map[string]interface{}{}
			if ref.Kind == "block" {
				c, qs, ok := grammarJourneyBlock(chapters[0], ref.BlockID)
				if !ok {
					return ErrDailyJourneyContentUnavailable
				}
				state.Chapter = c
				state.Questions = qs
			} else {
				var test *TestQuestions
				if ref.Kind == "category" {
					test, err = s.GenerateCategoryTest(ctx, ref.SectionID)
				} else {
					test, err = s.GenerateChapterTest(ctx, ref.ChapterID)
				}
				if err != nil {
					return err
				}
				banks := map[string]map[string]map[string]interface{}{}
				for _, c := range chapters {
					banks[c.ID] = grammarQuestionBank(c)
				}
				for _, item := range test.Questions {
					pub, ok := item.(map[string]interface{})
					if !ok {
						return fmt.Errorf("invalid grammar test question")
					}
					chapterID := ref.ChapterID
					if ref.Kind == "category" {
						chapterID, _ = pub["_category_test_chapter_id"].(string)
					}
					id, _ := pub["id"].(string)
					q := banks[chapterID][id]
					if q == nil {
						return ErrDailyJourneyContentUnavailable
					}
					clone := map[string]interface{}{}
					for k, v := range q {
						clone[k] = v
					}
					clone["chapter_id"] = chapterID
					if ref.Kind == "category" {
						clone["_category_test_chapter_id"] = chapterID
					}
					state.Questions = append(state.Questions, clone)
				}
			}
			if ref.Kind != "block" && len(state.Questions) != step.TargetCount {
				return ErrDailyJourneyContentUnavailable
			}
		}
	case "answer":
		if state.Result != nil || state.Feedback || state.Questions == nil || state.Phase == "read" {
			return dailyjourney.ErrInvalidAction
		}
		var payload struct {
			Index  int         `json:"index"`
			Answer interface{} `json:"answer"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if payload.Index != state.Index || state.Index >= len(state.Questions) || payload.Answer == nil {
			return dailyjourney.ErrInvalidAction
		}
		var lockedUser int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=? FOR UPDATE`, user).Scan(&lockedUser); err != nil {
			return err
		}
		question := state.Questions[state.Index]
		correct := s.compareQuestionAnswer(payload.Answer, question)
		if question["type"] == "true_false" {
			u, okU := normalizeTrueFalseValue(payload.Answer)
			c, okC := normalizeTrueFalseValue(question["correct_answer"])
			correct = okU && okC && u == c
		}
		resultCorrect := question["correct_answer"]
		if question["type"] == "true_false" {
			if normalized, ok := normalizeTrueFalseValue(resultCorrect); ok {
				resultCorrect = normalized
			}
		}
		result := map[string]interface{}{"question_id": question["id"], "chapter_id": question["chapter_id"], "prompt": question["prompt"], "correct": correct, "user_answer": payload.Answer, "correct_answer": resultCorrect, "explanation": question["explanation"]}
		state.Answers[state.Index] = payload.Answer
		state.Results = append(state.Results, result)
		if ref.Kind == "block" {
			chapterID, _ := question["chapter_id"].(string)
			concept, _ := question["concept_id"].(string)
			if err := s.SRSRepo.UpdateTheoryMemoryTx(tx, user, s.learning.TargetLang, s.learning.GrammarBundleID, chapterID, ref.BlockID, concept, correct, time.Now()); err != nil {
				return err
			}
			if a.MirrorSRS {
				if err := repository.NewLinglowGrammarSRSBackfillRepository(tx).SyncTheoryBlockForUser(ctx, s.learning, user, chapterID, ref.BlockID); err != nil {
					return err
				}
			}
			state.Feedback = true
		} else {
			state.Index++
			if state.Index == len(state.Questions) {
				if err := a.finishTest(ctx, tx, s, user, j, step, ref, &state); err != nil {
					return err
				}
			}
		}
		step.CompletedCount = len(state.Answers)
	case "ack_feedback":
		if ref.Kind == "block" {
			if len(state.Questions) > 0 {
				if !state.Feedback {
					return dailyjourney.ErrInvalidAction
				}
				state.Feedback = false
				state.Index++
			}
			if state.Index >= len(state.Questions) {
				if _, err := tx.ExecContext(ctx, `INSERT INTO grammar_block_progress(user_course_id,chapter_id,block_id,content_version) SELECT uc.id,?,?,? FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=? ON CONFLICT DO NOTHING`, ref.ChapterID, ref.BlockID, step.ContentVersion, user, j.CourseCode); err != nil {
					return err
				}
				step.Status = "completed"
				step.CompletedCount = step.TargetCount
				step.ResultDetails, _ = json.Marshal(map[string]interface{}{"kind": "block", "answered": len(state.Answers), "chapter_passed": false})
			}
		} else {
			if state.Result == nil || state.Result["passed"] != true {
				return dailyjourney.ErrInvalidAction
			}
			step.Status = "completed"
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	step.State, err = json.Marshal(state)
	return err
}
func (a *DailyJourneyGrammarAdapter) finishTest(ctx context.Context, tx *sql.Tx, s *GrammarService, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, ref journeyGrammarRef, state *journeyGrammarState) error {
	correct := 0
	answers := []AnswerItem{}
	for i, q := range state.Questions {
		if state.Results[i]["correct"] == true {
			correct++
		}
		answers = append(answers, AnswerItem{QuestionID: q["id"].(string), ChapterID: q["chapter_id"].(string), Answer: state.Answers[i]})
		block, _ := q["theory_block_id"].(string)
		concept, _ := q["concept_id"].(string)
		if err := s.SRSRepo.UpdateTheoryMemoryTx(tx, user, s.learning.TargetLang, s.learning.GrammarBundleID, q["chapter_id"].(string), block, concept, state.Results[i]["correct"] == true, time.Now()); err != nil {
			return err
		}
		if a.MirrorSRS {
			if err := repository.NewLinglowGrammarSRSBackfillRepository(tx).SyncTheoryBlockForUser(ctx, s.learning, user, q["chapter_id"].(string), block); err != nil {
				return err
			}
		}
	}
	total := len(state.Questions)
	score := correct * 100 / total
	passed := score >= 50
	now := time.Now()
	scopeID := ref.ChapterID
	if ref.Kind == "category" {
		scopeID = ref.SectionID
	}
	key := "daily-journey:" + j.ID + ":" + step.ID
	if state.TestAttempt > 0 {
		key += fmt.Sprintf(":retry:%d", state.TestAttempt)
	}
	answerJSON, _ := json.Marshal(answers)
	resultJSON, _ := json.Marshal(state.Results)
	attempt := &repository.TestAttempt{UserID: user, ScopeType: ref.Kind, ScopeID: scopeID, StartedAt: now, FinishedAt: &now, Score: score, Passed: passed, TotalQuestions: total, AnswersJSON: string(answerJSON), ResultsJSON: string(resultJSON), CourseVersion: &step.ContentVersion, ClientAttemptID: &key}
	id, err := s.AttemptRepo.CreateAttempt(attempt)
	if err != nil {
		return err
	}
	if ref.Kind == "chapter" {
		if err := s.AttemptRepo.UpdateProgress(user, scopeID, score, passed); err != nil {
			return err
		}
	}
	if _, err := repository.NewLinglowEventRepository(a.db).RecordGrammarTestAttemptTx(ctx, tx, s.learning, repository.GrammarTestEventInput{UserID: user, AttemptID: id, ScopeType: ref.Kind, ScopeID: scopeID, Score: score, Passed: passed, TotalQuestions: total, AnswersJSON: string(answerJSON), ResultsJSON: string(resultJSON), ClientAttemptID: key, AnsweredAt: now}); err != nil {
		return err
	}
	state.Result = map[string]interface{}{"attempt_id": id, "score": score, "passed": passed, "correct": correct, "total": total, "results": state.Results}
	step.ResultDetails, _ = json.Marshal(state.Result)
	return nil
}
func (a *DailyJourneyGrammarAdapter) Exercise(ctx context.Context, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (interface{}, error) {
	var state journeyGrammarState
	if err := json.Unmarshal(step.State, &state); err != nil {
		return nil, err
	}
	var ref journeyGrammarRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return nil, err
	}
	// Never expose hidden answer/explanation/choice feedback before grading.
	public := []map[string]interface{}{}
	for _, q := range state.Questions {
		clone := map[string]interface{}{}
		for k, v := range q {
			if k != "correct_answer" && k != "acceptable_answers" && k != "accepted_answers" && k != "explanation" && k != "choice_feedback" {
				clone[k] = v
			}
		}
		if choices, ok := q["choices"].([]interface{}); ok {
			clean := []interface{}{}
			for _, v := range choices {
				if choice, ok := v.(map[string]interface{}); ok {
					clean = append(clean, map[string]interface{}{"id": choice["id"], "text": choice["text"]})
				}
			}
			clone["choices"] = clean
		}
		if q["type"] == "reorder" {
			answer, _ := q["correct_answer"].(string)
			punct := ""
			if len(answer) > 0 && strings.ContainsAny(answer[len(answer)-1:], ".!?") {
				punct = answer[len(answer)-1:]
				answer = answer[:len(answer)-1]
			}
			words := strings.Fields(strings.ReplaceAll(answer, ",", " , "))
			sort.Strings(words)
			clone["reorder_words"] = words
			clone["reorder_punctuation"] = punct
		}
		public = append(public, clone)
	}
	chapter := state.Chapter
	if ref.Kind == "block" {
		engine := a.Resolve(j.CourseCode)
		var err error
		chapter, err = engine.ContentRepo.GetChapter(ref.ChapterID)
		if err != nil {
			return nil, err
		}
	}
	if chapter != nil {
		chapter = grammarJourneyStudyChapter(chapter)
	}
	return map[string]interface{}{"kind": ref.Kind, "phase": state.Phase, "chapter": chapter, "chapters": state.Chapters, "questions": public, "index": state.Index, "answers": state.Answers, "feedback": state.Feedback, "results": func() interface{} {
		if ref.Kind == "block" || ref.Kind == "srs" || state.Result != nil {
			return state.Results
		}
		return []interface{}{}
	}(), "result": state.Result}, nil
}

// Inline quizzes use the same local grading as the ordinary study page. Only
// their question bank is exposed; the full test queue remains private.
func grammarJourneyStudyChapter(c *repository.Chapter) *repository.Chapter {
	clone := *c
	ids := map[string]bool{}
	for _, item := range c.Blocks {
		block, _ := item.(map[string]interface{})
		quiz, _ := block["quiz_inline"].(map[string]interface{})
		questions, _ := quiz["question_ids"].([]interface{})
		for _, value := range questions {
			if id, ok := value.(string); ok {
				ids[id] = true
			}
		}
	}
	bank := []interface{}{}
	for id, question := range grammarQuestionBank(c) {
		if ids[id] {
			bank = append(bank, question)
		}
	}
	clone.QuestionBank = map[string]interface{}{"questions": bank}
	clone.ChapterTest = nil
	return &clone
}
