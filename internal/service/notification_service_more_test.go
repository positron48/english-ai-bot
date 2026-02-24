package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"
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
	db := testutil.SetupTestDatabase(t)

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
	for index := 0; index < 10; index++ {
		cardID, createErr := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wordID,
			WordEN:     "apple",
			WordRU:     "яблоко",
			MeaningEN:  "apple",
			SenseIndex: index,
		})
		if createErr != nil {
			t.Fatalf("CreateTrainingCard error: %v", createErr)
		}

		dueAt := time.Now().Add(-time.Hour)
		_, createErr = userCardRepo.CreateUserCard(&models.UserCard{
			UserID:         user.ID,
			TrainingCardID: cardID,
			Direction:      models.DirectionRUtoEN,
			State:          models.StateReview,
			EF:             models.InitialEF,
			NextDueAt:      &dueAt,
		})
		if createErr != nil {
			t.Fatalf("CreateUserCard error: %v", createErr)
		}
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

	// Message should still contain core CTA
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

func TestNotificationService_SkipsWhenDueCardsLessThanTen(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	nudgeRepo := repository.NewNudgeRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(45678)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "banana"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	for index := 0; index < 9; index++ {
		cardID, createErr := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wordID,
			WordEN:     "banana",
			WordRU:     "банан",
			MeaningEN:  "banana",
			SenseIndex: index,
		})
		if createErr != nil {
			t.Fatalf("CreateTrainingCard error: %v", createErr)
		}

		dueAt := time.Now().Add(-time.Hour)
		_, createErr = userCardRepo.CreateUserCard(&models.UserCard{
			UserID:         user.ID,
			TrainingCardID: cardID,
			Direction:      models.DirectionRUtoEN,
			State:          models.StateReview,
			EF:             models.InitialEF,
			NextDueAt:      &dueAt,
		})
		if createErr != nil {
			t.Fatalf("CreateUserCard error: %v", createErr)
		}
	}

	client := &mockTelegramClientNS{}
	bot := newTestBotNS(client)
	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	err = service.sendNotificationIfNeeded(user, time.Now())
	if err != nil {
		t.Fatalf("sendNotificationIfNeeded error: %v", err)
	}
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Fatalf("expected no notification when due cards < 10")
	}
}

func TestNotificationService_MessageContainsStreakWhenTrainedYesterday(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	nudgeRepo := repository.NewNudgeRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)

	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}
	_, _ = conn.Exec("UPDATE users SET timezone = $1, preferred_training_time = $2 WHERE id = $3", "UTC", "00:00", user.ID)
	user, _ = userRepo.GetUserByID(user.ID)
	if user == nil {
		t.Fatal("user not found")
	}

	wordID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "streak"})
	for i := 0; i < 10; i++ {
		cardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wordID, WordEN: "streak", WordRU: "серия", MeaningEN: "streak", SenseIndex: i,
		})
		dueAt := time.Now().UTC().Add(-time.Hour)
		_, _ = userCardRepo.CreateUserCard(&models.UserCard{
			UserID: user.ID, TrainingCardID: cardID, Direction: models.DirectionRUtoEN,
			State: models.StateReview, EF: models.InitialEF, NextDueAt: &dueAt,
		})
	}

	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02") + " 12:00:00+00"
	dayBefore := now.AddDate(0, 0, -2).Format("2006-01-02") + " 12:00:00+00"
	_, _ = conn.Exec(`INSERT INTO training_sessions (user_id, source, planned_count, done_count, session_json, started_at)
		VALUES ($1, $2, $3, $4, $5, $6::timestamptz), ($1, $2, $3, $4, $5, $7::timestamptz)`,
		user.ID, "manual", 5, 0, "{}", yesterday, dayBefore)

	client := &mockTelegramClientNS{}
	bot := newTestBotNS(client)
	svc := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	err = svc.sendNotificationIfNeeded(user, now)
	if err != nil {
		t.Fatalf("sendNotificationIfNeeded: %v", err)
	}
	text := client.lastParams.Get("text")
	if text == "" {
		t.Fatal("expected notification to be sent")
	}
	if !strings.Contains(text, "дней подряд") && !strings.Contains(text, "так держать") {
		t.Errorf("expected encouraging streak phrase in message, got %q", text)
	}
	if !strings.Contains(text, "К повторению") {
		t.Errorf("expected due count in message, got %q", text)
	}
}

func TestNotificationService_SkipsWhenDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

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
	db := testutil.SetupTestDatabase(t)

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

func TestNotificationService_DisablesNotificationsOnBlockedOrMissingChat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	nudgeRepo := repository.NewNudgeRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(888)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}
	_, err = db.GetConnection().Exec(`UPDATE users SET timezone = ?, preferred_training_time = ? WHERE id = ?`, "UTC", "00:00", user.ID)
	if err != nil {
		t.Fatalf("update user error: %v", err)
	}
	settingsJSON, marshalErr := json.Marshal(models.UserSettings{NotificationFrequency: "daily"})
	if marshalErr != nil {
		t.Fatalf("marshal error: %v", marshalErr)
	}
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("UpdateUserSettings error: %v", err)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "cherry"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	for index := 0; index < 10; index++ {
		cardID, createErr := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wordID,
			WordEN:     "cherry",
			WordRU:     "вишня",
			MeaningEN:  "cherry",
			SenseIndex: index,
		})
		if createErr != nil {
			t.Fatalf("CreateTrainingCard error: %v", createErr)
		}

		dueAt := time.Now().Add(-time.Hour)
		_, createErr = userCardRepo.CreateUserCard(&models.UserCard{
			UserID:         user.ID,
			TrainingCardID: cardID,
			Direction:      models.DirectionRUtoEN,
			State:          models.StateReview,
			EF:             models.InitialEF,
			NextDueAt:      &dueAt,
		})
		if createErr != nil {
			t.Fatalf("CreateUserCard error: %v", createErr)
		}
	}

	testCases := []string{
		"failed to send message: Forbidden: bot was blocked by the user",
		"failed to send message: Bad Request: chat not found",
	}

	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	for _, errText := range testCases {
		t.Run(errText, func(t *testing.T) {
			_, err := db.GetConnection().Exec(`UPDATE users SET settings_json = ? WHERE id = ?`, string(settingsJSON), user.ID)
			if err != nil {
				t.Fatalf("failed to reset settings: %v", err)
			}
			user.SettingsJSON = string(settingsJSON)

			if !service.shouldDisableNotificationsOnError(errors.New(errText)) {
				t.Fatalf("expected error to trigger notification disable")
			}
			if disableErr := service.disableUserNotifications(user); disableErr != nil {
				t.Fatalf("disableUserNotifications error: %v", disableErr)
			}

			updated, getErr := userRepo.GetUserByID(user.ID)
			if getErr != nil {
				t.Fatalf("GetUserByID error: %v", getErr)
			}

			var settings models.UserSettings
			if unmarshalErr := json.Unmarshal([]byte(updated.SettingsJSON), &settings); unmarshalErr != nil {
				t.Fatalf("failed to parse user settings: %v", unmarshalErr)
			}
			if settings.NotificationFrequency != "never" {
				t.Fatalf("expected notification_frequency=never, got %q", settings.NotificationFrequency)
			}
		})
	}
}

func TestNotificationService_DoesNotDisableNotificationsOnOtherErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewNotificationService(nil, nil, nil, nil, nil, logger)

	if service.shouldDisableNotificationsOnError(errors.New("failed to send message: Too Many Requests")) {
		t.Fatalf("expected unrelated error to not trigger notification disable")
	}
}
