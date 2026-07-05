package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// userCardRepoForWordSet is used by WordSetService for MarkKnown and EnsureUserCardsForWord (allows mocks in tests).
type userCardRepoForWordSet interface {
	CreateUserCard(card *models.UserCard) (int64, error)
	DeleteUserCardsByWordCardIDForUser(userID, wordCardID int64) (int64, error)
	CountUserCardsForWord(userID, wordCardID int64) (int, error)
}

// wordRepoForWordSet is used by WordSetService for word card operations (allows mocks in tests).
type wordRepoForWordSet interface {
	GetWordCardByLemmaForCourse(lemma, courseCode string) (*models.WordCard, error)
	GetWordCardByID(id int64) (*models.WordCard, error)
	SaveWordCard(word, content, courseCode string) error
	UpsertWordCardLemma(card *models.WordCard) (int64, error)
	UpsertWordFormMapping(form string, wordCardID int64) error
}

type courseAwareWordRepoForWordSet interface {
	TagWordCardCourse(wordCardID int64, courseCode string) error
}

// WordSetService handles word set business logic
type WordSetService struct {
	wordSetRepo           *repository.WordSetRepository
	wordSetCategoryRepo   *repository.WordSetCategoryRepository
	wordRepo              wordRepoForWordSet
	trainingCardRepo      *repository.TrainingCardRepository
	userCardRepo          userCardRepoForWordSet
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository
	userWordMasteringRepo *repository.UserWordMasteringRepository
	aiService             *ai.Service
	learning              config.LearningConfig
	courseCode            string
	modelHigh             string
	logger                *zap.Logger
}

// NewWordSetService creates a new word set service
func NewWordSetService(
	wordSetRepo *repository.WordSetRepository,
	wordSetCategoryRepo *repository.WordSetCategoryRepository,
	wordRepo wordRepoForWordSet,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo userCardRepoForWordSet,
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	modelHigh string,
	logger *zap.Logger,
) *WordSetService {
	return NewWordSetServiceWithMastering(wordSetRepo, wordSetCategoryRepo, wordRepo, trainingCardRepo, userCardRepo, userWordKnowledgeRepo, nil, aiService, learning, modelHigh, logger)
}

// NewWordSetServiceWithMastering creates a word set service with optional UserWordMasteringRepository for storing mastering score on card creation.
func NewWordSetServiceWithMastering(
	wordSetRepo *repository.WordSetRepository,
	wordSetCategoryRepo *repository.WordSetCategoryRepository,
	wordRepo wordRepoForWordSet,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo userCardRepoForWordSet,
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository,
	userWordMasteringRepo *repository.UserWordMasteringRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	modelHigh string,
	logger *zap.Logger,
) *WordSetService {
	return &WordSetService{
		wordSetRepo:           wordSetRepo,
		wordSetCategoryRepo:   wordSetCategoryRepo,
		wordRepo:              wordRepo,
		trainingCardRepo:      trainingCardRepo,
		userCardRepo:          userCardRepo,
		userWordKnowledgeRepo: userWordKnowledgeRepo,
		userWordMasteringRepo: userWordMasteringRepo,
		aiService:             aiService,
		learning:              learning,
		courseCode:            repository.CourseCodeForLearning(learning),
		modelHigh:             modelHigh,
		logger:                logger,
	}
}

// EnsureWordCardExistsMinimal ensures a word card exists with minimal data (only word)
// Returns the word card ID
// This is used for fast word set creation - word cards are created without LLM call
// Word cards will be filled asynchronously by TrainingWorker
func (s *WordSetService) EnsureWordCardExistsMinimal(word string) (int64, error) {
	return s.ensureWordCardExistsMinimalForCourse(word, s.courseCode)
}

func (s *WordSetService) ensureWordCardExistsMinimalForCourse(word, courseCode string) (int64, error) {
	normalizedWord := strings.TrimSpace(strings.ToLower(word))
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if normalizedWord == "" {
		return 0, fmt.Errorf("word is empty")
	}

	// Try to get existing word card
	wordCard, err := s.wordRepo.GetWordCardByLemmaForCourse(normalizedWord, courseCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get word card: %w", err)
	}

	if wordCard != nil {
		s.tryLinkVerbLemma(wordCard.ID, wordCard.Word)
		return wordCard.ID, nil
	}

	// Word not found, create minimal word card (only word, no LLM call)
	// Word card data will be filled asynchronously by TrainingWorker
	wordCardModel := &models.WordCard{
		Word:       normalizedWord,
		Definition: "", // Empty - will be filled later
		CourseCode: courseCode,
		// All other fields are nil - will be filled asynchronously
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCardModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create minimal word card: %w", err)
	}

	s.tryLinkVerbLemma(wordCardID, normalizedWord)

	s.logger.Debug("created minimal word card",
		zap.String("word", normalizedWord),
		zap.String("course_code", courseCode),
		zap.Int64("word_card_id", wordCardID),
	)

	return wordCardID, nil
}

// EnsureWordCardExists ensures a word card exists for a word
// Returns the word card ID
func (s *WordSetService) EnsureWordCardExists(ctx context.Context, word string) (int64, error) {
	normalizedWord := strings.TrimSpace(strings.ToLower(word))
	if normalizedWord == "" {
		return 0, fmt.Errorf("word is empty")
	}

	// Step 1: try word form mapping first (same lookup order as chat word flow),
	// but only when repository implementation supports this operation.
	type wordFormLookup interface {
		GetWordFormMappingForCourse(wordForm, courseCode string) (*models.WordForm, error)
	}
	if lookupRepo, ok := s.wordRepo.(wordFormLookup); ok {
		wordForm, err := lookupRepo.GetWordFormMappingForCourse(normalizedWord, s.courseCode)
		if err != nil {
			s.logger.Warn("failed to get word form mapping, fallback to lemma lookup",
				zap.String("word", normalizedWord),
				zap.Error(err),
			)
		}
		if err == nil && wordForm != nil {
			wordCard, err := s.wordRepo.GetWordCardByID(wordForm.WordCardID)
			if err != nil {
				return 0, fmt.Errorf("failed to get mapped word card: %w", err)
			}
			if wordCard != nil {
				s.tryLinkVerbLemma(wordCard.ID, wordCard.Word)
				return wordCard.ID, nil
			}
		}
	}

	// Step 2: direct lookup by lemma.
	wordCard, err := s.wordRepo.GetWordCardByLemmaForCourse(normalizedWord, s.courseCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get word card: %w", err)
	}

	if wordCard != nil {
		s.tryLinkVerbLemma(wordCard.ID, wordCard.Word)
		return wordCard.ID, nil
	}

	// Step 3: missing in DB -> create via AI.
	if s.aiService == nil {
		return 0, fmt.Errorf("AI service not available")
	}

	response, err := s.aiService.GenerateResponseForCourse(ctx, word, s.courseCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse JSON response
	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// Not JSON, save as legacy format
		if err := s.wordRepo.SaveWordCard(normalizedWord, response, s.courseCode); err != nil {
			return 0, fmt.Errorf("failed to save word card: %w", err)
		}
		// Get the saved word card
		wordCard, err := s.wordRepo.GetWordCardByLemmaForCourse(normalizedWord, s.courseCode)
		if err != nil {
			return 0, fmt.Errorf("failed to get saved word card: %w", err)
		}
		if wordCard == nil {
			return 0, fmt.Errorf("word card not found after save")
		}
		return wordCard.ID, nil
	}

	models.SyncWordInfoResponseNeutralAliases(&wordInfo)

	if !validDefinitionNativeForLearning(wordInfo.DefinitionRU, s.learning) {
		return 0, fmt.Errorf("word rejected by LLM: definition_native is not in expected language")
	}

	// Check for error from LLM
	if wordInfo.Error.IsTrue() {
		return 0, fmt.Errorf("word rejected by LLM: %s", wordInfo.Error.Message)
	}

	// Save structured word card (same logic as WordService)
	lemma := strings.ToLower(wordInfo.Lemma)
	if lemma == "" {
		lemma = normalizedWord
	}

	displayEN := lemma
	if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
		if strings.EqualFold(s.learning.TargetLang, "en") {
			displayEN = "to " + wordInfo.VerbForms.V1
		} else {
			displayEN = wordInfo.VerbForms.V1
		}
	}

	// Marshal examples and verb forms
	var examplesJSON *string
	if len(wordInfo.Examples) > 0 {
		examplesBytes, _ := json.Marshal(wordInfo.Examples)
		examplesStr := string(examplesBytes)
		examplesJSON = &examplesStr
	}

	var verbFormsJSON *string
	if wordInfo.VerbForms != nil {
		verbFormsBytes, _ := json.Marshal(wordInfo.VerbForms)
		verbFormsStr := string(verbFormsBytes)
		verbFormsJSON = &verbFormsStr
	}

	pos := wordInfo.POS
	transcription := wordInfo.Transcription
	definitionRU := wordInfo.DefinitionRU

	wordCardModel := &models.WordCard{
		Word:          lemma,
		Definition:    "", // Legacy field
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU:  &definitionRU,
		ExamplesJSON:  examplesJSON,
		VerbFormsJSON: verbFormsJSON,
		DisplayEN:     &displayEN,
		CourseCode:    s.courseCode,
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCardModel)
	if err != nil {
		return 0, fmt.Errorf("failed to save word card: %w", err)
	}
	s.tryLinkVerbLemma(wordCardID, lemma)

	// Get the saved word card
	wordCard, err = s.wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		return 0, fmt.Errorf("failed to get saved word card: %w", err)
	}
	if wordCard == nil {
		return 0, fmt.Errorf("word card not found after save")
	}

	// Create word form mapping if needed
	if normalizedWord != strings.ToLower(wordCard.Word) {
		if err := s.wordRepo.UpsertWordFormMapping(normalizedWord, wordCard.ID); err != nil {
			s.logger.Warn("failed to create word form mapping", zap.Error(err))
		}
	}

	return wordCard.ID, nil
}

// tryLinkVerbLemma links word_card to verb_lemmas when lemma exists in the Spanish conjugation dictionary (no POS gate).
func (s *WordSetService) tryLinkVerbLemma(wordCardID int64, lemma string) {
	if !strings.EqualFold(s.learning.TargetLang, "es") {
		return
	}
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	if lemma == "" {
		return
	}
	wr, ok := s.wordRepo.(*repository.WordRepository)
	if !ok {
		return
	}
	verbRepo := repository.NewVerbFormsRepository(wr.DB(), s.logger)
	_, err := verbRepo.LinkWordCardByLemma(wordCardID, lemma, s.learning.TargetLang, "word_set_service")
	if err != nil {
		s.logger.Warn("failed to link verb lemma for word set card",
			zap.Int64("word_card_id", wordCardID),
			zap.String("lemma", lemma),
			zap.Error(err))
	}
}

// EnsureTrainingCardsExist ensures training cards exist for a word card
// Uses the same logic as training worker
func (s *WordSetService) EnsureTrainingCardsExist(ctx context.Context, wordCardID int64) error {
	return s.ensureTrainingCardsExist(ctx, wordCardID, "")
}

func (s *WordSetService) ensureTrainingCardsExist(ctx context.Context, wordCardID int64, nativeLookupHint string) error {
	// Check if training cards already exist
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to check training cards: %w", err)
	}

	if len(trainingCards) > 0 {
		// Training cards already exist
		return nil
	}

	// Get word card
	wordCard, err := s.wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get word card: %w", err)
	}
	if wordCard == nil {
		return fmt.Errorf("word card not found")
	}

	// Generate training card via LLM
	if s.aiService == nil {
		return fmt.Errorf("AI service not available")
	}

	// Try to generate training card, first with default model, then with high model if validation fails
	var trainingResp models.TrainingCardResponse
	var response string
	var validationError string
	trainingContext := BuildTrainingCardNativeHintInstruction(nativeLookupHint, s.learning.NativeLang)
	generateTraining := func(modelOverride ...string) (string, error) {
		if trainingContext != "" {
			return s.aiService.GenerateTrainingCardForCourseWithContext(ctx, wordCard.Word, wordCard.CourseCode, trainingContext, modelOverride...)
		}
		return s.aiService.GenerateTrainingCardForCourse(ctx, wordCard.Word, wordCard.CourseCode, modelOverride...)
	}

	// First attempt with default model
	response, err = generateTraining()
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	if err := json.Unmarshal([]byte(response), &trainingResp); err != nil {
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}

	models.SyncTrainingCardResponseNeutralAliases(&trainingResp)

	// Check for error from LLM
	if trainingResp.Error != "" {
		return fmt.Errorf("word rejected by LLM: %s", trainingResp.Error)
	}

	// Validate response
	if len(trainingResp.Senses) == 0 {
		return fmt.Errorf("no senses in LLM response")
	}

	// Validate distractors
	validationError = ValidateTrainingCardResponse(s.learning.TargetLang, wordCard, &trainingResp)
	if validationError != "" {
		// Validation failed - try with high model if available
		if s.modelHigh != "" {
			s.logger.Info("validation failed with default model, trying with high model",
				zap.String("word", wordCard.Word),
				zap.String("error", validationError),
				zap.String("high_model", s.modelHigh),
			)

			// Try with high model
			response, err = generateTraining(s.modelHigh)
			if err != nil {
				s.logger.Warn("LLM generation with high model failed, using original validation error",
					zap.String("word", wordCard.Word),
					zap.Error(err),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			// Parse response from high model
			var highTrainingResp models.TrainingCardResponse
			if err := json.Unmarshal([]byte(response), &highTrainingResp); err != nil {
				s.logger.Warn("failed to parse LLM response from high model, using original validation error",
					zap.String("word", wordCard.Word),
					zap.Error(err),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			models.SyncTrainingCardResponseNeutralAliases(&highTrainingResp)

			// Check for error from LLM
			if highTrainingResp.Error != "" {
				return fmt.Errorf("word rejected by high model LLM: %s", highTrainingResp.Error)
			}

			// Validate response from high model
			if len(highTrainingResp.Senses) == 0 {
				s.logger.Warn("no senses in high model LLM response, using original validation error",
					zap.String("word", wordCard.Word),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			// Validate distractors from high model
			highValidationError := ValidateTrainingCardResponse(s.learning.TargetLang, wordCard, &highTrainingResp)
			if highValidationError == "" {
				// High model validation passed - use this response
				s.logger.Info("validation passed with high model",
					zap.String("word", wordCard.Word),
					zap.String("high_model", s.modelHigh),
				)
				trainingResp = highTrainingResp
				// Continue with creating training cards using high model response
			} else {
				s.logger.Warn("validation also failed with high model",
					zap.String("word", wordCard.Word),
					zap.String("error", highValidationError),
				)
				return fmt.Errorf("validation failed: %s", highValidationError)
			}
		} else {
			// High model not available, return original validation error
			return fmt.Errorf("validation failed: %s", validationError)
		}
	}

	// Create training cards
	for i, sense := range trainingResp.Senses {
		// Get POS
		pos := sense.POS
		if pos == "" && wordCard.POS != nil {
			pos = *wordCard.POS
		}

		// Marshal distractors
		distractorsRU, _ := json.Marshal(sense.DistractorsRU)
		distractorsTarget := normalizeTargetVerbDisplays(s.learning.TargetLang, pos, sense.DistractorsEN)
		distractorsEN, _ := json.Marshal(distractorsTarget)

		// Determine display_word
		displayWord := trainingResp.WordEN
		if sense.DisplayWord != "" {
			displayWord = sense.DisplayWord
		}
		displayWord = normalizeTargetVerbDisplay(s.learning.TargetLang, pos, displayWord)

		trainingCard := &models.TrainingCard{
			WordCardID:    wordCardID,
			WordEN:        displayWord,
			Transcription: trainingResp.Transcription,
			SenseIndex:    i,
			WordRU:        sense.WordRU,
			MeaningEN:     sense.MeaningEN,
			ExampleEN:     sense.ExampleEN,
			ExampleRU:     sense.ExampleRU,
			DistractorsRU: string(distractorsRU),
			DistractorsEN: string(distractorsEN),
			Hint:          sense.Hint,
		}

		if pos != "" {
			trainingCard.POS = &pos
		}
		if displayWord != "" {
			trainingCard.DisplayWord = &displayWord
		}

		_, err := s.trainingCardRepo.CreateTrainingCard(trainingCard)
		if err != nil {
			return fmt.Errorf("failed to create training card: %w", err)
		}
	}

	return nil
}

// EnsureCardsForWord ensures both training cards and per-direction user_cards exist for
// wordCardID. Equivalent to calling EnsureTrainingCardsExist followed by EnsureUserCardsForWord,
// but on the common fast path (training cards already generated) it fetches them once instead of
// twice — this is the hot path hit on every word-card open, so the extra round trip matters.
//
// Matches the existing caller convention (internal/web reading.go / vocab.go): a training-card
// generation failure is soft — logged and swallowed, leaving trainingCards empty — so callers
// still get their usual "word/card not found" response instead of a hard 500 on every transient
// LLM hiccup. Only a failure creating user_cards for cards that DO exist is a hard error.
func (s *WordSetService) EnsureCardsForWord(ctx context.Context, userID, wordCardID int64) error {
	return s.EnsureCardsForWordWithNativeHint(ctx, userID, wordCardID, "")
}

// EnsureCardsForWordWithNativeHint is like EnsureCardsForWord but passes the learner's original
// native-language lookup token into training-card generation (e.g. RU "капуста" for ES "col").
func (s *WordSetService) EnsureCardsForWordWithNativeHint(ctx context.Context, userID, wordCardID int64, nativeLookupHint string) error {
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to check training cards: %w", err)
	}
	if len(trainingCards) == 0 {
		if err := s.ensureTrainingCardsExist(ctx, wordCardID, nativeLookupHint); err != nil {
			s.logger.Warn("failed to ensure training cards", zap.Int64("word_card_id", wordCardID), zap.Error(err))
		} else if trainingCards, err = s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID); err != nil {
			return fmt.Errorf("failed to reload training cards after generation: %w", err)
		}
	}
	return s.ensureUserCardsForTrainingCards(userID, wordCardID, trainingCards)
}

// EnsureUserCardsForWord creates user_cards for all training_cards of a word
// Similar to WordService.ensureUserCardsForWord
func (s *WordSetService) EnsureUserCardsForWord(userID, wordCardID int64) error {
	// Get all training cards for this word
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get training cards: %w", err)
	}
	return s.ensureUserCardsForTrainingCards(userID, wordCardID, trainingCards)
}

// ensureUserCardsForTrainingCards is EnsureUserCardsForWord given an already-fetched
// trainingCards list, so callers that just fetched it (EnsureCardsForWord) don't pay for a second
// identical query.
func (s *WordSetService) ensureUserCardsForTrainingCards(userID, wordCardID int64, trainingCards []*models.TrainingCard) error {
	if len(trainingCards) == 0 {
		return nil
	}

	// Fast path: skip the per-card create-or-noop loop (up to 2 SELECT-then-INSERT round trips
	// per training card) when the user already has every (training_card, direction) pair — the
	// common case on every re-open of an already-learned word.
	if existing, err := s.userCardRepo.CountUserCardsForWord(userID, wordCardID); err == nil && existing == len(trainingCards)*2 {
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
		s.logger.Info("created user cards for word",
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

// MarkKnown marks a word as known and removes user_cards
func (s *WordSetService) MarkKnown(userID, wordCardID int64) error {
	// Mark as known
	if err := s.userWordKnowledgeRepo.MarkKnown(userID, wordCardID); err != nil {
		return fmt.Errorf("failed to mark as known: %w", err)
	}

	// Delete user_cards for this word
	_, err := s.userCardRepo.DeleteUserCardsByWordCardIDForUser(userID, wordCardID)
	if err != nil {
		s.logger.Warn("failed to delete user cards",
			zap.Int64("user_id", userID),
			zap.Int64("word_card_id", wordCardID),
			zap.Error(err),
		)
		// Don't fail if deletion fails, knowledge is already marked
	}

	return nil
}

// ProcessWordSetItems processes a comma-separated list of words for a word set
// Creates word_cards with minimal data (only word) and creates word_set_items
// Word card data and training cards are created asynchronously by TrainingWorker
// This allows saving word sets quickly even with many words
func (s *WordSetService) ProcessWordSetItems(ctx context.Context, wordSetID int64, wordsStr string) error {
	return s.ProcessWordSetItemsForCourse(ctx, wordSetID, "", wordsStr)
}

// ProcessWordSetItemsForCourse replaces a set's items within its declared course.
func (s *WordSetService) ProcessWordSetItemsForCourse(ctx context.Context, wordSetID int64, courseCode, wordsStr string) error {
	if courseCode != "" {
		wordSet, err := s.wordSetRepo.GetWordSetForCourse(wordSetID, courseCode)
		if err != nil {
			return fmt.Errorf("failed to get word set: %w", err)
		}
		if wordSet == nil {
			return fmt.Errorf("word set not found for course %s", courseCode)
		}
	}

	// Split by comma and normalize
	words := strings.Split(wordsStr, ",")
	var wordCardIDs []int64

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Ensure word card exists with minimal data (no LLM call)
		// Word card data will be filled asynchronously by TrainingWorker
		wordCardID, err := s.ensureWordCardExistsMinimalForCourse(word, courseCode)
		if err != nil {
			s.logger.Warn("failed to ensure minimal word card exists",
				zap.String("word", word),
				zap.Error(err),
			)
			// Continue with other words
			continue
		}
		if taggedRepo, ok := s.wordRepo.(courseAwareWordRepoForWordSet); ok {
			if err := taggedRepo.TagWordCardCourse(wordCardID, courseCode); err != nil {
				s.logger.Warn("failed to tag word card with course",
					zap.Int64("word_card_id", wordCardID),
					zap.String("course_code", courseCode),
					zap.Error(err),
				)
			}
		}

		// Word card data and training cards will be created asynchronously by TrainingWorker
		// This allows saving word sets quickly even with many words

		wordCardIDs = append(wordCardIDs, wordCardID)
	}

	// Remove duplicates
	seen := make(map[int64]bool)
	var uniqueWordCardIDs []int64
	for _, id := range wordCardIDs {
		if !seen[id] {
			seen[id] = true
			uniqueWordCardIDs = append(uniqueWordCardIDs, id)
		}
	}

	// Set word set items, validating that every linked word card belongs to the set's course.
	if err := s.wordSetRepo.SetWordSetItemsForCourse(wordSetID, courseCode, uniqueWordCardIDs); err != nil {
		return fmt.Errorf("failed to set word set items: %w", err)
	}

	return nil
}
