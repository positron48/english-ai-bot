package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/wordsetimport"
)

func main() {
	os.Exit(runCLI())
}

func runCLI() int {
	csvPath := flag.String("csv", "", "path to frequency CSV (default depends on --lang)")
	lang := flag.String("lang", "", "import language: en or es (default: LEARNING_TARGET_LANG)")
	commit := flag.Bool("commit", false, "apply changes to DB (default: dry-run)")
	onlySetsRaw := flag.String("only-set", "", "comma-separated exact set titles to import (optional)")
	limitSets := flag.Int("limit-sets", 0, "limit number of sets to process after filtering (0 = no limit)")
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
		fmt.Fprintf(os.Stderr, "failed to init database: %v\n", err)
		return 1
	}
	defer db.Close()

	importLang := strings.ToLower(strings.TrimSpace(*lang))
	if importLang == "" {
		importLang = strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	}
	res, err := wordsetimport.Import(ctx, cfg, db.GetConnection(), log, wordsetimport.ImportOptions{
		CSVPath:       *csvPath,
		Lang:          importLang,
		Commit:        *commit,
		OnlySetTitles: parseCSVSet(*onlySetsRaw),
		LimitSets:     *limitSets,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
		return 1
	}

	if len(res.Sets) == 0 {
		fmt.Println("No matching sets found to process.")
		return 0
	}

	fmt.Printf("[%s] Processing %d set(s)\n", res.Mode, len(res.Sets))
	for _, s := range res.Sets {
		if s.Skipped {
			fmt.Printf("SKIP set_id=%d title=%q: %s\n", s.SetID, s.Title, s.Reason)
			continue
		}
		fmt.Printf("SET id=%d title=%q pos=%s range=%d-%d imported=%d\n",
			s.SetID, s.Title, s.POS, s.RangeFrom, s.RangeTo, s.Imported)
	}
	fmt.Printf("Done. processed=%d (mode=%s)\n", res.Processed, res.Mode)
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
