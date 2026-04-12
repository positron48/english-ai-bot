package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		inputPath     = flag.String("input", "", "path to input file (JSON array of lemmas or Jehle CSV)")
		source        = flag.String("source", "custom", "dataset source name")
		sourceVersion = flag.String("source-version", "v1", "dataset source version")
		format        = flag.String("format", "json", "input format: json | jehle-csv")
	)
	flag.Parse()
	if strings.TrimSpace(*inputPath) == "" {
		fmt.Fprintln(os.Stderr, "--input is required")
		return 2
	}
	body, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		return 1
	}

	var payload []inputLemma
	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "json":
		if err := json.Unmarshal(body, &payload); err != nil {
			fmt.Fprintf(os.Stderr, "parse json: %v\n", err)
			return 1
		}
	case "jehle-csv":
		jlemmas, err2 := spanishverbs.ParseJehleVerbDatabaseCSV(bytes.NewReader(body))
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "parse jehle csv: %v\n", err2)
			return 1
		}
		payload = make([]inputLemma, 0, len(jlemmas))
		for _, jl := range jlemmas {
			forms := make([]inputForm, 0, len(jl.Forms))
			for _, f := range jl.Forms {
				forms = append(forms, inputForm{
					Mood:        f.Mood,
					Tense:       f.Tense,
					Person:      f.Person,
					Number:      f.Number,
					Form:        f.SurfaceForm,
					IsIrregular: false,
				})
			}
			payload = append(payload, inputLemma{Lemma: jl.Lemma, Forms: forms})
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown --format %q (use json or jehle-csv)\n", *format)
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		return 1
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		return 1
	}
	defer db.Close()

	repo := repository.NewVerbFormsRepository(db.GetConnection(), log)
	totalForms := 0
	for _, lemma := range payload {
		lemmaNorm := strings.ToLower(strings.TrimSpace(lemma.Lemma))
		if lemmaNorm == "" {
			continue
		}
		sum := sha256.Sum256([]byte(lemmaNorm + ":" + *sourceVersion))
		lemmaID, err := repo.UpsertVerbLemma(lemmaNorm, "es", *source, *sourceVersion, hex.EncodeToString(sum[:]), "{}")
		if err != nil {
			fmt.Fprintf(os.Stderr, "upsert lemma %q: %v\n", lemmaNorm, err)
			return 1
		}
		for _, f := range lemma.Forms {
			if strings.TrimSpace(f.Form) == "" {
				continue
			}
			_, err := repo.UpsertVerbForm(&models.VerbFormDict{
				VerbLemmaID: lemmaID,
				Mood:        f.Mood,
				Tense:       f.Tense,
				Person:      f.Person,
				Number:      f.Number,
				SurfaceForm: f.Form,
				IsIrregular: f.IsIrregular,
				TagsJSON:    "{}",
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "upsert form %q/%q: %v\n", lemmaNorm, f.Form, err)
				return 1
			}
			totalForms++
		}
	}
	fmt.Printf("imported lemmas=%d forms=%d source=%s version=%s format=%s\n", len(payload), totalForms, *source, *sourceVersion, *format)
	return 0
}
