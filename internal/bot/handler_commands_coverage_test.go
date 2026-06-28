package bot

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const coverageChatID int64 = 900010
const coverageUserID int64 = 900001

func newCoverageHandler(t *testing.T, client interface{ Do(*http.Request) (*http.Response, error) }, cfg *config.Config) *Handler {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	courseRepo := repository.NewCourseRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(conn, logger), 5, logger)
	if cfg == nil {
		cfg = &config.Config{}
	}
	return NewHandler(
		newTestBot(client),
		logger,
		nil,
		nil,
		testDefaultCourse,
		courseRepo,
		userRepo,
		nil,
		nil,
		cbService,
		cfg,
		conn,
	)
}

func TestHandleMigrationRedirect(t *testing.T) {
	t.Run("message update", func(t *testing.T) {
		client := &mockTelegramClient{}
		cfg := &config.Config{}
		cfg.Migration.Enabled = true
		cfg.Migration.Message = "Переезжаем в @newbot"
		h := newCoverageHandler(t, client, cfg)

		h.handleMigrationRedirect(tgbotapi.Update{
			Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: coverageChatID}},
		})

		if got := lastText(client); got != cfg.Migration.Message {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("callback update", func(t *testing.T) {
		client := &mockTelegramClient{}
		cfg := &config.Config{}
		cfg.Migration.Message = "migrate notice"
		h := newCoverageHandler(t, client, cfg)

		h.handleMigrationRedirect(tgbotapi.Update{
			CallbackQuery: callbackQuery("noop", coverageUserID),
		})

		if got := lastText(client); got != cfg.Migration.Message {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("no chat id noop", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)

		h.handleMigrationRedirect(tgbotapi.Update{})

		if client.DoCount != 0 {
			t.Fatalf("expected no send, DoCount=%d", client.DoCount)
		}
	})
}

func TestHandleGetIDCommand(t *testing.T) {
	client := &mockTelegramClient{}
	h := newCoverageHandler(t, client, nil)

	h.handleGetIDCommand(coverageChatID, coverageUserID)

	got := lastText(client)
	if !strings.Contains(got, "900001") {
		t.Fatalf("got %q", got)
	}
	if client.lastParams.Get("parse_mode") != "Markdown" {
		t.Fatalf("parse_mode=%q", client.lastParams.Get("parse_mode"))
	}
}

func TestHandleCommand_allBranches(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		adminID    int64
		wantSubstr string
		silent     bool
	}{
		{"start", "/start", 0, "Welcome start", false},
		{"help", "/help", 0, "Help text here", false},
		{"language", "/language", 0, "язык", false},
		{"get_id", "/get_id", 0, "900002", false},
		{"unknown", "/nope", 0, "unknown cmd", false},
		{"reset_circuit non-admin", "/reset_circuit", 999999, "", true},
		{"reset_circuit admin", "/reset_circuit", 900003, "Circuit breaker", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockTelegramClient{}
			cfg := &config.Config{}
			cfg.Bot.StartMessage = "Welcome start"
			cfg.Bot.HelpMessage = "Help text here"
			cfg.Bot.UnknownCommandMessage = "unknown cmd"
			cfg.Admin.TelegramID = tc.adminID
			h := newCoverageHandler(t, client, cfg)

			msg := commandMessage(tc.command)
			msg.Chat.ID = coverageChatID
			msg.From.ID = 900002
			if tc.name == "reset_circuit admin" {
				msg.From.ID = 900003
			}
			if tc.name == "get_id" {
				msg.From.ID = 900002
			}

			h.handleCommand(context.Background(), msg)

			if tc.silent {
				if client.DoCount != 0 {
					t.Fatalf("expected silent, DoCount=%d", client.DoCount)
				}
				return
			}
			got := lastText(client)
			if got == "" || !strings.Contains(got, tc.wantSubstr) {
				t.Fatalf("got %q want substring %q", got, tc.wantSubstr)
			}
		})
	}
}

func TestHandleLanguageCommand_branches(t *testing.T) {
	t.Run("user repo error", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)
		h.userRepo = &stubUserRepo{getErr: io.EOF}

		h.handleLanguageCommand(context.Background(), coverageChatID, coverageUserID)

		got := lastText(client)
		if !strings.Contains(got, "Произошла ошибка") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("no active courses", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)
		if _, err := h.db.Exec("UPDATE courses SET status = 'archived'"); err != nil {
			t.Fatalf("archive courses: %v", err)
		}
		t.Cleanup(func() {
			_, _ = h.db.Exec("UPDATE courses SET status = 'active'")
		})

		h.handleLanguageCommand(context.Background(), coverageChatID, coverageUserID)

		got := lastText(client)
		if !strings.Contains(got, "Сейчас нет доступных языков") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("marks current course", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)
		user, err := h.userRepo.GetOrCreateUser(coverageUserID)
		if err != nil {
			t.Fatalf("user: %v", err)
		}
		if _, err := h.courseRepo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
			t.Fatalf("select course: %v", err)
		}

		h.handleLanguageCommand(context.Background(), coverageChatID, coverageUserID)

		got := lastText(client)
		if !strings.Contains(got, "Выберите язык") {
			t.Fatalf("got %q", got)
		}
		if !strings.Contains(string(client.LastBody), "setlang%3A") {
			t.Fatalf("expected inline keyboard payload, body=%q", client.LastBody)
		}
	})

	t.Run("shows course keyboard", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)

		h.handleLanguageCommand(context.Background(), coverageChatID, coverageUserID)

		got := lastText(client)
		if got == "" || (!strings.Contains(got, "Выберите язык") && !strings.Contains(got, "Сейчас нет")) {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("send failure logs only", func(t *testing.T) {
		client := &failingTelegramClient{}
		h := newCoverageHandler(t, client, nil)

		h.handleLanguageCommand(context.Background(), coverageChatID, coverageUserID)
	})
}

type parseEntitiesThenOKClient struct {
	mockTelegramClient
	attempts int
}

func (c *parseEntitiesThenOKClient) Do(req *http.Request) (*http.Response, error) {
	c.attempts++
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if c.attempts == 1 {
		resp := `{"ok": false, "description": "Bad Request: can't parse entities at byte offset 1"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(resp)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return c.mockTelegramClient.Do(req)
}

type parseEntitiesFailBothClient struct{}

func (c *parseEntitiesFailBothClient) Do(req *http.Request) (*http.Response, error) {
	resp := `{"ok": false, "description": "Bad Request: can't parse entities at byte offset 1"}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(resp)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

type genericSendErrorClient struct{}

func (c *genericSendErrorClient) Do(req *http.Request) (*http.Response, error) {
	resp := `{"ok": false, "description": "Bad Request: chat not found"}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(resp)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func TestSendMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &mockTelegramClient{}
		h := newCoverageHandler(t, client, nil)

		h.sendMessage(coverageChatID, "hello *world*")

		if got := lastText(client); got != "hello *world*" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("parse entities fallback", func(t *testing.T) {
		client := &parseEntitiesThenOKClient{}
		h := newCoverageHandler(t, client, nil)

		h.sendMessage(coverageChatID, "bad _markdown")

		if client.attempts < 2 {
			t.Fatalf("expected retry, attempts=%d", client.attempts)
		}
	})

	t.Run("parse entities fallback also fails", func(t *testing.T) {
		client := &parseEntitiesFailBothClient{}
		h := newCoverageHandler(t, client, nil)

		h.sendMessage(coverageChatID, "bad _markdown")
	})

	t.Run("non parse error", func(t *testing.T) {
		client := &genericSendErrorClient{}
		h := newCoverageHandler(t, client, nil)

		h.sendMessage(coverageChatID, "plain text")
	})
}

func TestHandleUpdate_migrationEnabled(t *testing.T) {
	client := &mockTelegramClient{}
	cfg := &config.Config{}
	cfg.Migration.Enabled = true
	cfg.Migration.Message = "migrated"
	h := newCoverageHandler(t, client, cfg)

	h.HandleUpdate(context.Background(), tgbotapi.Update{
		Message: commandMessage("/start"),
	})

	if got := lastText(client); got != "migrated" {
		t.Fatalf("got %q", got)
	}
}
