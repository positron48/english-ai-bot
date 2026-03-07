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
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// migrationDelay is the pause between processing cards; tests may set to 0.
var migrationDelay = 500 * time.Millisecond

// dbConn is the DB interface used by run for query/exec (allows mocking in tests).
type dbConn interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// wordCardGetter returns a word card by ID (allows mocking in tests).
type wordCardGetter interface {
	GetWordCardByID(id int64) (*models.WordCard, error)
}

// aiResponder returns an LLM response for a prompt (allows mocking in tests).
type aiResponder interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

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

// configLoader loads config; replaced in tests to force init failure.
var configLoader = config.Load

// runFuncForTests if non-nil is used instead of run in runMigrationWithLoader (for coverage).
var runFuncForTests func(context.Context, dbConn, wordCardGetter, aiResponder, *zap.Logger) error

// runMigrationWithLoader runs the migration using the given config loader.
// Used by runMigration and by tests to inject config/logger failures.
func runMigrationWithLoader(ctx context.Context, load func() (*config.Config, error)) (exitMsg string, err error) {
	cfg, err := load()
	if err != nil {
		return "Failed to load config", err
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		return "Failed to initialize logger", err
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		return "Failed to initialize database", err
	}
	defer db.Close()
	wordRepo := repository.NewWordRepository(db.GetConnection(), log)
	aiService := ai.NewService(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, log)
	doRun := run
	if runFuncForTests != nil {
		doRun = runFuncForTests
	}
	if err := doRun(ctx, db.GetConnection(), wordRepo, aiService, log); err != nil {
		return "Failed to query training cards", err
	}
	return "", nil
}

// runMigration loads config, logger, DB, runs the migration and returns ("", nil) on success.
// On config or logger init failure returns (msg, err) so main can print and exit.
// On DB or query failure it calls log.Fatal and does not return.
func runMigration(ctx context.Context) (exitMsg string, err error) {
	return runMigrationWithLoader(ctx, configLoader)
}

// runCLI runs the migration and returns the process exit code (0 on success, 1 on error).
// Used by main and by tests to cover CLI exit paths.
func runCLI() int {
	msg, err := runMigration(context.Background())
	if err != nil {
		if msg != "" {
			fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		}
		return 1
	}
	return 0
}

func main() {
	os.Exit(runCLI())
}

// run performs the migration: queries training cards, calls AI, updates DB.
// conn must use the same placeholder style as the app (e.g. ? for compat driver).
func run(ctx context.Context, conn dbConn, wordRepo wordCardGetter, aiService aiResponder, log *zap.Logger) error {
	query := `SELECT id, word_card_id, word_en, word_ru, meaning_en, example_en, pos, display_word 
			  FROM training_cards 
			  WHERE pos IS NULL OR pos = '' OR display_word IS NULL OR display_word = ''
			  ORDER BY id`
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return err
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

	processed := 0
	errors := 0
	skipped := 0

	for _, tc := range trainingCards {
		log.Info("Processing training card",
			zap.Int64("id", tc.ID),
			zap.String("word_en", tc.WordEN),
			zap.String("word_ru", tc.WordRU),
		)

		if tc.POS.Valid && tc.POS.String != "" && tc.DisplayWord.Valid && tc.DisplayWord.String != "" {
			log.Info("Training card already migrated, skipping",
				zap.Int64("id", tc.ID),
			)
			skipped++
			continue
		}

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

		exampleText := ""
		if tc.ExampleEN != "" {
			exampleText = fmt.Sprintf("\nExample: %s", tc.ExampleEN)
		}

		lemma := wordCard.Word
		if wordCard.DisplayEN != nil && *wordCard.DisplayEN != "" {
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

		response = strings.TrimSpace(response)
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)

		responsePreview := response
		if len(response) > 300 {
			responsePreview = response[:min(300, len(response))] + "..."
		}
		log.Info("LLM raw response",
			zap.Int64("id", tc.ID),
			zap.String("response", responsePreview),
		)

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

		log.Info("LLM parsed response",
			zap.Int64("id", tc.ID),
			zap.String("pos", posResp.POS),
			zap.String("display_word", posResp.DisplayWord),
			zap.String("error", posResp.Error),
		)

		errorMsg := strings.TrimSpace(posResp.Error)
		errorMsgLower := strings.ToLower(errorMsg)

		nonErrorStrings := []string{
			"null", "none", "false", "no",
			"load", "master", "slave", "tor", "corm",
			"valid english word", "valid English word",
		}

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

		isRealError := false
		if errorMsg != "" {
			for _, keyword := range errorKeywords {
				if strings.Contains(errorMsgLower, keyword) {
					isRealError = true
					break
				}
			}
		}

		if errorMsg != "" && (isRealError || !isNonErrorString) {
			log.Info("LLM rejected training card",
				zap.Int64("id", tc.ID),
				zap.String("error", errorMsg),
			)
			skipped++
			continue
		}

		if posResp.POS == "" {
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

		if posResp.DisplayWord == "" {
			lemma := wordCard.Word
			if wordCard.DisplayEN != nil && *wordCard.DisplayEN != "" {
				lemma = strings.TrimPrefix(*wordCard.DisplayEN, "to ")
				lemma = strings.TrimSpace(lemma)
			}
			if lemma == "" {
				lemma = tc.WordEN
			}
			if posResp.POS == "verb" {
				posResp.DisplayWord = "to " + lemma
			} else {
				posResp.DisplayWord = lemma
			}
		}

		updateQuery := `UPDATE training_cards 
					   SET pos = ?, display_word = ?
					   WHERE id = ?`
		_, err = conn.ExecContext(ctx, updateQuery, posResp.POS, posResp.DisplayWord, tc.ID)
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

		time.Sleep(migrationDelay)
	}

	log.Info("Migration completed",
		zap.Int("processed", processed),
		zap.Int("skipped", skipped),
		zap.Int("errors", errors),
		zap.Int("total", len(trainingCards)),
	)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
