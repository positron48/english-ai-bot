package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type createWordReportRequest struct {
	UserCardID     int64                  `json:"user_card_id"`
	Word           string                 `json:"word"`
	Direction      string                 `json:"direction"`
	WordCardID     *int64                 `json:"word_card_id"`
	TrainingCardID *int64                 `json:"training_card_id"`
	WordCategory   string                 `json:"word_category"`
	ReportCategory string                 `json:"report_category"`
	Comment        string                 `json:"comment"`
	Extra          map[string]interface{} `json:"extra"`
}

func (r *Router) handleTrainingReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body createWordReportRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	repo := repository.NewContentReportRepository(r.db, r.logger)
	ctx := &repository.WordTrainingContext{
		Word:                 strings.TrimSpace(body.Word),
		TranslationDirection: strings.TrimSpace(body.Direction),
		WordCardID:           body.WordCardID,
		TrainingCardID:       body.TrainingCardID,
		WordCategory:         strings.TrimSpace(body.WordCategory),
	}
	if body.UserCardID > 0 {
		serverCtx, err := repo.GetWordTrainingContextByUserCard(body.UserCardID)
		if err != nil {
			r.logger.Error("failed to load word training context for report", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if serverCtx != nil {
			ctx = serverCtx
		}
	}
	if ctx.Word == "" {
		http.Error(w, "word or valid user_card_id required", http.StatusBadRequest)
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if !validateReportComment("word_training", body.ReportCategory, comment) {
		http.Error(w, "comment required", http.StatusBadRequest)
		return
	}
	if comment == "" {
		comment = models.NormalizeReportCategory("word_training", body.ReportCategory)
	}
	var reportUserCardID *int64
	if body.UserCardID > 0 {
		v := body.UserCardID
		reportUserCardID = &v
	} else if ctx.UserCardID != nil {
		reportUserCardID = ctx.UserCardID
	}
	payload := map[string]interface{}{
		"word":          ctx.Word,
		"direction":     ctx.TranslationDirection,
		"word_card_id":  ctx.WordCardID,
		"training_card": ctx.TrainingCardID,
		"comment":       comment,
	}
	for k, v := range body.Extra {
		payload[k] = v
	}
	id, err := repo.Create(repository.CreateContentReportInput{
		UserID:               userID,
		SourceType:           "word_training",
		Word:                 ctx.Word,
		TranslationDirection: ctx.TranslationDirection,
		WordCardID:           ctx.WordCardID,
		TrainingCardID:       ctx.TrainingCardID,
		UserCardID:           reportUserCardID,
		WordCategory:         ctx.WordCategory,
		ReportCategory:       body.ReportCategory,
		CommentText:          comment,
		Payload:              payload,
	})
	if err != nil {
		r.logger.Error("failed to create word training report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "report_id": id})
}

type createGrammarReportRequest struct {
	QuestionID   string                 `json:"question_id"`
	ChapterID    string                 `json:"chapter_id"`
	TheoryBlock  string                 `json:"theory_block_id"`
	ReportCategory string                 `json:"report_category"`
	Comment        string                 `json:"comment"`
	QuestionData   map[string]interface{} `json:"question_data"`
}

func (r *Router) handleLearningGrammarTrainingReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body createGrammarReportRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.QuestionID) == "" {
		http.Error(w, "question_id required", http.StatusBadRequest)
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if !validateReportComment("grammar_training", body.ReportCategory, comment) {
		http.Error(w, "comment required", http.StatusBadRequest)
		return
	}
	if comment == "" {
		comment = models.NormalizeReportCategory("grammar_training", body.ReportCategory)
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	payload := map[string]interface{}{
		"question_id":       body.QuestionID,
		"chapter_id":        body.ChapterID,
		"theory_block_id":   body.TheoryBlock,
		"question_snapshot": body.QuestionData,
		"comment":           comment,
	}
	id, err := repo.Create(repository.CreateContentReportInput{
		UserID:            userID,
		SourceType:        "grammar_training",
		GrammarChapterID:  strings.TrimSpace(body.ChapterID),
		TheoryBlockID:     strings.TrimSpace(body.TheoryBlock),
		GrammarQuestionID: strings.TrimSpace(body.QuestionID),
		ReportCategory:    body.ReportCategory,
		CommentText:       comment,
		Payload:           payload,
	})
	if err != nil {
		r.logger.Error("failed to create grammar training report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "report_id": id})
}

func (r *Router) handleAdminContentReports(w http.ResponseWriter, req *http.Request) {
	repo := repository.NewContentReportRepository(r.db, r.logger)
	switch req.Method {
	case http.MethodGet:
		status := strings.TrimSpace(req.URL.Query().Get("status"))
		reports, err := repo.List(status, 500)
		if err != nil {
			r.logger.Error("failed to list content reports", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		out := make([]map[string]interface{}, 0, len(reports))
		for _, item := range reports {
			out = append(out, contentReportToMap(item, repo))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"reports": out})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleAdminContentReportByID(w http.ResponseWriter, req *http.Request) {
	repo := repository.NewContentReportRepository(r.db, r.logger)
	idRaw := strings.TrimPrefix(req.URL.Path, "/api/admin/content-reports/")
	idRaw = strings.Trim(idRaw, "/")
	if idRaw == "" {
		http.NotFound(w, req)
		return
	}
	parts := strings.Split(idRaw, "/")
	reportID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid report id", http.StatusBadRequest)
		return
	}
	if len(parts) == 2 && parts[1] == "resolve" {
		if req.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		adminID := getUserIDFromContext(req.Context())
		if adminID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := repo.Resolve(reportID, adminID); err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			r.logger.Error("failed to resolve content report", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}
	if req.Method != http.MethodGet || len(parts) != 1 {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	report, err := repo.GetByID(reportID)
	if err != nil {
		r.logger.Error("failed to get content report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if report == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	resp := contentReportToMap(report, repo)
	if report.TrainingCardID != nil {
		trainingRepo := repository.NewTrainingCardRepository(r.db, r.logger)
		if card, e := trainingRepo.GetTrainingCard(*report.TrainingCardID); e == nil && card != nil {
			resp["training_card"] = card
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func asReportTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func (r *Router) authorizeInternalService(req *http.Request) bool {
	token := strings.TrimSpace(req.Header.Get("X-Service-Token"))
	if token == "" || len(r.internalServiceTokens) == 0 {
		return false
	}
	for _, expected := range r.internalServiceTokens {
		if expected == token {
			return true
		}
	}
	return false
}

func (r *Router) handleInternalGrammarContentReports(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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
			http.Error(w, "Invalid cursor", http.StatusBadRequest)
			return
		}
		cursorID = v
	}
	repo := repository.NewContentReportRepository(r.db, r.logger)
	reports, err := repo.ListActiveGrammarReports(repository.ListGrammarReportsFilter{
		Course:      strings.TrimSpace(req.URL.Query().Get("course")),
		ChapterID:   strings.TrimSpace(req.URL.Query().Get("chapter_id")),
		TheoryBlock: strings.TrimSpace(req.URL.Query().Get("theory_block_id")),
		CursorID:    cursorID,
		Limit:       limit,
	})
	if err != nil {
		r.logger.Error("failed to list internal grammar content reports", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	items := make([]map[string]interface{}, 0, len(reports))
	var nextCursor int64
	for _, item := range reports {
		nextCursor = item.ID
		m := contentReportToMap(item, repo)
		delete(m, "resolved_at")
		delete(m, "resolved_by_user_id")
		delete(m, "updated_at")
		delete(m, "user_id")
		items = append(items, m)
	}
	resp := map[string]interface{}{"reports": items}
	if len(reports) > 0 {
		resp["next_cursor"] = nextCursor
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type resolveBulkGrammarReportsRequest struct {
	ReportIDs []int64 `json:"report_ids"`
	Reason    string  `json:"reason"`
}

func (r *Router) handleInternalGrammarContentReportsResolveBulk(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body resolveBulkGrammarReportsRequest
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
		r.logger.Error("failed to resolve bulk grammar content reports", zap.Error(err))
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
