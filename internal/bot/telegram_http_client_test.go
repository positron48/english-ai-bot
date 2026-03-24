package bot

import (
	"testing"
	"time"
)

func TestTelegramLongPollHTTPTimeouts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		pollSec     int
		wantHeader  time.Duration
		wantClient  time.Duration
	}{
		{"zero uses default 30s poll", 0, 30*time.Second + 25*time.Second, 30*time.Second + 25*time.Second + 20*time.Second},
		{"30s poll", 30, 30*time.Second + 25*time.Second, 30*time.Second + 25*time.Second + 20*time.Second},
		{"50s poll capped", 50, 50*time.Second + 25*time.Second, 50*time.Second + 25*time.Second + 20*time.Second},
		{"above 50 capped to 50", 999, 50*time.Second + 25*time.Second, 50*time.Second + 25*time.Second + 20*time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotH, gotC := telegramLongPollHTTPTimeouts(tt.pollSec)
			if gotH != tt.wantHeader {
				t.Fatalf("response header timeout: got %v want %v", gotH, tt.wantHeader)
			}
			if gotC != tt.wantClient {
				t.Fatalf("client timeout: got %v want %v", gotC, tt.wantClient)
			}
		})
	}
}
