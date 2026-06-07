package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/config"
)

// LinglowSRSMirrorRepository keeps canonical srs_items aligned with legacy writes
// while LINGLOW_SRS_READ_ENABLED routes city/review reads to srs_items.
type LinglowSRSMirrorRepository struct {
	word    *LinglowWordSRSBackfillRepository
	grammar *LinglowGrammarSRSBackfillRepository
	db      *sql.DB
}

func NewLinglowSRSMirrorRepository(db *sql.DB) *LinglowSRSMirrorRepository {
	return &LinglowSRSMirrorRepository{
		word:    NewLinglowWordSRSBackfillRepository(db),
		grammar: NewLinglowGrammarSRSBackfillRepository(db),
		db:      db,
	}
}

func (r *LinglowSRSMirrorRepository) MirrorWordReview(ctx context.Context, lc config.LearningConfig, userID, userCardID int64) error {
	if userID == 0 || userCardID == 0 {
		return nil
	}
	var wordCardID int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT tc.word_card_id
		FROM user_cards uc
		JOIN training_cards tc ON tc.id = uc.training_card_id
		WHERE uc.id = ? AND uc.user_id = ?
	`, userCardID, userID).Scan(&wordCardID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("lookup word card for srs mirror: %w", err)
	}
	return r.word.SyncWordLearningItemForUser(ctx, lc, userID, wordCardID)
}

func (r *LinglowSRSMirrorRepository) MirrorGrammarTraining(ctx context.Context, lc config.LearningConfig, userID int64, chapterID, theoryBlockID string) error {
	chapterID = strings.TrimSpace(chapterID)
	theoryBlockID = strings.TrimSpace(theoryBlockID)
	if userID == 0 || chapterID == "" || theoryBlockID == "" {
		return nil
	}
	return r.grammar.SyncTheoryBlockForUser(ctx, lc, userID, chapterID, theoryBlockID)
}
