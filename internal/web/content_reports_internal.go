package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/grammartrainingpack"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func contentReportToMap(item *models.ContentReport, repo *repository.ContentReportRepository) map[string]interface{} {
	return map[string]interface{}{
		"id":                    item.ID,
		"user_id":               item.UserID,
		"source_type":           item.SourceType,
		"status":                item.Status,
		"report_category":       item.ReportCategory,
		"word":                  item.Word,
		"translation_direction": item.TranslationDirection,
		"word_card_id":          item.WordCardID,
		"training_card_id":      item.TrainingCardID,
		"user_card_id":          item.UserCardID,
		"word_category":         item.WordCategory,
		"grammar_chapter_id":    item.GrammarChapterID,
		"theory_block_id":       item.TheoryBlockID,
		"grammar_question_id":   item.GrammarQuestionID,
		"comment_text":          item.CommentText,
		"payload":               repo.ParsePayload(item.PayloadJSON),
		"resolved_at":           asReportTime(item.ResolvedAt),
		"resolved_by_user_id":   item.ResolvedByUserID,
		"created_at":            item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"updated_at":            item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func parseInternalReportsListQuery(req *http.Request) (repository.ListActiveReportsFilter, error) {
	limit := 200
	if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	var cursorID int64
	if raw := strings.TrimSpace(req.URL.Query().Get("cursor")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			return repository.ListActiveReportsFilter{}, fmt.Errorf("invalid cursor")
		}
		cursorID = v
	}
	status := strings.TrimSpace(req.URL.Query().Get("status"))
	if status == "" {
		status = string(models.ContentReportStatusActive)
	}
	return repository.ListActiveReportsFilter{
		Status:      status,
		SourceType:  strings.TrimSpace(req.URL.Query().Get("source_type")),
		Course:      strings.TrimSpace(req.URL.Query().Get("course")),
		ChapterID:   strings.TrimSpace(req.URL.Query().Get("chapter_id")),
		TheoryBlock: strings.TrimSpace(req.URL.Query().Get("theory_block_id")),
		Category:    strings.TrimSpace(req.URL.Query().Get("category")),
		CursorID:    cursorID,
		Limit:       limit,
	}, nil
}

func (r *Router) handleInternalContentReports(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	filter, err := parseInternalReportsListQuery(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	reports, err := repo.ListActiveReports(filter)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported course") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		r.logger.Error("failed to list internal content reports", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	items := make([]map[string]interface{}, 0, len(reports))
	var nextCursor int64
	for _, item := range reports {
		nextCursor = item.ID
		items = append(items, contentReportToMap(item, repo))
	}
	resp := map[string]interface{}{"reports": items}
	if len(reports) > 0 {
		resp["next_cursor"] = nextCursor
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *Router) handleInternalContentReportsSummary(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	rows, err := repo.SummaryActiveReports(strings.TrimSpace(req.URL.Query().Get("course")))
	if err != nil {
		if strings.Contains(err.Error(), "unsupported course") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		r.logger.Error("failed to summarize internal content reports", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"source_type":        row.SourceType,
			"report_category":    row.ReportCategory,
			"grammar_chapter_id": row.GrammarChapterID,
			"word_category":      row.WordCategory,
			"count":              row.Count,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"summary": out})
}

func (r *Router) handleInternalContentReportByID(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idRaw := strings.TrimPrefix(req.URL.Path, "/api/internal/content-reports/")
	idRaw = strings.Trim(idRaw, "/")
	if idRaw == "" || strings.Contains(idRaw, "/") {
		http.NotFound(w, req)
		return
	}
	reportID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		http.Error(w, "Invalid report id", http.StatusBadRequest)
		return
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	report, err := repo.GetByID(reportID)
	if err != nil {
		r.logger.Error("failed to get internal content report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if report == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	resp := contentReportToMap(report, repo)
	if report.SourceType == "reading_text" {
		textID := report.GrammarChapterID
		if payloadID, ok := repo.ParsePayload(report.PayloadJSON)["text_id"].(string); ok && strings.TrimSpace(payloadID) != "" {
			textID = strings.TrimSpace(payloadID)
		}
		doc, found, err := repository.NewReadingCatalogRepository(r.db).GetTextDocument(textID)
		if err != nil {
			r.logger.Error("failed to get reported reading text", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// The report payload is a historical snapshot. This is the current DB
		// document so triage can verify an import before resolving the report.
		resp["text_id"] = textID
		resp["reading_text_found"] = found
		resp["reading_text"] = doc
	}
	if report.TrainingCardID != nil {
		trainingRepo := repository.NewTrainingCardRepository(r.db, r.logger)
		if card, e := trainingRepo.GetTrainingCard(*report.TrainingCardID); e == nil && card != nil {
			resp["training_card"] = card
		}
	}
	if report.WordCardID != nil {
		wordRepo := repository.NewWordRepository(r.db, r.logger)
		if wc, e := wordRepo.GetWordCardByID(*report.WordCardID); e == nil && wc != nil {
			resp["word_card"] = wc
		}
	}
	if report.SourceType == "word_training" && report.Word != "" && r.pronunciationService != nil && r.pronunciationService.IsEnabled() {
		if status, e := r.pronunciationService.GetStatus(report.Word); e == nil {
			resp["tts_status"] = status
		}
	}
	if report.SourceType == "grammar_training" {
		if rel, packID := grammarTrainingPackRelPath(report.GrammarChapterID, report.TheoryBlockID); rel != "" {
			resp["training_pack_id"] = packID
			resp["training_pack_relpath"] = rel
			resp["courses_training_pack_hint"] = fmt.Sprintf("courses/%s-grammar/training_pack/chapters/%s", packID, rel)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func grammarTrainingPackRelPath(chapterID, theoryBlockID string) (relPath, packID string) {
	chapterID = strings.TrimSpace(chapterID)
	theoryBlockID = strings.TrimSpace(theoryBlockID)
	if chapterID == "" || theoryBlockID == "" {
		return "", ""
	}
	packID = "en"
	if strings.HasPrefix(strings.ToLower(chapterID), "es.") {
		packID = "es"
	}
	packFS, err := grammartrainingpack.PackFS(packID)
	if err != nil {
		return "", packID
	}
	raw, err := fs.ReadFile(packFS, "index.json")
	if err != nil {
		return "", packID
	}
	var idx struct {
		Blocks map[string]string `json:"blocks"`
	}
	if err := json.Unmarshal(raw, &idx); err != nil {
		return "", packID
	}
	key := chapterID + "::" + theoryBlockID
	return strings.TrimSpace(idx.Blocks[key]), packID
}

type resolveBulkContentReportsRequest struct {
	ReportIDs []int64 `json:"report_ids"`
	Reason    string  `json:"reason"`
}

func (r *Router) handleInternalContentReportsResolveBulk(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body resolveBulkContentReportsRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.ReportIDs) == 0 {
		http.Error(w, "report_ids required", http.StatusBadRequest)
		return
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	affected, err := repo.ResolveBulk(body.ReportIDs, nil)
	if err != nil {
		r.logger.Error("failed to resolve bulk content reports", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"affected":       affected,
		"resolved_by":    "service",
		"resolve_reason": strings.TrimSpace(body.Reason),
	})
}

func (r *Router) handleInternalContentReportsSubpath(w http.ResponseWriter, req *http.Request) {
	sub := strings.TrimPrefix(req.URL.Path, "/api/internal/content-reports/")
	sub = strings.Trim(sub, "/")
	if sub == "" {
		http.NotFound(w, req)
		return
	}
	switch sub {
	case "summary":
		r.handleInternalContentReportsSummary(w, req)
		return
	case "resolve-bulk":
		r.handleInternalContentReportsResolveBulk(w, req)
		return
	case "grammar":
		r.handleInternalGrammarContentReports(w, req)
		return
	case "grammar/resolve-bulk":
		r.handleInternalGrammarContentReportsResolveBulk(w, req)
		return
	}
	if _, err := strconv.ParseInt(sub, 10, 64); err == nil {
		r.handleInternalContentReportByID(w, req)
		return
	}
	http.NotFound(w, req)
}

func validateReportComment(sourceType, category, comment string) bool {
	if strings.TrimSpace(comment) != "" {
		return true
	}
	category = models.NormalizeReportCategory(sourceType, category)
	return category != "" && category != "other"
}
