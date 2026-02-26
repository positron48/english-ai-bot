package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func createWordSetStudyFixture(t *testing.T, router *Router) (userID int64, setID int64, wordCardID int64) {
	t.Helper()

	db := router.db
	userRepo := repository.NewUserRepository(db, router.logger)
	wordSetRepo := repository.NewWordSetRepository(db, router.logger)

	user, err := userRepo.GetOrCreateUser(777001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id",
		"studyword",
		"studyword",
	).Scan(&wordCardID); err != nil {
		t.Fatalf("create word card: %v", err)
	}

	setID, err = wordSetRepo.CreateWordSet(&models.WordSet{Title: "Study Set", IsPublished: true})
	if err != nil {
		t.Fatalf("create word set: %v", err)
	}

	if err := wordSetRepo.SetWordSetItems(setID, []int64{wordCardID}); err != nil {
		t.Fatalf("set word set items: %v", err)
	}

	return user.ID, setID, wordCardID
}

func TestHandleLearningWordsSetStudy_PreferredPOSSelection(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userID, setID, wordCardID := createWordSetStudyFixture(t, router)
	wordSetRepo := repository.NewWordSetRepository(router.db, router.logger)
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)

	preferredPOS := "verb"
	wordSet, err := wordSetRepo.GetWordSet(setID)
	if err != nil {
		t.Fatalf("get word set: %v", err)
	}
	wordSet.PreferredPOS = &preferredPOS
	if err := wordSetRepo.UpdateWordSet(wordSet); err != nil {
		t.Fatalf("update word set preferred_pos: %v", err)
	}

	noun := "noun"
	verb := "verb"
	if _, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "studyword (noun)",
		SenseIndex: 0,
		WordRU:     "слово",
		MeaningEN:  "noun meaning",
		POS:        &noun,
	}); err != nil {
		t.Fatalf("create noun training card: %v", err)
	}
	if _, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "studyword (verb)",
		SenseIndex: 1,
		WordRU:     "изучать",
		MeaningEN:  "verb meaning",
		POS:        &verb,
	}); err != nil {
		t.Fatalf("create verb training card: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/learning/words/sets/%d/study?word_card_id=%d", setID, wordCardID),
		nil,
	)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()

	router.handleLearningWordsSetStudy(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	card := payload["training_card"]
	if card["word_en"] != "studyword (verb)" {
		t.Fatalf("expected preferred POS card to be selected, got %v", card["word_en"])
	}
}

func TestHandleLearningWordsSetStudyLearnAndKnow_Flow(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userID, setID, wordCardID := createWordSetStudyFixture(t, router)
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)

	if _, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "studyword",
		SenseIndex: 0,
		WordRU:     "слово",
		MeaningEN:  "word",
	}); err != nil {
		t.Fatalf("create training card: %v", err)
	}

	if _, err := router.db.Exec(
		"INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known')",
		userID,
		wordCardID,
	); err != nil {
		t.Fatalf("seed known status: %v", err)
	}

	studyReq := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/learning/words/sets/%d/study?word_card_id=%d", setID, wordCardID),
		nil,
	)
	studyReq = setUserIDInContext(studyReq, userID)
	studyW := httptest.NewRecorder()
	router.handleLearningWordsSetDetailOrStudy(studyW, studyReq)
	if studyW.Code != http.StatusOK {
		t.Fatalf("study route expected 200, got %d: %s", studyW.Code, studyW.Body.String())
	}

	body := []byte(fmt.Sprintf(`{"word_card_id":%d}`, wordCardID))
	learnReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/learning/words/sets/%d/study/learn", setID),
		bytes.NewReader(body),
	)
	learnReq = setUserIDInContext(learnReq, userID)
	learnW := httptest.NewRecorder()
	router.handleLearningWordsSetDetailOrStudy(learnW, learnReq)

	if learnW.Code != http.StatusOK {
		t.Fatalf("learn route expected 200, got %d: %s", learnW.Code, learnW.Body.String())
	}

	var cardsAfterLearn int
	if err := router.db.QueryRow(
		`SELECT COUNT(*)
		 FROM user_cards uc
		 JOIN training_cards tc ON tc.id = uc.training_card_id
		 WHERE uc.user_id = $1 AND tc.word_card_id = $2`,
		userID,
		wordCardID,
	).Scan(&cardsAfterLearn); err != nil {
		t.Fatalf("count user cards after learn: %v", err)
	}
	if cardsAfterLearn == 0 {
		t.Fatal("expected user cards to be created after learn action")
	}

	var knownRowsAfterLearn int
	if err := router.db.QueryRow(
		"SELECT COUNT(*) FROM user_word_knowledge WHERE user_id = $1 AND word_card_id = $2 AND status = 'known'",
		userID,
		wordCardID,
	).Scan(&knownRowsAfterLearn); err != nil {
		t.Fatalf("count known rows after learn: %v", err)
	}
	if knownRowsAfterLearn != 0 {
		t.Fatalf("expected known status to be removed by learn action, got %d rows", knownRowsAfterLearn)
	}

	knowReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/learning/words/sets/%d/study/know", setID),
		bytes.NewReader(body),
	)
	knowReq = setUserIDInContext(knowReq, userID)
	knowW := httptest.NewRecorder()
	router.handleLearningWordsSetDetailOrStudy(knowW, knowReq)

	if knowW.Code != http.StatusOK {
		t.Fatalf("know route expected 200, got %d: %s", knowW.Code, knowW.Body.String())
	}

	var knownRowsAfterKnow int
	if err := router.db.QueryRow(
		"SELECT COUNT(*) FROM user_word_knowledge WHERE user_id = $1 AND word_card_id = $2 AND status = 'known'",
		userID,
		wordCardID,
	).Scan(&knownRowsAfterKnow); err != nil {
		t.Fatalf("count known rows after know: %v", err)
	}
	if knownRowsAfterKnow != 1 {
		t.Fatalf("expected known status to be set by know action, got %d rows", knownRowsAfterKnow)
	}

	var cardsAfterKnow int
	if err := router.db.QueryRow(
		`SELECT COUNT(*)
		 FROM user_cards uc
		 JOIN training_cards tc ON tc.id = uc.training_card_id
		 WHERE uc.user_id = $1 AND tc.word_card_id = $2`,
		userID,
		wordCardID,
	).Scan(&cardsAfterKnow); err != nil {
		t.Fatalf("count user cards after know: %v", err)
	}
	if cardsAfterKnow != 0 {
		t.Fatalf("expected user cards to be deleted by know action, got %d", cardsAfterKnow)
	}
}
