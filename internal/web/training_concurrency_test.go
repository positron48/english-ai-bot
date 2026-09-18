package web

import (
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"time"
)

func TestTextAnswerRetryDoesNotConsumeNextQuestion(t *testing.T) {
	state := &WebTrainingState{UserID: 42, SessionID: 7, Queue: []*models.TrainingQueueItem{
		{Type: "spell", Spell: &models.SpellChallenge{DisplayWord: "apple"}},
		{Type: "type", TypeChallenge: &models.TypeChallenge{DisplayWord: "banana"}},
	}}
	router := &Router{logger: zap.NewNop(), config: &config.Config{}, webTrainingHandler: &WebTrainingHandler{sessions: map[int64]*WebTrainingState{42: state}}}
	var first string
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader("answer_text=apple&session_id=7&card_index=1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.handleTrainingAnswer(w, setUserIDInContext(req, 42))
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		if i == 0 {
			first = w.Body.String()
		} else if first != w.Body.String() {
			t.Fatal("retry changed the answer")
		}
	}
	if state.CurrentIndex != 1 || state.CorrectCount != 1 {
		t.Fatalf("retry consumed next question: %+v", state)
	}
	for _, body := range []string{"answer_text=banana", "answer_text=banana&session_id=999&card_index=2"} {
		req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.handleTrainingAnswer(w, setUserIDInContext(req, 42))
		if w.Code != http.StatusConflict || state.CurrentIndex != 1 {
			t.Fatal("stale/missing identity accepted")
		}
	}
}

func TestSlowTrainingUserDoesNotBlockAnother(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	card := func(id int64) *models.TrainingQueueItem {
		return &models.TrainingQueueItem{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{ID: id}}}
	}
	router := &Router{logger: zap.NewNop(), config: &config.Config{}, webTrainingHandler: &WebTrainingHandler{sessions: map[int64]*WebTrainingState{
		1: {UserID: 1, Queue: []*models.TrainingQueueItem{card(1), card(2)}},
		2: {UserID: 2, Queue: []*models.TrainingQueueItem{card(3)}, Options: []string{"answer"}},
	}}, optionsService: inspectingTrainingOptions{inspect: func(*models.UserCardWithTraining) { close(entered); <-release }}}
	done := make(chan struct{})
	go func() {
		defer close(done)
		router.handleTrainingPrefetchNext(httptest.NewRecorder(), setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/prefetch-next", nil), 1))
	}()
	<-entered
	other := make(chan int, 1)
	go func() {
		w := httptest.NewRecorder()
		router.handleTrainingReveal(w, setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/reveal", nil), 2))
		other <- w.Code
	}()
	select {
	case code := <-other:
		if code != http.StatusOK {
			t.Errorf("status=%d", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("another user's training was blocked")
	}
	close(release)
	<-done
}
