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
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

func journeySubjectUser(t *testing.T, db *sql.DB, telegram int64) int64 {
	t.Helper()
	user, err := repository.NewUserRepository(db, zap.NewNop()).GetOrCreateUser(telegram)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(context.Background(), user.ID, "en_ru"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET settings_json='{"spell_mode_enabled":false,"type_mode_enabled":false,"options_delay_seconds":0}' WHERE id=?`, user.ID); err != nil {
		t.Fatal(err)
	}
	return user.ID
}
func journeySubjectWord(t *testing.T, db *sql.DB, index int) (int64, int64) {
	t.Helper()
	var id int64
	if err := db.QueryRow(`INSERT INTO word_cards(word,definition,course_code) VALUES(?,?,'en_ru') RETURNING id`, fmt.Sprintf("journey-word-%d", index), "{} ").Scan(&id); err != nil {
		t.Fatal(err)
	}
	tc := &models.TrainingCard{WordCardID: id, WordEN: fmt.Sprintf("word%d", index), WordRU: fmt.Sprintf("слово%d", index), SenseIndex: 0, DistractorsEN: `["other","something","another"]`, DistractorsRU: `["другое","что-то","иное"]`}
	cardID, err := repository.NewTrainingCardRepository(db, zap.NewNop()).CreateTrainingCard(tc)
	if err != nil {
		t.Fatal(err)
	}
	return id, cardID
}
func journeySubjectAct(t *testing.T, s *DailyJourneyService, user int64, j **dailyjourney.Journey, key, kind string, payload any) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	current := (*j).Current()
	if current == nil {
		t.Fatal("no current stop")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := s.Act(ctx, user, "en_ru", (*j).ID, current.ID, dailyjourney.Action{ClientActionID: key, Type: kind, ExpectedRevision: (*j).Revision, Payload: raw})
	if err != nil {
		t.Fatalf("%s %s: %v", key, kind, err)
	}
	*j = result
}
func TestDailyJourneyWordsRealSRSAtomicity(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	user := journeySubjectUser(t, db, 808301)
	if _, err := db.Exec(`UPDATE users SET settings_json='{"spell_mode_enabled":false,"type_mode_enabled":false,"options_delay_seconds":0,"wrong_answer_delay_seconds":7}' WHERE id=?`, user); err != nil {
		t.Fatal(err)
	}
	userCards := repository.NewUserCardRepository(db, zap.NewNop())
	for i := 0; i < 20; i++ {
		_, tc := journeySubjectWord(t, db, i)
		for _, direction := range []models.CardDirection{models.DirectionENtoRU, models.DirectionRUtoEN} {
			if _, err := userCards.CreateUserCard(&models.UserCard{UserID: user, TrainingCardID: tc, Direction: direction, State: models.StateReview, EF: 2.5, Reps: 3}); err != nil {
				t.Fatal(err)
			}
		}
	}
	normal := repository.NewSessionRepository(db, zap.NewNop())
	normalID, err := normal.CreateSession(&models.TrainingSession{UserID: user, Source: models.SourceManual, PlannedCount: 40, SessionJSON: `{"independent":true}`})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).MapLegacyContentForLearning(ctx, config.DefaultLearningConfig()); err != nil {
		t.Fatal(err)
	}
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	adapter := NewDailyJourneyWordsAdapter(db, zap.NewNop(), config.DefaultLearningConfig())
	adapter.MirrorSRS = true
	adapter.WrongAnswerDelaySeconds = 3
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	j, err := svc.Start(ctx, user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]any{})
	var state journeyWordsState
	if err := json.Unmarshal(j.Steps[0].State, &state); err != nil {
		t.Fatal(err)
	}
	if state.WrongAnswerDelaySeconds != 7 {
		t.Fatalf("user wrong-answer delay ignored: %d", state.WrongAnswerDelaySeconds)
	}
	if state.DueWords != 20 || state.AddTarget != 0 || len(state.Queue) != 20 {
		t.Fatalf("directional cards counted as words: %+v", state)
	}
	journeySubjectAct(t, svc, user, &j, "reveal", "start", map[string]any{"reveal": true})
	step := j.Current()
	if err := json.Unmarshal(step.State, &state); err != nil {
		t.Fatal(err)
	}
	wrong := 0
	for index, option := range state.Options {
		if option != state.CorrectAnswer {
			wrong = index
			break
		}
	}
	payload, _ := json.Marshal(map[string]any{"card_index": 1, "user_card_id": state.Card["user_card_id"], "option_index": wrong})
	action := dailyjourney.Action{ClientActionID: "grade", ExpectedRevision: j.Revision, Type: "answer", Payload: payload}
	deliberate := errors.New("abort after real SRS/event/stats writes")
	_, err = repository.NewDailyJourneyRepository(db).Apply(ctx, user, "en_ru", j.ID, step.ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
		if err := adapter.Apply(ctx, tx, user, j, step, action); err != nil {
			return err
		}
		return deliberate
	})
	if !errors.Is(err, deliberate) {
		t.Fatal(err)
	}
	count := func(table string) int {
		t.Helper()
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if count("review_events") != 0 || count("exercise_attempts") != 0 || count("learning_events") != 0 || count("mode_daily_stats") != 0 || count("user_word_mastering") != 0 || count("srs_items") != 0 {
		t.Fatal("subject side effects escaped rollback")
	}
	// Eight identical concurrent submissions: one SRS change and one set of events.
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Act(ctx, user, "en_ru", j.ID, step.ID, action)
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if count("review_events") != 1 || count("exercise_attempts") != 1 || count("learning_events") != 1 || count("user_word_mastering") != 1 || count("srs_items") != 1 {
		t.Fatal("duplicate subject credit")
	}
	var correct int
	if err := db.QueryRow(`SELECT is_correct FROM review_events LIMIT 1`).Scan(&correct); err != nil {
		t.Fatal(err)
	}
	if correct != 0 {
		t.Fatal("wrong answer turned correct")
	}
	active, err := normal.GetActiveSession(user)
	if err != nil {
		t.Fatal(err)
	}
	if active == nil || active.ID != normalID || active.SessionJSON != `{"independent":true}` {
		t.Fatal("daily journey changed normal session")
	}
	restored := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	view, err := restored.Exercise(ctx, user, "en_ru", j.ID, step.ID)
	if err != nil {
		t.Fatal(err)
	}
	data := view.Data.(JourneyWordsExercise)
	if data.Feedback["is_correct"] != false || data.Card["card_index"].(float64) != 1 {
		t.Fatal("feedback not restored")
	}
	var cardID int64
	raw, _ := json.Marshal(state.Card["user_card_id"])
	if err := json.Unmarshal(raw, &cardID); err != nil {
		t.Fatal(err)
	}
	graded, err := userCards.GetUserCard(cardID)
	if err != nil {
		t.Fatal(err)
	}
	if graded.LapseCount != 1 || graded.LastQuality == nil || *graded.LastQuality != 0 || graded.WrongAnswersJSON == "" {
		t.Fatalf("bad SRS grade: %+v", graded)
	}
	// Finish the real queue without touching the independent normal session.
	j = view.Journey
	journeySubjectAct(t, svc, user, &j, "ack-first", "ack_feedback", map[string]any{})
	for index := 1; index < 20; index++ {
		journeySubjectAct(t, svc, user, &j, fmt.Sprintf("reveal-%d", index), "start", map[string]any{"reveal": true})
		if err := json.Unmarshal(j.Current().State, &state); err != nil {
			t.Fatal(err)
		}
		option := -1
		for k, value := range state.Options {
			if value == state.CorrectAnswer {
				option = k
				break
			}
		}
		if option < 0 {
			t.Fatal("correct option missing")
		}
		journeySubjectAct(t, svc, user, &j, fmt.Sprintf("answer-%d", index), "answer", map[string]any{"card_index": index + 1, "user_card_id": state.Card["user_card_id"], "option_index": option})
		journeySubjectAct(t, svc, user, &j, fmt.Sprintf("ack-%d", index), "ack_feedback", map[string]any{})
	}
	if j.Status != "completed" || j.Steps[0].CompletedCount != 20 || count("review_events") != 20 {
		t.Fatal("queue completion is inconsistent")
	}
	active, err = normal.GetActiveSession(user)
	if err != nil {
		t.Fatal(err)
	}
	if active == nil || active.ID != normalID || active.DoneCount != 0 {
		t.Fatal("daily finish ended/advanced normal training")
	}

}
func TestDailyJourneyWordsAdditionUsesUniqueReadyWords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	user := journeySubjectUser(t, db, 808302)
	setRepo := repository.NewWordSetRepository(db, zap.NewNop())
	level := "A0"
	set, err := setRepo.CreateWordSet(&models.WordSet{CourseCode: "en_ru", Title: "Daily ready", IsPublished: true, LevelCode: &level})
	if err != nil {
		t.Fatal(err)
	}
	ids := []int64{}
	for i := 0; i < 20; i++ {
		word, _ := journeySubjectWord(t, db, i)
		ids = append(ids, word)
	}
	if err := setRepo.SetWordSetItems(set, ids); err != nil {
		t.Fatal(err)
	}
	adapter := NewDailyJourneyWordsAdapter(db, zap.NewNop(), config.DefaultLearningConfig())
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	j, err := svc.Start(ctx, user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]any{})
	var state journeyWordsState
	if err := json.Unmarshal(j.Steps[0].State, &state); err != nil {
		t.Fatal(err)
	}
	if state.AddTarget != 15 || state.DueWords != 0 || j.Steps[0].Launch.RouteName != "WordSets" {
		t.Fatal("missing addition phase")
	}
	journeySubjectAct(t, svc, user, &j, "know", "add_word", map[string]any{"set_id": set, "word_card_id": ids[0], "known": true})
	journeySubjectAct(t, svc, user, &j, "learn-1", "add_word", map[string]any{"set_id": set, "word_card_id": ids[1]})
	journeySubjectAct(t, svc, user, &j, "learn-duplicate", "add_word", map[string]any{"set_id": set, "word_card_id": ids[1]})
	if err := json.Unmarshal(j.Steps[0].State, &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Added) != 1 || j.Steps[0].CompletedCount != 1 {
		t.Fatal("know/repeated learn counted")
	}
	for i := 2; i <= 15; i++ {
		journeySubjectAct(t, svc, user, &j, fmt.Sprintf("learn-%d", i), "add_word", map[string]any{"set_id": set, "word_card_id": ids[i]})
	}
	if err := json.Unmarshal(j.Steps[0].State, &state); err != nil {
		t.Fatal(err)
	}
	if state.Phase != "train" || len(state.Queue) != 15 || j.Steps[0].Launch.RouteName != "Training" {
		t.Fatal("new words excluded by ordinary new-card limit")
	}
	seen := map[int64]bool{}
	for _, item := range state.Queue {
		if item.Card == nil {
			t.Fatal("unexpected format")
		}
		seen[item.Card.TrainingCard.WordCardID] = true
	}
	for _, id := range state.Added {
		if !seen[id] {
			t.Fatal("added word missing from queue")
		}
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM training_sessions WHERE user_id=?`, user).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("global session created")
	}
}
func TestDailyJourneyReadingAttemptsAndAtomicMark(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	catalog := repository.NewReadingCatalogRepository(db)
	passage := `{"title":"A small story","segments":[],"comprehension_questions":[{"id":"q1","prompt":"One?","type":"true_false","correct_answer":"true","explanation":"First explanation"},{"id":"q2","prompt":"Two?","type":"true_false","correct_answer":"true"},{"id":"q3","prompt":"Three?","type":"true_false","correct_answer":"true"}]}`
	if err := catalog.ReplaceCatalog("v1", "", []repository.ReadingCategoryUpsert{{CategoryID: "daily-a0", Title: "A0", Level: "A0", TextIDs: []string{"daily-text"}}}, []repository.ReadingTextUpsert{{TextID: "daily-text", CategoryID: "daily-a0", Title: "A small story", Level: "A0", TargetLanguage: "en", ReadingPassageJSON: passage}}); err != nil {
		t.Fatal(err)
	}
	for _, passed := range []bool{false, true} {
		t.Run(fmt.Sprintf("passed=%v", passed), func(t *testing.T) {
			user := journeySubjectUser(t, db, 808303+int64(map[bool]int{false: 0, true: 1}[passed]))
			adapter := NewDailyJourneyReadingAdapter(db)
			svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Reading: adapter})
			j, err := svc.Start(ctx, user, "en_ru")
			if err != nil {
				t.Fatal(err)
			}
			if j.Current() == nil || j.Current().Mode != dailyjourney.Reading {
				t.Fatalf("no accessible reading: %+v", j)
			}
			journeySubjectAct(t, svc, user, &j, "start", "start", map[string]any{"open_quiz": true})
			view, err := svc.Exercise(ctx, user, "en_ru", j.ID, j.Current().ID)
			if err != nil {
				t.Fatal(err)
			}
			public := view.Data.(JourneyReadingExercise)
			if public.Questions[0].CorrectAnswer != nil || public.Document.ReadingPassage["comprehension_questions"] != nil {
				t.Fatal("unanswered answer key exposed")
			}
			for index := 0; index < 3; index++ {
				answer := "false"
				if passed {
					answer = "true"
				}
				if passed && index == 2 {
					raw, _ := json.Marshal(map[string]any{"index": index, "answer": answer})
					action := dailyjourney.Action{ClientActionID: "atomic-last", ExpectedRevision: j.Revision, Type: "answer", Payload: raw}
					deliberate := errors.New("rollback reading mark/event")
					_, err := repository.NewDailyJourneyRepository(db).Apply(ctx, user, "en_ru", j.ID, j.Current().ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
						if err := adapter.Apply(ctx, tx, user, j, step, action); err != nil {
							return err
						}
						return deliberate
					})
					if !errors.Is(err, deliberate) {
						t.Fatal(err)
					}
					read, err := repository.NewReadingTextProgressRepository(db).Get(user, "daily-text")
					if err != nil {
						t.Fatal(err)
					}
					if read != nil {
						t.Fatal("read mark escaped rollback")
					}
					var count int
					if err := db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts ea JOIN user_courses uc ON uc.id=ea.user_course_id WHERE uc.user_id=?`, user).Scan(&count); err != nil {
						t.Fatal(err)
					}
					if count != 0 {
						t.Fatal("reading event escaped rollback")
					}
				}
				journeySubjectAct(t, svc, user, &j, fmt.Sprintf("answer-%d", index), "answer", map[string]any{"index": index, "answer": answer})
				if j.Current() == nil || j.Current().Status != "active" {
					t.Fatal("stop finished before feedback acknowledged")
				}
				journeySubjectAct(t, svc, user, &j, fmt.Sprintf("ack-%d", index), "ack_feedback", map[string]any{})
			}
			if j.Status != "completed" || j.Steps[1].Status != "completed" {
				t.Fatal("finished unsuccessful attempt did not resolve stop")
			}
			read, err := repository.NewReadingTextProgressRepository(db).Get(user, "daily-text")
			if err != nil {
				t.Fatal(err)
			}
			if (read != nil) != passed {
				t.Fatal("ordinary >50% reading threshold changed")
			}
		})
	}
}
