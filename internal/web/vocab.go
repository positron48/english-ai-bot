package web

import (
	"database/sql"
	"fmt"
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

	r.renderTemplate(w, "vocab.html", map[string]interface{}{
		"Title": "Vocabulary",
		"Words": words,
	})
}

// handleVocabDelete handles vocabulary deletion (confirm and delete)
func (r *Router) handleVocabDelete(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("handleVocabDelete called", zap.String("path", req.URL.Path), zap.String("method", req.Method))
	
	// Validate that path starts with /app/vocab/
	path := req.URL.Path
	if !strings.HasPrefix(path, "/app/vocab/") {
		r.logger.Error("handleVocabDelete called with invalid path", zap.String("path", path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract word from URL path: /app/vocab/{word}/confirm_delete or /app/vocab/{word}/delete
	// Remove leading /app/vocab/ prefix
	remainingPath := strings.TrimPrefix(path, "/app/vocab/")
	
	// If path is empty or just a slash, redirect to vocab list
	if remainingPath == "" || remainingPath == "/" {
		r.logger.Warn("empty path in handleVocabDelete, redirecting to vocab list", zap.String("original_path", path))
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	parts := strings.Split(remainingPath, "/")
	if len(parts) < 1 || parts[0] == "" {
		// Invalid path, redirect to vocab list
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	wordEN := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// Validate wordEN is not empty
	if wordEN == "" {
		// Invalid path, redirect to vocab list
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)

	if req.Method == http.MethodGet && action == "confirm_delete" {
		// Show confirmation page
		// Validate wordEN is not empty before querying
		if wordEN == "" {
			r.logger.Warn("empty wordEN in confirm_delete request", zap.String("path", req.URL.Path))
			http.Redirect(w, req, "/app/vocab", http.StatusFound)
			return
		}

		// Get word info
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

		// If word not found, redirect to vocab list
		if count == 0 {
			r.logger.Warn("word not found for user", zap.String("word", wordEN), zap.Int64("userID", userID))
			http.Redirect(w, req, "/app/vocab", http.StatusFound)
			return
		}

		r.renderTemplate(w, "vocab_confirm_delete.html", map[string]interface{}{
			"Title": "Confirm Delete",
			"WordEN": wordEN,
			"Count": count,
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Validate wordEN is not empty before deletion
		if wordEN == "" {
			r.logger.Warn("empty wordEN in delete request", zap.String("path", req.URL.Path))
			http.Redirect(w, req, "/app/vocab", http.StatusFound)
			return
		}

		// Perform deletion
		rowsAffected, err := userCardRepo.DeleteUserCardsByWordENForUser(userID, wordEN)
		if err != nil {
			r.logger.Error("failed to delete user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Redirect to vocab list with success message
		http.Redirect(w, req, "/app/vocab?deleted="+wordEN+"&count="+fmt.Sprintf("%d", rowsAffected), http.StatusFound)
		return
	}

	// If we get here, the request doesn't match any expected pattern
	r.logger.Warn("unexpected request in handleVocabDelete", 
		zap.String("path", req.URL.Path), 
		zap.String("method", req.Method),
		zap.String("wordEN", wordEN),
		zap.String("action", action))
	http.Error(w, "Invalid request", http.StatusBadRequest)
}

