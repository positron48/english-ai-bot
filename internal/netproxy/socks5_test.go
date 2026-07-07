package netproxy

import (
	"testing"
	"time"
)

func TestNewHTTPClientWithoutProxy(t *testing.T) {
	client, err := NewHTTPClient(3*time.Second, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Timeout != 3*time.Second {
		t.Fatalf("timeout = %s, want 3s", client.Timeout)
	}
	if client.Transport != nil {
		t.Fatalf("transport should be nil without proxy")
	}
}

func TestNewSocks5TransportRejectsUnsupportedScheme(t *testing.T) {
	if _, err := NewSocks5Transport("http://127.0.0.1:1080"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestNormalizeSocks5ProxyAcceptsAuthURL(t *testing.T) {
	hostPort, auth, err := normalizeSocks5Proxy("socks5h://user:pass@127.0.0.1:1080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostPort != "127.0.0.1:1080" {
		t.Fatalf("hostPort = %q", hostPort)
	}
	if auth == nil || auth.User != "user" || auth.Password != "pass" {
		t.Fatalf("auth = %#v", auth)
	}
}
