package service

import (
	"encoding/json"
	"math/rand"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// OptionsService generates multiple choice options for training
type OptionsService struct {
	trainingCardRepo *repository.TrainingCardRepository
	logger           *zap.Logger
	// targetLang is LEARNING_TARGET_LANG (e.g. en, es). English-specific "to " on verbs applies only when targetLang is en.
	targetLang string
}

// NewOptionsService creates a new options service.
// targetLang should match config learning target; empty defaults to "en" (tests / legacy callers).
func NewOptionsService(trainingCardRepo *repository.TrainingCardRepository, logger *zap.Logger, targetLang string) *OptionsService {
	tl := strings.ToLower(strings.TrimSpace(targetLang))
	if tl == "" {
		tl = "en"
	}
	return &OptionsService{
		trainingCardRepo: trainingCardRepo,
		logger:           logger,
		targetLang:       tl,
	}
}

// ForTargetLang returns an OptionsService scoped to targetLang (reuses receiver when unchanged).
func (s *OptionsService) ForTargetLang(targetLang string) *OptionsService {
	if s == nil {
		return nil
	}
	tl := strings.ToLower(strings.TrimSpace(targetLang))
	if tl == "" || tl == s.targetLang {
		return s
	}
	return NewOptionsService(s.trainingCardRepo, s.logger, tl)
}

// GenerateOptions generates multiple choice options for a card
// sessionWords: correct answers from other cards in the current session (to mix in as distractors)
// sessionWordENs: set of WordEN values from all cards in the session (to exclude distractors with matching English spelling)
// sessionWordRUs: set of WordRU values from all cards in the session (to exclude distractors with matching Russian spelling)
func (s *OptionsService) GenerateOptions(
	card *models.UserCardWithTraining,
	optionCount int,
	sessionWords []string,
	sessionWordENs map[string]bool,
	sessionWordRUs map[string]bool,
) ([]string, string, error) {
	var correctAnswer string
	var distractorsJSON string

	// Determine correct answer and distractors based on direction
	if card.UserCard.Direction == models.DirectionRUtoEN {
		// Question in Russian, answer in English
		// Use display_word if available (e.g., "to spy" for verbs), otherwise word_en
		if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
			correctAnswer = *card.TrainingCard.DisplayWord
		} else {
			correctAnswer = card.TrainingCard.WordEN
		}
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

	// Get POS of current card for filtering
	currentPOS := ""
	if card.TrainingCard.POS != nil && *card.TrainingCard.POS != "" {
		currentPOS = *card.TrainingCard.POS
	}

	// Get all meanings of this word to exclude them from distractors
	excludedMeanings := s.getOtherMeaningsOfWord(card.TrainingCard.WordCardID, card.UserCard.Direction)

	// Create a set of excluded words for fast lookup (includes correct answer and other meanings)
	// Normalize all excluded values to ensure we catch duplicates like "make" and "to make"
	excludedSet := make(map[string]bool)
	normalizedCorrectAnswer := s.normalizeVerbFormat(correctAnswer, currentPOS, card.UserCard.Direction)
	excludedSet[normalizedCorrectAnswer] = true
	// Also add original correct answer to catch both formats
	excludedSet[correctAnswer] = true
	for _, meaning := range excludedMeanings {
		normalizedMeaning := s.normalizeVerbFormat(meaning, currentPOS, card.UserCard.Direction)
		excludedSet[normalizedMeaning] = true
		// Also add original meaning to catch both formats
		excludedSet[meaning] = true
	}

	// Filter out other meanings of the same word, duplicates, and distractors that match English or Russian spelling of session words
	filteredDistractors := make([]string, 0, len(distractors))
	seenDistractors := make(map[string]bool)
	for _, d := range distractors {
		if !excludedSet[d] && !seenDistractors[d] {
			// Check if this distractor matches English or Russian spelling of any word in the session
			// For RU->EN direction, check if distractor (English word) matches WordEN of any session word
			// For EN->RU direction, check if distractor (Russian word) matches WordRU of any session word
			shouldExclude := false
			if card.UserCard.Direction == models.DirectionRUtoEN {
				// Remove "to " prefix for comparison if present
				lookupWord := strings.TrimPrefix(d, "to ")
				lookupWord = strings.TrimSpace(lookupWord)
				if sessionWordENs[lookupWord] || sessionWordENs[d] {
					shouldExclude = true
				}
			} else {
				if sessionWordRUs[d] {
					shouldExclude = true
				}
			}

			if !shouldExclude {
				filteredDistractors = append(filteredDistractors, d)
				seenDistractors[d] = true
			}
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
	// Also filter by POS if current card has POS
	// Note: recent correct answers should already be excluded by the caller (extractSessionWords)
	filteredSessionWords := make([]string, 0, len(sessionWords))
	seenSessionWords := make(map[string]bool)
	for _, sw := range sessionWords {
		if !excludedSet[sw] && !optionsPoolSet[sw] && !seenSessionWords[sw] {
			// Check POS match if current card has POS
			if currentPOS != "" {
				if !s.hasMatchingPOS(sw, currentPOS, card.UserCard.Direction) {
					continue
				}
			}
			filteredSessionWords = append(filteredSessionWords, sw)
			seenSessionWords[sw] = true
		}
	}

	// Select distractors (need optionCount - 1)
	neededDistractors := optionCount - 1
	selectedDistractors := make([]string, 0, neededDistractors)
	selectedDistractorsSet := make(map[string]bool) // Track selected distractors to avoid duplicates

	// STEP 1: First, take 1-2 distractors from the current card's distractors
	// This ensures that card-specific distractors are always used
	cardDistractorsToUse := 1
	if len(distractors) >= 2 && neededDistractors >= 3 {
		// Use 2 card distractors if we have enough and need at least 3 distractors
		cardDistractorsToUse = 2
	}
	if cardDistractorsToUse > len(distractors) {
		cardDistractorsToUse = len(distractors)
	}

	// Shuffle card distractors and take what we need
	shuffledCardDistractors := make([]string, len(distractors))
	copy(shuffledCardDistractors, distractors)
	rand.Shuffle(len(shuffledCardDistractors), func(i, j int) {
		shuffledCardDistractors[i], shuffledCardDistractors[j] = shuffledCardDistractors[j], shuffledCardDistractors[i]
	})

	// Add card distractors to selected distractors
	// Normalize verbs for RU->EN direction (add "to " if needed)
	for i := 0; i < cardDistractorsToUse && i < len(shuffledCardDistractors); i++ {
		word := shuffledCardDistractors[i]
		// Normalize verb format for RU->EN direction BEFORE checking duplicates
		normalizedWord := s.normalizeVerbFormat(word, currentPOS, card.UserCard.Direction)
		if !selectedDistractorsSet[normalizedWord] {
			selectedDistractors = append(selectedDistractors, normalizedWord)
			selectedDistractorsSet[normalizedWord] = true
		}
	}

	// STEP 2: Mix in 1-2 session words (familiar words from current training session)
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
	// Normalize verbs for RU->EN direction (add "to " if needed)
	for i := 0; i < sessionWordsToUse && i < len(shuffledSessionWords) && len(selectedDistractors) < neededDistractors; i++ {
		word := shuffledSessionWords[i]
		// Normalize verb format for RU->EN direction BEFORE checking duplicates
		normalizedWord := s.normalizeVerbFormat(word, currentPOS, card.UserCard.Direction)
		// Check both that it's not already selected AND that it doesn't match the correct answer
		if !selectedDistractorsSet[normalizedWord] && !excludedSet[normalizedWord] {
			selectedDistractors = append(selectedDistractors, normalizedWord)
			selectedDistractorsSet[normalizedWord] = true
		}
	}

	// STEP 3: Fill remaining slots with LLM-generated distractors (wrong answers + remaining card distractors)
	// Shuffle LLM distractors pool
	rand.Shuffle(len(optionsPool), func(i, j int) {
		optionsPool[i], optionsPool[j] = optionsPool[j], optionsPool[i]
	})

	// Add remaining options from pool (wrong answers + remaining card distractors)
	// Exclude already added distractors and ensure no duplicates
	// Filter by POS if current card has POS
	for _, d := range optionsPool {
		if !excludedSet[d] && len(selectedDistractors) < neededDistractors {
			// Check POS match if current card has POS
			if currentPOS != "" {
				if !s.hasMatchingPOS(d, currentPOS, card.UserCard.Direction) {
					continue
				}
			}
			// Normalize verb format for RU->EN direction BEFORE checking duplicates
			normalizedD := s.normalizeVerbFormat(d, currentPOS, card.UserCard.Direction)
			if !selectedDistractorsSet[normalizedD] {
				selectedDistractors = append(selectedDistractors, normalizedD)
				selectedDistractorsSet[normalizedD] = true
			}
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
		// Use fallbacks filtered by POS if current card has POS
		fallbacks := s.getFallbackDistractors(card.UserCard.Direction, currentPOS)
		for _, fb := range fallbacks {
			if !excludedSet[fb] {
				// Normalize verb format for RU->EN direction BEFORE checking duplicates
				normalizedFb := s.normalizeVerbFormat(fb, currentPOS, card.UserCard.Direction)
				if !selectedDistractorsSet[normalizedFb] {
					selectedDistractors = append(selectedDistractors, normalizedFb)
					selectedDistractorsSet[normalizedFb] = true
					if len(selectedDistractors) >= neededDistractors {
						break
					}
				}
			}
		}
	}

	// Build final options array with correct answer
	// Normalize correct answer for consistency (all options should be in the same format)
	normalizedCorrectAnswerForOptions := s.normalizeVerbFormat(correctAnswer, currentPOS, card.UserCard.Direction)
	options := make([]string, 0, optionCount)
	options = append(options, normalizedCorrectAnswerForOptions)
	options = append(options, selectedDistractors...)

	// Shuffle to randomize position
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	return options, normalizedCorrectAnswerForOptions, nil
}

// getFallbackDistractors returns generic fallback distractors filtered by POS
func (s *OptionsService) getFallbackDistractors(direction models.CardDirection, pos string) []string {
	if direction == models.DirectionRUtoEN {
		// English fallbacks
		verbs := []string{
			"make", "take", "get", "give", "come", "go", "see", "know",
			"think", "want", "look", "use", "find", "tell", "ask", "work",
			"call", "try", "need", "feel", "become", "leave", "put", "mean",
		}
		nouns := []string{
			"time", "person", "year", "way", "day", "thing", "man", "world",
			"life", "hand", "part", "child", "eye", "woman", "place", "work",
			"week", "case", "point", "government", "company", "number", "group",
		}
		adjectives := []string{
			"good", "new", "first", "last", "long", "great", "little", "own",
			"other", "old", "right", "big", "high", "small", "large", "next",
			"early", "young", "important", "few", "public", "bad", "same",
		}

		// Filter by POS
		switch pos {
		case "verb":
			return verbs
		case "noun":
			return nouns
		case "adjective":
			return adjectives
		default:
			// If no POS or unknown POS, return all
			return append(append(verbs, nouns...), adjectives...)
		}
	}
	// Russian fallbacks
	verbs := []string{
		"делать", "брать", "получать", "давать", "приходить", "идти",
		"видеть", "знать", "думать", "хотеть", "смотреть", "использовать",
		"находить", "говорить", "спрашивать", "работать", "звонить",
		"пытаться", "нуждаться", "чувствовать", "становиться", "покидать",
	}
	nouns := []string{
		"время", "человек", "год", "путь", "день", "вещь", "мужчина", "мир",
		"жизнь", "рука", "часть", "ребенок", "глаз", "женщина", "место", "работа",
		"неделя", "случай", "точка", "правительство", "компания", "число", "группа",
	}
	adjectives := []string{
		"хороший", "новый", "первый", "последний", "долгий", "великий", "маленький", "собственный",
		"другой", "старый", "правый", "большой", "высокий", "малый", "крупный", "следующий",
		"ранний", "молодой", "важный", "немного", "публичный", "плохой", "тот же",
	}

	// Filter by POS
	switch pos {
	case "verb":
		return verbs
	case "noun":
		return nouns
	case "adjective":
		return adjectives
	default:
		// If no POS or unknown POS, return all
		return append(append(verbs, nouns...), adjectives...)
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

// hasMatchingPOS checks if a word has the same POS as the current card
func (s *OptionsService) hasMatchingPOS(word string, targetPOS string, direction models.CardDirection) bool {
	if targetPOS == "" {
		return true // If no POS specified, accept all
	}

	// For RU->EN direction, search by word_en or display_word
	// For EN->RU direction, we can't easily determine POS from Russian word alone
	// So we'll be more lenient for EN->RU
	if direction == models.DirectionENtoRU {
		// For EN->RU, we can't easily check POS from Russian word
		// Accept it for now (could be improved by storing POS in word_cards)
		return true
	}

	// For RU->EN, try to find the word in training_cards
	// Remove "to " prefix if present for lookup
	lookupWord := strings.TrimPrefix(word, "to ")
	lookupWord = strings.TrimSpace(lookupWord)

	cards, err := s.trainingCardRepo.GetTrainingCardsByWordEN(lookupWord)
	if err != nil {
		s.logger.Debug("failed to get cards for POS check", zap.String("word", lookupWord), zap.Error(err))
		// If we can't find it, be lenient and accept it
		return true
	}

	// Check if any card has matching POS
	for _, card := range cards {
		if card.POS != nil && *card.POS == targetPOS {
			return true
		}
	}

	// If no cards found or no matching POS, reject
	return false
}

// normalizeVerbFormat keeps English infinitives as "to …" for RU->EN, and strips
// that English-only marker for Spanish/non-English RU->target cards.
func (s *OptionsService) normalizeVerbFormat(word string, pos string, direction models.CardDirection) string {
	if direction != models.DirectionRUtoEN {
		return word
	}
	if s.targetLang != "en" {
		return normalizeTargetVerbDisplay(s.targetLang, pos, word)
	}
	if !models.IsVerbPOS(pos) {
		return word
	}

	// Check if word already starts with "to "
	if strings.HasPrefix(word, "to ") {
		return word
	}

	// Add "to " prefix
	return "to " + word
}
