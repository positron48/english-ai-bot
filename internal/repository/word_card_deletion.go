package repository

import (
	"context"
	"errors"
	"fmt"
)

var ErrWordCardIdentityMismatch = errors.New("word card identity mismatch")

// WordCardDeletion describes the rows observed before a deletion. Foreign keys
// also remove dependent review history and retain content reports with null links.
type WordCardDeletion struct {
	WordCardID    int64  `json:"word_card_id"`
	Word          string `json:"word"`
	CourseCode    string `json:"course_code"`
	TrainingCards int64  `json:"training_cards"`
	UserCards     int64  `json:"user_cards"`
	DryRun        bool   `json:"dry_run"`
	Deleted       bool   `json:"deleted"`
}

// DeleteWordCardMatching guards the identity in the final DELETE as well as the
// preview. One parent DELETE and its FK cascades are atomic, including on failure.
func (r *WordRepository) DeleteWordCardMatching(ctx context.Context, id int64, course, word string, dryRun bool) (*WordCardDeletion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin word deletion: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	result := &WordCardDeletion{WordCardID: id, DryRun: dryRun}
	err = tx.QueryRowContext(ctx, `SELECT word, COALESCE(course_code, '') FROM word_cards WHERE id = ?`, id).Scan(&result.Word, &result.CourseCode)
	if err != nil {
		return nil, fmt.Errorf("read word deletion target: %w", err)
	}
	if result.Word != word || result.CourseCode != course {
		return nil, ErrWordCardIdentityMismatch
	}
	err = tx.QueryRowContext(ctx, `SELECT
 (SELECT COUNT(*) FROM training_cards WHERE word_card_id = ?),
 (SELECT COUNT(*) FROM user_cards WHERE training_card_id IN (SELECT id FROM training_cards WHERE word_card_id = ?))`, id, id).Scan(&result.TrainingCards, &result.UserCards)
	if err != nil {
		return nil, fmt.Errorf("preview word deletion: %w", err)
	}
	if dryRun {
		return result, nil
	}
	deleted, err := tx.ExecContext(ctx, `DELETE FROM word_cards WHERE id = ? AND course_code = ? AND word = ?`, id, course, word)
	if err != nil {
		return nil, fmt.Errorf("delete matching word card: %w", err)
	}
	count, err := deleted.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("count deleted word cards: %w", err)
	}
	if count != 1 {
		return nil, ErrWordCardIdentityMismatch
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit word deletion: %w", err)
	}
	result.Deleted = true
	return result, nil
}
