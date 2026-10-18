package service

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

// Opt-in timing probe against a disposable DB and a large synthetic pack.
func TestDailyJourneySRSPinnedValidationCost(t *testing.T) {
	if os.Getenv("DAILY_JOURNEY_SRS_PERF") != "1" {
		t.Skip("opt-in synthetic performance probe")
	}
	db := testutil.SetupTestDB(t)
	engine := journeyGrammarFixture(t, db, "en")
	user := journeySubjectUser(t, db, 808491)
	chapter := "en.grammar.journey.one"
	if _, err := db.Exec(`UPDATE grammar_content_chapters SET level='A0' WHERE bundle_id='en' AND chapter_id=?`, chapter); err != nil {
		t.Fatal(err)
	}
	if err := engine.AttemptRepo.UpdateProgress(user, chapter, 100, true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO grammar_training_content_questions(bundle_id,question_id,chapter_id,theory_block_id,raw_json,source_hash) SELECT 'en','perf-q-'||g,?,'perf-b-'||g,json_build_object('id','perf-q-'||g,'chapter_id',CAST(? AS text),'theory_block_id','perf-b-'||g,'type','true_false','prompt',repeat('Synthetic question. ',100),'correct_answer',true)::text,'perf' FROM generate_series(1,2000) g`, chapter, chapter); err != nil {
		t.Fatal(err)
	}
	engine.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryFromDB(db, "en", zap.NewNop()))
	adapter := NewDailyJourneyGrammarSRSAdapter(db, func(string) *GrammarService { return engine })
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.GrammarSRS: adapter})
	j, err := svc.Start(context.Background(), user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	started := time.Now()
	for i := 0; i < 10; i++ {
		if err := adapter.Validate(context.Background(), db, user, "en_ru", j.Current()); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("2000-question pack: selected=%d average validation=%s", j.Current().TargetCount, time.Since(started)/10)
}
