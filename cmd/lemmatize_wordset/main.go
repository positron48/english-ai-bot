// One-off cleanup for a seed word-set JSON (e.g. resources/wordsets/es_ru_seed_word_sets.json).
//
// The seed file was authored with surface forms in places ("tokenizado") instead of
// lemmas ("tokenizar"), and a lemma can appear via several forms across levels. This
// tool rewrites the file so it contains only lemmas, deduplicated across all levels
// (a lemma is kept at its first occurrence, preserving level order and sort_order).
//
// Lemmas are taken from the live DB (word_cards.word -> the lemma stored on its first
// training_card), the same derivation cmd/dedup_word_forms uses. Words with no card or
// no derivable lemma in the DB are kept unchanged and reported, so nothing is lost
// silently. Run cmd/dedup_word_forms first so the DB already holds canonical lemmas.
//
// Defaults to a dry run (prints a summary); pass -write to overwrite the JSON file.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type wordSet struct {
	Level       string   `json:"level"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
	TargetCount int      `json:"target_count"`
	Words       []string `json:"words"`
}

type seedFile struct {
	CourseCode      string    `json:"course_code"`
	SourceMigration string    `json:"source_migration"`
	Sets            []wordSet `json:"sets"`
}

func main() {
	os.Exit(run())
}

func run() int {
	jsonPath := flag.String("json", "resources/wordsets/es_ru_seed_word_sets.json", "path to seed word-set JSON")
	courseFilter := flag.String("course", "", "course_code for DB lemma lookup (default: file's course_code)")
	write := flag.Bool("write", false, "overwrite the JSON file (default: dry-run summary)")
	flag.Parse()

	raw, err := os.ReadFile(*jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", *jsonPath, err)
		return 1
	}
	var seed seedFile
	if err := json.Unmarshal(raw, &seed); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse %s: %v\n", *jsonPath, err)
		return 1
	}

	course := strings.TrimSpace(*courseFilter)
	if course == "" {
		course = seed.CourseCode
	}

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

	wordToLemma, err := loadLemmaMap(conn, course)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load lemma map: %v\n", err)
		return 1
	}

	seen := make(map[string]struct{}) // lemmas already kept (dedup across all levels)
	var unmapped []string
	droppedDup := 0
	mappedForm := 0
	totalBefore := 0

	for si := range seed.Sets {
		s := &seed.Sets[si]
		newWords := make([]string, 0, len(s.Words))
		for _, w := range s.Words {
			totalBefore++
			key := strings.ToLower(strings.TrimSpace(w))
			lemma, ok := wordToLemma[key]
			if !ok || lemma == "" {
				lemma = key
				unmapped = append(unmapped, w)
			} else if lemma != key {
				mappedForm++
			}
			if _, dup := seen[lemma]; dup {
				droppedDup++
				continue
			}
			seen[lemma] = struct{}{}
			newWords = append(newWords, lemma)
		}
		s.Words = newWords
	}

	totalAfter := 0
	for _, s := range seed.Sets {
		totalAfter += len(s.Words)
	}

	fmt.Printf("course=%s sets=%d words: %d -> %d (forms mapped to lemma: %d, cross-level dups dropped: %d)\n",
		course, len(seed.Sets), totalBefore, totalAfter, mappedForm, droppedDup)
	if len(unmapped) > 0 {
		sort.Strings(unmapped)
		uniq := dedupStrings(unmapped)
		fmt.Printf("%d word(s) had no DB lemma (kept as-is): %s\n", len(uniq), strings.Join(firstN(uniq, 40), ", "))
	}

	if !*write {
		fmt.Println("dry-run: pass -write to overwrite the file")
		return 0
	}

	out, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal: %v\n", err)
		return 1
	}
	out = append(out, '\n')
	if err := os.WriteFile(*jsonPath, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *jsonPath, err)
		return 1
	}
	fmt.Printf("wrote %s\n", *jsonPath)
	return 0
}

// loadLemmaMap returns lower-cased surface word -> lemma for a course, where the lemma
// is the word stored on the card's first training_card (sense_index 0). Cards without
// training cards are omitted.
func loadLemmaMap(conn *sql.DB, course string) (map[string]string, error) {
	rows, err := conn.Query(`
		SELECT wc.word,
		       (SELECT tc.word_en FROM training_cards tc
		        WHERE tc.word_card_id = wc.id ORDER BY tc.sense_index LIMIT 1) AS lemma_src
		FROM word_cards wc
		WHERE wc.course_code = ?`, course)
	if err != nil {
		return nil, fmt.Errorf("query lemmas: %w", err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var word string
		var lemmaSrc sql.NullString
		if err := rows.Scan(&word, &lemmaSrc); err != nil {
			return nil, fmt.Errorf("scan lemma row: %w", err)
		}
		if !lemmaSrc.Valid {
			continue
		}
		lemma := strings.ToLower(strings.TrimSpace(lemmaSrc.String))
		if lemma == "" {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(word))] = lemma
	}
	return out, rows.Err()
}

func dedupStrings(in []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func firstN(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return append(in[:n:n], "...")
}
