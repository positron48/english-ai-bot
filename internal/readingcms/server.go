package readingcms

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server is the local HTTP API for Reading CMS.
type Server struct {
	svc     *Service
	webRoot string
}

func NewServer(svc *Service, webRoot string) *Server {
	return &Server{svc: svc, webRoot: webRoot}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/courses", s.handleCourses)
	mux.HandleFunc("/api/drafts", s.handleDrafts)
	mux.HandleFunc("/api/drafts/generate", s.handleGenerate)
	mux.HandleFunc("/api/drafts/import-text", s.handleImportText)
	mux.HandleFunc("/api/drafts/import-json", s.handleImportJSON)
	mux.HandleFunc("/api/prompts/reading", s.handleReadingPrompt)
	mux.HandleFunc("/api/published", s.handlePublished)
	mux.HandleFunc("/api/audio/", s.handleAudio)
	mux.HandleFunc("/api/drafts/", s.handleDraftSubroutes)
	if s.webRoot != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.webRoot)))
	}
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, map[string]interface{}{"courses": s.svc.Courses()})
}

func (s *Server) handleDrafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	q := r.URL.Query()
	items, err := s.svc.ListDrafts(q.Get("course_code"), q.Get("level"), q.Get("status"), q.Get("audio"), q.Get("search"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]interface{}{"drafts": items, "total": len(items)})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req GenerateRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.GenerateBatch(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]interface{}{"drafts": created, "total": len(created)})
}

func (s *Server) handleImportText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req ImportTextRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	meta, doc, err := s.svc.ImportPlainText(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]interface{}{"draft": meta, "document": doc})
}

func (s *Server) handleImportJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req ImportJSONRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	meta, doc, err := s.svc.ImportJSON(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]interface{}{"draft": meta, "document": doc})
}

func (s *Server) handleReadingPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req PromptRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	prompt, err := s.svc.ReadingPrompt(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]interface{}{
		"prompt":      prompt,
		"course_code": req.CourseCode,
		"kind":        req.Kind,
	})
}

func (s *Server) handlePublished(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		items, err := s.svc.ListPublished(q.Get("course_code"), q.Get("level"), q.Get("search"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]interface{}{"texts": items, "total": len(items)})
	case http.MethodDelete:
		q := r.URL.Query()
		textID := strings.TrimSpace(q.Get("text_id"))
		courseCode := strings.TrimSpace(q.Get("course_code"))
		if textID == "" || courseCode == "" {
			writeError(w, http.StatusBadRequest, errInvalid("text_id and course_code required"))
			return
		}
		if err := s.svc.DeletePublished(courseCode, textID, true); err != nil {
			if IsNotFound(err) {
				writeError(w, http.StatusNotFound, err)
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/audio/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid audio path", http.StatusBadRequest)
		return
	}
	textID := parts[0]
	filename := parts[len(parts)-1]
	path := filepath.Join(s.svc.Paths().StagingDir(textID), "assets", "reading", textID, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	_, _ = w.Write(data)
}

func (s *Server) handleDraftSubroutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/drafts/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	textID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		s.handleDraftByID(w, r, textID)
	case "approve":
		s.handleDraftAction(w, r, textID, "approve")
	case "reject":
		s.handleDraftAction(w, r, textID, "reject")
	case "audio":
		s.handleDraftAction(w, r, textID, "audio")
	case "publish":
		s.handlePublish(w, r, textID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleDraftByID(w http.ResponseWriter, r *http.Request, textID string) {
	switch r.Method {
	case http.MethodGet:
		meta, doc, err := s.svc.GetDraft(textID)
		if err != nil {
			if IsNotFound(err) {
				writeError(w, http.StatusNotFound, err)
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]interface{}{"draft": meta, "document": doc})
	case http.MethodPut:
		var body struct {
			Document *TextDocument `json:"document"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		meta, err := s.svc.SaveDocument(textID, body.Document)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, map[string]interface{}{"draft": meta})
	case http.MethodDelete:
		if err := s.svc.DeleteDraft(textID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleDraftAction(w http.ResponseWriter, r *http.Request, textID, action string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var meta *DraftMeta
	var err error
	switch action {
	case "approve":
		meta, err = s.svc.Approve(textID)
	case "reject":
		meta, err = s.svc.Reject(textID)
	case "audio":
		meta, err = s.svc.GenerateAudio(context.Background(), textID)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		if IsNotFound(err) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]interface{}{"draft": meta})
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request, textID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var body struct {
		SyncBundle bool `json:"sync_bundle"`
	}
	_ = readJSON(r, &body)
	meta, err := s.svc.Publish(textID, body.SyncBundle)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]interface{}{"draft": meta})
}

func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, dst)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
