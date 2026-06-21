// One-off remediation for word_cards that are shared between multiple courses
// (e.g. an es_ru-tagged "real"/"social"/"hotel" word_card whose word_set_items
// also link it into the en_ru course, because the word happens to be spelled
// the same in both languages). A shared word_card has nothing in common
// between courses except spelling: translation, training cards (examples,
// distractors) and pronunciation are all language-specific.
//
// For every word_card linked (via word_set_items -> word_sets.course_code)
// into more than one distinct course, this tool:
//   - creates a brand new word_cards row per linked course (word, course_code),
//   - relinks that course's word_set_items rows from the old shared id to the
//     new per-course id,
//   - deletes the old shared word_cards row (cascades to word_forms,
//     training_cards, user_cards, review_events, user_word_knowledge,
//     user_word_mastering, word_request_history - i.e. resets all user
//     progress on the word, since it was tracked against the wrong content),
//   - deletes any tts_generation_status rows (and audio files) for the word
//     under every affected course so pronunciation is regenerated fresh.
//
// The new word_cards rows are left with processed_at=NULL and no training
// cards, so the normal pipelines (TrainingWorker fills dictionary fields and
// training cards; PronunciationService/tts-worker regenerates audio) pick
// them up and regenerate everything from scratch, this time using the
// correct per-course AI prompts.
//
// Defaults to a dry run; pass -commit to apply changes.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
)

type dbQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type dbExecQuerier interface {
	dbQuerier
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type wordSetLink struct {
	WordSetID int64
	SortOrder int
}

type sharedWordCard struct {
	WordCardID int64
	Word       string
	Courses    []string                // distinct course codes this word_card is linked into, sorted
	LinksByC   map[string][]wordSetLink // course_code -> word_set_items rows for that course
}

func main() {
	os.Exit(run())
}

func run() int {
	commit := flag.Bool("commit", false, "apply changes (default: dry-run)")
	onlyWord := flag.String("only-word", "", "process only this word (optional, for testing)")
	limit := flag.Int("limit", 0, "max number of shared words to fix (0 = no limit)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		return 1
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init db: %v\n", err)
		return 1
	}
	defer db.Close()
	conn := db.GetConnection()

	audioDir := strings.TrimSpace(cfg.TTS.AudioDir)
	if audioDir == "" {
		audioDir = "/app/data/tts"
	}

	shared, err := findSharedWordCards(conn, *onlyWord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find shared word_cards: %v\n", err)
		return 1
	}
	sort.Slice(shared, func(i, j int) bool { return shared[i].WordCardID < shared[j].WordCardID })
	if *limit > 0 && len(shared) > *limit {
		shared = shared[:*limit]
	}

	fmt.Printf("found %d word_card(s) shared across multiple courses\n", len(shared))
	for _, s := range shared {
		fmt.Printf("  id=%d word=%q courses=%v\n", s.WordCardID, s.Word, s.Courses)
	}

	if !*commit {
		fmt.Println("dry-run: pass -commit to apply")
		return 0
	}

	fixed := 0
	for _, s := range shared {
		if err := splitOne(context.Background(), conn, audioDir, s); err != nil {
			fmt.Fprintf(os.Stderr, "failed to split word=%q id=%d: %v\n", s.Word, s.WordCardID, err)
			continue
		}
		fixed++
	}
	fmt.Printf("split %d/%d word(s)\n", fixed, len(shared))
	return 0
}

func findSharedWordCards(conn dbQuerier, onlyWord string) ([]sharedWordCard, error) {
	query := `
		SELECT wc.id, wc.word
		FROM word_cards wc
		JOIN word_set_items wsi ON wsi.word_card_id = wc.id
		JOIN word_sets ws ON ws.id = wsi.word_set_id
	`
	args := []interface{}{}
	if strings.TrimSpace(onlyWord) != "" {
		query += " WHERE LOWER(wc.word) = LOWER(?)"
		args = append(args, strings.TrimSpace(onlyWord))
	}
	query += `
		GROUP BY wc.id, wc.word
		HAVING COUNT(DISTINCT ws.course_code) > 1
	`

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query shared word_cards: %w", err)
	}
	defer rows.Close()

	out := make([]sharedWordCard, 0)
	for rows.Next() {
		var s sharedWordCard
		if err := rows.Scan(&s.WordCardID, &s.Word); err != nil {
			return nil, fmt.Errorf("scan shared word_card: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		linksByC, err := loadLinksByCourse(conn, out[i].WordCardID)
		if err != nil {
			return nil, fmt.Errorf("load links for word_card_id=%d: %w", out[i].WordCardID, err)
		}
		out[i].LinksByC = linksByC
		courses := make([]string, 0, len(linksByC))
		for c := range linksByC {
			courses = append(courses, c)
		}
		sort.Strings(courses)
		out[i].Courses = courses
	}
	return out, nil
}

func loadLinksByCourse(conn dbQuerier, wordCardID int64) (map[string][]wordSetLink, error) {
	rows, err := conn.Query(`
		SELECT ws.course_code, wsi.word_set_id, wsi.sort_order
		FROM word_set_items wsi
		JOIN word_sets ws ON ws.id = wsi.word_set_id
		WHERE wsi.word_card_id = ?`, wordCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]wordSetLink)
	for rows.Next() {
		var course string
		var l wordSetLink
		if err := rows.Scan(&course, &l.WordSetID, &l.SortOrder); err != nil {
			return nil, err
		}
		out[course] = append(out[course], l)
	}
	return out, rows.Err()
}

func loadAudioRelPath(conn dbQuerier, courseCode, word string) (string, error) {
	var rel *string
	err := conn.QueryRow(`
		SELECT audio_rel_path FROM tts_generation_status
		WHERE course_code = ? AND LOWER(word) = LOWER(?)`, courseCode, word).Scan(&rel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if rel == nil {
		return "", nil
	}
	return *rel, nil
}

func splitOne(ctx context.Context, conn dbExecQuerier, audioDir string, s sharedWordCard) error {
	// Collect stray audio paths before the transaction (read-only) so we can
	// remove the files after a successful commit.
	audioPaths := make([]string, 0, len(s.Courses))
	for _, course := range s.Courses {
		rel, err := loadAudioRelPath(conn, course, s.Word)
		if err != nil {
			return fmt.Errorf("load audio path for course=%s: %w", course, err)
		}
		if rel != "" {
			audioPaths = append(audioPaths, rel)
		}
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, course := range s.Courses {
		var newID int64
		if err := tx.QueryRow(`
			INSERT INTO word_cards (word, definition, course_code, updated_at)
			VALUES (?, '', ?, CURRENT_TIMESTAMP)
			RETURNING id`, s.Word, course).Scan(&newID); err != nil {
			return fmt.Errorf("create per-course word_cards row for course=%s: %w", course, err)
		}

		for _, l := range s.LinksByC[course] {
			if _, err := tx.Exec(`
				UPDATE word_set_items SET word_card_id = ?
				WHERE word_set_id = ? AND word_card_id = ?`, newID, l.WordSetID, s.WordCardID); err != nil {
				return fmt.Errorf("relink word_set_items set=%d course=%s: %w", l.WordSetID, course, err)
			}
		}

		if _, err := tx.Exec(`DELETE FROM tts_generation_status WHERE course_code = ? AND LOWER(word) = LOWER(?)`, course, s.Word); err != nil {
			return fmt.Errorf("delete tts status for course=%s: %w", course, err)
		}
	}

	// All word_set_items referencing the old shared row have been relinked
	// above; deleting it now cascades to word_forms, training_cards,
	// user_cards, review_events, user_word_knowledge, user_word_mastering,
	// and word_request_history - resetting any progress tracked against the
	// shared (wrong-language) content.
	if _, err := tx.Exec(`DELETE FROM word_cards WHERE id = ?`, s.WordCardID); err != nil {
		return fmt.Errorf("delete shared word_cards row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	for _, rel := range audioPaths {
		full := filepath.Join(audioDir, rel)
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to remove stray audio file %s: %v\n", full, err)
		}
	}
	return nil
}
