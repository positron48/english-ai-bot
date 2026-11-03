package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/repository"
)

func (r *Router) handleInternalWordCardRename(w http.ResponseWriter, req *http.Request) {
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(req.URL.Path, "/api/internal/word-cards/"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid word card id", http.StatusBadRequest)
		return
	}
	var body struct {
		CourseCode    string  `json:"course_code"`
		ExpectedWord  string  `json:"expected_word"`
		NewWord       string  `json:"new_word"`
		Transcription *string `json:"transcription"`
		DryRun        *bool   `json:"dry_run"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 4096))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	if err = dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Expected one JSON object", http.StatusBadRequest)
		return
	}
	if body.CourseCode == "" || body.ExpectedWord == "" || strings.TrimSpace(body.NewWord) != body.NewWord || body.NewWord == "" || body.NewWord == body.ExpectedWord || len(body.NewWord) > 255 || strings.ContainsAny(body.NewWord, "/\\") || strings.IndexFunc(body.NewWord, unicode.IsControl) >= 0 {
		http.Error(w, "course_code, expected_word and a different valid new_word required", http.StatusBadRequest)
		return
	}
	result, err := repository.NewWordRepository(r.db, r.logger).RenameWordCardMatching(req.Context(), id, body.CourseCode, body.ExpectedWord, body.NewWord, body.Transcription, body.DryRun == nil || *body.DryRun)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Word card not found", http.StatusNotFound)
		return
	case errors.Is(err, repository.ErrWordCardIdentityMismatch), errors.Is(err, repository.ErrWordCardRenameConflict):
		http.Error(w, "Word identity changed or rename conflicts with existing content", http.StatusConflict)
		return
	case err != nil:
		r.logger.Error("internal word rename failed", zap.Int64("word_card_id", id), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
