package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestNormalizePronunciationWord(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "Spy", want: "spy", ok: true},
		{in: "to  Spy ", want: "to spy", ok: true},
		{in: "can't", want: "can't", ok: true},
		{in: "re-entry", want: "re-entry", ok: true},
		{in: "привет", want: "", ok: false},
		{in: "123", want: "", ok: false},
		{in: "spy!", want: "spy", ok: true},
	}

	for _, tt := range tests {
		got, ok := normalizePronunciationWord(tt.in)
		if ok != tt.ok {
			t.Fatalf("normalizePronunciationWord(%q) ok=%v want %v", tt.in, ok, tt.ok)
		}
		if got != tt.want {
			t.Fatalf("normalizePronunciationWord(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestPronunciationServiceLookupAndCache(t *testing.T) {
	audioBytes := []byte("ID3mock-audio")
	var lookupCalls int
	var audioCalls int

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v2/entries/en/"):
			lookupCalls++
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"` + srv.URL + `/audio/spy.mp3"}]}]`))
		case r.URL.Path == "/audio/spy.mp3":
			audioCalls++
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write(audioBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          tmpDir,
		PublicBasePath:    "/media/tts",
		RequestTimeout:    "3s",
		PrefetchEnabled:   false,
		PrefetchWorkers:   1,
		RetryBaseDelay:    "100ms",
		RetryMaxDelay:     "1s",
		MaxRetries:        2,
		DictionaryEnabled: true,
		DictionaryBaseURL: srv.URL + "/api/v2/entries/en",
	}

	logger := zap.NewNop()
	service := NewPronunciationService(cfg, nil, logger)
	if !service.IsEnabled() {
		t.Fatalf("service should be enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)

	initial := service.Lookup("spy")
	if initial.Available {
		t.Fatalf("expected pronunciation to be initially unavailable")
	}

	waitFor := time.After(4 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var final PronunciationLookupResult
	found := false
	for !found {
		select {
		case <-waitFor:
			t.Fatalf("timed out waiting for pronunciation file")
		case <-ticker.C:
			final = service.Lookup("spy")
			found = final.Available
		}
	}

	if !strings.HasPrefix(final.URL, "/media/tts/") {
		t.Fatalf("unexpected pronunciation URL: %q", final.URL)
	}

	fullPath := filepath.Join(tmpDir, strings.TrimPrefix(final.URL, "/media/tts/"))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read cached pronunciation: %v", err)
	}
	if string(data) != string(audioBytes) {
		t.Fatalf("unexpected cached audio bytes")
	}
	if lookupCalls == 0 || audioCalls == 0 {
		t.Fatalf("expected dictionary and audio endpoints to be called at least once")
	}
}

func TestPronunciationServiceBackoffOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		RequestTimeout:    "3s",
		PrefetchEnabled:   false,
		PrefetchWorkers:   1,
		RetryBaseDelay:    "100ms",
		RetryMaxDelay:     "1s",
		MaxRetries:        3,
		DictionaryEnabled: true,
		DictionaryBaseURL: srv.URL + "/api/v2/entries/en",
	}

	service := NewPronunciationService(cfg, nil, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)

	_ = service.ScheduleWord("missingword")

	waitFor := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-waitFor:
			t.Fatalf("timed out waiting for retry state")
		case <-ticker.C:
			service.mu.Lock()
			retry, ok := service.retries["missingword"]
			service.mu.Unlock()
			if ok {
				if retry.attempt <= 0 {
					t.Fatalf("expected retry attempt > 0")
				}
				if !retry.nextTry.After(time.Now()) {
					t.Fatalf("expected next retry in the future")
				}
				return
			}
		}
	}
}
