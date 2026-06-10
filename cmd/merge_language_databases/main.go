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
	AttemptSources []attemptSourceRow `json:"attempt_sources"`
}

type telegramMultiCourseUser struct {
	TelegramID    int64    `json:"telegram_id"`
	SourceLabels  []string `json:"source_labels"`
	SourceUserIDs []string `json:"source_user_ids"`
}

type writeSummary struct {
	Phase            string `json:"phase"`
	UsersScanned     int64  `json:"users_scanned"`
	MappingsCreated  int64  `json:"mappings_created"`
	MappingsExisting int64  `json:"mappings_existing"`
	UsersInserted    int64  `json:"users_inserted"`
	UsersReused      int64  `json:"users_reused"`
	UserCoursesAdded int64  `json:"user_courses_added"`
	ItemsScanned      int64 `json:"items_scanned,omitempty"`
	ItemsInserted     int64 `json:"items_inserted,omitempty"`
	AttemptsScanned   int64 `json:"attempts_scanned,omitempty"`
	AttemptsInserted  int64 `json:"attempts_inserted,omitempty"`
	EventsInserted    int64 `json:"events_inserted,omitempty"`
	SRSScanned        int64 `json:"srs_scanned,omitempty"`
	SRSInserted       int64 `json:"srs_inserted,omitempty"`
	SRSLinksUpdated   int64 `json:"srs_links_updated,omitempty"`
	ConflictsLogged   int64 `json:"conflicts_logged"`
	Skipped          int64  `json:"skipped"`
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
	SRSWordMissing           int64    `json:"srs_word_missing"`
	SRSGrammarMissing        int64    `json:"srs_grammar_missing"`
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
	Mode                string                    `json:"mode"`
	GeneratedAt         time.Time                 `json:"generated_at"`
	Sources             []dbSummary               `json:"sources"`
	Target              *dbSummary                `json:"target,omitempty"`
	TelegramMultiCourse []telegramMultiCourseUser `json:"telegram_multi_course_users"`
	TelegramConflicts   []identityConflict        `json:"telegram_conflicts"`
	IdentityConflicts   []stableIdentityConflict  `json:"identity_conflicts"`
	ReadyForSourceMerge bool   `json:"ready_for_source_merge"`
	ReadyForWriteMerge  bool   `json:"ready_for_write_merge"`
	WritePhase          string `json:"write_phase,omitempty"`
	WriteSummary        *writeSummary             `json:"write_summary,omitempty"`
	Notes               []string                  `json:"notes"`
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
	var commit bool
	var phase string
	flag.StringVar(&englishURL, "english-db-url", env("ENGLISH_DATABASE_URL"), "English source DATABASE_URL; defaults to ENGLISH_DATABASE_URL")
	flag.StringVar(&spanishURL, "spanish-db-url", env("SPANISH_DATABASE_URL"), "Spanish source DATABASE_URL; defaults to SPANISH_DATABASE_URL")
	flag.StringVar(&targetURL, "target-db-url", env("TARGET_DATABASE_URL"), "Target unified DATABASE_URL; defaults to TARGET_DATABASE_URL")
	flag.DurationVar(&timeout, "timeout", 5*time.Minute, "overall audit timeout (counts and telegram identity scans)")
	flag.BoolVar(&commit, "commit", false, "write merge data to target DB; requires --phase")
	flag.StringVar(&phase, "phase", "", "write phase: users|user-courses|course-mappings|content|attempts|srs|legacy-words|reset-word-items")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sources := []sourceDB{
		{Label: "english", URL: strings.TrimSpace(englishURL)},
		{Label: "spanish", URL: strings.TrimSpace(spanishURL)},
	}

	report := auditReport{
		Mode:                "dry-run",
		GeneratedAt:         time.Now().UTC(),
		TelegramMultiCourse: []telegramMultiCourseUser{},
		TelegramConflicts:   []identityConflict{},
		IdentityConflicts:   []stableIdentityConflict{},
		Notes: []string{
			"audit foundation: read-only counts, readiness, attempt sources, telegram overlap",
			"use --commit --phase=users|user-courses|course-mappings|content|attempts|srs for unified DB write slices",
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

	multiCourse, conflicts, err := auditTelegram(ctx, openSources, targetDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge audit telegram scan failed: %v\n", err)
		os.Exit(1)
	}
	identityConflicts := []stableIdentityConflict{}
	report.TelegramMultiCourse = multiCourse
	if conflicts == nil {
		conflicts = []identityConflict{}
	}
	report.TelegramConflicts = conflicts
	report.IdentityConflicts = identityConflicts

	allSourcesReady := true
	for _, src := range report.Sources {
		if !src.Readiness.ReadyForDryRunMerge {
			allSourcesReady = false
			break
		}
	}
	report.ReadyForSourceMerge = len(conflicts) == 0 && len(identityConflicts) == 0 && len(openSources) == 2 && allSourcesReady
	report.ReadyForWriteMerge = report.ReadyForSourceMerge && targetDB != nil && targetSummary.LegacyMappingTablesOK
	if targetDB == nil {
		report.Notes = append(report.Notes, "target DB URL is not set; ready_for_write_merge stays false until TARGET_DATABASE_URL is provided")
	}
	if len(openSources) < 2 {
		report.Notes = append(report.Notes, "provide both English and Spanish source URLs before real merge planning")
	}
	if len(multiCourse) > 0 {
		report.Notes = append(report.Notes, fmt.Sprintf("%d telegram users span both English and Spanish and will merge into one target user", len(multiCourse)))
	}
	if len(conflicts) > 0 {
		report.Notes = append(report.Notes, "telegram identity conflicts exist; real merge must map them explicitly")
	}
	if len(identityConflicts) > 0 {
		report.Notes = append(report.Notes, "stable identity conflicts exist; real merge must map them explicitly")
	}
	if !allSourcesReady {
		report.Notes = append(report.Notes, "one or more source DBs are not ready for merge; fix readiness gaps first")
	}

	if commit {
		if phase == "" {
			fmt.Fprintln(os.Stderr, "--commit requires --phase=users|user-courses|course-mappings|content|attempts|srs")
			os.Exit(1)
		}
		if targetDB == nil {
			fmt.Fprintln(os.Stderr, "TARGET_DATABASE_URL is required for write merge")
			os.Exit(1)
		}
		if !report.ReadyForWriteMerge {
			fmt.Fprintln(os.Stderr, "write merge blocked: resolve audit conflicts and source readiness first")
			os.Exit(1)
		}
		writeCtx, writeCancel := context.WithTimeout(context.Background(), timeout)
		defer writeCancel()
		summary, err := runWritePhase(writeCtx, phase, openSources, targetDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write merge failed: %v\n", err)
			os.Exit(1)
		}
		report.Mode = "commit"
		report.WritePhase = phase
		report.WriteSummary = summary
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
		AttemptSources: []attemptSourceRow{},
	}
	if url == "" {
		return s, nil, nil
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		return s, nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
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
	r.SRSWordMissing, err = countMissingWordSRS(ctx, db)
	if err != nil {
		return r, err
	}
	r.SRSGrammarMissing, err = countMissingGrammarSRS(ctx, db)
	if err != nil {
		return r, err
	}
	r.SRSBackfillGap = r.SRSWordMissing + r.SRSGrammarMissing
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
