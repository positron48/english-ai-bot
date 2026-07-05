package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// TrainingCardRepository handles database operations for training cards
type TrainingCardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTrainingCardRepository creates a new training card repository
func NewTrainingCardRepository(db *sql.DB, logger *zap.Logger) *TrainingCardRepository {
	return &TrainingCardRepository{
		db:     db,
		logger: logger,
	}
}

// GetTrainingCardByWordCardIDAndSenseIndex gets a training card by word_card_id and sense_index
func (r *TrainingCardRepository) GetTrainingCardByWordCardIDAndSenseIndex(wordCardID int64, senseIndex int) (*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  word_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  pos, display_word, created_at
			  FROM training_cards WHERE word_card_id = ? AND sense_index = ?`

	var card models.TrainingCard
	var createdAt string
	var pos, displayWord sql.NullString

	err := r.db.QueryRow(query, wordCardID, senseIndex).Scan(
		&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
		&card.WordRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
		&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
		&pos, &displayWord, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get training card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	if pos.Valid {
		card.POS = &pos.String
	}
	if displayWord.Valid {
		card.DisplayWord = &displayWord.String
	}

	models.SyncTrainingCardNeutralAliases(&card)

	return &card, nil
}

// CreateTrainingCard creates a new training card
// If a card with the same word_card_id and sense_index already exists, returns its ID
func (r *TrainingCardRepository) CreateTrainingCard(card *models.TrainingCard) (int64, error) {
	models.NormalizeTrainingCardLegacyBeforeWrite(card)

	// Check if training card already exists
	existing, err := r.GetTrainingCardByWordCardIDAndSenseIndex(card.WordCardID, card.SenseIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing training card: %w", err)
	}
	if existing != nil {
		r.logger.Debug("training card already exists, returning existing ID",
			zap.Int64("id", existing.ID),
			zap.String("word", card.WordEN),
			zap.Int("sense_index", card.SenseIndex),
		)
		return existing.ID, nil
	}

	query := `INSERT INTO training_cards (
		word_card_id, word_en, transcription, sense_index,
		word_ru, meaning_en, example_en, example_ru,
		distractors_ru, distractors_en, hint, pos, display_word, course_code
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		(SELECT NULLIF(course_code, '') FROM word_cards WHERE id = ?))`

	// Use display_word if provided, otherwise fall back to word_en
	displayWord := card.WordEN
	if card.DisplayWord != nil && *card.DisplayWord != "" {
		displayWord = *card.DisplayWord
	}

	id, err := database.InsertAndReturnID(r.db, query,
		card.WordCardID, card.WordEN, card.Transcription, card.SenseIndex,
		card.WordRU, card.MeaningEN, card.ExampleEN, card.ExampleRU,
		card.DistractorsRU, card.DistractorsEN, card.Hint,
		card.POS, displayWord, card.WordCardID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create training card: %w", err)
	}

	r.logger.Debug("created training card",
		zap.Int64("id", id),
		zap.String("word", card.WordEN),
		zap.Int("sense_index", card.SenseIndex),
	)

	return id, nil
}

// GetTrainingCard gets a training card by ID
func (r *TrainingCardRepository) GetTrainingCard(id int64) (*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  word_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  pos, display_word, created_at
			  FROM training_cards WHERE id = ?`

	var card models.TrainingCard
	var createdAt string
	var pos, displayWord sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
		&card.WordRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
		&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
		&pos, &displayWord, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get training card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	if pos.Valid {
		card.POS = &pos.String
	}
	if displayWord.Valid {
		card.DisplayWord = &displayWord.String
	}

	models.SyncTrainingCardNeutralAliases(&card)

	return &card, nil
}

// GetTrainingCardsByWordCardID gets all training cards for a word card
func (r *TrainingCardRepository) GetTrainingCardsByWordCardID(wordCardID int64) ([]*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  word_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  pos, display_word, created_at
			  FROM training_cards WHERE word_card_id = ? ORDER BY sense_index`

	rows, err := r.db.Query(query, wordCardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get training cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.TrainingCard
	for rows.Next() {
		var card models.TrainingCard
		var createdAt string
		var pos, displayWord sql.NullString

		err := rows.Scan(
			&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
			&card.WordRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
			&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
			&pos, &displayWord, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan training card: %w", err)
		}

		card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

		if pos.Valid {
			card.POS = &pos.String
		}
		if displayWord.Valid {
			card.DisplayWord = &displayWord.String
		}

		models.SyncTrainingCardNeutralAliases(&card)

		cards = append(cards, &card)
	}

	return cards, nil
}

// GetWordCardsWithoutTrainingCards gets word cards that don't have training cards yet and haven't been processed with an error
func (r *TrainingCardRepository) GetWordCardsWithoutTrainingCards(limit int) ([]*models.WordCard, error) {
	return r.getWordCardsWithoutTrainingCards("", limit)
}

// GetWordCardsWithoutTrainingCardsForCourse is like GetWordCardsWithoutTrainingCards but only
// returns cards belonging to the given course_code. Used by per-course training workers on the
// unified DB so each worker processes (and validates) only its own course's words.
func (r *TrainingCardRepository) GetWordCardsWithoutTrainingCardsForCourse(courseCode string, limit int) ([]*models.WordCard, error) {
	return r.getWordCardsWithoutTrainingCards(courseCode, limit)
}

func (r *TrainingCardRepository) getWordCardsWithoutTrainingCards(courseCode string, limit int) ([]*models.WordCard, error) {
	query := `SELECT wc.id, wc.word, wc.definition,
			  wc.pos, wc.transcription, wc.definition_ru,
			  wc.examples_json, wc.verb_forms_json, wc.display_en,
			  COALESCE(wc.course_code, '') as course_code,
			  COALESCE(CAST(wc.processed_at AS TEXT), '') as processed_at,
			  COALESCE(wc.processing_error, '') as processing_error,
			  CAST(wc.created_at AS TEXT) as created_at,
			  CAST(wc.updated_at AS TEXT) as updated_at
			  FROM word_cards wc
			  LEFT JOIN training_cards tc ON wc.id = tc.word_card_id
			  WHERE tc.id IS NULL AND wc.processed_at IS NULL`

	args := []interface{}{}
	if courseCode != "" {
		query += ` AND wc.course_code = ?`
		args = append(args, courseCode)
	}
	query += ` ORDER BY wc.created_at LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get word cards without training cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.WordCard
	for rows.Next() {
		var card models.WordCard
		var createdAt, updatedAt, processedAtStr, processingErrorStr, courseCode string
		var pos, transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN sql.NullString

		err := rows.Scan(&card.ID, &card.Word, &card.Definition,
			&pos, &transcription, &definitionRU,
			&examplesJSON, &verbFormsJSON, &displayEN,
			&courseCode,
			&processedAtStr, &processingErrorStr, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word card: %w", err)
		}

		card.CourseCode = courseCode

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

		models.SyncWordCardNeutralAliases(&card)

		cards = append(cards, &card)
	}

	return cards, nil
}

// HasTrainingCards checks if a word card has training cards
func (r *TrainingCardRepository) HasTrainingCards(wordCardID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM training_cards WHERE word_card_id = ?`
	var count int
	err := r.db.QueryRow(query, wordCardID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check training cards: %w", err)
	}
	return count > 0, nil
}

// DeleteTrainingCardsByWordEN deletes all training cards for a word by word_en
func (r *TrainingCardRepository) DeleteTrainingCardsByWordEN(wordEN string) (int64, error) {
	query := `DELETE FROM training_cards WHERE word_en = ?`
	result, err := r.db.Exec(query, wordEN)
	if err != nil {
		return 0, fmt.Errorf("failed to delete training cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deleted training cards",
		zap.String("word_en", wordEN),
		zap.Int64("rows_affected", rowsAffected),
	)

	return rowsAffected, nil
}

// DeleteAllTrainingCards deletes all training cards (cascades to user_cards and review_events)
func (r *TrainingCardRepository) DeleteAllTrainingCards() (int64, error) {
	query := `DELETE FROM training_cards`
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all training cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deleted all training cards",
		zap.Int64("rows_affected", rowsAffected),
	)

	return rowsAffected, nil
}

// GetTrainingCardsByWordEN gets all training cards for a word by word_en (or display_word)
func (r *TrainingCardRepository) GetTrainingCardsByWordEN(wordEN string) ([]*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  word_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  pos, display_word, created_at
			  FROM training_cards WHERE word_en = ? OR display_word = ? ORDER BY sense_index`

	rows, err := r.db.Query(query, wordEN, wordEN)
	if err != nil {
		return nil, fmt.Errorf("failed to get training cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.TrainingCard
	for rows.Next() {
		var card models.TrainingCard
		var createdAt string
		var pos, displayWord sql.NullString

		err := rows.Scan(
			&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
			&card.WordRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
			&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
			&pos, &displayWord, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan training card: %w", err)
		}

		card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

		if pos.Valid {
			card.POS = &pos.String
		}
		if displayWord.Valid {
			card.DisplayWord = &displayWord.String
		}

		models.SyncTrainingCardNeutralAliases(&card)

		cards = append(cards, &card)
	}

	return cards, nil
}

// UpdateTrainingCard updates a training card
func (r *TrainingCardRepository) UpdateTrainingCard(card *models.TrainingCard) error {
	models.NormalizeTrainingCardLegacyBeforeWrite(card)

	displayWord := card.WordEN
	if card.DisplayWord != nil && *card.DisplayWord != "" {
		displayWord = *card.DisplayWord
	}

	query := `UPDATE training_cards SET
			  word_ru = ?, meaning_en = ?, example_en = ?, example_ru = ?,
			  transcription = ?, distractors_ru = ?, distractors_en = ?, hint = ?,
			  pos = ?, display_word = ?
			  WHERE id = ?`

	_, err := r.db.Exec(query,
		card.WordRU, card.MeaningEN, card.ExampleEN, card.ExampleRU,
		card.Transcription, card.DistractorsRU, card.DistractorsEN, card.Hint,
		card.POS, displayWord,
		card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update training card: %w", err)
	}

	r.logger.Debug("updated training card",
		zap.Int64("id", card.ID),
		zap.String("word", card.WordEN),
		zap.Int("sense_index", card.SenseIndex),
	)

	return nil
}

// DeleteTrainingCard deletes a training card by ID
func (r *TrainingCardRepository) DeleteTrainingCard(id int64) error {
	query := `DELETE FROM training_cards WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete training card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("training card not found")
	}

	r.logger.Info("deleted training card",
		zap.Int64("id", id),
	)

	return nil
}

// OrphanedTrainingCardInfo represents an orphaned training card (without word_card)
type OrphanedTrainingCardInfo struct {
	TrainingCardID int64
	WordCardID     int64
	WordEN         string
	WordTarget     string
	Transcription  string
	SenseIndex     int
	WordRU         string
	WordNative     string
	MeaningEN      string
	MeaningTarget  string
	ExampleEN      string
	ExampleTarget  string
	ExampleRU      string
	ExampleNative  string
	POS            *string
	DisplayWord    *string
	CreatedAt      time.Time
	UserCardsCount int64
}

// ListOrphanedTrainingCards lists training_cards that reference non-existent word_cards
func (r *TrainingCardRepository) ListOrphanedTrainingCards(limit, offset int) ([]*OrphanedTrainingCardInfo, error) {
	query := `SELECT 
		tc.id as training_card_id,
		tc.word_card_id,
		tc.word_en,
		COALESCE(tc.transcription, '') as transcription,
		tc.sense_index,
		tc.word_ru,
		tc.meaning_en,
		COALESCE(tc.example_en, '') as example_en,
		COALESCE(tc.example_ru, '') as example_ru,
		tc.pos,
		tc.display_word,
		tc.created_at,
		COALESCE(COUNT(uc.id), 0) as user_cards_count
	FROM training_cards tc
	LEFT JOIN word_cards wc ON tc.word_card_id = wc.id
	LEFT JOIN user_cards uc ON uc.training_card_id = tc.id
	WHERE wc.id IS NULL
	GROUP BY tc.id, tc.word_card_id, tc.word_en, tc.transcription, tc.sense_index, 
	         tc.word_ru, tc.meaning_en, tc.example_en, tc.example_ru, tc.pos, tc.display_word, tc.created_at
	ORDER BY tc.created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orphaned training cards: %w", err)
	}
	defer rows.Close()

	var items []*OrphanedTrainingCardInfo
	for rows.Next() {
		var item OrphanedTrainingCardInfo
		var createdAt string
		var pos, displayWord sql.NullString

		err := rows.Scan(
			&item.TrainingCardID,
			&item.WordCardID,
			&item.WordEN,
			&item.Transcription,
			&item.SenseIndex,
			&item.WordRU,
			&item.MeaningEN,
			&item.ExampleEN,
			&item.ExampleRU,
			&pos,
			&displayWord,
			&createdAt,
			&item.UserCardsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan orphaned training card: %w", err)
		}

		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

		if pos.Valid {
			item.POS = &pos.String
		}
		if displayWord.Valid {
			item.DisplayWord = &displayWord.String
		}

		item.WordTarget = item.WordEN
		item.WordNative = item.WordRU
		item.MeaningTarget = item.MeaningEN
		item.ExampleTarget = item.ExampleEN
		item.ExampleNative = item.ExampleRU

		items = append(items, &item)
	}

	return items, nil
}

// CountOrphanedTrainingCards counts total orphaned training cards
func (r *TrainingCardRepository) CountOrphanedTrainingCards() (int, error) {
	query := `SELECT COUNT(*) 
	FROM training_cards tc
	LEFT JOIN word_cards wc ON tc.word_card_id = wc.id
	WHERE wc.id IS NULL`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orphaned training cards: %w", err)
	}

	return count, nil
}
