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

		// Count published chapters in this section
		publishedChapters := 0
		passedChapters := 0
		for _, chapterID := range section.ChapterIDs {
			chapterItem, _ := s.PublishRepo.GetPublishedItem("chapter", chapterID)
			if chapterItem.IsPublished {
				publishedChapters++
				progress, _ := s.AttemptRepo.GetChapterProgress(userID, chapterID)
				if progress.Passed {
					passedChapters++
				}
			}
		}

		result = append(result, &SectionWithProgress{
			Section:         section,
			Title:           title,
			PublishedChapters: publishedChapters,
			PassedChapters:  passedChapters,
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

	// Remove correct_answer from selected questions
	for _, q := range selectedQuestions {
		if qMap, ok := q.(map[string]interface{}); ok {
			delete(qMap, "correct_answer")
		}
	}

	return &TestQuestions{
		Questions: selectedQuestions,
		Total:     numQuestions,
	}, nil
}

// TestQuestions represents test questions
type TestQuestions struct {
	Questions []interface{}
	Total     int
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
