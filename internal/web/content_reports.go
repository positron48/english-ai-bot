package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	QuestionData map[string]interface{} `json:"question_data"`
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
	repo := repository.NewContentReportRepository(r.db, r.logger)
	payload := map[string]interface{}{
		"question_id":       body.QuestionID,
		"chapter_id":        body.ChapterID,
		"theory_block_id":   body.TheoryBlock,
		"question_snapshot": body.QuestionData,
	}
	id, err := repo.Create(repository.CreateContentReportInput{
		UserID:            userID,
		SourceType:        "grammar_training",
		GrammarChapterID:  strings.TrimSpace(body.ChapterID),
		TheoryBlockID:     strings.TrimSpace(body.TheoryBlock),
		GrammarQuestionID: strings.TrimSpace(body.QuestionID),
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
			out = append(out, map[string]interface{}{
				"id":                    item.ID,
				"user_id":               item.UserID,
				"source_type":           item.SourceType,
				"status":                item.Status,
				"word":                  item.Word,
				"translation_direction": item.TranslationDirection,
				"word_card_id":          item.WordCardID,
				"training_card_id":      item.TrainingCardID,
				"user_card_id":          item.UserCardID,
				"word_category":         item.WordCategory,
				"grammar_chapter_id":    item.GrammarChapterID,
				"theory_block_id":       item.TheoryBlockID,
				"grammar_question_id":   item.GrammarQuestionID,
				"payload":               repo.ParsePayload(item.PayloadJSON),
				"resolved_at":           asReportTime(item.ResolvedAt),
				"resolved_by_user_id":   item.ResolvedByUserID,
				"created_at":            item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
				"updated_at":            item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			})
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
	resp := map[string]interface{}{
		"id":                    report.ID,
		"user_id":               report.UserID,
		"source_type":           report.SourceType,
		"status":                report.Status,
		"word":                  report.Word,
		"translation_direction": report.TranslationDirection,
		"word_card_id":          report.WordCardID,
		"training_card_id":      report.TrainingCardID,
		"user_card_id":          report.UserCardID,
		"word_category":         report.WordCategory,
		"grammar_chapter_id":    report.GrammarChapterID,
		"theory_block_id":       report.TheoryBlockID,
		"grammar_question_id":   report.GrammarQuestionID,
		"payload":               repo.ParsePayload(report.PayloadJSON),
		"resolved_at":           asReportTime(report.ResolvedAt),
		"resolved_by_user_id":   report.ResolvedByUserID,
		"created_at":            report.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"updated_at":            report.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
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
