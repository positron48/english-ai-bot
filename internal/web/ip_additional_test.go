package web

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP_WithXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")

	ip := clientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got %q", ip)
	}
}

func TestClientIP_WithXRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "172.16.0.1")

	ip := clientIP(req)
	if ip != "172.16.0.1" {
		t.Errorf("Expected IP '172.16.0.1', got %q", ip)
	}
}

func TestClientIP_FromRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	ip := clientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("Expected IP '127.0.0.1', got %q", ip)
	}
}

func TestClientIP_XRealIPTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.Header.Set("X-Real-IP", "172.16.0.1")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := clientIP(req)
	if ip != "172.16.0.1" {
		t.Errorf("Expected IP '172.16.0.1' (X-Real-IP takes precedence), got %q", ip)
	}
}
