package readingcms

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServerHealth(t *testing.T) {
	root := t.TempDir()
	courseDir := filepath.Join(root, "courses", "english-grammar")
	if err := os.MkdirAll(filepath.Join(courseDir, "reading", "texts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(courseDir, "reading", "index.json"), []byte(`{"version":"1.0.0","categories":{},"texts":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServer(svc, "")
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d", rec.Code)
	}
}
