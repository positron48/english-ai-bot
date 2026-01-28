package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type mockTelegramClientNS struct {
	lastParams url.Values
}

func (c *mockTelegramClientNS) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	params, _ := url.ParseQuery(string(body))
	c.lastParams = params

	resp := `{"ok": true, "result": {"message_id": 1}}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(resp)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func newTestBotNS(client *mockTelegramClientNS) *tgbotapi.BotAPI {
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	return bot
}

func TestNotificationService_SendNotification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	nudgeRepo := repository.NewNudgeRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	_, err = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "apple",
		WordRU:     "яблоко",
		MeaningEN:  "apple",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard error: %v", err)
	}

	dueAt := time.Now().Add(-time.Hour)
	_, err = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             models.InitialEF,
		NextDueAt:      &dueAt,
	})
	if err != nil {
		t.Fatalf("CreateUserCard error: %v", err)
	}

	client := &mockTelegramClientNS{}
	bot := newTestBotNS(client)
	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	err = service.sendNotificationIfNeeded(user, time.Now())
	if err != nil {
		t.Fatalf("sendNotificationIfNeeded error: %v", err)
	}

	if got := client.lastParams.Get("text"); !strings.Contains(got, "К повторению") {
		t.Fatalf("expected notification message, got %q", got)
	}

	hasNudge, err := nudgeRepo.HasNudgeToday(user.ID, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("HasNudgeToday error: %v", err)
	}
	if !hasNudge {
		t.Fatalf("expected nudge to be recorded")
	}

	updated, err := userRepo.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID error: %v", err)
	}
	if updated == nil || !strings.Contains(updated.SettingsJSON, time.Now().Format("2006-01-02")) {
		t.Fatalf("expected last_notification_date in settings")
	}
}

func TestNotificationService_SkipsWhenDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	nudgeRepo := repository.NewNudgeRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(999)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	settings := models.UserSettings{NotificationFrequency: "never"}
	settingsJSON, _ := json.Marshal(settings)
	user.SettingsJSON = string(settingsJSON)

	client := &mockTelegramClientNS{}
	bot := newTestBotNS(client)
	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	err = service.sendNotificationIfNeeded(user, time.Now())
	if err != nil {
		t.Fatalf("sendNotificationIfNeeded error: %v", err)
	}
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Fatalf("expected no notification when disabled")
	}
}

func TestNotificationService_CheckAndSendNotifications(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	nudgeRepo := repository.NewNudgeRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(777)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}
	_, err = db.GetConnection().Exec(`UPDATE users SET timezone = ?, preferred_training_time = ? WHERE id = ?`, "UTC", "23:59", user.ID)
	if err != nil {
		t.Fatalf("update user error: %v", err)
	}

	client := &mockTelegramClientNS{}
	bot := newTestBotNS(client)
	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	service.checkAndSendNotifications()

	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Fatalf("expected no notification before preferred time")
	}
}
