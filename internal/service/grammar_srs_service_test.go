package service

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"path"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
	"testing/fstest"
)

type secondReadBadSectionsFS struct {
	valid fstest.MapFS
	reads int
}

func (s *secondReadBadSectionsFS) Open(name string) (fs.File, error) {
	clean := path.Clean(strings.TrimPrefix(name, "/"))
	if clean == "sections.json" {
		s.reads++
		if s.reads >= 2 {
			bad := fstest.MapFS{"sections.json": {Data: []byte(`{bad`)}} // force JSON parse error on 2nd read
			return bad.Open("sections.json")
		}
	}
	return s.valid.Open(clean)
}

func setupGrammarSRSServiceWithRepos(t *testing.T) (*GrammarService, *repository.GrammarAttemptRepository, int64, func()) {
	t.Helper()
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentFS := fstest.MapFS{
		"sections.json": {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":    {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{
			"schema_version":"1",
			"id":"ch1",
			"section_id":"s1",
			"title":"Chapter 1",
			"blocks":[{"id":"b1","type":"theory","theory":{"title":"T","content":"C"}}],
			"question_bank":{"questions":[{"id":"q1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A","explanation":"E","theory_block_id":"b1","chapter_id":"ch1","concept_id":"c1"}]},
			"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}
		}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(contentFS, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, err := userRepo.GetOrCreateUser(1)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	lc := config.DefaultLearningConfig()
	lc.TargetLang = "es"
	lc.GrammarBundleID = "es"
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, lc, logger)
	packFS := fstest.MapFS{
		"index.json": {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"one_questions.json"}}`)},
		"chapters/one_questions.json": {Data: []byte(`{
			"chapter_id":"ch1",
			"questions":[
				{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","concept_id":"c1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A","explanation":"E"},
				{"id":"q2","chapter_id":"ch1","theory_block_id":"b1","concept_id":"c1","type":"single_choice","question":"Q2?","options":["A","B"],"correct_answer":"B","explanation":"E2"}
			]
		}`)},
	}
	packRepo := repository.NewGrammarTrainingPackRepositoryWithFS(packFS, zap.NewNop())
	svc.SetTrainingPackRepository(packRepo)
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(db.GetConnection(), zap.NewNop()))

	if err := publishRepo.SetPublished("section", "s1", true, nil); err != nil {
		t.Fatalf("publish section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", "ch1", true, nil); err != nil {
		t.Fatalf("publish chapter: %v", err)
	}
	if err := svc.AttemptRepo.SavePlacementTestResult(user.ID, 100, 100, []string{"s1"}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	return svc, svc.AttemptRepo, user.ID, func() {}
}

func TestGrammarSRS_AvailabilityAndSessionAndAnswer_HappyPath(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	availability, err := svc.GetGrammarTrainingAvailability(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetGrammarTrainingAvailability: %v", err)
	}
	if !availability.Available || availability.QuestionCount != 2 || availability.TheoryBlockCount != 1 || availability.DueTheoryBlockCount != 1 {
		t.Fatalf("expected availability with 2 questions / 1 theory block / 1 due block, got %+v", availability)
	}

	session, err := svc.StartGrammarSrsSession(context.Background(), userID, 5)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession: %v", err)
	}
	if len(session.Items) == 0 {
		t.Fatal("expected non-empty grammar SRS session")
	}
	questionID, _ := session.Items[0].Question["id"].(string)
	if questionID == "" {
		t.Fatal("expected question id")
	}
	answerRes, err := svc.SubmitGrammarSrsAnswer(context.Background(), userID, questionID, session.Items[0].Question["correct_answer"])
	if err != nil {
		t.Fatalf("SubmitGrammarSrsAnswer: %v", err)
	}
	if !answerRes.Correct {
		t.Fatal("expected answer to be correct")
	}
}

func TestGrammarSRS_Availability_EdgeBranches(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	// training pack not configured
	availability, err := svc.GetGrammarTrainingAvailability(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGrammarTrainingAvailability without pack: %v", err)
	}
	if availability.Available || availability.QuestionCount != 0 || availability.TheoryBlockCount != 0 || availability.DueTheoryBlockCount != 0 {
		t.Fatalf("expected unavailable without pack, got %+v", availability)
	}

	// configured pack with missing index -> HasAnyQuestions error
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, zap.NewNop()))
	if _, err := svc.GetGrammarTrainingAvailability(context.Background(), 1); err == nil {
		t.Fatal("expected HasAnyQuestions error")
	}
}

func TestGrammarSRS_StartSession_ErrorAndEmptyBranches(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if _, err := svc.StartGrammarSrsSession(context.Background(), 1, 10); err == nil {
		t.Fatal("expected error when training pack repository is nil")
	}

	// invalid index fs
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, zap.NewNop()))
	if _, err := svc.StartGrammarSrsSession(context.Background(), 1, 10); err == nil {
		t.Fatal("expected error when reading questions by block fails")
	}

	// valid pack + no allowed chapters => empty session
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc = NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepository(logger))

	session, err := svc.StartGrammarSrsSession(context.Background(), 9999, 0) // unknown user, no placement/progress
	if err != nil {
		t.Fatalf("StartGrammarSrsSession empty: %v", err)
	}
	if len(session.Items) != 0 {
		t.Fatalf("expected empty session, got %d", len(session.Items))
	}
}

func TestGrammarSRS_SubmitAnswer_ErrorBranches(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if _, err := svc.SubmitGrammarSrsAnswer(context.Background(), 1, "q", "a"); err == nil {
		t.Fatal("expected error when training pack repository is nil")
	}

	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, zap.NewNop()))
	if _, err := svc.SubmitGrammarSrsAnswer(context.Background(), 1, "q", "a"); err == nil {
		t.Fatal("expected error when reading all questions fails")
	}
}

func TestGrammarSRS_SubmitAnswer_NotFoundAndNotAllowed(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	if _, err := svc.SubmitGrammarSrsAnswer(context.Background(), userID, "missing-question-id", "a"); err == nil {
		t.Fatal("expected not found for missing question id")
	}

	all, err := svc.TrainingPackRepo.GetAllQuestions()
	if err != nil || len(all) == 0 {
		t.Fatalf("GetAllQuestions: %v", err)
	}
	qID, _ := all[0]["id"].(string)
	if qID == "" {
		t.Fatal("expected question id")
	}
	if _, err := svc.SubmitGrammarSrsAnswer(context.Background(), 9999, qID, "a"); err == nil {
		t.Fatal("expected not found for inaccessible chapter")
	}

	// incorrect answer branch
	res, err := svc.SubmitGrammarSrsAnswer(context.Background(), userID, qID, "__wrong__")
	if err != nil {
		t.Fatalf("SubmitGrammarSrsAnswer incorrect: %v", err)
	}
	if res.Correct {
		t.Fatal("expected incorrect answer branch")
	}
}

func TestGrammarSRS_HelperBranches_FilterAndRecordAndUpdate(t *testing.T) {
	svc, _, _, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	// filter with empty inputs
	filtered := svc.filterBlocksByAllowedChapters(map[string][]map[string]interface{}{}, map[string]bool{})
	if len(filtered) != 0 {
		t.Fatalf("expected empty filtered map, got %+v", filtered)
	}

	// allowedTrainingChapters user=0 branch
	allowed, err := svc.allowedTrainingChapters(context.Background(), 0)
	if err != nil {
		t.Fatalf("allowedTrainingChapters user=0: %v", err)
	}
	if len(allowed) != 0 {
		t.Fatalf("expected empty allowed chapters for user=0, got %+v", allowed)
	}

	// RecordGrammarTheoryAttemptFromTest guard branches
	svc.RecordGrammarTheoryAttemptFromTest(1, nil, true, nil)
	svc.RecordGrammarTheoryAttemptFromTest(1, map[string]interface{}{"chapter_id": "x"}, true, nil)

	// updateTheoryMemory guard branches
	if err := svc.updateTheoryMemory(1, "", "", "", true); err != nil {
		t.Fatalf("updateTheoryMemory empty theory block should be nil error, got: %v", err)
	}
}

func TestGrammarSRS_Availability_AllowedEmptyBranch(t *testing.T) {
	// allowed empty branch
	svc, _, _, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()
	got, err := svc.GetGrammarTrainingAvailability(context.Background(), 999999)
	if err != nil {
		t.Fatalf("availability with unknown user: %v", err)
	}
	if got.Available || got.QuestionCount != 0 || got.TheoryBlockCount != 0 || got.DueTheoryBlockCount != 0 {
		t.Fatalf("expected unavailable for unknown user, got %+v", got)
	}
}

func TestGrammarSRS_StartSession_RemainingLoopBranches(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[{"id":"q1","correct_answer":"A"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}}`)},
	}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	svc.learning.TargetLang = "es"
	svc.learning.GrammarBundleID = "es"
	packFS := fstest.MapFS{
		"index.json": {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"q.json"}}`)},
		"chapters/q.json": {Data: []byte(`{"chapter_id":"ch1","questions":[
			{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","correct_answer":"A"},
			{"id":"q2","chapter_id":"ch1","theory_block_id":"b2","correct_answer":"B"},
			{"id":"q3","chapter_id":"ch1","theory_block_id":"b3","correct_answer":"C"}
		]}`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))
	svc.SetSRSRepository(nil) // force remaining path
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = svc.AttemptRepo.SavePlacementTestResult(user.ID, 100, 100, []string{"s1"})
	session, err := svc.StartGrammarSrsSession(context.Background(), user.ID, 1)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession remaining branches: %v", err)
	}
	if len(session.Items) != 1 {
		t.Fatalf("expected exactly one item, got %d", len(session.Items))
	}
}

func TestGrammarSRS_StartSession_DueLoopBreakBranch(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[{"id":"q1","correct_answer":"A"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}}`)},
	}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	svc.learning.TargetLang = "es"
	svc.learning.GrammarBundleID = "es"
	packFS := fstest.MapFS{
		"index.json": {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"q.json"}}`)},
		"chapters/q.json": {Data: []byte(`{"chapter_id":"ch1","questions":[
			{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","correct_answer":"A"},
			{"id":"q2","chapter_id":"ch1","theory_block_id":"b2","correct_answer":"B"}
		]}`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(db.GetConnection(), logger))
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = svc.AttemptRepo.SavePlacementTestResult(user.ID, 100, 100, []string{"s1"})
	// Preseed due list with two existing blocks, limit=1 should hit due-loop break on second iteration.
	_ = svc.SRSRepo.EnsureTheoryMemory(user.ID, "es", "es", "ch1", "b1", "c1")
	_, _ = db.GetConnection().Exec(
		`INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'new', CURRENT_TIMESTAMP)
		 ON CONFLICT(user_id, language, course_id, theory_block_id) DO UPDATE SET next_review_at=CURRENT_TIMESTAMP`,
		user.ID, "es", "es", "ch1", "b2", "c2",
	)
	if due, err := svc.SRSRepo.ListDueMemories(user.ID, "es", "es", time.Now(), 10); err != nil {
		t.Fatalf("ListDueMemories: %v", err)
	} else if len(due) == 0 {
		t.Fatal("expected non-empty due memories")
	}
	session, err := svc.StartGrammarSrsSession(context.Background(), user.ID, 1)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession due-loop break: %v", err)
	}
	if len(session.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(session.Items))
	}
}

func TestGrammarSRS_StartSession_DueTrimBranch_WithSQLMock(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[{"id":"q1","correct_answer":"A"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}}`)},
	}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	svc.learning.TargetLang = "es"
	svc.learning.GrammarBundleID = "es"
	packFS := fstest.MapFS{
		"index.json": {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"q.json"}}`)},
		"chapters/q.json": {Data: []byte(`{"chapter_id":"ch1","questions":[
			{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","correct_answer":"A"},
			{"id":"q2","chapter_id":"ch1","theory_block_id":"b2","correct_answer":"B"}
		]}`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = svc.AttemptRepo.SavePlacementTestResult(user.ID, 100, 100, []string{"s1"})

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer mockDB.Close()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectExec("INSERT INTO grammar_theory_memory").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO grammar_theory_memory").WillReturnResult(sqlmock.NewResult(1, 1))
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "language", "course_id", "chapter_id", "theory_block_id", "concept_id", "state",
		"review_count", "correct_count", "wrong_count", "lapse_count", "correct_streak", "wrong_streak",
		"ease", "interval_days", "mastery_score", "next_review_at", "last_review_at",
	}).AddRow(
		int64(1), user.ID, "es", "es", "ch1", "b1", "c1", "new", 0, 0, 0, 0, 0, 0, 2.5, 0, 0, now, nil,
	).AddRow(
		int64(2), user.ID, "es", "es", "ch1", "b2", "c2", "new", 0, 0, 0, 0, 0, 0, 2.5, 0, 0, now, nil,
	)
	mock.ExpectQuery("SELECT id, user_id, language, course_id").WillReturnRows(rows)
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(mockDB, logger))

	session, err := svc.StartGrammarSrsSession(context.Background(), user.ID, 1)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession with sqlmock due trim: %v", err)
	}
	if len(session.Items) != 1 {
		t.Fatalf("expected one item after due trim, got %d", len(session.Items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestGrammarSRS_AllowedTrainingChapters_IsSectionOpenedByPlacementErrorBranch(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	baseFS := fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[{"id":"q1","correct_answer":"A"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(&secondReadBadSectionsFS{valid: baseFS}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	// placement with unknown section to force level-based branch inside isSectionOpenedByPlacement,
	// which triggers the second ContentRepo.GetSections() call that now fails.
	if err := svc.AttemptRepo.SavePlacementTestResult(user.ID, 10, 10, []string{"unknown-section"}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}
	if _, err := svc.allowedTrainingChapters(context.Background(), user.ID); err == nil {
		t.Fatal("expected error from isSectionOpenedByPlacement second GetSections call")
	}
}

func TestGrammarSRS_AllowedTrainingChapters_ErrAndPassedBranches(t *testing.T) {
	// openedByPlacement error branch
	svcClosed, cleanup := grammarServiceWithClosedAttemptRepo(t)
	defer cleanup()
	if _, err := svcClosed.allowedTrainingChapters(context.Background(), 1); err == nil {
		t.Fatal("expected error from isSectionOpenedByPlacement with closed attempt repo")
	}

	// progress passed=true branch
	svc, _, _, _, _, cleanup2 := setupGrammarService(t)
	defer cleanup2()
	sections, err := svc.ContentRepo.GetSections()
	if err != nil || len(sections.Sections) == 0 || len(sections.Sections[0].ChapterIDs) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	ch := sections.Sections[0].ChapterIDs[0]
	if err := svc.AttemptRepo.UpdateProgress(1, ch, 90, true); err != nil {
		t.Fatalf("UpdateProgress: %v", err)
	}
	allowed, err := svc.allowedTrainingChapters(context.Background(), 1)
	if err != nil {
		t.Fatalf("allowedTrainingChapters: %v", err)
	}
	if !allowed[ch] {
		t.Fatalf("expected chapter %s to be allowed by passed progress", ch)
	}
}

func TestGrammarSRS_Availability_InvalidPackFilesAreSkipped(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	lc := config.DefaultLearningConfig()
	lc.TargetLang = "es"
	lc.GrammarBundleID = "es"
	svc.learning = lc

	// broken training pack: invalid chapter payload => HasAnyQuestions error path
	brokenFS := fstest.MapFS{
		"index.json":            {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"missing.json"}}`)},
		"chapters/missing.json": {Data: []byte(`{bad-json`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(brokenFS, zap.NewNop()))
	got, err := svc.GetGrammarTrainingAvailability(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error for invalid pack files: %v", err)
	}
	if got.Available || got.QuestionCount != 0 || got.TheoryBlockCount != 0 || got.DueTheoryBlockCount != 0 {
		t.Fatalf("expected unavailable because invalid files are skipped, got %+v", got)
	}
}

func TestGrammarSRS_StartSession_WithNilSRSAndClosedSRS(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	// branch: SRSRepo == nil
	svc.SetSRSRepository(nil)
	session, err := svc.StartGrammarSrsSession(context.Background(), userID, 50)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession nil SRS: %v", err)
	}
	if len(session.Items) == 0 {
		t.Fatal("expected items with nil SRS")
	}

	// branch: SRS repo list due returns error (closed connection)
	dsn := testutil.GetTestDSN(t)
	closedConn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skip("postgres_compat unavailable")
	}
	closedConn.Close()
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(closedConn, zap.NewNop()))
	session, err = svc.StartGrammarSrsSession(context.Background(), userID, 1)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession closed SRS should still continue: %v", err)
	}
	if len(session.Items) == 0 {
		t.Fatal("expected fallback selection when due memories fail")
	}
}

func TestGrammarSRS_SubmitAnswer_WithoutTheoryBlock(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentFS := fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(contentFS, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	packFS := fstest.MapFS{
		"index.json":                  {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"one_questions.json"}}`)},
		"chapters/one_questions.json": {Data: []byte(`{"chapter_id":"ch1","questions":[{"id":"q1","chapter_id":"ch1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A"}]}`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = svc.AttemptRepo.SavePlacementTestResult(user.ID, 100, 100, []string{"s1"})

	res, err := svc.SubmitGrammarSrsAnswer(context.Background(), user.ID, "q1", "A")
	if err != nil {
		t.Fatalf("SubmitGrammarSrsAnswer without theory block: %v", err)
	}
	if !res.Correct {
		t.Fatal("expected correct result")
	}
}

func TestGrammarSRS_AllowedTrainingChapters_ErrorPaths(t *testing.T) {
	// ContentRepo.GetSections error
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	if _, err := svc.allowedTrainingChapters(context.Background(), 1); err == nil {
		t.Fatal("expected error on GetSections")
	}

	// AttemptRepo.GetChapterProgress error (closed DB)
	svc2, cleanup := grammarServiceWithClosedAttemptRepo(t)
	defer cleanup()
	if _, err := svc2.allowedTrainingChapters(context.Background(), 1); err == nil {
		t.Fatal("expected error on GetChapterProgress with closed repo")
	}
}

func TestGrammarSRS_UpdateTheoryMemory_ErrorBranches(t *testing.T) {
	// EnsureTheoryMemory error with closed SRS repo
	svc, _, _, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()
	dsn := testutil.GetTestDSN(t)
	closedConn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skip("postgres_compat unavailable")
	}
	closedConn.Close()
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(closedConn, zap.NewNop()))
	if err := svc.updateTheoryMemory(1, "ch1", "b1", "c1", true); err == nil {
		t.Fatal("expected ensure/list due error for closed SRS repo")
	}
}

func TestGrammarSRS_ServiceErrorBranches_WithBrokenContentRepo(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	badContent := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)

	lc := config.DefaultLearningConfig()
	lc.TargetLang = "es"
	lc.GrammarBundleID = "es"
	svc := NewGrammarService(badContent, publishRepo, attemptRepo, lc, logger)
	packFS := fstest.MapFS{
		"index.json":                  {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"one_questions.json"}}`)},
		"chapters/one_questions.json": {Data: []byte(`{"chapter_id":"ch1","questions":[{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","correct_answer":"A"}]}`)},
	}
	svc.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))

	if _, err := svc.GetGrammarTrainingAvailability(context.Background(), user.ID); err == nil {
		t.Fatal("expected availability error when allowedTrainingChapters fails")
	}
	if _, err := svc.StartGrammarSrsSession(context.Background(), user.ID, 1); err == nil {
		t.Fatal("expected start session error when allowedTrainingChapters fails")
	}
	if _, err := svc.SubmitGrammarSrsAnswer(context.Background(), user.ID, "ch1::b1::q1", "A"); err == nil {
		t.Fatal("expected submit answer error when allowedTrainingChapters fails")
	}
}

func TestGrammarSRS_AllowedTrainingChapters_SkipsBlankChapterIDs(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	contentFS := fstest.MapFS{
		"sections.json":     {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":[" ","ch1"]}]}`)},
		"index.json":        {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{"schema_version":"1","id":"ch1","section_id":"s1","title":"Chapter 1","blocks":[],"question_bank":{"questions":[]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":[],"num_questions":1}}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(contentFS, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(1)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	allowed, err := svc.allowedTrainingChapters(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("allowedTrainingChapters: %v", err)
	}
	if allowed[" "] {
		t.Fatal("blank chapter id must be skipped")
	}
}

func TestGrammarSRS_UpdateTheoryMemory_NoDueMatchReturnsNil(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()
	if svc.SRSRepo == nil {
		t.Fatal("expected SRS repo")
	}
	// Ensure memory exists first.
	if err := svc.SRSRepo.EnsureTheoryMemory(userID, "es", "es", "ch1", "b1", "c1"); err != nil {
		t.Fatalf("EnsureTheoryMemory: %v", err)
	}
	// Move due date far into the future (> 365 days) by repeated correct updates.
	for i := 0; i < 25; i++ {
		mems, err := svc.SRSRepo.ListDueMemories(userID, "es", "es", time.Now().Add(10*365*24*time.Hour), 100)
		if err != nil {
			t.Fatalf("ListDueMemories: %v", err)
		}
		found := false
		for _, m := range mems {
			if m.TheoryBlockID == "b1" {
				found = true
				if err := svc.SRSRepo.UpdateAfterAnswer(m, true); err != nil {
					t.Fatalf("UpdateAfterAnswer: %v", err)
				}
				break
			}
		}
		if !found {
			break
		}
	}
	// If memory is beyond +365d window, updateTheoryMemory should return nil from the no-match tail path.
	if err := svc.updateTheoryMemory(userID, "ch1", "b1", "c1", true); err != nil {
		t.Fatalf("updateTheoryMemory no due match should not fail: %v", err)
	}
}

func TestGrammarSRS_UpdateTheoryMemory_ListDueErrorAndNoMatch(t *testing.T) {
	newSvc := func(mockDB *sql.DB) *GrammarService {
		svc, _, _, _, _, _ := setupGrammarService(t)
		svc.SRSRepo = repository.NewGrammarSRSRepository(mockDB, zap.NewNop())
		svc.learning.TargetLang = "es"
		svc.learning.GrammarBundleID = "es"
		return svc
	}

	t.Run("list due error branch", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectExec("INSERT INTO grammar_theory_memory").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT id, user_id, language, course_id").WillReturnError(errors.New("boom"))
		svc := newSvc(db)
		if err := svc.updateTheoryMemory(1, "ch1", "b1", "c1", true); err == nil {
			t.Fatal("expected list due error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	})

	t.Run("no match returns nil branch", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "user_id", "language", "course_id", "chapter_id", "theory_block_id", "concept_id", "state",
			"review_count", "correct_count", "wrong_count", "lapse_count", "correct_streak", "wrong_streak",
			"ease", "interval_days", "mastery_score", "next_review_at", "last_review_at",
		}).AddRow(
			int64(10), int64(1), "es", "es", "ch1", "other-block", "c1", "new",
			0, 0, 0, 0, 0, 0, 2.5, 0, 0, now, nil,
		)
		mock.ExpectExec("INSERT INTO grammar_theory_memory").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT id, user_id, language, course_id").WillReturnRows(rows)
		svc := newSvc(db)
		if err := svc.updateTheoryMemory(1, "ch1", "b1", "c1", true); err != nil {
			t.Fatalf("expected nil error on no matching block, got: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	})
}

func TestGrammarSRS_RecordAndUpdateTheoryMemory_ExtraBranches(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	// RecordGrammarTheoryAttemptFromTest with a valid question path.
	svc.RecordGrammarTheoryAttemptFromTest(userID, map[string]interface{}{
		"chapter_id":      "ch1",
		"theory_block_id": "b1",
		"concept_id":      "c1",
	}, false, nil)

	// updateTheoryMemory with nil SRS repo branch.
	svc.SetSRSRepository(nil)
	if err := svc.updateTheoryMemory(userID, "ch1", "b1", "c1", true); err != nil {
		t.Fatalf("expected nil error with nil SRS repo, got: %v", err)
	}

	// Restore SRS and call with empty chapter/concept to trigger TheoryIndex fallback branch.
	svc, _, userID, cleanup = setupGrammarSRSServiceWithRepos(t)
	defer cleanup()
	if err := svc.updateTheoryMemory(userID, "", "b1", "", true); err != nil {
		t.Fatalf("updateTheoryMemory TheoryIndex fallback: %v", err)
	}
}

func TestGrammarSRS_StartSession_DueIncludesUnknownBlock(t *testing.T) {
	svc, _, userID, cleanup := setupGrammarSRSServiceWithRepos(t)
	defer cleanup()

	if svc.SRSRepo == nil {
		t.Fatal("expected SRS repo")
	}
	// Seed due memory that is not present in byBlock to cover exists==false branch.
	_ = svc.SRSRepo.EnsureTheoryMemory(userID, "es", "es", "ch1", "missing-block", "cX")

	session, err := svc.StartGrammarSrsSession(context.Background(), userID, 2)
	if err != nil {
		t.Fatalf("StartGrammarSrsSession: %v", err)
	}
	if len(session.Items) == 0 {
		t.Fatal("expected session items")
	}
}
