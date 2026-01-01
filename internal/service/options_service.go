package service

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// OptionsService generates multiple choice options for training
type OptionsService struct {
	trainingCardRepo *repository.TrainingCardRepository
	logger           *zap.Logger
}

// NewOptionsService creates a new options service
func NewOptionsService(trainingCardRepo *repository.TrainingCardRepository, logger *zap.Logger) *OptionsService {
	return &OptionsService{
		trainingCardRepo: trainingCardRepo,
		logger:           logger,
	}
}

// GenerateOptions generates multiple choice options for a card
// sessionWords: correct answers from other cards in the current session (to mix in as distractors)
func (s *OptionsService) GenerateOptions(
	card *models.UserCardWithTraining,
	optionCount int,
	sessionWords []string,
) ([]string, string, error) {
	var correctAnswer string
	var distractorsJSON string

	// Determine correct answer and distractors based on direction
	if card.UserCard.Direction == models.DirectionRUtoEN {
		// Question in Russian, answer in English
		correctAnswer = card.TrainingCard.WordEN
		distractorsJSON = card.TrainingCard.DistractorsEN
	} else {
		// Question in English, answer in Russian
		correctAnswer = card.TrainingCard.WordRU
		distractorsJSON = card.TrainingCard.DistractorsRU
	}

	// Parse distractors
	var distractors []string
	if distractorsJSON != "" {
		if err := json.Unmarshal([]byte(distractorsJSON), &distractors); err != nil {
			s.logger.Warn("failed to parse distractors", zap.Error(err))
			distractors = []string{}
		}
	}

	// Get all meanings of this word to exclude them from distractors
	excludedMeanings := s.getOtherMeaningsOfWord(card.TrainingCard.WordCardID, card.UserCard.Direction)
	
	// Create a set of excluded words for fast lookup (includes correct answer and other meanings)
	excludedSet := make(map[string]bool)
	excludedSet[correctAnswer] = true
	for _, meaning := range excludedMeanings {
		excludedSet[meaning] = true
	}
	
	// Filter out other meanings of the same word and duplicates
	filteredDistractors := make([]string, 0, len(distractors))
	seenDistractors := make(map[string]bool)
	for _, d := range distractors {
		if !excludedSet[d] && !seenDistractors[d] {
			filteredDistractors = append(filteredDistractors, d)
			seenDistractors[d] = true
		}
	}
	distractors = filteredDistractors

	// Parse user's wrong answers for personalization
	// Exclude duplicates and other meanings of the same word
	var wrongAnswers []string
	seenWrongAnswers := make(map[string]bool)
	if card.UserCard.WrongAnswersJSON != "" {
		type WrongAnswer struct {
			Option string `json:"option"`
		}
		var wrongs []WrongAnswer
		if err := json.Unmarshal([]byte(card.UserCard.WrongAnswersJSON), &wrongs); err == nil {
			for _, w := range wrongs {
				if !excludedSet[w.Option] && !seenWrongAnswers[w.Option] {
					wrongAnswers = append(wrongAnswers, w.Option)
					seenWrongAnswers[w.Option] = true
				}
			}
		}
	}

	// Build options pool: wrong answers first, then distractors
	// Use a map to track duplicates
	optionsPool := make([]string, 0, len(wrongAnswers)+len(distractors))
	optionsPoolSet := make(map[string]bool)
	
	// Add wrong answers (already deduplicated)
	for _, wa := range wrongAnswers {
		if !optionsPoolSet[wa] {
			optionsPool = append(optionsPool, wa)
			optionsPoolSet[wa] = true
		}
	}
	
	// Add distractors that aren't already in wrong answers
	for _, d := range distractors {
		if !optionsPoolSet[d] && !excludedSet[d] {
			optionsPool = append(optionsPool, d)
			optionsPoolSet[d] = true
		}
	}

	// Filter session words: exclude current card's answer, already included options, and other meanings of the same word
	// Note: recent correct answers should already be excluded by the caller (extractSessionWords)
	filteredSessionWords := make([]string, 0, len(sessionWords))
	seenSessionWords := make(map[string]bool)
	for _, sw := range sessionWords {
		if !excludedSet[sw] && !optionsPoolSet[sw] && !seenSessionWords[sw] {
			filteredSessionWords = append(filteredSessionWords, sw)
			seenSessionWords[sw] = true
		}
	}

	// Select distractors (need optionCount - 1)
	neededDistractors := optionCount - 1
	selectedDistractors := make([]string, 0, neededDistractors)
	selectedDistractorsSet := make(map[string]bool) // Track selected distractors to avoid duplicates

	// Mix in 1-2 session words (familiar words from current training session)
	// This prevents guessing by word recognition since all options look familiar
	sessionWordsToUse := 1
	if len(filteredSessionWords) >= 2 && neededDistractors >= 3 {
		// Use 2 session words if we have enough and need at least 3 distractors
		sessionWordsToUse = 2
	}
	if sessionWordsToUse > len(filteredSessionWords) {
		sessionWordsToUse = len(filteredSessionWords)
	}

	// Shuffle session words and take what we need
	shuffledSessionWords := make([]string, len(filteredSessionWords))
	copy(shuffledSessionWords, filteredSessionWords)
	rand.Shuffle(len(shuffledSessionWords), func(i, j int) {
		shuffledSessionWords[i], shuffledSessionWords[j] = shuffledSessionWords[j], shuffledSessionWords[i]
	})

	// Add session words to selected distractors
	for i := 0; i < sessionWordsToUse && i < len(shuffledSessionWords); i++ {
		word := shuffledSessionWords[i]
		if !selectedDistractorsSet[word] {
			selectedDistractors = append(selectedDistractors, word)
			selectedDistractorsSet[word] = true
		}
	}

	// Shuffle LLM distractors pool
	rand.Shuffle(len(optionsPool), func(i, j int) {
		optionsPool[i], optionsPool[j] = optionsPool[j], optionsPool[i]
	})

	// Fill remaining slots with LLM-generated distractors (wrong answers + distractors)
	// Exclude session words we already added and ensure no duplicates
	for _, d := range optionsPool {
		if !selectedDistractorsSet[d] && !excludedSet[d] && len(selectedDistractors) < neededDistractors {
			selectedDistractors = append(selectedDistractors, d)
			selectedDistractorsSet[d] = true
		}
		if len(selectedDistractors) >= neededDistractors {
			break
		}
	}

	// If we don't have enough distractors, try to get from other cards of the same word
	if len(selectedDistractors) < neededDistractors {
		s.logger.Debug("not enough distractors, using fallback",
			zap.Int("have", len(selectedDistractors)),
			zap.Int("need", neededDistractors),
		)
		// For now, just use generic fallbacks
		// In production, you might query similar words
		fallbacks := s.getFallbackDistractors(card.UserCard.Direction)
		for _, fb := range fallbacks {
			if !selectedDistractorsSet[fb] && !excludedSet[fb] {
				selectedDistractors = append(selectedDistractors, fb)
				selectedDistractorsSet[fb] = true
				if len(selectedDistractors) >= neededDistractors {
					break
				}
			}
		}
	}

	// Build final options array with correct answer
	options := make([]string, 0, optionCount)
	options = append(options, correctAnswer)
	options = append(options, selectedDistractors...)

	// Shuffle to randomize position
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	if len(options) < 2 {
		return nil, "", fmt.Errorf("not enough options generated")
	}

	return options, correctAnswer, nil
}

// getFallbackDistractors returns generic fallback distractors
func (s *OptionsService) getFallbackDistractors(direction models.CardDirection) []string {
	if direction == models.DirectionRUtoEN {
		// English fallbacks
		return []string{
			"make", "take", "get", "give", "come", "go", "see", "know",
			"think", "want", "look", "use", "find", "tell", "ask", "work",
			"call", "try", "need", "feel", "become", "leave", "put", "mean",
		}
	}
	// Russian fallbacks
	return []string{
		"делать", "брать", "получать", "давать", "приходить", "идти",
		"видеть", "знать", "думать", "хотеть", "смотреть", "использовать",
		"находить", "говорить", "спрашивать", "работать", "звонить",
		"пытаться", "нуждаться", "чувствовать", "становиться", "покидать",
	}
}

// getOtherMeaningsOfWord gets all meanings of the word to exclude from distractors
// Returns all correct answers from other training cards of the same word (WordCardID)
func (s *OptionsService) getOtherMeaningsOfWord(wordCardID int64, direction models.CardDirection) []string {
	// Get all training cards for this word
	cards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		s.logger.Warn("failed to get other meanings", zap.Error(err))
		return []string{}
	}

	meanings := make([]string, 0, len(cards))
	for _, c := range cards {
		if direction == models.DirectionRUtoEN {
			// For RU->EN, exclude all English word meanings from other cards
			// (same English word can have different Russian meanings)
			if c.WordEN != "" {
				meanings = append(meanings, c.WordEN)
			}
		} else {
			// For EN->RU, exclude all Russian meanings from other cards
			// (same English word can have different Russian translations)
			if c.WordRU != "" {
				meanings = append(meanings, c.WordRU)
			}
		}
	}

	return meanings
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

