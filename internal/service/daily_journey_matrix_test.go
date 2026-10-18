package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

func TestDailyJourneyCourseTierMatrix(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	engines := map[string]*GrammarService{}
	texts := []repository.ReadingTextUpsert{}
	for _, lang := range []string{"en", "es"} {
		engine := journeyGrammarFixture(t, db, lang)
		engines[lang+"_ru"] = engine
		chapter := lang + ".grammar.journey.one"
		if _, err := db.Exec(`UPDATE grammar_content_chapters SET level='A0' WHERE bundle_id=? AND chapter_id=?`, lang, chapter); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(map[string]any{"id": lang + "-matrix-q", "chapter_id": chapter, "theory_block_id": "b1", "type": "true_false", "prompt": "Ready?", "correct_answer": true})
		if _, err := db.Exec(`INSERT INTO grammar_training_content_questions(bundle_id,question_id,chapter_id,theory_block_id,raw_json,source_hash) VALUES(?,?,?,'b1',?,'matrix')`, lang, lang+"-matrix-q", chapter, string(raw)); err != nil {
			t.Fatal(err)
		}
		engine.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryFromDB(db, lang, logger))
		texts = append(texts, repository.ReadingTextUpsert{TextID: lang + "-matrix-text", CategoryID: "matrix-a0", Title: "Story", Level: "A0", TargetLanguage: lang, ReadingPassageJSON: `{"title":"Story","segments":[],"comprehension_questions":[{"id":"q","prompt":"Ready?","type":"true_false","correct_answer":"true"}]}`})
	}
	if err := repository.NewReadingCatalogRepository(db).ReplaceCatalog("matrix", "", []repository.ReadingCategoryUpsert{{CategoryID: "matrix-a0", Title: "A0", Level: "A0", TextIDs: []string{"en-matrix-text", "es-matrix-text"}}}, texts); err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"en", "es"} {
		if _, err := repository.NewCourseRepository(db, logger).MapLegacyContent(ctx, lang+"_ru", lang); err != nil {
			t.Fatal(err)
		}
	}
	resolve := func(course string) *GrammarService { return engines[course] }
	adapters := map[dailyjourney.Mode]DailyJourneyAdapter{
		dailyjourney.Words:   NewDailyJourneyWordsAdapter(db, logger, config.DefaultLearningConfig()),
		dailyjourney.Reading: NewDailyJourneyReadingAdapter(db), dailyjourney.Grammar: NewDailyJourneyGrammarAdapter(db, resolve), dailyjourney.GrammarSRS: NewDailyJourneyGrammarSRSAdapter(db, resolve),
		dailyjourney.VerbForms: NewDailyJourneyVerbsAdapter(db, config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 10}, config.DefaultLearningConfig(), logger),
	}
	for ci, course := range []string{"en_ru", "es_ru"} {
		for ti, tier := range []string{"free", "pro", "pro_plus"} {
			t.Run(course+"/"+tier, func(t *testing.T) {
				idx := ci*3 + ti
				user := journeySubjectUser(t, db, int64(808800+idx))
				if _, err := repository.NewCourseRepository(db, logger).EnsureUserCourse(ctx, user, course); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE users SET subscription_tier=? WHERE id=?`, tier, user); err != nil {
					t.Fatal(err)
				}
				word, tc := journeySubjectWord(t, db, 20000+idx)
				if _, err := db.Exec(`UPDATE word_cards SET course_code=? WHERE id=?`, course, word); err != nil {
					t.Fatal(err)
				}
				if _, err := repository.NewUserCardRepository(db, logger).CreateUserCard(&models.UserCard{UserID: user, TrainingCardID: tc, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.5, Reps: 3}); err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(course, "es_") {
					journeyVerbSeed(t, db, user, 21000+idx, "presente")
				}
				for mode, adapter := range adapters {
					if mode == dailyjourney.GrammarSRS {
						continue
					}
					items, err := adapter.Candidates(ctx, db, user, course)
					if err != nil {
						t.Fatalf("%s: %v", mode, err)
					}
					want := mode != dailyjourney.VerbForms || course == "es_ru"
					if (len(items) > 0) != want {
						t.Fatalf("%s available=%v want=%v (%s)", mode, len(items) > 0, want, fmt.Sprint(items))
					}
					for _, item := range items {
						if err := adapter.Validate(ctx, db, user, course, &item); err != nil {
							t.Fatalf("%s selected invalid content: %v", mode, err)
						}
					}
				}
				srs := adapters[dailyjourney.GrammarSRS]
				items, err := srs.Candidates(ctx, db, user, course)
				if err != nil || len(items) != 0 {
					t.Fatal("unlearned SRS allowed", items, err)
				}
				lang := strings.Split(course, "_")[0]
				if err := engines[course].AttemptRepo.UpdateProgress(user, lang+".grammar.journey.one", 100, true); err != nil {
					t.Fatal(err)
				}
				items, err = srs.Candidates(ctx, db, user, course)
				if err != nil || len(items) == 0 {
					t.Fatal("passed chapter SRS unavailable", items, err)
				}
				for _, item := range items {
					if err := srs.Validate(ctx, db, user, course, &item); err != nil {
						t.Fatal(err)
					}
				}

			})
		}
	}
}
