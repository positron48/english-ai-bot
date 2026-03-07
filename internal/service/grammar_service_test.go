package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
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

	service := NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)

	cleanup := func() {} // shared db, do not close

	return service, contentRepo, publishRepo, attemptRepo, userRepo, cleanup
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
