package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"

	"go.uber.org/zap"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		batchSize   = flag.Int("batch-size", 28, "Spanish lemmas per LLM request")
		sleepMs     = flag.Int("sleep-ms", 800, "pause between LLM batches (rate limits)")
		dryRun      = flag.Bool("dry-run", false, "print actions only, no LLM or DB updates")
		force       = flag.Bool("force", false, "overwrite existing ru.gloss in metadata")
		limitLemmas = flag.Int("limit", 0, "max lemmas to process (0 = all)")
		fillClass   = flag.Bool("fill-class", false, "LLM batch: set verb_class only (motion|speech|transfer|generic); skips ru gloss pass")
	)
	flag.Parse()

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
	lemmas, err := repo.ListSpanishVerbLemmas()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list lemmas: %v\n", err)
		return 1
	}
	if *limitLemmas > 0 && len(lemmas) > *limitLemmas {
		lemmas = lemmas[:*limitLemmas]
	}

	if *fillClass {
		return runFillClass(cfg, log, repo, lemmas, *batchSize, *sleepMs, *dryRun, *force)
	}

	metaAll, err := repo.GetVerbLemmaMetadataJSONBatch(lemmas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load metadata: %v\n", err)
		return 1
	}

	todo := make([]string, 0, len(lemmas))
	for _, lem := range lemmas {
		k := strings.ToLower(strings.TrimSpace(lem))
		if k == "" {
			continue
		}
		if !*force {
			if g := spanishverbs.RuGlossFromLemmaMetadataJSON(metaAll[k]); g != "" {
				continue
			}
		}
		todo = append(todo, k)
	}
	fmt.Printf("lemmas total=%d to_fill=%d batch_size=%d dry_run=%v\n", len(lemmas), len(todo), *batchSize, *dryRun)
	if len(todo) == 0 {
		fmt.Println("nothing to do")
		return 0
	}

	if *dryRun {
		for i := 0; i < len(todo) && i < 15; i++ {
			fmt.Printf("would batch: %s\n", todo[i])
		}
		if len(todo) > 15 {
			fmt.Printf("... and %d more\n", len(todo)-15)
		}
		return 0
	}

	aiSvc := ai.NewServiceWithTimeout(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, ai.ParseHTTPTimeout(cfg.AI.RequestTimeout), log)
	const sys = `You are a lexicographer. Output ONLY one minified JSON object, no markdown.
Keys: Spanish infinitive lemmas (lowercase, ASCII letters and Spanish letters as in input).
Values: short Russian literary gloss for dictionary main sense — 2–7 words, Cyrillic, no nested JSON, no quotes inside values, no line breaks inside values.
If a lemma is unknown, omit it from the object.`

	ctx := context.Background()
	updated := 0
	for i := 0; i < len(todo); i += *batchSize {
		end := i + *batchSize
		if end > len(todo) {
			end = len(todo)
		}
		chunk := todo[i:end]
		userPayload, _ := json.Marshal(map[string]interface{}{"lemmas": chunk})
		raw, err := aiSvc.ChatSystemUser(ctx, sys, "Return the JSON object for these lemmas: "+string(userPayload))
		if err != nil {
			fmt.Fprintf(os.Stderr, "llm batch %d-%d: %v\n", i, end, err)
			return 1
		}
		var glosses map[string]string
		if err := json.Unmarshal([]byte(raw), &glosses); err != nil {
			fmt.Fprintf(os.Stderr, "json decode batch %d-%d: %v\nbody: %s\n", i, end, err, raw)
			return 1
		}
		existingChunk, err := repo.GetVerbLemmaMetadataJSONBatch(chunk)
		if err != nil {
			fmt.Fprintf(os.Stderr, "metadata batch: %v\n", err)
			return 1
		}
		chunkSet := map[string]struct{}{}
		for _, lem := range chunk {
			chunkSet[lem] = struct{}{}
		}
		for k, g := range glosses {
			lem := strings.ToLower(strings.TrimSpace(k))
			if _, ok := chunkSet[lem]; !ok {
				continue
			}
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			prev := existingChunk[lem]
			merged, err := spanishverbs.MergeRuGlossIntoLemmaMetadataJSON(prev, g, "llm-batch-es-ru-v1")
			if err != nil {
				fmt.Fprintf(os.Stderr, "merge %s: %v\n", lem, err)
				return 1
			}
			if err := repo.UpdateVerbLemmaMetadataJSON(lem, merged); err != nil {
				fmt.Fprintf(os.Stderr, "update %s: %v\n", lem, err)
				return 1
			}
			updated++
		}
		if *sleepMs > 0 && end < len(todo) {
			time.Sleep(time.Duration(*sleepMs) * time.Millisecond)
		}
	}
	fmt.Printf("updated lemma metadata rows=%d\n", updated)
	return 0
}

func runFillClass(cfg *config.Config, log *zap.Logger, repo *repository.VerbFormsRepository, lemmas []string, batchSize, sleepMs int, dryRun, force bool) int {
	metaAll, err := repo.GetVerbLemmaMetadataJSONBatch(lemmas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load metadata: %v\n", err)
		return 1
	}
	todo := make([]string, 0, len(lemmas))
	for _, lem := range lemmas {
		k := strings.ToLower(strings.TrimSpace(lem))
		if k == "" {
			continue
		}
		if !force {
			if spanishverbs.VerbClassFromLemmaMetadataJSON(metaAll[k]) != "" {
				continue
			}
		}
		todo = append(todo, k)
	}
	fmt.Printf("fill-class: lemmas total=%d to_fill=%d batch_size=%d dry_run=%v\n", len(lemmas), len(todo), batchSize, dryRun)
	if len(todo) == 0 {
		fmt.Println("nothing to do")
		return 0
	}
	if dryRun {
		for i := 0; i < len(todo) && i < 15; i++ {
			fmt.Printf("would classify: %s\n", todo[i])
		}
		if len(todo) > 15 {
			fmt.Printf("... and %d more\n", len(todo)-15)
		}
		return 0
	}

	aiSvc := ai.NewServiceWithTimeout(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, ai.ParseHTTPTimeout(cfg.AI.RequestTimeout), log)
	const sys = `You are a Spanish lexicographer. Output ONLY one minified JSON object, no markdown.
Keys: Spanish infinitive lemmas (lowercase).
Values: exactly one of: motion, speech, transfer, generic — coarse verb semantics for example templates.
Use "motion" for movement/location change (ir, venir, salir, llegar, ...).
Use "speech" for saying/telling (decir, hablar, ...).
Use "transfer" for giving/taking/bringing (dar, traer, llevar, ...).
Use "generic" when none fits clearly.
If a lemma is unknown, omit it.`

	ctx := context.Background()
	updated := 0
	for i := 0; i < len(todo); i += batchSize {
		end := i + batchSize
		if end > len(todo) {
			end = len(todo)
		}
		chunk := todo[i:end]
		userPayload, _ := json.Marshal(map[string]interface{}{"lemmas": chunk})
		raw, err := aiSvc.ChatSystemUser(ctx, sys, "Return the JSON object for these lemmas: "+string(userPayload))
		if err != nil {
			fmt.Fprintf(os.Stderr, "llm batch %d-%d: %v\n", i, end, err)
			return 1
		}
		var classes map[string]string
		if err := json.Unmarshal([]byte(raw), &classes); err != nil {
			fmt.Fprintf(os.Stderr, "json decode batch %d-%d: %v\nbody: %s\n", i, end, err, raw)
			return 1
		}
		existingChunk, err := repo.GetVerbLemmaMetadataJSONBatch(chunk)
		if err != nil {
			fmt.Fprintf(os.Stderr, "metadata batch: %v\n", err)
			return 1
		}
		chunkSet := map[string]struct{}{}
		for _, lem := range chunk {
			chunkSet[lem] = struct{}{}
		}
		for k, cls := range classes {
			lem := strings.ToLower(strings.TrimSpace(k))
			if _, ok := chunkSet[lem]; !ok {
				continue
			}
			cls = strings.ToLower(strings.TrimSpace(cls))
			switch cls {
			case "motion", "speech", "transfer", "generic":
			default:
				continue
			}
			prev := existingChunk[lem]
			merged, err := spanishverbs.MergeVerbClassIntoLemmaMetadataJSON(prev, cls, "llm-verb-class-v1")
			if err != nil {
				fmt.Fprintf(os.Stderr, "merge %s: %v\n", lem, err)
				return 1
			}
			if err := repo.UpdateVerbLemmaMetadataJSON(lem, merged); err != nil {
				fmt.Fprintf(os.Stderr, "update %s: %v\n", lem, err)
				return 1
			}
			updated++
		}
		if sleepMs > 0 && end < len(todo) {
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		}
	}
	fmt.Printf("fill-class updated rows=%d\n", updated)
	return 0
}
