package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeyFuncIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	key := KeyFuncIP(req)
	if key != "ip:192.168.1.1" {
		t.Errorf("Expected key 'ip:192.168.1.1', got %q", key)
	}
}

func TestKeyFuncIPAndUserID(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "10.0.0.1:54321"
	req.ParseForm()
	req.Form.Set("user_id", "12345")

	key := KeyFuncIPAndUserID(req)
	if key != "ip:10.0.0.1:user:12345" {
		t.Errorf("Expected key 'ip:10.0.0.1:user:12345', got %q", key)
	}
}

func TestKeyFuncIPAndUserID_NoUserID(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "10.0.0.1:54321"

	key := KeyFuncIPAndUserID(req)
	if key != "ip:10.0.0.1" {
		t.Errorf("Expected key 'ip:10.0.0.1', got %q", key)
	}
}

func TestKeyFuncIPAndUsername(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "172.16.0.1:9999"
	req.ParseForm()
	req.Form.Set("username", "testuser")

	key := KeyFuncIPAndUsername(req)
	if key != "ip:172.16.0.1:user:testuser" {
		t.Errorf("Expected key 'ip:172.16.0.1:user:testuser', got %q", key)
	}
}

func TestKeyFuncIPAndUsername_NoUsername(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "10.0.0.2:8080"

	key := KeyFuncIPAndUsername(req)
	if key != "ip:10.0.0.2" {
		t.Errorf("Expected key 'ip:10.0.0.2', got %q", key)
	}
}

func TestKeyFuncIPAndUsername_UsernameFromJSONBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"username":"jsonuser"}`))
	req.RemoteAddr = "192.168.1.10:443"
	req.Header.Set("Content-Type", "application/json")

	key := KeyFuncIPAndUsername(req)
	if key != "ip:192.168.1.10:user:jsonuser" {
		t.Errorf("Expected key 'ip:192.168.1.10:user:jsonuser', got %q", key)
	}
}

func TestKeyFuncIPAndUserIDFromContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	ctx := req.Context()
	ctx = context.WithValue(ctx, userIDKey, int64(999))
	req = req.WithContext(ctx)

	key := KeyFuncIPAndUserIDFromContext(req)
	if key != "ip:127.0.0.1:user:999" {
		t.Errorf("Expected key 'ip:127.0.0.1:user:999', got %q", key)
	}
}

func TestKeyFuncIPAndUserIDFromContext_NoUserID(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	key := KeyFuncIPAndUserIDFromContext(req)
	if key != "ip:127.0.0.1" {
		t.Errorf("Expected key 'ip:127.0.0.1', got %q", key)
	}
}
