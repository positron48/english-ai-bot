package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

type ContentReportRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewContentReportRepository(db *sql.DB, logger *zap.Logger) *ContentReportRepository {
	return &ContentReportRepository{db: db, logger: logger}
}

type CreateContentReportInput struct {
	UserID               int64
	SourceType           string
	Word                 string
	TranslationDirection string
	WordCardID           *int64
	TrainingCardID       *int64
	UserCardID           *int64
	WordCategory         string
	GrammarChapterID     string
	TheoryBlockID        string
	GrammarQuestionID    string
	Payload              map[string]interface{}
}

func (r *ContentReportRepository) Create(input CreateContentReportInput) (int64, error) {
	payloadJSON := ""
	if input.Payload != nil {
		raw, err := json.Marshal(input.Payload)
		if err != nil {
			return 0, fmt.Errorf("marshal report payload: %w", err)
		}
		payloadJSON = string(raw)
	}

	q := `INSERT INTO content_reports (
		user_id, source_type, status, word, translation_direction,
		word_card_id, training_card_id, user_card_id, word_category,
		grammar_chapter_id, theory_block_id, grammar_question_id, payload_json
	) VALUES (?, ?, 'active', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	id, err := database.InsertAndReturnID(r.db, q,
		input.UserID, input.SourceType, input.Word, input.TranslationDirection,
		input.WordCardID, input.TrainingCardID, input.UserCardID, input.WordCategory,
		input.GrammarChapterID, input.TheoryBlockID, input.GrammarQuestionID, payloadJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("create content report: %w", err)
	}
	return id, nil
}

func (r *ContentReportRepository) Resolve(reportID, resolvedByUserID int64) error {
	q := `UPDATE content_reports
	      SET status='resolved', resolved_at=CURRENT_TIMESTAMP, resolved_by_user_id=?, updated_at=CURRENT_TIMESTAMP
	      WHERE id=?`
	res, err := r.db.Exec(q, resolvedByUserID, reportID)
	if err != nil {
		return fmt.Errorf("resolve content report: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("resolve content report rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("content report not found")
	}
	return nil
}

func (r *ContentReportRepository) List(status string, limit int) ([]*models.ContentReport, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	base := `SELECT id, user_id, source_type, status, COALESCE(word,''), COALESCE(translation_direction,''),
	                word_card_id, training_card_id, user_card_id, COALESCE(word_category,''),
	                COALESCE(grammar_chapter_id,''), COALESCE(theory_block_id,''), COALESCE(grammar_question_id,''),
	                COALESCE(payload_json,''), resolved_at, resolved_by_user_id, created_at, updated_at
	         FROM content_reports`
	args := make([]interface{}, 0, 2)
	if status != "" {
		base += ` WHERE status = ?`
		args = append(args, status)
	}
	base += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.Query(base, args...)
	if err != nil {
		return nil, fmt.Errorf("list content reports: %w", err)
	}
	defer rows.Close()

	result := make([]*models.ContentReport, 0, limit)
	for rows.Next() {
		item, err := scanContentReport(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *ContentReportRepository) GetByID(reportID int64) (*models.ContentReport, error) {
	q := `SELECT id, user_id, source_type, status, COALESCE(word,''), COALESCE(translation_direction,''),
	             word_card_id, training_card_id, user_card_id, COALESCE(word_category,''),
	             COALESCE(grammar_chapter_id,''), COALESCE(theory_block_id,''), COALESCE(grammar_question_id,''),
	             COALESCE(payload_json,''), resolved_at, resolved_by_user_id, created_at, updated_at
	      FROM content_reports
	      WHERE id = ?`
	rows, err := r.db.Query(q, reportID)
	if err != nil {
		return nil, fmt.Errorf("get content report: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	return scanContentReport(rows)
}

func scanContentReport(rows *sql.Rows) (*models.ContentReport, error) {
	var (
		item         models.ContentReport
		wordCardID   sql.NullInt64
		trainingID   sql.NullInt64
		userCardID   sql.NullInt64
		resolvedAt   sql.NullTime
		resolvedByID sql.NullInt64
	)
	if err := rows.Scan(
		&item.ID, &item.UserID, &item.SourceType, &item.Status, &item.Word, &item.TranslationDirection,
		&wordCardID, &trainingID, &userCardID, &item.WordCategory,
		&item.GrammarChapterID, &item.TheoryBlockID, &item.GrammarQuestionID,
		&item.PayloadJSON, &resolvedAt, &resolvedByID, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan content report: %w", err)
	}
	if wordCardID.Valid {
		v := wordCardID.Int64
		item.WordCardID = &v
	}
	if trainingID.Valid {
		v := trainingID.Int64
		item.TrainingCardID = &v
	}
	if userCardID.Valid {
		v := userCardID.Int64
		item.UserCardID = &v
	}
	if resolvedAt.Valid {
		t := resolvedAt.Time
		item.ResolvedAt = &t
	}
	if resolvedByID.Valid {
		v := resolvedByID.Int64
		item.ResolvedByUserID = &v
	}
	return &item, nil
}

type WordTrainingContext struct {
	Word                 string
	TranslationDirection string
	WordCardID           *int64
	TrainingCardID       *int64
	UserCardID           *int64
	WordCategory         string
}

func (r *ContentReportRepository) GetWordTrainingContextByUserCard(userCardID int64) (*WordTrainingContext, error) {
	q := `SELECT COALESCE(tc.word_en, ''), COALESCE(uc.direction, ''), tc.word_card_id, tc.id, uc.id, COALESCE(tc.pos, '')
	      FROM user_cards uc
	      JOIN training_cards tc ON tc.id = uc.training_card_id
	      WHERE uc.id = ?`
	row := r.db.QueryRow(q, userCardID)
	var (
		ctx                    WordTrainingContext
		wordCardID, trainingID sql.NullInt64
		userCardIDNullable     sql.NullInt64
	)
	if err := row.Scan(&ctx.Word, &ctx.TranslationDirection, &wordCardID, &trainingID, &userCardIDNullable, &ctx.WordCategory); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get word training context: %w", err)
	}
	if wordCardID.Valid {
		v := wordCardID.Int64
		ctx.WordCardID = &v
	}
	if trainingID.Valid {
		v := trainingID.Int64
		ctx.TrainingCardID = &v
	}
	if userCardIDNullable.Valid {
		v := userCardIDNullable.Int64
		ctx.UserCardID = &v
	}
	return &ctx, nil
}

func (r *ContentReportRepository) ParsePayload(payloadJSON string) map[string]interface{} {
	if payloadJSON == "" {
		return map[string]interface{}{}
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		if r.logger != nil {
			r.logger.Warn("failed to parse content report payload", zap.Error(err))
		}
		return map[string]interface{}{}
	}
	return payload
}
