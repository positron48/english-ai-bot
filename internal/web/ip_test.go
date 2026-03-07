package web

import (
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedPrefix string
	}{
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			remoteAddr:     "127.0.0.1:12345",
			expectedPrefix: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1, 192.168.1.1",
			},
			remoteAddr:     "127.0.0.1:12345",
			expectedPrefix: "10.0.0.1",
		},
		{
			name:           "RemoteAddr fallback",
			headers:        map[string]string{},
			remoteAddr:     "192.168.1.100:54321",
			expectedPrefix: "192.168.1.100",
		},
		{
			name: "X-Real-IP takes priority over X-Forwarded-For",
			headers: map[string]string{
				"X-Real-IP":       "192.168.1.1",
				"X-Forwarded-For": "10.0.0.1",
			},
			remoteAddr:     "127.0.0.1:12345",
			expectedPrefix: "192.168.1.1",
		},
		{
			name: "IPv6 address",
			headers: map[string]string{
				"X-Real-IP": "2001:db8::1",
			},
			remoteAddr:     "127.0.0.1:12345",
			expectedPrefix: "2001:db8::1",
		},
		{
			name:           "invalid X-Real-IP falls back to RemoteAddr",
			headers:        map[string]string{"X-Real-IP": "not.an.ip"},
			remoteAddr:     "192.168.1.1:8080",
			expectedPrefix: "192.168.1.1",
		},
		{
			name:           "all invalid yields unknown",
			headers:        map[string]string{"X-Real-IP": "x", "X-Forwarded-For": "y"},
			remoteAddr:     "invalid",
			expectedPrefix: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := clientIP(req)
			if ip != tt.expectedPrefix {
				t.Errorf("Expected IP %s, got %s", tt.expectedPrefix, ip)
			}
		})
	}
}

func TestParseIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"whitespace only", "  ", ""},
		{"IPv4", "192.168.1.1", "192.168.1.1"},
		{"IPv4 with port", "192.168.1.1:8080", "192.168.1.1"},
		{"IPv6", "2001:db8::1", "2001:db8::1"},
		{"invalid", "not.an.ip", ""},
		{"invalid with port", "hostname:80", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseIP(tt.input)
			if got != tt.expected {
				t.Errorf("parseIP(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
