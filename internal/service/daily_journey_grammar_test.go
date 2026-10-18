package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

func journeyGrammarFixture(t *testing.T, db *sql.DB, lang string) *GrammarService {
	t.Helper()
	section := lang + ".grammar.journey"
	chapter := section + ".one"
	raw := fmt.Sprintf(`{"id":%q,"section_id":%q,"title":"Journey grammar","target_language":%q,"blocks":[{"id":"b1","type":"theory","theory":{"content_md":"Theory"}},{"id":"quiz","type":"quiz_inline","quiz_inline":{"question_ids":["q1","q2"]}}],"question_bank":{"questions":[{"id":"q1","type":"true_false","prompt":"First?","correct_answer":true,"explanation":"First explanation","chapter_id":%q,"theory_block_id":"b1"},{"id":"q2","type":"reorder","prompt":"Put in order","translation_ru":"Я здесь.","correct_answer":"I am here.","accepted_answers":["Here I am."],"explanation":"Second explanation","chapter_id":%q,"theory_block_id":"b1"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1","q2"],"num_questions":2}}`, chapter, section, lang, chapter, chapter)
	sections := fmt.Sprintf(`{"sections":[{"section_id":%q,"level":"A0","chapter_ids":[%q]}]}`, section, chapter)
	index := fmt.Sprintf(`{"chapters":{%q:"one.json"}}`, chapter)
	if _, err := db.Exec(`INSERT INTO grammar_content_bundle_meta(bundle_id,app_code,native_lang,target_lang,version,source_hash,sections_json,index_json) VALUES(?,?,'ru',?,'1','hash',?,?) ON CONFLICT(bundle_id) DO UPDATE SET sections_json=excluded.sections_json,index_json=excluded.index_json`, lang, lang, lang, sections, index); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO grammar_content_chapters(bundle_id,chapter_id,section_id,title,ui_language,target_language,raw_json,source_hash) VALUES(?,?,?,'Grammar','ru',?,?,'hash') ON CONFLICT(bundle_id,chapter_id) DO UPDATE SET raw_json=excluded.raw_json`, lang, chapter, section, lang, raw); err != nil {
		t.Fatal(err)
	}
	pub := repository.NewGrammarPublishRepository(db, zap.NewNop())
	for kind, id := range map[string]string{"section": section, "chapter": chapter} {
		if err := pub.SetPublished(kind, id, true, nil); err != nil {
			t.Fatal(err)
		}
	}
	lc := config.DefaultLearningConfig()
	lc.TargetLang = lang
	lc.GrammarBundleID = lang
	svc := NewGrammarService(repository.NewGrammarContentRepositoryFromDB(db, lang, zap.NewNop()), pub, repository.NewGrammarAttemptRepository(db, zap.NewNop()), lc, zap.NewNop())
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(db, zap.NewNop()))
	return svc
}
func journeyGrammarStart(t *testing.T, db *sql.DB, adapter *DailyJourneyGrammarAdapter, user int64, course, kind string) *dailyjourney.Journey {
	t.Helper()
	repo := repository.NewDailyJourneyRepository(db)
	j, err := repo.Start(context.Background(), user, course, time.Now(), func(tx *sql.Tx) ([]dailyjourney.Step, error) {
		candidates, err := adapter.Candidates(context.Background(), tx, user, course)
		if err != nil {
			return nil, err
		}
		if len(candidates) != 1 {
			return nil, fmt.Errorf("expected one candidate, got %d", len(candidates))
		}
		var ref journeyGrammarRef
		_ = json.Unmarshal(candidates[0].ContentRef, &ref)
		if ref.Kind != kind {
			return nil, fmt.Errorf("kind %s, want %s", ref.Kind, kind)
		}
		candidates[0].Status = "pending"
		return candidates, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return j
}
func TestDailyJourneyGrammarBlockAndAtomicAnswers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeySubjectUser(t, db, 808401)
	engine := journeyGrammarFixture(t, db, "en")
	adapter := NewDailyJourneyGrammarAdapter(db, func(course string) *GrammarService {
		if course == "en_ru" {
			return engine
		}
		return nil
	})
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Grammar: adapter})
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	j := journeyGrammarStart(t, db, adapter, user, "en_ru", "block")
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	view, err := adapter.Exercise(context.Background(), user, j, j.Current())
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(view)
	var public map[string]interface{}
	_ = json.Unmarshal(raw, &public)
	qs := public["questions"].([]interface{})
	if len(qs) != 2 {
		t.Fatalf("expected both bound questions: %s", raw)
	}
	if qs[0].(map[string]interface{})["correct_answer"] != nil || qs[0].(map[string]interface{})["explanation"] != nil || qs[1].(map[string]interface{})["accepted_answers"] != nil {
		t.Fatal("hidden answer leaked")
	}
	if len(qs[1].(map[string]interface{})["reorder_words"].([]interface{})) != 3 {
		t.Fatal("reorder tokens missing")
	}
	step := j.Current()
	action := dailyjourney.Action{ClientActionID: "answer", ExpectedRevision: j.Revision, Type: "answer", Payload: json.RawMessage(`{"index":0,"answer":"true"}`)}
	forced := errors.New("after SRS failure")
	_, err = svc.repo.Apply(context.Background(), user, "en_ru", j.ID, step.ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
		if err := adapter.Apply(context.Background(), tx, user, j, s, action); err != nil {
			return err
		}
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM grammar_theory_memory WHERE user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatalf("memory didn't roll back: %d %v", n, err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, e := svc.Act(ctx, user, "en_ru", j.ID, step.ID, action)
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	j, err = svc.Get(context.Background(), user, "en_ru", j.ID)
	if err != nil {
		t.Fatal(err)
	}
	var state journeyGrammarState
	_ = json.Unmarshal(j.Current().State, &state)
	if !state.Feedback || state.Results[0]["correct"] != true {
		t.Fatal("saved feedback missing")
	}
	if err := db.QueryRow(`SELECT review_count FROM grammar_theory_memory WHERE user_id=?`, user).Scan(&n); err != nil || n != 1 {
		t.Fatalf("memory duplicate: %d %v", n, err)
	}
	journeySubjectAct(t, svc, user, &j, "ack1", "ack_feedback", map[string]interface{}{})
	journeySubjectAct(t, svc, user, &j, "answer2", "answer", map[string]interface{}{"index": 1, "answer": "Here I am."})
	journeySubjectAct(t, svc, user, &j, "ack2", "ack_feedback", map[string]interface{}{})
	if j.Status != "completed" {
		t.Fatal(j.Status)
	}
	progress, err := engine.AttemptRepo.GetChapterProgress(user, "en.grammar.journey.one")
	if err != nil || progress.Passed {
		t.Fatalf("block granted chapter pass: %+v %v", progress, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM grammar_block_progress bp JOIN user_courses uc ON uc.id=bp.user_course_id WHERE uc.user_id=?`, user).Scan(&n); err != nil || n != 1 {
		t.Fatalf("missing block progress %d %v", n, err)
	}
	candidates, err := adapter.Candidates(context.Background(), db, user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	var next journeyGrammarRef
	_ = json.Unmarshal(candidates[0].ContentRef, &next)
	if next.Kind != "chapter" {
		t.Fatalf("next=%+v", next)
	}
}
func TestDailyJourneyGrammarFullTestsAndAvailability(t *testing.T) {
	db := testutil.SetupTestDB(t)
	en := journeyGrammarFixture(t, db, "en")
	es := journeyGrammarFixture(t, db, "es")
	adapter := NewDailyJourneyGrammarAdapter(db, func(course string) *GrammarService {
		if course == "en_ru" {
			return en
		}
		if course == "es_ru" {
			return es
		}
		return nil
	})
	for i, kind := range []string{"chapter", "category"} {
		t.Run(kind, func(t *testing.T) {
			user := journeySubjectUser(t, db, int64(808410+i))
			chapter := "en.grammar.journey.one"
			if kind == "chapter" {
				c, err := en.ContentRepo.GetChapter(chapter)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO grammar_block_progress(user_course_id,chapter_id,block_id,content_version) SELECT uc.id,?,'b1',? FROM user_courses uc JOIN courses co ON co.id=uc.course_id WHERE uc.user_id=? AND co.code='en_ru'`, chapter, grammarJourneyVersion(c), user); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := en.AttemptRepo.UpdateProgress(user, chapter, 50, true); err != nil {
					t.Fatal(err)
				}
			}
			svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Grammar: adapter})
			j := journeyGrammarStart(t, db, adapter, user, "en_ru", kind)
			journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
			var state journeyGrammarState
			_ = json.Unmarshal(j.Current().State, &state)
			if len(state.Questions) != 2 {
				t.Fatal("truncated full test")
			}
			for index, q := range state.Questions {
				if index == len(state.Questions)-1 {
					snapshot := j
					current := j.Current()
					payload, _ := json.Marshal(map[string]interface{}{"index": index, "answer": "wrong"})
					action := dailyjourney.Action{ClientActionID: "rollback-final", ExpectedRevision: j.Revision, Type: "answer", Payload: payload}
					forced := errors.New("rollback full test")
					_, err := svc.repo.Apply(context.Background(), user, "en_ru", j.ID, current.ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
						if err := adapter.Apply(context.Background(), tx, user, j, step, action); err != nil {
							return err
						}
						return forced
					})
					if !errors.Is(err, forced) {
						t.Fatal(err)
					}
					var attempts int
					if err := db.QueryRow(`SELECT COUNT(*) FROM grammar_test_attempts WHERE user_id=?`, user).Scan(&attempts); err != nil || attempts != 0 {
						t.Fatalf("test failed to roll back: %d %v", attempts, err)
					}
					j = snapshot
				}

				answer := q["correct_answer"]
				if index == 1 {
					answer = "wrong"
				}
				journeySubjectAct(t, svc, user, &j, fmt.Sprint("answer", index), "answer", map[string]interface{}{"index": index, "answer": answer})
			}
			_ = json.Unmarshal(j.Current().State, &state)
			if state.Result["score"] != float64(50) || state.Result["passed"] != true {
				t.Fatalf("threshold: %+v", state.Result)
			}
			if j.Status != "active" {
				t.Fatal("result skipped acknowledgment")
			}
			var n int
			for _, table := range []string{"grammar_test_attempts", "exercise_attempts", "learning_events", "mode_daily_stats"} {
				query := "SELECT COUNT(*) FROM " + table + " WHERE user_course_id IN (SELECT id FROM user_courses WHERE user_id=?)"
				if table == "grammar_test_attempts" {
					query = "SELECT COUNT(*) FROM grammar_test_attempts WHERE user_id=?"
				}
				if err := db.QueryRow(query, user).Scan(&n); err != nil || n != 1 {
					t.Fatalf("%s %d %v", table, n, err)
				}
			}
			journeySubjectAct(t, svc, user, &j, "ack", "ack_feedback", map[string]interface{}{})
			if j.Status != "completed" {
				t.Fatal(j.Status)
			}
			if kind == "chapter" {
				p, err := en.AttemptRepo.GetChapterProgress(user, chapter)
				if err != nil || !p.Passed {
					t.Fatalf("50%% must pass: %+v %v", p, err)
				}
			} else {
				passed, err := en.AttemptRepo.GetCategoryTestProgress(user, "en.grammar.journey")
				if err != nil || !passed {
					t.Fatal(passed, err)
				}
			}
		})
	}
	user := journeySubjectUser(t, db, 808420)
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(context.Background(), user, "es_ru"); err != nil {
		t.Fatal(err)
	}
	candidates, err := adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 1 {
		t.Fatalf("ES %v %v", candidates, err)
	}
	var ref journeyGrammarRef
	_ = json.Unmarshal(candidates[0].ContentRef, &ref)
	if ref.ChapterID != "es.grammar.journey.one" {
		t.Fatal("cross-course candidate")
	}
	if err := en.PublishRepo.SetPublished("chapter", "en.grammar.journey.one", false, nil); err != nil {
		t.Fatal(err)
	}
	// EN withdrawal must not affect the distinct ES content.
	if err := adapter.Validate(context.Background(), db, user, "es_ru", &candidates[0]); err != nil {
		t.Fatal(err)
	}
	if err := es.PublishRepo.SetPublished("chapter", ref.ChapterID, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Validate(context.Background(), db, user, "es_ru", &candidates[0]); !errors.Is(err, ErrDailyJourneyContentUnavailable) {
		t.Fatalf("withdrawn: %v", err)
	}
}

func TestDailyJourneyGrammarSRSQueueAndAtomicity(t *testing.T) {
	db := testutil.SetupTestDB(t)
	engine := journeyGrammarFixture(t, db, "en")
	user := journeySubjectUser(t, db, 808430)
	chapter := "en.grammar.journey.one"
	if _, err := db.Exec(`UPDATE grammar_content_chapters SET level='A0' WHERE bundle_id='en' AND chapter_id=?`, chapter); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 6; i++ {
		block := fmt.Sprint("srs-b", i)
		id := fmt.Sprint("journey-srs-q", i)
		raw, _ := json.Marshal(map[string]interface{}{"id": id, "chapter_id": chapter, "theory_block_id": block, "type": "mcq_single", "prompt": "Choose A", "choices": []interface{}{map[string]interface{}{"id": "a", "text": "A", "feedback": "Answer leak"}, map[string]interface{}{"id": "b", "text": "B"}}, "correct_answer": "a", "explanation": "Saved SRS explanation"})
		if _, err := db.Exec(`INSERT INTO grammar_training_content_questions(bundle_id,question_id,chapter_id,theory_block_id,raw_json,source_hash) VALUES('en',?,?,?,?,'hash')`, id, chapter, block, string(raw)); err != nil {
			t.Fatal(err)
		}
	}
	engine.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryFromDB(db, "en", zap.NewNop()))
	adapter := NewDailyJourneyGrammarSRSAdapter(db, func(course string) *GrammarService {
		if course == "en_ru" {
			return engine
		}
		return nil
	})
	adapter.MirrorSRS = true
	// Without a passed chapter/placement there is no candidate and no writes.
	candidates, err := adapter.Candidates(context.Background(), db, user, "en_ru")
	if err != nil || len(candidates) != 0 {
		t.Fatalf("locked SRS %v %v", candidates, err)
	}
	if err := engine.AttemptRepo.UpdateProgress(user, chapter, 100, true); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).MapLegacyContent(context.Background(), "en_ru", "en"); err != nil {
		t.Fatal(err)
	}
	if err := engine.SRSRepo.EnsureTheoryMemory(user, "en", "en", chapter, "srs-b6", ""); err != nil {
		t.Fatal(err)
	}
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	candidates, err = adapter.Candidates(context.Background(), db, user, "en_ru")
	if err != nil || len(candidates) != 1 || candidates[0].TargetCount != 5 {
		t.Fatalf("queue candidate %v %v", candidates, err)
	}
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.GrammarSRS: adapter})
	j, err := svc.Start(context.Background(), user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	var state journeyGrammarState
	_ = json.Unmarshal(j.Current().State, &state)
	if len(state.Questions) != 5 || state.Questions[0]["theory_block_id"] != "srs-b6" {
		t.Fatalf("due priority %+v", state.Questions)
	}
	step := j.Current()
	action := dailyjourney.Action{ClientActionID: "answer0", ExpectedRevision: j.Revision, Type: "answer", Payload: json.RawMessage(`{"index":0,"answer":"b"}`)}
	forced := errors.New("rollback SRS result")
	_, err = svc.repo.Apply(context.Background(), user, "en_ru", j.ID, step.ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
		if err := adapter.Apply(context.Background(), tx, user, j, step, action); err != nil {
			return err
		}
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM grammar_attempts WHERE user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatalf("attempt rollback %d %v", n, err)
	}
	if err := db.QueryRow(`SELECT review_count FROM grammar_theory_memory WHERE user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatalf("memory rollback %d %v", n, err)
	}
	for i := 0; i < 5; i++ {
		journeySubjectAct(t, svc, user, &j, fmt.Sprint("answer", i), "answer", map[string]interface{}{"index": i, "answer": "b"})
		view, err := adapter.Exercise(context.Background(), user, j, j.Current())
		if err != nil {
			t.Fatal(err)
		}
		stateRaw, _ := json.Marshal(view)
		var public map[string]interface{}
		_ = json.Unmarshal(stateRaw, &public)
		if public["feedback"] != true {
			t.Fatal("feedback not restored")
		}
		if public["questions"].([]interface{})[i].(map[string]interface{})["correct_answer"] != nil {
			t.Fatal("hidden answer leaked")
		}
		journeySubjectAct(t, svc, user, &j, fmt.Sprint("ack", i), "ack_feedback", map[string]interface{}{})
	}
	if j.Status != "completed" {
		t.Fatal(j.Status)
	}
	for _, table := range []string{"grammar_attempts", "exercise_attempts", "learning_events", "srs_items"} {
		query := "SELECT COUNT(*) FROM " + table + " WHERE user_course_id IN (SELECT id FROM user_courses WHERE user_id=?)"
		if table == "grammar_attempts" {
			query = "SELECT COUNT(*) FROM grammar_attempts WHERE user_id=?"
		}
		if err := db.QueryRow(query, user).Scan(&n); err != nil || n != 5 {
			t.Fatalf("%s %d %v", table, n, err)
		}
	}
	if err := db.QueryRow(`SELECT attempt_count FROM mode_daily_stats WHERE user_course_id IN (SELECT id FROM user_courses WHERE user_id=?) AND mode='grammar_training'`, user).Scan(&n); err != nil || n != 5 {
		t.Fatalf("stats %d %v", n, err)
	}
}

func TestDailyJourneyGrammarRespectsPlacementContinueAndWithdrawal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	s := journeyGrammarFixture(t, db, "en")
	user := journeySubjectUser(t, db, 808440)
	section2 := "en.grammar.journey.advanced"
	chapter2 := section2 + ".one"
	old, err := s.ContentRepo.GetChapter("en.grammar.journey.one")
	if err != nil {
		t.Fatal(err)
	}
	old.ID = chapter2
	old.SectionID = section2
	raw, _ := json.Marshal(old)
	if _, err := db.Exec(`INSERT INTO grammar_content_chapters(bundle_id,chapter_id,section_id,title,ui_language,target_language,raw_json,source_hash) VALUES('en',?,?,'Advanced','ru','en',?,'hash')`, chapter2, section2, string(raw)); err != nil {
		t.Fatal(err)
	}
	sections := fmt.Sprintf(`{"sections":[{"section_id":"en.grammar.journey","level":"A0","chapter_ids":["en.grammar.journey.one"]},{"section_id":%q,"level":"B1","chapter_ids":[%q]}]}`, section2, chapter2)
	if _, err := db.Exec(`UPDATE grammar_content_bundle_meta SET sections_json=? WHERE bundle_id='en'`, sections); err != nil {
		t.Fatal(err)
	}
	for kind, id := range map[string]string{"section": section2, "chapter": chapter2} {
		if err := s.PublishRepo.SetPublished(kind, id, true, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AttemptRepo.SavePlacementTestResult(user, 90, 10, []string{section2}); err != nil {
		t.Fatal(err)
	}
	adapter := NewDailyJourneyGrammarAdapter(db, func(course string) *GrammarService { return s })
	candidates, err := adapter.Candidates(context.Background(), db, user, "en_ru")
	if err != nil || len(candidates) != 1 {
		t.Fatal(candidates, err)
	}
	var ref journeyGrammarRef
	_ = json.Unmarshal(candidates[0].ContentRef, &ref)
	if ref.ChapterID != chapter2 {
		t.Fatalf("placement ignored: %+v", ref)
	}
	if err := s.AttemptRepo.DeletePlacementTestResult(user); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Validate(context.Background(), db, user, "en_ru", &candidates[0]); !errors.Is(err, ErrDailyJourneyContentUnavailable) {
		t.Fatalf("revoked placement access: %v", err)
	}
}
