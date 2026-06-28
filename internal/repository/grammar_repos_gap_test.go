package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const grammarGapUserBase int64 = 900001

const grammarGapInvalidDSN = "postgres://x:x@127.0.0.1:1/nodb?connect_timeout=1"

func grammarGapBrokenDB(t *testing.T) *sql.DB {
	t.Helper()
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", grammarGapInvalidDSN)
	if err != nil {
		t.Skip("postgres_compat open:", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func grammarGapDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
}

func grammarGapUser(t *testing.T, conn *sql.DB, telegramID int64) int64 {
	t.Helper()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user %d: %v", telegramID, err)
	}
	return user.ID
}

func grammarGapAttemptRepo(t *testing.T, conn *sql.DB) *GrammarAttemptRepository {
	t.Helper()
	return NewGrammarAttemptRepository(conn, zap.NewNop())
}

func grammarGapSeedContentBundle(t *testing.T, conn *sql.DB, bundleID string) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO grammar_content_bundle_meta (bundle_id, app_code, native_lang, target_lang, version, source_hash, sections_json, index_json)
		VALUES (?, 'english', 'ru', 'en', '1.0.0', 'gap-hash-abc', '{"version":"1.0.0","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["gap.ch1"]}]}', '{"version":"1.0.0","chapters":{"gap.ch1":"gap/ch1.json"}}')
		ON CONFLICT (bundle_id) DO UPDATE SET source_hash = EXCLUDED.source_hash`,
		bundleID)
	if err != nil {
		t.Fatalf("seed bundle meta: %v", err)
	}
	_, err = conn.Exec(`
		INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, ui_language, target_language, level, sort_order, raw_json, source_hash)
		VALUES (?, 'gap.ch1', 's1', 'Gap Chapter', 'ru', 'en', 'A1', 1, '{"id":"gap.ch1","section_id":"s1","title":"Gap Chapter","ui_language":"ru","target_language":"en","blocks":[],"question_bank":{},"chapter_test":{}}', 'ch-hash')
		ON CONFLICT (bundle_id, chapter_id) DO NOTHING`,
		bundleID)
	if err != nil {
		t.Fatalf("seed chapter: %v", err)
	}
}

func grammarGapSeedTrainingPack(t *testing.T, conn *sql.DB, bundleID string) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO grammar_training_content_meta (bundle_id, language, course_id, version, index_json, source_hash)
		VALUES (?, 'en', 'gap-course', '1.0.0', '{"version":"1.0.0","chapters":{"gap.ch1":"gap/q.json"},"blocks":{}}', 'tp-hash')
		ON CONFLICT (bundle_id) DO UPDATE SET index_json = EXCLUDED.index_json`,
		bundleID)
	if err != nil {
		t.Fatalf("seed training meta: %v", err)
	}
	_, err = conn.Exec(`
		INSERT INTO grammar_training_content_questions (bundle_id, chapter_id, theory_block_id, concept_id, question_id, source_hash, raw_json)
		VALUES (?, 'gap.ch1', 'tb1', 'c1', 'q1', 'qh1', '{"id":"q1","chapter_id":"gap.ch1","theory_block_id":"tb1","prompt":"?"}'),
		        (?, 'gap.ch1', 'tb1', 'c1', 'q2', 'qh2', '{"id":"q2","chapter_id":"gap.ch1","theory_block_id":"tb1","prompt":"??"}')
		ON CONFLICT (bundle_id, question_id) DO NOTHING`,
		bundleID, bundleID)
	if err != nil {
		t.Fatalf("seed training questions: %v", err)
	}
}

func TestGrammarReposGap_AttemptHasClientAttempt(t *testing.T) {
	conn := grammarGapDB(t)
	repo := grammarGapAttemptRepo(t, conn)
	userID := grammarGapUser(t, conn, grammarGapUserBase)

	exists, err := repo.HasClientAttempt(userID, "")
	if err != nil {
		t.Fatalf("HasClientAttempt(empty): %v", err)
	}
	if exists {
		t.Fatal("expected false for empty client attempt id")
	}

	clientID := "offline-gap-attempt-1"
	exists, err = repo.HasClientAttempt(userID, clientID)
	if err != nil {
		t.Fatalf("HasClientAttempt(missing): %v", err)
	}
	if exists {
		t.Fatal("expected false before insert")
	}

	finished := time.Now()
	cid := clientID
	_, err = repo.CreateAttempt(&TestAttempt{
		UserID:          userID,
		ScopeType:       "chapter",
		ScopeID:         "gap-ch-1",
		StartedAt:       time.Now(),
		FinishedAt:      &finished,
		Score:           70,
		Passed:          true,
		TotalQuestions:  5,
		AnswersJSON:     "[]",
		ResultsJSON:     "[]",
		ClientAttemptID: &cid,
	})
	if err != nil {
		t.Fatalf("CreateAttempt: %v", err)
	}

	exists, err = repo.HasClientAttempt(userID, clientID)
	if err != nil {
		t.Fatalf("HasClientAttempt(after): %v", err)
	}
	if !exists {
		t.Fatal("expected true after synced attempt")
	}
}

func TestGrammarReposGap_AttemptGetCategoryTestBestScores(t *testing.T) {
	conn := grammarGapDB(t)
	repo := grammarGapAttemptRepo(t, conn)
	userID := grammarGapUser(t, conn, grammarGapUserBase+1)
	finished := time.Now()

	for _, tc := range []struct {
		section string
		score   int
	}{
		{"gap-sec-a", 40},
		{"gap-sec-a", 75},
		{"gap-sec-b", 60},
	} {
		_, err := repo.CreateAttempt(&TestAttempt{
			UserID:         userID,
			ScopeType:      "category",
			ScopeID:        tc.section,
			StartedAt:      time.Now(),
			FinishedAt:     &finished,
			Score:          tc.score,
			Passed:         tc.score >= 50,
			TotalQuestions: 10,
			AnswersJSON:    "[]",
			ResultsJSON:    "[]",
		})
		if err != nil {
			t.Fatalf("CreateAttempt(%s): %v", tc.section, err)
		}
	}

	scores, err := repo.GetCategoryTestBestScores(userID)
	if err != nil {
		t.Fatalf("GetCategoryTestBestScores: %v", err)
	}
	if scores["gap-sec-a"] != 75 {
		t.Fatalf("gap-sec-a best = %d, want 75", scores["gap-sec-a"])
	}
	if scores["gap-sec-b"] != 60 {
		t.Fatalf("gap-sec-b best = %d, want 60", scores["gap-sec-b"])
	}

	emptyUser := grammarGapUser(t, conn, grammarGapUserBase+2)
	emptyScores, err := repo.GetCategoryTestBestScores(emptyUser)
	if err != nil {
		t.Fatalf("GetCategoryTestBestScores(empty): %v", err)
	}
	if len(emptyScores) != 0 {
		t.Fatalf("expected empty map, got %v", emptyScores)
	}
}

func TestGrammarReposGap_AttemptGetAllChapterProgress(t *testing.T) {
	conn := grammarGapDB(t)
	repo := grammarGapAttemptRepo(t, conn)
	userID := grammarGapUser(t, conn, grammarGapUserBase+3)

	if err := repo.UpdateProgress(userID, "gap-ch-a", 80, true); err != nil {
		t.Fatalf("UpdateProgress(a): %v", err)
	}
	if err := repo.UpdateProgress(userID, "gap-ch-b", 45, false); err != nil {
		t.Fatalf("UpdateProgress(b): %v", err)
	}

	all, err := repo.GetAllChapterProgress(userID)
	if err != nil {
		t.Fatalf("GetAllChapterProgress: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(all))
	}
	if all["gap-ch-a"].BestScore != 80 || !all["gap-ch-a"].Passed {
		t.Fatalf("gap-ch-a progress: %+v", all["gap-ch-a"])
	}
	if all["gap-ch-b"].BestScore != 45 || all["gap-ch-b"].Passed {
		t.Fatalf("gap-ch-b progress: %+v", all["gap-ch-b"])
	}

	emptyUser := grammarGapUser(t, conn, grammarGapUserBase+4)
	emptyAll, err := repo.GetAllChapterProgress(emptyUser)
	if err != nil {
		t.Fatalf("GetAllChapterProgress(empty): %v", err)
	}
	if len(emptyAll) != 0 {
		t.Fatalf("expected empty progress map, got %v", emptyAll)
	}
}

func TestGrammarReposGap_AttemptUpsertAndDeletePlacement(t *testing.T) {
	conn := grammarGapDB(t)
	repo := grammarGapAttemptRepo(t, conn)
	userID := grammarGapUser(t, conn, grammarGapUserBase+5)

	if err := repo.UpsertPlacementByAdmin(userID, 30, 20, []string{"s1"}); err != nil {
		t.Fatalf("UpsertPlacementByAdmin(insert): %v", err)
	}
	got, err := repo.GetPlacementTestResult(userID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult: %v", err)
	}
	if got == nil || got.Score != 30 || !got.AdminOverride || len(got.OpenedSections) != 1 {
		t.Fatalf("unexpected first upsert: %+v", got)
	}

	if err := repo.UpsertPlacementByAdmin(userID, 90, 20, []string{"s1", "s2", "s3"}); err != nil {
		t.Fatalf("UpsertPlacementByAdmin(update): %v", err)
	}
	got, err = repo.GetPlacementTestResult(userID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult(update): %v", err)
	}
	if got.Score != 90 || len(got.OpenedSections) != 3 || !got.AdminOverride {
		t.Fatalf("unexpected updated upsert: %+v", got)
	}

	if err := repo.DeletePlacementTestResult(userID); err != nil {
		t.Fatalf("DeletePlacementTestResult: %v", err)
	}
	got, err = repo.GetPlacementTestResult(userID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult(after delete): %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestGrammarReposGap_AttemptSavePlacementAdminOverrideReplace(t *testing.T) {
	conn := grammarGapDB(t)
	repo := grammarGapAttemptRepo(t, conn)
	userID := grammarGapUser(t, conn, grammarGapUserBase+6)

	if err := repo.UpsertPlacementByAdmin(userID, 80, 20, []string{"admin"}); err != nil {
		t.Fatalf("UpsertPlacementByAdmin: %v", err)
	}
	if err := repo.SavePlacementTestResult(userID, 40, 20, []string{"user"}); err != nil {
		t.Fatalf("SavePlacementTestResult(lower with admin override): %v", err)
	}
	got, err := repo.GetPlacementTestResult(userID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult: %v", err)
	}
	if got.Score != 40 || got.AdminOverride {
		t.Fatalf("user attempt should replace admin row: %+v", got)
	}
}

func TestGrammarReposGap_ContentConstructorsAndBundleVersionHash(t *testing.T) {
	logger := zap.NewNop()

	t.Run("NewGrammarContentRepositoryFromDB trims bundle id", func(t *testing.T) {
		conn := grammarGapDB(t)
		grammarGapSeedContentBundle(t, conn, "en")
		repo := NewGrammarContentRepositoryFromDB(conn, " EN ", logger)
		if repo == nil || repo.bundleID != "en" {
			t.Fatalf("repo bundleID = %q", repo.bundleID)
		}
		hash, err := repo.BundleVersionHash()
		if err != nil {
			t.Fatalf("BundleVersionHash(db): %v", err)
		}
		if hash != "gap-hash-abc" {
			t.Fatalf("hash = %q, want gap-hash-abc", hash)
		}
	})

	t.Run("BundleVersionHash from filesystem", func(t *testing.T) {
		mfs := fstest.MapFS{
			"sections.json": &fstest.MapFile{Data: []byte(`{"version":"1","sections":[]}`)},
			"index.json":    &fstest.MapFile{Data: []byte(`{"version":"1","chapters":{}}`)},
		}
		repo := NewGrammarContentRepositoryWithFS(mfs, logger)
		hash, err := repo.BundleVersionHash()
		if err != nil {
			t.Fatalf("BundleVersionHash(fs): %v", err)
		}
		if len(hash) != 64 {
			t.Fatalf("expected sha256 hex, got %q", hash)
		}
	})

	t.Run("NewGrammarContentRepositoryForLearning unknown bundle fails at read", func(t *testing.T) {
		lc := config.DefaultLearningConfig()
		lc.GrammarBundleID = "does-not-exist-gap"
		repo, err := NewGrammarContentRepositoryForLearning(lc, logger)
		if err != nil {
			return // constructor error path covered
		}
		_, err = repo.GetIndex()
		if err == nil {
			t.Fatal("expected error reading unknown embedded bundle")
		}
	})

	t.Run("DB-backed GetSections GetIndex GetChapterRawJSON", func(t *testing.T) {
		conn := grammarGapDB(t)
		grammarGapSeedContentBundle(t, conn, "gap-en")
		repo := NewGrammarContentRepositoryFromDB(conn, "gap-en", logger)
		sections, err := repo.GetSections()
		if err != nil || sections == nil || len(sections.Sections) != 1 {
			t.Fatalf("GetSections: err=%v sections=%+v", err, sections)
		}
		index, err := repo.GetIndex()
		if err != nil || index == nil || index.Chapters["gap.ch1"] == "" {
			t.Fatalf("GetIndex: err=%v index=%+v", err, index)
		}
		raw, err := repo.GetChapterRawJSON("gap.ch1")
		if err != nil || len(raw) == 0 {
			t.Fatalf("GetChapterRawJSON: err=%v len=%d", err, len(raw))
		}
		_, err = repo.GetChapterRawJSON("missing.chapter")
		if err == nil {
			t.Fatal("expected error for missing chapter")
		}
	})
}

func TestGrammarReposGap_TrainingPackConstructorsAndDB(t *testing.T) {
	logger := zap.NewNop()

	t.Run("NewGrammarTrainingPackRepository embedded", func(t *testing.T) {
		repo := NewGrammarTrainingPackRepository(logger)
		ok, n, err := repo.HasAnyQuestions()
		if err != nil {
			t.Fatalf("HasAnyQuestions: %v", err)
		}
		if !ok || n < 1 {
			t.Fatalf("expected embedded questions, ok=%v n=%d", ok, n)
		}
	})

	t.Run("NewGrammarTrainingPackRepositoryForLearning en and es", func(t *testing.T) {
		for _, id := range []string{"en", "es"} {
			lc := config.DefaultLearningConfig()
			lc.GrammarBundleID = id
			repo, err := NewGrammarTrainingPackRepositoryForLearning(lc, logger)
			if err != nil {
				t.Fatalf("ForLearning(%s): %v", id, err)
			}
			idx, err := repo.GetIndex()
			if err != nil || idx == nil {
				t.Fatalf("GetIndex(%s): err=%v", id, err)
			}
		}
	})

	t.Run("NewGrammarTrainingPackRepositoryForLearning unknown pack fails at read", func(t *testing.T) {
		lc := config.DefaultLearningConfig()
		lc.GrammarBundleID = "missing-pack-gap"
		repo, err := NewGrammarTrainingPackRepositoryForLearning(lc, logger)
		if err != nil {
			return // constructor error path covered
		}
		_, err = repo.GetIndex()
		if err == nil {
			t.Fatal("expected error reading unknown embedded pack")
		}
	})

	t.Run("NewGrammarTrainingPackRepositoryFromDB and getDBQuestions", func(t *testing.T) {
		conn := grammarGapDB(t)
		grammarGapSeedTrainingPack(t, conn, "gap-en")
		repo := NewGrammarTrainingPackRepositoryFromDB(conn, " GAP-EN ", logger)
		if repo.bundleID != "gap-en" {
			t.Fatalf("bundleID = %q", repo.bundleID)
		}
		idx, err := repo.GetIndex()
		if err != nil || idx.Chapters["gap.ch1"] == "" {
			t.Fatalf("GetIndex(db): err=%v idx=%+v", err, idx)
		}
		all, err := repo.GetAllQuestions()
		if err != nil || len(all) != 2 {
			t.Fatalf("GetAllQuestions(db): err=%v len=%d", err, len(all))
		}
		chQs, err := repo.GetChapterQuestions("gap.ch1")
		if err != nil || len(chQs) != 2 {
			t.Fatalf("GetChapterQuestions(db): err=%v len=%d", err, len(chQs))
		}
		byBlock, err := repo.QuestionsByTheoryBlock()
		if err != nil || len(byBlock["tb1"]) != 2 {
			t.Fatalf("QuestionsByTheoryBlock(db): err=%v by=%v", err, byBlock)
		}
	})
}

func TestGrammarReposGap_TrainingPackParseAndFSBranches(t *testing.T) {
	logger := zap.NewNop()

	t.Run("parseTrainingPackChapterFiles invalid", func(t *testing.T) {
		_, err := parseTrainingPackChapterFiles([]byte(`{"c1": 123}`))
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("parseTrainingPackIndex invalid top-level", func(t *testing.T) {
		_, err := parseTrainingPackIndex([]byte(`not-json`))
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("parseTrainingPackIndex invalid chapters", func(t *testing.T) {
		_, err := parseTrainingPackIndex([]byte(`{"chapters": {"c1": 1}}`))
		if err == nil {
			t.Fatal("expected chapters parse error")
		}
	})

	t.Run("parseTrainingPackChapterFiles skips empty paths", func(t *testing.T) {
		out, err := parseTrainingPackChapterFiles([]byte(`{"c1": "", "c2": "ok.json"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(out["c1"]) != 0 || out["c2"][0] != "ok.json" {
			t.Fatalf("unexpected: %#v", out)
		}
	})

	t.Run("GetAllQuestions skips invalid file with logger", func(t *testing.T) {
		mfs := fstest.MapFS{
			"index.json": &fstest.MapFile{Data: []byte(`{"chapters":{"c1":"bad.json","c2":"good.json"}}`)},
			"chapters/bad.json":  &fstest.MapFile{Data: []byte(`{invalid`)},
			"chapters/good.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":"q1","chapter_id":"c2","theory_block_id":"t1"}]}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, logger)
		qs, err := repo.GetAllQuestions()
		if err != nil {
			t.Fatalf("GetAllQuestions: %v", err)
		}
		if len(qs) != 1 {
			t.Fatalf("len=%d", len(qs))
		}
	})

	t.Run("GetChapterQuestions not found", func(t *testing.T) {
		mfs := fstest.MapFS{
			"index.json": &fstest.MapFile{Data: []byte(`{"chapters":{}}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
		_, err := repo.GetChapterQuestions("missing-chapter")
		if err == nil {
			t.Fatal("expected not found error")
		}
	})

	t.Run("readQuestionsFile null questions", func(t *testing.T) {
		mfs := fstest.MapFS{
			"chapters/null.json": &fstest.MapFile{Data: []byte(`{"questions":null}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
		qs, err := repo.readQuestionsFile("null.json")
		if err != nil {
			t.Fatalf("readQuestionsFile: %v", err)
		}
		if qs != nil {
			t.Fatalf("expected nil slice, got %#v", qs)
		}
	})

	t.Run("assignStableQuestionIDs partial fields", func(t *testing.T) {
		qs := []map[string]interface{}{
			nil,
			{"id": "keep"},
			{"id": "q1", "chapter_id": "c1"},
			{"id": "q2", "chapter_id": "c1", "theory_block_id": "t1"},
		}
		assignStableQuestionIDs(qs)
		if qs[1]["id"] != "keep" {
			t.Fatalf("id without block unchanged: %v", qs[1]["id"])
		}
		if qs[3]["id"] != "c1::t1::q2" {
			t.Fatalf("stable id = %v", qs[3]["id"])
		}
	})

	t.Run("QuestionsByTheoryBlock skips empty theory_block_id", func(t *testing.T) {
		mfs := fstest.MapFS{
			"index.json": &fstest.MapFile{Data: []byte(`{"chapters":{"c1":"q.json"}}`)},
			"chapters/q.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":"q1"},{"id":"q2","theory_block_id":"tb1","chapter_id":"c1","theory_block_id":"tb1"}]}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
		by, err := repo.QuestionsByTheoryBlock()
		if err != nil {
			t.Fatal(err)
		}
		if len(by) != 1 || len(by["tb1"]) != 1 {
			t.Fatalf("by=%v", by)
		}
	})

	t.Run("collectTrainingPackFilePaths skips blank paths", func(t *testing.T) {
		idx, err := parseTrainingPackIndex([]byte(`{"blocks":{"a::b":"  ","c::d":"path.json"}}`))
		if err != nil {
			t.Fatal(err)
		}
		paths := collectTrainingPackFilePaths(idx)
		if len(paths) != 1 || paths[0] != "path.json" {
			t.Fatalf("paths=%v", paths)
		}
	})

	t.Run("GetChapterQuestions skips invalid file with logger", func(t *testing.T) {
		mfs := fstest.MapFS{
			"index.json": &fstest.MapFile{Data: []byte(`{"blocks":{"c1::b1":"bad.json","c1::b2":"good.json"}}`)},
			"chapters/bad.json":  &fstest.MapFile{Data: []byte(`{`)},
			"chapters/good.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":"q1","chapter_id":"c1","theory_block_id":"b2"}]}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, logger)
		qs, err := repo.GetChapterQuestions("c1")
		if err != nil {
			t.Fatalf("GetChapterQuestions: %v", err)
		}
		if len(qs) != 1 {
			t.Fatalf("len=%d", len(qs))
		}
	})

	t.Run("HasAnyQuestions error path", func(t *testing.T) {
		repo := NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, nil)
		_, _, err := repo.HasAnyQuestions()
		if err == nil {
			t.Fatal("expected error for missing index")
		}
	})

	t.Run("readQuestionsFile invalid json", func(t *testing.T) {
		mfs := fstest.MapFS{
			"chapters/bad.json": &fstest.MapFile{Data: []byte(`{`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
		_, err := repo.readQuestionsFile("bad.json")
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("GetChapterQuestions skips empty block rel", func(t *testing.T) {
		mfs := fstest.MapFS{
			"index.json":         &fstest.MapFile{Data: []byte(`{"blocks":{"c1::b1":"  ","c1::b2":"good.json"}}`)},
			"chapters/good.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":"q1","chapter_id":"c1","theory_block_id":"b2"}]}`)},
		}
		repo := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
		qs, err := repo.GetChapterQuestions("c1")
		if err != nil || len(qs) != 1 {
			t.Fatalf("empty rel skip: err=%v len=%d", err, len(qs))
		}
	})

	t.Run("QuestionsByTheoryBlock propagates error", func(t *testing.T) {
		repo := NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, nil)
		_, err := repo.QuestionsByTheoryBlock()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("assignStableQuestionIDs blank ids", func(t *testing.T) {
		qs := []map[string]interface{}{
			{"id": "q1", "chapter_id": "  ", "theory_block_id": "t1"},
			{"id": "q2", "chapter_id": "c1", "theory_block_id": "  "},
		}
		assignStableQuestionIDs(qs)
		if qs[0]["id"] != "q1" || qs[1]["id"] != "q2" {
			t.Fatalf("ids changed: %#v", qs)
		}
	})

	t.Run("GetIndex db nil chapters guard", func(t *testing.T) {
		conn := grammarGapDB(t)
		_, err := conn.Exec(`
			INSERT INTO grammar_training_content_meta (bundle_id, language, course_id, version, index_json, source_hash)
			VALUES ('nil-ch', 'en', 'c', '1', '{"version":"1","chapters":null,"blocks":{}}', 'h')
			ON CONFLICT (bundle_id) DO UPDATE SET index_json = EXCLUDED.index_json`)
		if err != nil {
			t.Fatal(err)
		}
		repo := NewGrammarTrainingPackRepositoryFromDB(conn, "nil-ch", zap.NewNop())
		idx, err := repo.GetIndex()
		if err != nil || idx.Chapters == nil {
			t.Fatalf("GetIndex nil chapters guard: err=%v idx=%+v", err, idx)
		}
	})

	t.Run("GetSections and GetIndex fs error paths", func(t *testing.T) {
		repo := NewGrammarContentRepositoryWithFS(fstest.MapFS{}, zap.NewNop())
		if _, err := repo.GetSections(); err == nil {
			t.Fatal("expected sections error")
		}
		if _, err := repo.GetIndex(); err == nil {
			t.Fatal("expected index error")
		}
	})

	t.Run("GetChapterRawJSON fs missing file", func(t *testing.T) {
		repo := NewGrammarContentRepositoryWithFS(fstest.MapFS{
			"index.json": &fstest.MapFile{Data: []byte(`{"version":"1","chapters":{"ch1":"missing.json"}}`)},
		}, zap.NewNop())
		_, err := repo.GetChapterRawJSON("ch1")
		if err == nil {
			t.Fatal("expected read error")
		}
	})

	t.Run("collectTrainingPackFilePaths empty block rel", func(t *testing.T) {
		idx, err := parseTrainingPackIndex([]byte(`{"blocks":{"a::b":"","c::d":"ok.json"}}`))
		if err != nil {
			t.Fatal(err)
		}
		paths := collectTrainingPackFilePaths(idx)
		if len(paths) != 1 || paths[0] != "ok.json" {
			t.Fatalf("paths=%v", paths)
		}
	})
}

func TestGrammarReposGap_DBErrorPaths(t *testing.T) {
	logger := zap.NewNop()
	testutil.SetupTestDB(t)

	t.Run("HasClientAttempt db error", func(t *testing.T) {
		repo := NewGrammarAttemptRepository(grammarGapBrokenDB(t), logger)
		_, err := repo.HasClientAttempt(grammarGapUserBase, "cid-1")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetCategoryTestBestScores query error", func(t *testing.T) {
		repo := NewGrammarAttemptRepository(grammarGapBrokenDB(t), logger)
		_, err := repo.GetCategoryTestBestScores(grammarGapUserBase)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetAllChapterProgress query error", func(t *testing.T) {
		repo := NewGrammarAttemptRepository(grammarGapBrokenDB(t), logger)
		_, err := repo.GetAllChapterProgress(grammarGapUserBase)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("SavePlacementTestResult insert error", func(t *testing.T) {
		repo := NewGrammarAttemptRepository(grammarGapBrokenDB(t), logger)
		err := repo.SavePlacementTestResult(grammarGapUserBase, 50, 10, []string{"s1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("SavePlacementTestResult check query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT score").WillReturnError(fmt.Errorf("check failed"))
		repo := NewGrammarAttemptRepository(db, logger)
		err = repo.SavePlacementTestResult(grammarGapUserBase, 50, 10, []string{"s1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetCategoryTestBestScores scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"scope_id", "max"}).AddRow(nil, 10)
		mock.ExpectQuery("SELECT scope_id").WillReturnRows(rows)
		repo := NewGrammarAttemptRepository(db, logger)
		_, err = repo.GetCategoryTestBestScores(grammarGapUserBase)
		if err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("GetAllChapterProgress scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"chapter_id", "best_score", "passed_at", "last_attempt_at"}).AddRow(nil, 10, nil, nil)
		mock.ExpectQuery("SELECT chapter_id").WillReturnRows(rows)
		repo := NewGrammarAttemptRepository(db, logger)
		_, err = repo.GetAllChapterProgress(grammarGapUserBase)
		if err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("BundleVersionHash missing bundle", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT source_hash").WillReturnError(sql.ErrConnDone)
		repo := NewGrammarContentRepositoryFromDB(db, "en", logger)
		_, err = repo.BundleVersionHash()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("getDBQuestions query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT raw_json").WillReturnError(fmt.Errorf("query failed"))
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetAllQuestions()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("getDBQuestions invalid json", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"raw_json"}).AddRow("{invalid")
		mock.ExpectQuery("SELECT raw_json").WillReturnRows(rows)
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetAllQuestions()
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("getDBQuestions scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"raw_json"}).AddRow(nil)
		mock.ExpectQuery("SELECT raw_json").WillReturnRows(rows)
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetAllQuestions()
		if err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("getDBQuestions empty result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT raw_json").WillReturnRows(sqlmock.NewRows([]string{"raw_json"}))
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetAllQuestions()
		if err == nil {
			t.Fatal("expected no questions error")
		}
	})

	t.Run("GetSections db query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT sections_json").WillReturnError(sql.ErrConnDone)
		repo := NewGrammarContentRepositoryFromDB(db, "en", logger)
		_, err = repo.GetSections()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetIndex content db query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT index_json").WillReturnError(sql.ErrConnDone)
		repo := NewGrammarContentRepositoryFromDB(db, "en", logger)
		_, err = repo.GetIndex()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetIndex db invalid json", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"index_json"}).AddRow("{bad")
		mock.ExpectQuery("SELECT index_json").WillReturnRows(rows)
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetIndex()
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("GetIndex db missing bundle", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT index_json").WillReturnError(sql.ErrNoRows)
		repo := NewGrammarTrainingPackRepositoryFromDB(db, "en", logger)
		_, err = repo.GetIndex()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GetSections db parse error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{"sections_json"}).AddRow("{bad")
		mock.ExpectQuery("SELECT sections_json").WillReturnRows(rows)
		repo := NewGrammarContentRepositoryFromDB(db, "en", logger)
		_, err = repo.GetSections()
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("UpsertPlacementByAdmin db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectExec("INSERT INTO grammar_placement_test").WillReturnError(fmt.Errorf("exec failed"))
		repo := NewGrammarAttemptRepository(db, logger)
		err = repo.UpsertPlacementByAdmin(grammarGapUserBase, 10, 5, []string{"s1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("DeletePlacementTestResult db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectExec("DELETE FROM grammar_placement_test").WillReturnError(fmt.Errorf("delete failed"))
		repo := NewGrammarAttemptRepository(db, logger)
		err = repo.DeletePlacementTestResult(grammarGapUserBase)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGrammarReposGap_ContentBundleVersionHashFSErrors(t *testing.T) {
	logger := zap.NewNop()

	t.Run("missing sections.json", func(t *testing.T) {
		repo := NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
		_, err := repo.BundleVersionHash()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing index.json", func(t *testing.T) {
		repo := NewGrammarContentRepositoryWithFS(fstest.MapFS{
			"sections.json": &fstest.MapFile{Data: []byte(`{"sections":[]}`)},
		}, logger)
		_, err := repo.BundleVersionHash()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGrammarReposGap_GrammarFSForLearningDir(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "grammarbundle", "en"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "index.json")); err != nil {
		t.Skip("grammar bundle dir not available")
	}
	lc := config.LearningConfig{
		GrammarBundleDir: dir,
		GrammarBundleID:  "ignored-when-dir-set",
	}
	repo, err := NewGrammarContentRepositoryForLearning(lc, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetSections(); err != nil {
		t.Fatalf("GetSections from dir: %v", err)
	}
}
