package wordmerge

import (
	"strconv"
	"testing"

	"tgbot-skeleton/internal/testutil"
)

func TestMergeWordCardTx_SameIDRejected(t *testing.T) {
	db := testutil.SetupTestDB(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := MergeWordCardTx(tx, 10, 10); err == nil {
		t.Fatal("expected error for same IDs")
	}
}

func TestMergeWordCardTx_RelinksAndDeletesForm(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var formID, canonID int64
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('running', 'to run', 'running', 'бег', 'en_ru') RETURNING id
	`).Scan(&formID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('run', 'to run', 'run', 'бегать', 'en_ru') RETURNING id
	`).Scan(&canonID); err != nil {
		t.Fatal(err)
	}
	var catID int64
	if err := db.QueryRow(`
		INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order, course_code)
		VALUES (NULL, 'Merge cat', 'test', 1, 1, 'en_ru') RETURNING id
	`).Scan(&catID); err != nil {
		t.Fatal(err)
	}
	var wordSetID int64
	if err := db.QueryRow(`
		INSERT INTO word_sets (category_id, title, description, is_published, sort_order, course_code)
		VALUES ($1, 'Merge set', 'test', 1, 1, 'en_ru') RETURNING id
	`, catID).Scan(&wordSetID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order) VALUES ($1, $2, 1)`, wordSetID, formID); err != nil {
		t.Fatal(err)
	}
	var userID int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (900501) RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO user_word_knowledge (user_id, word_card_id, status)
		VALUES ($1, $2, 'known')
	`, userID, formID); err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := MergeWordCardTx(tx, formID, canonID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("merge: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	var formCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM word_cards WHERE id = $1`, formID).Scan(&formCount); err != nil {
		t.Fatal(err)
	}
	if formCount != 0 {
		t.Fatal("form word_card should be deleted")
	}
	var itemCardID int64
	if err := db.QueryRow(`SELECT word_card_id FROM word_set_items WHERE word_set_id = $1`, wordSetID).Scan(&itemCardID); err != nil {
		t.Fatal(err)
	}
	if itemCardID != canonID {
		t.Fatalf("word_set_items card = %d, want %d", itemCardID, canonID)
	}
	var knowCardID int64
	if err := db.QueryRow(`SELECT word_card_id FROM user_word_knowledge WHERE user_id = $1`, userID).Scan(&knowCardID); err != nil {
		t.Fatal(err)
	}
	if knowCardID != canonID {
		t.Fatalf("user_word_knowledge card = %d, want %d", knowCardID, canonID)
	}
}

func TestMergeWordCardTx_DedupConflicts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var formID, canonID int64
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('forms', 'plural', 'forms', 'формы', 'en_ru') RETURNING id
	`).Scan(&formID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('form', 'shape', 'form', 'форма', 'en_ru') RETURNING id
	`).Scan(&canonID); err != nil {
		t.Fatal(err)
	}
	var userID int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (900502) RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	for _, cardID := range []int64{formID, canonID} {
		if _, err := db.Exec(`
			INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score)
			VALUES ($1, $2, 50)
		`, userID, cardID); err != nil {
			t.Fatal(err)
		}
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := MergeWordCardTx(tx, formID, canonID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("merge: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var masteringCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_word_mastering WHERE user_id = $1`, userID).Scan(&masteringCount); err != nil {
		t.Fatal(err)
	}
	if masteringCount != 1 {
		t.Fatalf("expected 1 mastering row, got %d", masteringCount)
	}
}

func TestMergeWordCardTx_RelinksLearningItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var formID, canonID int64
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('dup-a', 'a', 'dup-a', 'a', 'en_ru') RETURNING id
	`).Scan(&formID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru, course_code)
		VALUES ('dup-b', 'b', 'dup-b', 'b', 'en_ru') RETURNING id
	`).Scan(&canonID); err != nil {
		t.Fatal(err)
	}
	var courseID int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE code = 'en_ru'`).Scan(&courseID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO learning_items (course_id, item_type, source_kind, source_id, title, status)
		VALUES ($1, 'word', 'word_card', $2, 'dup item', 'published')
	`, courseID, strconv.FormatInt(formID, 10)); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := MergeWordCardTx(tx, formID, canonID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("merge: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var sourceID string
	if err := db.QueryRow(`
		SELECT source_id FROM learning_items WHERE course_id = $1 AND source_kind = 'word_card'
	`, courseID).Scan(&sourceID); err != nil {
		t.Fatal(err)
	}
	if sourceID != strconv.FormatInt(canonID, 10) {
		t.Fatalf("learning_items source_id = %q, want %q", sourceID, strconv.FormatInt(canonID, 10))
	}
}
