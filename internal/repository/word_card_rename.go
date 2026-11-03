package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrWordCardRenameConflict = errors.New("word rename conflicts with existing content")

type WordCardRename struct {
	WordCardID    int64  `json:"word_card_id"`
	CourseCode    string `json:"course_code"`
	OldWord       string `json:"old_word"`
	NewWord       string `json:"new_word"`
	TrainingCards int64  `json:"training_cards"`
	DryRun        bool   `json:"dry_run"`
	Renamed       bool   `json:"renamed"`
}

// RenameWordCardMatching preserves identities and review history. It never merges
// duplicate dictionary entries or rewrites historical report payloads.
func (r *WordRepository) RenameWordCardMatching(ctx context.Context, id int64, course, expected, replacement string, transcription *string, dryRun bool) (*WordCardRename, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck
	result := &WordCardRename{WordCardID: id, NewWord: replacement, DryRun: dryRun}
	var forms sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT word, COALESCE(course_code,''), verb_forms_json FROM word_cards WHERE id=? FOR UPDATE`, id).Scan(&result.OldWord, &result.CourseCode, &forms)
	if err != nil {
		return nil, err
	}
	if result.OldWord != expected || result.CourseCode != course {
		return nil, ErrWordCardIdentityMismatch
	}
	var conflict bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_cards WHERE LOWER(course_code)=LOWER(?) AND LOWER(word)=LOWER(?) AND id<>?)`, course, replacement, id).Scan(&conflict)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, ErrWordCardRenameConflict
	}
	// Lock children before inspecting their spelling; unexpected content is not overwritten.
	rows, err := tx.QueryContext(ctx, `SELECT word_en FROM training_cards WHERE word_card_id=? FOR UPDATE`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var word string
		if err = rows.Scan(&word); err != nil {
			rows.Close()
			return nil, err
		}
		if word != expected {
			rows.Close()
			return nil, ErrWordCardRenameConflict
		}
		result.TrainingCards++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if forms.Valid && forms.String != "" {
		var obj map[string]interface{}
		if err = json.Unmarshal([]byte(forms.String), &obj); err != nil {
			return nil, fmt.Errorf("invalid verb forms: %w", ErrWordCardRenameConflict)
		}
		if v, ok := obj["v1"]; ok && v == expected {
			obj["v1"] = replacement
			raw, e := json.Marshal(obj)
			if e != nil {
				return nil, e
			}
			forms.String = string(raw)
		}
	}
	if dryRun {
		return result, nil
	}
	_, err = tx.ExecContext(ctx, `UPDATE word_cards SET word=?, display_en=CASE WHEN display_en IS NULL OR display_en='' OR display_en=? THEN ? ELSE display_en END, transcription=COALESCE(?,transcription), verb_forms_json=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, replacement, expected, replacement, transcription, forms, id)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return nil, ErrWordCardRenameConflict
		}
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE training_cards SET word_en=?, display_word=CASE WHEN display_word IS NULL OR display_word='' OR display_word=? THEN ? ELSE display_word END, transcription=COALESCE(?,transcription) WHERE word_card_id=?`, replacement, expected, replacement, transcription, id)
	if err != nil {
		return nil, err
	}
	// Remove the typo's lookup alias; the canonical word lookup resolves the new name.
	_, err = tx.ExecContext(ctx, `DELETE FROM word_forms WHERE word_card_id=? AND LOWER(form)=LOWER(?)`, id, expected)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	result.Renamed = true
	return result, nil
}
