package wordsetimport

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/resources/wordsets"

	"go.uber.org/zap"
)

type csvWord struct {
	Lemma           string
	POS             string
	PopularityCount int
}

type ImportOptions struct {
	CSVPath       string
	Lang          string
	Commit        bool
	OnlySetTitles map[string]struct{}
	LimitSets     int
}

type SetResult struct {
	SetID     int64
	Title     string
	POS       string
	RangeFrom int
	RangeTo   int
	Imported  int
	Skipped   bool
	Reason    string
}

type ImportResult struct {
	Mode      string
	Processed int
	Sets      []SetResult
}

var ranksRe = regexp.MustCompile(`(?i)(?:ranks?|rangos?)\s+(\d+)\s*[\p{Pd}-]\s*(\d+)`)
var spanishLemmaRe = regexp.MustCompile(`^[a-záéíóúüñ]+(?:-[a-záéíóúüñ]+)*$`)
var englishLemmaRe = regexp.MustCompile(`^[a-z]+(?:[-'][a-z]+)*$`)

func defaultCSVPathForLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "en":
		return "data/english_word_freq_pos_ud_top6000.filtered.csv"
	default:
		return "data/spanish_word_freq_pos_ud_top6000.csv"
	}
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "es"
	}
	return lang
}

func assertImportProfile(cfg *config.Config, lang, csvPath string) error {
	target := strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	appCode := strings.ToLower(strings.TrimSpace(cfg.Learning.AppCode))
	csvLower := strings.ToLower(csvPath)
	switch lang {
	case "en":
		if target != "en" || appCode != "english" {
			return fmt.Errorf("english import requires LEARNING_TARGET_LANG=en and LEARNING_APP_CODE=english (got target=%q app=%q)", target, appCode)
		}
		if strings.Contains(csvLower, "spanish") {
			return fmt.Errorf("english import refuses spanish csv path: %s", csvPath)
		}
	case "es":
		if target != "es" || appCode != "spanish" {
			return fmt.Errorf("spanish import requires LEARNING_TARGET_LANG=es and LEARNING_APP_CODE=spanish (got target=%q app=%q)", target, appCode)
		}
		if strings.Contains(csvLower, "english") {
			return fmt.Errorf("spanish import refuses english csv path: %s", csvPath)
		}
	default:
		return fmt.Errorf("unsupported lang %q (expected en or es)", lang)
	}
	return nil
}

func parseRankRange(title string) (start int, end int, ok bool) {
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
	if s, e, ok := parseRankRange(ws.Title); ok {
		return s, e, true
	}
	if ws.Description != nil {
		if s, e, ok := parseRankRange(*ws.Description); ok {
			return s, e, true
		}
	}
	return 0, 0, false
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

func isValidLemma(lang, lemma string) bool {
	if lemma == "" || utf8.RuneCountInString(lemma) == 1 {
		return false
	}
	if wordsets.IsLemmaBlockedForLang(lang, lemma) {
		return false
	}
	if wordsets.IsVowellessAbbrevASCII(lemma) {
		return false
	}
	switch lang {
	case "en":
		return englishLemmaRe.MatchString(lemma)
	default:
		return spanishLemmaRe.MatchString(lemma)
	}
}

func loadWordsByTrainingPOS(path, lang string) (map[string][]csvWord, error) {
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
		if !isValidLemma(lang, lemma) {
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

func Import(ctx context.Context, cfg *config.Config, conn *sql.DB, log *zap.Logger, opts ImportOptions) (*ImportResult, error) {
	lang := normalizeLang(opts.Lang)
	csvPath := strings.TrimSpace(opts.CSVPath)
	if csvPath == "" {
		csvPath = defaultCSVPathForLang(lang)
	}
	if err := assertImportProfile(cfg, lang, csvPath); err != nil {
		return nil, err
	}

	wordsByPOS, err := loadWordsByTrainingPOS(csvPath, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	wordSetRepo := repository.NewWordSetRepository(conn, log)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, log)
	wordRepo := repository.NewWordRepository(conn, log)
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
		return nil, fmt.Errorf("failed to list word sets: %w", err)
	}
	targetSets := selectRankedSets(sets, opts.OnlySetTitles)
	if opts.LimitSets > 0 && len(targetSets) > opts.LimitSets {
		targetSets = targetSets[:opts.LimitSets]
	}

	mode := "DRY-RUN"
	if opts.Commit {
		mode = "COMMIT"
	}
	out := &ImportResult{
		Mode: mode,
		Sets: make([]SetResult, 0, len(targetSets)),
	}
	if len(targetSets) == 0 {
		return out, nil
	}

	for _, ws := range targetSets {
		r := SetResult{SetID: ws.ID, Title: ws.Title}
		if ws.PreferredPOS == nil || strings.TrimSpace(*ws.PreferredPOS) == "" {
			r.Skipped = true
			r.Reason = "preferred_pos is empty"
			out.Sets = append(out.Sets, r)
			continue
		}
		pos := strings.ToLower(strings.TrimSpace(*ws.PreferredPOS))
		r.POS = pos
		rankStart, rankEnd, ok := extractRankRangeFromSet(ws)
		if !ok {
			r.Skipped = true
			r.Reason = "no rank range in title/description"
			out.Sets = append(out.Sets, r)
			continue
		}
		r.RangeFrom = rankStart
		r.RangeTo = rankEnd

		candidates := wordsByPOS[pos]
		if len(candidates) == 0 {
			r.Skipped = true
			r.Reason = "no words for preferred_pos"
			out.Sets = append(out.Sets, r)
			continue
		}
		if rankStart > len(candidates) {
			r.Skipped = true
			r.Reason = fmt.Sprintf("range starts at %d but only %d words in pos=%s", rankStart, len(candidates), pos)
			out.Sets = append(out.Sets, r)
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
		r.Imported = len(lemmas)

		if opts.Commit {
			if err := svc.ProcessWordSetItems(ctx, ws.ID, strings.Join(lemmas, ",")); err != nil {
				r.Skipped = true
				r.Reason = err.Error()
				out.Sets = append(out.Sets, r)
				continue
			}
		}
		out.Processed++
		out.Sets = append(out.Sets, r)
	}

	return out, nil
}
