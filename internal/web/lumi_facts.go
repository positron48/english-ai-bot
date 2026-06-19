package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleLumiFact returns the fact of the day for the given context.
// @Summary      Факт дня от Lumi
// @Description  Возвращает факт дня для пары (курс, контекст, локаль) с ротацией «самый давно не показанный — следующий»
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Success      204  "Фактов нет"
// @Router       /api/linglow/lumi-fact [get]
func (r *Router) handleLumiFact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	courseCode := strings.TrimSpace(req.URL.Query().Get("course_code"))
	if courseCode == "" {
		courseCode = r.currentCourseCodeForUser(req.Context(), userID)
	}
	locale := strings.TrimSpace(req.URL.Query().Get("locale"))
	if locale == "" {
		locale = "ru"
	}
	factContext := req.URL.Query().Get("context")

	repo := repository.NewLumiFactRepository(r.db)
	fact, err := repo.GetDailyFact(req.Context(), courseCode, factContext, locale)
	if err != nil {
		r.logger.Error("failed to get lumi fact", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if fact == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, map[string]interface{}{
		"id":      fact.ID,
		"body":    fact.Body,
		"context": fact.Context,
	})
}

// handleAdminLumiFacts serves GET (list), POST (bulk add) and PUT (update one).
// @Summary      Управление Lumi-фактами
// @Tags         Admin
// @Router       /api/admin/lumi-facts [get]
func (r *Router) handleAdminLumiFacts(w http.ResponseWriter, req *http.Request) {
	repo := repository.NewLumiFactRepository(r.db)
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		offset, _ := strconv.Atoi(q.Get("offset"))
		facts, total, err := repo.List(req.Context(), repository.LumiFactFilter{
			CourseCode: strings.TrimSpace(q.Get("course_code")),
			Context:    strings.TrimSpace(q.Get("context")),
			Locale:     strings.TrimSpace(q.Get("locale")),
			Status:     strings.TrimSpace(q.Get("status")),
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			r.logger.Error("failed to list lumi facts", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"facts": facts, "total": total})
	case http.MethodPost:
		var body struct {
			CourseCode string `json:"course_code"`
			Context    string `json:"context"`
			Locale     string `json:"locale"`
			Text       string `json:"text"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		// One fact per paragraph: blocks separated by blank lines.
		blocks := splitFactBlocks(body.Text)
		if len(blocks) == 0 {
			http.Error(w, "text is empty", http.StatusBadRequest)
			return
		}
		inserted, err := repo.BulkInsert(req.Context(), strings.TrimSpace(body.CourseCode), body.Context, strings.TrimSpace(body.Locale), blocks, getUserIDFromContext(req.Context()))
		if err != nil {
			r.logger.Error("failed to bulk insert lumi facts", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"inserted": inserted})
	case http.MethodPut:
		var fact repository.LumiFact
		if err := json.NewDecoder(req.Body).Decode(&fact); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := repo.Update(req.Context(), fact); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func splitFactBlocks(text string) []string {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
