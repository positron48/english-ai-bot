package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func insertEntreEntrarFixture(t *testing.T, db *sql.DB, courseCode string) (entreID, entrarID int64) {
	t.Helper()

	if err := db.QueryRow(
		`INSERT INTO word_cards (word, definition, course_code) VALUES (?, ?, ?) RETURNING id`,
		"entre", "preposition", courseCode,
	).Scan(&entreID); err != nil {
		t.Fatalf("insert entre card: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos) VALUES (?, ?, 0, ?, ?, ?)`,
		entreID, "entre", "между", "between", "preposition",
	); err != nil {
		t.Fatalf("insert entre training card: %v", err)
	}

	if err := db.QueryRow(
		`INSERT INTO word_cards (word, definition, course_code) VALUES (?, ?, ?) RETURNING id`,
		"entrar", "verb", courseCode,
	).Scan(&entrarID); err != nil {
		t.Fatalf("insert entrar card: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos) VALUES (?, ?, 0, ?, ?, ?)`,
		entrarID, "entrar", "входить", "to enter", "verb",
	); err != nil {
		t.Fatalf("insert entrar training card: %v", err)
	}

	var verbLemmaID int64
	err := db.QueryRow(`SELECT id FROM verb_lemmas WHERE lemma = ? AND language = 'es'`, "entrar").Scan(&verbLemmaID)
	if err == sql.ErrNoRows {
		if err := db.QueryRow(
			`INSERT INTO verb_lemmas (lemma, language) VALUES (?, 'es') RETURNING id`,
			"entrar",
		).Scan(&verbLemmaID); err != nil {
			t.Fatalf("insert verb lemma: %v", err)
		}
	} else if err != nil {
		t.Fatalf("lookup verb lemma: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO word_verb_lemmas (word_card_id, verb_lemma_id) VALUES (?, ?) ON CONFLICT (word_card_id) DO UPDATE SET verb_lemma_id = EXCLUDED.verb_lemma_id`,
		entrarID, verbLemmaID,
	); err != nil {
		t.Fatalf("link entrar verb lemma: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO verb_forms_dict (verb_lemma_id, mood, tense, person, number, surface_form) VALUES (?, 'subjunctive', 'present', '3', 'singular', 'entre')
ON CONFLICT (verb_lemma_id, mood, tense, person, number) DO UPDATE SET surface_form = EXCLUDED.surface_form`,
		verbLemmaID,
	); err != nil {
		t.Fatalf("insert verb surface entre: %v", err)
	}
	return entreID, entrarID
}

func TestFindReadingWordCardByInput_EntrePrepositionBeforeEntrarVerbForm(t *testing.T) {
	router, db, _ := setupReadingCoverageDB(t)
	entreID, entrarID := insertEntreEntrarFixture(t, db, "es_ru")

	lemma, id, found, err := router.findReadingWordCardByInput("entre", "es_ru")
	if err != nil || !found || id != entreID || lemma != "entre" {
		t.Fatalf("expected entre preposition card (id=%d), got lemma=%q id=%d found=%v err=%v (entrar=%d)", entreID, lemma, id, found, err, entrarID)
	}
}

func TestFindReadingWordCardByNativeTranslation_Mezhdu(t *testing.T) {
	router, db, _ := setupReadingCoverageDB(t)
	entreID, _ := insertEntreEntrarFixture(t, db, "es_ru")

	lemma, id, found, err := router.findReadingWordCardByNativeTranslation("между", "es_ru")
	if err != nil || !found || id != entreID || lemma != "entre" {
		t.Fatalf("expected entre via native translation, got lemma=%q id=%d found=%v err=%v", lemma, id, found, err)
	}
}

func TestHandleReadingWordLookup_CyrillicMezhduReturnsEntre(t *testing.T) {
	router, db, userID := setupReadingCoverageDB(t)
	entreID, _ := insertEntreEntrarFixture(t, db, "es_ru")

	req := httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=%D0%BC%D0%B5%D0%B6%D0%B4%D1%83&course_code=es_ru", nil)
	req = setUserIDInContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleReadingWordLookup(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("lookup status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["lemma"] != "entre" {
		t.Fatalf("expected lemma entre, got %v", payload["lemma"])
	}
	if int64(payload["word_card_id"].(float64)) != entreID {
		t.Fatalf("expected word_card_id %d, got %v", entreID, payload["word_card_id"])
	}
}
