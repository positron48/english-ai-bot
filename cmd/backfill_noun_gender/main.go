package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

type wordRow struct {
	ID         int64
	Word       string
	Definition string
	POS        string
}

type nounMorph struct {
	Gender             string
	OppositeGenderWord string
}

func isNounPOS(pos string) bool {
	return models.IsNounPOS(pos)
}

func normalizeGender(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "m", "f", "mf", "n":
		return s
	default:
		return ""
	}
}

func normalizeOpposite(v, word string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return ""
	}
	w := strings.ToLower(strings.TrimSpace(word))
	if s == w {
		return ""
	}
	if !models.IsLikelySimpleSpanishGenderPair(w, s) {
		return ""
	}
	return s
}

func detectNounMorph(ctx context.Context, svc *ai.Service, targetLang string, lex map[string]models.SpanishGenderLexiconEntry, row wordRow) (*nounMorph, error) {
	lang := strings.ToLower(strings.TrimSpace(targetLang))
	if lang == "en" {
		return &nounMorph{Gender: "n"}, nil
	}
	if lang == "es" && len(lex) > 0 {
		if entry, ok := lex[strings.ToLower(strings.TrimSpace(row.Word))]; ok {
			return &nounMorph{Gender: entry.Gender, OppositeGenderWord: entry.OppositeGenderWord}, nil
		}
	}

	// Spanish nouns: ask model to return two tokens separated by "|": gender|opposite_word.
	prompt := fmt.Sprintf(
		"Classify grammatical morphology for one Spanish noun. "+
			"Return exactly one line in format gender|opposite_word where gender is one of m,f,mf,n. "+
			"For nouns without natural opposite form return empty opposite_word. "+
			"If uncertain about gender return n. Word: %q. Optional definition in Russian: %q.",
		row.Word, row.Definition,
	)
	resp, err := svc.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(strings.TrimSpace(resp), "|", 2)
	gender := ""
	opp := ""
	if len(parts) > 0 {
		gender = normalizeGender(parts[0])
	}
	if len(parts) == 2 {
		opp = normalizeOpposite(parts[1], row.Word)
	}
	if gender == "" {
		gender = "n"
	}
	return &nounMorph{Gender: gender, OppositeGenderWord: opp}, nil
}

func main() {
	var (
		limit  = flag.Int("limit", 0, "max nouns to process; 0 means all")
		batch  = flag.Int("batch", 100, "batch size for DB reads")
		dryRun = flag.Bool("dry-run", true, "preview only; do not write to database")
	)
	flag.Parse()

	if *batch <= 0 {
		fmt.Fprintln(os.Stderr, "batch must be > 0")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	aiService := ai.NewServiceWithTimeoutAndSocks5Proxy(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, "You are a strict morphology classifier.", ai.ParseHTTPTimeout(cfg.AI.RequestTimeout), cfg.AI.Socks5Proxy, log)
	lexicon, lexPath, lexErr := models.LoadSpanishGenderLexiconDefault()
	if lexErr != nil {
		log.Warn("spanish gender lexicon unavailable; fallback to LLM only", zap.Error(lexErr))
	} else {
		log.Info("spanish gender lexicon loaded", zap.String("path", lexPath), zap.Int("entries", len(lexicon)))
	}

	conn := db.GetConnection()
	ctx := context.Background()

	var (
		lastID    int64
		processed int
		updated   int
		failed    int
	)

	for {
		rows, queryErr := conn.Query(`
			SELECT id, word, COALESCE(definition_ru, ''), COALESCE(pos, '')
			FROM word_cards
			WHERE id > ?
			  AND (
			    noun_gender IS NULL OR LOWER(noun_gender) NOT IN ('m','f','mf','n')
			    OR COALESCE(TRIM(opposite_gender_word), '') = ''
			  )
			ORDER BY id
			LIMIT ?
		`, lastID, *batch)
		if queryErr != nil {
			log.Fatal("failed to query noun rows", zap.Error(queryErr))
		}

		chunk := make([]wordRow, 0, *batch)
		for rows.Next() {
			var r wordRow
			if err := rows.Scan(&r.ID, &r.Word, &r.Definition, &r.POS); err != nil {
				_ = rows.Close()
				log.Fatal("failed to scan noun row", zap.Error(err))
			}
			chunk = append(chunk, r)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			log.Fatal("failed while iterating noun rows", zap.Error(err))
		}
		_ = rows.Close()

		if len(chunk) == 0 {
			break
		}

		for _, r := range chunk {
			if *limit > 0 && processed >= *limit {
				log.Info("limit reached", zap.Int("processed", processed), zap.Int("updated", updated), zap.Int("failed", failed))
				if *dryRun {
					log.Info("dry-run completed", zap.Int("processed", processed), zap.Int("updated_candidates", updated), zap.Int("failed", failed))
				} else {
					log.Info("backfill completed", zap.Int("processed", processed), zap.Int("updated", updated), zap.Int("failed", failed))
				}
				return
			}

			processed++
			lastID = r.ID

			// For non-noun POS do NOT use AI: only lexicon lookup is allowed.
			if !isNounPOS(r.POS) {
				if strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang)) != "es" || len(lexicon) == 0 {
					continue
				}
				entry, ok := lexicon[strings.ToLower(strings.TrimSpace(r.Word))]
				if !ok {
					continue
				}
				morph := &nounMorph{
					Gender:             normalizeGender(entry.Gender),
					OppositeGenderWord: normalizeOpposite(entry.OppositeGenderWord, r.Word),
				}
				if morph.Gender == "" {
					continue
				}
				if *dryRun {
					updated++
					log.Info("dry-run lexicon morphology candidate (non-noun POS)",
						zap.Int64("word_card_id", r.ID),
						zap.String("word", r.Word),
						zap.String("pos", r.POS),
						zap.String("noun_gender", morph.Gender),
						zap.String("opposite_gender_word", morph.OppositeGenderWord),
					)
					continue
				}
				res, execErr := conn.Exec(`
					UPDATE word_cards
					SET noun_gender = ?, opposite_gender_word = ?, updated_at = CURRENT_TIMESTAMP
					WHERE id = ?
				`, morph.Gender, morph.OppositeGenderWord, r.ID)
				if execErr != nil {
					failed++
					log.Warn("failed to update lexicon morphology (non-noun POS)",
						zap.Int64("word_card_id", r.ID),
						zap.String("word", r.Word),
						zap.String("pos", r.POS),
						zap.Error(execErr),
					)
					continue
				}
				affected, _ := res.RowsAffected()
				if affected > 0 {
					updated++
				}
				continue
			}

			morph, detectErr := detectNounMorph(ctx, aiService, cfg.Learning.TargetLang, lexicon, r)
			if detectErr != nil {
				failed++
				log.Warn("failed to detect noun morphology",
					zap.Int64("word_card_id", r.ID),
					zap.String("word", r.Word),
					zap.Error(detectErr),
				)
				continue
			}

			if *dryRun {
				updated++
				log.Info("dry-run noun morphology candidate",
					zap.Int64("word_card_id", r.ID),
					zap.String("word", r.Word),
					zap.String("noun_gender", morph.Gender),
					zap.String("opposite_gender_word", morph.OppositeGenderWord),
				)
				continue
			}

			res, execErr := conn.Exec(`
				UPDATE word_cards
				SET noun_gender = ?, opposite_gender_word = ?, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, morph.Gender, morph.OppositeGenderWord, r.ID)
			if execErr != nil {
				failed++
				log.Warn("failed to update noun morphology",
					zap.Int64("word_card_id", r.ID),
					zap.String("word", r.Word),
					zap.Error(execErr),
				)
				continue
			}
			affected, _ := res.RowsAffected()
			if affected > 0 {
				updated++
			}
		}

		log.Info("batch done",
			zap.Int("processed", processed),
			zap.Int("updated", updated),
			zap.Int("failed", failed),
		)

		// Small pause to reduce provider burst pressure in long runs.
		time.Sleep(100 * time.Millisecond)
	}

	if *dryRun {
		log.Info("dry-run completed", zap.Int("processed", processed), zap.Int("updated_candidates", updated), zap.Int("failed", failed))
	} else {
		log.Info("backfill completed", zap.Int("processed", processed), zap.Int("updated", updated), zap.Int("failed", failed))
	}
}
