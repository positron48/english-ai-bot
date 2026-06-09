package repository

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestCourseRepository_TagLegacyWordTablesForLearning(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())

	var wordCardID int64
	if err := conn.QueryRow(`
		INSERT INTO word_cards (word, definition) VALUES ('hola', 'hi') RETURNING id
	`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO training_cards (word_card_id, word_en, word_ru, meaning_en, sense_index)
		VALUES (?, 'hola', 'привет', 'hi', 0)
	`, wordCardID); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}

	lc := config.LearningConfig{TargetLang: "es", NativeLang: "ru"}

	summary, err := repo.TagLegacyWordTablesForLearning(context.Background(), lc)
	if err != nil {
		t.Fatalf("TagLegacyWordTablesForLearning: %v", err)
	}
	if summary.CourseCode != "es_ru" {
		t.Fatalf("course code = %q want es_ru", summary.CourseCode)
	}
	if summary.Tagged["word_cards"] != 1 || summary.Tagged["training_cards"] != 1 {
		t.Fatalf("tagged = %+v want word_cards=1 training_cards=1", summary.Tagged)
	}

	var wcCourse, tcCourse string
	if err := conn.QueryRow(`SELECT course_code FROM word_cards WHERE id = ?`, wordCardID).Scan(&wcCourse); err != nil {
		t.Fatalf("read word_card course: %v", err)
	}
	if err := conn.QueryRow(`SELECT course_code FROM training_cards WHERE word_card_id = ?`, wordCardID).Scan(&tcCourse); err != nil {
		t.Fatalf("read training_card course: %v", err)
	}
	if wcCourse != "es_ru" || tcCourse != "es_ru" {
		t.Fatalf("course tags = word_card:%q training_card:%q want es_ru", wcCourse, tcCourse)
	}

	// Second run is a no-op (idempotent): nothing left with NULL course_code.
	again, err := repo.TagLegacyWordTablesForLearning(context.Background(), lc)
	if err != nil {
		t.Fatalf("second TagLegacyWordTablesForLearning: %v", err)
	}
	if again.Tagged["word_cards"] != 0 || again.Tagged["training_cards"] != 0 {
		t.Fatalf("second run tagged = %+v want all zero", again.Tagged)
	}
}
