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

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"tgbot-skeleton/internal/logger"
)

type setWord struct {
	WordCardID int64
	Word       string
}

func main() {
	os.Exit(runCLI())
}

func runCLI() int {
	commit := flag.Bool("commit", false, "apply changes to DB (default: dry-run)")
	onlySetsRaw := flag.String("only-set", "", "comma-separated exact set titles to process (optional)")
	limitSets := flag.Int("limit-sets", 0, "limit number of sets to process after filtering (0 = no limit)")
	limitWordsPerSet := flag.Int("limit-words-per-set", 0, "limit words per set (0 = no limit)")
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
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		return 1
	}
	defer db.Close()

	aiService := ai.NewServiceWithTimeoutAndSocks5Proxy(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, ai.ParseHTTPTimeout(cfg.AI.RequestTimeout), cfg.AI.Socks5Proxy, log)
	trainingPrompt, err := ai.LoadRenderedPromptFile(
		cfg.Training.PromptFile,
		cfg.Learning.NativeLang,
		cfg.Learning.TargetLang,
		cfg.Learning.Pair,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load training prompt file %q: %v\n", cfg.Training.PromptFile, err)
		return 1
	}
	aiService.SetTrainingPrompt(trainingPrompt)
	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), log)
	wordRepo := repository.NewWordRepository(db.GetConnection(), log)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), log)

	sets, err := wordSetRepo.ListWordSets(nil, 1000, 0, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list word sets: %v\n", err)
		return 1
	}
	selectedSetTitles := parseCSVSet(*onlySetsRaw)
	targetSets := selectSetsWithPreferredPOS(sets, selectedSetTitles)
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

	ctx := context.Background()
	var created, skippedNoBase, skippedHasPOS, skippedBadPOS, failed int

	for _, ws := range targetSets {
		if ws.PreferredPOS == nil || strings.TrimSpace(*ws.PreferredPOS) == "" {
			continue
		}
		requiredPOS := normalizePOS(*ws.PreferredPOS)
		if requiredPOS == "" {
			fmt.Printf("SKIP set_id=%d title=%q: unsupported preferred_pos=%q\n", ws.ID, ws.Title, *ws.PreferredPOS)
			continue
		}

		words, err := listSetWords(db.GetConnection(), ws.ID)
		if err != nil {
			fmt.Printf("ERROR set_id=%d title=%q: list words failed: %v\n", ws.ID, ws.Title, err)
			failed++
			continue
		}
		if *limitWordsPerSet > 0 && len(words) > *limitWordsPerSet {
			words = words[:*limitWordsPerSet]
		}
		fmt.Printf("SET id=%d title=%q preferred_pos=%s words=%d\n", ws.ID, ws.Title, requiredPOS, len(words))

		for _, w := range words {
			existingCards, err := trainingCardRepo.GetTrainingCardsByWordCardID(w.WordCardID)
			if err != nil {
				fmt.Printf("  FAIL word=%q id=%d: get training cards: %v\n", w.Word, w.WordCardID, err)
				failed++
				continue
			}

			// Ensure ordinary generation already happened for this word.
			if len(existingCards) == 0 {
				skippedNoBase++
				continue
			}
			if hasTrainingPOS(existingCards, requiredPOS) {
				skippedHasPOS++
				continue
			}

			wordCard, err := wordRepo.GetWordCardByID(w.WordCardID)
			if err != nil || wordCard == nil {
				fmt.Printf("  FAIL word=%q id=%d: get word card: %v\n", w.Word, w.WordCardID, err)
				failed++
				continue
			}

			resp, err := generateOnePOSCard(ctx, aiService, wordCard, requiredPOS, cfg.Learning.TargetLang, cfg.AI.ModelHigh)
			if err != nil {
				fmt.Printf("  FAIL word=%q id=%d pos=%s: %v\n", w.Word, w.WordCardID, requiredPOS, err)
				failed++
				continue
			}
			if len(resp.Senses) == 0 {
				fmt.Printf("  FAIL word=%q id=%d pos=%s: empty senses after generation\n", w.Word, w.WordCardID, requiredPOS)
				failed++
				continue
			}
			sense := resp.Senses[0]
			genPOS := normalizePOS(sense.POS)
			if genPOS != requiredPOS {
				skippedBadPOS++
				continue
			}

			if *commit {
				card, err := buildTrainingCardFromSense(wordCard.ID, existingCards, &resp, &sense)
				if err != nil {
					fmt.Printf("  FAIL word=%q id=%d: build card: %v\n", w.Word, w.WordCardID, err)
					failed++
					continue
				}
				if _, err := trainingCardRepo.CreateTrainingCard(card); err != nil {
					fmt.Printf("  FAIL word=%q id=%d: create training card: %v\n", w.Word, w.WordCardID, err)
					failed++
					continue
				}
			}
			created++
			fmt.Printf("  OK word=%q id=%d created_pos=%s\n", w.Word, w.WordCardID, requiredPOS)
		}
	}

	fmt.Printf("Done. mode=%s created=%d skipped_no_base=%d skipped_has_pos=%d skipped_bad_pos=%d failed=%d\n",
		mode, created, skippedNoBase, skippedHasPOS, skippedBadPOS, failed)
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

func selectSetsWithPreferredPOS(all []*models.WordSet, only map[string]struct{}) []*models.WordSet {
	out := make([]*models.WordSet, 0, len(all))
	for _, ws := range all {
		if ws == nil || ws.PreferredPOS == nil || strings.TrimSpace(*ws.PreferredPOS) == "" {
			continue
		}
		if normalizePOS(*ws.PreferredPOS) == "" {
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

func listSetWords(conn *sql.DB, setID int64) ([]setWord, error) {
	rows, err := conn.Query(`
		SELECT wc.id, wc.word
		FROM word_set_items wsi
		INNER JOIN word_cards wc ON wc.id = wsi.word_card_id
		WHERE wsi.word_set_id = ?
		ORDER BY wsi.sort_order, wc.word
	`, setID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []setWord
	for rows.Next() {
		var w setWord
		if err := rows.Scan(&w.WordCardID, &w.Word); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func hasTrainingPOS(cards []*models.TrainingCard, requiredPOS string) bool {
	for _, c := range cards {
		if c == nil || c.POS == nil {
			continue
		}
		if normalizePOS(*c.POS) == requiredPOS {
			return true
		}
	}
	return false
}

func normalizePOS(pos string) string {
	p := strings.ToLower(strings.TrimSpace(pos))
	switch p {
	case "noun", "proper noun":
		return "noun"
	case "verb", "aux", "auxiliary", "auxiliary verb":
		return "verb"
	case "adjective", "adj":
		return "adjective"
	case "adverb", "adv":
		return "adverb"
	default:
		return ""
	}
}

func generateOnePOSCard(
	ctx context.Context,
	aiService *ai.Service,
	wordCard *models.WordCard,
	requiredPOS string,
	targetLang string,
	modelHigh string,
) (models.TrainingCardResponse, error) {
	lemma := strings.TrimSpace(wordCard.Word)
	if lemma == "" {
		return models.TrainingCardResponse{}, fmt.Errorf("empty word lemma")
	}
	constraints := fmt.Sprintf(
		"Target POS: %s.\nReturn exactly one sense.\nThe generated sense.pos must be exactly %q.\nIf this POS is not valid for this word, return JSON with non-empty error field and no fabricated sense.",
		requiredPOS,
		requiredPOS,
	)

	tryGenerate := func(modelOverride ...string) (models.TrainingCardResponse, error) {
		raw, err := aiService.GenerateAdditionalTrainingCard(ctx, lemma, constraints, modelOverride...)
		if err != nil {
			return models.TrainingCardResponse{}, err
		}
		var resp models.TrainingCardResponse
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			return models.TrainingCardResponse{}, fmt.Errorf("parse llm response: %w", err)
		}
		models.SyncTrainingCardResponseNeutralAliases(&resp)
		if strings.TrimSpace(resp.Error) != "" {
			return models.TrainingCardResponse{}, fmt.Errorf("llm error: %s", strings.TrimSpace(resp.Error))
		}
		if len(resp.Senses) == 0 {
			return models.TrainingCardResponse{}, fmt.Errorf("llm returned empty senses")
		}
		resp.Senses = resp.Senses[:1]
		resp.Senses[0].POS = requiredPOS
		validationErr := service.ValidateTrainingCardResponse(targetLang, wordCard, &resp)
		if validationErr != "" {
			return models.TrainingCardResponse{}, fmt.Errorf("validation failed: %s", validationErr)
		}
		return resp, nil
	}

	resp, err := tryGenerate()
	if err == nil {
		return resp, nil
	}
	if modelHigh != "" {
		return tryGenerate(modelHigh)
	}
	return models.TrainingCardResponse{}, err
}

func buildTrainingCardFromSense(wordCardID int64, existing []*models.TrainingCard, resp *models.TrainingCardResponse, sense *models.TrainingCardSense) (*models.TrainingCard, error) {
	maxSense := -1
	for _, c := range existing {
		if c != nil && c.SenseIndex > maxSense {
			maxSense = c.SenseIndex
		}
	}
	nextSense := maxSense + 1

	dRU, err := json.Marshal(sense.DistractorsRU)
	if err != nil {
		return nil, err
	}
	dEN, err := json.Marshal(sense.DistractorsEN)
	if err != nil {
		return nil, err
	}

	displayWord := strings.TrimSpace(sense.DisplayWord)
	if displayWord == "" {
		displayWord = strings.TrimSpace(resp.WordEN)
	}
	pos := normalizePOS(sense.POS)
	if pos == "" {
		pos = "noun"
	}

	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        displayWord,
		WordTarget:    displayWord,
		Transcription: resp.Transcription,
		SenseIndex:    nextSense,
		WordRU:        sense.WordRU,
		WordNative:    sense.WordNative,
		MeaningEN:     sense.MeaningEN,
		MeaningTarget: sense.MeaningTarget,
		ExampleEN:     sense.ExampleEN,
		ExampleTarget: sense.ExampleTarget,
		ExampleRU:     sense.ExampleRU,
		ExampleNative: sense.ExampleNative,
		DistractorsRU: string(dRU),
		DistractorsEN: string(dEN),
		Hint:          sense.Hint,
		POS:           &pos,
		DisplayWord:   &displayWord,
	}
	return card, nil
}
