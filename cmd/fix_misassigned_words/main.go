// One-off remediation for word_cards that ended up tagged with the wrong
// course_code (e.g. Spanish words attached to the English course because of
// a course-tagging gap in the word-set importer) and got TTS audio generated
// in the wrong voice as a result.
//
// For every word that:
//   - is linked into a word_set belonging to --to-course, and
//   - has word_cards.course_code = --from-course (or is untagged but has a
//     ready/terminal TTS status recorded under --from-course),
//
// this tool deletes the word_cards row (cascading to word_forms,
// training_cards, user_cards, word_set_items, etc.), removes the stray
// --from-course TTS audio file and status row, then re-creates the word_card
// tagged with --to-course and re-links it into the same word_sets it was
// removed from. The word is left with no TTS status so the normal pipeline
// (PronunciationService / tts-worker) regenerates audio in the correct voice.
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

// dbQuerier is the read-only subset of *sql.DB used while gathering candidates.
type dbQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// dbExecQuerier additionally supports starting a transaction, used while applying fixes.
type dbExecQuerier interface {
	dbQuerier
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type wordSetLink struct {
	WordSetID int64
	SortOrder int
}

type candidate struct {
	WordCardID   int64
	Word         string
	CourseCode   string // current (wrong) course_code on word_cards, "" if untagged
	AudioRelPath string // from-course tts_generation_status audio path, "" if none
	Links        []wordSetLink
}

func main() {
	os.Exit(run())
}

func run() int {
	fromCourse := flag.String("from-course", "en_ru", "course_code the words are wrongly tagged with")
	toCourse := flag.String("to-course", "es_ru", "course_code the words actually belong to")
	commit := flag.Bool("commit", false, "apply changes (default: dry-run)")
	onlyWord := flag.String("only-word", "", "process only this word (optional, for testing)")
	limit := flag.Int("limit", 0, "max number of words to fix (0 = no limit)")
	flag.Parse()

	from := strings.ToLower(strings.TrimSpace(*fromCourse))
	to := strings.ToLower(strings.TrimSpace(*toCourse))
	if from == "" || to == "" || from == to {
		fmt.Fprintln(os.Stderr, "from-course and to-course must be set and different")
		return 1
	}

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

	candidates, err := findCandidates(conn, from, to, *onlyWord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find candidates: %v\n", err)
		return 1
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].WordCardID < candidates[j].WordCardID })
	if *limit > 0 && len(candidates) > *limit {
		candidates = candidates[:*limit]
	}

	fmt.Printf("found %d word(s) tagged %q but linked into %q word sets\n", len(candidates), from, to)
	for _, c := range candidates {
		linkDesc := make([]string, 0, len(c.Links))
		for _, l := range c.Links {
			linkDesc = append(linkDesc, fmt.Sprintf("set#%d@%d", l.WordSetID, l.SortOrder))
		}
		fmt.Printf("  id=%d word=%q course_code=%q audio=%q links=[%s]\n",
			c.WordCardID, c.Word, c.CourseCode, c.AudioRelPath, strings.Join(linkDesc, ","))
	}

	if !*commit {
		fmt.Println("dry-run: pass -commit to apply")
		return 0
	}

	fixed := 0
	for _, c := range candidates {
		if err := fixOne(context.Background(), conn, audioDir, from, to, c); err != nil {
			fmt.Fprintf(os.Stderr, "failed to fix word=%q id=%d: %v\n", c.Word, c.WordCardID, err)
			continue
		}
		fixed++
	}
	fmt.Printf("fixed %d/%d word(s)\n", fixed, len(candidates))
	return 0
}

func findCandidates(conn dbQuerier, from, to, onlyWord string) ([]candidate, error) {
	query := `
		SELECT DISTINCT wc.id, wc.word, COALESCE(wc.course_code, '')
		FROM word_cards wc
		JOIN word_set_items wsi ON wsi.word_card_id = wc.id
		JOIN word_sets ws ON ws.id = wsi.word_set_id
		WHERE ws.course_code = ?
		  AND (
			wc.course_code = ?
			OR EXISTS (
				SELECT 1 FROM tts_generation_status tgs
				WHERE LOWER(tgs.word) = LOWER(wc.word) AND tgs.course_code = ?
				  AND tgs.state IN ('ready', 'failed_terminal')
			)
		  )
	`
	args := []interface{}{to, from, from}
	if strings.TrimSpace(onlyWord) != "" {
		query += " AND LOWER(wc.word) = LOWER(?)"
		args = append(args, strings.TrimSpace(onlyWord))
	}

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query candidates: %w", err)
	}
	defer rows.Close()

	out := make([]candidate, 0)
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.WordCardID, &c.Word, &c.CourseCode); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		links, err := loadLinks(conn, out[i].WordCardID)
		if err != nil {
			return nil, fmt.Errorf("load links for word_card_id=%d: %w", out[i].WordCardID, err)
		}
		out[i].Links = links

		rel, err := loadAudioRelPath(conn, from, out[i].Word)
		if err != nil {
			return nil, fmt.Errorf("load audio path for word=%q: %w", out[i].Word, err)
		}
		out[i].AudioRelPath = rel
	}
	return out, nil
}

func loadLinks(conn dbQuerier, wordCardID int64) ([]wordSetLink, error) {
	rows, err := conn.Query(`SELECT word_set_id, sort_order FROM word_set_items WHERE word_card_id = ?`, wordCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []wordSetLink
	for rows.Next() {
		var l wordSetLink
		if err := rows.Scan(&l.WordSetID, &l.SortOrder); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
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

func fixOne(ctx context.Context, conn dbExecQuerier, audioDir, from, to string, c candidate) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete the stray from-course TTS status row (and its audio file, after commit).
	if _, err := tx.Exec(`DELETE FROM tts_generation_status WHERE course_code = ? AND LOWER(word) = LOWER(?)`, from, c.Word); err != nil {
		return fmt.Errorf("delete tts status: %w", err)
	}

	// word_cards delete cascades to word_forms, training_cards, user_cards,
	// user_word_mastering, user_word_knowledge, word_set_items, word_request_history.
	if _, err := tx.Exec(`DELETE FROM word_cards WHERE id = ?`, c.WordCardID); err != nil {
		return fmt.Errorf("delete word_cards: %w", err)
	}

	var newID int64
	if err := tx.QueryRow(`
		INSERT INTO word_cards (word, definition, course_code, updated_at)
		VALUES (?, '', ?, CURRENT_TIMESTAMP)
		RETURNING id`, c.Word, to).Scan(&newID); err != nil {
		return fmt.Errorf("recreate word_cards: %w", err)
	}

	for _, l := range c.Links {
		if _, err := tx.Exec(`
			INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
			VALUES (?, ?, ?)
			ON CONFLICT (word_set_id, word_card_id) DO NOTHING`, l.WordSetID, newID, l.SortOrder); err != nil {
			return fmt.Errorf("relink word_set_items set=%d: %w", l.WordSetID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	if c.AudioRelPath != "" {
		full := filepath.Join(audioDir, c.AudioRelPath)
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to remove stray audio file %s: %v\n", full, err)
		}
	}
	return nil
}
