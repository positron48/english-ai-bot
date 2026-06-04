// Backfills historical legacy attempts into Linglow exercise_attempts and learning_events.
//
// Dry-run audit:
//
//	go run ./cmd/backfill_linglow_events
//
// Commit all currently supported sources:
//
//	go run ./cmd/backfill_linglow_events --commit
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func main() {
	var opts repository.LinglowEventBackfillOptions
	flag.BoolVar(&opts.Commit, "commit", false, "write missing Linglow exercise attempts/events")
	flag.StringVar(&opts.Source, "source", repository.LinglowBackfillSourceAll, "source to backfill: all, grammar-tests, grammar-training, word-reviews")
	flag.IntVar(&opts.Limit, "limit", 0, "max rows to process per source when --commit is set; 0 means no limit")
	flag.Parse()

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

	repo := repository.NewLinglowEventBackfillRepository(db.GetConnection())
	summaries, err := repo.Backfill(context.Background(), cfg.Learning, opts)
	if err != nil {
		log.Fatal("Linglow event backfill failed", zap.Error(err))
	}

	mode := "dry-run"
	if opts.Commit {
		mode = "commit"
	}
	for _, s := range summaries {
		fmt.Printf("[%s] source=%s legacy_total=%d mirrored_total=%d missing=%d processed=%d inserted=%d failed=%d\n",
			mode, s.Source, s.LegacyTotal, s.MirroredTotal, s.Missing, s.Processed, s.Inserted, s.Failed)
	}
}
