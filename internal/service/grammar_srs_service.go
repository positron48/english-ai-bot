package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type GrammarTrainingAvailability struct {
	Available           bool   `json:"available"`
	QuestionCount       int    `json:"question_count"`         // all questions in training pack (many per theory block); legacy/diagnostic
	TheoryBlockCount    int    `json:"theory_block_count"`     // SRS deck size — one spaced-repetition unit per theory block
	DueTheoryBlockCount int    `json:"due_theory_block_count"` // blocks with no row yet or next_review_at <= now
	BundleID            string `json:"bundle_id"`
	BundleLanguage      string `json:"bundle_language"`
}

type GrammarSrsSessionItem struct {
	Question map[string]interface{} `json:"question"`
}

type GrammarSrsSession struct {
	Items []GrammarSrsSessionItem `json:"items"`
}

type GrammarSrsAnswerResult struct {
	Correct       bool        `json:"correct"`
	CorrectAnswer interface{} `json:"correct_answer"`
	Explanation   interface{} `json:"explanation"`
}

func (s *GrammarService) GetGrammarTrainingAvailability(ctx context.Context, userID int64) (*GrammarTrainingAvailability, error) {
	out := &GrammarTrainingAvailability{
		Available:           false,
		QuestionCount:       0,
		TheoryBlockCount:    0,
		DueTheoryBlockCount: 0,
		BundleID:            s.learning.GrammarBundleID,
		BundleLanguage:      s.learning.TargetLang,
	}
	if s.TrainingPackRepo == nil {
		return out, nil
	}
	byBlock, err := s.TrainingPackRepo.QuestionsByTheoryBlock()
	if err != nil {
		return nil, err
	}
	if len(byBlock) == 0 {
		return out, nil
	}
	allowedByChapter, err := s.allowedTrainingChapters(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(allowedByChapter) == 0 {
		return out, nil
	}
	filtered := s.filterBlocksByAllowedChapters(byBlock, allowedByChapter)
	theoryBlockCount := len(filtered)
	totalQuestions := 0
	blockIDs := make([]string, 0, theoryBlockCount)
	for tbID, qs := range filtered {
		totalQuestions += len(qs)
		blockIDs = append(blockIDs, tbID)
	}
	out.Available = theoryBlockCount > 0
	out.QuestionCount = totalQuestions
	out.TheoryBlockCount = theoryBlockCount
	out.DueTheoryBlockCount = theoryBlockCount
	if s.SRSRepo != nil && userID != 0 && theoryBlockCount > 0 {
		if n, err := s.SRSRepo.CountDueOrNewTheoryBlocks(userID, s.learning.TargetLang, s.learning.GrammarBundleID, blockIDs, time.Now()); err == nil {
			out.DueTheoryBlockCount = n
		}
	}
	return out, nil
}

func (s *GrammarService) StartGrammarSrsSession(ctx context.Context, userID int64, limit int) (*GrammarSrsSession, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}
	if s.TrainingPackRepo == nil {
		return nil, fmt.Errorf("training pack repository is not configured")
	}
	byBlock, err := s.TrainingPackRepo.QuestionsByTheoryBlock()
	if err != nil {
		return nil, err
	}
	allowedByChapter, err := s.allowedTrainingChapters(ctx, userID)
	if err != nil {
		return nil, err
	}
	byBlock = s.filterBlocksByAllowedChapters(byBlock, allowedByChapter)
	if len(byBlock) == 0 {
		return &GrammarSrsSession{Items: []GrammarSrsSessionItem{}}, nil
	}

	// Ensure memory cards for all available theory blocks (idempotent).
	if s.SRSRepo != nil {
		for theoryBlockID := range byBlock {
			chapterID := ""
			conceptID := ""
			if s.TheoryIndex != nil {
				if info, ok := s.TheoryIndex.ByBlockID[theoryBlockID]; ok {
					chapterID = info.ChapterID
					conceptID = info.ConceptID
				}
			}
			_ = s.SRSRepo.EnsureTheoryMemory(userID, s.learning.TargetLang, s.learning.GrammarBundleID, chapterID, theoryBlockID, conceptID)
		}
	}

	selectedBlocks := make([]string, 0, limit)
	if s.SRSRepo != nil {
		due, err := s.SRSRepo.ListDueMemories(userID, s.learning.TargetLang, s.learning.GrammarBundleID, time.Now(), limit*3)
		if err == nil {
			for _, m := range due {
				if _, exists := byBlock[m.TheoryBlockID]; exists {
					selectedBlocks = append(selectedBlocks, m.TheoryBlockID)
				}
			}
			if len(selectedBlocks) > limit {
				selectedBlocks = selectedBlocks[:limit]
			}
		}
	}

	if len(selectedBlocks) < limit {
		remaining := make([]string, 0, len(byBlock))
		used := make(map[string]bool, len(selectedBlocks))
		for _, id := range selectedBlocks {
			used[id] = true
		}
		for theoryBlockID := range byBlock {
			if !used[theoryBlockID] {
				remaining = append(remaining, theoryBlockID)
			}
		}
		rand.Shuffle(len(remaining), func(i, j int) { remaining[i], remaining[j] = remaining[j], remaining[i] })
		selectedBlocks = append(selectedBlocks, remaining...)
		if len(selectedBlocks) > limit {
			selectedBlocks = selectedBlocks[:limit]
		}
	}

	items := make([]GrammarSrsSessionItem, 0, len(selectedBlocks))
	for _, theoryBlockID := range selectedBlocks {
		qs := byBlock[theoryBlockID]
		q := qs[rand.Intn(len(qs))]
		items = append(items, GrammarSrsSessionItem{Question: q})
	}
	return &GrammarSrsSession{Items: items}, nil
}

func (s *GrammarService) SubmitGrammarSrsAnswer(ctx context.Context, userID int64, questionID string, userAnswer interface{}) (*GrammarSrsAnswerResult, error) {
	if s.TrainingPackRepo == nil {
		return nil, fmt.Errorf("training pack repository is not configured")
	}
	all, err := s.TrainingPackRepo.GetAllQuestions()
	if err != nil {
		return nil, err
	}
	var question map[string]interface{}
	for _, q := range all {
		if id, _ := q["id"].(string); id == questionID {
			question = q
			break
		}
	}
	if question == nil {
		return nil, fmt.Errorf("question not found: %s", questionID)
	}
	chapterID, _ := question["chapter_id"].(string)
	allowedByChapter, err := s.allowedTrainingChapters(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !allowedByChapter[chapterID] {
		return nil, fmt.Errorf("question not found: %s", questionID)
	}
	correctAnswer := question["correct_answer"]
	isCorrect := s.compareAnswers(userAnswer, correctAnswer)

	theoryBlockID, _ := question["theory_block_id"].(string)
	conceptID, _ := question["concept_id"].(string)
	if s.SRSRepo != nil && theoryBlockID != "" {
		_ = s.updateTheoryMemory(userID, chapterID, theoryBlockID, conceptID, isCorrect)
		_ = s.SRSRepo.SaveAttempt(userID, s.learning.TargetLang, s.learning.GrammarBundleID, chapterID, theoryBlockID, conceptID, questionID, userAnswer, correctAnswer, isCorrect)
	}

	return &GrammarSrsAnswerResult{
		Correct:       isCorrect,
		CorrectAnswer: correctAnswer,
		Explanation:   question["explanation"],
	}, nil
}

func (s *GrammarService) allowedTrainingChapters(ctx context.Context, userID int64) (map[string]bool, error) {
	if userID == 0 {
		return map[string]bool{}, nil
	}
	sectionsData, err := s.ContentRepo.GetSections()
	if err != nil {
		return nil, fmt.Errorf("get sections for training availability: %w", err)
	}
	allowed := make(map[string]bool)
	for _, section := range sectionsData.Sections {
		openedByPlacement, err := s.isSectionOpenedByPlacement(ctx, userID, section.SectionID)
		if err != nil {
			return nil, err
		}
		if openedByPlacement {
			for _, chapterID := range section.ChapterIDs {
				if strings.TrimSpace(chapterID) != "" {
					allowed[chapterID] = true
				}
			}
			continue
		}
		for _, chapterID := range section.ChapterIDs {
			if strings.TrimSpace(chapterID) == "" {
				continue
			}
			progress, err := s.AttemptRepo.GetChapterProgress(userID, chapterID)
			if err != nil {
				return nil, fmt.Errorf("get chapter progress for %s: %w", chapterID, err)
			}
			if progress != nil && progress.Passed {
				allowed[chapterID] = true
			}
		}
	}
	return allowed, nil
}

func (s *GrammarService) filterBlocksByAllowedChapters(
	byBlock map[string][]map[string]interface{},
	allowedByChapter map[string]bool,
) map[string][]map[string]interface{} {
	if len(byBlock) == 0 || len(allowedByChapter) == 0 {
		return map[string][]map[string]interface{}{}
	}
	out := make(map[string][]map[string]interface{}, len(byBlock))
	for theoryBlockID, qs := range byBlock {
		filteredQuestions := make([]map[string]interface{}, 0, len(qs))
		for _, q := range qs {
			chapterID, _ := q["chapter_id"].(string)
			if allowedByChapter[chapterID] {
				filteredQuestions = append(filteredQuestions, q)
			}
		}
		if len(filteredQuestions) > 0 {
			out[theoryBlockID] = filteredQuestions
		}
	}
	return out
}

func (s *GrammarService) RecordGrammarTheoryAttemptFromTest(userID int64, question map[string]interface{}, isCorrect bool) {
	if s.SRSRepo == nil || question == nil {
		return
	}
	theoryBlockID, _ := question["theory_block_id"].(string)
	if theoryBlockID == "" {
		return
	}
	chapterID, _ := question["chapter_id"].(string)
	conceptID, _ := question["concept_id"].(string)
	_ = s.updateTheoryMemory(userID, chapterID, theoryBlockID, conceptID, isCorrect)
}

func (s *GrammarService) updateTheoryMemory(userID int64, chapterID, theoryBlockID, conceptID string, isCorrect bool) error {
	if s.SRSRepo == nil || theoryBlockID == "" {
		return nil
	}
	if chapterID == "" && s.TheoryIndex != nil {
		if info, ok := s.TheoryIndex.ByBlockID[theoryBlockID]; ok {
			chapterID = info.ChapterID
			if conceptID == "" {
				conceptID = info.ConceptID
			}
		}
	}
	if err := s.SRSRepo.EnsureTheoryMemory(userID, s.learning.TargetLang, s.learning.GrammarBundleID, chapterID, theoryBlockID, conceptID); err != nil {
		return err
	}
	memories, err := s.SRSRepo.ListDueMemories(userID, s.learning.TargetLang, s.learning.GrammarBundleID, time.Now().Add(365*24*time.Hour), 2000)
	if err != nil {
		return err
	}
	for _, m := range memories {
		if m.TheoryBlockID == theoryBlockID {
			return s.SRSRepo.UpdateAfterAnswer(m, isCorrect)
		}
	}
	return nil
}
