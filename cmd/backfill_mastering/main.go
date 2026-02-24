// One-time backfill of user_word_mastering from review_events.
// Run after deploying the user_word_mastering table and code.
// Usage: go run ./cmd/backfill_mastering or build and run the binary with DATABASE_URL set.
package main

import (
	"fmt"
	"os"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

const batchSize = 500

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	conn := db.GetConnection()
	masteringRepo := repository.NewUserWordMasteringRepository(conn, log)

	rows, err := conn.Query(`SELECT DISTINCT uc.user_id, tc.word_card_id
		FROM user_cards uc
		JOIN training_cards tc ON uc.training_card_id = tc.id
		ORDER BY uc.user_id, tc.word_card_id`)
	if err != nil {
		log.Fatal("Failed to query user-word pairs", zap.Error(err))
	}
	defer rows.Close()

	var pairs []repository.UserWordPair
	for rows.Next() {
		var p repository.UserWordPair
		if err := rows.Scan(&p.UserID, &p.WordCardID); err != nil {
			log.Fatal("Failed to scan pair", zap.Error(err))
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		log.Fatal("Iterating pairs", zap.Error(err))
	}

	log.Info("Backfilling mastering scores", zap.Int("pairs", len(pairs)))
	upserted := 0
	for i := 0; i < len(pairs); i += batchSize {
		end := i + batchSize
		if end > len(pairs) {
			end = len(pairs)
		}
		batch := pairs[i:end]
		statsMap, err := masteringRepo.GetWordMasteringStatsBatch(batch)
		if err != nil {
			log.Fatal("GetWordMasteringStatsBatch failed", zap.Error(err))
		}
		knownMap, err := masteringRepo.GetKnownForPairs(batch)
		if err != nil {
			log.Fatal("GetKnownForPairs failed", zap.Error(err))
		}
		entries := make([]struct {
			UserID     int64
			WordCardID int64
			Score      int
		}, 0, len(batch))
		for _, p := range batch {
			stats, hasStats := statsMap[p]
			known := knownMap[p]
			var total, correct, recentTotal, recentCorrect int64
			if hasStats {
				total = stats.Total
				correct = stats.Correct
				recentTotal = stats.RecentTotal
				recentCorrect = stats.RecentCorrect
			}
			score := service.ComputeMasteringScore(total, correct, recentTotal, recentCorrect, known)
			entries = append(entries, struct {
				UserID     int64
				WordCardID int64
				Score      int
			}{p.UserID, p.WordCardID, score})
		}
		if err := masteringRepo.UpsertBatch(entries); err != nil {
			log.Fatal("UpsertBatch failed", zap.Error(err))
		}
		upserted += len(entries)
		log.Info("Batch done", zap.Int("progress", upserted), zap.Int("total", len(pairs)))
	}
	log.Info("Backfill completed", zap.Int("upserted", upserted))
}
