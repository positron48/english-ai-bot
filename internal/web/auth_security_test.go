package web

import (
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"strconv"
	"testing"
	"tgbot-skeleton/internal/config"
	"time"
)

func TestJWTTokenPurposeIsolation(t *testing.T) {
	svc, _ := NewJWTService(&config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-only"}}, zap.NewNop())
	access, _ := svc.GenerateToken(42, []int64{1})
	refresh, _ := svc.GenerateRefreshToken(42)
	if _, err := svc.ValidateRefreshToken(access); err == nil {
		t.Fatal("access token accepted for refresh")
	}
	if _, _, err := svc.ValidateToken(refresh); err == nil {
		t.Fatal("refresh token accepted as access")
	}
	legacy := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 42, "iss": "english-bot", "exp": time.Now().Add(time.Hour).Unix()})
	raw, _ := legacy.SignedString(svc.secret)
	if _, _, err := svc.ValidateToken(raw); err == nil {
		t.Fatal("untyped token accepted")
	}
	if _, err := svc.ValidateRefreshToken(raw); err == nil {
		t.Fatal("untyped token accepted for refresh")
	}
}

func TestTelegramInitDataLifetime(t *testing.T) {
	middleware := &AuthMiddleware{botToken: "test-only"}
	for _, test := range []struct {
		name, date string
		valid      bool
	}{
		{"fresh", strconv.FormatInt(time.Now().Unix(), 10), true},
		{"expired", "1", false}, {"missing", "", false}, {"invalid", "abc", false},
		{"future", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			signed := buildInitDataWithHash(middleware.botToken, map[string]string{"auth_date": test.date, "user": `{"id":12345}`})
			_, err := middleware.ValidateTelegramInitData(signed)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v err=%v", test.valid, err)
			}
		})
	}
}
