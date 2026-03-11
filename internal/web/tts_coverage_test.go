package web

// Tests to cover remaining branches in tts.go not covered by tts_test.go.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

// TestHandleTTSMedia_EmptyRelative covers line 82-85:
// When relative path is empty (request to /media/tts/ with no file) → 404.
func TestHandleTTSMedia_EmptyRelative(t *testing.T) {
	mock := &mockPronunciationService{
		enabled:    true,
		publicBase: "/media/tts",
		audioDir:   "/tmp/tts",
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: mock}

	// Request to the base path with trailing slash → relative = ""
	req := httptest.NewRequest(http.MethodGet, "/media/tts/", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty relative path, got %d", w.Code)
	}
}

// TestHandleTTSMedia_NoMp3Extension covers line 82-85:
// When relative path has no .mp3 extension → 404.
func TestHandleTTSMedia_NoMp3Extension(t *testing.T) {
	mock := &mockPronunciationService{
		enabled:    true,
		publicBase: "/media/tts",
		audioDir:   "/tmp/tts",
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: mock}

	// Request to a file without .mp3 extension
	req := httptest.NewRequest(http.MethodGet, "/media/tts/word.txt", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-mp3 file, got %d", w.Code)
	}
}

// TestHandleTTSMedia_PathTraversalEscapesRoot covers lines 89-92:
// The cleaned target path does not start with root+separator → 404.
// This is triggered when AudioDir is empty: root = "." and target = "file.mp3"
// which does not start with "./", hitting the security check.
// (Same scenario as TestHandleTTSMedia_PathTraversal in tts_test.go but with
// a named file to exercise the branch more explicitly.)
func TestHandleTTSMedia_PathTraversalEscapesRoot(t *testing.T) {
	// With empty AudioDir, root = filepath.Clean("") = "."
	// target = filepath.Clean(filepath.Join(".", "word.mp3")) = "word.mp3"
	// "word.mp3" does not start with "./" → security check triggers → 404
	mock := &mockPronunciationService{
		enabled:    true,
		publicBase: "/media/tts",
		audioDir:   "", // empty → root = "." → target doesn't start with "./"
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: mock}

	req := httptest.NewRequest(http.MethodGet, "/media/tts/word.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when target not under root, got %d", w.Code)
	}
}
