package service

import (
	"encoding/json"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func newVerbTrainingServiceForTest(t *testing.T, enabled bool) (*VerbTrainingService, *repository.VerbFormsRepository, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	cfg := config.TrainingConfig{
		SpanishVerbFormsEnabled: enabled,
		VerbFormsMaxCards:       10,
		VerbFormsMaxNew:         5,
	}
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "es"}, cfg, zap.NewNop())

	var userID int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (88002) RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return svc, repo, userID
}

func TestVerbTrainingService_Enabled(t *testing.T) {
	repo := repository.NewVerbFormsRepository(nil, zap.NewNop())

	cases := []struct {
		name     string
		learning config.LearningConfig
		training config.TrainingConfig
		want     bool
	}{
		{"es enabled", config.LearningConfig{TargetLang: "es"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, true},
		{"es disabled flag", config.LearningConfig{TargetLang: "es"}, config.TrainingConfig{SpanishVerbFormsEnabled: false}, false},
		{"en target", config.LearningConfig{TargetLang: "en"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewVerbTrainingService(repo, tc.learning, tc.training, zap.NewNop())
			if got := svc.Enabled(); got != tc.want {
				t.Fatalf("Enabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVerbTrainingService_StartSession_disabled(t *testing.T) {
	svc, _, userID := newVerbTrainingServiceForTest(t, false)
	_, err := svc.StartSession(userID, []string{"es.presente.indicativo"})
	if err == nil || err.Error() != "verb forms training is disabled" {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestVerbTrainingService_StartSession_noCards(t *testing.T) {
	svc, _, userID := newVerbTrainingServiceForTest(t, true)
	_, err := svc.StartSession(userID, []string{"es.presente.indicativo"})
	if err == nil || err.Error() != "no cards available for training" {
		t.Fatalf("expected no cards error, got %v", err)
	}
}

func TestVerbTrainingService_StartSession_success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	cfg := config.TrainingConfig{
		SpanishVerbFormsEnabled: true,
		VerbFormsMaxCards:       5,
		VerbFormsMaxNew:         3,
	}
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "es"}, cfg, zap.NewNop())

	var userID int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (88003) RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var wordCardID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('hablar', 'говорить', 'verb') RETURNING id`).
		Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, 'hablar', 0, 'говорить', 'to speak', 'verb') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`,
		userID, trainingCardID); err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	lemmaID, err := repo.UpsertVerbLemma("hablar", "es", "test", "v1", "chk", `{"ru":{"gloss":"говорить"}}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	formID, err := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "hablo",
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm: %v", err)
	}
	_ = formID
	if err := repo.LinkWordCardToLemma(wordCardID, lemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}

	session, err := svc.StartSession(userID, []string{"es.presente.indicativo"})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if session == nil || session.SessionID <= 0 || len(session.Queue) == 0 {
		t.Fatalf("expected session with queue, got %+v", session)
	}
	for _, card := range session.Queue {
		var prompt map[string]interface{}
		if err := json.Unmarshal([]byte(card.PromptJSON), &prompt); err != nil {
			t.Fatalf("prompt json: %v", err)
		}
		if prompt["example_translation"] == nil || prompt["example_translation"] == "" {
			t.Fatalf("runtime card should include example_translation: %+v", prompt)
		}
	}
}

func TestVerbTrainingService_Grade_correctAndIncorrect(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "es"},
		config.TrainingConfig{SpanishVerbFormsEnabled: true}, zap.NewNop())

	var userID int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (88004) RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var wordCardID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('hablar', 'def') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	lemmaID, _ := repo.UpsertVerbLemma("hablar", "es", "t", "v", "c", `{}`)
	formID, _ := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "hablo",
	})
	_ = repo.LinkWordCardToLemma(wordCardID, lemmaID, 1.0, "test")
	vtcID, _ := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID: wordCardID, VerbFormDictID: formID, CardType: models.VerbCardTypeCloze,
		PromptJSON: `{"example_translation":"Я говорю."}`, AnswerJSON: `{"surface_form":"hablo"}`,
	})
	uvcID, err := repo.GetOrCreateUserVerbCard(userID, vtcID)
	if err != nil {
		t.Fatalf("GetOrCreateUserVerbCard: %v", err)
	}
	sessionID, err := repo.StartVerbSession(userID, 1, `{}`)
	if err != nil {
		t.Fatalf("StartVerbSession: %v", err)
	}

	if err := svc.Grade(userID, sessionID, uvcID, true); err != nil {
		t.Fatalf("Grade correct: %v", err)
	}
	srs, err := repo.GetVerbUserCardSRS(uvcID)
	if err != nil || srs == nil {
		t.Fatalf("GetVerbUserCardSRS: %+v err=%v", srs, err)
	}
	if srs.Reps != 1 || srs.State != "review" {
		t.Fatalf("after correct grade: reps=%d state=%q", srs.Reps, srs.State)
	}

	if err := svc.Grade(userID, sessionID, uvcID, false); err != nil {
		t.Fatalf("Grade incorrect: %v", err)
	}
	srs, _ = repo.GetVerbUserCardSRS(uvcID)
	if srs.Reps != 0 || srs.State != "learning" || srs.LapseCount != 1 {
		t.Fatalf("after incorrect grade: reps=%d state=%q lapses=%d", srs.Reps, srs.State, srs.LapseCount)
	}
}

func TestVerbTrainingService_Grade_notFound(t *testing.T) {
	svc, _, userID := newVerbTrainingServiceForTest(t, true)
	err := svc.Grade(userID, 1, 99999, true)
	if err == nil || err.Error() != "user verb card not found" {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestBuildRuntimeVerbTrainingCard_incompleteRow(t *testing.T) {
	_, err := buildRuntimeVerbTrainingCard(repository.LinkedVerbFormRow{Lemma: "hablar"}, `{}`)
	if err == nil {
		t.Fatal("expected error for incomplete row")
	}
}

func TestStableVerbTrainingSeed_deterministic(t *testing.T) {
	row := repository.LinkedVerbFormRow{
		Lemma: "hablar", Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", VerbFormDictID: 42,
	}
	a := stableVerbTrainingSeed(row)
	b := stableVerbTrainingSeed(row)
	if a != b {
		t.Fatalf("seed should be deterministic: %d vs %d", a, b)
	}
	row.VerbFormDictID = 43
	c := stableVerbTrainingSeed(row)
	if c == a {
		t.Fatal("different form id should change seed")
	}
}

func TestVerbTrainingService_Grade_disabled(t *testing.T) {
	svc, _, userID := newVerbTrainingServiceForTest(t, false)
	err := svc.Grade(userID, 1, 1, true)
	if err == nil || err.Error() != "verb forms training is disabled" {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestVerbTrainingService_StartSession_createsSessionRecord(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "es"},
		config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 5}, zap.NewNop())

	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (88005) RETURNING id`).Scan(&userID)
	var wordCardID int64
	_ = db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('comer', 'есть', 'verb') RETURNING id`).Scan(&wordCardID)
	var tcID int64
	_ = db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, 'comer', 0, 'есть', 'to eat', 'verb') RETURNING id`, wordCardID).Scan(&tcID)
	_, _ = db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`, userID, tcID)
	lemmaID, _ := repo.UpsertVerbLemma("comer", "es", "t", "v", "c", `{}`)
	formID, _ := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "como",
	})
	_ = formID
	_ = repo.LinkWordCardToLemma(wordCardID, lemmaID, 1.0, "test")

	before := time.Now()
	session, err := svc.StartSession(userID, []string{"es.presente.indicativo"})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if session.SessionID <= 0 {
		t.Fatal("expected positive session id")
	}
	var startedAt time.Time
	if err := db.QueryRow(`SELECT started_at FROM verb_training_sessions WHERE id=$1`, session.SessionID).Scan(&startedAt); err != nil {
		t.Fatalf("load session: %v", err)
	}
	if startedAt.Before(before.Add(-time.Second)) {
		t.Fatalf("session started_at looks wrong: %v", startedAt)
	}
}
