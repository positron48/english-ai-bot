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
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type dbSummary struct {
	Label                 string               `json:"label"`
	URLProvided           bool                 `json:"url_provided"`
	Users                 int64                `json:"users"`
	Courses               int64                `json:"courses"`
	UserCourses           int64                `json:"user_courses"`
	LearningItems         int64                `json:"learning_items"`
	ExerciseAttempts      int64                `json:"exercise_attempts"`
	LearningEvents        int64                `json:"learning_events"`
	SRSItems              int64                `json:"srs_items"`
	ReviewEvents          int64                `json:"review_events"`
	GrammarTestAttempts   int64                `json:"grammar_test_attempts"`
	GrammarSRSAttempts    int64                `json:"grammar_srs_attempts"`
	ReadingTextProgress   int64                `json:"reading_text_progress"`
	SpeakingAttempts      int64                `json:"speaking_attempts"`
	LegacyMappingTablesOK bool                 `json:"legacy_mapping_tables_ok"`
	LatestActivityAt      string               `json:"latest_activity_at,omitempty"`
	Readiness             readinessReport      `json:"readiness"`
	CourseBreakdown       []courseBreakdownRow `json:"course_breakdown"`
	AttemptSources        []attemptSourceRow   `json:"attempt_sources"`
}

type courseBreakdownRow struct {
	CourseCode       string `json:"course_code"`
	Users            int64  `json:"users"`
	UserCourses      int64  `json:"user_courses"`
	LearningItems    int64  `json:"learning_items"`
	ExerciseAttempts int64  `json:"exercise_attempts"`
	LearningEvents   int64  `json:"learning_events"`
	SRSItems         int64  `json:"srs_items"`
}

type attemptSourceRow struct {
	SourceTable string `json:"source_table"`
	Mode        string `json:"mode"`
	Count       int64  `json:"count"`
}

type readinessReport struct {
	UserCoursesMissing       int64    `json:"user_courses_missing"`
	UnmappedContentHints     int64    `json:"unmapped_content_hints"`
	LegacyAttemptsTotal      int64    `json:"legacy_attempts_total"`
	CanonicalAttemptsTotal   int64    `json:"canonical_attempts_total"`
	AttemptBackfillGap       int64    `json:"attempt_backfill_gap"`
	LegacySRSSnapshotsTotal  int64    `json:"legacy_srs_snapshots_total"`
	CanonicalSRSItemsTotal   int64    `json:"canonical_srs_items_total"`
	SRSBackfillGap           int64    `json:"srs_backfill_gap"`
	LegacyMediaProgressTotal int64    `json:"legacy_media_progress_total"`
	CanonicalMediaAttempts   int64    `json:"canonical_media_attempts"`
	MediaProgressBackfillGap int64    `json:"media_progress_backfill_gap"`
	ReadyForDryRunMerge      bool     `json:"ready_for_dry_run_merge"`
	BlockingReasons          []string `json:"blocking_reasons"`
}

type identityConflict struct {
	TelegramID    int64    `json:"telegram_id"`
	SourceLabels  []string `json:"source_labels"`
	SourceUserIDs []string `json:"source_user_ids"`
	TargetUserIDs []string `json:"target_user_ids,omitempty"`
}

type stableIdentityConflict struct {
	IdentityType  string   `json:"identity_type"`
	IdentityValue string   `json:"identity_value"`
	SourceLabels  []string `json:"source_labels"`
	SourceUserIDs []string `json:"source_user_ids"`
	TargetUserIDs []string `json:"target_user_ids,omitempty"`
}

type auditReport struct {
	Mode               string                   `json:"mode"`
	GeneratedAt        time.Time                `json:"generated_at"`
	Sources            []dbSummary              `json:"sources"`
	Target             *dbSummary               `json:"target,omitempty"`
	TelegramConflicts  []identityConflict       `json:"telegram_conflicts"`
	IdentityConflicts  []stableIdentityConflict `json:"identity_conflicts"`
	ReadyForWriteMerge bool                     `json:"ready_for_write_merge"`
	Notes              []string                 `json:"notes"`
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
		IdentityConflicts: []stableIdentityConflict{},
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
	if conflicts == nil {
		conflicts = []identityConflict{}
	}
	report.TelegramConflicts = conflicts
	identityConflicts, err := findStableIdentityConflicts(ctx, openSources, targetDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge audit stable identity scan failed: %v\n", err)
		os.Exit(1)
	}
	if identityConflicts == nil {
		identityConflicts = []stableIdentityConflict{}
	}
	report.IdentityConflicts = identityConflicts
	report.ReadyForWriteMerge = len(conflicts) == 0 && len(identityConflicts) == 0 && len(openSources) > 0 && targetDB != nil && targetSummary.LegacyMappingTablesOK
	if targetDB == nil {
		report.Notes = append(report.Notes, "target DB URL is not set; target readiness could not be checked")
	}
	if len(openSources) < 2 {
		report.Notes = append(report.Notes, "provide both English and Spanish source URLs before real merge planning")
	}
	if len(conflicts) > 0 {
		report.Notes = append(report.Notes, "telegram identity conflicts exist; real merge must map them explicitly")
	}
	if len(identityConflicts) > 0 {
		report.Notes = append(report.Notes, "stable identity conflicts exist; real merge must map them explicitly")
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
	s := dbSummary{
		Label:       label,
		URLProvided: url != "",
		Readiness: readinessReport{
			BlockingReasons: []string{},
		},
		CourseBreakdown: []courseBreakdownRow{},
		AttemptSources:  []attemptSourceRow{},
	}
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
	latest, err := latestActivity(ctx, db)
	if err != nil {
		db.Close()
		return s, nil, err
	}
	s.LatestActivityAt = latest
	readiness, err := computeReadiness(ctx, db, s)
	if err != nil {
		db.Close()
		return s, nil, err
	}
	s.Readiness = readiness
	courses, err := courseBreakdown(ctx, db)
	if err != nil {
		db.Close()
		return s, nil, err
	}
	s.CourseBreakdown = courses
	sources, err := attemptSources(ctx, db)
	if err != nil {
		db.Close()
		return s, nil, err
	}
	s.AttemptSources = sources
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

func latestActivity(ctx context.Context, db *sql.DB) (string, error) {
	tables := map[string]string{
		"exercise_attempts":     "answered_at",
		"learning_events":       "event_time",
		"review_events":         "answered_at",
		"grammar_test_attempts": "finished_at",
		"grammar_attempts":      "answered_at",
		"reading_text_progress": "read_at",
		"speaking_attempts":     "created_at",
	}
	var latest *time.Time
	for table, column := range tables {
		exists, err := columnExists(ctx, db, table, column)
		if err != nil {
			return "", err
		}
		if !exists {
			continue
		}
		var value sql.NullTime
		if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT MAX(%s) FROM %s", column, table)).Scan(&value); err != nil {
			return "", err
		}
		if value.Valid && (latest == nil || value.Time.After(*latest)) {
			t := value.Time.UTC()
			latest = &t
		}
	}
	if latest == nil {
		return "", nil
	}
	return latest.Format(time.RFC3339), nil
}

func columnExists(ctx context.Context, db *sql.DB, table, column string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)
	`, table, column).Scan(&exists)
	return exists, err
}

func computeReadiness(ctx context.Context, db *sql.DB, s dbSummary) (readinessReport, error) {
	r := readinessReport{
		LegacyAttemptsTotal:      s.ReviewEvents + s.GrammarTestAttempts + s.GrammarSRSAttempts,
		CanonicalAttemptsTotal:   s.ExerciseAttempts,
		LegacySRSSnapshotsTotal:  0,
		CanonicalSRSItemsTotal:   s.SRSItems,
		LegacyMediaProgressTotal: s.ReadingTextProgress + s.SpeakingAttempts,
		BlockingReasons:          []string{},
	}
	if s.Users > s.UserCourses {
		r.UserCoursesMissing = s.Users - s.UserCourses
	}
	if s.ExerciseAttempts < r.LegacyAttemptsTotal {
		r.AttemptBackfillGap = r.LegacyAttemptsTotal - s.ExerciseAttempts
	}
	wordSRS, err := countTable(ctx, db, "user_cards")
	if err != nil {
		return r, err
	}
	grammarSRS, err := countTable(ctx, db, "grammar_theory_memory")
	if err != nil {
		return r, err
	}
	r.LegacySRSSnapshotsTotal = wordSRS + grammarSRS
	if r.CanonicalSRSItemsTotal < r.LegacySRSSnapshotsTotal {
		r.SRSBackfillGap = r.LegacySRSSnapshotsTotal - r.CanonicalSRSItemsTotal
	}
	mediaCanonical, err := countCanonicalMediaAttempts(ctx, db)
	if err != nil {
		return r, err
	}
	r.CanonicalMediaAttempts = mediaCanonical
	if r.CanonicalMediaAttempts < r.LegacyMediaProgressTotal {
		r.MediaProgressBackfillGap = r.LegacyMediaProgressTotal - r.CanonicalMediaAttempts
	}
	if s.LearningItems == 0 && (s.Courses > 0 || s.UserCourses > 0) {
		r.UnmappedContentHints = 1
	}

	if r.UserCoursesMissing > 0 {
		r.BlockingReasons = append(r.BlockingReasons, "user_courses backfill is incomplete")
	}
	if s.LearningItems == 0 {
		r.BlockingReasons = append(r.BlockingReasons, "learning_items mapping is empty")
	}
	if r.AttemptBackfillGap > 0 {
		r.BlockingReasons = append(r.BlockingReasons, "exercise_attempts backfill is behind legacy attempts")
	}
	if r.SRSBackfillGap > 0 {
		r.BlockingReasons = append(r.BlockingReasons, "srs_items snapshot backfill is behind legacy SRS state")
	}
	if r.MediaProgressBackfillGap > 0 {
		r.BlockingReasons = append(r.BlockingReasons, "reading/speaking progress backfill is behind legacy progress")
	}
	r.ReadyForDryRunMerge = len(r.BlockingReasons) == 0 && s.Users > 0 && s.UserCourses > 0 && s.LearningItems > 0
	return r, nil
}

func countCanonicalMediaAttempts(ctx context.Context, db *sql.DB) (int64, error) {
	exists, err := tableExists(ctx, db, "exercise_attempts")
	if err != nil || !exists {
		return 0, err
	}
	var count int64
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM exercise_attempts
		WHERE source_table IN ('reading_text_progress', 'speaking_attempts')
		   OR mode IN ('reading', 'speaking')
	`).Scan(&count)
	return count, err
}

func courseBreakdown(ctx context.Context, db *sql.DB) ([]courseBreakdownRow, error) {
	required := []string{"courses", "user_courses", "learning_items", "exercise_attempts", "learning_events", "srs_items"}
	for _, table := range required {
		exists, err := tableExists(ctx, db, table)
		if err != nil || !exists {
			return []courseBreakdownRow{}, err
		}
	}
	rows, err := db.QueryContext(ctx, `
		SELECT
			c.code,
			COUNT(DISTINCT uc.user_id) AS users,
			COUNT(DISTINCT uc.id) AS user_courses,
			COUNT(DISTINCT li.id) AS learning_items,
			COUNT(DISTINCT ea.id) AS exercise_attempts,
			COUNT(DISTINCT le.id) AS learning_events,
			COUNT(DISTINCT si.id) AS srs_items
		FROM courses c
		LEFT JOIN user_courses uc ON uc.course_id = c.id
		LEFT JOIN learning_items li ON li.course_id = c.id
		LEFT JOIN exercise_attempts ea ON ea.user_course_id = uc.id
		LEFT JOIN learning_events le ON le.user_course_id = uc.id
		LEFT JOIN srs_items si ON si.user_course_id = uc.id
		GROUP BY c.code
		ORDER BY c.code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []courseBreakdownRow
	for rows.Next() {
		var row courseBreakdownRow
		if err := rows.Scan(&row.CourseCode, &row.Users, &row.UserCourses, &row.LearningItems, &row.ExerciseAttempts, &row.LearningEvents, &row.SRSItems); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []courseBreakdownRow{}
	}
	return out, nil
}

func attemptSources(ctx context.Context, db *sql.DB) ([]attemptSourceRow, error) {
	exists, err := tableExists(ctx, db, "exercise_attempts")
	if err != nil || !exists {
		return []attemptSourceRow{}, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(source_table, ''), mode, COUNT(*)
		FROM exercise_attempts
		GROUP BY COALESCE(source_table, ''), mode
		ORDER BY COALESCE(source_table, ''), mode
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []attemptSourceRow
	for rows.Next() {
		var row attemptSourceRow
		if err := rows.Scan(&row.SourceTable, &row.Mode, &row.Count); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []attemptSourceRow{}
	}
	return out, nil
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
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].TelegramID < conflicts[j].TelegramID
	})
	return conflicts, nil
}

func findStableIdentityConflicts(ctx context.Context, dbs []openedSourceDB, targetDB *sql.DB) ([]stableIdentityConflict, error) {
	seen := map[string]*stableIdentityConflict{}
	for _, src := range dbs {
		if src.DB == nil {
			continue
		}
		if err := collectStableIdentities(ctx, seen, src.Label, src.DB, false); err != nil {
			return nil, err
		}
	}
	if targetDB != nil {
		if err := collectStableIdentities(ctx, seen, "target", targetDB, true); err != nil {
			return nil, err
		}
	}

	var conflicts []stableIdentityConflict
	for _, entry := range seen {
		sourceSet := map[string]bool{}
		for _, label := range entry.SourceLabels {
			sourceSet[label] = true
		}
		if len(sourceSet) > 1 || len(entry.TargetUserIDs) > 0 {
			conflicts = append(conflicts, *entry)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].IdentityType != conflicts[j].IdentityType {
			return conflicts[i].IdentityType < conflicts[j].IdentityType
		}
		return conflicts[i].IdentityValue < conflicts[j].IdentityValue
	})
	return conflicts, nil
}

func collectStableIdentities(ctx context.Context, seen map[string]*stableIdentityConflict, label string, db *sql.DB, target bool) error {
	exists, err := tableExists(ctx, db, "users")
	if err != nil || !exists {
		return err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, telegram_id, COALESCE(NULLIF(TRIM(telegram_username), ''), '')
		FROM users
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, telegramID int64
		var username string
		if err := rows.Scan(&userID, &telegramID, &username); err != nil {
			return err
		}
		addStableIdentity(seen, "telegram_id", fmt.Sprintf("%d", telegramID), label, userID, target)
		if username != "" {
			addStableIdentity(seen, "telegram_username", strings.ToLower(username), label, userID, target)
		}
	}
	return rows.Err()
}

func addStableIdentity(seen map[string]*stableIdentityConflict, identityType, identityValue, label string, userID int64, target bool) {
	key := identityType + "\x00" + identityValue
	entry := seen[key]
	if entry == nil {
		entry = &stableIdentityConflict{
			IdentityType:  identityType,
			IdentityValue: identityValue,
			SourceLabels:  []string{},
			SourceUserIDs: []string{},
			TargetUserIDs: []string{},
		}
		seen[key] = entry
	}
	if target {
		entry.TargetUserIDs = append(entry.TargetUserIDs, fmt.Sprintf("%s:%d", label, userID))
		return
	}
	entry.SourceLabels = append(entry.SourceLabels, label)
	entry.SourceUserIDs = append(entry.SourceUserIDs, fmt.Sprintf("%s:%d", label, userID))
}
