package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"

	"tgbot-skeleton/internal/config"
)

func TestHandleMediaServeCacheHeaders(t *testing.T) {
	dir := t.TempDir()
	pictureDir := filepath.Join(dir, "picture")
	if err := os.MkdirAll(pictureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pictureDir, "ok.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := &Router{
		logger: zap.NewNop(),
		config: &config.Config{Media: config.MediaConfig{
			Dir:            dir,
			PublicBasePath: "/api/media",
		}},
	}

	existing := httptest.NewRecorder()
	router.handleMediaServe(existing, httptest.NewRequest(http.MethodGet, "/api/media/picture/ok.png", nil))
	if existing.Code != http.StatusOK {
		t.Fatalf("existing media status = %d, want 200", existing.Code)
	}
	if got := existing.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("existing media Cache-Control = %q, want immutable", got)
	}

	missing := httptest.NewRecorder()
	router.handleMediaServe(missing, httptest.NewRequest(http.MethodGet, "/api/media/picture/missing.png", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing media status = %d, want 404", missing.Code)
	}
	if got := missing.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("missing media Cache-Control = %q, want no-store", got)
	}
}
