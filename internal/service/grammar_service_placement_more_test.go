package service

import (
	"context"
	"strings"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func pickSectionsWithQuestionChapters(t *testing.T, contentRepo *repository.GrammarContentRepository, need int) []repository.Section {
	t.Helper()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections error: %v", err)
	}

	picked := make([]repository.Section, 0, need)
	for _, section := range sectionsData.Sections {
		if len(section.ChapterIDs) == 0 {
			continue
		}
		chapter, err := contentRepo.GetChapter(section.ChapterIDs[0])
		if err != nil {
			continue
		}
		questions, ok := chapter.QuestionBank["questions"].([]interface{})
		if !ok || len(questions) == 0 {
			continue
		}
		picked = append(picked, section)
		if len(picked) == need {
			break
		}
	}

	if len(picked) < need {
		t.Fatalf("expected at least %d sections with question chapters, got %d", need, len(picked))
	}
	return picked
}

func publishSectionFirstChapter(t *testing.T, publishRepo *repository.GrammarPublishRepository, section repository.Section) {
	t.Helper()
	if err := publishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("SetPublished section error: %v", err)
	}
	if err := publishRepo.SetPublished("chapter", section.ChapterIDs[0], true, nil); err != nil {
		t.Fatalf("SetPublished chapter error: %v", err)
	}
}

func chapterAnswersByID(t *testing.T, contentRepo *repository.GrammarContentRepository, chapterID string) map[string]interface{} {
	t.Helper()
	chapter, err := contentRepo.GetChapter(chapterID)
	if err != nil {
		t.Fatalf("GetChapter error: %v", err)
	}
	questions, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok {
		t.Fatalf("chapter %s has invalid question bank", chapterID)
	}
	out := make(map[string]interface{}, len(questions))
	for _, raw := range questions {
		qMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		qid, _ := qMap["id"].(string)
		if qid == "" {
			continue
		}
		out[qid] = qMap["correct_answer"]
	}
	return out
}

func wrongAnswerFor(correct interface{}) interface{} {
	switch v := correct.(type) {
	case bool:
		return !v
	case string:
		return "__wrong__"
	case float64:
		return v + 1
	case int:
		return v + 1
	case []interface{}:
		if len(v) == 0 {
			return []interface{}{"__wrong__"}
		}
		return []interface{}{v[0], "__wrong__"}
	case map[string]interface{}:
		return map[string]interface{}{"__wrong__": true}
	default:
		return "__wrong__"
	}
}

func levelOrder(level string) int {
	order := map[string]int{
		"A0": 0,
		"A1": 1,
		"A2": 2,
		"B1": 3,
		"B2": 4,
		"C1": 5,
		"C2": 6,
	}
	if v, ok := order[level]; ok {
		return v
	}
	return -1
}

func TestGrammarService_GeneratePlacementTest_SanitizesAndKeepsMetadata(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections := pickSectionsWithQuestionChapters(t, contentRepo, 3)
	publishedSectionSet := make(map[string]bool)
	publishedChapterSet := make(map[string]bool)
	for _, section := range sections {
		publishSectionFirstChapter(t, publishRepo, section)
		publishedSectionSet[section.SectionID] = true
		publishedChapterSet[section.ChapterIDs[0]] = true
	}

	placement, err := svc.GeneratePlacementTest(context.Background())
	if err != nil {
		t.Fatalf("GeneratePlacementTest error: %v", err)
	}
	if placement.Total == 0 {
		t.Fatal("expected placement questions")
	}
	if placement.Total > 25 {
		t.Fatalf("expected at most 25 questions, got %d", placement.Total)
	}

	seenSections := make(map[string]int)
	for _, raw := range placement.Questions {
		qMap, ok := raw.(map[string]interface{})
		if !ok {
			t.Fatalf("expected question map, got %T", raw)
		}
		qID, _ := qMap["id"].(string)
		if !strings.Contains(qID, ":") {
			t.Fatalf("expected composite question id with chapter prefix, got %q", qID)
		}

		chapterID, _ := qMap["placement_chapter_id"].(string)
		if !publishedChapterSet[chapterID] {
			t.Fatalf("question from unpublished chapter: %s", chapterID)
		}

		sectionID, _ := qMap["placement_section_id"].(string)
		if !publishedSectionSet[sectionID] {
			t.Fatalf("question from unpublished section: %s", sectionID)
		}
		seenSections[sectionID]++

		if _, ok := qMap["placement_section_order"]; !ok {
			t.Fatal("expected placement_section_order metadata")
		}
		if _, ok := qMap["placement_chapter_order"]; !ok {
			t.Fatal("expected placement_chapter_order metadata")
		}

		qType, _ := qMap["type"].(string)
		if qType != "reorder" {
			if _, ok := qMap["correct_answer"]; ok {
				t.Fatalf("expected correct_answer removed for %s", qType)
			}
		}
	}

	for _, section := range sections {
		if seenSections[section.SectionID] == 0 {
			t.Fatalf("expected at least one question from section %s", section.SectionID)
		}
	}
}

func TestGrammarService_SubmitPlacementTest_LevelExpansionAndBelowA1(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	t.Run("expands opened sections by level", func(t *testing.T) {
		sections := pickSectionsWithQuestionChapters(t, contentRepo, 4)
		for _, section := range sections {
			publishSectionFirstChapter(t, publishRepo, section)
		}

		placement, err := svc.GeneratePlacementTest(context.Background())
		if err != nil {
			t.Fatalf("GeneratePlacementTest error: %v", err)
		}
		if placement.Total == 0 {
			t.Fatal("expected placement questions")
		}

		chapterAnswerCache := make(map[string]map[string]interface{})
		answers := make(map[string]interface{}, len(placement.Questions))
		passSection := map[string]bool{
			sections[0].SectionID: true,
			sections[1].SectionID: true,
		}

		for _, raw := range placement.Questions {
			qMap, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			compID, _ := qMap["id"].(string)
			chapterID, _ := qMap["placement_chapter_id"].(string)
			sectionID, _ := qMap["placement_section_id"].(string)
			parts := strings.SplitN(compID, ":", 2)
			if len(parts) != 2 {
				continue
			}
			qid := parts[1]

			if _, ok := chapterAnswerCache[chapterID]; !ok {
				chapterAnswerCache[chapterID] = chapterAnswersByID(t, contentRepo, chapterID)
			}
			correct, ok := chapterAnswerCache[chapterID][qid]
			if !ok {
				continue
			}

			if passSection[sectionID] {
				answers[compID] = correct
			} else {
				answers[compID] = wrongAnswerFor(correct)
			}
		}

		result, err := svc.SubmitPlacementTest(context.Background(), 1, answers)
		if err != nil {
			t.Fatalf("SubmitPlacementTest error: %v", err)
		}
		if result.TotalQuestions == 0 {
			t.Fatal("expected scored questions")
		}
		if result.Level != sections[1].Level {
			t.Fatalf("expected level %s, got %s", sections[1].Level, result.Level)
		}

		expectedOpened := make([]string, 0)
		maxLevelOrder := levelOrder(result.Level)
		for _, section := range sections {
			if levelOrder(section.Level) <= maxLevelOrder {
				expectedOpened = append(expectedOpened, section.SectionID)
			}
		}
		if len(result.OpenedSections) != len(expectedOpened) {
			t.Fatalf("expected %d opened sections, got %d", len(expectedOpened), len(result.OpenedSections))
		}
		for i, sectionID := range expectedOpened {
			if result.OpenedSections[i] != sectionID {
				t.Fatalf("expected opened section %s at index %d, got %s", sectionID, i, result.OpenedSections[i])
			}
		}

		storedPlacement, err := attemptRepo.GetPlacementTestResult(1)
		if err != nil {
			t.Fatalf("GetPlacementTestResult error: %v", err)
		}
		if storedPlacement == nil {
			t.Fatal("expected stored placement result")
		}
		if storedPlacement.Score != result.Score {
			t.Fatalf("expected stored score %d, got %d", result.Score, storedPlacement.Score)
		}

		attempts, err := attemptRepo.GetUserAttempts(1, "placement", "placement", 5)
		if err != nil {
			t.Fatalf("GetUserAttempts error: %v", err)
		}
		if len(attempts) == 0 {
			t.Fatal("expected placement attempt to be stored")
		}
	})

	t.Run("below A1 when no valid answers", func(t *testing.T) {
		sections := pickSectionsWithQuestionChapters(t, contentRepo, 1)
		publishSectionFirstChapter(t, publishRepo, sections[0])

		result, err := svc.SubmitPlacementTest(context.Background(), 1, map[string]interface{}{
			"missing:question": "anything",
		})
		if err != nil {
			t.Fatalf("SubmitPlacementTest error: %v", err)
		}
		if result.Level != "Below A1" {
			t.Fatalf("expected level Below A1, got %s", result.Level)
		}
		if result.Score != 0 {
			t.Fatalf("expected score 0, got %d", result.Score)
		}
		if len(result.OpenedSections) != 0 {
			t.Fatalf("expected no opened sections, got %d", len(result.OpenedSections))
		}
	})
}

func TestGrammarService_GetNextPublishedChapterID(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sectionsData, err := contentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections error: %v", err)
	}

	var section repository.Section
	found := false
	for _, s := range sectionsData.Sections {
		if len(s.ChapterIDs) >= 3 {
			section = s
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected a section with at least 3 chapters")
	}

	if err := publishRepo.SetPublished("chapter", section.ChapterIDs[2], true, nil); err != nil {
		t.Fatalf("SetPublished chapter error: %v", err)
	}

	next, isLast, gotSectionID, err := svc.GetNextPublishedChapterID(context.Background(), section.ChapterIDs[0])
	if err != nil {
		t.Fatalf("GetNextPublishedChapterID error: %v", err)
	}
	if gotSectionID != section.SectionID {
		t.Fatalf("expected section %s, got %s", section.SectionID, gotSectionID)
	}
	if isLast {
		t.Fatal("expected not last chapter")
	}
	if next != section.ChapterIDs[2] {
		t.Fatalf("expected next published chapter %s, got %s", section.ChapterIDs[2], next)
	}

	next, isLast, _, err = svc.GetNextPublishedChapterID(context.Background(), section.ChapterIDs[2])
	if err != nil {
		t.Fatalf("GetNextPublishedChapterID error: %v", err)
	}
	if !isLast {
		t.Fatal("expected last chapter")
	}
	if next != "" {
		t.Fatalf("expected empty next chapter for last, got %s", next)
	}

	_, _, _, err = svc.GetNextPublishedChapterID(context.Background(), "missing.chapter.id")
	if err == nil {
		t.Fatal("expected error for missing chapter id")
	}
}

func TestGrammarService_IsSectionOpenedByPlacement_EffectiveLevel(t *testing.T) {
	svc, contentRepo, _, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections := pickSectionsWithQuestionChapters(t, contentRepo, 5)
	if levelOrder(sections[0].Level) >= levelOrder(sections[4].Level) {
		t.Fatalf("expected later section to have higher level, got %s and %s", sections[0].Level, sections[4].Level)
	}

	if err := attemptRepo.SavePlacementTestResult(1, 80, 10, []string{sections[1].SectionID}); err != nil {
		t.Fatalf("SavePlacementTestResult error: %v", err)
	}

	isOpened, err := svc.isSectionOpenedByPlacement(context.Background(), 1, sections[1].SectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement error: %v", err)
	}
	if !isOpened {
		t.Fatal("expected explicitly opened section to be accessible")
	}

	isOpened, err = svc.isSectionOpenedByPlacement(context.Background(), 1, sections[0].SectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement error: %v", err)
	}
	if !isOpened {
		t.Fatal("expected lower-level section to be opened via effective level")
	}

	isOpened, err = svc.isSectionOpenedByPlacement(context.Background(), 1, sections[4].SectionID)
	if err != nil {
		t.Fatalf("isSectionOpenedByPlacement error: %v", err)
	}
	if isOpened {
		t.Fatal("expected higher-level section to remain locked")
	}
}

func TestGrammarService_CanAccessSection_FallbackAllPublishedChaptersPassed(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections := pickSectionsWithQuestionChapters(t, contentRepo, 2)
	prev := sections[0]
	next := sections[1]

	publishSectionFirstChapter(t, publishRepo, prev)
	if err := publishRepo.SetPublished("chapter", prev.ChapterIDs[1], true, nil); err != nil {
		t.Fatalf("SetPublished chapter error: %v", err)
	}

	if err := attemptRepo.UpdateProgress(1, prev.ChapterIDs[0], 80, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}
	if err := attemptRepo.UpdateProgress(1, prev.ChapterIDs[1], 70, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	canAccess, err := svc.CanAccessSection(context.Background(), 1, next.SectionID)
	if err != nil {
		t.Fatalf("CanAccessSection error: %v", err)
	}
	if !canAccess {
		t.Fatal("expected next section unlocked by fallback all-published-chapters-passed")
	}
}
