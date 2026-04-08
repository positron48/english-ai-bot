package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

type row struct {
	ID         int64
	Word       string
	POS        string
	NounGender string
	Opposite   string
}

func main() {
	var (
		limit  = flag.Int("limit", 0, "max rows to process; 0 means all")
		batch  = flag.Int("batch", 200, "batch size for DB reads")
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

	conn := db.GetConnection()
	lexicon, _, _ := models.LoadSpanishGenderLexiconDefault()

	var (
		lastID    int64
		processed int
		updated   int
	)

	for {
		rows, qErr := conn.Query(`
			SELECT id, word, COALESCE(pos, ''), COALESCE(noun_gender, ''), COALESCE(opposite_gender_word, '')
			FROM word_cards
			WHERE id > ?
			  AND COALESCE(TRIM(pos), '') <> ''
			ORDER BY id
			LIMIT ?
		`, lastID, *batch)
		if qErr != nil {
			log.Fatal("failed to query rows", zap.Error(qErr))
		}

		chunk := make([]row, 0, *batch)
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.ID, &r.Word, &r.POS, &r.NounGender, &r.Opposite); err != nil {
				_ = rows.Close()
				log.Fatal("failed to scan row", zap.Error(err))
			}
			chunk = append(chunk, r)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			log.Fatal("iteration error", zap.Error(err))
		}
		_ = rows.Close()

		if len(chunk) == 0 {
			break
		}

		for _, r := range chunk {
			if *limit > 0 && processed >= *limit {
				log.Info("limit reached", zap.Int("processed", processed), zap.Int("updated", updated))
				if *dryRun {
					log.Info("dry-run completed", zap.Int("processed", processed), zap.Int("updated_candidates", updated))
				} else {
					log.Info("normalization completed", zap.Int("processed", processed), zap.Int("updated", updated))
				}
				return
			}

			processed++
			lastID = r.ID

			rawPOS := strings.TrimSpace(r.POS)
			canonicalPOS := models.CanonicalWordPOS(rawPOS)
			if canonicalPOS == "" {
				continue
			}

			curGender := models.NormalizeNounGenderValue(r.NounGender)
			nextGender := curGender
			if models.IsNounPOS(rawPOS) && nextGender == "" {
				nextGender = models.InferNounGenderFromPOSText(rawPOS)
			}
			curOpp := strings.ToLower(strings.TrimSpace(r.Opposite))
			nextOpp := curOpp
			lemmaNorm := strings.ToLower(strings.TrimSpace(r.Word))
			if curOpp != "" && !models.IsLikelySimpleSpanishGenderPair(lemmaNorm, curOpp) {
				if entry, ok := lexicon[lemmaNorm]; !ok || entry.OppositeGenderWord != curOpp {
					nextOpp = ""
				}
			}

			changedPOS := strings.ToLower(rawPOS) != canonicalPOS
			changedGender := nextGender != "" && nextGender != curGender
			changedOpp := nextOpp != curOpp
			if !changedPOS && !changedGender && !changedOpp {
				continue
			}

			if *dryRun {
				updated++
				log.Info("dry-run pos normalization candidate",
					zap.Int64("word_card_id", r.ID),
					zap.String("word", r.Word),
					zap.String("pos_from", rawPOS),
					zap.String("pos_to", canonicalPOS),
					zap.String("noun_gender_from", curGender),
					zap.String("noun_gender_to", nextGender),
					zap.String("opposite_from", curOpp),
					zap.String("opposite_to", nextOpp),
				)
				continue
			}

			if changedGender || changedOpp {
				if _, err := conn.Exec(`
					UPDATE word_cards
					SET pos = ?, noun_gender = ?, opposite_gender_word = ?, updated_at = CURRENT_TIMESTAMP
					WHERE id = ?
				`, canonicalPOS, nextGender, nextOpp, r.ID); err != nil {
					log.Warn("failed to update pos+gender+opposite",
						zap.Int64("word_card_id", r.ID),
						zap.String("word", r.Word),
						zap.Error(err),
					)
					continue
				}
			} else {
				if _, err := conn.Exec(`
					UPDATE word_cards
					SET pos = ?, updated_at = CURRENT_TIMESTAMP
					WHERE id = ?
				`, canonicalPOS, r.ID); err != nil {
					log.Warn("failed to update pos",
						zap.Int64("word_card_id", r.ID),
						zap.String("word", r.Word),
						zap.Error(err),
					)
					continue
				}
			}
			updated++
		}

		log.Info("batch done", zap.Int("processed", processed), zap.Int("updated", updated))
		time.Sleep(100 * time.Millisecond)
	}

	if *dryRun {
		log.Info("dry-run completed", zap.Int("processed", processed), zap.Int("updated_candidates", updated))
	} else {
		log.Info("normalization completed", zap.Int("processed", processed), zap.Int("updated", updated))
	}
}
