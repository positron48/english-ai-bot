package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// GrammarService handles grammar course business logic
type GrammarService struct {
	ContentRepo      *repository.GrammarContentRepository
	PublishRepo      *repository.GrammarPublishRepository
	AttemptRepo      *repository.GrammarAttemptRepository
	TrainingPackRepo *repository.GrammarTrainingPackRepository
	SRSRepo          *repository.GrammarSRSRepository
	TheoryIndex      *repository.TheoryBlockIndex
	learning         config.LearningConfig
	logger           *zap.Logger
}

// NewGrammarService creates a new grammar service
func NewGrammarService(
	contentRepo *repository.GrammarContentRepository,
	publishRepo *repository.GrammarPublishRepository,
	attemptRepo *repository.GrammarAttemptRepository,
	learning config.LearningConfig,
	logger *zap.Logger,
) *GrammarService {
	bundleID := strings.ToLower(strings.TrimSpace(learning.GrammarBundleID))
	if bundleID == "" {
		bundleID = strings.ToLower(strings.TrimSpace(learning.TargetLang))
	}
	nativeLang := strings.ToLower(strings.TrimSpace(learning.NativeLang))
	if bundleID != "" && nativeLang != "" {
		attemptRepo = attemptRepo.ForCourse(bundleID + "_" + nativeLang)
	}
	theoryIndex, err := repository.BuildTheoryBlockIndex(contentRepo)
	if err != nil && logger != nil {
		logger.Warn("failed to build theory block index", zap.Error(err))
	}
	return &GrammarService{
		ContentRepo: contentRepo,
		PublishRepo: publishRepo,
		AttemptRepo: attemptRepo,
		TheoryIndex: theoryIndex,
		learning:    learning,
		logger:      logger,
	}
}

func (s *GrammarService) SetTrainingPackRepository(repo *repository.GrammarTrainingPackRepository) {
	s.TrainingPackRepo = repo
}

func (s *GrammarService) SetSRSRepository(repo *repository.GrammarSRSRepository) {
	s.SRSRepo = repo
}

// GetPublishedSections returns published sections with progress
func (s *GrammarService) GetPublishedSections(ctx context.Context, userID int64) ([]*SectionWithProgress, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("section")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	var result []*SectionWithProgress
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		item, exists := publishedItems[section.SectionID]
		if !exists || !item.IsPublished {
			continue // Skip unpublished sections
		}

		// Get override name if exists
		title := section.Title
		if item.Name != nil && *item.Name != "" {
			title = *item.Name
		}

		// Count published chapters in this section and calculate progress percentage
		publishedChapters := 0
		passedChapters := 0
		totalScore := 0
		for _, chapterID := range section.ChapterIDs {
			chapterItem, _ := s.PublishRepo.GetPublishedItem("chapter", chapterID)
			if chapterItem.IsPublished {
				publishedChapters++
				progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
				if progress.Passed {
					passedChapters++
				}
				// Add best_score to total for percentage calculation
				totalScore += progress.BestScore
			}
		}

		// Calculate average percentage (sum of chapter percentages / number of chapters)
		progressPercentage := 0
		if publishedChapters > 0 {
			progressPercentage = totalScore / publishedChapters
		}

		result = append(result, &SectionWithProgress{
			Section:            section,
			Title:              title,
			IsPublished:        true,
			PublishedChapters:  publishedChapters,
			PassedChapters:     passedChapters,
			ProgressPercentage: progressPercentage,
		})
	}

	return result, nil
}

// SectionWithProgress represents a section with user progress
type SectionWithProgress struct {
	Section            *repository.Section
	Title              string
	IsPublished        bool
	PublishedChapters  int
	PassedChapters     int
	ProgressPercentage int // Average percentage based on chapter best_scores
}

// GetAllSectionsWithProgress returns all sections (published and unpublished) with progress and IsPublished flag
func (s *GrammarService) GetAllSectionsWithProgress(ctx context.Context, userID int64) ([]*SectionWithProgress, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("section")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	var result []*SectionWithProgress
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		item, exists := publishedItems[section.SectionID]
		isPublished := exists && item.IsPublished

		title := section.Title
		if isPublished && item.Name != nil && *item.Name != "" {
			title = *item.Name
		}

		publishedChapters := 0
		passedChapters := 0
		totalScore := 0
		for _, chapterID := range section.ChapterIDs {
			chapterItem, _ := s.PublishRepo.GetPublishedItem("chapter", chapterID)
			if chapterItem.IsPublished {
				publishedChapters++
				progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
				if progress.Passed {
					passedChapters++
				}
				totalScore += progress.BestScore
			}
		}

		progressPercentage := 0
		if publishedChapters > 0 {
			progressPercentage = totalScore / publishedChapters
		}

		result = append(result, &SectionWithProgress{
			Section:            section,
			Title:              title,
			IsPublished:        isPublished,
			PublishedChapters:  publishedChapters,
			PassedChapters:     passedChapters,
			ProgressPercentage: progressPercentage,
		})
	}

	return result, nil
}

// IsSectionPublished returns whether the section is published
func (s *GrammarService) IsSectionPublished(ctx context.Context, sectionID string) (bool, error) {
	item, err := s.PublishRepo.GetPublishedItem("section", sectionID)
	if err != nil || item == nil {
		return false, err
	}
	return item.IsPublished, nil
}

// GetPublishedChapters returns published chapters for a section
func (s *GrammarService) GetPublishedChapters(ctx context.Context, sectionID string, userID int64) ([]*ChapterWithProgress, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	var section *repository.Section
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			section = &sectionsData.Sections[i]
			break
		}
	}

	if section == nil {
		return nil, fmt.Errorf("section not found: %s", sectionID)
	}

	published, _ := s.IsSectionPublished(ctx, sectionID)
	if !published {
		return nil, fmt.Errorf("section not published: %s", sectionID)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	var result []*ChapterWithProgress
	for _, chapterID := range section.ChapterIDs {
		item, exists := publishedItems[chapterID]
		if !exists || !item.IsPublished {
			continue // Skip unpublished chapters
		}

		chapter, err := s.ContentRepo.GetChapter(chapterID)
		if err != nil {
			s.logger.Warn("failed to load chapter", zap.String("chapter_id", chapterID), zap.Error(err))
			continue
		}

		// Get override name if exists
		title := chapter.Title
		if item.Name != nil && *item.Name != "" {
			title = *item.Name
		}

		progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)

		result = append(result, &ChapterWithProgress{
			Chapter:  chapter,
			Title:    title,
			Progress: progress,
		})
	}

	sectionAccess, _ := s.CanAccessSection(ctx, userID, sectionID)
	// Check if section was opened by placement test (in this case, all chapters are accessible)
	isOpenedByPlacement, _ := s.isSectionOpenedByPlacement(ctx, userID, sectionID)

	for i := range result {
		if isOpenedByPlacement {
			// If opened by placement test, all chapters are accessible
			result[i].CanAccess = true
		} else if i == 0 {
			// First chapter is always accessible if section is accessible
			result[i].CanAccess = sectionAccess
		} else {
			// Other chapters are accessible only if previous chapter was passed
			result[i].CanAccess = result[i-1].Progress.Passed
		}
	}

	return result, nil
}

// ContinueChapter is the grammar chapter the user should resume on home quick access.
type ContinueChapter struct {
	ChapterID         string            `json:"chapter_id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	SectionID         string            `json:"section_id"`
}

// GetContinueChapter returns the chapter the user should continue studying.
// Priority: (1) first accessible unpassed chapter that needs study,
// (2) most recently attempted accessible chapter unless placement unlocked higher levels,
// (3) furthest accessible chapter in canonical course order.
func (s *GrammarService) GetContinueChapter(ctx context.Context, userID int64) (*ContinueChapter, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	placementLevelOrder := s.placementEffectiveLevelOrder(userID, sectionsData)

	var (
		studyTarget       *ContinueChapter
		byRecentAttempt   *ContinueChapter
		recentAttemptTime time.Time
		frontier          *ContinueChapter
	)

	for si := range sectionsData.Sections {
		section := &sectionsData.Sections[si]
		sectionAccess, err := s.CanAccessSection(ctx, userID, section.SectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check section access: %w", err)
		}
		if !sectionAccess {
			break
		}

		isOpenedByPlacement, err := s.isSectionOpenedByPlacement(ctx, userID, section.SectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check placement section: %w", err)
		}

		chapters, err := s.GetPublishedChapters(ctx, section.SectionID, userID)
		if err != nil {
			if strings.Contains(err.Error(), "not published") {
				continue
			}
			return nil, fmt.Errorf("failed to get chapters for section %s: %w", section.SectionID, err)
		}

		for _, chapter := range chapters {
			canAccess := chapter.CanAccess
			passed := chapter.Progress.Passed
			if !canAccess && !passed {
				break
			}

			cc := s.continueChapterFromProgress(chapter, section.SectionID)
			frontier = cc

			hasAttempt := !chapter.Progress.LastAttemptAt.IsZero()
			if canAccess && hasAttempt {
				if byRecentAttempt == nil || chapter.Progress.LastAttemptAt.After(recentAttemptTime) {
					byRecentAttempt = cc
					recentAttemptTime = chapter.Progress.LastAttemptAt
				}
			}

			if studyTarget == nil && canAccess && !passed {
				if isOpenedByPlacement && !hasAttempt {
					continue
				}
				sectionLevelOrder := s.sectionLevelOrder(sectionsData, section.SectionID)
				if isOpenedByPlacement && placementLevelOrder >= 0 && sectionLevelOrder >= 0 && placementLevelOrder > sectionLevelOrder {
					continue
				}
				studyTarget = cc
			}
		}
	}

	if studyTarget != nil {
		return studyTarget, nil
	}
	if byRecentAttempt != nil {
		attemptLevelOrder := s.sectionLevelOrder(sectionsData, byRecentAttempt.SectionID)
		if placementLevelOrder >= 0 && attemptLevelOrder >= 0 && placementLevelOrder > attemptLevelOrder {
			if frontier != nil {
				return frontier, nil
			}
		}
		return byRecentAttempt, nil
	}
	if frontier != nil {
		return frontier, nil
	}
	return nil, nil
}

func (s *GrammarService) continueChapterFromProgress(chapter *ChapterWithProgress, sectionID string) *ContinueChapter {
	title := chapter.Title
	if title == "" && chapter.Chapter != nil {
		title = chapter.Chapter.Title
	}
	var titleTranslations map[string]string
	if chapter.Chapter != nil {
		titleTranslations = chapter.Chapter.TitleTranslations
	}
	chapterID := ""
	if chapter.Chapter != nil {
		chapterID = chapter.Chapter.ID
	}
	return &ContinueChapter{
		ChapterID:         chapterID,
		Title:             title,
		TitleTranslations: titleTranslations,
		SectionID:         sectionID,
	}
}

func (s *GrammarService) sectionLevelOrder(sectionsData *repository.SectionsData, sectionID string) int {
	levelOrder := grammarLevelOrderMap()
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID != sectionID {
			continue
		}
		if ord, ok := levelOrder[sectionsData.Sections[i].Level]; ok {
			return ord
		}
		return -1
	}
	return -1
}

func (s *GrammarService) placementEffectiveLevelOrder(userID int64, sectionsData *repository.SectionsData) int {
	placementResult, _ := s.AttemptRepo.GetPlacementTestResult(userID)
	if placementResult == nil || len(placementResult.OpenedSections) == 0 {
		return -1
	}
	levelOrder := grammarLevelOrderMap()
	effectiveOrder := -1
	for _, openedSectionID := range placementResult.OpenedSections {
		for j := range sectionsData.Sections {
			if sectionsData.Sections[j].SectionID != openedSectionID {
				continue
			}
			if ord, ok := levelOrder[sectionsData.Sections[j].Level]; ok && ord >= 0 && ord > effectiveOrder {
				effectiveOrder = ord
			}
			break
		}
	}
	return effectiveOrder
}

// GetNextPublishedChapterID returns the next published chapter within the same section,
// using the canonical order from sections.json (section.ChapterIDs).
func (s *GrammarService) GetNextPublishedChapterID(ctx context.Context, chapterID string) (nextChapterID string, isLast bool, sectionID string, err error) {
	chapter, err := s.ContentRepo.GetChapter(chapterID)
	if err != nil {
		return "", false, "", fmt.Errorf("chapter not found: %s", chapterID)
	}

	sectionID = chapter.SectionID
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return "", false, "", fmt.Errorf("failed to get sections: %w", err)
	}

	var section *repository.Section
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			section = &sectionsData.Sections[i]
			break
		}
	}
	if section == nil {
		return "", false, "", fmt.Errorf("section not found: %s", sectionID)
	}

	// Use canonical order from sections.json (section.ChapterIDs) and then walk forward
	// to the next published chapter. This avoids any "backwards" jumps even if publish data
	// is incomplete or chapter IDs contain dots, etc.
	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return "", false, "", fmt.Errorf("failed to get published items: %w", err)
	}

	// Find the chapter index in the canonical list.
	// If there are duplicates (shouldn't happen), we use the last occurrence to ensure we move forward.
	idx := -1
	for i := range section.ChapterIDs {
		if section.ChapterIDs[i] == chapterID {
			idx = i
		}
	}
	if idx == -1 {
		return "", false, sectionID, fmt.Errorf("chapter not found in section list: %s", chapterID)
	}

	// Walk forward to the next published chapter
	for i := idx + 1; i < len(section.ChapterIDs); i++ {
		id := section.ChapterIDs[i]
		item, exists := publishedItems[id]
		if !exists || !item.IsPublished {
			continue
		}
		return id, false, sectionID, nil
	}

	return "", true, sectionID, nil
}

// ChapterWithProgress represents a chapter with user progress
type ChapterWithProgress struct {
	Chapter   *repository.Chapter
	Title     string
	Progress  *repository.ChapterProgress
	CanAccess bool // true if section is unlocked (placement/previous) or first/prev passed
}

// GetChapterContent returns chapter content (without answers for tests)
func (s *GrammarService) GetChapterContent(ctx context.Context, chapterID string, includeAnswers bool) (*ChapterContent, error) {
	// Check if chapter is published
	isPublished, err := s.PublishRepo.IsPublished("chapter", chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check published status: %w", err)
	}
	if !isPublished {
		return nil, fmt.Errorf("chapter not found or not published: %s", chapterID)
	}

	chapter, err := s.ContentRepo.GetChapter(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	// Get override name if exists
	item, _ := s.PublishRepo.GetPublishedItem("chapter", chapterID)
	title := chapter.Title
	if item.Name != nil && *item.Name != "" {
		title = *item.Name
	}

	content := &ChapterContent{
		Chapter: chapter,
		Title:   title,
	}

	if !includeAnswers {
		// Remove correct_answer from questions for inline quizzes (we can show explanation)
		// For chapter tests, we'll handle this in GenerateChapterTest
		content.Chapter = s.sanitizeChapterForDisplay(chapter)
	}

	// Filter question bank to only include questions used in inline quizzes
	content.Chapter = s.filterQuestionBankForQuizzes(content.Chapter)

	return content, nil
}

// GetSectionBySectionID returns section metadata from the bundle (categories list).
func (s *GrammarService) GetSectionBySectionID(_ context.Context, sectionID string) (*repository.Section, error) {
	if strings.TrimSpace(sectionID) == "" {
		return nil, fmt.Errorf("section_id empty")
	}
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, err
	}
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			sec := sectionsData.Sections[i]
			return &sec, nil
		}
	}
	return nil, fmt.Errorf("section not found: %s", sectionID)
}

// ChapterContent represents chapter content for display
type ChapterContent struct {
	Chapter *repository.Chapter
	Title   string
}

// sanitizeChapterForDisplay removes sensitive data but keeps enough for inline quiz feedback
func (s *GrammarService) sanitizeChapterForDisplay(chapter *repository.Chapter) *repository.Chapter {
	// For now, return as-is for inline quizzes (they can show immediate feedback)
	// Chapter tests will be handled separately
	return chapter
}

// filterQuestionBankForQuizzes filters question bank to only include questions used in inline quizzes
func (s *GrammarService) filterQuestionBankForQuizzes(chapter *repository.Chapter) *repository.Chapter {
	// Create a copy to avoid modifying the original
	filteredChapter := *chapter

	// Collect all question IDs used in inline quizzes
	usedQuestionIDs := make(map[string]bool)

	// Check blocks for quiz_inline types
	blocks := chapter.Blocks
	for _, blockInterface := range blocks {
		block, ok := blockInterface.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)
		if blockType == "quiz_inline" {
			quizInline, ok := block["quiz_inline"].(map[string]interface{})
			if !ok {
				continue
			}

			questionIDs, ok := quizInline["question_ids"].([]interface{})
			if !ok {
				continue
			}

			for _, idInterface := range questionIDs {
				if id, ok := idInterface.(string); ok {
					usedQuestionIDs[id] = true
				}
			}
		}
	}

	// Filter question bank if it exists
	questionBank := make(map[string]interface{}, len(chapter.QuestionBank))
	for key, value := range chapter.QuestionBank {
		questionBank[key] = value
	}
	availableIDs := make(map[string]bool)
	if questionBank != nil {
		if questions, ok := questionBank["questions"].([]interface{}); ok {
			filteredQuestions := make([]interface{}, 0)

			for _, qInterface := range questions {
				q, ok := qInterface.(map[string]interface{})
				if !ok {
					continue
				}

				qID, ok := q["id"].(string)
				if !ok {
					continue
				}

				// Only include questions that are used in quizzes
				if usedQuestionIDs[qID] && repository.GrammarQuestionAvailable(q) {
					filteredQuestions = append(filteredQuestions, q)
					availableIDs[qID] = true
				}
			}

			// Update question bank with filtered questions
			questionBank["questions"] = filteredQuestions
			filteredChapter.QuestionBank = questionBank
		}
	}

	// Prune quiz references too, including quizzes left empty after filtering.
	filteredChapter.Blocks = make([]interface{}, 0, len(chapter.Blocks))
	for _, raw := range chapter.Blocks {
		block, ok := raw.(map[string]interface{})
		if !ok || block["type"] != "quiz_inline" {
			filteredChapter.Blocks = append(filteredChapter.Blocks, raw)
			continue
		}
		quiz, _ := block["quiz_inline"].(map[string]interface{})
		ids, _ := quiz["question_ids"].([]interface{})
		keptIDs := make([]interface{}, 0, len(ids))
		for _, rawID := range ids {
			if id, ok := rawID.(string); ok && availableIDs[id] {
				keptIDs = append(keptIDs, id)
			}
		}
		if len(keptIDs) == 0 {
			continue
		}
		blockCopy := make(map[string]interface{}, len(block))
		for key, value := range block {
			blockCopy[key] = value
		}
		quizCopy := make(map[string]interface{}, len(quiz))
		for key, value := range quiz {
			quizCopy[key] = value
		}
		quizCopy["question_ids"] = keptIDs
		blockCopy["quiz_inline"] = quizCopy
		filteredChapter.Blocks = append(filteredChapter.Blocks, blockCopy)
	}
	return &filteredChapter
}

// GenerateChapterTest generates a test from chapter's question bank
func (s *GrammarService) GenerateChapterTest(ctx context.Context, chapterID string) (*TestQuestions, error) {
	chapter, err := s.ContentRepo.GetChapter(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	// Extract test config
	testConfig, ok := chapter.ChapterTest["selection_strategy"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid test config")
	}

	numQuestions := 10
	if n, ok := chapter.ChapterTest["num_questions"].(float64); ok {
		numQuestions = int(n)
	}

	poolIDs, ok := chapter.ChapterTest["pool_question_ids"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid pool_question_ids")
	}

	// Get question bank
	questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid question bank")
	}

	// Build question map
	questionMap := make(map[string]interface{})
	for _, q := range questionBank {
		qMap, ok := q.(map[string]interface{})
		if !ok || !repository.GrammarQuestionAvailable(qMap) {
			continue
		}
		if id, ok := qMap["id"].(string); ok {
			questionMap[id] = q
		}
	}

	// Select questions based on strategy
	selectedQuestions := s.selectQuestions(poolIDs, questionMap, testConfig, numQuestions)

	// Remove correct_answer from selected questions (except for reorder type, which needs it for word display)
	for _, q := range selectedQuestions {
		if qMap, ok := q.(map[string]interface{}); ok {
			// Keep correct_answer for reorder questions - needed to split into words for UI
			questionType, _ := qMap["type"].(string)
			if questionType != "reorder" {
				delete(qMap, "correct_answer")
			}
		}
	}

	return &TestQuestions{
		Questions: selectedQuestions,
		Total:     len(selectedQuestions),
	}, nil
}

// GenerateCategoryTest generates a test from all chapters in a category
// It selects at least 2 questions from each chapter, then fills up to 20 questions randomly
func (s *GrammarService) GenerateCategoryTest(ctx context.Context, sectionID string) (*TestQuestions, error) {
	// Get sections to find the section
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	// Find the section
	var section *repository.Section
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			section = &sectionsData.Sections[i]
			break
		}
	}

	if section == nil {
		return nil, fmt.Errorf("section not found: %s", sectionID)
	}

	// Get published items to filter chapters
	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Collect questions from all published chapters in the section
	type chapterQuestions struct {
		chapterID   string
		questions   []interface{}
		questionMap map[string]interface{}
	}

	var allChapterQuestions []chapterQuestions
	minQuestionsPerChapter := 2
	targetTotalQuestions := 20

	// Collect questions from each chapter
	for _, chapterID := range section.ChapterIDs {
		chapterItem, exists := publishedItems[chapterID]
		if !exists || !chapterItem.IsPublished {
			continue // Skip unpublished chapters
		}

		chapter, err := s.ContentRepo.GetChapter(chapterID)
		if err != nil {
			s.logger.Warn("failed to get chapter for category test", zap.String("chapter_id", chapterID), zap.Error(err))
			continue
		}

		// Get question bank
		questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
		if !ok || len(questionBank) == 0 {
			continue // Skip chapters without questions
		}

		// Get pool_question_ids from chapter_test if available
		var poolIDs []interface{}
		if chapterTest, ok := chapter.ChapterTest["pool_question_ids"].([]interface{}); ok {
			poolIDs = chapterTest
		} else {
			// Fallback: use all questions from question bank
			for _, q := range questionBank {
				if qMap, ok := q.(map[string]interface{}); ok {
					if id, ok := qMap["id"].(string); ok {
						poolIDs = append(poolIDs, id)
					}
				}
			}
		}

		// Build question map for this chapter
		questionMap := make(map[string]interface{})
		for _, q := range questionBank {
			qMap, ok := q.(map[string]interface{})
			if !ok || !repository.GrammarQuestionAvailable(qMap) {
				continue
			}
			if id, ok := qMap["id"].(string); ok {
				questionMap[id] = q
			}
		}

		// Filter poolIDs to only include questions that exist in questionMap
		validPoolIDs := make([]interface{}, 0)
		for _, idInterface := range poolIDs {
			id, ok := idInterface.(string)
			if !ok {
				continue
			}
			if _, exists := questionMap[id]; exists {
				validPoolIDs = append(validPoolIDs, id)
			}
		}

		if len(validPoolIDs) > 0 {
			allChapterQuestions = append(allChapterQuestions, chapterQuestions{
				chapterID:   chapterID,
				questions:   validPoolIDs,
				questionMap: questionMap,
			})
		}
	}

	if len(allChapterQuestions) == 0 {
		return nil, fmt.Errorf("no questions available in section: %s", sectionID)
	}

	// Select questions: at least minQuestionsPerChapter from each chapter, then fill up to targetTotalQuestions
	selectedQuestions := make([]interface{}, 0)
	selectedIDs := make(map[string]bool)

	// First pass: select minQuestionsPerChapter from each chapter
	for _, cq := range allChapterQuestions {
		// Shuffle questions for this chapter
		shuffled := make([]interface{}, len(cq.questions))
		copy(shuffled, cq.questions)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// Select up to minQuestionsPerChapter questions
		count := 0
		for _, idInterface := range shuffled {
			if count >= minQuestionsPerChapter || len(selectedQuestions) >= targetTotalQuestions {
				break
			}
			id := idInterface.(string) // validPoolIDs only contains strings
			if !selectedIDs[id] {
				if q, exists := cq.questionMap[id]; exists {
					// CRITICAL: Add chapter ID to question for category tests
					// This allows SubmitTest to find the correct question when IDs are duplicated across chapters
					if qMap, ok := q.(map[string]interface{}); ok {
						qMap["_category_test_chapter_id"] = cq.chapterID
					}
					selectedQuestions = append(selectedQuestions, q)
					selectedIDs[id] = true
					count++
				}
			}
		}
	}

	// Second pass: fill remaining slots randomly from all chapters
	if len(selectedQuestions) < targetTotalQuestions {
		// Collect all remaining questions with their chapter IDs
		type questionWithChapter struct {
			question  interface{}
			chapterID string
		}
		allRemaining := make([]questionWithChapter, 0)
		for _, cq := range allChapterQuestions {
			for _, idInterface := range cq.questions {
				id := idInterface.(string) // validPoolIDs only contains strings
				if !selectedIDs[id] {
					if q, exists := cq.questionMap[id]; exists {
						// CRITICAL: Store question with its chapter ID to avoid wrong assignment
						allRemaining = append(allRemaining, questionWithChapter{
							question:  q,
							chapterID: cq.chapterID,
						})
					}
				}
			}
		}

		// Shuffle and select remaining
		rand.Shuffle(len(allRemaining), func(i, j int) {
			allRemaining[i], allRemaining[j] = allRemaining[j], allRemaining[i]
		})

		for len(selectedQuestions) < targetTotalQuestions && len(allRemaining) > 0 {
			item := allRemaining[0]
			q := item.question
			if qMap, ok := q.(map[string]interface{}); ok {
				if id, ok := qMap["id"].(string); ok {
					// CRITICAL: Use the chapter ID we stored, don't search for it
					// This ensures the question is correctly assigned to the chapter it came from
					qMap["_category_test_chapter_id"] = item.chapterID
					selectedIDs[id] = true
				}
			}
			selectedQuestions = append(selectedQuestions, q)
			allRemaining = allRemaining[1:]
		}
	}

	// Remove correct_answer from selected questions (except for reorder type)
	for _, q := range selectedQuestions {
		if qMap, ok := q.(map[string]interface{}); ok {
			questionType, _ := qMap["type"].(string)
			if questionType != "reorder" {
				delete(qMap, "correct_answer")
			}
		}
	}

	return &TestQuestions{
		Questions: selectedQuestions,
		Total:     len(selectedQuestions),
	}, nil
}

// TestQuestions represents test questions
type TestQuestions struct {
	Questions []interface{} `json:"questions"`
	Total     int           `json:"total"`
}

// selectQuestions selects questions based on strategy
func (s *GrammarService) selectQuestions(poolIDs []interface{}, questionMap map[string]interface{}, config map[string]interface{}, numQuestions int) []interface{} {
	strategyType, _ := config["type"].(string)

	if strategyType == "stratified_by_theory_block" {
		return s.selectStratified(poolIDs, questionMap, config, numQuestions)
	}

	// Default: random selection
	return s.selectRandom(poolIDs, questionMap, numQuestions)
}

// selectStratified selects questions stratified by theory block
func (s *GrammarService) selectStratified(poolIDs []interface{}, questionMap map[string]interface{}, config map[string]interface{}, numQuestions int) []interface{} {
	minPerBlock := 1
	if m, ok := config["min_per_theory_block"].(float64); ok {
		minPerBlock = int(m)
	}

	// Group questions by theory_block_id
	blockGroups := make(map[string][]interface{})
	for _, idInterface := range poolIDs {
		id, ok := idInterface.(string)
		if !ok {
			continue
		}
		q, exists := questionMap[id]
		if !exists {
			continue
		}

		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}

		blockID, _ := qMap["theory_block_id"].(string)
		if blockID == "" {
			blockID = "unknown"
		}

		blockGroups[blockID] = append(blockGroups[blockID], q)
	}

	// Select min_per_block from each group, then fill remaining randomly
	selected := make([]interface{}, 0, numQuestions)
	selectedIDs := make(map[string]bool)

	// First pass: min from each block
	for _, group := range blockGroups {
		rand.Shuffle(len(group), func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
		for i := 0; i < minPerBlock && i < len(group) && len(selected) < numQuestions; i++ {
			qMap := group[i].(map[string]interface{}) // blockGroups only contains maps
			id, _ := qMap["id"].(string)
			if !selectedIDs[id] {
				selected = append(selected, group[i])
				selectedIDs[id] = true
			}
		}
	}

	// Second pass: fill remaining randomly
	allRemaining := make([]interface{}, 0)
	for _, idInterface := range poolIDs {
		id, ok := idInterface.(string)
		if !ok {
			continue
		}
		if !selectedIDs[id] {
			if q, exists := questionMap[id]; exists {
				allRemaining = append(allRemaining, q)
			}
		}
	}

	rand.Shuffle(len(allRemaining), func(i, j int) {
		allRemaining[i], allRemaining[j] = allRemaining[j], allRemaining[i]
	})

	for len(selected) < numQuestions && len(allRemaining) > 0 {
		selected = append(selected, allRemaining[0])
		allRemaining = allRemaining[1:]
	}

	return selected
}

// selectRandom selects questions randomly
func (s *GrammarService) selectRandom(poolIDs []interface{}, questionMap map[string]interface{}, numQuestions int) []interface{} {
	available := make([]interface{}, 0)
	for _, idInterface := range poolIDs {
		id, ok := idInterface.(string)
		if !ok {
			continue
		}
		if q, exists := questionMap[id]; exists {
			available = append(available, q)
		}
	}

	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	if len(available) > numQuestions {
		return available[:numQuestions]
	}
	return available
}

// AnswerItem represents a single answer with explicit question identification
type AnswerItem struct {
	QuestionID string      `json:"question_id"`          // question ID (unique within chapter)
	ChapterID  string      `json:"chapter_id,omitempty"` // chapter ID (required for category tests)
	Answer     interface{} `json:"answer"`               // user's answer (can be null if not answered)
}

// SubmitTest checks answers and saves attempt
// answers is an array of AnswerItem objects in the order they appear in the test
// Each AnswerItem explicitly links an answer to its question via question_id and chapter_id (for category tests)
func (s *GrammarService) SubmitTest(ctx context.Context, userID int64, scopeType, scopeID string, answers []AnswerItem) (*TestResult, error) {
	return s.SubmitTestWithClientAttemptID(ctx, userID, scopeType, scopeID, answers, "", nil)
}

// SubmitTestWithClientAttemptID checks answers and saves an attempt with an optional
// client-side idempotency key used by offline sync. attemptAt, when set, is stored as
// the attempt completion time instead of the sync time.
func (s *GrammarService) SubmitTestWithClientAttemptID(ctx context.Context, userID int64, scopeType, scopeID string, answers []AnswerItem, clientAttemptID string, attemptAt *time.Time) (*TestResult, error) {
	var questionMap map[string]map[string]interface{}
	var questionMapByChapter map[string]map[string]map[string]interface{} // For category tests
	var err error

	switch scopeType {
	case "chapter":
		chapter, err := s.ContentRepo.GetChapter(scopeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chapter: %w", err)
		}

		// Get question bank
		questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid question bank")
		}

		// Build question map with correct answers
		questionMap = make(map[string]map[string]interface{})
		for _, q := range questionBank {
			qMap, ok := q.(map[string]interface{})
			if !ok {
				continue
			}
			if id, ok := qMap["id"].(string); ok {
				questionMap[id] = qMap
			}
		}
	case "category":
		// For category tests, we need to get questions from all chapters in the section
		// CRITICAL: Questions from different chapters may have the same ID, so we need
		// to use a composite key (chapterID:questionID) or track which chapter each question came from
		sectionsData, err := s.ContentRepo.GetSections()
		if err != nil {
			return nil, fmt.Errorf("failed to get sections: %w", err)
		}

		// Find the section
		var section *repository.Section
		for i := range sectionsData.Sections {
			if sectionsData.Sections[i].SectionID == scopeID {
				section = &sectionsData.Sections[i]
				break
			}
		}

		if section == nil {
			return nil, fmt.Errorf("section not found: %s", scopeID)
		}

		// Get published items to filter chapters
		publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
		if err != nil {
			return nil, fmt.Errorf("failed to get published items: %w", err)
		}

		// Build question map from all chapters in the section
		// CRITICAL: Questions from different chapters may have the same ID.
		// We need to build a map that can handle this. Since we process chapters
		// in the same order as GenerateCategoryTest (section.ChapterIDs), we should
		// get the same questions. But to be safe, we'll use a two-level map:
		// first by chapter ID, then by question ID.
		questionMapByChapter = make(map[string]map[string]map[string]interface{})
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue // Skip unpublished chapters
			}

			chapter, err := s.ContentRepo.GetChapter(chapterID)
			if err != nil {
				s.logger.Warn("failed to get chapter for category test submission", zap.String("chapter_id", chapterID), zap.Error(err))
				continue
			}

			// Get question bank
			questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
			if !ok {
				continue
			}

			// Build chapter-level question map
			chapterQuestionMap := make(map[string]map[string]interface{})
			for _, q := range questionBank {
				qMap, ok := q.(map[string]interface{})
				if !ok {
					continue
				}
				if id, ok := qMap["id"].(string); ok {
					chapterQuestionMap[id] = qMap
				}
			}
			questionMapByChapter[chapterID] = chapterQuestionMap
		}

		// Build question map using chapter information from questionOrder
		// CRITICAL: For category tests, questions may have duplicate IDs across chapters.
		// We need to use the chapter ID stored in each question from GenerateCategoryTest
		// to find the correct question.
		questionMap = make(map[string]map[string]interface{})

		// First, build a map by composite key (chapterID:questionID) for all questions
		questionMapByCompositeKey := make(map[string]map[string]interface{})
		for chapterID, chapterQuestionMap := range questionMapByChapter {
			for id, qMap := range chapterQuestionMap {
				compositeKey := chapterID + ":" + id
				questionMapByCompositeKey[compositeKey] = qMap
			}
		}

		// Then, for each question in questionOrder, find it using chapter ID if available
		// If chapter ID is not in questionOrder, fall back to simple ID lookup
		// (this handles the case where questionOrder doesn't have chapter info)
		for _, chapterID := range section.ChapterIDs {
			chapterQuestionMap, exists := questionMapByChapter[chapterID]
			if !exists {
				continue
			}
			for id, qMap := range chapterQuestionMap {
				// Only add if not already present (first occurrence wins for backward compatibility)
				if _, exists := questionMap[id]; !exists {
					questionMap[id] = qMap
				}
			}
		}

		if len(questionMap) == 0 {
			return nil, fmt.Errorf("no questions found in section: %s", scopeID)
		}
	default:
		return nil, fmt.Errorf("unsupported scope type: %s", scopeType)
	}

	// Process answers in the order they appear in the test
	// Each answer item explicitly links question_id, chapter_id (for category tests), and answer
	results := make([]interface{}, 0)
	correct := 0
	total := 0

	// Process each answer item in order
	for answerIndex, answerItem := range answers {
		questionID := answerItem.QuestionID
		chapterID := answerItem.ChapterID
		userAnswer := answerItem.Answer
		hasAnswer := userAnswer != nil

		// Validate that chapter_id is provided for category tests
		if scopeType == "category" && chapterID == "" {
			s.logger.Warn("missing chapter_id for category test answer",
				zap.String("question_id", questionID),
				zap.String("scope_id", scopeID),
				zap.Int("answer_index", answerIndex))
			// Try to find question in any chapter (fallback)
		}

		// Find question using chapterID (for category tests) or simple questionID (for chapter tests)
		var q map[string]interface{}
		var qExists bool

		if scopeType == "category" && chapterID != "" && questionMapByChapter != nil {
			// Use question from specific chapter
			if chapterQuestionMap, chapterExists := questionMapByChapter[chapterID]; chapterExists {
				if chapterQ, chapterQExists := chapterQuestionMap[questionID]; chapterQExists {
					q = chapterQ
					qExists = true
				}
			}
		}

		// Fallback to simple lookup if chapter-specific lookup failed
		if !qExists {
			q, qExists = questionMap[questionID]
		}

		if !qExists {
			// Question not found - log error and skip
			s.logger.Warn("question not found in questionMap during submission",
				zap.String("question_id", questionID),
				zap.String("chapter_id", chapterID),
				zap.String("scope_type", scopeType),
				zap.String("scope_id", scopeID),
				zap.Int("answer_index", answerIndex))
			continue
		}

		// Process the question
		total++
		correctAnswer := q["correct_answer"]
		prompt, _ := q["prompt"].(string)

		// If no answer provided, mark as incorrect
		if !hasAnswer {
			resultItem := map[string]interface{}{
				"question_id":    questionID,
				"prompt":         prompt,
				"correct":        false,
				"user_answer":    nil,
				"correct_answer": correctAnswer,
				"explanation":    q["explanation"],
			}
			// For category tests, include chapter_id to uniquely identify questions
			if scopeType == "category" && chapterID != "" {
				resultItem["chapter_id"] = chapterID
			}
			results = append(results, resultItem)
			s.RecordGrammarTheoryAttemptFromTest(userID, q, false, attemptAt)
			continue
		}

		// Check if answer is correct
		qType, _ := q["type"].(string)
		var isCorrect bool
		var resultCorrect = correctAnswer

		if qType == "true_false" {
			uNorm, okU := normalizeTrueFalseValue(userAnswer)
			cNorm, okC := normalizeTrueFalseValue(correctAnswer)
			if okU && okC {
				isCorrect = (uNorm == cNorm)
				resultCorrect = cNorm
			} else {
				isCorrect = s.compareAnswers(userAnswer, correctAnswer)
			}
		} else {
			isCorrect = s.compareAnswers(userAnswer, correctAnswer)
		}

		if isCorrect {
			correct++
		}

		// Include prompt in results to avoid confusion when displaying
		resultItem := map[string]interface{}{
			"question_id":    questionID,
			"prompt":         prompt,
			"correct":        isCorrect,
			"user_answer":    userAnswer,
			"correct_answer": resultCorrect,
			"explanation":    q["explanation"],
		}
		// For category tests, include chapter_id to uniquely identify questions
		if scopeType == "category" && chapterID != "" {
			resultItem["chapter_id"] = chapterID
		}
		results = append(results, resultItem)
		s.RecordGrammarTheoryAttemptFromTest(userID, q, isCorrect, attemptAt)
	}

	score := 0
	if total > 0 {
		score = (correct * 100) / total
	}
	// UI copy says "at least 50% to pass", so 50% counts as passed.
	passed := score >= 50

	// Save attempt
	// Convert answers array back to JSON for storage
	answersJSON, _ := json.Marshal(answers)
	resultsJSON, _ := json.Marshal(results)

	finishedAt := time.Now().UTC()
	if attemptAt != nil && !attemptAt.IsZero() {
		finishedAt = attemptAt.UTC()
	}

	attempt := &repository.TestAttempt{
		UserID:         userID,
		ScopeType:      scopeType,
		ScopeID:        scopeID,
		StartedAt:      finishedAt,
		FinishedAt:     &finishedAt,
		Score:          score,
		Passed:         passed,
		TotalQuestions: total,
		AnswersJSON:    string(answersJSON),
		ResultsJSON:    string(resultsJSON),
	}
	if strings.TrimSpace(clientAttemptID) != "" {
		id := strings.TrimSpace(clientAttemptID)
		attempt.ClientAttemptID = &id
	}

	attemptID, err := s.AttemptRepo.CreateAttempt(attempt)
	if err != nil {
		s.logger.Error("failed to save attempt", zap.Error(err))
	}

	// Update progress for chapter tests
	switch scopeType {
	case "chapter":
		if err := s.AttemptRepo.UpdateProgress(userID, scopeID, score, passed); err != nil {
			s.logger.Error("failed to update progress", zap.Error(err))
		}
	case "category":
		// For category tests, we need to save category test progress
		// This will be used to unlock the next category
		_ = s.AttemptRepo.UpdateCategoryTestProgress(userID, scopeID, score, passed)
	}

	return &TestResult{
		AttemptID:  attemptID,
		Score:      score,
		Passed:     passed,
		Correct:    correct,
		Total:      total,
		Results:    results,
		AnsweredAt: finishedAt,
	}, nil
}

// TestResult represents test submission result
type TestResult struct {
	AttemptID  int64
	Score      int
	Passed     bool
	Correct    int
	Total      int
	Results    []interface{}
	AnsweredAt time.Time
}

// normalizeTrueFalseValue maps Да/Нет, true/false, bool, etc. to "true" or "false"
// so that comparison and frontend display (formatAnswer) work consistently.
// Returns (normalized, true) when recognized; otherwise ("", false).
func normalizeTrueFalseValue(v interface{}) (string, bool) {
	switch val := v.(type) {
	case bool:
		if val {
			return "true", true
		}
		return "false", true
	case string:
		lower := strings.TrimSpace(strings.ToLower(val))
		switch lower {
		case "true", "да", "yes", "1":
			return "true", true
		case "false", "нет", "no", "0":
			return "false", true
		}
	}
	return "", false
}

// normalizeAnswerString trims spaces, collapses multiple spaces to one, lowercases for comparison
func normalizeAnswerString(str string) string {
	s := strings.TrimSpace(str)
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToLower(s)
}

// compareAnswers compares user answer with correct answer (case-insensitive, spaces normalized)
func (s *GrammarService) compareAnswers(userAnswer, correctAnswer interface{}) bool {
	// Handle different answer types
	switch v := correctAnswer.(type) {
	case string:
		userStr, ok := userAnswer.(string)
		if !ok {
			// Allow string comparison if user sent e.g. number as string
			userStr = fmt.Sprintf("%v", userAnswer)
		}
		return normalizeAnswerString(userStr) == normalizeAnswerString(v)
	case []interface{}:
		userArr, ok := userAnswer.([]interface{})
		if !ok {
			return false
		}
		if len(userArr) != len(v) {
			return false
		}
		// Simple comparison - may need more sophisticated logic
		for i, correctItem := range v {
			if i >= len(userArr) || !s.compareAnswers(userArr[i], correctItem) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		userMap, ok := userAnswer.(map[string]interface{})
		if !ok {
			return false
		}
		// Compare map keys/values
		for k, correctVal := range v {
			userVal, exists := userMap[k]
			if !exists || !s.compareAnswers(userVal, correctVal) {
				return false
			}
		}
		return true
	default:
		return userAnswer == correctAnswer
	}
}

// CanAccessChapter checks if a user can access a chapter
// A chapter can be accessed if:
// 1. The user has access to the whole section (placement test or previous section passed), OR
// 2. It's the first chapter in its section, OR
// 3. The previous chapter has been passed (score >= 50%)
func (s *GrammarService) CanAccessChapter(ctx context.Context, userID int64, chapterID string) (bool, error) {
	// Get the chapter to find its section
	chapter, err := s.ContentRepo.GetChapter(chapterID)
	if err != nil {
		return false, fmt.Errorf("failed to get chapter: %w", err)
	}

	// Get sections to find chapter order
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return false, fmt.Errorf("failed to get sections: %w", err)
	}

	// Find the section containing this chapter
	var section *repository.Section
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == chapter.SectionID {
			section = &sectionsData.Sections[i]
			break
		}
	}

	if section == nil {
		return false, fmt.Errorf("section not found for chapter: %s", chapterID)
	}

	// Check if section was opened by placement test (in this case, all chapters are accessible)
	isOpenedByPlacement, _ := s.isSectionOpenedByPlacement(ctx, userID, section.SectionID)
	if isOpenedByPlacement {
		return true, nil
	}

	// If section is not accessible, first chapter is not accessible either
	canSection, _ := s.CanAccessSection(ctx, userID, section.SectionID)
	if !canSection {
		return false, nil
	}

	// Find the index of this chapter in the section
	chapterIndex := -1
	for i, id := range section.ChapterIDs {
		if id == chapterID {
			chapterIndex = i
			break
		}
	}

	if chapterIndex == -1 {
		return false, fmt.Errorf("chapter not found in section: %s", chapterID)
	}

	// First chapter is always accessible
	if chapterIndex == 0 {
		return true, nil
	}

	// Check if previous chapter has been passed
	previousChapterID := section.ChapterIDs[chapterIndex-1]
	progress, err := s.AttemptRepo.GetChapterProgress(userID, previousChapterID)
	if err != nil {
		return false, fmt.Errorf("failed to get progress for previous chapter: %w", err)
	}

	// Chapter is accessible if previous chapter was passed (score >= 50%)
	return progress.Passed, nil
}

// IsSectionOpenedByPlacement reports whether the user unlocked this section via placement test.
func (s *GrammarService) IsSectionOpenedByPlacement(ctx context.Context, userID int64, sectionID string) (bool, error) {
	return s.isSectionOpenedByPlacement(ctx, userID, sectionID)
}

// isSectionOpenedByPlacement checks if a section was opened by placement test
// This is used to determine if all chapters should be accessible (placement) or only first chapter (category test)
func (s *GrammarService) isSectionOpenedByPlacement(ctx context.Context, userID int64, sectionID string) (bool, error) {
	placementResult, _ := s.AttemptRepo.GetPlacementTestResult(userID)
	if placementResult == nil {
		return false, nil
	}

	// Check if section is in OpenedSections from placement test
	for _, openedSectionID := range placementResult.OpenedSections {
		if openedSectionID == sectionID {
			return true, nil
		}
	}

	// Check placement "effective level" - if section level <= max level among opened sections
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return false, fmt.Errorf("failed to get sections: %w", err)
	}

	levelOrder := map[string]int{"A0": 0, "A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6, "mixed": -1}
	if len(placementResult.OpenedSections) > 0 {
		effectiveOrder := -1
		for _, oid := range placementResult.OpenedSections {
			for j := range sectionsData.Sections {
				if sectionsData.Sections[j].SectionID == oid {
					if ord, ok := levelOrder[sectionsData.Sections[j].Level]; ok && ord >= 0 && ord > effectiveOrder {
						effectiveOrder = ord
					}
					break
				}
			}
		}
		if effectiveOrder >= 0 {
			var section *repository.Section
			for i := range sectionsData.Sections {
				if sectionsData.Sections[i].SectionID == sectionID {
					section = &sectionsData.Sections[i]
					break
				}
			}
			if section != nil {
				secOrd, ok := levelOrder[section.Level]
				if ok && secOrd >= 0 && secOrd <= effectiveOrder {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CanAccessSection checks if a user can access a section (category)
// A section can be accessed if:
//  1. It's the first section, OR
//  2. Category test for previous section was passed (score >= 50%), OR
//  3. It was opened by placement test (in OpenedSections), OR
//  4. Placement "effective level": section level <= max level among OpenedSections (fixes old DB rows
//     where OpenedSections missed sections that had no questions in the 25-question test)
func (s *GrammarService) CanAccessSection(ctx context.Context, userID int64, sectionID string) (bool, error) {
	placementResult, _ := s.AttemptRepo.GetPlacementTestResult(userID)
	if placementResult != nil {
		for _, openedSectionID := range placementResult.OpenedSections {
			if openedSectionID == sectionID {
				return true, nil
			}
		}
	}

	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return false, fmt.Errorf("failed to get sections: %w", err)
	}

	var section *repository.Section
	var sectionIndex = -1
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			section = &sectionsData.Sections[i]
			sectionIndex = i
			break
		}
	}

	if section == nil {
		return false, fmt.Errorf("section not found: %s", sectionID)
	}

	// Placement "effective level": if stored OpenedSections is incomplete (old bug), treat the
	// highest level among opened sections as the placement level and open any section at or below it.
	levelOrder := map[string]int{"A0": 0, "A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6, "mixed": -1}
	if placementResult != nil && len(placementResult.OpenedSections) > 0 {
		effectiveOrder := -1
		for _, oid := range placementResult.OpenedSections {
			for j := range sectionsData.Sections {
				if sectionsData.Sections[j].SectionID == oid {
					if ord, ok := levelOrder[sectionsData.Sections[j].Level]; ok && ord >= 0 && ord > effectiveOrder {
						effectiveOrder = ord
					}
					break
				}
			}
		}
		if effectiveOrder >= 0 {
			secOrd, ok := levelOrder[section.Level]
			if ok && secOrd >= 0 && secOrd <= effectiveOrder {
				return true, nil
			}
		}
	}

	// First section is always accessible
	if sectionIndex == 0 {
		return true, nil
	}

	// Check if category test for previous section was passed (score >= 50%)
	previousSection := sectionsData.Sections[sectionIndex-1]

	// Require category test to be passed (score >= 50%) to unlock next section
	// Exception: if section was opened by placement test (checked above)
	categoryTestPassed, err := s.AttemptRepo.GetCategoryTestProgress(userID, previousSection.SectionID)
	if err != nil {
		return false, fmt.Errorf("failed to get category test progress: %w", err)
	}
	if categoryTestPassed {
		return true, nil
	}

	// Fallback for migrated/legacy data: if category test attempt is missing,
	// unlock next section when all published chapters in the previous section are passed.
	previousPublishedChapters, chaptersErr := s.GetPublishedChapters(ctx, previousSection.SectionID, userID)
	if chaptersErr == nil && len(previousPublishedChapters) > 0 {
		allPublishedPassed := true
		for _, chapter := range previousPublishedChapters {
			if chapter.Progress == nil || !chapter.Progress.Passed {
				allPublishedPassed = false
				break
			}
		}
		if allPublishedPassed {
			return true, nil
		}
	}

	// Next section is only accessible if category test was passed (score >= 50%)
	return false, nil
}

// GrammarStatistics represents overall grammar course statistics
type GrammarStatistics struct {
	ConfirmedLevel           string `json:"confirmed_level"`             // Highest level where all chapters are passed
	CourseCompletionPct      int    `json:"course_completion_pct"`       // Completion % over published chapters only (зачётка)
	WholeCourseCompletionPct int    `json:"whole_course_completion_pct"` // Completion % over entire course (all chapters in bundle)
	AverageTestScore         int    `json:"average_test_score"`          // Average percentage across all test attempts
	PassedChapters           int    `json:"passed_chapters"`             // Number of passed (published) chapters
	TotalChapters            int    `json:"total_chapters"`              // Total number of published chapters
	TotalChaptersInCourse    int    `json:"total_chapters_in_course"`    // Total chapters in course (whole bundle)
}

// GetGrammarStatistics calculates overall grammar statistics for a user
func (s *GrammarService) GetGrammarStatistics(ctx context.Context, userID int64) (*GrammarStatistics, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Level hierarchy for comparison
	levelOrder := map[string]int{
		"A0":    0,
		"A1":    1,
		"A2":    2,
		"B1":    3,
		"B2":    4,
		"C1":    5,
		"C2":    6,
		"mixed": -1, // Mixed levels don't count for confirmed level
	}

	// Track confirmed level (highest level where all chapters are passed)
	confirmedLevel := ""
	confirmedLevelOrder := -1

	// Count total chapters in course (whole bundle, all sections)
	totalChaptersInCourse := 0
	for i := range sectionsData.Sections {
		totalChaptersInCourse += len(sectionsData.Sections[i].ChapterIDs)
	}

	// Total score over entire course (unpublished chapters count as 0)
	totalScoreWholeCourse := 0

	// Track overall progress (published only)
	totalScore := 0
	totalPublishedChapters := 0
	passedChaptersCount := 0

	// Process each section
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]

		// Check if section is published
		sectionItem, err := s.PublishRepo.GetPublishedItem("section", section.SectionID)
		if err != nil || sectionItem == nil || !sectionItem.IsPublished {
			continue
		}

		sectionLevel := section.Level
		sectionLevelOrder, hasLevel := levelOrder[sectionLevel]
		if !hasLevel || sectionLevelOrder < 0 {
			continue // Skip sections without valid level
		}

		// Check all published chapters in this section
		allChaptersPassed := true
		sectionTotalScore := 0
		sectionPublishedChapters := 0

		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue // Skip unpublished chapters
			}

			sectionPublishedChapters++

			// Placement grants access; only actual chapter attempts establish progress.
			progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
			if progress == nil {
				allChaptersPassed = false
				continue
			}
			sectionTotalScore += progress.BestScore
			if progress.Passed {
				passedChaptersCount++
			} else {
				allChaptersPassed = false
			}
		}

		// Update confirmed level if all chapters in this section are passed
		if allChaptersPassed && sectionPublishedChapters > 0 && sectionLevelOrder > confirmedLevelOrder {
			confirmedLevel = sectionLevel
			confirmedLevelOrder = sectionLevelOrder
		}

		// Add to overall totals
		totalScore += sectionTotalScore
		totalPublishedChapters += sectionPublishedChapters
	}

	// Calculate overall completion percentage
	completionPct := 0
	if totalPublishedChapters > 0 {
		completionPct = totalScore / totalPublishedChapters
	}

	// Get average test score across all attempts
	averageTestScore, err := s.AttemptRepo.GetAverageTestScore(userID)
	if err != nil {
		s.logger.Warn("failed to get average test score", zap.Error(err))
		averageTestScore = 0
	}

	// If no confirmed level, set to empty string or "A0" if user has any progress
	if confirmedLevel == "" && totalPublishedChapters > 0 {
		// Check if user has any progress at all
		if totalScore > 0 {
			confirmedLevel = "A0" // Starting level
		}
	}

	// Whole course: sum of best scores for every chapter in bundle (unpublished = 0)
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				totalScoreWholeCourse += 0
				continue
			}
			progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
			if progress != nil {
				totalScoreWholeCourse += progress.BestScore
			}
		}
	}
	wholeCourseCompletionPct := 0
	if totalChaptersInCourse > 0 {
		wholeCourseCompletionPct = totalScoreWholeCourse / totalChaptersInCourse
	}

	return &GrammarStatistics{
		ConfirmedLevel:           confirmedLevel,
		CourseCompletionPct:      completionPct,
		WholeCourseCompletionPct: wholeCourseCompletionPct,
		AverageTestScore:         averageTestScore,
		PassedChapters:           passedChaptersCount,
		TotalChapters:            totalPublishedChapters,
		TotalChaptersInCourse:    totalChaptersInCourse,
	}, nil
}

// GeneratePlacementTest generates a placement test with at least 1 question from each
// published section (category), then fills up to 25. Questions are ordered by section then chapter.
func (s *GrammarService) GeneratePlacementTest(ctx context.Context) (*TestQuestions, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Collect all questions, grouped by section; add placement_section_id, _section_order, _chapter_order
	allQuestions := make([]interface{}, 0)
	bySection := make(map[string][]interface{})

	for _, section := range sectionsData.Sections {
		for ci, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue
			}

			chapter, err := s.ContentRepo.GetChapter(chapterID)
			if err != nil {
				s.logger.Warn("failed to load chapter for placement test", zap.String("chapter_id", chapterID), zap.Error(err))
				continue
			}

			questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
			if !ok {
				continue
			}

			for _, q := range questionBank {
				qMap, ok := q.(map[string]interface{})
				if !ok {
					continue
				}
				if !repository.GrammarQuestionAvailable(qMap) {
					continue
				}
				if id, ok := qMap["id"].(string); ok && id != "" {
					compID := chapterID + ":" + id
					qMap["id"] = compID
					qMap["placement_chapter_id"] = chapterID
					qMap["placement_chapter_title"] = chapter.Title
					qMap["placement_section_id"] = section.SectionID
					qMap["placement_section_order"] = section.Order
					qMap["placement_chapter_order"] = ci
					allQuestions = append(allQuestions, qMap)
					bySection[section.SectionID] = append(bySection[section.SectionID], qMap)
				}
			}
		}
	}

	// Phase 1: at least 1 question from each section that has questions
	selectedIDs := make(map[string]bool)
	selected := make([]interface{}, 0)
	for _, section := range sectionsData.Sections {
		list := bySection[section.SectionID]
		if len(list) == 0 {
			continue
		}
		idx := rand.Intn(len(list))
		q := list[idx].(map[string]interface{})
		selected = append(selected, q)
		selectedIDs[q["id"].(string)] = true
	}

	// Phase 2: fill up to 25 with random questions from the rest
	const placementNumQuestions = 25
	pool := make([]interface{}, 0, len(allQuestions)-len(selected))
	for _, q := range allQuestions {
		qMap := q.(map[string]interface{})
		if !selectedIDs[qMap["id"].(string)] {
			pool = append(pool, q)
		}
	}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	need := placementNumQuestions - len(selected)
	if need > 0 && len(pool) > 0 {
		n := need
		if len(pool) < n {
			n = len(pool)
		}
		for i := 0; i < n; i++ {
			selected = append(selected, pool[i])
		}
	}

	// Sort by section order, then chapter order
	sort.Slice(selected, func(i, j int) bool {
		a, b := selected[i].(map[string]interface{}), selected[j].(map[string]interface{})
		soA, _ := a["placement_section_order"].(int)
		soB, _ := b["placement_section_order"].(int)
		if soA != soB {
			return soA < soB
		}
		coA, _ := a["placement_chapter_order"].(int)
		coB, _ := b["placement_chapter_order"].(int)
		return coA < coB
	})

	// Remove correct_answer from selected questions (except for reorder type)
	selectedQuestions := selected
	for _, q := range selectedQuestions {
		if qMap, ok := q.(map[string]interface{}); ok {
			questionType, _ := qMap["type"].(string)
			if questionType != "reorder" {
				delete(qMap, "correct_answer")
			}
		}
	}

	return &TestQuestions{
		Questions: selectedQuestions,
		Total:     len(selectedQuestions),
	}, nil
}

// SubmitPlacementTest submits placement test answers and determines user level.
// Level is determined by evaluating correctness from first section to last: we open
// sections up to the point where the user answered confidently (>=50% on that section's
// questions). The level is the CEFR level of the last opened section.
func (s *GrammarService) SubmitPlacementTest(ctx context.Context, userID int64, answers map[string]interface{}) (*PlacementTestResult, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Chapter order for sorting results by course sequence (sections then chapters)
	chapterOrder := make(map[string]int)
	ord := 0
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue
			}
			chapterOrder[chapterID] = ord
			ord++
		}
	}

	// Build question map with correct answers; track chapter, section, level per question
	questionMap := make(map[string]map[string]interface{})
	questionSectionMap := make(map[string]string)
	questionChapterTitleMap := make(map[string]string)
	questionChapterIDMap := make(map[string]string)
	questionLevelMap := make(map[string]string)

	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue
			}

			chapter, err := s.ContentRepo.GetChapter(chapterID)
			if err != nil {
				continue
			}

			questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
			if !ok {
				continue
			}

			for _, q := range questionBank {
				qMap, ok := q.(map[string]interface{})
				if !ok {
					continue
				}
				if id, ok := qMap["id"].(string); ok && id != "" {
					key := chapterID + ":" + id
					questionMap[key] = qMap
					questionSectionMap[key] = section.SectionID
					questionChapterTitleMap[key] = chapter.Title
					questionChapterIDMap[key] = chapterID
					questionLevelMap[key] = section.Level
				}
			}
		}
	}

	// Score per section: from first to last, we open until the first section where user scored <50%
	sectionScores := make(map[string]struct{ correct, total int })

	totalCorrect := 0
	totalQuestions := 0

	// Collect question IDs from answers; sort by chapter order (course sequence)
	answerQuestionIDs := make([]string, 0, len(answers))
	for qid := range answers {
		answerQuestionIDs = append(answerQuestionIDs, qid)
	}
	sort.Slice(answerQuestionIDs, func(i, j int) bool {
		ci := questionChapterIDMap[answerQuestionIDs[i]]
		cj := questionChapterIDMap[answerQuestionIDs[j]]
		oi, oki := chapterOrder[ci]
		oj, okj := chapterOrder[cj]
		if !oki {
			oi = 999999
		}
		if !okj {
			oj = 999999
		}
		if oi != oj {
			return oi < oj
		}
		return answerQuestionIDs[i] < answerQuestionIDs[j]
	})

	// Build per-question results and section scores
	results := make([]interface{}, 0, len(answers))

	for _, questionID := range answerQuestionIDs {
		userAnswer := answers[questionID]
		q, exists := questionMap[questionID]
		if !exists {
			continue
		}

		totalQuestions++
		correctAnswer := q["correct_answer"]
		qType, _ := q["type"].(string)
		var isCorrect bool
		var resultCorrect = correctAnswer

		if qType == "true_false" {
			uNorm, okU := normalizeTrueFalseValue(userAnswer)
			cNorm, okC := normalizeTrueFalseValue(correctAnswer)
			if okU && okC {
				isCorrect = (uNorm == cNorm)
				resultCorrect = cNorm
			} else {
				isCorrect = s.compareAnswers(userAnswer, correctAnswer)
			}
		} else {
			isCorrect = s.compareAnswers(userAnswer, correctAnswer)
		}

		if isCorrect {
			totalCorrect++
		}

		sectionID := questionSectionMap[questionID]
		if sectionID != "" {
			sc := sectionScores[sectionID]
			sc.total++
			if isCorrect {
				sc.correct++
			}
			sectionScores[sectionID] = sc
		}

		explanation, _ := q["explanation"].(string)
		chapterTitle := questionChapterTitleMap[questionID]
		qLevel := questionLevelMap[questionID]
		results = append(results, map[string]interface{}{
			"question_id":             questionID,
			"correct":                 isCorrect,
			"user_answer":             userAnswer,
			"correct_answer":          resultCorrect,
			"explanation":             explanation,
			"placement_chapter_title": chapterTitle,
			"level":                   qLevel,
		})
	}

	overallScore := 0
	if totalQuestions > 0 {
		overallScore = (totalCorrect * 100) / totalQuestions
	}

	// Open sections in order: open while section has >=1 question and >=50% correct; stop at first fail
	openedSectionsList := make([]string, 0)
	var level string

	for i := range sectionsData.Sections {
		sec := &sectionsData.Sections[i]
		sc := sectionScores[sec.SectionID]
		if sc.total < 1 {
			// No questions from this section in the test — skip, do not open, do not stop
			continue
		}
		pct := (sc.correct * 100) / sc.total
		if pct >= 50 {
			openedSectionsList = append(openedSectionsList, sec.SectionID)
			if sec.Level != "" {
				level = sec.Level
			}
		} else {
			break
		}
	}

	if level == "" {
		if len(openedSectionsList) > 0 {
			level = "—"
		} else {
			level = "Below A1"
		}
	}

	// Expand opened sections: if we have a clear CEFR level, open ALL published sections
	// at or below that level. The 25-question test only samples the course, so many
	// sections have zero questions and were skipped above — they would stay locked
	// otherwise even though the user's level (e.g. B1) implies they should be open.
	levelOrder := map[string]int{
		"A0": 0, "A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6,
		"mixed": -1,
	}
	if level != "Below A1" && level != "—" {
		if maxOrder, ok := levelOrder[level]; ok && maxOrder >= 0 {
			openedSectionsList = make([]string, 0)
			for i := range sectionsData.Sections {
				sec := &sectionsData.Sections[i]
				sectionItem, errPub := s.PublishRepo.GetPublishedItem("section", sec.SectionID)
				if errPub != nil || sectionItem == nil || !sectionItem.IsPublished {
					continue
				}
				secOrder, has := levelOrder[sec.Level]
				if !has || secOrder < 0 {
					continue
				}
				if secOrder <= maxOrder {
					openedSectionsList = append(openedSectionsList, sec.SectionID)
				}
			}
		}
	}

	// Save result (only if better than existing)
	err = s.AttemptRepo.SavePlacementTestResult(userID, overallScore, totalQuestions, openedSectionsList)
	if err != nil {
		s.logger.Error("failed to save placement test result", zap.Error(err))
	}

	resultsJSON, _ := json.Marshal(results)
	answersJSON, _ := json.Marshal(answers)

	attempt := &repository.TestAttempt{
		UserID:         userID,
		ScopeType:      "placement",
		ScopeID:        "placement",
		StartedAt:      time.Now(),
		FinishedAt:     &[]time.Time{time.Now()}[0],
		Score:          overallScore,
		Passed:         overallScore >= 50,
		TotalQuestions: totalQuestions,
		AnswersJSON:    string(answersJSON),
		ResultsJSON:    string(resultsJSON),
	}

	_, err = s.AttemptRepo.CreateAttempt(attempt)
	if err != nil {
		s.logger.Error("failed to save placement test attempt", zap.Error(err))
	}

	return &PlacementTestResult{
		Score:          overallScore,
		TotalQuestions: totalQuestions,
		Correct:        totalCorrect,
		OpenedSections: openedSectionsList,
		Level:          level,
		Results:        results,
	}, nil
}

func grammarLevelOrderMap() map[string]int {
	return map[string]int{"A0": 0, "A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6, "mixed": -1}
}

// OpenPublishedSectionsThroughLevel returns published section IDs with CEFR level <= targetLevel (same expansion as placement submit).
func (s *GrammarService) OpenPublishedSectionsThroughLevel(targetLevel string) ([]string, error) {
	targetLevel = strings.TrimSpace(targetLevel)
	levelOrder := grammarLevelOrderMap()
	maxOrder, ok := levelOrder[targetLevel]
	if !ok || maxOrder < 0 {
		return nil, fmt.Errorf("unknown grammar level %q", targetLevel)
	}
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}
	opened := make([]string, 0)
	for i := range sectionsData.Sections {
		sec := &sectionsData.Sections[i]
		sectionItem, errPub := s.PublishRepo.GetPublishedItem("section", sec.SectionID)
		if errPub != nil || sectionItem == nil || !sectionItem.IsPublished {
			continue
		}
		secOrder, has := levelOrder[sec.Level]
		if !has || secOrder < 0 {
			continue
		}
		if secOrder <= maxOrder {
			opened = append(opened, sec.SectionID)
		}
	}
	return opened, nil
}

// AdminSetGrammarPlacementLevel sets or clears placement-based grammar unlocks (admin). Empty level deletes the placement row.
// Level values: A0..C2, "below_a1" / "below a1" / "Below A1" for no sections opened.
func (s *GrammarService) AdminSetGrammarPlacementLevel(_ context.Context, userID int64, levelInput string) error {
	levelInput = strings.TrimSpace(levelInput)
	if levelInput == "" {
		return s.AttemptRepo.DeletePlacementTestResult(userID)
	}
	normSpace := strings.ToLower(strings.ReplaceAll(levelInput, "_", " "))
	if normSpace == "below a1" {
		return s.AttemptRepo.UpsertPlacementByAdmin(userID, 0, 0, []string{})
	}
	norm := strings.ToUpper(strings.TrimSpace(levelInput))
	valid := map[string]struct{}{
		"A0": {}, "A1": {}, "A2": {}, "B1": {}, "B2": {}, "C1": {}, "C2": {},
	}
	if _, ok := valid[norm]; !ok {
		return fmt.Errorf("invalid grammar level %q", levelInput)
	}
	opened, err := s.OpenPublishedSectionsThroughLevel(norm)
	if err != nil {
		return err
	}
	return s.AttemptRepo.UpsertPlacementByAdmin(userID, 0, 0, opened)
}

// FormatPlacementLevelDisplay returns a short CEFR label for admin UI (empty if no placement row).
func (s *GrammarService) FormatPlacementLevelDisplay(opened []string, hasPlacementRow bool) string {
	if !hasPlacementRow {
		return ""
	}
	if len(opened) == 0 {
		return "Below A1"
	}
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return "—"
	}
	levelOrder := grammarLevelOrderMap()
	maxOrd := -1
	for _, oid := range opened {
		for j := range sectionsData.Sections {
			if sectionsData.Sections[j].SectionID != oid {
				continue
			}
			if ord, ok := levelOrder[sectionsData.Sections[j].Level]; ok && ord >= 0 && ord > maxOrd {
				maxOrd = ord
			}
			break
		}
	}
	if maxOrd < 0 {
		return "—"
	}
	labels := []string{"A0", "A1", "A2", "B1", "B2", "C1", "C2"}
	return labels[maxOrd]
}

// PlacementTestResult represents placement test submission result
type PlacementTestResult struct {
	Score          int           `json:"score"`
	TotalQuestions int           `json:"total_questions"`
	Correct        int           `json:"correct"`
	OpenedSections []string      `json:"opened_sections"`
	Level          string        `json:"level"`   // CEFR level of last opened section, or "Below A1" / "—"
	Results        []interface{} `json:"results"` // per-question: question_id, correct, user_answer, correct_answer, explanation, placement_chapter_title
}
