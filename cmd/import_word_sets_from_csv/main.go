package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/resources/wordsets"
)

type csvWord struct {
	Lemma           string
	POS             string
	PopularityCount int
}

var ranksRe = regexp.MustCompile(`(?i)(?:ranks?|rangos?)\s+(\d+)\s*[\p{Pd}-]\s*(\d+)`)
var spanishLemmaRe = regexp.MustCompile(`^[a-záéíóúüñ]+(?:-[a-záéíóúüñ]+)*$`)

func main() {
	os.Exit(runCLI())
}

func runCLI() int {
	csvPath := flag.String("csv", "data/spanish_word_freq_pos_ud_top6000.csv", "path to frequency CSV")
	commit := flag.Bool("commit", false, "apply changes to DB (default: dry-run)")
	onlySetsRaw := flag.String("only-set", "", "comma-separated exact set titles to import (optional)")
	limitSets := flag.Int("limit-sets", 0, "limit number of sets to process after filtering (0 = no limit)")
	flag.Parse()

	ctx := context.Background()

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
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		return 1
	}
	defer db.Close()

	wordsByPOS, err := loadWordsByTrainingPOS(*csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read csv: %v\n", err)
		return 1
	}

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), log)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), log)
	wordRepo := repository.NewWordRepository(db.GetConnection(), log)
	svc := service.NewWordSetService(
		wordSetRepo,
		wordSetCategoryRepo,
		wordRepo,
		nil,
		nil,
		nil,
		nil,
		cfg.Learning,
		"",
		log,
	)

	sets, err := wordSetRepo.ListWordSets(nil, 1000, 0, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list word sets: %v\n", err)
		return 1
	}

	selectedSetTitles := parseCSVSet(*onlySetsRaw)
	targetSets := selectRankedSets(sets, selectedSetTitles)
	if *limitSets > 0 && len(targetSets) > *limitSets {
		targetSets = targetSets[:*limitSets]
	}

	if len(targetSets) == 0 {
		fmt.Println("No matching sets found to process.")
		return 0
	}

	mode := "DRY-RUN"
	if *commit {
		mode = "COMMIT"
	}
	fmt.Printf("[%s] Processing %d set(s)\n", mode, len(targetSets))

	processed := 0
	for _, ws := range targetSets {
		if ws.PreferredPOS == nil || strings.TrimSpace(*ws.PreferredPOS) == "" {
			fmt.Printf("SKIP set_id=%d title=%q: preferred_pos is empty\n", ws.ID, ws.Title)
			continue
		}
		pos := strings.ToLower(strings.TrimSpace(*ws.PreferredPOS))
		rankStart, rankEnd, ok := extractRankRangeFromSet(ws)
		if !ok {
			fmt.Printf("SKIP set_id=%d title=%q: no rank range in title/description\n", ws.ID, ws.Title)
			continue
		}

		candidates := wordsByPOS[pos]
		if len(candidates) == 0 {
			fmt.Printf("SKIP set_id=%d title=%q: no words for preferred_pos=%q\n", ws.ID, ws.Title, pos)
			continue
		}
		if rankStart > len(candidates) {
			fmt.Printf("SKIP set_id=%d title=%q: range starts at %d but only %d words in pos=%s\n",
				ws.ID, ws.Title, rankStart, len(candidates), pos)
			continue
		}

		effectiveEnd := rankEnd
		if effectiveEnd > len(candidates) {
			effectiveEnd = len(candidates)
		}
		lemmas := make([]string, 0, effectiveEnd-rankStart+1)
		for i := rankStart - 1; i < effectiveEnd; i++ {
			lemmas = append(lemmas, candidates[i].Lemma)
		}

		fmt.Printf("SET id=%d title=%q pos=%s range=%d-%d imported=%d\n",
			ws.ID, ws.Title, pos, rankStart, rankEnd, len(lemmas))

		if *commit {
			if err := svc.ProcessWordSetItems(ctx, ws.ID, strings.Join(lemmas, ",")); err != nil {
				fmt.Printf("ERROR set_id=%d title=%q: %v\n", ws.ID, ws.Title, err)
				continue
			}
		}
		processed++
	}

	fmt.Printf("Done. processed=%d (mode=%s)\n", processed, mode)
	return 0
}

func parseCSVSet(v string) map[string]struct{} {
	res := map[string]struct{}{}
	for _, part := range strings.Split(v, ",") {
		s := strings.TrimSpace(part)
		if s == "" {
			continue
		}
		res[s] = struct{}{}
	}
	return res
}

func selectRankedSets(all []*models.WordSet, only map[string]struct{}) []*models.WordSet {
	out := make([]*models.WordSet, 0, len(all))
	for _, ws := range all {
		if ws == nil {
			continue
		}
		if _, _, ok := extractRankRangeFromSet(ws); !ok {
			continue
		}
		if len(only) > 0 {
			if _, ok := only[ws.Title]; !ok {
				continue
			}
		}
		out = append(out, ws)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder == out[j].SortOrder {
			return out[i].Title < out[j].Title
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out
}

func extractRankRange(title string) (start int, end int, ok bool) {
	m := ranksRe.FindStringSubmatch(title)
	if len(m) != 3 {
		return 0, 0, false
	}
	s, err1 := strconv.Atoi(m[1])
	e, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil || s <= 0 || e < s {
		return 0, 0, false
	}
	return s, e, true
}

func extractRankRangeFromSet(ws *models.WordSet) (start int, end int, ok bool) {
	if ws == nil {
		return 0, 0, false
	}
	if s, e, ok := extractRankRange(ws.Title); ok {
		return s, e, true
	}
	if ws.Description != nil {
		if s, e, ok := extractRankRange(*ws.Description); ok {
			return s, e, true
		}
	}
	return 0, 0, false
}

func loadWordsByTrainingPOS(path string) (map[string][]csvWord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("csv has no data rows")
	}

	// header indexes
	header := make(map[string]int)
	for i, name := range rows[0] {
		header[strings.TrimSpace(strings.ToLower(name))] = i
	}
	idxLemma, okLemma := header["lemma"]
	idxPOS, okPOS := header["pos"]
	idxPop, okPop := header["popularity_count"]
	if !okLemma || !okPOS || !okPop {
		return nil, fmt.Errorf("csv must contain columns: lemma,pos,popularity_count")
	}

	byPOS := map[string][]csvWord{
		"noun":      {},
		"verb":      {},
		"adjective": {},
		"adverb":    {},
	}
	seen := map[string]map[string]struct{}{
		"noun":      {},
		"verb":      {},
		"adjective": {},
		"adverb":    {},
	}

	for _, row := range rows[1:] {
		if idxLemma >= len(row) || idxPOS >= len(row) || idxPop >= len(row) {
			continue
		}
		lemma := wordsets.NormalizeLemmaImport(strings.TrimSpace(row[idxLemma]))
		udPOS := strings.TrimSpace(strings.ToUpper(row[idxPOS]))
		popStr := strings.TrimSpace(row[idxPop])
		if lemma == "" || udPOS == "" || popStr == "" {
			continue
		}
		if !isValidSpanishLemma(lemma) {
			continue
		}
		pop, err := strconv.Atoi(popStr)
		if err != nil {
			continue
		}

		trainingPOS, ok := mapUDToTrainingPOS(udPOS)
		if !ok {
			continue
		}
		if _, exists := seen[trainingPOS][lemma]; exists {
			continue
		}
		seen[trainingPOS][lemma] = struct{}{}
		byPOS[trainingPOS] = append(byPOS[trainingPOS], csvWord{
			Lemma:           lemma,
			POS:             trainingPOS,
			PopularityCount: pop,
		})
	}

	for pos := range byPOS {
		sort.Slice(byPOS[pos], func(i, j int) bool {
			if byPOS[pos][i].PopularityCount == byPOS[pos][j].PopularityCount {
				return byPOS[pos][i].Lemma < byPOS[pos][j].Lemma
			}
			return byPOS[pos][i].PopularityCount > byPOS[pos][j].PopularityCount
		})
	}
	return byPOS, nil
}

func mapUDToTrainingPOS(udPOS string) (string, bool) {
	switch udPOS {
	case "NOUN":
		return "noun", true
	case "VERB", "AUX":
		return "verb", true
	case "ADJ":
		return "adjective", true
	case "ADV":
		return "adverb", true
	default:
		return "", false
	}
}

func isValidSpanishLemma(lemma string) bool {
	if lemma == "" || utf8.RuneCountInString(lemma) == 1 {
		return false
	}
	if wordsets.IsLemmaBlocked(lemma) {
		return false
	}
	if wordsets.IsVowellessAbbrevASCII(lemma) {
		return false
	}
	return spanishLemmaRe.MatchString(lemma)
}

