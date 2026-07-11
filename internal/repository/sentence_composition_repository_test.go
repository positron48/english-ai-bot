package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestSentenceComposition_SelectCandidateWords_OrdersByLeastUsed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewSentenceCompositionRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(7001)

	const course = "es_ru"
	// Seed three well-learned words (review state, mastering >= 70).
	ids := map[string]int64{}
	for _, w := range []struct{ word, ru string }{{"gato", "кот"}, {"perro", "собака"}, {"casa", "дом"}} {
		var wcID, tcID int64
		if err := db.QueryRow(`INSERT INTO word_cards (word, definition, course_code) VALUES (?, 'd', ?) RETURNING id`, w.word, course).Scan(&wcID); err != nil {
			t.Fatalf("insert word_card: %v", err)
		}
		if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code) VALUES (?, ?, 0, ?, 'm', ?) RETURNING id`, wcID, w.word, w.ru, course).Scan(&tcID); err != nil {
			t.Fatalf("insert training_card: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, course_code) VALUES (?, ?, 'ru_en', 'review', 2.5, ?)`, user.ID, tcID, course); err != nil {
			t.Fatalf("insert user_card: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, course_code) VALUES (?, ?, 80, ?)`, user.ID, wcID, course); err != nil {
			t.Fatalf("insert mastering: %v", err)
		}
		ids[w.word] = wcID
	}

	// "gato" has already participated twice, "perro" once, "casa" never -> expect casa, perro, gato.
	if _, err := db.Exec(`INSERT INTO sentence_word_usage (user_id, word_card_id, course_code, used_count, last_used_on) VALUES (?, ?, ?, 2, CAST(? AS date))`, user.ID, ids["gato"], course, "2026-06-01"); err != nil {
		t.Fatalf("seed usage gato: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sentence_word_usage (user_id, word_card_id, course_code, used_count, last_used_on) VALUES (?, ?, ?, 1, CAST(? AS date))`, user.ID, ids["perro"], course, "2026-06-02"); err != nil {
		t.Fatalf("seed usage perro: %v", err)
	}

	got, err := repo.SelectCandidateWords(user.ID, course, 70, 10)
	if err != nil {
		t.Fatalf("SelectCandidateWords: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(got))
	}
	if got[0].Lemma != "casa" || got[1].Lemma != "perro" || got[2].Lemma != "gato" {
		t.Fatalf("expected order [casa perro gato], got [%s %s %s]", got[0].Lemma, got[1].Lemma, got[2].Lemma)
	}
}

func TestSentenceComposition_SelectCandidateWords_IncludesExplicitlyKnownWithoutStudy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewSentenceCompositionRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(7006)
	const course = "es_ru"

	var knownWC int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, definition_ru, course_code) VALUES ('ventana', 'window', 'окно', ?) RETURNING id`, course).Scan(&knownWC); err != nil {
		t.Fatalf("insert known word: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code) VALUES (?, 'ventana', 0, 'окно', 'window', ?)`, knownWC, course); err != nil {
		t.Fatalf("insert known training card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status, course_code) VALUES (?, ?, 'known', ?)`, user.ID, knownWC, course); err != nil {
		t.Fatalf("mark known: %v", err)
	}

	var lowWC, lowTC int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, course_code) VALUES ('raro', 'rare', ?) RETURNING id`, course).Scan(&lowWC); err != nil {
		t.Fatalf("insert low word: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code) VALUES (?, 'raro', 0, 'редкий', 'rare', ?) RETURNING id`, lowWC, course).Scan(&lowTC); err != nil {
		t.Fatalf("insert low training card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, course_code) VALUES (?, ?, 'ru_en', 'review', 2.5, ?)`, user.ID, lowTC, course); err != nil {
		t.Fatalf("insert low user card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, course_code) VALUES (?, ?, 69, ?)`, user.ID, lowWC, course); err != nil {
		t.Fatalf("insert low mastery: %v", err)
	}

	got, err := repo.SelectCandidateWords(user.ID, course, 70, 10)
	if err != nil {
		t.Fatalf("SelectCandidateWords: %v", err)
	}
	if len(got) != 1 || got[0].WordCardID != knownWC || got[0].Lemma != "ventana" || got[0].Translation != "окно" {
		t.Fatalf("expected only explicitly known word, got %+v", got)
	}
}

func TestSentenceComposition_CreateAndRecordAttempt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewSentenceCompositionRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(7002)
	const course = "es_ru"

	var wcID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, course_code) VALUES ('gato', 'd', ?) RETURNING id`, course).Scan(&wcID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}

	set := &models.SentenceSet{UserID: user.ID, CourseCode: course, GenerationDate: "2026-06-29", Scopes: []string{"es.presente.indicativo"}}
	items := []models.SentenceItem{
		{Position: 0, PromptRU: "У меня есть кот", ReferenceES: "Tengo un gato", WordCardIDs: []int64{wcID}},
		{Position: 1, PromptRU: "Кот спит", ReferenceES: "El gato duerme", WordCardIDs: []int64{wcID}},
	}
	setID, err := repo.CreateSet(set, items, []int64{wcID})
	if err != nil {
		t.Fatalf("CreateSet: %v", err)
	}

	// Participation counter bumped.
	if n, _ := repo.ParticipationCount(user.ID, wcID, course); n != 1 {
		t.Fatalf("expected participation 1, got %d", n)
	}

	// Latest set must be 'ready' (never started) -> regeneration guard would block.
	latest, _ := repo.LatestSet(user.ID, course)
	if latest == nil || latest.Status != models.SentenceSetReady {
		t.Fatalf("expected ready set, got %+v", latest)
	}

	got, _ := repo.GetItems(setID)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}

	// Attempt first item: 0 errors -> star, set becomes 'started'.
	updated, err := repo.RecordAttempt(got[0].ID, "Tengo un gato", 0, models.SentenceOutcomeStar, `{"error_count":0}`)
	if err != nil {
		t.Fatalf("RecordAttempt 1: %v", err)
	}
	if updated.Status != models.SentenceSetStarted {
		t.Fatalf("expected started after first attempt, got %s", updated.Status)
	}
	if updated.StarCount != 1 {
		t.Fatalf("expected 1 star, got %d", updated.StarCount)
	}

	// Re-attempting the same item must be rejected (single shot).
	if _, err := repo.RecordAttempt(got[0].ID, "x", 0, models.SentenceOutcomeStar, `{}`); err == nil {
		t.Fatalf("expected error on re-attempt")
	}

	// Attempt second item with 2 errors -> failed, set completes.
	updated, err = repo.RecordAttempt(got[1].ID, "gato dormir", 2, models.SentenceOutcomeFailed, `{"error_count":2}`)
	if err != nil {
		t.Fatalf("RecordAttempt 2: %v", err)
	}
	if updated.Status != models.SentenceSetCompleted {
		t.Fatalf("expected completed, got %s", updated.Status)
	}
	if updated.FailedCount != 1 || updated.StarCount != 1 {
		t.Fatalf("expected stars=1 failed=1, got stars=%d failed=%d", updated.StarCount, updated.FailedCount)
	}
	if updated.CompletedAt == nil {
		t.Fatalf("expected completed_at set")
	}
}

func TestSentenceComposition_CreateSet_AllowsMultipleSetsPerDay(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewSentenceCompositionRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(7003)
	const course = "es_ru"

	var wcID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, course_code) VALUES ('casa', 'd', ?) RETURNING id`, course).Scan(&wcID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}

	set := &models.SentenceSet{UserID: user.ID, CourseCode: course, GenerationDate: "2026-07-01", Scopes: []string{"es.presente.indicativo"}}
	items := []models.SentenceItem{
		{Position: 0, PromptRU: "Это дом", ReferenceES: "Es una casa", WordCardIDs: []int64{wcID}},
	}
	firstID, err := repo.CreateSet(set, items, []int64{wcID})
	if err != nil {
		t.Fatalf("CreateSet first: %v", err)
	}
	secondID, err := repo.CreateSet(set, items, []int64{wcID})
	if err != nil {
		t.Fatalf("CreateSet second same day: %v", err)
	}
	if secondID == firstID {
		t.Fatalf("expected distinct set ids, got %d", secondID)
	}

	latest, err := repo.LatestSet(user.ID, course)
	if err != nil {
		t.Fatalf("LatestSet: %v", err)
	}
	if latest == nil || latest.ID != secondID {
		t.Fatalf("expected latest set %d, got %+v", secondID, latest)
	}
	if n, _ := repo.ParticipationCount(user.ID, wcID, course); n != 2 {
		t.Fatalf("expected participation 2, got %d", n)
	}
}

func TestOutcomeForErrorCount(t *testing.T) {
	cases := []struct {
		errors int
		want   string
	}{
		{0, models.SentenceOutcomeStar},
		{1, models.SentenceOutcomePassed},
		{2, models.SentenceOutcomeFailed},
		{5, models.SentenceOutcomeFailed},
	}
	for _, c := range cases {
		if got := models.OutcomeForErrorCount(c.errors); got != c.want {
			t.Errorf("OutcomeForErrorCount(%d) = %s, want %s", c.errors, got, c.want)
		}
	}
}
