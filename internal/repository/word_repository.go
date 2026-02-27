package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// WordRepository handles database operations for word cards
type WordRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWordRepository creates a new word repository
func NewWordRepository(db *sql.DB, logger *zap.Logger) *WordRepository {
	return &WordRepository{
		db:     db,
		logger: logger,
	}
}

func (r *WordRepository) DB() *sql.DB {
	return r.db
}

// GetWordCard retrieves a word card by word (backward compatibility - searches by lemma)
func (r *WordRepository) GetWordCard(word string) (*models.WordCard, error) {
	return r.GetWordCardByLemma(word)
}

// GetWordCardByID retrieves a word card by ID
func (r *WordRepository) GetWordCardByID(id int64) (*models.WordCard, error) {
	query := `SELECT id, word, definition, pos, transcription, definition_ru, 
			  examples_json, verb_forms_json, display_en, 
			  COALESCE(CAST(processed_at AS TEXT), '') as processed_at,
			  COALESCE(processing_error, '') as processing_error,
			  CAST(created_at AS TEXT) as created_at,
			  CAST(updated_at AS TEXT) as updated_at
			  FROM word_cards 
			  WHERE id = ?`

	var card models.WordCard
	var createdAt, updatedAt, processedAtStr, processingErrorStr string
	var pos, transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.Word,
		&card.Definition,
		&pos,
		&transcription,
		&definitionRU,
		&examplesJSON,
		&verbFormsJSON,
		&displayEN,
		&processedAtStr,
		&processingErrorStr,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	if pos.Valid {
		card.POS = &pos.String
	}
	if transcription.Valid {
		card.Transcription = &transcription.String
	}
	if definitionRU.Valid {
		card.DefinitionRU = &definitionRU.String
	}
	if examplesJSON.Valid {
		card.ExamplesJSON = &examplesJSON.String
	}
	if verbFormsJSON.Valid {
		card.VerbFormsJSON = &verbFormsJSON.String
	}
	if displayEN.Valid {
		card.DisplayEN = &displayEN.String
	}
	if processedAtStr != "" {
		processedAt, _ := time.Parse("2006-01-02 15:04:05", processedAtStr)
		card.ProcessedAt = &processedAt
	}
	if processingErrorStr != "" {
		card.ProcessingError = &processingErrorStr
	}

	return &card, nil
}

// GetWordCardByLemma retrieves a word card by lemma (base form)
func (r *WordRepository) GetWordCardByLemma(lemma string) (*models.WordCard, error) {
	query := `SELECT id, word, definition, pos, transcription, definition_ru, 
			  examples_json, verb_forms_json, display_en,
			  COALESCE(CAST(processed_at AS TEXT), '') as processed_at,
			  COALESCE(processing_error, '') as processing_error,
			  CAST(created_at AS TEXT) as created_at,
			  CAST(updated_at AS TEXT) as updated_at
			  FROM word_cards 
			  WHERE LOWER(word) = LOWER(?)`

	var card models.WordCard
	var createdAt, updatedAt, processedAtStr, processingErrorStr string
	var pos, transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN sql.NullString

	err := r.db.QueryRow(query, lemma).Scan(
		&card.ID,
		&card.Word,
		&card.Definition,
		&pos,
		&transcription,
		&definitionRU,
		&examplesJSON,
		&verbFormsJSON,
		&displayEN,
		&processedAtStr,
		&processingErrorStr,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	if pos.Valid {
		card.POS = &pos.String
	}
	if transcription.Valid {
		card.Transcription = &transcription.String
	}
	if definitionRU.Valid {
		card.DefinitionRU = &definitionRU.String
	}
	if examplesJSON.Valid {
		card.ExamplesJSON = &examplesJSON.String
	}
	if verbFormsJSON.Valid {
		card.VerbFormsJSON = &verbFormsJSON.String
	}
	if displayEN.Valid {
		card.DisplayEN = &displayEN.String
	}
	if processedAtStr != "" {
		processedAt, _ := time.Parse("2006-01-02 15:04:05", processedAtStr)
		card.ProcessedAt = &processedAt
	}
	if processingErrorStr != "" {
		card.ProcessingError = &processingErrorStr
	}

	return &card, nil
}

// SaveWordCard saves a new word card or updates existing one (backward compatibility)
func (r *WordRepository) SaveWordCard(word, definition string) error {
	query := `INSERT INTO word_cards (word, definition, updated_at) 
			  VALUES (?, ?, CURRENT_TIMESTAMP)
			  ON CONFLICT(word) DO UPDATE SET 
			  	definition = excluded.definition,
			  	updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, word, definition)
	if err != nil {
		return fmt.Errorf("failed to save word card: %w", err)
	}

	r.logger.Debug("word card saved",
		zap.String("word", word),
	)

	return nil
}

// UpsertWordCardLemma saves or updates a word card (lemma) with structured data
func (r *WordRepository) UpsertWordCardLemma(card *models.WordCard) (int64, error) {
	query := `INSERT INTO word_cards (
		word, definition, pos, transcription, definition_ru, 
		examples_json, verb_forms_json, display_en, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(word) DO UPDATE SET 
		definition = COALESCE(excluded.definition, word_cards.definition),
		pos = COALESCE(excluded.pos, word_cards.pos),
		transcription = COALESCE(excluded.transcription, word_cards.transcription),
		definition_ru = COALESCE(excluded.definition_ru, word_cards.definition_ru),
		examples_json = COALESCE(excluded.examples_json, word_cards.examples_json),
		verb_forms_json = COALESCE(excluded.verb_forms_json, word_cards.verb_forms_json),
		display_en = COALESCE(excluded.display_en, word_cards.display_en),
		updated_at = CURRENT_TIMESTAMP`

	result, err := r.db.Exec(query,
		card.Word,
		card.Definition,
		card.POS,
		card.Transcription,
		card.DefinitionRU,
		card.ExamplesJSON,
		card.VerbFormsJSON,
		card.DisplayEN,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert word card: %w", err)
	}

	id, err := result.LastInsertId()
	// LastInsertId() can return 0 or error when ON CONFLICT triggers UPDATE instead of INSERT
	// In that case, we need to get the ID by querying the database
	if err != nil || id == 0 {
		// Get ID by lemma (works for both INSERT and UPDATE cases)
		existing, err := r.GetWordCardByLemma(card.Word)
		if err != nil {
			return 0, fmt.Errorf("failed to get word card ID: %w", err)
		}
		if existing == nil {
			return 0, fmt.Errorf("word card not found after upsert")
		}
		id = existing.ID
	}

	r.logger.Debug("word card lemma upserted",
		zap.Int64("id", id),
		zap.String("lemma", card.Word),
	)

	return id, nil
}

// GetWordFormMapping retrieves word_card_id for a given word form
func (r *WordRepository) GetWordFormMapping(form string) (*models.WordForm, error) {
	query := `SELECT form, word_card_id FROM word_forms WHERE LOWER(form) = LOWER(?)`

	var wf models.WordForm
	err := r.db.QueryRow(query, form).Scan(&wf.Form, &wf.WordCardID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word form mapping: %w", err)
	}

	return &wf, nil
}

// UpsertWordFormMapping creates or updates a mapping from word form to lemma
func (r *WordRepository) UpsertWordFormMapping(form string, wordCardID int64) error {
	query := `INSERT INTO word_forms (form, word_card_id) 
			  VALUES (?, ?)
			  ON CONFLICT(form) DO UPDATE SET 
			  	word_card_id = excluded.word_card_id`

	_, err := r.db.Exec(query, strings.ToLower(form), wordCardID)
	if err != nil {
		return fmt.Errorf("failed to upsert word form mapping: %w", err)
	}

	r.logger.Debug("word form mapping upserted",
		zap.String("form", form),
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}

// AddWordRequestHistory adds a history entry for a word request (backward compatibility)
func (r *WordRepository) AddWordRequestHistory(userID int64, word string) error {
	return r.AddWordRequestHistoryWithCard(userID, word, nil, nil)
}

// AddWordRequestHistoryWithCard adds a history entry with word_card_id and input_word
func (r *WordRepository) AddWordRequestHistoryWithCard(userID int64, inputWord string, wordCardID *int64, word *string) error {
	query := `INSERT INTO word_request_history (user_id, word, word_card_id, input_word, requested_at) 
			  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := r.db.Exec(query, userID, word, wordCardID, inputWord)
	if err != nil {
		return fmt.Errorf("failed to add word request history: %w", err)
	}

	r.logger.Debug("word request history added",
		zap.Int64("user_id", userID),
		zap.String("input_word", inputWord),
		zap.Int64p("word_card_id", wordCardID),
	)

	return nil
}

// GetUserIDsByWord gets telegram user IDs who requested a specific word (by lemma or input_word)
func (r *WordRepository) GetUserIDsByWord(word string) ([]int64, error) {
	// Try to find by word_card_id first (if word is a lemma)
	wordCard, err := r.GetWordCardByLemma(word)
	if err != nil {
		return nil, fmt.Errorf("failed to get word card: %w", err)
	}

	var query string
	var args []interface{}

	if wordCard != nil {
		// Search by word_card_id
		query = `SELECT DISTINCT user_id FROM word_request_history 
				 WHERE word_card_id = ? OR LOWER(input_word) = LOWER(?) OR LOWER(word) = LOWER(?)`
		args = []interface{}{wordCard.ID, word, word}
	} else {
		// Search by input_word or legacy word field
		query = `SELECT DISTINCT user_id FROM word_request_history 
				 WHERE LOWER(input_word) = LOWER(?) OR LOWER(word) = LOWER(?)`
		args = []interface{}{word, word}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// ListPronunciationCandidates returns recent distinct canonical words (lemmas)
// suitable for pronunciation prefetch.
func (r *WordRepository) ListPronunciationCandidates(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 5000 {
		limit = 5000
	}

	query := `
		SELECT wc.word AS candidate
		FROM word_cards wc
		WHERE wc.word IS NOT NULL AND wc.word <> ''
		ORDER BY wc.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit*3)
	if err != nil {
		return nil, fmt.Errorf("failed to list pronunciation candidates: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]struct{}, limit)
	candidates := make([]string, 0, limit)
	for rows.Next() {
		var candidate string
		if err := rows.Scan(&candidate); err != nil {
			return nil, fmt.Errorf("failed to scan pronunciation candidate: %w", err)
		}
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		key := strings.ToLower(candidate)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
		if len(candidates) >= limit {
			break
		}
	}

	return candidates, nil
}

// MarkWordCardProcessedError marks a word card as processed with an error
func (r *WordRepository) MarkWordCardProcessedError(wordCardID int64, errorText string) error {
	query := `UPDATE word_cards 
			  SET processed_at = CURRENT_TIMESTAMP, 
			      processing_error = ?,
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, errorText, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to mark word card as processed with error: %w", err)
	}

	r.logger.Debug("word card marked as processed with error",
		zap.Int64("word_card_id", wordCardID),
		zap.String("error", errorText),
	)

	return nil
}

// ResetWordCardProcessed resets the processed status of a word card
func (r *WordRepository) ResetWordCardProcessed(wordCardID int64) error {
	query := `UPDATE word_cards 
			  SET processed_at = NULL, 
			      processing_error = NULL,
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to reset word card processed status: %w", err)
	}

	r.logger.Debug("word card processed status reset",
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}

// UpdateWordCardDefinition updates the definition of a word card
func (r *WordRepository) UpdateWordCardDefinition(wordCardID int64, definition string) error {
	query := `UPDATE word_cards 
			  SET definition = ?, 
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, definition, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to update word card definition: %w", err)
	}

	r.logger.Debug("word card definition updated",
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}

// UpdateWordCard updates all fields of a word card
func (r *WordRepository) UpdateWordCard(card *models.WordCard) error {
	query := `UPDATE word_cards 
			  SET word = ?, definition = ?, pos = ?, transcription = ?, 
			      definition_ru = ?, examples_json = ?, verb_forms_json = ?, 
			      display_en = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query,
		card.Word,
		card.Definition,
		card.POS,
		card.Transcription,
		card.DefinitionRU,
		card.ExamplesJSON,
		card.VerbFormsJSON,
		card.DisplayEN,
		card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update word card: %w", err)
	}

	r.logger.Debug("word card updated",
		zap.Int64("word_card_id", card.ID),
		zap.String("word", card.Word),
	)

	return nil
}

// WordCardAdminItem represents a word card with additional info for admin view
type WordCardAdminItem struct {
	models.WordCard
	HasTrainingCards bool
	RequestingUsers  []int64
	TTSState         *string
	TTSError         *string
	TTSAudioURL      *string
}

// ListWordCardsAdmin lists word cards for admin view with optional filters.
// missingTrainingPOS: when set, only words that have no training card with this part of speech are returned (e.g. "noun" => words without a noun card).
func (r *WordRepository) ListWordCardsAdmin(filterUserID *int64, onlyWithErrors bool, hasAudio *bool, searchQuery string, missingTrainingPOS string, limit, offset int, sortBy string, sortOrder string) ([]*WordCardAdminItem, error) {
	// Use LEFT JOIN with GROUP BY to check for training cards - more reliable than subquery
	query := `SELECT wc.id, wc.word, wc.definition,
			  COALESCE(wc.pos, '') as pos,
			  COALESCE(wc.transcription, '') as transcription,
			  COALESCE(wc.definition_ru, '') as definition_ru,
			  COALESCE(wc.examples_json, '') as examples_json,
			  COALESCE(wc.verb_forms_json, '') as verb_forms_json,
			  COALESCE(wc.display_en, '') as display_en,
			  COALESCE(CAST(wc.processed_at AS TEXT), '') as processed_at,
			  COALESCE(wc.processing_error, '') as processing_error,
			  COALESCE(tts.state, '') as tts_state,
			  COALESCE(tts.last_error_message, '') as tts_error,
			  COALESCE(tts.audio_rel_path, '') as tts_audio_rel_path,
			  CAST(wc.created_at AS TEXT) as created_at,
			  CAST(wc.updated_at AS TEXT) as updated_at,
			  MAX(CASE WHEN tc.id IS NOT NULL THEN 1 ELSE 0 END) as has_training_cards
			  FROM word_cards wc
			  LEFT JOIN training_cards tc ON tc.word_card_id = wc.id
			  LEFT JOIN tts_generation_status tts ON LOWER(tts.word) = LOWER(wc.word)`

	args := []interface{}{}
	conditions := []string{}

	// Filter by user if specified - use subquery to avoid duplicates from JOIN
	if filterUserID != nil {
		conditions = append(conditions, "wc.id IN (SELECT DISTINCT COALESCE(word_card_id, (SELECT id FROM word_cards WHERE LOWER(word) = LOWER(word_request_history.word))) FROM word_request_history WHERE user_id = ?)")
		args = append(args, *filterUserID)
	}

	// Filter by errors if specified
	if onlyWithErrors {
		conditions = append(conditions, "((wc.processing_error IS NOT NULL AND wc.processing_error != '') OR tts.state IN ('failed_retryable','failed_terminal'))")
	}

	// Filter by audio existence if specified
	if hasAudio != nil {
		if *hasAudio {
			conditions = append(conditions, "tts.audio_rel_path IS NOT NULL AND tts.audio_rel_path != ''")
		} else {
			conditions = append(conditions, "(tts.audio_rel_path IS NULL OR tts.audio_rel_path = '')")
		}
	}

	// Filter: only words that have no training card with the given part of speech
	if missingTrainingPOS != "" {
		conditions = append(conditions, "wc.id NOT IN (SELECT word_card_id FROM training_cards WHERE pos = ?)")
		args = append(args, missingTrainingPOS)
	}

	// Filter by search query if specified (by word or translation) - case-insensitive
	if searchQuery != "" {
		searchLower := strings.ToLower(searchQuery)
		conditions = append(conditions, "(LOWER(wc.word) LIKE ? OR LOWER(wc.definition_ru) LIKE ?)")
		searchPattern := "%" + searchLower + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY wc.id, wc.word, wc.definition, wc.pos, wc.transcription, wc.definition_ru, wc.examples_json, wc.verb_forms_json, wc.display_en, wc.processed_at, wc.processing_error, tts.state, tts.last_error_message, tts.audio_rel_path, wc.created_at, wc.updated_at"

	// Build ORDER BY clause
	orderBy := "wc.created_at"

	// Validate and set sort column
	validSortColumns := map[string]string{
		"id":        "wc.id",
		"word":      "LOWER(wc.word)",
		"pos":       "wc.pos",
		"has_cards": "MAX(CASE WHEN tc.id IS NOT NULL THEN 1 ELSE 0 END)",
	}

	if sortBy != "" {
		if column, ok := validSortColumns[sortBy]; ok {
			orderBy = column
		}
	}

	// Validate sort order
	var orderDir string
	switch sortOrder {
	case "asc":
		orderDir = "ASC"
	case "desc":
		orderDir = "DESC"
	default:
		orderDir = "DESC"
	}

	query += " ORDER BY " + orderBy + " " + orderDir + " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list word cards: %w", err)
	}
	defer rows.Close()

	var items []*WordCardAdminItem
	for rows.Next() {
		var item WordCardAdminItem
		var createdAt, updatedAt, processedAtStr, processingErrorStr, ttsStateStr, ttsErrorStr, ttsAudioRelPath string
		var posStr, transcriptionStr, definitionRUStr, examplesJSONStr, verbFormsJSONStr, displayENStr string
		var hasTrainingCards int

		err := rows.Scan(&item.ID, &item.Word, &item.Definition,
			&posStr, &transcriptionStr, &definitionRUStr, &examplesJSONStr, &verbFormsJSONStr, &displayENStr,
			&processedAtStr, &processingErrorStr,
			&ttsStateStr, &ttsErrorStr, &ttsAudioRelPath,
			&createdAt, &updatedAt, &hasTrainingCards)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word card: %w", err)
		}

		// Set optional fields
		if posStr != "" {
			item.POS = &posStr
		}
		if transcriptionStr != "" {
			item.Transcription = &transcriptionStr
		}
		if definitionRUStr != "" {
			item.DefinitionRU = &definitionRUStr
		}
		if examplesJSONStr != "" {
			item.ExamplesJSON = &examplesJSONStr
		}
		if verbFormsJSONStr != "" {
			item.VerbFormsJSON = &verbFormsJSONStr
		}
		if displayENStr != "" {
			item.DisplayEN = &displayENStr
		}

		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		item.HasTrainingCards = hasTrainingCards > 0

		if processedAtStr != "" {
			processedAt, _ := time.Parse("2006-01-02 15:04:05", processedAtStr)
			item.ProcessedAt = &processedAt
		}
		if processingErrorStr != "" {
			item.ProcessingError = &processingErrorStr
		}
		if ttsStateStr != "" {
			item.TTSState = &ttsStateStr
		}
		if ttsErrorStr != "" {
			item.TTSError = &ttsErrorStr
		}
		if ttsAudioRelPath != "" {
			audioURL := "/media/tts/" + strings.TrimLeft(ttsAudioRelPath, "/")
			item.TTSAudioURL = &audioURL
		}

		// Get requesting users for this word
		userIDs, err := r.GetUserIDsByWord(item.Word)
		if err != nil {
			r.logger.Warn("failed to get requesting users", zap.Error(err), zap.String("word", item.Word))
		} else {
			item.RequestingUsers = userIDs
		}

		items = append(items, &item)
	}

	return items, nil
}

// CountWordCardsAdmin counts total word cards matching filters (for pagination).
// missingTrainingPOS: when set, only words that have no training card with this part of speech are counted.
func (r *WordRepository) CountWordCardsAdmin(filterUserID *int64, onlyWithErrors bool, hasAudio *bool, searchQuery string, missingTrainingPOS string) (int, error) {
	query := `SELECT COUNT(DISTINCT wc.id) FROM word_cards wc
			  LEFT JOIN tts_generation_status tts ON LOWER(tts.word) = LOWER(wc.word)`

	args := []interface{}{}
	conditions := []string{}

	// Filter by user if specified
	if filterUserID != nil {
		conditions = append(conditions, "wc.word IN (SELECT DISTINCT word FROM word_request_history WHERE user_id = ?)")
		args = append(args, *filterUserID)
	}

	// Filter by errors if specified
	if onlyWithErrors {
		conditions = append(conditions, "((wc.processing_error IS NOT NULL AND wc.processing_error != '') OR tts.state IN ('failed_retryable','failed_terminal'))")
	}

	// Filter by audio existence if specified
	if hasAudio != nil {
		if *hasAudio {
			conditions = append(conditions, "tts.audio_rel_path IS NOT NULL AND tts.audio_rel_path != ''")
		} else {
			conditions = append(conditions, "(tts.audio_rel_path IS NULL OR tts.audio_rel_path = '')")
		}
	}

	// Filter: only words that have no training card with the given part of speech
	if missingTrainingPOS != "" {
		conditions = append(conditions, "wc.id NOT IN (SELECT word_card_id FROM training_cards WHERE pos = ?)")
		args = append(args, missingTrainingPOS)
	}

	// Filter by search query if specified (by word or translation) - case-insensitive
	if searchQuery != "" {
		searchLower := strings.ToLower(searchQuery)
		conditions = append(conditions, "(LOWER(wc.word) LIKE ? OR LOWER(wc.definition_ru) LIKE ?)")
		searchPattern := "%" + searchLower + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count word cards: %w", err)
	}

	return count, nil
}

// GetWordCardRequestingUsers gets user IDs who requested a specific word card
func (r *WordRepository) GetWordCardRequestingUsers(wordCardID int64) ([]int64, error) {
	// First get the word
	var word string
	err := r.db.QueryRow(`SELECT word FROM word_cards WHERE id = ?`, wordCardID).Scan(&word)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	return r.GetUserIDsByWord(word)
}

// DeleteWordCard deletes a word card by ID and all related data
// Explicitly deletes training_cards and user_cards to ensure cleanup even if CASCADE doesn't work
func (r *WordRepository) DeleteWordCard(wordCardID int64) error {
	// Get word before deletion for logging
	var word string
	err := r.db.QueryRow("SELECT word FROM word_cards WHERE id = ?", wordCardID).Scan(&word)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("word card not found")
		}
		return fmt.Errorf("failed to get word: %w", err)
	}

	// Step 1: Delete user_cards that reference training_cards for this word_card_id
	// This must be done first to avoid foreign key violations
	deleteUserCardsQuery := `DELETE FROM user_cards 
		WHERE training_card_id IN (
			SELECT id FROM training_cards WHERE word_card_id = ?
		)`
	userCardsResult, err := r.db.Exec(deleteUserCardsQuery, wordCardID)
	if err != nil {
		r.logger.Warn("failed to delete user cards", zap.Error(err), zap.Int64("word_card_id", wordCardID))
		// Continue with deletion even if this fails
	} else {
		userCardsAffected, _ := userCardsResult.RowsAffected()
		r.logger.Info("deleted user cards",
			zap.Int64("word_card_id", wordCardID),
			zap.Int64("user_cards_deleted", userCardsAffected),
		)
	}

	// Step 2: Delete training_cards for this word_card_id
	deleteTrainingCardsQuery := `DELETE FROM training_cards WHERE word_card_id = ?`
	trainingCardsResult, err := r.db.Exec(deleteTrainingCardsQuery, wordCardID)
	if err != nil {
		r.logger.Warn("failed to delete training cards", zap.Error(err), zap.Int64("word_card_id", wordCardID))
		// Continue with deletion even if this fails
	} else {
		trainingCardsAffected, _ := trainingCardsResult.RowsAffected()
		r.logger.Info("deleted training cards",
			zap.Int64("word_card_id", wordCardID),
			zap.Int64("training_cards_deleted", trainingCardsAffected),
		)
	}

	// Step 3: Delete the word card itself
	query := `DELETE FROM word_cards WHERE id = ?`
	result, err := r.db.Exec(query, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to delete word card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("word card not found")
	}

	r.logger.Info("deleted word card and all related data",
		zap.Int64("word_card_id", wordCardID),
		zap.String("word", word),
	)

	return nil
}
