package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/utils"

	"go.uber.org/zap"
)

// WordService handles word-related business logic
type WordService struct {
	wordRepo              *repository.WordRepository
	trainingCardRepo      *repository.TrainingCardRepository
	userCardRepo          *repository.UserCardRepository
	userWordMasteringRepo *repository.UserWordMasteringRepository
	pronunciationService  *PronunciationService
	aiService             *ai.Service
	learning              config.LearningConfig
	logger                *zap.Logger
}

// NewWordService creates a new word service
func NewWordService(
	wordRepo *repository.WordRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	logger *zap.Logger,
) *WordService {
	return NewWordServiceWithMastering(wordRepo, trainingCardRepo, userCardRepo, nil, aiService, learning, logger)
}

// NewWordServiceWithMastering creates a word service with optional UserWordMasteringRepository for storing mastering score on card creation.
func NewWordServiceWithMastering(
	wordRepo *repository.WordRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	userWordMasteringRepo *repository.UserWordMasteringRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	logger *zap.Logger,
) *WordService {
	return &WordService{
		wordRepo:              wordRepo,
		trainingCardRepo:      trainingCardRepo,
		userCardRepo:          userCardRepo,
		userWordMasteringRepo: userWordMasteringRepo,
		aiService:             aiService,
		learning:              learning,
		logger:                logger,
	}
}

// SetPronunciationService connects background pronunciation prefetch to word creation/lookups.
func (s *WordService) SetPronunciationService(pronunciationService *PronunciationService) {
	s.pronunciationService = pronunciationService
}

func (s *WordService) typoOrInvalidWordMessage() string {
	if s.learning.TargetLang == "en" {
		return "💡 Это слово, скорее всего, опечатка или несуществующее английское слово."
	}
	return "💡 Это слово, скорее всего, опечатка или несуществующее слово."
}

// IsSingleWord checks if the input is a single word
func (s *WordService) IsSingleWord(text string) bool {
	trimmed := strings.TrimSpace(text)
	// Remove common punctuation
	trimmed = strings.Trim(trimmed, ".,!?;:()[]{}'\"")
	trimmed = strings.TrimSpace(trimmed)

	// Check if it's a single word (no spaces)
	parts := strings.Fields(trimmed)
	if len(parts) != 1 || len(trimmed) == 0 {
		return false
	}

	// Single-word dictionary lookup is only for target-language tokens in Latin script.
	// Cyrillic words should go through normal translation/correction routing.
	hasLatin := false
	for _, r := range trimmed {
		if unicode.IsDigit(r) {
			return false
		}
		if unicode.Is(unicode.Cyrillic, r) {
			return false
		}
		if unicode.Is(unicode.Latin, r) {
			hasLatin = true
		}
	}
	return hasLatin
}

func buildForcedSingleWordLookupPrompt(word string) string {
	return fmt.Sprintf(`SINGLE-WORD LOOKUP ONLY.
Return ONLY valid JSON object (no markdown, no explanations, no prose).
Use this exact schema and fill all required fields for a valid existing word:
{
  "error": false,
  "hint": "",
  "input_word": "%s",
  "lemma": "",
  "pos": "",
  "transcription": "",
  "definition_native": "",
  "examples": [
    { "example_target": "", "gloss_native": "" },
    { "example_target": "", "gloss_native": "" }
  ]
}
If the token is clearly nonsense, return:
{"error": true, "hint": "..." , "input_word":"%s", "lemma":"", "pos":"", "transcription":"", "definition_native":"", "examples":[]}
Word: %s`, word, word, word)
}

func buildForcedNativeLanguagePrompt(word, nativeLang, targetLang, pair string) string {
	return fmt.Sprintf(`SINGLE-WORD LOOKUP ONLY for pair %s (%s -> %s).
Return ONLY valid JSON object (no markdown, no prose).
Critical language constraint:
- definition_native MUST be in %s.
- every gloss_native in examples MUST be in %s.
- example_target MUST stay in %s.
Word: %s`, pair, nativeLang, targetLang, nativeLang, nativeLang, targetLang, word)
}

func (s *WordService) nativeLanguageRequiresCyrillic() bool {
	// Keep strict native-language guard for Spanish deployment where drift was observed.
	// Do not enforce globally for all pairs to preserve legacy EN tests/data behavior.
	return strings.EqualFold(s.learning.NativeLang, "ru") && strings.EqualFold(s.learning.TargetLang, "es")
}

func hasCyrillicText(text string) bool {
	return ContainsCyrillic(strings.TrimSpace(text))
}

func nativeFieldsLookValidForConfig(definitionNative string, examples []models.WordInfoExample, requireCyrillic bool) bool {
	if !requireCyrillic {
		return true
	}
	if !hasCyrillicText(definitionNative) {
		return false
	}
	for _, ex := range examples {
		gloss := strings.TrimSpace(ex.GlossNative)
		if gloss == "" {
			gloss = strings.TrimSpace(ex.GlossRU)
		}
		if gloss != "" && !hasCyrillicText(gloss) {
			return false
		}
	}
	return true
}

func (s *WordService) wordCardNativeFieldsLookValid(card *models.WordCard) bool {
	if card == nil {
		return false
	}
	definition := ""
	if card.DefinitionNative != nil {
		definition = strings.TrimSpace(*card.DefinitionNative)
	}
	if definition == "" && card.DefinitionRU != nil {
		definition = strings.TrimSpace(*card.DefinitionRU)
	}
	var examples []models.WordInfoExample
	if card.ExamplesJSON != nil && strings.TrimSpace(*card.ExamplesJSON) != "" {
		_ = json.Unmarshal([]byte(*card.ExamplesJSON), &examples)
	}
	return nativeFieldsLookValidForConfig(definition, examples, s.nativeLanguageRequiresCyrillic())
}

// NormalizeWord normalizes a word for storage/lookup
func (s *WordService) NormalizeWord(word string) string {
	return strings.TrimSpace(strings.ToLower(word))
}

// GetWordDefinition gets word definition from DB or AI
// New flow: resolve word form -> lemma -> render markdown from structured data
func (s *WordService) GetWordDefinition(ctx context.Context, userID int64, word string) (string, error) {
	normalizedWord := s.NormalizeWord(word)
	inputWord := word // Keep original for history

	// Step 1: Try to resolve word form to lemma via word_forms table
	wordForm, err := s.wordRepo.GetWordFormMapping(normalizedWord)
	if err != nil {
		s.logger.Warn("failed to get word form mapping", zap.Error(err))
	}

	var wordCard *models.WordCard
	if wordForm != nil {
		// Found mapping, get lemma
		wordCard, err = s.wordRepo.GetWordCardByID(wordForm.WordCardID)
		if err != nil {
			s.logger.Warn("failed to get word card by ID", zap.Error(err))
		}
	}

	// Step 2: If no mapping found, try direct lookup by lemma
	if wordCard == nil {
		wordCard, err = s.wordRepo.GetWordCardByLemma(normalizedWord)
		if err != nil {
			s.logger.Warn("failed to get word card by lemma", zap.Error(err))
		}
	}

	// Step 3: If found in DB, render markdown and return
	if wordCard != nil {
		if !s.wordCardNativeFieldsLookValid(wordCard) {
			s.logger.Warn("word card from database has invalid native-language fields, forcing refresh from AI",
				zap.String("word", wordCard.Word),
				zap.Int64("word_card_id", wordCard.ID),
				zap.String("native_lang", s.learning.NativeLang),
			)
			wordCard = nil
		}
	}
	if wordCard != nil {
		s.logger.Info("word found in database",
			zap.String("input", inputWord),
			zap.String("lemma", wordCard.Word),
		)

		// Record request history
		wordCardID := wordCard.ID
		if err := s.wordRepo.AddWordRequestHistoryWithCard(userID, inputWord, &wordCardID, nil); err != nil {
			s.logger.Warn("failed to add word request history", zap.Error(err))
		}

		// Create user_cards for existing training_cards if they exist
		// This ensures that when a user requests a word that already has training cards,
		// they get linked to the user for training
		if s.trainingCardRepo != nil && s.userCardRepo != nil {
			if err := s.ensureUserCardsForWord(userID, wordCardID); err != nil {
				s.logger.Warn("failed to ensure user cards for word",
					zap.Int64("user_id", userID),
					zap.Int64("word_card_id", wordCardID),
					zap.Error(err),
				)
				// Don't fail the request if user cards creation fails
			}
		}

		// Render markdown from structured data
		if s.pronunciationService != nil {
			// Pronunciation is canonical per word (lemma), not per training-card display form.
			s.pronunciationService.ScheduleWord(wordCard.Word)
		}
		markdown := s.renderWordCardMarkdown(wordCard)
		return markdown, nil
	}

	// Step 4: Word not found, request from AI (expecting JSON)
	s.logger.Info("word not found in database, requesting from AI",
		zap.String("word", normalizedWord),
	)

	if s.aiService == nil {
		return "", fmt.Errorf("AI service not available")
	}

	response, err := s.aiService.GenerateResponse(ctx, word)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	// Cyrillic input: don't save to DB, just return LLM response to chat
	if ContainsCyrillic(normalizedWord) {
		s.logger.Info("word contains Cyrillic, returning AI response without saving",
			zap.String("word", normalizedWord),
			zap.Int64("user_id", userID),
		)
		return response, nil
	}

	// Parse JSON response
	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// LLM can occasionally return correction text for single-word lookup.
		// Retry once with a strict JSON-only lookup instruction.
		forcedPrompt := buildForcedSingleWordLookupPrompt(word)
		forcedResponse, forcedErr := s.aiService.GenerateResponse(ctx, forcedPrompt)
		if forcedErr == nil {
			if retryErr := json.Unmarshal([]byte(forcedResponse), &wordInfo); retryErr == nil {
				response = forcedResponse
			} else {
				// Not JSON after retry, keep legacy fallback below.
				s.logger.Warn("forced single-word lookup response is not JSON",
					zap.Error(retryErr),
				)
			}
		} else {
			s.logger.Warn("forced single-word lookup retry failed",
				zap.Error(forcedErr),
			)
		}
	}
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// Not JSON, might be old format - try to save as-is for backward compatibility
		s.logger.Warn("failed to parse AI response as JSON, saving as legacy format",
			zap.Error(err),
			zap.String("response", response[:min(100, len(response))]),
		)
		if err := s.wordRepo.SaveWordCard(normalizedWord, response); err != nil {
			s.logger.Warn("failed to save word card", zap.Error(err))
		} else {
			if err := s.wordRepo.AddWordRequestHistory(userID, normalizedWord); err != nil {
				s.logger.Warn("failed to add word request history", zap.Error(err))
			}
		}
		return response, nil
	}

	models.SyncWordInfoResponseNeutralAliases(&wordInfo)

	requireNativeCyrillic := s.nativeLanguageRequiresCyrillic()
	if !nativeFieldsLookValidForConfig(wordInfo.DefinitionNative, wordInfo.Examples, requireNativeCyrillic) {
		forcedPrompt := buildForcedNativeLanguagePrompt(word, s.learning.NativeLang, s.learning.TargetLang, s.learning.Pair)
		forcedResponse, forcedErr := s.aiService.GenerateResponse(ctx, forcedPrompt)
		if forcedErr == nil {
			var forcedInfo models.WordInfoResponse
			if err := json.Unmarshal([]byte(forcedResponse), &forcedInfo); err == nil {
				models.SyncWordInfoResponseNeutralAliases(&forcedInfo)
				if nativeFieldsLookValidForConfig(forcedInfo.DefinitionNative, forcedInfo.Examples, requireNativeCyrillic) {
					wordInfo = forcedInfo
				}
			}
		}
	}
	if !nativeFieldsLookValidForConfig(wordInfo.DefinitionNative, wordInfo.Examples, requireNativeCyrillic) {
		s.logger.Warn("LLM response rejected: native fields are not in expected language",
			zap.String("word", normalizedWord),
			zap.String("native_lang", s.learning.NativeLang),
			zap.String("target_lang", s.learning.TargetLang),
		)
		return "⚠️ Не удалось получить карточку со значением на русском. Попробуйте ещё раз.", nil
	}

	// Check for error from LLM
	// If definition_ru is present, ignore error field - we have valid data
	hasDefinitionRU := strings.TrimSpace(wordInfo.DefinitionRU) != ""

	// Priority 1: If error is true (bool or string "true"), check hint first
	// BUT skip if we have definition_ru (valid data)
	if !hasDefinitionRU && wordInfo.Error.IsTrue() {
		hint := strings.TrimSpace(wordInfo.Hint)
		if hint != "" {
			// If hint is present, return only hint
			message := fmt.Sprintf("💡 %s", hint)
			return message, nil
		} else {
			// If hint is empty, return default error message
			return s.typoOrInvalidWordMessage(), nil
		}
	}

	// Priority 3: Legacy handling for string error messages
	// LLM sometimes puts non-error strings in error field (like "load", "master", "none", "valid English word")
	// Skip this check if we have definition_ru (valid data)
	if !hasDefinitionRU {
		errorMsg := strings.TrimSpace(wordInfo.Error.Message)
		hasValidData := wordInfo.Lemma != "" && wordInfo.POS != ""

		// List of known non-error strings that LLM sometimes puts in error field
		nonErrorStrings := []string{
			"null", "none", "false", "no",
			"load", "master", "slave", "tor", "corm",
			"valid english word", "valid English word",
		}

		isNonErrorString := false
		errorMsgLower := strings.ToLower(errorMsg)
		for _, nonError := range nonErrorStrings {
			if errorMsgLower == strings.ToLower(nonError) {
				isNonErrorString = true
				break
			}
		}

		// Treat as error if: error field is not empty AND it's not a known non-error string AND we don't have valid data
		if errorMsg != "" && !isNonErrorString && !hasValidData {
			// Check hint first
			hint := strings.TrimSpace(wordInfo.Hint)
			if hint != "" {
				// If hint is present, return only hint
				message := fmt.Sprintf("💡 %s", hint)
				return message, nil
			} else {
				// If hint is empty, return default error message
				return s.typoOrInvalidWordMessage(), nil
			}
		}
	}

	// Step 5: Check if we have valid data (definition_ru is required for valid word card)
	// If no valid data, don't save to database
	// hasDefinitionRU already declared above, just check it again
	if !hasDefinitionRU {
		s.logger.Info("LLM did not return valid word card (no definition_ru), not saving to database",
			zap.String("word", normalizedWord),
			zap.Int64("user_id", userID),
		)
		// Return hint if available, otherwise default message
		hint := strings.TrimSpace(wordInfo.Hint)
		if hint != "" {
			return fmt.Sprintf("💡 %s", hint), nil
		}
		return s.typoOrInvalidWordMessage(), nil
	}

	// Step 6: Save structured data to word_cards (lemma)
	lemma := strings.ToLower(wordInfo.Lemma)
	if lemma == "" {
		lemma = normalizedWord
	}
	if strings.Contains(lemma, ",") {
		s.logger.Warn("lemma contains comma-separated values, using normalized input word",
			zap.String("lemma", lemma),
			zap.String("input", normalizedWord),
		)
		lemma = normalizedWord
	}

	displayEN := lemma
	if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
		if strings.EqualFold(s.learning.TargetLang, "en") {
			displayEN = "to " + wordInfo.VerbForms.V1
		} else {
			// Spanish and other targets: infinitive without English "to " prefix
			displayEN = wordInfo.VerbForms.V1
		}
	}

	wordCard = &models.WordCard{
		Word:          lemma,
		Definition:    "", // Legacy field, keep empty
		POS:           &wordInfo.POS,
		Transcription: &wordInfo.Transcription,
		DefinitionRU:  &wordInfo.DefinitionRU,
		DisplayEN:     &displayEN,
	}

	// Serialize examples
	if len(wordInfo.Examples) > 0 {
		examplesJSON, _ := json.Marshal(wordInfo.Examples)
		examplesStr := string(examplesJSON)
		wordCard.ExamplesJSON = &examplesStr
	}

	// Serialize verb forms
	if wordInfo.VerbForms != nil {
		verbFormsJSON, _ := json.Marshal(wordInfo.VerbForms)
		verbFormsStr := string(verbFormsJSON)
		wordCard.VerbFormsJSON = &verbFormsStr
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		s.logger.Warn("failed to save word card lemma", zap.Error(err))
		return "", fmt.Errorf("failed to save word card: %w", err)
	}

	// Step 7: Create word form mappings
	// Map input word to lemma
	if normalizedWord != lemma {
		if err := s.wordRepo.UpsertWordFormMapping(normalizedWord, wordCardID); err != nil {
			s.logger.Warn("failed to create word form mapping", zap.Error(err))
		}
	}

	// Map lemma to itself
	if err := s.wordRepo.UpsertWordFormMapping(lemma, wordCardID); err != nil {
		s.logger.Warn("failed to create lemma mapping", zap.Error(err))
	}

	// Map verb forms if present
	if wordInfo.VerbForms != nil {
		forms := []string{wordInfo.VerbForms.V1, wordInfo.VerbForms.V2, wordInfo.VerbForms.V3}
		if wordInfo.VerbForms.Gerund != "" {
			forms = append(forms, wordInfo.VerbForms.Gerund)
		}
		if wordInfo.VerbForms.ThirdPerson != "" {
			forms = append(forms, wordInfo.VerbForms.ThirdPerson)
		}
		for _, form := range forms {
			if form != "" && strings.ToLower(form) != lemma {
				if err := s.wordRepo.UpsertWordFormMapping(strings.ToLower(form), wordCardID); err != nil {
					s.logger.Warn("failed to create verb form mapping", zap.String("form", form), zap.Error(err))
				}
			}
		}
	}

	// Step 8: Record request history
	if err := s.wordRepo.AddWordRequestHistoryWithCard(userID, inputWord, &wordCardID, nil); err != nil {
		s.logger.Warn("failed to add word request history", zap.Error(err))
	}

	// Step 9: Render and return markdown
	if s.pronunciationService != nil {
		// Pronunciation is canonical per word (lemma), not per training-card display form.
		s.pronunciationService.ScheduleWord(lemma)
	}
	markdown := s.renderWordCardMarkdown(wordCard)
	return markdown, nil
}

// renderWordCardMarkdown renders markdown from structured WordCard data
func (s *WordService) renderWordCardMarkdown(card *models.WordCard) string {
	var examples []models.WordInfoExample
	var verbForms *models.WordInfoVerbForms

	// Parse examples
	if card.ExamplesJSON != nil && *card.ExamplesJSON != "" {
		if err := json.Unmarshal([]byte(*card.ExamplesJSON), &examples); err != nil {
			s.logger.Warn("failed to parse examples JSON", zap.Error(err))
		}
	}

	// Parse verb forms
	if card.VerbFormsJSON != nil && *card.VerbFormsJSON != "" {
		var vf models.WordInfoVerbForms
		if err := json.Unmarshal([]byte(*card.VerbFormsJSON), &vf); err != nil {
			s.logger.Warn("failed to parse verb forms JSON", zap.Error(err))
		} else {
			verbForms = &vf
		}
	}

	return utils.RenderWordCardMarkdownForTarget(card, examples, verbForms, s.learning.TargetLang)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetWordCard retrieves a word card from database
func (s *WordService) GetWordCard(word string) (*models.WordCard, error) {
	return s.wordRepo.GetWordCard(word)
}

// ensureUserCardsForWord creates user_cards for all training_cards of a word if they don't exist
// This is called when a user requests a word that already exists in the database with training cards
func (s *WordService) ensureUserCardsForWord(userID, wordCardID int64) error {
	// Get all training cards for this word
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get training cards: %w", err)
	}

	if len(trainingCards) == 0 {
		// No training cards yet, nothing to do
		return nil
	}

	// Create user_cards for each training card (both directions)
	createdCount := 0
	for _, trainingCard := range trainingCards {
		// Create ru_en card
		ruEnCard := &models.UserCard{
			UserID:         userID,
			TrainingCardID: trainingCard.ID,
			Direction:      models.DirectionRUtoEN,
			State:          models.StateNew,
			EF:             models.InitialEF,
		}
		if _, err := s.userCardRepo.CreateUserCard(ruEnCard); err != nil {
			s.logger.Warn("failed to create ru_en user card",
				zap.Int64("user_id", userID),
				zap.Int64("training_card_id", trainingCard.ID),
				zap.Error(err),
			)
		} else {
			createdCount++
		}

		// Create en_ru card
		enRuCard := &models.UserCard{
			UserID:         userID,
			TrainingCardID: trainingCard.ID,
			Direction:      models.DirectionENtoRU,
			State:          models.StateNew,
			EF:             models.InitialEF,
		}
		if _, err := s.userCardRepo.CreateUserCard(enRuCard); err != nil {
			s.logger.Warn("failed to create en_ru user card",
				zap.Int64("user_id", userID),
				zap.Int64("training_card_id", trainingCard.ID),
				zap.Error(err),
			)
		} else {
			createdCount++
		}
	}

	if createdCount > 0 {
		s.logger.Info("created user cards for existing word",
			zap.Int64("user_id", userID),
			zap.Int64("word_card_id", wordCardID),
			zap.Int("training_cards", len(trainingCards)),
			zap.Int("user_cards_created", createdCount),
		)
		if s.userWordMasteringRepo != nil {
			if err := s.userWordMasteringRepo.Upsert(userID, wordCardID, 0); err != nil {
				s.logger.Warn("failed to upsert initial mastering score", zap.Error(err))
			}
		}
	}

	return nil
}
