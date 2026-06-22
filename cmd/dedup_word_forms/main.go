// One-off remediation for word_cards that are surface forms rather than lemmas.
//
// Seed word sets sometimes contain a word form (e.g. "tokenizado") instead of its
// lemma ("tokenizar"). Each such word_card still gets a full set of training_cards
// that teach the *lemma*, so a single lemma ends up taught by several duplicate
// training-card sets — which makes spaced repetition study the same word repeatedly.
//
// Going forward the TrainingWorker canonicalizes new cards on the lemma. This tool
// fixes the historical data: for every course it groups already-processed word_cards
// by their lemma (taken from the stored training_cards, no AI calls), and for each
// group of >1 card it:
//   - picks a canonical card (prefer word == lemma, else most training cards, else
//     lowest id) and, if its surface word differs from the lemma, renames it to the
//     lemma (dropping stale pronunciation so it regenerates),
//   - merges every other card in the group into the canonical one via the shared
//     repository.MergeWordCardTx helper (relinks word_set_items, learning_items,
//     user_word_knowledge, user_word_mastering; deletes the form, cascading away its
//     training_cards/user_cards/review_events). Canonical progress wins on conflicts.
//
// Cards without training_cards (not yet processed) are skipped and reported: they have
// produced no duplicate training cards yet, and the worker will canonicalize them when
// it processes them.
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

	"tgbot-skeleton/internal/wordmerge"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type wordCardRow struct {
	ID          int64
	Word        string
	CourseCode  string
	Lemma       string // derived from training_cards.word_en; "" if not derivable
	TCount      int
	HasTraining bool
}

type groupKey struct {
	course string
	lemma  string
}

func main() {
	os.Exit(run())
}

func run() int {
	commit := flag.Bool("commit", false, "apply changes (default: dry-run)")
	courseFilter := flag.String("course", "", "restrict to a single course_code (optional)")
	limit := flag.Int("limit", 0, "max number of duplicate groups to fix (0 = no limit)")
	flag.Parse()

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		return 1
	}
	conn, err := sql.Open("pgx", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		return 1
	}
	defer conn.Close()
	if err := conn.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect db: %v\n", err)
		return 1
	}

	audioDir := strings.TrimSpace(os.Getenv("TTS_AUDIO_DIR"))
	if audioDir == "" {
		audioDir = "/app/data/tts"
	}

	rows, err := loadWordCards(conn, strings.TrimSpace(*courseFilter))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load word cards: %v\n", err)
		return 1
	}

	// Group by (course, lemma); only cards with a derivable lemma participate.
	groups := make(map[groupKey][]wordCardRow)
	skippedNoLemma := 0
	for _, r := range rows {
		if r.Lemma == "" {
			skippedNoLemma++
			continue
		}
		k := groupKey{course: r.CourseCode, lemma: r.Lemma}
		groups[k] = append(groups[k], r)
	}

	// Keep only groups with duplicates, in a stable order.
	dupKeys := make([]groupKey, 0)
	for k, g := range groups {
		if len(g) > 1 {
			dupKeys = append(dupKeys, k)
		}
	}
	sort.Slice(dupKeys, func(i, j int) bool {
		if dupKeys[i].course != dupKeys[j].course {
			return dupKeys[i].course < dupKeys[j].course
		}
		return dupKeys[i].lemma < dupKeys[j].lemma
	})
	if *limit > 0 && len(dupKeys) > *limit {
		dupKeys = dupKeys[:*limit]
	}

	totalForms := 0
	for _, k := range dupKeys {
		g := groups[k]
		canonical := pickCanonical(g)
		forms := make([]wordCardRow, 0, len(g)-1)
		for _, c := range g {
			if c.ID != canonical.ID {
				forms = append(forms, c)
			}
		}
		totalForms += len(forms)
		formWords := make([]string, 0, len(forms))
		for _, f := range forms {
			formWords = append(formWords, fmt.Sprintf("%q(id=%d,tc=%d)", f.Word, f.ID, f.TCount))
		}
		rename := ""
		if !strings.EqualFold(canonical.Word, k.lemma) {
			rename = fmt.Sprintf(" [rename canonical %q -> %q]", canonical.Word, k.lemma)
		}
		fmt.Printf("course=%s lemma=%q canonical=id%d(%q,tc=%d)%s merge %d form(s): %s\n",
			k.course, k.lemma, canonical.ID, canonical.Word, canonical.TCount, rename, len(forms), strings.Join(formWords, " "))
	}

	fmt.Printf("\n%d duplicate lemma group(s), %d form card(s) to merge; %d card(s) skipped (no lemma yet)\n",
		len(dupKeys), totalForms, skippedNoLemma)

	if !*commit {
		fmt.Println("dry-run: pass -commit to apply")
		return 0
	}

	fixed := 0
	for _, k := range dupKeys {
		if err := dedupGroup(context.Background(), conn, audioDir, k, groups[k]); err != nil {
			fmt.Fprintf(os.Stderr, "failed to dedup course=%s lemma=%q: %v\n", k.course, k.lemma, err)
			continue
		}
		fixed++
	}
	fmt.Printf("deduped %d/%d group(s)\n", fixed, len(dupKeys))
	return 0
}

func loadWordCards(conn *sql.DB, courseFilter string) ([]wordCardRow, error) {
	query := `
		SELECT wc.id, wc.word, COALESCE(wc.course_code, ''),
		       (SELECT tc.word_en FROM training_cards tc
		        WHERE tc.word_card_id = wc.id ORDER BY tc.sense_index LIMIT 1) AS lemma_src,
		       (SELECT COUNT(*) FROM training_cards tc WHERE tc.word_card_id = wc.id) AS tc_count
		FROM word_cards wc`
	args := []interface{}{}
	if courseFilter != "" {
		query += " WHERE wc.course_code = ?"
		args = append(args, courseFilter)
	}

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query word_cards: %w", err)
	}
	defer rows.Close()

	out := make([]wordCardRow, 0)
	for rows.Next() {
		var r wordCardRow
		var lemmaSrc sql.NullString
		if err := rows.Scan(&r.ID, &r.Word, &r.CourseCode, &lemmaSrc, &r.TCount); err != nil {
			return nil, fmt.Errorf("scan word_card: %w", err)
		}
		r.HasTraining = r.TCount > 0
		if lemmaSrc.Valid {
			r.Lemma = strings.ToLower(strings.TrimSpace(lemmaSrc.String))
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// pickCanonical chooses the card to keep for a lemma group: prefer one whose surface
// word already equals the lemma, else the one with the most training cards, else the
// lowest id (stable).
func pickCanonical(g []wordCardRow) wordCardRow {
	best := g[0]
	for _, c := range g[1:] {
		if better(c, best) {
			best = c
		}
	}
	return best
}

func better(a, b wordCardRow) bool {
	aMatch := strings.EqualFold(a.Word, a.Lemma)
	bMatch := strings.EqualFold(b.Word, b.Lemma)
	if aMatch != bMatch {
		return aMatch
	}
	if a.TCount != b.TCount {
		return a.TCount > b.TCount
	}
	return a.ID < b.ID
}

func loadAudioRelPath(conn *sql.DB, courseCode, word string) (string, error) {
	var rel sql.NullString
	err := conn.QueryRow(`SELECT audio_rel_path FROM tts_generation_status
		WHERE course_code = ? AND LOWER(word) = LOWER(?)`, courseCode, word).Scan(&rel)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !rel.Valid {
		return "", nil
	}
	return rel.String, nil
}

func dedupGroup(ctx context.Context, conn *sql.DB, audioDir string, k groupKey, g []wordCardRow) error {
	canonical := pickCanonical(g)
	renameCanonical := !strings.EqualFold(canonical.Word, k.lemma)

	// Collect audio paths to remove after commit (form words + canonical's old word
	// when it is renamed) — file removal isn't transactional.
	audioToRemove := make([]string, 0)
	staleTTSWords := make([]string, 0)
	for _, c := range g {
		if c.ID == canonical.ID && !renameCanonical {
			continue
		}
		rel, err := loadAudioRelPath(conn, k.course, c.Word)
		if err != nil {
			return fmt.Errorf("load audio path for %q: %w", c.Word, err)
		}
		if rel != "" {
			audioToRemove = append(audioToRemove, rel)
		}
		staleTTSWords = append(staleTTSWords, c.Word)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if renameCanonical {
		if _, err := tx.Exec(`UPDATE word_cards SET word = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			k.lemma, canonical.ID); err != nil {
			return fmt.Errorf("rename canonical word_card: %w", err)
		}
	}

	for _, c := range g {
		if c.ID == canonical.ID {
			continue
		}
		if err := wordmerge.MergeWordCardTx(tx, c.ID, canonical.ID); err != nil {
			return fmt.Errorf("merge form id=%d: %w", c.ID, err)
		}
	}

	// Drop stale pronunciation for merged forms and (when renamed) the canonical's old
	// surface word, so audio regenerates for the lemma.
	for _, word := range staleTTSWords {
		if _, err := tx.Exec(`DELETE FROM tts_generation_status WHERE course_code = ? AND LOWER(word) = LOWER(?)`,
			k.course, word); err != nil {
			return fmt.Errorf("delete tts status for %q: %w", word, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	for _, rel := range audioToRemove {
		full := filepath.Join(audioDir, rel)
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to remove stale audio file %s: %v\n", full, err)
		}
	}
	return nil
}
