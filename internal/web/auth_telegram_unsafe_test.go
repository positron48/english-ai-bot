package web

import (
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tgbot-skeleton/internal/config"
)

func TestUnsignedTelegramRouteRejected(t *testing.T) {
	router := NewRouter(zap.NewNop(), &config.Config{}, nil, nil, nil, nil, nil)
	for _, body := range []string{"user_id=12345", "", "user_id=invalid"} {
		req := httptest.NewRequest(http.MethodPost, "/auth/telegram_unsafe", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusGone || strings.Contains(w.Body.String(), "access_token") {
			t.Fatalf("unsafe auth: %d %s", w.Code, w.Body.String())
		}
	}
}
