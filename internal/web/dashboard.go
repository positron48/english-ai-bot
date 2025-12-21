package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"tgbot-skeleton/internal/utils"

	"go.uber.org/zap"
)

// handleDashboard shows the user dashboard
func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get due count
	query := `SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND (next_due_at IS NULL OR next_due_at <= ?)`
	var dueCount int
	err := r.db.QueryRow(query, userID, time.Now()).Scan(&dueCount)
	if err != nil {
		r.logger.Error("failed to get due count", zap.Error(err))
		dueCount = 0
	}

	r.renderTemplate(w, "dashboard.html", map[string]interface{}{
		"Title":    "Dashboard",
		"DueCount": dueCount,
	})
}

// handleChat handles AI chat requests
func (r *Router) handleChat(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	message := req.FormValue("message")
	if message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// Get AI service - need to properly type it
	// For now, access via interface assertion
	type AIService interface {
		GenerateResponse(ctx context.Context, text string) (string, error)
	}
	aiService, ok := r.aiService.(AIService)
	if !ok {
		r.logger.Error("AI service does not implement GenerateResponse")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate response
	ctx := req.Context()
	response, err := aiService.GenerateResponse(ctx, message)
	if err != nil {
		r.logger.Error("failed to generate AI response", zap.Error(err))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<div class="error">Sorry, an error occurred while processing your message. Please try again.</div>`)
		return
	}

	// Convert Markdown to HTML (simplified - just escape HTML for now)
	// In production, use a proper markdown library
	htmlResponse := utils.ConvertMarkdownToHTML(response)

	// Return response as HTML fragment
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="ai-response">
		<div class="message">%s</div>
	</div>`, htmlResponse)
}

