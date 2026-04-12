package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		lemma   = flag.String("lemma", "", "Spanish infinitive (required), e.g. hablar")
		lang    = flag.String("lang", "es", "verb_lemmas.language")
		maxRows = flag.Int("max", 0, "max rows to print (0 = all)")
		tsv     = flag.Bool("tsv", false, "tab-separated output (no table header)")
	)
	flag.Parse()

	lem := strings.TrimSpace(strings.ToLower(*lemma))
	if lem == "" {
		fmt.Fprintln(os.Stderr, "usage: preview_verb_templates -lemma=hablar [-lang=es] [-max=N] [-tsv]")
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
	forms, err := repo.ListVerbFormsForLemma(lem, *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list forms: %v\n", err)
		return 1
	}
	if len(forms) == 0 {
		fmt.Fprintf(os.Stderr, "no verb_forms_dict rows for lemma %q (language %q). Import paradigms first.\n", lem, *lang)
		return 1
	}

	if tpl, err := repo.ListVerbExampleCatalogTemplatesCached(); err == nil {
		spanishverbs.RegisterExternalTemplates(tpl)
	}

	meta, _ := repo.GetVerbLemmaMetadataJSONBatch([]string{lem})
	rawMeta := meta[lem]
	ruGloss := spanishverbs.RuGlossFromLemmaMetadataJSON(rawMeta)
	if ruGloss == "" {
		ruGloss = spanishverbs.DefaultRuGloss(lem)
	}
	verbClass := spanishverbs.VerbClassFromLemmaMetadataJSON(rawMeta)
	allowed := spanishverbs.AllowedTemplateIDsFromLemmaMetadataJSON(rawMeta)

	fmt.Printf("# lemma=%s language=%s forms=%d ru_gloss=%q verb_class=%q allowed_template_ids=%v\n",
		lem, *lang, len(forms), ruGloss, verbClass, allowed)

	limit := len(forms)
	if *maxRows > 0 && *maxRows < limit {
		limit = *maxRows
	}

	var tw *tabwriter.Writer
	if *tsv {
		tw = tabwriter.NewWriter(os.Stdout, 0, 0, 1, '\t', 0)
	} else {
		tw = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "mood\ttense\tp\tnumber\tsurface\tsource\tES\tRU")
	}

	for i := 0; i < limit; i++ {
		row := forms[i]
		seed := row.VerbFormDictID ^ int64(stringsHash(lem))
		es, ru := spanishverbs.GenerateVerbExamplePair(
			seed, lem, row.Mood, row.Tense, row.Person, row.Number, row.SurfaceForm, ruGloss, verbClass, allowed,
		)
		src := "generic"
		if esC, ruC, ok := spanishverbs.TryGenerateCatalogPair(
			seed, lem, row.Mood, row.Tense, row.Person, row.Number, row.SurfaceForm, ruGloss, verbClass, allowed,
		); ok && spanishverbs.ExampleContainsSurface(esC, row.SurfaceForm) && esC == es && ruC == ru {
			src = "catalog"
		}
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
			row.Mood, row.Tense, row.Person, row.Number, row.SurfaceForm, src, es, ru)
		_, _ = fmt.Fprintln(tw, line)
	}
	_ = tw.Flush()
	return 0
}

func stringsHash(s string) uint32 {
	var h uint32 = 5381
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return h
}
