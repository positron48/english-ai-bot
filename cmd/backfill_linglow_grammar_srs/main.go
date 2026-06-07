// Backfills legacy grammar_theory_memory state into Linglow srs_items.
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
	var opts repository.LinglowGrammarSRSBackfillOptions
	var noPruneOrphans bool
	flag.BoolVar(&opts.Commit, "commit", false, "write missing Linglow grammar srs_items")
	flag.BoolVar(&opts.Resync, "resync", false, "re-upsert all mapped legacy grammar_theory_memory into srs_items; use with --commit")
	flag.BoolVar(&noPruneOrphans, "no-prune-orphans", false, "with --resync --commit, keep canonical due grammar srs_items that have no legacy memory row")
	flag.IntVar(&opts.Limit, "limit", 0, "max rows to process when --commit is set; 0 means no limit")
	flag.Parse()
	if opts.Resync && opts.Commit {
		opts.PruneOrphans = !noPruneOrphans
	}

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

	repo := repository.NewLinglowGrammarSRSBackfillRepository(db.GetConnection())
	summary, err := repo.Backfill(context.Background(), cfg.Learning, opts)
	if err != nil {
		log.Fatal("Linglow grammar SRS backfill failed", zap.Error(err))
	}

	mode := "dry-run"
	if opts.Commit {
		if opts.Resync {
			mode = "resync"
		} else {
			mode = "commit"
		}
	}
	fmt.Printf("[%s] course=%s legacy_total=%d mapped_total=%d srs_total=%d missing=%d processed=%d upserted=%d pruned_orphans=%d unmapped_total=%d\n",
		mode,
		summary.CourseCode,
		summary.LegacyTotal,
		summary.MappedTotal,
		summary.SRSTotal,
		summary.Missing,
		summary.Processed,
		summary.Upserted,
		summary.PrunedOrphans,
		summary.UnmappedTotal,
	)
}
