// Backfills srs_item_id links for Linglow exercise_attempts.
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
	var opts repository.LinglowAttemptSRSLinkBackfillOptions
	flag.BoolVar(&opts.Commit, "commit", false, "write missing exercise_attempts.srs_item_id links")
	flag.IntVar(&opts.Limit, "limit", 5000, "batch size per UPDATE when --commit is set; 0 updates all linkable rows in one pass")
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

	repo := repository.NewLinglowAttemptSRSLinkBackfillRepository(db.GetConnection())
	summaries, err := repo.Backfill(context.Background(), cfg.Learning, opts)
	if err != nil {
		log.Fatal("Linglow attempt SRS link backfill failed", zap.Error(err))
	}
	mode := "dry-run"
	if opts.Commit {
		mode = "commit"
	}
	for _, s := range summaries {
		fmt.Printf("[%s] source=%s missing=%d updated=%d unmapped_total=%d\n", mode, s.Source, s.Missing, s.Updated, s.UnmappedTotal)
	}
}
