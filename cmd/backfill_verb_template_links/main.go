package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
)

func main() {
	os.Exit(run())
}

// Offline rules: extend map as more lemmas get curated templates in code/DB.
var lemmaTemplateRules = map[string]struct {
	verbClass string
	templateIDs []string
}{
	"ir": {verbClass: spanishverbs.VerbClassMotion, templateIDs: spanishverbs.IrTemplateCodes()},
}

func run() int {
	var (
		dryRun  = flag.Bool("dry-run", false, "print planned merges only")
		limit   = flag.Int("limit", 0, "max lemmas to touch (0 = all known rules)")
		force   = flag.Bool("force", false, "overwrite existing allowed_template_ids / verb_class from rules")
		lemmasF = flag.String("lemmas", "", "comma-separated lemmas to process (default: all keys in offline rules)")
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

	var lemmas []string
	if strings.TrimSpace(*lemmasF) != "" {
		for _, p := range strings.Split(*lemmasF, ",") {
			k := strings.ToLower(strings.TrimSpace(p))
			if k != "" {
				lemmas = append(lemmas, k)
			}
		}
	} else {
		for k := range lemmaTemplateRules {
			lemmas = append(lemmas, k)
		}
	}
	if *limit > 0 && len(lemmas) > *limit {
		lemmas = lemmas[:*limit]
	}

	meta, err := repo.GetVerbLemmaMetadataJSONBatch(lemmas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "metadata: %v\n", err)
		return 1
	}

	updated := 0
	for _, lem := range lemmas {
		rule, ok := lemmaTemplateRules[lem]
		if !ok {
			fmt.Fprintf(os.Stderr, "no offline rule for lemma %q (add to lemmaTemplateRules)\n", lem)
			return 1
		}
		prev := meta[lem]
		if !*force {
			hasClass := spanishverbs.VerbClassFromLemmaMetadataJSON(prev) != ""
			hasTpl := len(spanishverbs.AllowedTemplateIDsFromLemmaMetadataJSON(prev)) > 0
			if hasClass && hasTpl {
				fmt.Printf("skip %s (already has class+templates, use --force)\n", lem)
				continue
			}
		}
		merged := prev
		var err error
		merged, err = spanishverbs.MergeVerbClassIntoLemmaMetadataJSON(merged, rule.verbClass, "offline-verb-template-links")
		if err != nil {
			fmt.Fprintf(os.Stderr, "merge class %s: %v\n", lem, err)
			return 1
		}
		merged, err = spanishverbs.MergeAllowedTemplateIDsIntoLemmaMetadataJSON(merged, rule.templateIDs, "offline-verb-template-links")
		if err != nil {
			fmt.Fprintf(os.Stderr, "merge templates %s: %v\n", lem, err)
			return 1
		}
		if merged == prev {
			fmt.Printf("skip %s (unchanged)\n", lem)
			continue
		}
		if *dryRun {
			fmt.Printf("dry-run: would update %s -> %s\n", lem, merged)
			updated++
			continue
		}
		if err := repo.UpdateVerbLemmaMetadataJSON(lem, merged); err != nil {
			fmt.Fprintf(os.Stderr, "update %s: %v\n", lem, err)
			return 1
		}
		updated++
		fmt.Printf("updated %s\n", lem)
	}
	fmt.Printf("done rows=%d dry_run=%v\n", updated, *dryRun)
	return 0
}
