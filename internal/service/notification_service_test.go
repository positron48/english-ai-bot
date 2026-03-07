package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestNewNotificationService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	bot := (*tgbotapi.BotAPI)(nil)
	userRepo := (*repository.UserRepository)(nil)
	userCardRepo := (*repository.UserCardRepository)(nil)
	nudgeRepo := (*repository.NudgeRepository)(nil)
	sessionRepo := (*repository.SessionRepository)(nil)

	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	_ = service // Verify service is created
}

func TestNotificationService_CheckAndSendNotifications_NoUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	// No users in DB — checkAndSendNotifications should not panic and should return without sending
	service.checkAndSendNotifications()
}

// TestNotificationService_CheckAndSendNotifications_InvalidTimezone covers user with invalid timezone (uses UTC).
func TestNotificationService_CheckAndSendNotifications_InvalidTimezone(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(7777)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, err = db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "Invalid/Zone", user.ID)
	if err != nil {
		t.Fatalf("UPDATE timezone: %v", err)
	}
	_ = userRepo.UpdateUserPreferredTime(user.ID, "19:00")
	service.checkAndSendNotifications()
	// Should not panic; invalid timezone path uses UTC
}

// TestNotificationService_CheckAndSendNotifications_InvalidPreferredTime covers user with unparseable preferred time (uses 19:00).
func TestNotificationService_CheckAndSendNotifications_InvalidPreferredTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(7778)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, _ = db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "UTC", user.ID)
	_ = userRepo.UpdateUserPreferredTime(user.ID, "25:99")
	service.checkAndSendNotifications()
	// Should not panic; invalid preferred time path uses 19:00
}

// TestNotificationService_SendNotificationIfNeeded_CustomFrequencyNotEnoughDays covers N-day frequency when not enough days passed.
func TestNotificationService_SendNotificationIfNeeded_CustomFrequencyNotEnoughDays(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(7779)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	settings := models.UserSettings{
		NotificationFrequency: "5",
		LastNotificationDate:  time.Now().UTC().AddDate(0, 0, -2).Format("2006-01-02"),
	}
	js, _ := json.Marshal(settings)
	_ = userRepo.UpdateUserSettings(user.ID, string(js))
	user, _ = userRepo.GetUserByTelegramID(7779)
	userNow := time.Now().UTC()

	err = svc.sendNotificationIfNeeded(user, userNow)
	if err != nil {
		t.Errorf("sendNotificationIfNeeded with custom frequency (not enough days) should return nil, got %v", err)
	}
}

// TestBuildNotificationMessage covers all branches: minute wording, weekLine, streak/trainedYesterday, no streak.
func TestBuildNotificationMessage(t *testing.T) {
	logger := zap.NewNop()
	svc := NewNotificationService(nil, nil, nil, nil, nil, logger)

	tests := []struct {
		name            string
		streak          int
		trainedYesterday bool
		newCardsWeek    int
		dueCount        int
		estimatedMin    int
		wantContains    []string
	}{
		{"1 minute wording", 0, false, 0, 15, 1, []string{"минута", "К повторению: 15", "Начать?"}},
		{"2-4 minutes wording", 0, false, 0, 20, 3, []string{"минуты", "К повторению: 20", "~3"}},
		{"5+ minutes wording", 0, false, 0, 50, 13, []string{"минут", "~13"}},
		{"week growth line", 0, false, 5, 10, 3, []string{"За неделю +5 слов", "К повторению: 10"}},
		{"streak 1 trained yesterday", 1, true, 0, 12, 2, []string{"день подряд в деле", "К повторению: 12"}},
		{"streak 1 trained yesterday with week", 1, true, 3, 10, 2, []string{"день подряд", "За неделю", "Продолжим?"}},
		{"streak 2+ trained yesterday", 2, true, 0, 10, 2, []string{"Уже 2 дней подряд", "Продолжим?"}},
		{"streak 2+ trained yesterday with week", 2, true, 1, 15, 4, []string{"Уже 2 дней подряд", "+1 слов в словаре"}},
		{"streak broken invite back", 1, false, 0, 20, 5, []string{"отличный день, чтобы вернуться", "Начать?"}},
		{"streak broken with week", 1, false, 2, 18, 4, []string{"вернуться", "За неделю ты уже добавил 2 слов"}},
		{"no streak new user with week", 0, false, 1, 10, 2, []string{"За неделю", "Начать?"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := svc.buildNotificationMessage(tt.streak, tt.trainedYesterday, tt.newCardsWeek, tt.dueCount, tt.estimatedMin)
			for _, sub := range tt.wantContains {
				if len(msg) == 0 {
					t.Fatalf("buildNotificationMessage returned empty")
				}
				if !strings.Contains(msg, sub) {
					t.Errorf("message %q should contain %q", msg, sub)
				}
			}
		})
	}
}

func TestNotificationService_ShouldDisableNotificationsOnError(t *testing.T) {
	logger := zap.NewNop()
	svc := NewNotificationService(nil, nil, nil, nil, nil, logger)

	if svc.shouldDisableNotificationsOnError(nil) {
		t.Error("nil error should not disable")
	}
	if !svc.shouldDisableNotificationsOnError(assertAnError("forbidden: bot was blocked by the user")) {
		t.Error("blocked by user should disable")
	}
	if !svc.shouldDisableNotificationsOnError(assertAnError("bad request: chat not found")) {
		t.Error("chat not found should disable")
	}
	if svc.shouldDisableNotificationsOnError(assertAnError("some other error")) {
		t.Error("other error should not disable")
	}
}

func assertAnError(s string) error { return &errString{s} }
type errString struct{ s string }
func (e *errString) Error() string { return e.s }

func TestNotificationService_SendNotificationIfNeeded_Never(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(1001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	settings := models.UserSettings{NotificationFrequency: "never"}
	js, _ := json.Marshal(settings)
	if err := userRepo.UpdateUserSettings(user.ID, string(js)); err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}
	user, _ = userRepo.GetUserByTelegramID(1001)
	userNow := time.Now().UTC().Truncate(time.Hour).Add(20 * time.Hour) // 20:00 UTC

	err = svc.sendNotificationIfNeeded(user, userNow)
	if err != nil {
		t.Errorf("sendNotificationIfNeeded with never should return nil, got %v", err)
	}
}

func TestNotificationService_SendNotificationIfNeeded_DueCountLow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(1002)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_ = userRepo.UpdateUserPreferredTime(user.ID, "00:00")
	_, _ = db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "UTC", user.ID)
	user, _ = userRepo.GetUserByTelegramID(1002)
	userNow := time.Now().UTC().Truncate(24 * time.Hour).Add(12 * time.Hour) // noon today

	// No user_cards => GetDueCount returns 0 => sendNotificationIfNeeded returns nil (not enough cards)
	err = svc.sendNotificationIfNeeded(user, userNow)
	if err != nil {
		t.Errorf("sendNotificationIfNeeded with dueCount < 10 should return nil, got %v", err)
	}
}

func TestNotificationService_SendNotificationIfNeeded_HasNudgeToday(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(1003)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	settings := models.UserSettings{NotificationFrequency: "daily"}
	js, _ := json.Marshal(settings)
	_ = userRepo.UpdateUserSettings(user.ID, string(js))
	_ = userRepo.UpdateUserPreferredTime(user.ID, "00:00")
	_, _ = db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "UTC", user.ID)
	localDate := time.Now().UTC().Format("2006-01-02")
	nudge := &models.TrainingNudge{UserID: user.ID, LocalDate: localDate, DueCountAtSend: 15}
	if _, err := nudgeRepo.CreateNudge(nudge); err != nil {
		t.Fatalf("CreateNudge: %v", err)
	}
	user, _ = userRepo.GetUserByTelegramID(1003)
	userNow := time.Now().UTC().Truncate(24 * time.Hour).Add(14 * time.Hour)

	err = svc.sendNotificationIfNeeded(user, userNow)
	if err != nil {
		t.Errorf("sendNotificationIfNeeded when already nudged today should return nil, got %v", err)
	}
}

func TestNotificationService_CheckAndSendNotifications_WithUserInWindow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	user, err := userRepo.GetOrCreateUser(1004)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// Preferred 00:00 UTC so any hour is in window; dueCount will be 0 so we never actually send
	_ = userRepo.UpdateUserPreferredTime(user.ID, "00:00")
	_, _ = db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "UTC", user.ID)
	settings := models.UserSettings{NotificationFrequency: "daily"}
	js, _ := json.Marshal(settings)
	_ = userRepo.UpdateUserSettings(user.ID, string(js))

	svc.checkAndSendNotifications()
	// No panic; with nil bot we skip send and return nil; with 0 due cards we might not reach send
}

func TestNotificationService_DisableUserNotifications(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	svc := NewNotificationService(nil, userRepo, nil, nudgeRepo, nil, logger)

	user, err := userRepo.GetOrCreateUser(1005)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	settings := models.UserSettings{NotificationFrequency: "daily"}
	js, _ := json.Marshal(settings)
	_ = userRepo.UpdateUserSettings(user.ID, string(js))
	user, _ = userRepo.GetUserByTelegramID(1005)

	err = svc.disableUserNotifications(user)
	if err != nil {
		t.Fatalf("disableUserNotifications: %v", err)
	}
	user, _ = userRepo.GetUserByTelegramID(1005)
	if !strings.Contains(user.SettingsJSON, "never") {
		t.Errorf("expected settings to contain never, got %q", user.SettingsJSON)
	}
}
