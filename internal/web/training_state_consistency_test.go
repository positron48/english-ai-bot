package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

type inspectingTrainingOptions struct {
	inspect func(*models.UserCardWithTraining)
}

func (o inspectingTrainingOptions) GenerateOptions(card *models.UserCardWithTraining, _ int, _ []string, _, _ map[string]bool) ([]string, string, error) {
	o.inspect(card)
	// Stop before DB enrichment; the invariant must hold even during generation,
	// including when generation fails.
	return nil, "", errors.New("generation failed")
}

func TestTrainingPrefetchKeepsActiveStateDuringGeneration(t *testing.T) {
	shownAt := time.Now().Add(-time.Second)
	state := &WebTrainingState{
		UserID: 1, SessionID: 7, CourseCode: "es_ru", CurrentIndex: 0,
		ShownAt: shownAt, OptionsShownAt: &shownAt,
		Options: []string{"дом", "дерево"}, CorrectAnswer: "дом",
		Queue: []*models.TrainingQueueItem{
			{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{ID: 10, Direction: models.DirectionENtoRU}, TrainingCard: models.TrainingCard{WordEN: "casa", WordRU: "дом"}}},
			{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{ID: 20, Direction: models.DirectionRUtoEN}, TrainingCard: models.TrainingCard{WordEN: "árbol", WordRU: "дерево"}}},
		},
	}
	router := &Router{
		logger: zap.NewNop(), config: &config.Config{},
		webTrainingHandler: &WebTrainingHandler{sessions: map[int64]*WebTrainingState{1: state}},
		optionsService: inspectingTrainingOptions{inspect: func(card *models.UserCardWithTraining) {
			if card.UserCard.ID != 20 {
				t.Errorf("prefetch generated card %d, want 20", card.UserCard.ID)
			}
			if state.CurrentIndex != 0 || !state.ShownAt.Equal(shownAt) || state.OptionsShownAt != &shownAt ||
				state.CorrectAnswer != "дом" || !reflect.DeepEqual(state.Options, []string{"дом", "дерево"}) {
				t.Errorf("prefetch exposed the next card as active during generation: %+v", state)
			}
		}},
	}
	w := httptest.NewRecorder()
	router.handleTrainingPrefetchNext(w, setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/prefetch-next", nil), 1))
	if w.Code != http.StatusInternalServerError || len(state.PrefetchedCards) != 0 {
		t.Fatalf("failed prefetch should not be cached: status=%d cache=%v", w.Code, state.PrefetchedCards)
	}
}

type trainingResponseHook struct {
	*httptest.ResponseRecorder
	beforeWrite func()
}

func (w trainingResponseHook) WriteHeader(status int) {
	w.beforeWrite()
	w.ResponseRecorder.WriteHeader(status)
}

func TestTrainingRevealSnapshotsOptionsBeforeWritingResponse(t *testing.T) {
	state := &WebTrainingState{
		UserID: 1, SessionID: 7, Options: []string{"дом", "дерево"},
		Queue: []*models.TrainingQueueItem{{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{ID: 10}}}},
	}
	router := &Router{webTrainingHandler: &WebTrainingHandler{sessions: map[int64]*WebTrainingState{1: state}}}
	w := trainingResponseHook{httptest.NewRecorder(), func() {
		// Simulate another request changing state after the handler unlocks but
		// before JSON is written. The response must still describe card #1.
		router.webTrainingHandler.sessionsMutex.Lock()
		defer router.webTrainingHandler.sessionsMutex.Unlock()
		state.Options[0] = "árbol"
		state.Options = []string{"árbol", "flor"}
		state.CurrentIndex++
	}}
	router.handleTrainingReveal(w, setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/reveal", nil), 1))
	var response struct {
		Options    []string `json:"options"`
		UserCardID int64    `json:"user_card_id"`
		SessionID  int64    `json:"session_id"`
		CardIndex  int      `json:"card_index"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(response.Options, []string{"дом", "дерево"}) || response.UserCardID != 10 || response.SessionID != 7 || response.CardIndex != 1 {
		t.Fatalf("reveal mixed different cards: %+v", response)
	}
}

func TestTrainingRevealAfterLastAnswer(t *testing.T) {
	router := &Router{webTrainingHandler: &WebTrainingHandler{sessions: map[int64]*WebTrainingState{1: {UserID: 1}}}}
	w := httptest.NewRecorder()
	router.handleTrainingReveal(w, setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/reveal", nil), 1))
	if w.Code != http.StatusNotFound {
		t.Fatalf("late reveal status=%d, want 404", w.Code)
	}
}
