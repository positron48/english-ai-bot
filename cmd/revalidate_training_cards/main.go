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

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/service"
)

type row struct {
	WordCardID    int64
	Lemma         string
	POS           sql.NullString
	DefinitionRU  sql.NullString
	SenseIndex    int
	WordEN        string
	Transcription sql.NullString
	WordRU        string
	MeaningEN     string
	ExampleEN     sql.NullString
	ExampleRU     sql.NullString
	DistractorsEN sql.NullString
	DistractorsRU sql.NullString
	Hint          sql.NullString
	DisplayWord   sql.NullString
}

type group struct {
	wordCard models.WordCard
	resp     models.TrainingCardResponse
}

type invalidGroup struct {
	g      group
	reason string
}

func main() {
	os.Exit(runCLI())
}

func runCLI() int {
	commit := flag.Bool("commit", false, "apply DB changes (default: dry-run)")
	limit := flag.Int("limit", 0, "max number of word_cards to requeue (0 = no limit)")
	onlyWord := flag.String("only-word", "", "process only this lemma (optional)")
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
		fmt.Fprintf(os.Stderr, "failed to init db: %v\n", err)
		return 1
	}
	defer db.Close()

	groups, err := loadGroups(db.GetConnection(), *onlyWord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load cards: %v\n", err)
		return 1
	}

	invalid := make([]invalidGroup, 0)
	for _, g := range groups {
		if !isNativeDefinitionValid(g.wordCard, cfg.Learning) {
			invalid = append(invalid, invalidGroup{
				g:      g,
				reason: "definition_ru_not_cyrillic",
			})
			continue
		}
		if errMsg := service.ValidateTrainingCardResponse(cfg.Learning.TargetLang, &g.wordCard, &g.resp); errMsg != "" {
			invalid = append(invalid, invalidGroup{
				g:      g,
				reason: errMsg,
			})
			continue
		}
		if dupErr := duplicateSenseError(g.resp); dupErr != "" {
			invalid = append(invalid, invalidGroup{
				g:      g,
				reason: dupErr,
			})
		}
	}

	sort.Slice(invalid, func(i, j int) bool {
		return invalid[i].g.wordCard.ID < invalid[j].g.wordCard.ID
	})
	if *limit > 0 && len(invalid) > *limit {
		invalid = invalid[:*limit]
	}

	mode := "DRY-RUN"
	if *commit {
		mode = "COMMIT"
	}
	fmt.Printf("[%s] total_word_cards=%d invalid=%d\n", mode, len(groups), len(invalid))
	for _, g := range invalid {
		fmt.Printf("INVALID word_card_id=%d lemma=%q senses=%d reason=%q\n", g.g.wordCard.ID, g.g.wordCard.Word, len(g.g.resp.Senses), g.reason)
	}

	if !*commit || len(invalid) == 0 {
		return 0
	}

	requeued := 0
	for _, g := range invalid {
		if err := requeueWordCard(ctx, db.GetConnection(), g.g.wordCard.ID, shouldNullDefinitionRU(g.g.wordCard, cfg.Learning)); err != nil {
			fmt.Printf("ERROR requeue word_card_id=%d lemma=%q: %v\n", g.g.wordCard.ID, g.g.wordCard.Word, err)
			continue
		}
		requeued++
	}
	fmt.Printf("Done. requeued=%d\n", requeued)
	return 0
}

func loadGroups(conn *sql.DB, onlyWord string) ([]group, error) {
	query := `SELECT
		wc.id, wc.word, wc.pos, wc.definition_ru,
		tc.sense_index, tc.word_en, tc.transcription, tc.word_ru, tc.meaning_en,
		tc.example_en, tc.example_ru, tc.distractors_en, tc.distractors_ru, tc.hint, tc.display_word
	FROM word_cards wc
	INNER JOIN training_cards tc ON tc.word_card_id = wc.id`
	args := []any{}
	if strings.TrimSpace(onlyWord) != "" {
		query += ` WHERE LOWER(wc.word) = LOWER(?)`
		args = append(args, strings.TrimSpace(onlyWord))
	}
	query += ` ORDER BY wc.id, tc.sense_index`

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[int64]*group{}
	for rows.Next() {
		var r row
		if err := rows.Scan(
			&r.WordCardID, &r.Lemma, &r.POS, &r.DefinitionRU,
			&r.SenseIndex, &r.WordEN, &r.Transcription, &r.WordRU, &r.MeaningEN,
			&r.ExampleEN, &r.ExampleRU, &r.DistractorsEN, &r.DistractorsRU, &r.Hint, &r.DisplayWord,
		); err != nil {
			return nil, err
		}

		g, ok := byID[r.WordCardID]
		if !ok {
			wc := models.WordCard{
				ID:   r.WordCardID,
				Word: r.Lemma,
			}
			if r.POS.Valid && strings.TrimSpace(r.POS.String) != "" {
				pos := strings.TrimSpace(r.POS.String)
				wc.POS = &pos
			}
			if r.DefinitionRU.Valid && strings.TrimSpace(r.DefinitionRU.String) != "" {
				defRU := strings.TrimSpace(r.DefinitionRU.String)
				wc.DefinitionRU = &defRU
			}
			gr := &group{
				wordCard: wc,
				resp: models.TrainingCardResponse{
					WordEN:        r.Lemma,
					WordTarget:    r.Lemma,
					Lemma:         r.Lemma,
					Transcription: strings.TrimSpace(r.Transcription.String),
					Senses:        []models.TrainingCardSense{},
				},
			}
			byID[r.WordCardID] = gr
			g = gr
		}

		var de, dr []string
		if strings.TrimSpace(r.DistractorsEN.String) != "" {
			_ = json.Unmarshal([]byte(r.DistractorsEN.String), &de)
		}
		if strings.TrimSpace(r.DistractorsRU.String) != "" {
			_ = json.Unmarshal([]byte(r.DistractorsRU.String), &dr)
		}
		sense := models.TrainingCardSense{
			POS:           strings.TrimSpace(r.POS.String),
			DisplayWord:   strings.TrimSpace(r.DisplayWord.String),
			WordRU:        strings.TrimSpace(r.WordRU),
			MeaningEN:     strings.TrimSpace(r.MeaningEN),
			ExampleEN:     strings.TrimSpace(r.ExampleEN.String),
			ExampleRU:     strings.TrimSpace(r.ExampleRU.String),
			DistractorsEN: de,
			DistractorsRU: dr,
			Hint:          strings.TrimSpace(r.Hint.String),
		}
		g.resp.Senses = append(g.resp.Senses, sense)
	}

	out := make([]group, 0, len(byID))
	for _, g := range byID {
		out = append(out, *g)
	}
	return out, rows.Err()
}

func isNativeDefinitionValid(card models.WordCard, learning config.LearningConfig) bool {
	if !strings.EqualFold(learning.NativeLang, "ru") || !strings.EqualFold(learning.TargetLang, "es") {
		return true
	}
	if card.DefinitionRU == nil {
		return false
	}
	return service.ContainsCyrillic(strings.TrimSpace(*card.DefinitionRU))
}

func shouldNullDefinitionRU(card models.WordCard, learning config.LearningConfig) bool {
	if !strings.EqualFold(learning.NativeLang, "ru") || !strings.EqualFold(learning.TargetLang, "es") {
		return false
	}
	if card.DefinitionRU == nil {
		return false
	}
	return !service.ContainsCyrillic(strings.TrimSpace(*card.DefinitionRU))
}

func requeueWordCard(ctx context.Context, conn *sql.DB, wordCardID int64, nullDefinitionRU bool) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM training_cards WHERE word_card_id = ?`, wordCardID); err != nil {
		return err
	}
	if nullDefinitionRU {
		if _, err := tx.ExecContext(ctx, `UPDATE word_cards
			SET definition_ru = NULL, processed_at = NULL, processing_error = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, wordCardID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE word_cards
			SET processed_at = NULL, processing_error = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, wordCardID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func duplicateSenseError(resp models.TrainingCardResponse) string {
	seen := make(map[string]int, len(resp.Senses))
	for i, s := range resp.Senses {
		key := strings.Join([]string{
			normalizeDupField(s.POS),
			normalizeDupField(s.DisplayWord),
			normalizeDupField(s.WordRU),
			normalizeDupField(s.MeaningEN),
		}, "|")
		if prevIdx, ok := seen[key]; ok {
			return fmt.Sprintf("duplicate_training_sense sense=%d duplicates sense=%d", i, prevIdx)
		}
		seen[key] = i
	}
	return ""
}

func normalizeDupField(v string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
}

