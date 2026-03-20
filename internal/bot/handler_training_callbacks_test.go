package bot

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func seedDueCardForTelegramUser(t *testing.T, h *Handler, telegramID int64) int64 {
	t.Helper()

	user, err := h.userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	spell := false
	typing := false
	settings := models.UserSettings{SpellModeEnabled: &spell, TypeModeEnabled: &typing}
	settingsJSON, _ := json.Marshal(settings)
	if err := h.userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("UpdateUserSettings failed: %v", err)
	}

	var wordCardID int64
	if err := h.db.QueryRow(
		"INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id",
		"train",
		"train",
	).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card failed: %v", err)
	}

	trainingCardID, err := h.trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "train",
		SenseIndex:    0,
		WordRU:        "тренировать",
		MeaningEN:     "train",
		DistractorsRU: `["проверять","делать","писать"]`,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard failed: %v", err)
	}

	dueAt := time.Now().Add(-time.Hour)
	if _, err := h.userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &dueAt,
	}); err != nil {
		t.Fatalf("CreateUserCard failed: %v", err)
	}

	return user.ID
}

func TestHandleTrainCommand_NoCards(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	h.handleTrainCommand(context.Background(), 10, 42)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "нет карточек") {
		t.Fatalf("expected no-cards message, got %q", got)
	}
}

func TestHandleTrainCommand_WithCards(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)
	seedDueCardForTelegramUser(t, h, 42)

	h.handleTrainCommand(context.Background(), 10, 42)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "Карточка") {
		t.Fatalf("expected training question message, got %q", got)
	}
}

func TestHandleCallbackQuery_TrainStart(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	q := &tgbotapi.CallbackQuery{
		ID:   "cb-train-start",
		Data: "train_start",
		From: &tgbotapi.User{ID: 42, UserName: "tester"},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 10},
		},
	}

	h.handleCallbackQuery(context.Background(), q)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "нет карточек") {
		t.Fatalf("expected no-cards train_start message, got %q", got)
	}
}

func TestHandleCallbackQuery_AnswerNoActiveSession(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	q := &tgbotapi.CallbackQuery{
		ID:   "cb-answer",
		Data: "answer_0",
		From: &tgbotapi.User{ID: 42, UserName: "tester"},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 10},
		},
	}

	h.handleCallbackQuery(context.Background(), q)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "Сессия не найдена") {
		t.Fatalf("expected no-session message, got %q", got)
	}
}

func TestHandleCallbackQuery_InvalidOptionIndex(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	q := &tgbotapi.CallbackQuery{
		ID:   "cb-invalid",
		Data: "answer_bad",
		From: &tgbotapi.User{ID: 42, UserName: "tester"},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 10},
		},
	}

	h.handleCallbackQuery(context.Background(), q)

	if got := client.lastParams.Get("callback_query_id"); got != "cb-invalid" {
		t.Fatalf("expected callback ack to be sent, got callback_query_id=%q", got)
	}
}
