// Backfills reading and speaking legacy progress into Linglow exercise_attempts and learning_events.
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
	var opts repository.LinglowMediaProgressBackfillOptions
	flag.BoolVar(&opts.Commit, "commit", false, "write missing Linglow reading/speaking attempts/events")
	flag.StringVar(&opts.Source, "source", "all", "source to backfill: all, reading, speaking")
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

	repo := repository.NewLinglowMediaProgressBackfillRepository(db.GetConnection())
	summaries, err := repo.Backfill(context.Background(), cfg.Learning, opts)
	if err != nil {
		log.Fatal("Linglow media progress backfill failed", zap.Error(err))
	}
	mode := "dry-run"
	if opts.Commit {
		mode = "commit"
	}
	for _, s := range summaries {
		fmt.Printf("[%s] source=%s legacy_total=%d mirrored_total=%d missing=%d processed=%d inserted=%d unmapped_total=%d\n",
			mode, s.Source, s.LegacyTotal, s.MirroredTotal, s.Missing, s.Processed, s.Inserted, s.UnmappedTotal)
	}
}
