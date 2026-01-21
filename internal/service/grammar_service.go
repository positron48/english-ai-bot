package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// GrammarService handles grammar course business logic
type GrammarService struct {
	ContentRepo  *repository.GrammarContentRepository
	PublishRepo  *repository.GrammarPublishRepository
	AttemptRepo  *repository.GrammarAttemptRepository
	logger       *zap.Logger
}

// NewGrammarService creates a new grammar service
func NewGrammarService(
	contentRepo *repository.GrammarContentRepository,
	publishRepo *repository.GrammarPublishRepository,
	attemptRepo *repository.GrammarAttemptRepository,
	logger *zap.Logger,
) *GrammarService {
	return &GrammarService{
		ContentRepo: contentRepo,
		PublishRepo: publishRepo,
		AttemptRepo: attemptRepo,
		logger:      logger,
	}
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
			Section:         section,
			Title:           title,
			PublishedChapters: publishedChapters,
			PassedChapters:  passedChapters,
			ProgressPercentage: progressPercentage,
		})
	}

	return result, nil
}

// SectionWithProgress represents a section with user progress
type SectionWithProgress struct {
	Section          *repository.Section
	Title            string
	PublishedChapters int
	PassedChapters   int
	ProgressPercentage int // Average percentage based on chapter best_scores
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
			Chapter: chapter,
			Title:   title,
			Progress: progress,
		})
	}

	return result, nil
}

// ChapterWithProgress represents a chapter with user progress
type ChapterWithProgress struct {
	Chapter  *repository.Chapter
	Title    string
	Progress *repository.ChapterProgress
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
	questionBank := filteredChapter.QuestionBank
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
				if usedQuestionIDs[qID] {
					filteredQuestions = append(filteredQuestions, q)
				}
			}
			
			// Update question bank with filtered questions
			questionBank["questions"] = filteredQuestions
			filteredChapter.QuestionBank = questionBank
		}
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
		if !ok {
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
		Total:     numQuestions,
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
			qMap, ok := group[i].(map[string]interface{})
			if !ok {
				continue
			}
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

// SubmitTest checks answers and saves attempt
func (s *GrammarService) SubmitTest(ctx context.Context, userID int64, scopeType, scopeID string, answers map[string]interface{}) (*TestResult, error) {
	var chapter *repository.Chapter
	var err error

	if scopeType == "chapter" {
		chapter, err = s.ContentRepo.GetChapter(scopeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chapter: %w", err)
		}
	} else {
		return nil, fmt.Errorf("category tests not yet implemented")
	}

	// Get question bank
	questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid question bank")
	}

	// Build question map with correct answers
	questionMap := make(map[string]map[string]interface{})
	for _, q := range questionBank {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := qMap["id"].(string); ok {
			questionMap[id] = qMap
		}
	}

	// Check answers
	results := make([]interface{}, 0)
	correct := 0
	total := 0

	for questionID, userAnswer := range answers {
		q, exists := questionMap[questionID]
		if !exists {
			continue
		}

		total++
		correctAnswer := q["correct_answer"]
		isCorrect := s.compareAnswers(userAnswer, correctAnswer)

		if isCorrect {
			correct++
		}

		results = append(results, map[string]interface{}{
			"question_id": questionID,
			"correct":     isCorrect,
			"user_answer": userAnswer,
			"correct_answer": correctAnswer,
			"explanation": q["explanation"],
		})
	}

	score := 0
	if total > 0 {
		score = (correct * 100) / total
	}
	passed := score > 50

	// Save attempt
	answersJSON, _ := json.Marshal(answers)
	resultsJSON, _ := json.Marshal(results)

	attempt := &repository.TestAttempt{
		UserID:         userID,
		ScopeType:      scopeType,
		ScopeID:        scopeID,
		StartedAt:      time.Now(),
		FinishedAt:     &[]time.Time{time.Now()}[0],
		Score:          score,
		Passed:         passed,
		TotalQuestions: total,
		AnswersJSON:    string(answersJSON),
		ResultsJSON:    string(resultsJSON),
	}

	_, err = s.AttemptRepo.CreateAttempt(attempt)
	if err != nil {
		s.logger.Error("failed to save attempt", zap.Error(err))
	}

	// Update progress for chapter tests
	if scopeType == "chapter" {
		if err := s.AttemptRepo.UpdateProgress(userID, scopeID, score, passed); err != nil {
			s.logger.Error("failed to update progress", zap.Error(err))
		}
	}

	return &TestResult{
		Score:   score,
		Passed:  passed,
		Correct: correct,
		Total:   total,
		Results: results,
	}, nil
}

// TestResult represents test submission result
type TestResult struct {
	Score   int
	Passed  bool
	Correct int
	Total   int
	Results []interface{}
}

// compareAnswers compares user answer with correct answer
func (s *GrammarService) compareAnswers(userAnswer, correctAnswer interface{}) bool {
	// Handle different answer types
	switch v := correctAnswer.(type) {
	case string:
		userStr, ok := userAnswer.(string)
		return ok && userStr == v
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
// 1. It's the first chapter in its section, OR
// 2. The previous chapter has been passed (score >= 50%)
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

// CanAccessSection checks if a user can access a section (category)
// A section can be accessed if:
// 1. It's the first section, OR
// 2. All chapters in the previous section have been passed, OR
// 3. It was opened by placement test
func (s *GrammarService) CanAccessSection(ctx context.Context, userID int64, sectionID string) (bool, error) {
	// Get placement test result to check opened sections
	placementResult, _ := s.AttemptRepo.GetPlacementTestResult(userID)
	if placementResult != nil {
		for _, openedSectionID := range placementResult.OpenedSections {
			if openedSectionID == sectionID {
				return true, nil // Section was opened by placement test
			}
		}
	}

	// Get sections to find section order
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return false, fmt.Errorf("failed to get sections: %w", err)
	}

	// Find the section
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

	// First section is always accessible
	if sectionIndex == 0 {
		return true, nil
	}

	// Check if all chapters in previous section have been passed
	previousSection := sectionsData.Sections[sectionIndex-1]
	
	// Get published items to filter chapters
	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return false, fmt.Errorf("failed to get published items: %w", err)
	}

	// Check each published chapter in previous section
	for _, chapterID := range previousSection.ChapterIDs {
		chapterItem, exists := publishedItems[chapterID]
		if !exists || !chapterItem.IsPublished {
			continue // Skip unpublished chapters
		}

		progress, err := s.AttemptRepo.GetChapterProgress(userID, chapterID)
		if err != nil {
			return false, fmt.Errorf("failed to get progress for chapter %s: %w", chapterID, err)
		}

		// If any chapter in previous section is not passed, section is not accessible
		if !progress.Passed {
			return false, nil
		}
	}

	return true, nil
}

// GrammarStatistics represents overall grammar course statistics
type GrammarStatistics struct {
	ConfirmedLevel      string  `json:"confirmed_level"`       // Highest level where all chapters are passed
	CourseCompletionPct int     `json:"course_completion_pct"` // Overall completion percentage
	AverageTestScore    int     `json:"average_test_score"`    // Average percentage across all test attempts
	PassedChapters      int     `json:"passed_chapters"`       // Number of passed chapters
	TotalChapters       int     `json:"total_chapters"`        // Total number of published chapters
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

	// Get placement test result to check opened sections
	placementResult, _ := s.AttemptRepo.GetPlacementTestResult(userID)
	openedSectionsMap := make(map[string]bool)
	if placementResult != nil {
		for _, sectionID := range placementResult.OpenedSections {
			openedSectionsMap[sectionID] = true
		}
	}

	// Level hierarchy for comparison
	levelOrder := map[string]int{
		"A0":   0,
		"A1":   1,
		"A2":   2,
		"B1":   3,
		"B2":   4,
		"C1":   5,
		"C2":   6,
		"mixed": -1, // Mixed levels don't count for confirmed level
	}

	// Track confirmed level (highest level where all chapters are passed)
	confirmedLevel := ""
	confirmedLevelOrder := -1

	// Track overall progress
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

		// Check if section was opened by placement test
		isOpenedByPlacement := openedSectionsMap[section.SectionID]

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
			
			// If section was opened by placement test, count all chapters as passed (100%)
			if isOpenedByPlacement {
				sectionTotalScore += 100
				passedChaptersCount++
			} else {
				progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
				sectionTotalScore += progress.BestScore

				if progress.Passed {
					passedChaptersCount++
				}

				if !progress.Passed {
					allChaptersPassed = false
				}
			}
		}

		// If opened by placement test, consider all chapters passed
		if isOpenedByPlacement {
			allChaptersPassed = true
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

	return &GrammarStatistics{
		ConfirmedLevel:      confirmedLevel,
		CourseCompletionPct: completionPct,
		AverageTestScore:    averageTestScore,
		PassedChapters:      passedChaptersCount,
		TotalChapters:       totalPublishedChapters,
	}, nil
}

// GeneratePlacementTest generates a placement test with 20-30 questions from all published chapters
func (s *GrammarService) GeneratePlacementTest(ctx context.Context) (*TestQuestions, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Collect all questions from all published chapters
	allQuestions := make([]interface{}, 0)
	questionMap := make(map[string]interface{})

	for _, section := range sectionsData.Sections {
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue // Skip unpublished chapters
			}

			chapter, err := s.ContentRepo.GetChapter(chapterID)
			if err != nil {
				s.logger.Warn("failed to load chapter for placement test", zap.String("chapter_id", chapterID), zap.Error(err))
				continue
			}

			// Get question bank
			questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
			if !ok {
				continue
			}

			// Add questions to pool
			for _, q := range questionBank {
				qMap, ok := q.(map[string]interface{})
				if !ok {
					continue
				}
				if id, ok := qMap["id"].(string); ok {
					// Store chapter_id with question for tracking
					qMap["placement_chapter_id"] = chapterID
					questionMap[id] = qMap
					allQuestions = append(allQuestions, qMap)
				}
			}
		}
	}

	// Select 20-30 questions randomly
	var numQuestions int
	if len(allQuestions) < 20 {
		numQuestions = len(allQuestions)
	} else if len(allQuestions) >= 30 {
		// Random between 20-30
		numQuestions = 20 + rand.Intn(11) // 20-30
	} else {
		numQuestions = len(allQuestions)
	}

	// Shuffle and select
	rand.Shuffle(len(allQuestions), func(i, j int) {
		allQuestions[i], allQuestions[j] = allQuestions[j], allQuestions[i]
	})

	selectedQuestions := allQuestions
	if len(allQuestions) > numQuestions {
		selectedQuestions = allQuestions[:numQuestions]
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

// SubmitPlacementTest submits placement test answers and determines user level
func (s *GrammarService) SubmitPlacementTest(ctx context.Context, userID int64, answers map[string]interface{}) (*PlacementTestResult, error) {
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	publishedItems, err := s.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		return nil, fmt.Errorf("failed to get published items: %w", err)
	}

	// Build question map with correct answers from all published chapters
	// Also track which chapter each question belongs to
	questionMap := make(map[string]map[string]interface{})
	questionChapterMap := make(map[string]string) // question_id -> chapter_id

	for _, section := range sectionsData.Sections {
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
				if id, ok := qMap["id"].(string); ok {
					questionMap[id] = qMap
					questionChapterMap[id] = chapterID
				}
			}
		}
	}

	// Check answers and calculate score per chapter
	chapterScores := make(map[string]struct {
		correct int
		total   int
	})

	totalCorrect := 0
	totalQuestions := 0

	for questionID, userAnswer := range answers {
		q, exists := questionMap[questionID]
		if !exists {
			continue
		}

		totalQuestions++
		correctAnswer := q["correct_answer"]
		isCorrect := s.compareAnswers(userAnswer, correctAnswer)

		if isCorrect {
			totalCorrect++
		}

		// Track score per chapter using questionChapterMap
		chapterID, exists := questionChapterMap[questionID]
		if exists {
			if _, exists := chapterScores[chapterID]; !exists {
				chapterScores[chapterID] = struct {
					correct int
					total   int
				}{0, 0}
			}
			score := chapterScores[chapterID]
			score.total++
			if isCorrect {
				score.correct++
			}
			chapterScores[chapterID] = score
		}
	}

	// Calculate overall score
	overallScore := 0
	if totalQuestions > 0 {
		overallScore = (totalCorrect * 100) / totalQuestions
	}

	// Determine which sections to open based on chapter performance
	// A chapter is considered "passed" if user got >= 50% correct
	openedSections := make(map[string]bool)

	// Process sections in order
	for _, section := range sectionsData.Sections {
		// Check if all tested chapters in this section are "passed"
		allTestedChaptersPassed := true
		testedChaptersCount := 0

		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := publishedItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue
			}

			// Only check chapters that were actually in the test
			if score, exists := chapterScores[chapterID]; exists {
				testedChaptersCount++
				chapterScore := 0
				if score.total > 0 {
					chapterScore = (score.correct * 100) / score.total
				}
				if chapterScore < 50 {
					allTestedChaptersPassed = false
					break
				}
			}
			// If chapter wasn't in test, we don't count it (it's OK to skip)
		}

		// Open section if all tested chapters passed and at least one chapter was tested
		if allTestedChaptersPassed && testedChaptersCount > 0 {
			openedSections[section.SectionID] = true
		} else {
			// Stop at first section that's not fully passed
			break
		}
	}

	// Convert to slice
	openedSectionsList := make([]string, 0, len(openedSections))
	for sectionID := range openedSections {
		openedSectionsList = append(openedSectionsList, sectionID)
	}

	// Save result (only if better than existing)
	err = s.AttemptRepo.SavePlacementTestResult(userID, overallScore, totalQuestions, openedSectionsList)
	if err != nil {
		s.logger.Error("failed to save placement test result", zap.Error(err))
	}

	// Save attempt record
	answersJSON, _ := json.Marshal(answers)
	resultsJSON, _ := json.Marshal([]interface{}{}) // Empty results for placement test

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
	}, nil
}

// PlacementTestResult represents placement test submission result
type PlacementTestResult struct {
	Score          int      `json:"score"`
	TotalQuestions int     `json:"total_questions"`
	Correct        int      `json:"correct"`
	OpenedSections []string `json:"opened_sections"`
}
