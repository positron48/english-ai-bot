package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
	"testing/fstest"
)

func setupGrammarService(t *testing.T) (*GrammarService, *repository.GrammarContentRepository, *repository.GrammarPublishRepository, *repository.GrammarAttemptRepository, *repository.UserRepository, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	// Create user 1 for grammar_progress, grammar_placement_test FK
	_, _ = userRepo.GetOrCreateUser(1)

	service := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	cleanup := func() {} // shared db, do not close

	return service, contentRepo, publishRepo, attemptRepo, userRepo, cleanup
}

func TestGrammarService_GetPublishedSections_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.GetPublishedSections(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

func TestGrammarService_GetAllSectionsWithProgress_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.GetAllSectionsWithProgress(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

func TestGrammarService_GetPublishedChapters_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.GetPublishedChapters(context.Background(), "any-section", 1)
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

func grammarServiceWithClosedPublishRepo(t *testing.T) (*GrammarService, func()) {
	t.Helper()
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepository(logger)
	db := testutil.SetupTestDatabase(t)
	dsn := testutil.GetTestDSN(t)
	closedConn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	closedConn.Close()
	publishRepo := repository.NewGrammarPublishRepository(closedConn, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	return svc, func() {}
}

// grammarServiceWithClosedAttemptRepo returns a service with closed attempt DB so CreateAttempt/UpdateProgress etc. fail.
func grammarServiceWithClosedAttemptRepo(t *testing.T) (*GrammarService, func()) {
	t.Helper()
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepository(logger)
	db := testutil.SetupTestDatabase(t)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	_, _ = userRepo.GetOrCreateUser(1)
	dsn := testutil.GetTestDSN(t)
	closedConn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	closedConn.Close()
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(closedConn, logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	return svc, func() {}
}

func TestGrammarService_GetPublishedSections_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.GetPublishedSections(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

func TestGrammarService_GetAllSectionsWithProgress_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.GetAllSectionsWithProgress(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

func TestGrammarService_IsSectionPublished_RepoError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.IsSectionPublished(context.Background(), "any-section")
	if err == nil {
		t.Fatal("expected error when GetPublishedItem fails")
	}
}

func TestGrammarService_GetNextPublishedChapterID_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	contentRepo := repository.NewGrammarContentRepository(zap.NewNop())
	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Skip("need sections and chapters from bundle")
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]

	_, _, _, err = svc.GetNextPublishedChapterID(context.Background(), chapterID)
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

func TestGrammarService_GetPublishedSectionsAndChapters(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}

	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) == 0 {
		t.Fatal("expected chapter IDs")
	}
	chapterID := section.ChapterIDs[0]

	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("failed to publish section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}
	if err := attemptRepo.UpdateProgress(1, chapterID, 80, true); err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	sections, err := svc.GetPublishedSections(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetPublishedSections error: %v", err)
	}
	if len(sections) == 0 {
		t.Fatal("expected published sections")
	}

	chapters, err := svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters error: %v", err)
	}
	if len(chapters) == 0 {
		t.Fatal("expected published chapters")
	}
}

func TestGrammarService_GetAllSectionsWithProgress_IncludesUnpublished(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	if len(sectionsData.Sections) < 2 {
		t.Fatal("need at least 2 sections")
	}
	// Publish first section only
	section0 := sectionsData.Sections[0]
	_ = publishRepo.SetPublished("section", section0.SectionID, true, nil)
	_ = publishRepo.SetPublished("chapter", section0.ChapterIDs[0], true, nil)

	all, err := svc.GetAllSectionsWithProgress(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAllSectionsWithProgress error: %v", err)
	}
	if len(all) != len(sectionsData.Sections) {
		t.Fatalf("expected %d sections, got %d", len(sectionsData.Sections), len(all))
	}
	var publishedCount, unpublishedCount int
	for _, s := range all {
		if s.IsPublished {
			publishedCount++
		} else {
			unpublishedCount++
		}
	}
	if publishedCount != 1 {
		t.Fatalf("expected 1 published section, got %d", publishedCount)
	}
	if unpublishedCount == 0 {
		t.Fatal("expected at least one unpublished section")
	}
}

func TestGrammarService_GetPublishedChapters_UnpublishedSection_Error(t *testing.T) {
	svc, contentRepo, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]
	// Do not publish section

	_, err = svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err == nil {
		t.Fatal("expected error for unpublished section")
	}
	if !strings.Contains(err.Error(), "not published") {
		t.Fatalf("expected 'not published' in error, got: %v", err)
	}
}

func TestGrammarService_GetGrammarStatistics(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]

	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("failed to publish section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}
	if err := attemptRepo.UpdateProgress(1, chapterID, 80, true); err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	stats, err := svc.GetGrammarStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGrammarStatistics error: %v", err)
	}
	if stats.TotalChaptersInCourse <= 0 {
		t.Fatalf("expected TotalChaptersInCourse > 0, got %d", stats.TotalChaptersInCourse)
	}
	if stats.TotalChapters != 1 {
		t.Fatalf("expected TotalChapters 1, got %d", stats.TotalChapters)
	}
	if stats.PassedChapters != 1 {
		t.Fatalf("expected PassedChapters 1, got %d", stats.PassedChapters)
	}
	if stats.CourseCompletionPct != 80 {
		t.Fatalf("expected CourseCompletionPct 80, got %d", stats.CourseCompletionPct)
	}
	// Whole course: one published chapter with 80%, rest (unpublished or not attempted) count as 0
	if stats.WholeCourseCompletionPct < 0 || stats.WholeCourseCompletionPct > 80 {
		t.Fatalf("expected WholeCourseCompletionPct in [0, 80], got %d", stats.WholeCourseCompletionPct)
	}
}

func TestGrammarService_GetGrammarStatistics_WithPlacementOpenedSections(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	section := sectionsData.Sections[0]
	sectionID := section.SectionID
	chapterID := section.ChapterIDs[0]

	_ = publishRepo.SetPublished("section", sectionID, true, nil)
	_ = publishRepo.SetPublished("chapter", chapterID, true, nil)
	// Section opened by placement: all its chapters count as 100% and passed
	if err := attemptRepo.SavePlacementTestResult(1, 50, 10, []string{sectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	stats, err := svc.GetGrammarStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGrammarStatistics error: %v", err)
	}
	if stats.PassedChapters != 1 {
		t.Fatalf("expected PassedChapters 1 when section opened by placement, got %d", stats.PassedChapters)
	}
	if stats.CourseCompletionPct != 100 {
		t.Fatalf("expected CourseCompletionPct 100 when section opened by placement, got %d", stats.CourseCompletionPct)
	}
}

func TestGrammarService_GetContinueChapter(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) < 2 {
		t.Fatal("expected at least 2 sections")
	}

	section0 := sectionsData.Sections[0]
	section1 := sectionsData.Sections[1]
	ch0 := section0.ChapterIDs[0]
	ch1 := section0.ChapterIDs[1]
	chNextSection := section1.ChapterIDs[0]

	for _, id := range []string{section0.SectionID, section1.SectionID} {
		if err := publishRepo.SetPublished("section", id, true, nil); err != nil {
			t.Fatalf("publish section: %v", err)
		}
	}
	for _, id := range []string{ch0, ch1, chNextSection} {
		if err := publishRepo.SetPublished("chapter", id, true, nil); err != nil {
			t.Fatalf("publish chapter: %v", err)
		}
	}

	t.Run("no progress returns first accessible chapter", func(t *testing.T) {
		ch, err := svc.GetContinueChapter(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetContinueChapter: %v", err)
		}
		if ch == nil || ch.ChapterID != ch0 {
			t.Fatalf("expected first chapter %q, got %+v", ch0, ch)
		}
	})

	t.Run("passed first chapter returns next accessible chapter", func(t *testing.T) {
		if err := attemptRepo.UpdateProgress(1, ch0, 80, true); err != nil {
			t.Fatalf("UpdateProgress: %v", err)
		}
		ch, err := svc.GetContinueChapter(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetContinueChapter: %v", err)
		}
		if ch == nil || ch.ChapterID != ch1 {
			t.Fatalf("expected next chapter %q, got %+v", ch1, ch)
		}
	})

	t.Run("placement beyond first section returns frontier not first chapter", func(t *testing.T) {
		if err := attemptRepo.UpdateProgress(1, ch0, 10, false); err != nil {
			t.Fatalf("UpdateProgress: %v", err)
		}
		if err := attemptRepo.SavePlacementTestResult(1, 100, 25, []string{section0.SectionID, section1.SectionID}); err != nil {
			t.Fatalf("SavePlacementTestResult: %v", err)
		}
		ch, err := svc.GetContinueChapter(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetContinueChapter: %v", err)
		}
		if ch == nil || ch.ChapterID != chNextSection {
			t.Fatalf("expected frontier chapter %q, got %+v", chNextSection, ch)
		}
	})
}

func TestGrammarService_GenerateChapterAndCategoryTests(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]

	chapterID := section.ChapterIDs[0]
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}

	chapterTest, err := svc.GenerateChapterTest(context.Background(), chapterID)
	if err != nil {
		t.Fatalf("GenerateChapterTest error: %v", err)
	}
	if chapterTest.Total == 0 {
		t.Fatal("expected chapter test questions")
	}

	for _, q := range chapterTest.Questions {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		qType, _ := qMap["type"].(string)
		if qType != "reorder" {
			if _, ok := qMap["correct_answer"]; ok {
				t.Fatalf("expected correct_answer to be removed for %s", qType)
			}
		}
	}

	if len(section.ChapterIDs) > 1 {
		if err := publishRepo.SetPublished("chapter", section.ChapterIDs[1], true, nil); err != nil {
			t.Fatalf("failed to publish chapter: %v", err)
		}
	}
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("failed to publish section: %v", err)
	}

	categoryTest, err := svc.GenerateCategoryTest(context.Background(), section.SectionID)
	if err != nil {
		t.Fatalf("GenerateCategoryTest error: %v", err)
	}
	if categoryTest.Total == 0 {
		t.Fatal("expected category test questions")
	}

	for _, q := range categoryTest.Questions {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		qType, _ := qMap["type"].(string)
		if qType != "reorder" {
			if _, ok := qMap["correct_answer"]; ok {
				t.Fatalf("expected correct_answer to be removed for %s", qType)
			}
		}
	}
}

func TestGrammarService_SubmitTest_Chapter(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]

	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}

	chapter, err := contentRepo.GetChapter(chapterID)
	if err != nil {
		t.Fatalf("failed to get chapter: %v", err)
	}
	questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok || len(questionBank) < 2 {
		t.Fatalf("expected question bank")
	}

	var answers []AnswerItem
	for _, q := range questionBank {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		qID, _ := qMap["id"].(string)
		qType, _ := qMap["type"].(string)
		correct := qMap["correct_answer"]
		if qType == "true_false" {
			answers = append(answers, AnswerItem{QuestionID: qID, Answer: "да"})
			break
		}
		if correct != nil {
			answers = append(answers, AnswerItem{QuestionID: qID, Answer: correct})
			if len(answers) >= 2 {
				break
			}
		}
	}

	if len(answers) == 0 {
		t.Fatalf("expected answers to submit")
	}

	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest error: %v", err)
	}
	if result.Total == 0 {
		t.Fatalf("expected total questions")
	}

	progress, err := attemptRepo.GetChapterProgress(1, chapterID)
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress == nil {
		t.Fatalf("expected progress")
	}
}

func TestGrammarService_CompareAnswers(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if !svc.compareAnswers("a", "a") {
		t.Fatal("expected string comparison to match")
	}
	if svc.compareAnswers("a", "b") {
		t.Fatal("expected string comparison to fail")
	}
	// Case-insensitive
	if !svc.compareAnswers("Answer", "answer") {
		t.Fatal("expected case-insensitive match")
	}
	if !svc.compareAnswers("ANSWER", "answer") {
		t.Fatal("expected case-insensitive match")
	}
	// Spaces: trim and collapse
	if !svc.compareAnswers("  answer  ", "answer") {
		t.Fatal("expected trim match")
	}
	if !svc.compareAnswers("two   words", "two words") {
		t.Fatal("expected collapsed spaces match")
	}
	if !svc.compareAnswers("  Two  Words  ", "two words") {
		t.Fatal("expected trim + collapse + case match")
	}
	if svc.compareAnswers("two words", "twowords") {
		t.Fatal("expected different words to fail")
	}
	if !svc.compareAnswers([]interface{}{1, 2}, []interface{}{1, 2}) {
		t.Fatal("expected slice comparison to match")
	}
	if svc.compareAnswers([]interface{}{1, 2}, []interface{}{1, 2, 3}) {
		t.Fatal("expected slice length mismatch")
	}
	if !svc.compareAnswers(map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1}) {
		t.Fatal("expected map comparison to match")
	}
	if svc.compareAnswers(map[string]interface{}{"a": 1}, map[string]interface{}{"a": 2}) {
		t.Fatal("expected map comparison to fail")
	}
	// User map missing key that correct has
	if svc.compareAnswers(map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1, "b": 2}) {
		t.Fatal("expected map with missing key to fail")
	}
	// Slice recursive: inner element mismatch
	if svc.compareAnswers([]interface{}{1, 2}, []interface{}{1, 3}) {
		t.Fatal("expected slice with different element to fail")
	}
	// User answer not slice when correct is slice
	if svc.compareAnswers("not-a-slice", []interface{}{1, 2}) {
		t.Fatal("expected non-slice user answer to fail for slice correct")
	}
	// User answer not map when correct is map
	if svc.compareAnswers("not-a-map", map[string]interface{}{"a": 1}) {
		t.Fatal("expected non-map user answer to fail for map correct")
	}
}

func TestNormalizeTrueFalseValue(t *testing.T) {
	cases := map[interface{}]string{
		true:  "true",
		false: "false",
		"да":  "true",
		"нет": "false",
		"1":   "true",
		"0":   "false",
	}

	for input, expected := range cases {
		got, ok := normalizeTrueFalseValue(input)
		if !ok {
			t.Fatalf("expected value for %v", input)
		}
		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}

	if _, ok := normalizeTrueFalseValue(123); ok {
		t.Fatal("expected unsupported value")
	}
}

func TestGrammarService_SelectQuestions_Strategies(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	pool := []interface{}{"q1", "q2", "q3"}
	questionMap := map[string]interface{}{
		"q1": map[string]interface{}{"id": "q1", "theory_block_id": "a"},
		"q2": map[string]interface{}{"id": "q2", "theory_block_id": "a"},
		"q3": map[string]interface{}{"id": "q3", "theory_block_id": "b"},
	}

	config := map[string]interface{}{"type": "stratified_by_theory_block", "min_per_theory_block": 1.0}
	selected := svc.selectQuestions(pool, questionMap, config, 2)
	if len(selected) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(selected))
	}

	selectedRandom := svc.selectQuestions(pool, questionMap, map[string]interface{}{}, 2)
	if len(selectedRandom) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(selectedRandom))
	}

	// selectStratified: min_per_theory_block from config, theory_block_id empty -> "unknown"
	configMin := map[string]interface{}{"type": "stratified_by_theory_block", "min_per_theory_block": 2.0}
	pool3 := []interface{}{"q1", "q2", "q3", "q4"}
	qm3 := map[string]interface{}{
		"q1": map[string]interface{}{"id": "q1"},
		"q2": map[string]interface{}{"id": "q2"},
		"q3": map[string]interface{}{"id": "q3", "theory_block_id": "b1"},
		"q4": map[string]interface{}{"id": "q4", "theory_block_id": "b1"},
	}
	selectedStrat := svc.selectQuestions(pool3, qm3, configMin, 3)
	if len(selectedStrat) < 2 {
		t.Fatalf("expected at least 2 from stratified, got %d", len(selectedStrat))
	}

	// selectRandom: non-string in poolIDs skipped; len(available) <= numQuestions returns all
	poolMixed := []interface{}{"q1", 42, "q2"}
	qmSmall := map[string]interface{}{
		"q1": map[string]interface{}{"id": "q1"},
		"q2": map[string]interface{}{"id": "q2"},
	}
	selectedSmall := svc.selectRandom(poolMixed, qmSmall, 10)
	if len(selectedSmall) != 2 {
		t.Fatalf("expected 2 when available < numQuestions, got %d", len(selectedSmall))
	}

	// selectStratified: second pass fill remaining when numQuestions > sum of min per block
	poolStrat := []interface{}{"a1", "a2", "b1", "b2", "c1"}
	qmStrat := map[string]interface{}{
		"a1": map[string]interface{}{"id": "a1", "theory_block_id": "blockA"},
		"a2": map[string]interface{}{"id": "a2", "theory_block_id": "blockA"},
		"b1": map[string]interface{}{"id": "b1", "theory_block_id": "blockB"},
		"b2": map[string]interface{}{"id": "b2", "theory_block_id": "blockB"},
		"c1": map[string]interface{}{"id": "c1", "theory_block_id": "blockC"},
	}
	configStrat := map[string]interface{}{"type": "stratified_by_theory_block", "min_per_theory_block": 1.0}
	selectedStratFull := svc.selectQuestions(poolStrat, qmStrat, configStrat, 5)
	if len(selectedStratFull) != 5 {
		t.Fatalf("expected 5 from stratified with second pass fill, got %d", len(selectedStratFull))
	}

	// selectStratified: min_per_theory_block from config as float64; questions with empty theory_block_id go to "unknown"
	configMinFloat := map[string]interface{}{"type": "stratified_by_theory_block", "min_per_theory_block": 2.0}
	poolUnknown := []interface{}{"u1", "u2", "u3"}
	qmUnknown := map[string]interface{}{
		"u1": map[string]interface{}{"id": "u1"}, // no theory_block_id -> "unknown"
		"u2": map[string]interface{}{"id": "u2"},
		"u3": map[string]interface{}{"id": "u3", "theory_block_id": "b1"},
	}
	selectedUnknown := svc.selectStratified(poolUnknown, qmUnknown, configMinFloat, 3)
	if len(selectedUnknown) < 2 {
		t.Fatalf("expected at least 2 from stratified with min_per_theory_block 2, got %d", len(selectedUnknown))
	}

	// selectStratified: two questions in same block with same id ""; implementation may skip duplicate id
	poolDup := []interface{}{"p1", "p2"}
	qmDup := map[string]interface{}{
		"p1": map[string]interface{}{"id": "", "theory_block_id": "same"},
		"p2": map[string]interface{}{"id": "", "theory_block_id": "same"},
	}
	selectedDup := svc.selectStratified(poolDup, qmDup, map[string]interface{}{"type": "stratified_by_theory_block", "min_per_theory_block": 1.0}, 2)
	if len(selectedDup) < 1 || len(selectedDup) > 2 {
		t.Fatalf("expected 1 or 2 questions from stratified with duplicate empty id, got %d", len(selectedDup))
	}
}

func TestGrammarService_SubmitTest_Category(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]

	for i, chapterID := range section.ChapterIDs {
		if i >= 2 {
			break
		}
		if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
			t.Fatalf("failed to publish chapter: %v", err)
		}
	}
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("failed to publish section: %v", err)
	}

	categoryTest, err := svc.GenerateCategoryTest(context.Background(), section.SectionID)
	if err != nil {
		t.Fatalf("GenerateCategoryTest error: %v", err)
	}

	if len(categoryTest.Questions) == 0 {
		t.Fatalf("expected category test questions")
	}

	first := categoryTest.Questions[0].(map[string]interface{})
	questionID, _ := first["id"].(string)
	chapterID, _ := first["_category_test_chapter_id"].(string)
	answers := []AnswerItem{
		{QuestionID: questionID, ChapterID: chapterID, Answer: first["correct_answer"]},
	}

	result, err := svc.SubmitTest(context.Background(), 1, "category", section.SectionID, answers)
	if err != nil {
		t.Fatalf("SubmitTest error: %v", err)
	}
	if result.Total == 0 {
		t.Fatalf("expected total questions")
	}
}

func TestGrammarService_GetChapterContent_Filtering(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]

	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}

	chapter, err := svc.GetChapterContent(context.Background(), chapterID, false)
	if err != nil {
		t.Fatalf("GetChapterContent error: %v", err)
	}
	if chapter == nil {
		t.Fatal("expected chapter content")
	}

	if chapter.Chapter != nil && chapter.Chapter.QuestionBank != nil {
		payload, _ := json.Marshal(chapter.Chapter.QuestionBank)
		if len(payload) == 0 {
			t.Fatal("expected question bank payload")
		}
	}
}

func TestGrammarService_CanAccessSection(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	if len(sectionsData.Sections) < 2 {
		t.Fatalf("expected at least 2 sections")
	}

	firstSection := sectionsData.Sections[0]
	secondSection := sectionsData.Sections[1]

	canFirst, err := svc.CanAccessSection(context.Background(), 1, firstSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if !canFirst {
		t.Fatalf("expected first section to be accessible")
	}

	canSecond, err := svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if canSecond {
		t.Fatalf("expected second section to be locked without progress")
	}

	// Allow via placement result
	if err := attemptRepo.SavePlacementTestResult(1, 10, 10, []string{secondSection.SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult error: %v", err)
	}

	canSecond, err = svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if !canSecond {
		t.Fatalf("expected second section to be accessible via placement")
	}
}

func TestGrammarService_CanAccessSection_SectionNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.CanAccessSection(context.Background(), 1, "nonexistent-section-id")
	if err == nil {
		t.Fatal("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "section not found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'section not found' in error, got: %v", err)
	}
}

// TestGrammarService_CanAccessSection_AccessViaCategoryTestPassed covers access to next section when category test for previous section was passed (score >= 50%).
func TestGrammarService_CanAccessSection_AccessViaCategoryTestPassed(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) < 2 {
		t.Fatalf("need at least 2 sections")
	}
	firstSection := sectionsData.Sections[0]
	secondSection := sectionsData.Sections[1]

	// No placement result; second section is locked until category test for first section is passed.
	canSecond, err := svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if canSecond {
		t.Fatal("expected second section to be locked without category test passed")
	}

	// Create passed category test for first section (score >= 50, passed = true)
	now := time.Now()
	attempt := &repository.TestAttempt{
		UserID:         1,
		ScopeType:      "category",
		ScopeID:        firstSection.SectionID,
		StartedAt:      now,
		FinishedAt:     &now,
		Score:          50,
		Passed:         true,
		TotalQuestions: 10,
		AnswersJSON:    "[]",
		ResultsJSON:    "{}",
	}
	if _, err := attemptRepo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt category: %v", err)
	}

	canSecond, err = svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if !canSecond {
		t.Fatal("expected second section to be accessible after category test passed for first section")
	}
}

// TestGrammarService_CanAccessSection_EffectiveLevel covers access via placement effective level (section not in OpenedSections but level <= max opened level).
func TestGrammarService_CanAccessSection_EffectiveLevel(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections := pickSectionsWithQuestionChapters(t, contentRepo, 3)
	if len(sections) < 3 {
		t.Fatalf("need at least 3 sections with question chapters")
	}
	// Placements open a higher-level section; lower-level section should be accessible via effective level.
	higherSection := sections[2]
	middleSection := sections[1]
	lowerSection := sections[0]

	// Placement opened only the middle (or higher) section, not the lower one.
	if err := attemptRepo.SavePlacementTestResult(1, 80, 10, []string{higherSection.SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	// Higher section: in OpenedSections → accessible
	can, err := svc.CanAccessSection(context.Background(), 1, higherSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if !can {
		t.Fatal("expected explicitly opened section to be accessible")
	}

	// Middle/lower section: not in OpenedSections but level <= effective → accessible via effective level
	can, err = svc.CanAccessSection(context.Background(), 1, middleSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if !can {
		t.Fatal("expected section with level <= placement effective level to be accessible")
	}

	can, err = svc.CanAccessSection(context.Background(), 1, lowerSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if !can {
		t.Fatal("expected lower-level section to be accessible via effective level")
	}
}

// TestGrammarService_CanAccessSection_DeniedWhenCategoryNotPassed covers denial when second section has no placement and category test not passed.
func TestGrammarService_CanAccessSection_DeniedWhenCategoryNotPassed(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) < 2 {
		t.Fatalf("need at least 2 sections")
	}
	secondSection := sectionsData.Sections[1]

	// No placement, no category attempt for first section → second section locked
	can, err := svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if can {
		t.Fatal("expected second section to be denied without placement or category test passed")
	}

	// Category attempt exists but not passed (score < 50 or passed = false) → still denied
	now := time.Now()
	attempt := &repository.TestAttempt{
		UserID:         1,
		ScopeType:      "category",
		ScopeID:        sectionsData.Sections[0].SectionID,
		StartedAt:      now,
		FinishedAt:     &now,
		Score:          40,
		Passed:         false,
		TotalQuestions: 10,
		AnswersJSON:    "[]",
		ResultsJSON:    "{}",
	}
	if _, err := attemptRepo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt: %v", err)
	}
	can, err = svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if can {
		t.Fatal("expected second section to remain denied when category test not passed")
	}
}

// TestGrammarService_CanAccessSection_FallbackAllChaptersPassed covers access to next section when category test is missing but all published chapters of previous section are passed.
func TestGrammarService_CanAccessSection_FallbackAllChaptersPassed(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) < 2 {
		t.Fatalf("GetSections or need 2 sections: %v", err)
	}
	firstSection := sectionsData.Sections[0]
	secondSection := sectionsData.Sections[1]
	// Publish first section and its chapters; do not create category test attempt
	for _, chID := range firstSection.ChapterIDs {
		_ = publishRepo.SetPublished("chapter", chID, true, nil)
	}
	_ = publishRepo.SetPublished("section", firstSection.SectionID, true, nil)
	_ = publishRepo.SetPublished("section", secondSection.SectionID, true, nil)
	if len(secondSection.ChapterIDs) > 0 {
		_ = publishRepo.SetPublished("chapter", secondSection.ChapterIDs[0], true, nil)
	}
	// Pass all published chapters of first section (no category test)
	for _, chID := range firstSection.ChapterIDs {
		_ = attemptRepo.UpdateProgress(1, chID, 80, true)
	}

	can, err := svc.CanAccessSection(context.Background(), 1, secondSection.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection: %v", err)
	}
	if !can {
		t.Fatal("expected second section accessible when all chapters of first section passed (fallback)")
	}
}

func TestGrammarService_IsSectionPublished(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	// Unpublished
	ok, err := svc.IsSectionPublished(context.Background(), sectionID)
	if err != nil {
		t.Fatalf("IsSectionPublished error: %v", err)
	}
	if ok {
		t.Fatal("expected false when section not published")
	}

	_ = publishRepo.SetPublished("section", sectionID, true, nil)
	ok, err = svc.IsSectionPublished(context.Background(), sectionID)
	if err != nil {
		t.Fatalf("IsSectionPublished error: %v", err)
	}
	if !ok {
		t.Fatal("expected true when section published")
	}
}

func TestGrammarService_GetChapterContent_IncludeAnswers_And_Unpublished(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]

	// Unpublished chapter
	_, err = svc.GetChapterContent(context.Background(), chapterID, true)
	if err == nil {
		t.Fatal("expected error for unpublished chapter")
	}

	_ = publishRepo.SetPublished("chapter", chapterID, true, nil)
	content, err := svc.GetChapterContent(context.Background(), chapterID, true)
	if err != nil {
		t.Fatalf("GetChapterContent error: %v", err)
	}
	if content == nil {
		t.Fatal("expected content")
	}
	// includeAnswers=true: chapter not sanitized, question bank still filtered
	if content.Chapter == nil {
		t.Fatal("expected chapter")
	}
}

func TestGrammarService_SubmitTest_UnsupportedScopeType(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.SubmitTest(context.Background(), 1, "invalid", "x", nil)
	if err == nil {
		t.Fatal("expected error for unsupported scope type")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitTest_Chapter_ChapterNotFound covers error when chapter ID does not exist.
func TestGrammarService_SubmitTest_Chapter_ChapterNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.SubmitTest(context.Background(), 1, "chapter", "nonexistent-chapter-id", []AnswerItem{})
	if err == nil {
		t.Fatal("expected error for nonexistent chapter")
	}
	if !strings.Contains(err.Error(), "failed to get chapter") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected get chapter or not found in error, got: %v", err)
	}
}

// TestGrammarService_SubmitTest_Category_SectionNotFound covers error when section ID does not exist for category scope.
func TestGrammarService_SubmitTest_Category_SectionNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.SubmitTest(context.Background(), 1, "category", "nonexistent-section-id", []AnswerItem{})
	if err == nil {
		t.Fatal("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "section not found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected section not found in error, got: %v", err)
	}
}

func TestGrammarService_CompareAnswers_NonStringUserAnswer(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	// User answer as number, correct is string -> fmt.Sprintf
	if !svc.compareAnswers(42, "42") {
		t.Fatal("expected 42 and '42' to compare equal via string")
	}
	if svc.compareAnswers(43, "42") {
		t.Fatal("expected 43 and '42' to differ")
	}
}

func TestGrammarService_CompareAnswers_DefaultBranch(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	// Same scalar (e.g. float64)
	if !svc.compareAnswers(3.14, 3.14) {
		t.Fatal("expected equal scalars to match")
	}
	if svc.compareAnswers(3.14, 3.15) {
		t.Fatal("expected different scalars to differ")
	}
}

func TestGrammarService_SubmitTest_QuestionNotFound_SkipsAnswer(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	_ = publishRepo.SetPublished("chapter", chapterID, true, nil)

	// Submit with a question ID that doesn't exist in the chapter
	answers := []AnswerItem{
		{QuestionID: "nonexistent-id", Answer: "x"},
	}
	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest error: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("expected Total 0 when no valid questions, got %d", result.Total)
	}
	// Attempt may still be created with 0 questions
	_ = attemptRepo
}

func TestGrammarService_SubmitTest_NoAnswer_MarksIncorrect(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	_ = publishRepo.SetPublished("chapter", chapterID, true, nil)

	chapter, _ := contentRepo.GetChapter(chapterID)
	questionBank, _ := chapter.QuestionBank["questions"].([]interface{})
	var firstID string
	for _, q := range questionBank {
		qMap, _ := q.(map[string]interface{})
		if id, ok := qMap["id"].(string); ok && id != "" {
			firstID = id
			break
		}
	}
	if firstID == "" {
		t.Fatal("no question id in chapter")
	}

	// Answer with nil (no answer)
	answers := []AnswerItem{
		{QuestionID: firstID, Answer: nil},
	}
	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest error: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
	if result.Correct != 0 {
		t.Fatalf("expected Correct 0 for no answer, got %d", result.Correct)
	}
}

func TestGrammarService_SubmitTest_TrueFalse_Normalization(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	_ = publishRepo.SetPublished("chapter", chapterID, true, nil)

	chapter, _ := contentRepo.GetChapter(chapterID)
	questionBank, _ := chapter.QuestionBank["questions"].([]interface{})
	var qID string
	var correct interface{}
	for _, q := range questionBank {
		qMap, _ := q.(map[string]interface{})
		if tpe, _ := qMap["type"].(string); tpe == "true_false" {
			qID, _ = qMap["id"].(string)
			correct = qMap["correct_answer"]
			break
		}
	}
	if qID == "" {
		t.Skip("no true_false question in chapter")
	}

	var userAnswer interface{}
	if correct == true || correct == "true" {
		userAnswer = "да"
	} else {
		userAnswer = "нет"
	}
	answers := []AnswerItem{{QuestionID: qID, Answer: userAnswer}}
	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest error: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
	// Should match via normalizeTrueFalseValue
	if result.Correct != 1 {
		t.Fatalf("expected Correct 1 for normalized true_false, got %d", result.Correct)
	}
}

// TestGrammarService_SubmitTest_TrueFalse_UnnormalizableUsesCompare covers true_false when user or correct does not normalize; compareAnswers is used.
func TestGrammarService_SubmitTest_TrueFalse_UnnormalizableUsesCompare(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished: %v", err)
	}

	chapter, _ := contentRepo.GetChapter(chapterID)
	questionBank, _ := chapter.QuestionBank["questions"].([]interface{})
	var qID string
	for _, q := range questionBank {
		qMap, _ := q.(map[string]interface{})
		if tpe, _ := qMap["type"].(string); tpe == "true_false" {
			qID, _ = qMap["id"].(string)
			break
		}
	}
	if qID == "" {
		t.Skip("no true_false question in chapter")
	}

	// Answer with value that normalizeTrueFalseValue does not recognize; service falls back to compareAnswers
	answers := []AnswerItem{{QuestionID: qID, Answer: 42}}
	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
	// 42 vs correct (true/false) -> compareAnswers -> false
	if result.Correct != 0 {
		t.Fatalf("expected Correct 0 for unnormalizable answer, got %d", result.Correct)
	}
}

func TestGrammarService_CanAccessChapter(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Fatalf("expected at least 2 chapters")
	}

	// Publish section and first two chapters
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("failed to publish section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", section.ChapterIDs[0], true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", section.ChapterIDs[1], true, nil); err != nil {
		t.Fatalf("failed to publish chapter: %v", err)
	}

	canFirst, err := svc.CanAccessChapter(context.Background(), 1, section.ChapterIDs[0])
	if err != nil {
		t.Fatalf("CanAccessChapter error: %v", err)
	}
	if !canFirst {
		t.Fatalf("expected first chapter to be accessible")
	}

	canSecond, err := svc.CanAccessChapter(context.Background(), 1, section.ChapterIDs[1])
	if err != nil {
		t.Fatalf("CanAccessChapter error: %v", err)
	}
	if canSecond {
		t.Fatalf("expected second chapter to be locked before passing first")
	}

	if err := attemptRepo.UpdateProgress(1, section.ChapterIDs[0], 80, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	canSecond, err = svc.CanAccessChapter(context.Background(), 1, section.ChapterIDs[1])
	if err != nil {
		t.Fatalf("CanAccessChapter error: %v", err)
	}
	if !canSecond {
		t.Fatalf("expected second chapter to be accessible after passing first")
	}
}

// TestGrammarService_GetPublishedSections_OverrideName covers section title override from publish item name.
func TestGrammarService_GetPublishedSections_OverrideName(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]

	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	customName := "Custom Section Title"
	if err := publishRepo.SetName("section", section.SectionID, &customName, nil); err != nil {
		t.Fatalf("SetName: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}
	_ = attemptRepo.UpdateProgress(1, chapterID, 60, false)

	sections, err := svc.GetPublishedSections(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetPublishedSections: %v", err)
	}
	if len(sections) == 0 {
		t.Fatal("expected published sections")
	}
	if sections[0].Title != customName {
		t.Fatalf("expected title %q, got %q", customName, sections[0].Title)
	}
	if sections[0].ProgressPercentage != 60 {
		t.Fatalf("expected ProgressPercentage 60, got %d", sections[0].ProgressPercentage)
	}
}

// TestGrammarService_GetPublishedChapters_SectionNotFound covers error when section ID does not exist.
func TestGrammarService_GetPublishedChapters_SectionNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.GetPublishedChapters(context.Background(), "nonexistent-section-id", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestGrammarService_GetPublishedChapters_GetChapterSkipsInvalid covers skipping a chapter when GetChapter fails (e.g. missing in content).
func TestGrammarService_GetPublishedChapters_GetChapterSkipsInvalid(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1","ch2"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"Ch1","blocks":[],"question_bank":{"questions":[]},"chapter_test":{}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/one.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch2", true, nil)

	chapters, err := svc.GetPublishedChapters(context.Background(), "s1", 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters: %v", err)
	}
	// ch2 is not in content (index has only ch1), so only ch1 is returned
	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter (ch2 skipped), got %d", len(chapters))
	}
	if chapters[0].Chapter.ID != "ch1" {
		t.Fatalf("expected chapter ch1, got %s", chapters[0].Chapter.ID)
	}
}

// TestGrammarService_GetPublishedChapters_CanAccessWhenOpenedByPlacement covers CanAccess=true for all chapters when section opened by placement.
func TestGrammarService_GetPublishedChapters_CanAccessWhenOpenedByPlacement(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Skip("need at least 2 chapters in section")
	}

	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := publishRepo.SetPublished("chapter", section.ChapterIDs[i], true, nil); err != nil {
			t.Fatalf("SetPublished chapter: %v", err)
		}
	}
	if err := attemptRepo.SavePlacementTestResult(1, 50, 10, []string{section.SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	chapters, err := svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters: %v", err)
	}
	for i, ch := range chapters {
		if !ch.CanAccess {
			t.Fatalf("chapter at index %d expected CanAccess true when section opened by placement", i)
		}
	}
}

// TestGrammarService_GetPublishedChapters_FirstAndSecondAccessByPrevPassed covers CanAccess: first chapter by section access, second by previous passed (no placement).
func TestGrammarService_GetPublishedChapters_FirstAndSecondAccessByPrevPassed(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Skip("need at least 2 chapters")
	}
	ch0, ch1 := section.ChapterIDs[0], section.ChapterIDs[1]
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", ch0, true, nil); err != nil {
		t.Fatalf("SetPublished ch0: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", ch1, true, nil); err != nil {
		t.Fatalf("SetPublished ch1: %v", err)
	}
	// First section is accessible; pass first chapter so second gets CanAccess
	if err := attemptRepo.UpdateProgress(1, ch0, 80, true); err != nil {
		t.Fatalf("UpdateProgress: %v", err)
	}

	chapters, err := svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters: %v", err)
	}
	if len(chapters) < 2 {
		t.Fatalf("expected at least 2 chapters, got %d", len(chapters))
	}
	if !chapters[0].CanAccess {
		t.Fatal("expected first chapter CanAccess true (section accessible)")
	}
	if !chapters[1].CanAccess {
		t.Fatal("expected second chapter CanAccess true (previous passed)")
	}
}

// TestGrammarService_GetPublishedChapters_ChapterNameOverride covers chapter title override from publish item in GetPublishedChapters result.
func TestGrammarService_GetPublishedChapters_ChapterNameOverride(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}
	customTitle := "Custom Chapter Title in List"
	if err := publishRepo.SetName("chapter", chapterID, &customTitle, nil); err != nil {
		t.Fatalf("SetName: %v", err)
	}

	chapters, err := svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters: %v", err)
	}
	if len(chapters) == 0 {
		t.Fatal("expected at least one chapter")
	}
	if chapters[0].Title != customTitle {
		t.Fatalf("expected chapter title %q, got %q", customTitle, chapters[0].Title)
	}
}

// TestGrammarService_GetPublishedChapters_SkipUnpublishedChapters covers that only published chapters appear in result.
func TestGrammarService_GetPublishedChapters_SkipUnpublishedChapters(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Skip("need at least 2 chapters in section")
	}
	ch0, ch1 := section.ChapterIDs[0], section.ChapterIDs[1]
	_ = ch1 // second chapter intentionally not published
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	// Publish only first chapter
	if err := publishRepo.SetPublished("chapter", ch0, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}

	chapters, err := svc.GetPublishedChapters(context.Background(), section.SectionID, 1)
	if err != nil {
		t.Fatalf("GetPublishedChapters: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("expected 1 published chapter (second is unpublished), got %d", len(chapters))
	}
	if chapters[0].Chapter.ID != ch0 {
		t.Fatalf("expected chapter id %q, got %q", ch0, chapters[0].Chapter.ID)
	}
}

// TestGrammarService_GetNextPublishedChapterID_ChapterNotFound covers error when chapter ID does not exist.
func TestGrammarService_GetNextPublishedChapterID_ChapterNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, _, _, err := svc.GetNextPublishedChapterID(context.Background(), "nonexistent.chapter.id")
	if err == nil {
		t.Fatal("expected error for nonexistent chapter")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestGrammarService_FilterQuestionBankForQuizzes covers filterQuestionBankForQuizzes branches.
func TestGrammarService_FilterQuestionBankForQuizzes(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	t.Run("nil question bank", func(t *testing.T) {
		ch := &repository.Chapter{QuestionBank: nil, Blocks: []interface{}{}}
		out := svc.filterQuestionBankForQuizzes(ch)
		if out.QuestionBank != nil {
			t.Fatal("expected nil question bank unchanged")
		}
	})

	t.Run("block not map", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{"not-a-map"},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{"id": "q1"},
				},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		// usedQuestionIDs is empty, so no questions match
		qs, _ := out.QuestionBank["questions"].([]interface{})
		if len(qs) != 0 {
			t.Fatalf("expected 0 questions when no quiz_inline, got %d", len(qs))
		}
	})

	t.Run("quiz_inline with question_ids filters bank", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{
				map[string]interface{}{
					"type": "quiz_inline",
					"quiz_inline": map[string]interface{}{
						"question_ids": []interface{}{"q1", "q2"},
					},
				},
			},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{"id": "q1", "prompt": "Q1"},
					map[string]interface{}{"id": "q2", "prompt": "Q2"},
					map[string]interface{}{"id": "q3", "prompt": "Q3"},
				},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		qs, ok := out.QuestionBank["questions"].([]interface{})
		if !ok || len(qs) != 2 {
			t.Fatalf("expected 2 questions in bank, got %d", len(qs))
		}
		ids := make(map[string]bool)
		for _, q := range qs {
			qMap, _ := q.(map[string]interface{})
			ids[qMap["id"].(string)] = true
		}
		if !ids["q1"] || !ids["q2"] || ids["q3"] {
			t.Fatalf("expected q1 and q2 only, got %v", ids)
		}
	})

	t.Run("block type not quiz_inline", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{
				map[string]interface{}{"type": "theory", "content": "x"},
			},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{map[string]interface{}{"id": "q1"}},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		qs, _ := out.QuestionBank["questions"].([]interface{})
		if len(qs) != 0 {
			t.Fatalf("expected 0 questions, got %d", len(qs))
		}
	})

	t.Run("question_ids not slice", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{
				map[string]interface{}{
					"type":        "quiz_inline",
					"quiz_inline": map[string]interface{}{"question_ids": "not-a-slice"},
				},
			},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{map[string]interface{}{"id": "q1"}},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		qs, _ := out.QuestionBank["questions"].([]interface{})
		if len(qs) != 0 {
			t.Fatalf("expected 0 questions, got %d", len(qs))
		}
	})

	t.Run("question without id string skipped", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{
				map[string]interface{}{
					"type": "quiz_inline",
					"quiz_inline": map[string]interface{}{
						"question_ids": []interface{}{"q1"},
					},
				},
			},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{"id": 123},
					map[string]interface{}{"prompt": "no id"},
				},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		qs, _ := out.QuestionBank["questions"].([]interface{})
		if len(qs) != 0 {
			t.Fatalf("expected 0 questions (no valid id), got %d", len(qs))
		}
	})

	t.Run("quiz_inline value not map", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{
				map[string]interface{}{
					"type":        "quiz_inline",
					"quiz_inline": "not-a-map",
				},
			},
			QuestionBank: map[string]interface{}{
				"questions": []interface{}{map[string]interface{}{"id": "q1"}},
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		qs, _ := out.QuestionBank["questions"].([]interface{})
		if len(qs) != 0 {
			t.Fatalf("expected 0 questions when quiz_inline is not map, got %d", len(qs))
		}
	})

	t.Run("question bank questions not slice", func(t *testing.T) {
		ch := &repository.Chapter{
			Blocks: []interface{}{},
			QuestionBank: map[string]interface{}{
				"questions": "not-a-slice",
			},
		}
		out := svc.filterQuestionBankForQuizzes(ch)
		if out.QuestionBank == nil {
			t.Fatal("expected non-nil question bank")
		}
		// questions key is not []interface{}, so inner block is skipped; bank is unchanged
		if out.QuestionBank["questions"] != "not-a-slice" {
			t.Fatalf("expected questions to remain unchanged when not a slice, got %v", out.QuestionBank["questions"])
		}
	})
}

// TestGrammarService_GenerateChapterTest_ChapterNotFound covers error when chapter ID does not exist.
func TestGrammarService_GenerateChapterTest_ChapterNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.GenerateChapterTest(context.Background(), "nonexistent-chapter-id")
	if err == nil {
		t.Fatal("expected error for nonexistent chapter")
	}
	if !strings.Contains(err.Error(), "failed to get chapter") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected get chapter or not found in error, got: %v", err)
	}
}

// grammarServiceWithCustomChapterFS returns a service whose content repo serves a single chapter from MapFS (for error-path tests).
func grammarServiceWithCustomChapterFS(t *testing.T, sectionsJSON, indexJSON, chapterJSON string) *GrammarService {
	t.Helper()
	logger := zap.NewNop()
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/one.json": {Data: []byte(chapterJSON)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	return NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
}

// TestGrammarService_GenerateChapterTest_InvalidTestConfig covers error when selection_strategy is not a map.
func TestGrammarService_GenerateChapterTest_InvalidTestConfig(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill","correct_answer":"x"}]},"chapter_test":{"selection_strategy":"not-a-map","pool_question_ids":["q1"],"num_questions":10}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)

	_, err := svc.GenerateChapterTest(context.Background(), "ch1")
	if err == nil {
		t.Fatal("expected error for invalid test config")
	}
	if !strings.Contains(err.Error(), "invalid test config") {
		t.Errorf("expected 'invalid test config' in error, got: %v", err)
	}
}

// TestGrammarService_GenerateChapterTest_InvalidPoolQuestionIDs covers error when pool_question_ids is not a slice.
func TestGrammarService_GenerateChapterTest_InvalidPoolQuestionIDs(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":"not-a-slice","num_questions":10}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)

	_, err := svc.GenerateChapterTest(context.Background(), "ch1")
	if err == nil {
		t.Fatal("expected error for invalid pool_question_ids")
	}
	if !strings.Contains(err.Error(), "invalid pool_question_ids") {
		t.Errorf("expected 'invalid pool_question_ids' in error, got: %v", err)
	}
}

// TestGrammarService_GenerateChapterTest_InvalidQuestionBank covers error when question bank questions is not a slice.
func TestGrammarService_GenerateChapterTest_InvalidQuestionBank(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":"not-a-slice"},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":10}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)

	_, err := svc.GenerateChapterTest(context.Background(), "ch1")
	if err == nil {
		t.Fatal("expected error for invalid question bank")
	}
	if !strings.Contains(err.Error(), "invalid question bank") {
		t.Errorf("expected 'invalid question bank' in error, got: %v", err)
	}
}

// TestGrammarService_GenerateChapterTest_QuestionWithNonStringID covers question bank entry with id not a string (skipped when building questionMap).
func TestGrammarService_GenerateChapterTest_QuestionWithNonStringID(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	// One question with string id "q1", one with numeric id so it is skipped in questionMap
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill","correct_answer":"x"},{"id":123,"type":"fill","correct_answer":"y"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":10}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)

	out, err := svc.GenerateChapterTest(context.Background(), "ch1")
	if err != nil {
		t.Fatalf("GenerateChapterTest: %v", err)
	}
	// Only q1 is in questionMap (the other has id 123 non-string), so only one question is selected
	if len(out.Questions) != 1 {
		t.Fatalf("expected 1 question in result (only q1 has string id), got %d", len(out.Questions))
	}
}

// TestGrammarService_GenerateCategoryTest_GetChapterFailsForOneChapter covers skipping a published chapter when GetChapter fails (e.g. missing in content).
func TestGrammarService_GenerateCategoryTest_GetChapterFailsForOneChapter(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1","ch2"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"Ch1","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill","correct_answer":"a"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":10}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/one.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch2", true, nil) // ch2 not in content; GetChapter will fail

	out, err := svc.GenerateCategoryTest(context.Background(), "s1")
	if err != nil {
		t.Fatalf("GenerateCategoryTest: %v", err)
	}
	// Only ch1 has content; ch2 is skipped due to GetChapter error
	if out.Total == 0 {
		t.Fatal("expected at least one question from ch1")
	}
}

// TestGrammarService_SubmitTest_Chapter_InvalidQuestionBank covers SubmitTest when chapter question bank is not a slice.
func TestGrammarService_SubmitTest_Chapter_InvalidQuestionBank(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":"not-a-slice"}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)

	_, err := svc.SubmitTest(context.Background(), 1, "chapter", "ch1", []AnswerItem{})
	if err == nil {
		t.Fatal("expected error for invalid question bank in SubmitTest")
	}
	if !strings.Contains(err.Error(), "invalid question bank") {
		t.Errorf("expected 'invalid question bank' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitTest_Chapter_QuestionNotMap covers question bank entry that is not a map (skipped when building questionMap).
func TestGrammarService_SubmitTest_Chapter_QuestionNotMap(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	// questions slice contains a string (not a map) so it is skipped in the loop
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill","correct_answer":"x"},"not-a-map"]}}`
	svc := grammarServiceWithCustomChapterFS(t, sectionsJSON, indexJSON, chapterJSON)
	_ = svc.PublishRepo.SetPublished("chapter", "ch1", true, nil)

	result, err := svc.SubmitTest(context.Background(), 1, "chapter", "ch1", []AnswerItem{{QuestionID: "q1", Answer: "x"}})
	if err != nil {
		t.Fatalf("SubmitTest: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
	if result.Correct != 1 {
		t.Fatalf("expected Correct 1, got %d", result.Correct)
	}
}

// TestGrammarService_SubmitTest_SaveAttemptFails covers SubmitTest when CreateAttempt fails (e.g. closed attempt DB); result is still returned.
func TestGrammarService_SubmitTest_SaveAttemptFails(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedAttemptRepo(t)
	defer cleanup()

	contentRepo := repository.NewGrammarContentRepository(zap.NewNop())
	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Skip("need sections and chapters from bundle")
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	// Publish chapter via the same DB that the service's publishRepo uses (we can't access it; use setup then closed attempt)
	// grammarServiceWithClosedAttemptRepo uses content from bundle and publish from db, so we need to publish the chapter.
	// We don't have direct access to publishRepo in the test after getting svc. So we need to publish in the helper or pass db.
	// Actually we have svc.PublishRepo - so we can call svc.PublishRepo.SetPublished. So do that.
	if err := svc.PublishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished: %v", err)
	}
	chapter, err := contentRepo.GetChapter(chapterID)
	if err != nil {
		t.Fatalf("GetChapter: %v", err)
	}
	questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok || len(questionBank) == 0 {
		t.Skip("chapter has no questions")
	}
	var qID string
	var correct interface{}
	for _, q := range questionBank {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := qMap["id"].(string); ok && id != "" {
			qID = id
			correct = qMap["correct_answer"]
			break
		}
	}
	if qID == "" {
		t.Skip("no question id in chapter")
	}

	answers := []AnswerItem{{QuestionID: qID, Answer: correct}}
	result, err := svc.SubmitTest(context.Background(), 1, "chapter", chapterID, answers)
	if err != nil {
		t.Fatalf("SubmitTest should succeed and return result even when SaveAttempt fails: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
}

// TestGrammarService_GenerateCategoryTest_GetPublishedItemsError covers error when GetPublishedItemsByType fails.
func TestGrammarService_GenerateCategoryTest_GetPublishedItemsError(t *testing.T) {
	_, contentRepo, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()
	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Skip("need sections from bundle")
	}
	sectionID := sectionsData.Sections[0].SectionID

	svcClosed, cleanupClosed := grammarServiceWithClosedPublishRepo(t)
	defer cleanupClosed()

	_, err = svcClosed.GenerateCategoryTest(context.Background(), sectionID)
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

// TestGrammarService_GenerateCategoryTest_NoQuestionsInSection covers error when section has no published chapters with questions.
func TestGrammarService_GenerateCategoryTest_NoQuestionsInSection(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	// Publish section but no chapters — no questions available
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}

	_, err = svc.GenerateCategoryTest(context.Background(), section.SectionID)
	if err == nil {
		t.Fatal("expected error when no questions in section")
	}
	if !strings.Contains(err.Error(), "no questions available") {
		t.Errorf("expected 'no questions available' in error, got: %v", err)
	}
}

// TestGrammarService_GeneratePlacementTest_GetSectionsError covers error when GetSections fails.
func TestGrammarService_GeneratePlacementTest_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.GeneratePlacementTest(context.Background())
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

// TestGrammarService_GeneratePlacementTest_GetPublishedItemsError covers error when GetPublishedItemsByType fails.
func TestGrammarService_GeneratePlacementTest_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.GeneratePlacementTest(context.Background())
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

// TestGrammarService_GetNextPublishedChapterID_GetSectionsError covers error when GetSections fails after GetChapter succeeds.
func TestGrammarService_GetNextPublishedChapterID_GetSectionsError(t *testing.T) {
	// MapFS with index + chapter so GetChapter works, but no sections.json so GetSections fails
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"T","blocks":[],"question_bank":{},"chapter_test":{}}`
	fs := fstest.MapFS{
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/one.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, _, _, err := svc.GetNextPublishedChapterID(context.Background(), "ch1")
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

// TestGrammarService_GetNextPublishedChapterID_SectionNotFound covers error when chapter's section_id is not in sections list.
func TestGrammarService_GetNextPublishedChapterID_SectionNotFound(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch2":"two.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch2","section_id":"s2","title":"T2","blocks":[],"question_bank":{},"chapter_test":{}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/two.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	_ = publishRepo.SetPublished("chapter", "ch2", true, nil)

	_, _, _, err := svc.GetNextPublishedChapterID(context.Background(), "ch2")
	if err == nil {
		t.Fatal("expected error when section not found for chapter's section_id")
	}
	if !strings.Contains(err.Error(), "section not found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'section not found' in error, got: %v", err)
	}
}

// TestGrammarService_GetNextPublishedChapterID_ChapterNotInSectionList covers error when chapter is not in section's ChapterIDs.
func TestGrammarService_GetNextPublishedChapterID_ChapterNotInSectionList(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch2":"two.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch2","section_id":"s1","title":"T2","blocks":[],"question_bank":{},"chapter_test":{}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/two.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	_ = publishRepo.SetPublished("chapter", "ch2", true, nil)

	_, _, _, err := svc.GetNextPublishedChapterID(context.Background(), "ch2")
	if err == nil {
		t.Fatal("expected error when chapter not in section list")
	}
	if !strings.Contains(err.Error(), "chapter not found in section list") {
		t.Errorf("expected 'chapter not found in section list' in error, got: %v", err)
	}
}

// TestGrammarService_GenerateCategoryTest_SectionNotFound covers error when section does not exist.
func TestGrammarService_GenerateCategoryTest_SectionNotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.GenerateCategoryTest(context.Background(), "nonexistent-section-id")
	if err == nil {
		t.Fatal("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitTest_Category_NoQuestions covers error when section has no published chapters.
func TestGrammarService_SubmitTest_Category_NoQuestions(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	// Publish section but no chapters
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}

	_, err = svc.SubmitTest(context.Background(), 1, "category", section.SectionID, []AnswerItem{
		{QuestionID: "q1", ChapterID: section.ChapterIDs[0], Answer: "x"},
	})
	if err == nil {
		t.Fatal("expected error when no published chapters in section")
	}
	if !strings.Contains(err.Error(), "no questions") {
		t.Errorf("expected 'no questions' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitPlacementTest_GetSectionsError covers error when GetSections fails.
func TestGrammarService_SubmitPlacementTest_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.SubmitPlacementTest(context.Background(), 1, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitPlacementTest_GetPublishedItemsError covers error when GetPublishedItemsByType fails.
func TestGrammarService_SubmitPlacementTest_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.SubmitPlacementTest(context.Background(), 1, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

// TestGrammarService_SubmitTest_Category_GetChapterFailsForOneChapter covers category submit when one published chapter fails GetChapter (skipped when building questionMapByChapter).
func TestGrammarService_SubmitTest_Category_GetChapterFailsForOneChapter(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1","ch2"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch1","section_id":"s1","title":"Ch1","blocks":[],"question_bank":{"questions":[{"id":"q1","type":"fill","correct_answer":"ans1"}]},"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":10}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/one.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	_, _ = userRepo.GetOrCreateUser(1)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch2", true, nil) // ch2 not in content; GetChapter fails, chapter skipped

	result, err := svc.SubmitTest(context.Background(), 1, "category", "s1", []AnswerItem{
		{QuestionID: "q1", ChapterID: "ch1", Answer: "ans1"},
	})
	if err != nil {
		t.Fatalf("SubmitTest: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
	if result.Correct != 1 {
		t.Fatalf("expected Correct 1, got %d", result.Correct)
	}
}

// TestGrammarService_SubmitTest_Category_MissingChapterID exercises fallback search when chapter_id is empty.
func TestGrammarService_SubmitTest_Category_MissingChapterID(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	var firstChapterID string
	for i, chID := range section.ChapterIDs {
		if i >= 2 {
			break
		}
		if err := publishRepo.SetPublished("chapter", chID, true, nil); err != nil {
			t.Fatalf("SetPublished chapter: %v", err)
		}
		firstChapterID = chID
	}
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}

	catTest, err := svc.GenerateCategoryTest(context.Background(), section.SectionID)
	if err != nil {
		t.Fatalf("GenerateCategoryTest: %v", err)
	}
	if len(catTest.Questions) == 0 {
		t.Fatal("expected category test questions")
	}
	first := catTest.Questions[0].(map[string]interface{})
	qID, _ := first["id"].(string)
	// correct_answer is stripped in GenerateCategoryTest; get it from content
	ch, err := contentRepo.GetChapter(firstChapterID)
	if err != nil {
		t.Fatalf("GetChapter: %v", err)
	}
	qBank, _ := ch.QuestionBank["questions"].([]interface{})
	var correctAns interface{}
	for _, q := range qBank {
		qMap, _ := q.(map[string]interface{})
		if qMap["id"] == qID {
			correctAns = qMap["correct_answer"]
			break
		}
	}
	if correctAns == nil {
		t.Skip("no correct_answer in chapter for question")
	}

	// Submit with empty ChapterID; service should find question via fallback search in all chapters
	answers := []AnswerItem{
		{QuestionID: qID, ChapterID: "", Answer: correctAns},
	}
	result, err := svc.SubmitTest(context.Background(), 1, "category", section.SectionID, answers)
	if err != nil {
		t.Fatalf("SubmitTest: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected Total 1, got %d", result.Total)
	}
}

// TestGrammarService_CanAccessChapter_OpenedByPlacement covers access when section is opened by placement test.
func TestGrammarService_CanAccessChapter_OpenedByPlacement(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Skip("need at least 2 chapters")
	}
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", section.ChapterIDs[1], true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}
	if err := attemptRepo.SavePlacementTestResult(1, 60, 10, []string{section.SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	can, err := svc.CanAccessChapter(context.Background(), 1, section.ChapterIDs[1])
	if err != nil {
		t.Fatalf("CanAccessChapter: %v", err)
	}
	if !can {
		t.Fatal("expected second chapter accessible when section opened by placement")
	}
}

// TestGrammarService_CanAccessChapter_NotFound covers error when chapter ID does not exist.
func TestGrammarService_CanAccessChapter_NotFound(t *testing.T) {
	svc, _, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	_, err := svc.CanAccessChapter(context.Background(), 1, "nonexistent-chapter-id")
	if err == nil {
		t.Fatal("expected error for nonexistent chapter")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "failed to get chapter") {
		t.Errorf("expected not found or failed to get chapter in error, got: %v", err)
	}
}

// TestGrammarService_CanAccessChapter_ChapterNotInSectionList covers error when chapter is not in section's ChapterIDs.
func TestGrammarService_CanAccessChapter_ChapterNotInSectionList(t *testing.T) {
	sectionsJSON := `{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`
	indexJSON := `{"version":"1","generated_at":"","chapters":{"ch2":"two.json"}}`
	chapterJSON := `{"schema_version":"1","id":"ch2","section_id":"s1","title":"T2","blocks":[],"question_bank":{},"chapter_test":{}}`
	fs := fstest.MapFS{
		"sections.json":     {Data: []byte(sectionsJSON)},
		"index.json":        {Data: []byte(indexJSON)},
		"chapters/two.json": {Data: []byte(chapterJSON)},
	}
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.CanAccessChapter(context.Background(), 1, "ch2")
	if err == nil {
		t.Fatal("expected error when chapter not in section list")
	}
	if !strings.Contains(err.Error(), "chapter not found in section list") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'chapter not found in section list' or 'not found' in error, got: %v", err)
	}
}

// TestGrammarService_IsSectionOpenedByPlacement_ViaEffectiveLevel covers true when section level <= placement effective level (section not in OpenedSections).
func TestGrammarService_IsSectionOpenedByPlacement_ViaEffectiveLevel(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections := pickSectionsWithQuestionChapters(t, contentRepo, 2)
	if len(sections) < 2 {
		t.Fatalf("need at least 2 sections")
	}
	higherSection := sections[1]
	lowerSection := sections[0]
	// Placement opened only the higher-level section; lower should be accessible via effective level
	if err := attemptRepo.SavePlacementTestResult(1, 80, 10, []string{higherSection.SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	opened, err := svc.isSectionOpenedByPlacement(context.Background(), 1, lowerSection.SectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement: %v", err)
	}
	if !opened {
		t.Fatal("expected true for lower-level section via effective level")
	}
}

// TestGrammarService_IsSectionOpenedByPlacement_NoPlacement returns false when user has no placement result.
func TestGrammarService_IsSectionOpenedByPlacement_NoPlacement(t *testing.T) {
	svc, contentRepo, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	if len(sectionsData.Sections) == 0 {
		t.Fatal("expected sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	opened, err := svc.isSectionOpenedByPlacement(context.Background(), 1, sectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement: %v", err)
	}
	if opened {
		t.Fatal("expected false when no placement result")
	}
}

// TestGrammarService_IsSectionOpenedByPlacement_DirectHit returns true when section ID is in placement OpenedSections.
func TestGrammarService_IsSectionOpenedByPlacement_DirectHit(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID
	if err := attemptRepo.SavePlacementTestResult(1, 50, 10, []string{sectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	opened, err := svc.isSectionOpenedByPlacement(context.Background(), 1, sectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement: %v", err)
	}
	if !opened {
		t.Fatal("expected true when section is in OpenedSections")
	}
}

// TestGrammarService_IsSectionOpenedByPlacement_GetSectionsError covers error when effective-level path calls GetSections and it fails.
func TestGrammarService_IsSectionOpenedByPlacement_GetSectionsError(t *testing.T) {
	_, _, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()
	if err := attemptRepo.SavePlacementTestResult(1, 50, 10, []string{"some-section"}); err != nil {
		t.Fatalf("SavePlacementTestResult: %v", err)
	}

	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	// Use same DB so placement result is visible; only content repo fails GetSections
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.isSectionOpenedByPlacement(context.Background(), 1, "other-section")
	if err == nil {
		t.Fatal("expected error when GetSections fails in effective-level path")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

// TestGrammarService_GetAllSectionsWithProgress_OverrideName covers title override from publish item name for published sections.
func TestGrammarService_GetAllSectionsWithProgress_OverrideName(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]
	customName := "All Sections Override"
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	if err := publishRepo.SetName("section", section.SectionID, &customName, nil); err != nil {
		t.Fatalf("SetName: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}

	all, err := svc.GetAllSectionsWithProgress(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAllSectionsWithProgress: %v", err)
	}
	var found bool
	for _, s := range all {
		if s.Section != nil && s.Section.SectionID == section.SectionID {
			found = true
			if s.Title != customName {
				t.Fatalf("expected title %q, got %q", customName, s.Title)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected to find section in result")
	}
}

// TestGrammarService_GetNextPublishedChapterID_NextAndIsLast covers next chapter and isLast.
func TestGrammarService_GetNextPublishedChapterID_NextAndIsLast(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) < 2 {
		t.Skip("need at least 2 chapters in section")
	}
	ch0, ch1 := section.ChapterIDs[0], section.ChapterIDs[1]
	for _, cid := range []string{ch0, ch1} {
		if err := publishRepo.SetPublished("chapter", cid, true, nil); err != nil {
			t.Fatalf("SetPublished chapter: %v", err)
		}
	}

	nextID, isLast, secID, err := svc.GetNextPublishedChapterID(context.Background(), ch0)
	if err != nil {
		t.Fatalf("GetNextPublishedChapterID: %v", err)
	}
	if nextID != ch1 {
		t.Fatalf("expected next chapter %q, got %q", ch1, nextID)
	}
	if isLast {
		t.Fatal("expected isLast false when next chapter exists")
	}
	if secID != section.SectionID {
		t.Fatalf("expected sectionID %q, got %q", section.SectionID, secID)
	}

	nextID2, isLast2, _, err := svc.GetNextPublishedChapterID(context.Background(), ch1)
	if err != nil {
		t.Fatalf("GetNextPublishedChapterID last: %v", err)
	}
	if nextID2 != "" {
		t.Fatalf("expected empty next ID for last chapter, got %q", nextID2)
	}
	if !isLast2 {
		t.Fatal("expected isLast true for last chapter")
	}
}

// TestGrammarService_GetChapterContent_ChapterNameOverride covers chapter title override from publish item.
func TestGrammarService_GetChapterContent_IsPublishedError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.GetChapterContent(context.Background(), "any-chapter", false)
	if err == nil {
		t.Fatal("expected error when IsPublished check fails")
	}
	if !strings.Contains(err.Error(), "failed to check published status") && !strings.Contains(err.Error(), "published") {
		t.Errorf("expected published status error, got: %v", err)
	}
}

func TestGrammarService_GetChapterContent_GetChapterError(t *testing.T) {
	// Content repo has sections but no index -> GetChapter fails after IsPublished succeeds
	logger := zap.NewNop()
	fs := fstest.MapFS{
		"sections.json": {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	if err := publishRepo.SetPublished("chapter", "ch1", true, nil); err != nil {
		t.Fatalf("SetPublished: %v", err)
	}

	_, err := svc.GetChapterContent(context.Background(), "ch1", false)
	if err == nil {
		t.Fatal("expected error when GetChapter fails (no index)")
	}
	if !strings.Contains(err.Error(), "failed to get chapter") && !strings.Contains(err.Error(), "chapter") {
		t.Errorf("expected get chapter error, got: %v", err)
	}
}

func TestGrammarService_GetChapterContent_ChapterNameOverride(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]
	if err := publishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished: %v", err)
	}
	customTitle := "Custom Chapter Title"
	if err := publishRepo.SetName("chapter", chapterID, &customTitle, nil); err != nil {
		t.Fatalf("SetName: %v", err)
	}

	content, err := svc.GetChapterContent(context.Background(), chapterID, true)
	if err != nil {
		t.Fatalf("GetChapterContent: %v", err)
	}
	if content.Title != customTitle {
		t.Fatalf("expected title %q, got %q", customTitle, content.Title)
	}
}

// TestGrammarService_GetGrammarStatistics_GetSectionsError covers error when GetSections fails.
func TestGrammarService_GetGrammarStatistics_GetSectionsError(t *testing.T) {
	logger := zap.NewNop()
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	db := testutil.SetupTestDatabase(t)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)

	_, err := svc.GetGrammarStatistics(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetSections fails")
	}
	if !strings.Contains(err.Error(), "failed to get sections") {
		t.Errorf("expected 'failed to get sections' in error, got: %v", err)
	}
}

// TestGrammarService_GetGrammarStatistics_GetPublishedItemsError covers error when GetPublishedItemsByType fails.
func TestGrammarService_GetGrammarStatistics_GetPublishedItemsError(t *testing.T) {
	svc, cleanup := grammarServiceWithClosedPublishRepo(t)
	defer cleanup()

	_, err := svc.GetGrammarStatistics(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when GetPublishedItemsByType fails")
	}
	if !strings.Contains(err.Error(), "failed to get published items") {
		t.Errorf("expected 'failed to get published items' in error, got: %v", err)
	}
}

// TestGrammarService_GetGrammarStatistics_PartialProgress sets confirmedLevel to A0 when user has progress but no section fully passed.
func TestGrammarService_GetGrammarStatistics_PartialProgress(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections: %v", err)
	}
	// Use a section with at least 2 chapters: pass one, not the other
	var section *repository.Section
	for i := range sectionsData.Sections {
		s := &sectionsData.Sections[i]
		if len(s.ChapterIDs) >= 2 {
			section = s
			break
		}
	}
	if section == nil {
		t.Skip("need section with at least 2 chapters")
	}
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section: %v", err)
	}
	ch0, ch1 := section.ChapterIDs[0], section.ChapterIDs[1]
	if err := publishRepo.SetPublished("chapter", ch0, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", ch1, true, nil); err != nil {
		t.Fatalf("SetPublished chapter: %v", err)
	}
	// Pass only first chapter
	if err := attemptRepo.UpdateProgress(1, ch0, 80, true); err != nil {
		t.Fatalf("UpdateProgress: %v", err)
	}
	// No progress on second chapter -> section not fully passed -> confirmedLevel stays "" then set to "A0"
	stats, err := svc.GetGrammarStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGrammarStatistics: %v", err)
	}
	if stats.ConfirmedLevel != "A0" {
		t.Fatalf("expected ConfirmedLevel A0 with partial progress, got %q", stats.ConfirmedLevel)
	}
}

func TestGrammarAttemptRepository_SavePlacement_ReplacesWhenAdminOverride(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	attemptRepo := repository.NewGrammarAttemptRepository(conn, zap.NewNop())
	userRepo := repository.NewUserRepository(conn, zap.NewNop())
	u, err := userRepo.GetOrCreateUser(990001)
	if err != nil {
		t.Fatal(err)
	}
	if err := attemptRepo.UpsertPlacementByAdmin(u.ID, 0, 0, []string{"sec-x"}); err != nil {
		t.Fatal(err)
	}
	if err := attemptRepo.SavePlacementTestResult(u.ID, 30, 5, []string{"sec-y"}); err != nil {
		t.Fatal(err)
	}
	res, err := attemptRepo.GetPlacementTestResult(u.ID)
	if err != nil || res == nil {
		t.Fatalf("GetPlacementTestResult: %v, %v", res, err)
	}
	if res.Score != 30 {
		t.Fatalf("score want 30 got %d", res.Score)
	}
	if len(res.OpenedSections) != 1 || res.OpenedSections[0] != "sec-y" {
		t.Fatalf("opened sections: %+v", res.OpenedSections)
	}
	if res.AdminOverride {
		t.Fatal("admin_override should be false after user placement save")
	}
}

func TestGrammarAttemptRepository_SavePlacement_KeepsHigherScoreWithoutAdmin(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	attemptRepo := repository.NewGrammarAttemptRepository(conn, zap.NewNop())
	userRepo := repository.NewUserRepository(conn, zap.NewNop())
	u, err := userRepo.GetOrCreateUser(990002)
	if err != nil {
		t.Fatal(err)
	}
	if err := attemptRepo.SavePlacementTestResult(u.ID, 80, 10, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	if err := attemptRepo.SavePlacementTestResult(u.ID, 50, 10, []string{"b"}); err != nil {
		t.Fatal(err)
	}
	res, err := attemptRepo.GetPlacementTestResult(u.ID)
	if err != nil || res == nil {
		t.Fatal(err)
	}
	if res.Score != 80 {
		t.Fatalf("score want 80 got %d", res.Score)
	}
	if len(res.OpenedSections) != 1 || res.OpenedSections[0] != "a" {
		t.Fatalf("opened: %+v", res.OpenedSections)
	}
}
