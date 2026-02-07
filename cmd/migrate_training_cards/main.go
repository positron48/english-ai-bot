package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// TrainingCardPOSRequest represents a request to determine POS for a training card
type TrainingCardPOSRequest struct {
	WordEN    string `json:"word_en"`
	WordRU    string `json:"word_ru"`
	MeaningEN string `json:"meaning_en"`
	ExampleEN string `json:"example_en,omitempty"`
}

// TrainingCardPOSResponse represents LLM response for POS determination
type TrainingCardPOSResponse struct {
	Error       string `json:"error,omitempty"`
	POS         string `json:"pos"`         // Part of speech for this specific sense
	DisplayWord string `json:"display_word"` // Display form (e.g., "spy" or "to spy")
}

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	wordRepo := repository.NewWordRepository(db.GetConnection(), log)

	// Initialize AI service
	aiService := ai.NewService(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, log)

	ctx := context.Background()

	// Get all training cards that need migration (pos is NULL or empty)
	query := `SELECT id, word_card_id, word_en, word_ru, meaning_en, example_en, pos, display_word 
			  FROM training_cards 
			  WHERE pos IS NULL OR pos = '' OR display_word IS NULL OR display_word = ''
			  ORDER BY id`
	rows, err := db.GetConnection().Query(query)
	if err != nil {
		log.Fatal("Failed to query training cards", zap.Error(err))
	}
	defer rows.Close()

	var trainingCards []struct {
		ID          int64
		WordCardID  int64
		WordEN      string
		WordRU      string
		MeaningEN   string
		ExampleEN   string
		POS         sql.NullString
		DisplayWord sql.NullString
	}

	for rows.Next() {
		var tc struct {
			ID          int64
			WordCardID  int64
			WordEN      string
			WordRU      string
			MeaningEN   string
			ExampleEN   string
			POS         sql.NullString
			DisplayWord sql.NullString
		}
		if err := rows.Scan(&tc.ID, &tc.WordCardID, &tc.WordEN, &tc.WordRU, &tc.MeaningEN, &tc.ExampleEN, &tc.POS, &tc.DisplayWord); err != nil {
			log.Warn("Failed to scan training card", zap.Error(err))
			continue
		}
		trainingCards = append(trainingCards, tc)
	}

	log.Info("Found training cards to migrate", zap.Int("count", len(trainingCards)))

	// Process each training card
	processed := 0
	errors := 0
	skipped := 0

	for _, tc := range trainingCards {
		log.Info("Processing training card",
			zap.Int64("id", tc.ID),
			zap.String("word_en", tc.WordEN),
			zap.String("word_ru", tc.WordRU),
		)

		// Check if already migrated
		if tc.POS.Valid && tc.POS.String != "" && tc.DisplayWord.Valid && tc.DisplayWord.String != "" {
			log.Info("Training card already migrated, skipping",
				zap.Int64("id", tc.ID),
			)
			skipped++
			continue
		}

		// Get word_card to get lemma and base POS
		wordCard, err := wordRepo.GetWordCardByID(tc.WordCardID)
		if err != nil {
			log.Warn("Failed to get word card",
				zap.Int64("id", tc.ID),
				zap.Int64("word_card_id", tc.WordCardID),
				zap.Error(err),
			)
			errors++
			continue
		}
		if wordCard == nil {
			log.Warn("Word card not found",
				zap.Int64("id", tc.ID),
				zap.Int64("word_card_id", tc.WordCardID),
			)
			errors++
			continue
		}

		// Determine POS and display_word for this specific training card
		// Use LLM to determine POS based on meaning_en and word_ru
		exampleText := ""
		if tc.ExampleEN != "" {
			exampleText = fmt.Sprintf("\nExample: %s", tc.ExampleEN)
		}
		
		lemma := wordCard.Word
		if wordCard.DisplayEN != nil && *wordCard.DisplayEN != "" {
			// Extract lemma from display_en (remove "to " for verbs)
			lemma = strings.TrimPrefix(*wordCard.DisplayEN, "to ")
			lemma = strings.TrimSpace(lemma)
		}
		
		prompt := fmt.Sprintf(`Determine the part of speech (POS) for this SPECIFIC sense/meaning of a word.

Word: %s
Russian translation: %s
English meaning: %s%s

Return ONLY valid JSON (no markdown, no explanations, no code blocks) with the following structure:
{
  "error": "string (only if this is clearly not a valid sense)",
  "pos": "string (part of speech: noun, verb, adjective, adverb, etc. - must be ONE part of speech for THIS specific sense)",
  "display_word": "string (display form: for verbs use 'to %s', for others use '%s')"
}

CRITICAL: Each training card represents ONE specific sense/meaning. Determine the POS for THIS specific sense based on the meaning provided, not the word in general. The same word can have different POS for different senses (e.g., "bank" can be noun or verb).`, 
			tc.WordEN, tc.WordRU, tc.MeaningEN, exampleText, lemma, lemma)

		response, err := aiService.GenerateResponse(ctx, prompt)
		if err != nil {
			log.Warn("Failed to get AI response",
				zap.Int64("id", tc.ID),
				zap.Error(err),
			)
			errors++
			continue
		}

		// Clean up response - remove markdown code blocks if present
		response = strings.TrimSpace(response)
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)

		// Log raw response for debugging (first 300 chars)
		responsePreview := response
		if len(response) > 300 {
			responsePreview = response[:min(300, len(response))] + "..."
		}
		log.Info("LLM raw response",
			zap.Int64("id", tc.ID),
			zap.String("response", responsePreview),
		)

		// Parse JSON response
		var posResp TrainingCardPOSResponse
		if err := json.Unmarshal([]byte(response), &posResp); err != nil {
			log.Warn("Failed to parse AI response as JSON",
				zap.Int64("id", tc.ID),
				zap.String("response_preview", response[:min(200, len(response))]),
				zap.Error(err),
			)
			errors++
			continue
		}

		// Log parsed response
		log.Info("LLM parsed response",
			zap.Int64("id", tc.ID),
			zap.String("pos", posResp.POS),
			zap.String("display_word", posResp.DisplayWord),
			zap.String("error", posResp.Error),
		)

		// Check for error from LLM
		// LLM sometimes puts non-error strings in error field (like "load", "master", "none", "valid English word")
		errorMsg := strings.TrimSpace(posResp.Error)
		errorMsgLower := strings.ToLower(errorMsg)
		
		// List of known non-error strings that LLM sometimes puts in error field
		nonErrorStrings := []string{
			"null", "none", "false", "no",
			"load", "master", "slave", "tor", "corm",
			"valid english word", "valid English word",
		}
		
		// Keywords that indicate a real error (word doesn't exist, gibberish, etc.)
		errorKeywords := []string{
			"gibberish", "does not exist", "not exist", "non-standard", "not a valid",
			"not an english", "not english", "not recognized", "not a word",
			"doesn't exist", "not found", "invalid word", "not a real",
		}
		
		isNonErrorString := false
		for _, nonError := range nonErrorStrings {
			if errorMsgLower == strings.ToLower(nonError) {
				isNonErrorString = true
				break
			}
		}
		
		// Check if error message contains keywords indicating a real error
		isRealError := false
		if errorMsg != "" {
			for _, keyword := range errorKeywords {
				if strings.Contains(errorMsgLower, keyword) {
					isRealError = true
					break
				}
			}
		}
		
		// Treat as error if:
		// 1. Error field contains keywords indicating real error (word doesn't exist, etc.) OR
		// 2. Error field is not empty AND it's not a known non-error string
		if errorMsg != "" && (isRealError || !isNonErrorString) {
			log.Info("LLM rejected training card",
				zap.Int64("id", tc.ID),
				zap.String("error", errorMsg),
			)
			skipped++
			continue
		}

		// Validate response
		if posResp.POS == "" {
			// Try to use word_card POS as fallback
			if wordCard.POS != nil && *wordCard.POS != "" {
				posResp.POS = *wordCard.POS
				log.Info("Using word_card POS as fallback",
					zap.Int64("id", tc.ID),
					zap.String("pos", posResp.POS),
				)
			} else {
				log.Warn("LLM response missing POS",
					zap.Int64("id", tc.ID),
					zap.String("response", response),
				)
				errors++
				continue
			}
		}

		// Determine display_word if not provided
		if posResp.DisplayWord == "" {
			// Determine lemma
			lemma := wordCard.Word
			if wordCard.DisplayEN != nil && *wordCard.DisplayEN != "" {
				// Extract lemma from display_en (remove "to " for verbs)
				lemma = strings.TrimPrefix(*wordCard.DisplayEN, "to ")
				lemma = strings.TrimSpace(lemma)
			}
			if lemma == "" {
				lemma = tc.WordEN
			}
			
			// For verbs, prepend "to"
			if posResp.POS == "verb" {
				posResp.DisplayWord = "to " + lemma
			} else {
				posResp.DisplayWord = lemma
			}
		}

		// Update training card
		updateQuery := `UPDATE training_cards 
					   SET pos = ?, display_word = ?
					   WHERE id = ?`
		_, err = db.GetConnection().Exec(updateQuery, posResp.POS, posResp.DisplayWord, tc.ID)
		if err != nil {
			log.Warn("Failed to update training card",
				zap.Int64("id", tc.ID),
				zap.Error(err),
			)
			errors++
			continue
		}

		processed++
		log.Info("Migrated training card",
			zap.Int64("id", tc.ID),
			zap.String("word_en", tc.WordEN),
			zap.String("pos", posResp.POS),
			zap.String("display_word", posResp.DisplayWord),
		)

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	log.Info("Migration completed",
		zap.Int("processed", processed),
		zap.Int("skipped", skipped),
		zap.Int("errors", errors),
		zap.Int("total", len(trainingCards)),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
