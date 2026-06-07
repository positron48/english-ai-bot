package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type createGrammarChapterReportRequest struct {
	ChapterID      string                 `json:"chapter_id"`
	TheoryBlockID  string                 `json:"theory_block_id"`
	ReportCategory string                 `json:"report_category"`
	Comment        string                 `json:"comment"`
	ContentSnapshot map[string]interface{} `json:"content_snapshot"`
	ClientReportID string                 `json:"client_report_id"`
}

type createGrammarTestReportRequest struct {
	QuestionID     string                 `json:"question_id"`
	ChapterID      string                 `json:"chapter_id"`
	Scope          string                 `json:"scope"`
	ScopeID        string                 `json:"scope_id"`
	ReportCategory string                 `json:"report_category"`
	Comment        string                 `json:"comment"`
	QuestionData   map[string]interface{} `json:"question_data"`
	ClientReportID string                 `json:"client_report_id"`
}

type offlineContentReportSyncItem struct {
	ClientReportID string                 `json:"client_report_id"`
	SourceType     string                 `json:"source_type"`
	ReportCategory string                 `json:"report_category"`
	Comment        string                 `json:"comment"`
	Word           string                 `json:"word"`
	Direction      string                 `json:"direction"`
	WordCardID     *int64                 `json:"word_card_id"`
	TrainingCardID *int64                 `json:"training_card_id"`
	UserCardID     *int64                 `json:"user_card_id"`
	WordCategory   string                 `json:"word_category"`
	GrammarChapterID string               `json:"grammar_chapter_id"`
	TheoryBlockID  string                 `json:"theory_block_id"`
	GrammarQuestionID string              `json:"grammar_question_id"`
	Payload        map[string]interface{} `json:"payload"`
}

func (r *Router) handleLearningGrammarChapterReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body createGrammarChapterReportRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.ChapterID) == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if !validateReportComment("grammar_chapter", body.ReportCategory, comment) {
		http.Error(w, "comment required", http.StatusBadRequest)
		return
	}
	if comment == "" {
		comment = models.NormalizeReportCategory("grammar_chapter", body.ReportCategory)
	}
	payload := map[string]interface{}{
		"chapter_id":       body.ChapterID,
		"theory_block_id":  body.TheoryBlockID,
		"content_snapshot": body.ContentSnapshot,
		"comment":          comment,
	}
	id, err := r.createContentReport(userID, repository.CreateContentReportInput{
		UserID:           userID,
		SourceType:       "grammar_chapter",
		ClientReportID:   body.ClientReportID,
		GrammarChapterID: strings.TrimSpace(body.ChapterID),
		TheoryBlockID:    strings.TrimSpace(body.TheoryBlockID),
		ReportCategory:   body.ReportCategory,
		CommentText:      comment,
		Payload:          payload,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "duplicate": true})
			return
		}
		r.logger.Error("failed to create grammar chapter report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "report_id": id})
}

func (r *Router) handleLearningGrammarTestReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body createGrammarTestReportRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.QuestionID) == "" {
		http.Error(w, "question_id required", http.StatusBadRequest)
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if !validateReportComment("grammar_test", body.ReportCategory, comment) {
		http.Error(w, "comment required", http.StatusBadRequest)
		return
	}
	if comment == "" {
		comment = models.NormalizeReportCategory("grammar_test", body.ReportCategory)
	}
	payload := map[string]interface{}{
		"question_id":     body.QuestionID,
		"chapter_id":      body.ChapterID,
		"scope":           body.Scope,
		"scope_id":        body.ScopeID,
		"question_snapshot": body.QuestionData,
		"comment":         comment,
	}
	id, err := r.createContentReport(userID, repository.CreateContentReportInput{
		UserID:            userID,
		SourceType:        "grammar_test",
		ClientReportID:    body.ClientReportID,
		GrammarChapterID:  strings.TrimSpace(body.ChapterID),
		GrammarQuestionID: strings.TrimSpace(body.QuestionID),
		ReportCategory:    body.ReportCategory,
		CommentText:       comment,
		Payload:           payload,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "duplicate": true})
			return
		}
		r.logger.Error("failed to create grammar test report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "report_id": id})
}

func (r *Router) handleContentReportsOfflineSync(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Reports []offlineContentReportSyncItem `json:"reports"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	results := make([]map[string]interface{}, 0, len(body.Reports))
	synced := 0
	for _, item := range body.Reports {
		result := map[string]interface{}{"client_report_id": item.ClientReportID}
		clientID := strings.TrimSpace(item.ClientReportID)
		sourceType := strings.TrimSpace(item.SourceType)
		if clientID == "" || sourceType == "" {
			result["synced"] = false
			result["error"] = "client_report_id and source_type required"
			results = append(results, result)
			continue
		}
		repo := repository.NewContentReportRepository(r.db, r.logger)
		exists, err := repo.HasClientReport(userID, clientID)
		if err != nil {
			result["synced"] = false
			result["error"] = "idempotency_check_failed"
			results = append(results, result)
			continue
		}
		if exists {
			result["synced"] = true
			result["duplicate"] = true
			results = append(results, result)
			synced++
			continue
		}
		comment := strings.TrimSpace(item.Comment)
		if !validateReportComment(sourceType, item.ReportCategory, comment) {
			result["synced"] = false
			result["error"] = "comment required"
			results = append(results, result)
			continue
		}
		if comment == "" {
			comment = models.NormalizeReportCategory(sourceType, item.ReportCategory)
		}
		payload := item.Payload
		if payload == nil {
			payload = map[string]interface{}{}
		}
		payload["comment"] = comment
		var userCardID *int64
		if item.UserCardID != nil && *item.UserCardID > 0 {
			userCardID = item.UserCardID
		}
		reportID, err := r.createContentReport(userID, repository.CreateContentReportInput{
			UserID:               userID,
			SourceType:           sourceType,
			ClientReportID:       clientID,
			Word:                 strings.TrimSpace(item.Word),
			TranslationDirection: strings.TrimSpace(item.Direction),
			WordCardID:           item.WordCardID,
			TrainingCardID:       item.TrainingCardID,
			UserCardID:           userCardID,
			WordCategory:         strings.TrimSpace(item.WordCategory),
			GrammarChapterID:     strings.TrimSpace(item.GrammarChapterID),
			TheoryBlockID:        strings.TrimSpace(item.TheoryBlockID),
			GrammarQuestionID:    strings.TrimSpace(item.GrammarQuestionID),
			ReportCategory:       item.ReportCategory,
			CommentText:          comment,
			Payload:              payload,
		})
		if err != nil {
			result["synced"] = false
			result["error"] = err.Error()
			results = append(results, result)
			continue
		}
		result["synced"] = true
		result["report_id"] = reportID
		results = append(results, result)
		synced++
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": results, "synced": synced})
}

func (r *Router) createContentReport(userID int64, input repository.CreateContentReportInput) (int64, error) {
	repo := repository.NewContentReportRepository(r.db, r.logger)
	clientID := strings.TrimSpace(input.ClientReportID)
	if clientID != "" {
		exists, err := repo.HasClientReport(userID, clientID)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, fmt.Errorf("duplicate client_report_id")
		}
	}
	input.UserID = userID
	return repo.Create(input)
}
