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
		correctAnswer = card.TrainingCard.MeaningRU
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
	
	// Filter out other meanings of the same word
	filteredDistractors := make([]string, 0, len(distractors))
	for _, d := range distractors {
		if !contains(excludedMeanings, d) && d != correctAnswer {
			filteredDistractors = append(filteredDistractors, d)
		}
	}
	distractors = filteredDistractors

	// Parse user's wrong answers for personalization
	var wrongAnswers []string
	if card.UserCard.WrongAnswersJSON != "" {
		type WrongAnswer struct {
			Option string `json:"option"`
		}
		var wrongs []WrongAnswer
		if err := json.Unmarshal([]byte(card.UserCard.WrongAnswersJSON), &wrongs); err == nil {
			for _, w := range wrongs {
				if w.Option != correctAnswer {
					wrongAnswers = append(wrongAnswers, w.Option)
				}
			}
		}
	}

	// Build options pool: wrong answers first, then distractors
	optionsPool := make([]string, 0, len(wrongAnswers)+len(distractors))
	optionsPool = append(optionsPool, wrongAnswers...)
	
	// Add distractors that aren't already in wrong answers
	for _, d := range distractors {
		if !contains(optionsPool, d) && d != correctAnswer {
			optionsPool = append(optionsPool, d)
		}
	}

	// Filter session words: exclude current card's answer and already included options
	filteredSessionWords := make([]string, 0, len(sessionWords))
	for _, sw := range sessionWords {
		if sw != correctAnswer && !contains(optionsPool, sw) {
			filteredSessionWords = append(filteredSessionWords, sw)
		}
	}

	// Select distractors (need optionCount - 1)
	neededDistractors := optionCount - 1
	selectedDistractors := make([]string, 0, neededDistractors)

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
		selectedDistractors = append(selectedDistractors, shuffledSessionWords[i])
	}

	// Shuffle LLM distractors pool
	rand.Shuffle(len(optionsPool), func(i, j int) {
		optionsPool[i], optionsPool[j] = optionsPool[j], optionsPool[i]
	})

	// Fill remaining slots with LLM-generated distractors (wrong answers + distractors)
	// Exclude session words we already added
	for _, d := range optionsPool {
		if !contains(selectedDistractors, d) && len(selectedDistractors) < neededDistractors {
			selectedDistractors = append(selectedDistractors, d)
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
			if !contains(selectedDistractors, fb) && fb != correctAnswer {
				selectedDistractors = append(selectedDistractors, fb)
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
			// For RU->EN, exclude English word meanings (but we use the same word, so skip)
			// Not applicable here as we're translating to the same word
			continue
		} else {
			// For EN->RU, exclude all Russian meanings
			if c.MeaningRU != "" {
				meanings = append(meanings, c.MeaningRU)
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

