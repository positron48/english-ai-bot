// Rebuilds daily_course_stats and mode_daily_stats from exercise_attempts history.
package main

import (
	"context"
	"fmt"
	"os"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

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

	repo := repository.NewLinglowDailyStatsRepository(db.GetConnection())
	summary, err := repo.Backfill(context.Background())
	if err != nil {
		log.Fatal("Linglow daily stats backfill failed", zap.Error(err))
	}
	fmt.Printf("backfill done: course_days=%d mode_days=%d\n", summary.CourseDays, summary.ModeDays)
}
