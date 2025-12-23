package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// VocabWord represents a word with statistics
type VocabWord struct {
	WordEN      string
	TotalCards  int
	DueCount    int
	LastReview  *time.Time
}

// handleVocab shows the vocabulary list
func (r *Router) handleVocab(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all words for this user with statistics
	query := `SELECT 
		tc.word_en,
		COUNT(DISTINCT uc.id) as total_cards,
		SUM(CASE WHEN uc.next_due_at IS NULL OR uc.next_due_at <= ? THEN 1 ELSE 0 END) as due_count,
		MAX(uc.last_review_at) as last_review
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ?
	GROUP BY tc.word_en
	ORDER BY tc.word_en`

	rows, err := r.db.Query(query, time.Now(), userID)
	if err != nil {
		r.logger.Error("failed to get vocabulary", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var words []VocabWord
	for rows.Next() {
		var word VocabWord
		var totalCards, dueCount int
		var lastReview sql.NullString

		err := rows.Scan(&word.WordEN, &totalCards, &dueCount, &lastReview)
		if err != nil {
			r.logger.Error("failed to scan word", zap.Error(err))
			continue
		}

		word.TotalCards = totalCards
		word.DueCount = dueCount
		if lastReview.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lastReview.String)
			word.LastReview = &t
		}

		words = append(words, word)
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"words": words,
	})
}

// handleVocabDelete handles vocabulary deletion (confirm and delete)
func (r *Router) handleVocabDelete(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract word from URL path: /app/vocab/{word}/confirm_delete or /app/vocab/{word}/delete
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/vocab/"), "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	wordEN := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if wordEN == "" {
		// Invalid path, redirect to vocab list
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)

	if req.Method == http.MethodGet && action == "confirm_delete" {
		// Get word info for confirmation
		query := `SELECT COUNT(*) FROM user_cards uc
				  JOIN training_cards tc ON uc.training_card_id = tc.id
				  WHERE uc.user_id = ? AND tc.word_en = ?`
		var count int
		err := r.db.QueryRow(query, userID, wordEN).Scan(&count)
		if err != nil {
			r.logger.Error("failed to get word count", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If word not found or empty, return error
		if count == 0 || wordEN == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Word not found",
			})
			return
		}

		// Return word info for confirmation
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"word_en": wordEN,
			"count":   count,
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Perform deletion
		rowsAffected, err := userCardRepo.DeleteUserCardsByWordENForUser(userID, wordEN)
		if err != nil {
			r.logger.Error("failed to delete user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"word_en":       wordEN,
			"rows_affected": rowsAffected,
		})
		return
	}

	http.Error(w, "Invalid request", http.StatusBadRequest)
}

