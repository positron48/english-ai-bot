package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

const maxMediaUploadBytes = 5 << 20 // 5 MB

var allowedMediaExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

func (r *Router) mediaDir() string {
	if r.config != nil && r.config.Media.Dir != "" {
		return r.config.Media.Dir
	}
	return "./data/media"
}

func (r *Router) mediaPublicBasePath() string {
	if r.config != nil && r.config.Media.PublicBasePath != "" {
		return "/" + strings.Trim(r.config.Media.PublicBasePath, "/")
	}
	return "/api/media"
}

// handleAdminMediaUpload handles POST /api/admin/media/upload?type=npc|quest.
// Accepts a multipart "file" field, saves it under MEDIA_DIR/<type>/<random>.<ext>,
// and returns the public URL.
func (r *Router) handleAdminMediaUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	kind := strings.TrimSpace(req.URL.Query().Get("type"))
	if kind != "npc" && kind != "quest" {
		http.Error(w, "type must be npc or quest", http.StatusBadRequest)
		return
	}
	if err := req.ParseMultipartForm(maxMediaUploadBytes); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedMediaExtensions[ext] {
		http.Error(w, "Unsupported file type (png/jpg/jpeg/webp only)", http.StatusBadRequest)
		return
	}

	destDir := filepath.Join(r.mediaDir(), kind)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		r.logger.Error("create media dir", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	name, err := randomHexFilename(ext)
	if err != nil {
		r.logger.Error("generate media filename", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	out, err := os.Create(filepath.Join(destDir, name))
	if err != nil {
		r.logger.Error("create media file", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(file, maxMediaUploadBytes+1)); err != nil {
		r.logger.Error("save media file", zap.Error(err))
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	url := r.mediaPublicBasePath() + "/" + kind + "/" + name
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// handleMediaServe serves uploaded NPC/quest images with path-traversal protection.
func (r *Router) handleMediaServe(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	prefix := r.mediaPublicBasePath() + "/"
	relative := strings.TrimPrefix(req.URL.Path, prefix)
	relative = filepath.ToSlash(relative)
	relative = strings.TrimPrefix(filepath.Clean("/"+relative), "/")

	if relative == "" || strings.HasSuffix(relative, "/") {
		http.NotFound(w, req)
		return
	}
	ext := strings.ToLower(filepath.Ext(relative))
	if !allowedMediaExtensions[ext] {
		http.NotFound(w, req)
		return
	}

	root := filepath.Clean(r.mediaDir())
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	if !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, req, target)
}

func randomHexFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}
