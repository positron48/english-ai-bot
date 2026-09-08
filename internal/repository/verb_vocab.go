package repository

import (
	"fmt"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/verbtraining"
)

// LinkSpanishVerbForWord only examines the word just added to vocabulary.
func (r *VerbFormsRepository) LinkSpanishVerbForWord(wordCardID int64) (bool, error) {
	var lemma string
	if err := r.db.QueryRow(`SELECT LOWER(TRIM(word)) FROM word_cards WHERE id = ?`, wordCardID).Scan(&lemma); err != nil {
		return false, err
	}
	return r.LinkWordCardByLemma(wordCardID, lemma, "es", "auto_user_vocab")
}

// EnsureUserVerbCardsForWord preserves existing SRS state and inserts all missing
// cards in one statement, independently of the size of the user's vocabulary.
func (r *VerbFormsRepository) EnsureUserVerbCardsForWord(userID, wordCardID int64, scopes []string) error {
	scopes = verbtraining.ExpandScopes(scopes)
	if len(scopes) == 0 {
		return nil
	}
	args := []interface{}{userID, wordCardID, models.VerbCardTypeCloze}
	for _, scope := range scopes {
		args = append(args, strings.ToLower(strings.TrimSpace(scope)))
	}
	q := `INSERT INTO user_verb_cards (user_id, verb_training_card_id, state, ef)
 SELECT ?, c.id, 'new', 2.5 FROM verb_training_cards c
 JOIN verb_forms_dict d ON d.id = c.verb_form_dict_id
 WHERE c.word_card_id = ? AND c.card_type = ?
 AND ('es.' || d.tense || '.' || d.mood) IN (` + strings.TrimSuffix(strings.Repeat("?,", len(scopes)), ",") + `)` +
		verbTrainingEligibleByWordCardSQL("c") + verbTrainingPromptHasExampleTranslationSQL("c") + `
 ON CONFLICT(user_id, verb_training_card_id) DO NOTHING`
	if _, err := r.db.Exec(q, args...); err != nil {
		return fmt.Errorf("insert verb cards for vocabulary word: %w", err)
	}
	return nil
}
