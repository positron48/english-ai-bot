package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		return 1
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		return 1
	}
	defer db.Close()

	conn := db.GetConnection()
	rows, err := conn.Query(`SELECT id, word, COALESCE(pos,'') FROM word_cards WHERE LOWER(COALESCE(pos,'')) IN ('verb','aux')`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query words: %v\n", err)
		return 1
	}
	defer rows.Close()
	repo := repository.NewVerbFormsRepository(conn, log)
	total := 0
	linked := 0
	for rows.Next() {
		var wordCardID int64
		var lemma string
		var pos string
		if err := rows.Scan(&wordCardID, &lemma, &pos); err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			return 1
		}
		total++
		ok, err := repo.LinkWordCardByLemma(wordCardID, strings.ToLower(strings.TrimSpace(lemma)), "es", "backfill")
		if err != nil && err != sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "link %q: %v\n", lemma, err)
			continue
		}
		if ok {
			linked++
		}
	}
	fmt.Printf("backfill done total_verbs=%d linked=%d unmatched=%d\n", total, linked, total-linked)
	return 0
}
