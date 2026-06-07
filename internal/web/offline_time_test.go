package web

import (
	"testing"
	"time"
)

func TestParseOfflineTimestamp(t *testing.T) {
	t.Parallel()
	raw := "2026-06-06T18:30:00.000Z"
	got := parseOfflineTimestamp("", raw)
	if got == nil {
		t.Fatal("expected parsed timestamp")
	}
	want := time.Date(2026, 6, 6, 18, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if second := parseOfflineTimestamp(raw, "2026-06-07T10:00:00Z"); second == nil || !second.Equal(want) {
		t.Fatalf("expected first non-empty value to win, got %v", second)
	}
}

func TestOfflineTimestampOrNow(t *testing.T) {
	t.Parallel()
	before := time.Now().UTC()
	got := offlineTimestampOrNow("2026-06-06T18:30:00Z")
	want := time.Date(2026, 6, 6, 18, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	now := offlineTimestampOrNow("")
	if now.Before(before) {
		t.Fatalf("expected current time fallback, got %v before %v", now, before)
	}
}
