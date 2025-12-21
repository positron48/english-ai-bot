package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// RequireAdmin wraps a handler to require admin access
func (r *Router) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := getUserIDFromContext(req.Context())
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user's telegram_id
		userRepo := r.userRepo.(*repository.UserRepository)
		user, err := userRepo.GetUserByID(userID)
		if err != nil || user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Check if user is admin
		if user.TelegramID != int64(r.config.Admin.TelegramID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, req)
	}
}

// handleAdmin shows the admin panel
func (r *Router) handleAdmin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get circuit breaker status
	cbRepo := repository.NewCircuitBreakerRepository(r.db, r.logger)
	cbState, err := cbRepo.GetState()
	if err != nil {
		r.logger.Error("failed to get circuit breaker state", zap.Error(err))
		cbState = nil
	}

	r.renderTemplate(w, "admin.html", map[string]interface{}{
		"Title":      "Admin Panel",
		"CBState":    cbState,
		"AdminID":    r.config.Admin.TelegramID,
	})
}

// handleAdminCircuitReset resets the circuit breaker
func (r *Router) handleAdminCircuitReset(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.cbService.Reset(); err != nil {
		r.logger.Error("failed to reset circuit breaker", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to admin panel
	http.Redirect(w, req, "/app/admin?reset=success", http.StatusFound)
}

// handleAdminTraining handles training card management
func (r *Router) handleAdminTraining(w http.ResponseWriter, req *http.Request) {
	// Extract action and word from path: /app/admin/training/{word}/{action}
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/admin/training/"), "/")
	
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	wordEN := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)

	if req.Method == http.MethodGet && wordEN != "" && action == "" {
		// Get training data for word
		// Extract word from query parameter if path doesn't have it
		if wordEN == "" {
			wordEN = req.URL.Query().Get("word")
		}
		if wordEN == "" {
			http.Error(w, "word is required", http.StatusBadRequest)
			return
		}
		
		cards, err := trainingCardRepo.GetTrainingCardsByWordEN(wordEN)
		if err != nil {
			r.logger.Error("failed to get training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Format as JSON for display
		cardsJSON, _ := json.MarshalIndent(cards, "", "  ")

		r.renderTemplate(w, "admin_training_data.html", map[string]interface{}{
			"Title":     "Training Data",
			"WordEN":    wordEN,
			"Cards":     cards,
			"CardsJSON": string(cardsJSON),
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Delete training cards by word
		// Get word from form or path
		if wordEN == "" {
			wordEN = req.FormValue("word")
		}
		if wordEN == "" {
			http.Error(w, "word is required", http.StatusBadRequest)
			return
		}
		
		rowsAffected, err := trainingCardRepo.DeleteTrainingCardsByWordEN(wordEN)
		if err != nil {
			r.logger.Error("failed to delete training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, fmt.Sprintf("/app/admin?deleted=%s&count=%d", wordEN, rowsAffected), http.StatusFound)
		return
	}

	if req.Method == http.MethodPost && wordEN == "delete_all" {
		// Delete all training cards
		rowsAffected, err := trainingCardRepo.DeleteAllTrainingCards()
		if err != nil {
			r.logger.Error("failed to delete all training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, fmt.Sprintf("/app/admin?deleted_all=success&count=%d", rowsAffected), http.StatusFound)
		return
	}

	http.Error(w, "Invalid request", http.StatusBadRequest)
}

