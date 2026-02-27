package web

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (r *Router) handleAdminTTS(w http.ResponseWriter, req *http.Request) {
	ctx := r.loadUserPermissionsIntoContext(req.Context())
	req = req.WithContext(ctx)

	if r.pronunciationService == nil || !r.pronunciationService.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/api/admin/tts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.Error(w, "word is required", http.StatusBadRequest)
		return
	}
	word := strings.TrimSpace(parts[0])
	action := ""
	if len(parts) > 1 {
		action = strings.TrimSpace(parts[1])
	}

	switch req.Method {
	case http.MethodGet:
		if !r.HasPermission(req.Context(), PermissionWordsReadAll) {
			http.Error(w, "Forbidden: read permission required", http.StatusForbidden)
			return
		}
		if action != "" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status, err := r.pronunciationService.GetStatus(word)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeTTSStatusResponse(w, status)
		return
	case http.MethodPost:
		if !r.HasPermission(req.Context(), PermissionWordsEditAll) {
			http.Error(w, "Forbidden: edit permission required", http.StatusForbidden)
			return
		}
		switch action {
		case "regenerate":
			status, err := r.pronunciationService.ForceRegenerate(word)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeTTSStatusResponse(w, status)
			return
		case "recheck":
			status, err := r.pronunciationService.Recheck(word)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeTTSStatusResponse(w, status)
			return
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeTTSStatusResponse(w http.ResponseWriter, status interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
