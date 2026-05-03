package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/verbtraining"

	"go.uber.org/zap"
)

func (r *Router) spanishVerbFormsAdminAvailable() bool {
	if r == nil || r.config == nil {
		return false
	}
	if !r.config.Training.SpanishVerbFormsEnabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(r.config.Learning.TargetLang), "es")
}

// handleAdminVerbTrainingLemmas GET /api/admin/verb-training/lemmas?q=&limit=&cursor=
func (r *Router) handleAdminVerbTrainingLemmas(w http.ResponseWriter, req *http.Request) {
	ctx := r.loadUserPermissionsIntoContext(req.Context())
	req = req.WithContext(ctx)
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.HasPermission(ctx, PermissionWordsReadAll) {
		http.Error(w, "Forbidden: read permission required", http.StatusForbidden)
		return
	}
	if !r.spanishVerbFormsAdminAvailable() {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	q := strings.TrimSpace(req.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	cursor, _ := strconv.ParseInt(req.URL.Query().Get("cursor"), 10, 64)

	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	rows, err := repo.ListAdminVerbTrainingLemmas(q, limit, cursor)
	if err != nil {
		r.logger.Error("admin verb training lemmas", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	full := verbtraining.FullCoverageClozeCardCountV1()
	type lemmaOut struct {
		WordCardID   int64  `json:"word_card_id"`
		Lemma        string `json:"lemma"`
		ClozeCount   int64  `json:"cloze_count"`
		RuGloss      string `json:"ru_gloss,omitempty"`
		FullCoverage bool   `json:"full_coverage"`
	}
	out := make([]lemmaOut, 0, len(rows))
	var nextCursor int64
	for _, row := range rows {
		out = append(out, lemmaOut{
			WordCardID:   row.WordCardID,
			Lemma:        row.Lemma,
			ClozeCount:   row.ClozeCount,
			RuGloss:      row.RuGloss,
			FullCoverage: row.ClozeCount >= int64(full),
		})
		nextCursor = row.WordCardID
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"lemmas":                   out,
		"next_cursor":              nextCursor,
		"expected_cloze_per_lemma": full,
	})
}

// handleAdminVerbTrainingCards GET /api/admin/verb-training/cards?word_card_id=
func (r *Router) handleAdminVerbTrainingCards(w http.ResponseWriter, req *http.Request) {
	ctx := r.loadUserPermissionsIntoContext(req.Context())
	req = req.WithContext(ctx)
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.HasPermission(ctx, PermissionWordsReadAll) {
		http.Error(w, "Forbidden: read permission required", http.StatusForbidden)
		return
	}
	if !r.spanishVerbFormsAdminAvailable() {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	wordCardID, err := strconv.ParseInt(strings.TrimSpace(req.URL.Query().Get("word_card_id")), 10, 64)
	if err != nil || wordCardID <= 0 {
		http.Error(w, "missing or invalid word_card_id", http.StatusBadRequest)
		return
	}

	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	lemma, ok, err := repo.AdminVerbLemmaLookup(wordCardID)
	if err != nil {
		r.logger.Error("admin verb lemma lookup", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "word not linked to a Spanish verb lemma", http.StatusNotFound)
		return
	}

	cards, err := repo.ListAdminVerbTrainingCardsByWordCard(wordCardID)
	if err != nil {
		r.logger.Error("admin verb training cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"word_card_id": wordCardID,
		"lemma":        lemma,
		"cards":        cards,
	})
}
