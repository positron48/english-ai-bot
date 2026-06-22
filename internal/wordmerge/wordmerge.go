// Package wordmerge holds the logic for collapsing a duplicate surface-form word_card
// into its canonical lemma word_card. It deliberately depends only on the standard
// library so one-off maintenance binaries that import it stay small (the repository
// package transitively embeds large grammar assets via go:embed).
package wordmerge

import (
	"database/sql"
	"fmt"
	"strconv"
)

// MergeWordCardTx relinks every reference from a duplicate "form" word_card onto a
// canonical word_card (the lemma) inside tx, then deletes the form word_card row.
//
// Deleting the form cascades to its language-specific content and per-user progress
// (training_cards, user_cards, review_events, word_forms, verb training cards;
// content_reports.word_card_id is set NULL). Membership-style references that we want
// to preserve under the canonical card are relinked first:
//   - word_set_items                (set membership)
//   - learning_items                (unified-course items; source_id is TEXT, NOT a FK,
//     so a plain delete would orphan these rows)
//   - user_word_knowledge           ("known" status)
//   - user_word_mastering           (mastery score)
//
// Where the canonical card already has a conflicting row (same unique key), the form's
// row is dropped and the canonical row kept — i.e. canonical progress wins.
//
// tts_generation_status rows and audio files are NOT touched here: file removal isn't
// transactional, so the caller handles TTS cleanup around this call.
func MergeWordCardTx(tx *sql.Tx, formID, canonicalID int64) error {
	if formID == canonicalID {
		return fmt.Errorf("MergeWordCardTx: formID == canonicalID (%d)", formID)
	}

	// word_set_items: UNIQUE(word_set_id, word_card_id)
	if _, err := tx.Exec(`
		DELETE FROM word_set_items wsi
		WHERE wsi.word_card_id = ?
		  AND EXISTS (SELECT 1 FROM word_set_items o
		              WHERE o.word_set_id = wsi.word_set_id AND o.word_card_id = ?)`,
		formID, canonicalID); err != nil {
		return fmt.Errorf("dedup word_set_items: %w", err)
	}
	if _, err := tx.Exec(`UPDATE word_set_items SET word_card_id = ? WHERE word_card_id = ?`,
		canonicalID, formID); err != nil {
		return fmt.Errorf("relink word_set_items: %w", err)
	}

	// user_word_knowledge: UNIQUE(user_id, word_card_id)
	if _, err := tx.Exec(`
		DELETE FROM user_word_knowledge k
		WHERE k.word_card_id = ?
		  AND EXISTS (SELECT 1 FROM user_word_knowledge o
		              WHERE o.user_id = k.user_id AND o.word_card_id = ?)`,
		formID, canonicalID); err != nil {
		return fmt.Errorf("dedup user_word_knowledge: %w", err)
	}
	if _, err := tx.Exec(`UPDATE user_word_knowledge SET word_card_id = ? WHERE word_card_id = ?`,
		canonicalID, formID); err != nil {
		return fmt.Errorf("relink user_word_knowledge: %w", err)
	}

	// user_word_mastering: UNIQUE(user_id, word_card_id)
	if _, err := tx.Exec(`
		DELETE FROM user_word_mastering m
		WHERE m.word_card_id = ?
		  AND EXISTS (SELECT 1 FROM user_word_mastering o
		              WHERE o.user_id = m.user_id AND o.word_card_id = ?)`,
		formID, canonicalID); err != nil {
		return fmt.Errorf("dedup user_word_mastering: %w", err)
	}
	if _, err := tx.Exec(`UPDATE user_word_mastering SET word_card_id = ? WHERE word_card_id = ?`,
		canonicalID, formID); err != nil {
		return fmt.Errorf("relink user_word_mastering: %w", err)
	}

	// learning_items: source_id is TEXT, UNIQUE(course_id, source_kind, source_id).
	formStr := strconv.FormatInt(formID, 10)
	canonStr := strconv.FormatInt(canonicalID, 10)
	if _, err := tx.Exec(`
		DELETE FROM learning_items li
		WHERE li.source_kind = 'word_card' AND li.source_id = ?
		  AND EXISTS (SELECT 1 FROM learning_items o
		              WHERE o.course_id = li.course_id
		                AND o.source_kind = 'word_card' AND o.source_id = ?)`,
		formStr, canonStr); err != nil {
		return fmt.Errorf("dedup learning_items: %w", err)
	}
	if _, err := tx.Exec(`
		UPDATE learning_items SET source_id = ?
		WHERE source_kind = 'word_card' AND source_id = ?`,
		canonStr, formStr); err != nil {
		return fmt.Errorf("relink learning_items: %w", err)
	}

	// Delete the form word_card; cascade removes its training_cards, user_cards,
	// review_events, word_forms and verb training cards.
	if _, err := tx.Exec(`DELETE FROM word_cards WHERE id = ?`, formID); err != nil {
		return fmt.Errorf("delete form word_card: %w", err)
	}

	return nil
}
