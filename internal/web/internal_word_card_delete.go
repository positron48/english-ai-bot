package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleInternalWordCardDelete deletes one verified dictionary entry, including
// its training cards and review history. Omitted dry_run means preview only.
func (r *Router) handleInternalWordCardDelete(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawID := strings.TrimPrefix(req.URL.Path, "/api/internal/word-cards/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid word card id", http.StatusBadRequest)
		return
	}
	body := struct {
		CourseCode   string `json:"course_code"`
		ExpectedWord string `json:"expected_word"`
		DryRun       *bool  `json:"dry_run"`
	}{}
	decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Expected one JSON object", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.CourseCode) == "" || strings.TrimSpace(body.ExpectedWord) == "" {
		http.Error(w, "course_code and expected_word required", http.StatusBadRequest)
		return
	}
	dryRun := body.DryRun == nil || *body.DryRun
	result, err := repository.NewWordRepository(r.db, r.logger).DeleteWordCardMatching(req.Context(), id, body.CourseCode, body.ExpectedWord, dryRun)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Word card not found", http.StatusNotFound)
		return
	case errors.Is(err, repository.ErrWordCardIdentityMismatch):
		http.Error(w, "Word card course or spelling changed", http.StatusConflict)
		return
	case err != nil:
		r.logger.Error("internal word card deletion failed", zap.Int64("word_card_id", id), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if result.Deleted {
		r.logger.Info("internal word card deleted", zap.Int64("word_card_id", id), zap.String("course_code", result.CourseCode), zap.String("word", result.Word))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
