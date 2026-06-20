package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// setupPronunciationMediaRoute registers static media route for cached pronunciation files.
func (r *Router) setupPronunciationMediaRoute() {
	if r.pronunciationMediaRouteRegistered {
		return
	}
	if r.pronunciationService == nil || !r.pronunciationService.IsEnabled() {
		return
	}

	basePath := r.pronunciationService.PublicBasePath()
	if basePath == "" {
		basePath = "/media/tts"
	}
	basePath = "/" + strings.Trim(basePath, "/")
	routePrefix := basePath + "/"
	r.mux.HandleFunc(routePrefix, r.handleTTSMedia)
	r.pronunciationMediaRouteRegistered = true
}

// handleTTSWord returns pronunciation availability and URL for a given word.
func (r *Router) handleTTSWord(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	word := strings.TrimSpace(req.URL.Query().Get("word"))
	if word == "" {
		http.Error(w, "word is required", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"available": false,
		"url":       "",
		"word":      "",
	}

	userID := getUserIDFromContext(req.Context())
	svc := r.pronunciationServiceForRequest(req, userID)
	if svc != nil && svc.IsEnabled() {
		lookup := svc.Lookup(word)
		result["available"] = lookup.Available
		result["url"] = lookup.URL
		result["word"] = lookup.NormalizedWord
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// handleTTSMedia serves cached pronunciation mp3 files with long-lived cache headers.
func (r *Router) handleTTSMedia(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.pronunciationService == nil || !r.pronunciationService.IsEnabled() {
		http.NotFound(w, req)
		return
	}

	basePath := r.pronunciationService.PublicBasePath()
	if basePath == "" {
		basePath = "/media/tts"
	}
	basePath = "/" + strings.Trim(basePath, "/")
	prefix := basePath + "/"
	relative := strings.TrimPrefix(req.URL.Path, prefix)
	relative = filepath.ToSlash(relative)
	relative = strings.TrimPrefix(filepath.Clean("/"+relative), "/")

	if relative == "" || strings.HasSuffix(relative, "/") || filepath.Ext(relative) != ".mp3" {
		http.NotFound(w, req)
		return
	}

	root := filepath.Clean(r.pronunciationService.AudioDir())
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	if !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, req, target)
}
