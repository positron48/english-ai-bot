package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordAdminTestDB(t *testing.T) (*sql.DB, *WordRepository) {
	db := testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	wordRepo := NewWordRepository(db, logger)
	return db, wordRepo
}

func TestWordRepository_ListWordCardsAdmin(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("admin1", "definition 1")
	repo.SaveWordCard("admin2", "definition 2")

	// List word cards
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) < 2 {
		t.Errorf("Expected at least 2 cards, got %d", len(cards))
	}
}

func TestWordRepository_ListWordCardsAdminForCourse(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	if err := repo.SaveWordCard("course-en-word", "definition"); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWordCard("course-es-word", "definition"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE word_cards SET course_code = 'en_ru' WHERE word = 'course-en-word'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE word_cards SET course_code = 'es_ru' WHERE word = 'course-es-word'`); err != nil {
		t.Fatal(err)
	}

	cards, err := repo.ListWordCardsAdminForCourse("es_ru", nil, false, nil, "", "", 10, 0, "word", "asc")
	if err != nil {
		t.Fatalf("ListWordCardsAdminForCourse: %v", err)
	}
	if len(cards) != 1 || cards[0].Word != "course-es-word" {
		t.Fatalf("cards = %+v, want only Spanish course word", cards)
	}
	if len(cards[0].CourseCodes) != 1 || cards[0].CourseCodes[0] != "es_ru" {
		t.Fatalf("course_codes = %v, want [es_ru]", cards[0].CourseCodes)
	}
}

func TestWordRepository_ListWordCardsAdminForCourseCanonicalMappingOverridesLegacyTag(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	if err := repo.SaveWordCard("canonical-spanish-word", "definition"); err != nil {
		t.Fatal(err)
	}
	var wordCardID int64
	if err := db.QueryRow(`SELECT id FROM word_cards WHERE word = 'canonical-spanish-word'`).Scan(&wordCardID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE word_cards SET course_code = 'en_ru' WHERE id = ?`, wordCardID); err != nil {
		t.Fatal(err)
	}
	var spanishCourseID int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE code = 'es_ru'`).Scan(&spanishCourseID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO learning_items (course_id, item_type, source_kind, source_id, title, status)
		VALUES (?, 'word', 'word_card', ?, 'canonical-spanish-word', 'published')
	`, spanishCourseID, fmt.Sprintf("%d", wordCardID)); err != nil {
		t.Fatal(err)
	}

	englishCards, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "word", "asc")
	if err != nil {
		t.Fatal(err)
	}
	if len(englishCards) != 0 {
		t.Fatalf("English cards = %+v, want none", englishCards)
	}
	spanishCards, err := repo.ListWordCardsAdminForCourse("es_ru", nil, false, nil, "", "", 10, 0, "word", "asc")
	if err != nil {
		t.Fatal(err)
	}
	if len(spanishCards) != 1 || spanishCards[0].Word != "canonical-spanish-word" {
		t.Fatalf("Spanish cards = %+v, want canonical Spanish word", spanishCards)
	}
	if total, err := repo.CountWordCardsAdminForCourse("en_ru", nil, false, nil, "", ""); err != nil || total != 0 {
		t.Fatalf("English count = %d, err = %v", total, err)
	}
	if total, err := repo.CountWordCardsAdminForCourse("es_ru", nil, false, nil, "", ""); err != nil || total != 1 {
		t.Fatalf("Spanish count = %d, err = %v", total, err)
	}
}

func TestWordRepository_ListWordCardsAdmin_WithSearch(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("searchword", "definition")
	repo.SaveWordCard("otherword", "definition")

	// List word cards with search
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "search", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) == 0 {
		t.Error("Expected at least one card matching search")
	}
}

func TestWordRepository_ListWordCardsAdmin_SortByAndOrder(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)
	repo.SaveWordCard("zword", "def")
	repo.SaveWordCard("aword", "def")

	// sortBy=word, order=asc
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "word", "asc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(sortBy=word, asc) error = %v", err)
	}
	if len(cards) < 2 {
		t.Skip("need at least 2 cards to check order")
	}
	// First should be aword (lower)
	if cards[0].Word > cards[1].Word {
		t.Errorf("expected asc order by word, got %q then %q", cards[0].Word, cards[1].Word)
	}

	// sortBy=id
	_, err = repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "id", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(sortBy=id) error = %v", err)
	}

	// sortBy=pos and sortBy=has_cards (valid columns)
	_, err = repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "pos", "asc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(sortBy=pos) error = %v", err)
	}
	_, err = repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "has_cards", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(sortBy=has_cards) error = %v", err)
	}
}

func TestWordRepository_ListWordCardsAdmin_OnlyWithErrors(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)
	repo.SaveWordCard("errorword", "definition")
	card, err := repo.GetWordCard("errorword")
	if err != nil {
		t.Fatalf("GetWordCard: %v", err)
	}
	if err := repo.MarkWordCardProcessedError(card.ID, "test error"); err != nil {
		t.Fatalf("MarkWordCardProcessedError: %v", err)
	}
	cards, err := repo.ListWordCardsAdmin(nil, true, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(onlyWithErrors=true) error = %v", err)
	}
	var found bool
	for _, c := range cards {
		if c.Word == "errorword" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected errorword in list when onlyWithErrors=true")
	}
}

func TestWordRepository_ListWordCardsAdmin_HasAudioFilter(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	repo.SaveWordCard("withaudio", "def")
	repo.SaveWordCard("noaudio", "def")
	repo.SaveWordCard("pendingstale", "def")
	// Insert TTS row for one word so has_audio filter can match
	_, err := db.Exec(`INSERT INTO tts_generation_status (word, state, audio_rel_path) VALUES ('withaudio', 'ready', 'ab/cd/withaudio.mp3')
		ON CONFLICT (course_code, word) DO UPDATE SET state = 'ready', audio_rel_path = EXCLUDED.audio_rel_path`)
	if err != nil {
		t.Fatalf("insert tts_generation_status: %v", err)
	}
	_, err = db.Exec(`INSERT INTO tts_generation_status (word, state, audio_rel_path) VALUES ('pendingstale', 'pending', 'ab/cd/pendingstale.mp3')
		ON CONFLICT (course_code, word) DO UPDATE SET state = 'pending', audio_rel_path = EXCLUDED.audio_rel_path`)
	if err != nil {
		t.Fatalf("insert pending tts_generation_status: %v", err)
	}
	hasAudioTrue := true
	cardsWith, err := repo.ListWordCardsAdmin(nil, false, &hasAudioTrue, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(hasAudio=true) error = %v", err)
	}
	var foundWith, foundPendingInWith bool
	for _, c := range cardsWith {
		if c.Word == "withaudio" {
			foundWith = true
		}
		if c.Word == "pendingstale" {
			foundPendingInWith = true
		}
	}
	if !foundWith {
		t.Error("Expected withaudio in list when hasAudio=true")
	}
	if foundPendingInWith {
		t.Error("pendingstale must not appear when hasAudio=true because state is not ready")
	}
	hasAudioFalse := false
	cardsWithout, err := repo.ListWordCardsAdmin(nil, false, &hasAudioFalse, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(hasAudio=false) error = %v", err)
	}
	// Should include words without TTS or with empty audio
	if len(cardsWithout) == 0 {
		t.Error("Expected some cards when hasAudio=false")
	}
	var foundPendingInWithout bool
	for _, c := range cardsWithout {
		if c.Word == "pendingstale" {
			foundPendingInWithout = true
			if c.TTSAudioURL != nil {
				t.Error("pendingstale must not expose TTSAudioURL while state is pending")
			}
		}
	}
	if !foundPendingInWithout {
		t.Error("Expected pendingstale in list when hasAudio=false")
	}
}

func TestWordRepository_CountWordCardsAdmin(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("count1", "definition 1")
	repo.SaveWordCard("count2", "definition 2")
	repo.SaveWordCard("count3", "definition 3")

	// Count word cards
	count, err := repo.CountWordCardsAdmin(nil, false, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 cards, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithFilterUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, repo := setupWordAdminTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(100)

	// Create word cards first
	repo.SaveWordCard("userword1", "definition 1")
	repo.SaveWordCard("userword2", "definition 2")

	// Get word cards to get their IDs
	card1, err := repo.GetWordCard("userword1")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	card2, err := repo.GetWordCard("userword2")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Add request history for user with word_card_id
	word1 := "userword1"
	word2 := "userword2"
	repo.AddWordRequestHistoryWithCard(user.ID, "userword1", &card1.ID, &word1)
	repo.AddWordRequestHistoryWithCard(user.ID, "userword2", &card2.ID, &word2)

	// Count word cards for user
	userID := user.ID
	count, err := repo.CountWordCardsAdmin(&userID, false, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 cards for user, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithErrors(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word card with error
	repo.SaveWordCard("errorword", "definition")
	card, err := repo.GetWordCard("errorword")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Mark as processed with error
	err = repo.MarkWordCardProcessedError(card.ID, "test error")
	if err != nil {
		t.Fatalf("Failed to mark error: %v", err)
	}

	// Count word cards with errors
	count, err := repo.CountWordCardsAdmin(nil, true, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card with error, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithSearch(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("searchable", "definition")
	repo.SaveWordCard("other", "definition")

	// Count word cards with search
	count, err := repo.CountWordCardsAdmin(nil, false, nil, "search", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card matching search, got %d", count)
	}
}

func TestWordRepository_ListWordCardsAdmin_MissingTrainingPOS(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	logger, _ := zap.NewDevelopment()
	tcRepo := NewTrainingCardRepository(db, logger)

	// Word A: has a training card with pos=noun
	repo.SaveWordCard("wordwithnoun", "definition")
	cardA, _ := repo.GetWordCard("wordwithnoun")
	noun := "noun"
	_, err := tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: cardA.ID, WordEN: "wordwithnoun", SenseIndex: 0,
		WordRU: "слово", MeaningEN: "meaning", POS: &noun,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	// Word B: no training card with pos=noun (no cards at all)
	repo.SaveWordCard("wordwithoutnoun", "definition")

	// Filter: missing card for noun -> only words that have no training card with pos=noun
	list, err := repo.ListWordCardsAdmin(nil, false, nil, "", "noun", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	// Should contain only wordwithoutnoun (wordwithnoun has a noun card)
	var found bool
	for _, w := range list {
		if w.Word == "wordwithoutnoun" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find wordwithoutnoun in list when filtering by missing_training_pos=noun, got %d words", len(list))
	}
	for _, w := range list {
		if w.Word == "wordwithnoun" {
			t.Error("wordwithnoun has a noun card and must not appear when filtering by missing_training_pos=noun")
			break
		}
	}

	count, err := repo.CountWordCardsAdmin(nil, false, nil, "", "noun")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 word without noun card, got count %d", count)
	}
}

func TestWordRepository_GetWordCardRequestingUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, repo := setupWordAdminTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(100)
	u2, _ := userRepo.GetOrCreateUser(200)

	// Create a word card
	repo.SaveWordCard("requesting", "definition")

	// Get word card to get its ID
	card, err := repo.GetWordCard("requesting")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Add request history
	repo.AddWordRequestHistory(u1.ID, "requesting")
	repo.AddWordRequestHistory(u2.ID, "requesting")

	// Get requesting users
	users, err := repo.GetWordCardRequestingUsers(card.ID)
	if err != nil {
		t.Fatalf("GetWordCardRequestingUsers() error = %v", err)
	}
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}

func TestWordRepository_DeleteWordCard(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create a word card
	repo.SaveWordCard("deletecard", "definition")

	// Get word card to get its ID
	card, err := repo.GetWordCard("deletecard")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Delete word card
	err = repo.DeleteWordCard(card.ID)
	if err != nil {
		t.Fatalf("DeleteWordCard() error = %v", err)
	}

	// Verify deletion
	deleted, err := repo.GetWordCard("deletecard")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if deleted != nil {
		t.Error("GetWordCard() should return nil for deleted card")
	}
}

func TestWordRepository_DeleteWordCard_NotFound(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)
	err := repo.DeleteWordCard(99999)
	if err == nil {
		t.Fatal("DeleteWordCard(99999) expected error")
	}
	if err != nil && err.Error() != "word card not found" {
		t.Errorf("DeleteWordCard() error = %v, want 'word card not found'", err)
	}
}

func TestWordRepository_DeleteWordCard_WithTrainingAndUserCards(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	logger, _ := zap.NewDevelopment()
	userRepo := NewUserRepository(db, logger)
	tcRepo := NewTrainingCardRepository(db, logger)
	ucRepo := NewUserCardRepository(db, logger)

	repo.SaveWordCard("cascadeword", "definition")
	card, err := repo.GetWordCard("cascadeword")
	if err != nil {
		t.Fatalf("GetWordCard: %v", err)
	}
	user, _ := userRepo.GetOrCreateUser(200)
	tc := &models.TrainingCard{WordCardID: card.ID, WordEN: "cascadeword", SenseIndex: 0, WordRU: "каскад", MeaningEN: "cascade"}
	tcID, err := tcRepo.CreateTrainingCard(tc)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5, NextDueAt: &now}
	_, err = ucRepo.CreateUserCard(uc)
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}

	err = repo.DeleteWordCard(card.ID)
	if err != nil {
		t.Fatalf("DeleteWordCard() error = %v", err)
	}
	deleted, err := repo.GetWordCard("cascadeword")
	if err != nil {
		t.Fatalf("GetWordCard after delete: %v", err)
	}
	if deleted != nil {
		t.Error("GetWordCard() should return nil after DeleteWordCard")
	}
}
