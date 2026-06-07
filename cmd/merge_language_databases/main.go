// Audits source language databases before a future Linglow unified DB merge.
//
// This command is intentionally dry-run only for now. It does not write to any
// database; it reports source/target counts and obvious identity conflicts that
// must be resolved before a real merge command is allowed to move data.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type dbSummary struct {
	Label                 string `json:"label"`
	URLProvided           bool   `json:"url_provided"`
	Users                 int64  `json:"users"`
	Courses               int64  `json:"courses"`
	UserCourses           int64  `json:"user_courses"`
	LearningItems         int64  `json:"learning_items"`
	ExerciseAttempts      int64  `json:"exercise_attempts"`
	LearningEvents        int64  `json:"learning_events"`
	SRSItems              int64  `json:"srs_items"`
	ReviewEvents          int64  `json:"review_events"`
	GrammarTestAttempts   int64  `json:"grammar_test_attempts"`
	GrammarSRSAttempts    int64  `json:"grammar_srs_attempts"`
	ReadingTextProgress   int64  `json:"reading_text_progress"`
	SpeakingAttempts      int64  `json:"speaking_attempts"`
	LegacyMappingTablesOK bool   `json:"legacy_mapping_tables_ok"`
}

type identityConflict struct {
	TelegramID    int64    `json:"telegram_id"`
	SourceLabels  []string `json:"source_labels"`
	SourceUserIDs []string `json:"source_user_ids"`
	TargetUserIDs []string `json:"target_user_ids,omitempty"`
}

type auditReport struct {
	Mode               string             `json:"mode"`
	GeneratedAt        time.Time          `json:"generated_at"`
	Sources            []dbSummary        `json:"sources"`
	Target             *dbSummary         `json:"target,omitempty"`
	TelegramConflicts  []identityConflict `json:"telegram_conflicts"`
	ReadyForWriteMerge bool               `json:"ready_for_write_merge"`
	Notes              []string           `json:"notes"`
}

type sourceDB struct {
	Label string
	URL   string
}

type openedSourceDB struct {
	Label string
	DB    *sql.DB
}

func main() {
	var englishURL, spanishURL, targetURL string
	var timeout time.Duration
	flag.StringVar(&englishURL, "english-db-url", env("ENGLISH_DATABASE_URL"), "English source DATABASE_URL; defaults to ENGLISH_DATABASE_URL")
	flag.StringVar(&spanishURL, "spanish-db-url", env("SPANISH_DATABASE_URL"), "Spanish source DATABASE_URL; defaults to SPANISH_DATABASE_URL")
	flag.StringVar(&targetURL, "target-db-url", env("TARGET_DATABASE_URL"), "Target unified DATABASE_URL; defaults to TARGET_DATABASE_URL")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "audit query timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sources := []sourceDB{
		{Label: "english", URL: strings.TrimSpace(englishURL)},
		{Label: "spanish", URL: strings.TrimSpace(spanishURL)},
	}

	report := auditReport{
		Mode:              "dry-run",
		GeneratedAt:       time.Now().UTC(),
		TelegramConflicts: []identityConflict{},
		Notes: []string{
			"audit-only foundation: no data is written",
			"real merge remains blocked until source identity conflicts are reviewed",
		},
	}

	var openSources []openedSourceDB
	for _, src := range sources {
		summary, db, err := summarizeDB(ctx, src.Label, src.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "merge audit failed for %s: %v\n", src.Label, err)
			os.Exit(1)
		}
		report.Sources = append(report.Sources, summary)
		if db != nil {
			defer db.Close()
			openSources = append(openSources, openedSourceDB{Label: src.Label, DB: db})
		}
	}

	targetSummary, targetDB, err := summarizeDB(ctx, "target", strings.TrimSpace(targetURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge audit failed for target: %v\n", err)
		os.Exit(1)
	}
	if targetDB != nil {
		defer targetDB.Close()
		report.Target = &targetSummary
	}

	conflicts, err := findTelegramConflicts(ctx, openSources, targetDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge audit conflict scan failed: %v\n", err)
		os.Exit(1)
	}
	report.TelegramConflicts = conflicts
	report.ReadyForWriteMerge = len(conflicts) == 0 && len(openSources) > 0 && targetDB != nil && targetSummary.LegacyMappingTablesOK
	if targetDB == nil {
		report.Notes = append(report.Notes, "target DB URL is not set; target readiness could not be checked")
	}
	if len(openSources) < 2 {
		report.Notes = append(report.Notes, "provide both English and Spanish source URLs before real merge planning")
	}
	if len(conflicts) > 0 {
		report.Notes = append(report.Notes, "telegram identity conflicts exist; real merge must map them explicitly")
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "encode report: %v\n", err)
		os.Exit(1)
	}
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func summarizeDB(ctx context.Context, label, url string) (dbSummary, *sql.DB, error) {
	s := dbSummary{Label: label, URLProvided: url != ""}
	if url == "" {
		return s, nil, nil
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		return s, nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return s, nil, err
	}
	tableCounts := map[string]*int64{
		"users":                 &s.Users,
		"courses":               &s.Courses,
		"user_courses":          &s.UserCourses,
		"learning_items":        &s.LearningItems,
		"exercise_attempts":     &s.ExerciseAttempts,
		"learning_events":       &s.LearningEvents,
		"srs_items":             &s.SRSItems,
		"review_events":         &s.ReviewEvents,
		"grammar_test_attempts": &s.GrammarTestAttempts,
		"grammar_attempts":      &s.GrammarSRSAttempts,
		"reading_text_progress": &s.ReadingTextProgress,
		"speaking_attempts":     &s.SpeakingAttempts,
	}
	for table, target := range tableCounts {
		count, err := countTable(ctx, db, table)
		if err != nil {
			db.Close()
			return s, nil, err
		}
		*target = count
	}
	ok, err := mappingTablesPresent(ctx, db)
	if err != nil {
		db.Close()
		return s, nil, err
	}
	s.LegacyMappingTablesOK = ok
	return s, db, nil
}

func countTable(ctx context.Context, db *sql.DB, table string) (int64, error) {
	exists, err := tableExists(ctx, db, table)
	if err != nil || !exists {
		return 0, err
	}
	var count int64
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, table).Scan(&exists)
	return exists, err
}

func mappingTablesPresent(ctx context.Context, db *sql.DB) (bool, error) {
	required := []string{
		"legacy_user_mappings",
		"legacy_course_mappings",
		"legacy_content_mappings",
		"legacy_attempt_mappings",
		"legacy_merge_conflicts",
	}
	for _, table := range required {
		exists, err := tableExists(ctx, db, table)
		if err != nil || !exists {
			return false, err
		}
	}
	return true, nil
}

func findTelegramConflicts(ctx context.Context, dbs []openedSourceDB, targetDB *sql.DB) ([]identityConflict, error) {
	seen := map[int64]*identityConflict{}
	for _, src := range dbs {
		if src.DB == nil {
			continue
		}
		rows, err := src.DB.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var userID, telegramID int64
			if err := rows.Scan(&userID, &telegramID); err != nil {
				rows.Close()
				return nil, err
			}
			entry := seen[telegramID]
			if entry == nil {
				entry = &identityConflict{TelegramID: telegramID}
				seen[telegramID] = entry
			}
			entry.SourceLabels = append(entry.SourceLabels, src.Label)
			entry.SourceUserIDs = append(entry.SourceUserIDs, fmt.Sprintf("%s:%d", src.Label, userID))
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	if targetDB != nil {
		rows, err := targetDB.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var userID, telegramID int64
			if err := rows.Scan(&userID, &telegramID); err != nil {
				rows.Close()
				return nil, err
			}
			if entry := seen[telegramID]; entry != nil {
				entry.TargetUserIDs = append(entry.TargetUserIDs, fmt.Sprintf("target:%d", userID))
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	var conflicts []identityConflict
	for _, entry := range seen {
		sourceSet := map[string]bool{}
		for _, label := range entry.SourceLabels {
			sourceSet[label] = true
		}
		if len(sourceSet) > 1 || len(entry.TargetUserIDs) > 0 {
			conflicts = append(conflicts, *entry)
		}
	}
	return conflicts, nil
}
