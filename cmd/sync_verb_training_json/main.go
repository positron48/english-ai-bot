package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/verbtraining"
)

type packIndex struct {
	Version string            `json:"version"`
	Language string           `json:"language"`
	Lemmas map[string]string  `json:"lemmas"`
}

type syncStats struct {
	LemmasTotal int
	FormsUpserted int
	CardsUpserted int
	FormsDeleted int
	CardsDeleted int
	LemmasDeleted int
}

func main() {
	os.Exit(run())
}

func run() int {
	var (
		courseRoot = flag.String("course-root", "courses/spanish-grammar", "path to spanish grammar course root")
		dryRun = flag.Bool("dry-run", false, "validate and report only, no DB changes")
	)
	flag.Parse()

	root := filepath.Clean(*courseRoot)
	artifactRoot := filepath.Join(root, "training_pack", "verb_forms")
	indexPath := filepath.Join(artifactRoot, "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		fallbackRoot := filepath.Join("internal", "grammartrainingpack", "es", "verb_forms")
		fallbackIndex := filepath.Join(fallbackRoot, "index.json")
		if _, ferr := os.Stat(fallbackIndex); ferr == nil {
			artifactRoot = fallbackRoot
			indexPath = fallbackIndex
		} else {
			fmt.Fprintf(os.Stderr, "index.json not found at %s and fallback %s\n", indexPath, fallbackIndex)
			return 1
		}
	}
	indexRaw, err := os.ReadFile(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read index: %v\n", err)
		return 1
	}
	var idx packIndex
	if err := json.Unmarshal(indexRaw, &idx); err != nil {
		fmt.Fprintf(os.Stderr, "parse index: %v\n", err)
		return 1
	}
	if strings.ToLower(strings.TrimSpace(idx.Language)) != "es" {
		fmt.Fprintf(os.Stderr, "unsupported language in index: %q\n", idx.Language)
		return 1
	}
	if len(idx.Lemmas) == 0 {
		fmt.Println("index.lemmas is empty; nothing to sync")
		return 0
	}
	lemmas := make([]string, 0, len(idx.Lemmas))
	for lemma := range idx.Lemmas {
		lemmas = append(lemmas, strings.ToLower(strings.TrimSpace(lemma)))
	}
	sort.Strings(lemmas)

	artifacts := make([]verbtraining.LemmaArtifact, 0, len(lemmas))
	for _, lemma := range lemmas {
		rel := strings.TrimSpace(idx.Lemmas[lemma])
		if rel == "" {
			fmt.Fprintf(os.Stderr, "empty path for lemma %s\n", lemma)
			return 1
		}
		path := filepath.Join(artifactRoot, filepath.FromSlash(rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read lemma file %s: %v\n", path, err)
			return 1
		}
		var a verbtraining.LemmaArtifact
		if err := json.Unmarshal(raw, &a); err != nil {
			fmt.Fprintf(os.Stderr, "parse lemma file %s: %v\n", path, err)
			return 1
		}
		if err := a.ValidateStrictCoverage(); err != nil {
			fmt.Fprintf(os.Stderr, "validate lemma %s: %v\n", lemma, err)
			return 1
		}
		if a.Lemma != lemma {
			fmt.Fprintf(os.Stderr, "lemma mismatch in file %s: index=%s file=%s\n", path, lemma, a.Lemma)
			return 1
		}
		artifacts = append(artifacts, a)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		return 1
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		return 1
	}
	defer db.Close()

	stats, err := syncArtifacts(db.GetConnection(), artifacts, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sync failed: %v\n", err)
		return 1
	}
	fmt.Printf("sync_verb_training_json: lemmas=%d forms_upserted=%d cards_upserted=%d forms_deleted=%d cards_deleted=%d lemmas_deleted=%d dry_run=%v\n",
		stats.LemmasTotal, stats.FormsUpserted, stats.CardsUpserted, stats.FormsDeleted, stats.CardsDeleted, stats.LemmasDeleted, *dryRun)
	return 0
}

func syncArtifacts(db *sql.DB, artifacts []verbtraining.LemmaArtifact, dryRun bool) (*syncStats, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	stats := &syncStats{LemmasTotal: len(artifacts)}
	seenLemma := make(map[string]struct{}, len(artifacts))
	for _, a := range artifacts {
		seenLemma[a.Lemma] = struct{}{}
		lemmaID, err := upsertLemma(tx, a.Lemma, a.WordCardID)
		if err != nil {
			return nil, fmt.Errorf("upsert lemma %s: %w", a.Lemma, err)
		}
		if err := linkWordCard(tx, a.WordCardID, lemmaID); err != nil {
			return nil, fmt.Errorf("link word-card %d -> lemma %s: %w", a.WordCardID, a.Lemma, err)
		}
		expectedFormIDs := make(map[int64]struct{}, len(a.Cards))
		expectedCardIDs := make(map[int64]struct{}, len(a.Cards))
		for _, card := range a.Cards {
			tense, mood, ok := verbtraining.ParseScope(card.Scope)
			if !ok {
				return nil, fmt.Errorf("invalid scope %s for lemma %s", card.Scope, a.Lemma)
			}
			formID, err := upsertForm(tx, lemmaID, mood, tense, card.Person, card.Number, card.SurfaceForm)
			if err != nil {
				return nil, fmt.Errorf("upsert form %s %s: %w", a.Lemma, card.Scope, err)
			}
			expectedFormIDs[formID] = struct{}{}
			stats.FormsUpserted++

			promptJSON, err := verbtraining.EncodePromptJSON(a.Lemma, card)
			if err != nil {
				return nil, fmt.Errorf("encode prompt %s %s: %w", a.Lemma, card.Scope, err)
			}
			answerJSON, _ := json.Marshal(map[string]string{"surface_form": card.SurfaceForm})
			optionsJSON, _ := json.Marshal(card.Options)
			cardID, err := upsertTrainingCard(tx, a.WordCardID, formID, promptJSON, string(answerJSON), string(optionsJSON))
			if err != nil {
				return nil, fmt.Errorf("upsert training card %s %s: %w", a.Lemma, card.Scope, err)
			}
			expectedCardIDs[cardID] = struct{}{}
			stats.CardsUpserted++
		}
		nf, err := pruneFormsForLemma(tx, lemmaID, expectedFormIDs)
		if err != nil {
			return nil, fmt.Errorf("prune forms %s: %w", a.Lemma, err)
		}
		stats.FormsDeleted += nf
		nc, err := pruneCardsForWord(tx, a.WordCardID, expectedCardIDs)
		if err != nil {
			return nil, fmt.Errorf("prune cards %s: %w", a.Lemma, err)
		}
		stats.CardsDeleted += nc
	}
	nd, err := pruneLemmasMissingFromJSON(tx, seenLemma)
	if err != nil {
		return nil, fmt.Errorf("prune removed lemmas: %w", err)
	}
	stats.LemmasDeleted = nd

	if dryRun {
		return stats, nil
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return stats, nil
}

func upsertLemma(tx *sql.Tx, lemma string, wordCardID int64) (int64, error) {
	meta, _ := json.Marshal(map[string]interface{}{
		"source": "verb_forms_json",
		"word_card_id": wordCardID,
	})
	q := `INSERT INTO verb_lemmas (lemma, language, source, source_version, checksum, metadata_json, updated_at)
	      VALUES (?, 'es', ?, 'v1', '', ?, CURRENT_TIMESTAMP)
	      ON CONFLICT(lemma, language) DO UPDATE SET
	        source=excluded.source,
	        source_version=excluded.source_version,
	        metadata_json=excluded.metadata_json,
	        updated_at=CURRENT_TIMESTAMP`
	if _, err := tx.Exec(q, lemma, verbtraining.CardSourceLLMJSON, string(meta)); err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM verb_lemmas WHERE lemma=? AND language='es'`, lemma).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func linkWordCard(tx *sql.Tx, wordCardID, lemmaID int64) error {
	_, err := tx.Exec(`INSERT INTO word_verb_lemmas (word_card_id, verb_lemma_id, confidence, source, updated_at)
		VALUES (?, ?, 1.0, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(word_card_id) DO UPDATE SET
		  verb_lemma_id=excluded.verb_lemma_id,
		  confidence=excluded.confidence,
		  source=excluded.source,
		  updated_at=CURRENT_TIMESTAMP`,
		wordCardID, lemmaID, verbtraining.CardSourceLLMJSON)
	return err
}

func upsertForm(tx *sql.Tx, lemmaID int64, mood, tense, person, number, surface string) (int64, error) {
	_, err := tx.Exec(`INSERT INTO verb_forms_dict (verb_lemma_id, mood, tense, person, number, surface_form, is_irregular, tags_json, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, '{}', CURRENT_TIMESTAMP)
		ON CONFLICT(verb_lemma_id, mood, tense, person, number) DO UPDATE SET
		  surface_form=excluded.surface_form,
		  is_irregular=excluded.is_irregular,
		  tags_json=excluded.tags_json,
		  updated_at=CURRENT_TIMESTAMP`,
		lemmaID, mood, tense, person, number, surface)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM verb_forms_dict WHERE verb_lemma_id=? AND mood=? AND tense=? AND person=? AND number=?`,
		lemmaID, mood, tense, person, number).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func upsertTrainingCard(tx *sql.Tx, wordCardID, formID int64, promptJSON, answerJSON, optionsJSON string) (int64, error) {
	_, err := tx.Exec(`INSERT INTO verb_training_cards (word_card_id, verb_form_dict_id, card_type, prompt_json, answer_json, distractors_json, example_id, updated_at)
		VALUES (?, ?, 'cloze_form', ?, ?, ?, NULL, CURRENT_TIMESTAMP)
		ON CONFLICT(word_card_id, verb_form_dict_id, card_type) DO UPDATE SET
		  prompt_json=excluded.prompt_json,
		  answer_json=excluded.answer_json,
		  distractors_json=excluded.distractors_json,
		  updated_at=CURRENT_TIMESTAMP`,
		wordCardID, formID, promptJSON, answerJSON, optionsJSON)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM verb_training_cards WHERE word_card_id=? AND verb_form_dict_id=? AND card_type='cloze_form'`,
		wordCardID, formID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func pruneFormsForLemma(tx *sql.Tx, lemmaID int64, expected map[int64]struct{}) (int, error) {
	rows, err := tx.Query(`SELECT id FROM verb_forms_dict WHERE verb_lemma_id=?`, lemmaID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var del []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if _, ok := expected[id]; !ok {
			del = append(del, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range del {
		if _, err := tx.Exec(`DELETE FROM verb_forms_dict WHERE id=?`, id); err != nil {
			return 0, err
		}
	}
	return len(del), nil
}

func pruneCardsForWord(tx *sql.Tx, wordCardID int64, expected map[int64]struct{}) (int, error) {
	rows, err := tx.Query(`SELECT id FROM verb_training_cards WHERE word_card_id=? AND card_type='cloze_form'`, wordCardID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var del []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if _, ok := expected[id]; !ok {
			del = append(del, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range del {
		if _, err := tx.Exec(`DELETE FROM verb_training_cards WHERE id=?`, id); err != nil {
			return 0, err
		}
	}
	return len(del), nil
}

func pruneLemmasMissingFromJSON(tx *sql.Tx, expected map[string]struct{}) (int, error) {
	rows, err := tx.Query(`SELECT id, lemma FROM verb_lemmas WHERE language='es' AND source=?`, verbtraining.CardSourceLLMJSON)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var del []int64
	for rows.Next() {
		var id int64
		var lemma string
		if err := rows.Scan(&id, &lemma); err != nil {
			return 0, err
		}
		lemma = strings.ToLower(strings.TrimSpace(lemma))
		if _, ok := expected[lemma]; !ok {
			del = append(del, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range del {
		if _, err := tx.Exec(`DELETE FROM verb_lemmas WHERE id=?`, id); err != nil {
			return 0, err
		}
	}
	return len(del), nil
}

