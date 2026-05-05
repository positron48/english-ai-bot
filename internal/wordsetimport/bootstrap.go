package wordsetimport

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/resources/wordsets"

	"go.yaml.in/yaml/v3"
	"go.uber.org/zap"
)

const (
	englishCSVChecksumSetting = "word_sets.english.csv_sha256"
	englishCSVImportedSetting = "word_sets.english.last_import_summary"
	defaultEnglishCSVPath     = "/app/data/english_word_freq_pos_ud_top6000.filtered.csv"
	fallbackEnglishCSVPath    = "resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv"
	englishMustHaveChecksum   = "word_sets.english.must_have_sha256"
	englishMustHaveSummary    = "word_sets.english.must_have_last_import_summary"
	defaultEnglishMustHaveYAMLPath  = "/app/data/english_word_sets_must_have.yaml"
	fallbackEnglishMustHaveYAMLPath = "courses/english-grammar/word-sets-must-have.yaml"
	spanishMustHaveChecksum   = "word_sets.spanish.must_have_sha256"
	spanishMustHaveSummary    = "word_sets.spanish.must_have_last_import_summary"
	defaultSpanishMustHaveYAMLPath  = "/app/data/spanish_word_sets_must_have.yaml"
	fallbackSpanishMustHaveYAMLPath = "courses/spanish-grammar/word-sets-must-have.yaml"
	systemUserID              = int64(0)
)

type mustHaveFile struct {
	MustHave mustHaveRoot `yaml:"must_have"`
}

type mustHaveRoot struct {
	Title         string                 `yaml:"title"`
	Description   string                 `yaml:"description"`
	Subcategories []mustHaveSubcategory  `yaml:"subcategories"`
}

type mustHaveSubcategory struct {
	ID    string          `yaml:"id"`
	Title string          `yaml:"title"`
	Sets  []mustHaveSet   `yaml:"sets"`
}

type mustHaveSet struct {
	ID    string          `yaml:"id"`
	Title string          `yaml:"title"`
	Words []mustHaveWord  `yaml:"words"`
}

type mustHaveWord struct {
	ES string `yaml:"es"`
	EN string `yaml:"en"`
	RU string `yaml:"ru"`
}

func detectEnglishCSVPath() (string, error) {
	candidates := []string{
		defaultEnglishCSVPath,
		fallbackEnglishCSVPath,
	}
	for _, path := range candidates {
		st, err := os.Stat(path)
		if err == nil && !st.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("english frequency csv not found: %s or %s", defaultEnglishCSVPath, fallbackEnglishCSVPath)
}

func detectMustHaveYAMLPath(candidates []string) (string, error) {
	for _, path := range candidates {
		st, err := os.Stat(path)
		if err == nil && !st.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("must-have yaml not found: %s", strings.Join(candidates, " or "))
}

func fileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func upsertCategory(conn *sql.DB, parentID *int64, name, description string, sortOrder int) (int64, error) {
	var parent any
	if parentID != nil {
		parent = *parentID
	}
	_, err := conn.Exec(`
		INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
		SELECT $1::bigint, $2, $3, 1, $4
		WHERE NOT EXISTS (
			SELECT 1 FROM word_set_categories
			WHERE (($1::bigint IS NULL AND parent_id IS NULL) OR parent_id = $1::bigint) AND name = $2
		)`,
		parent, name, description, sortOrder)
	if err != nil {
		return 0, err
	}
	var id int64
	err = conn.QueryRow(`
		SELECT id FROM word_set_categories
		WHERE (($1::bigint IS NULL AND parent_id IS NULL) OR parent_id = $1::bigint) AND name = $2
		ORDER BY id LIMIT 1`,
		parent, name).Scan(&id)
	return id, err
}

func ensureCategoryMetadata(conn *sql.DB, id int64, description string, sortOrder int) error {
	_, err := conn.Exec(`
		UPDATE word_set_categories
		SET description = CASE
			WHEN (description IS NULL OR btrim(description) = '') AND btrim($2) <> '' THEN $2
			ELSE description
		END,
		sort_order = $3,
		is_published = 1
		WHERE id = $1`,
		id, strings.TrimSpace(description), sortOrder)
	return err
}

func upsertWordSet(conn *sql.DB, categoryID int64, title, description, preferredPOS string, sortOrder int) error {
	_, err := conn.Exec(`
		INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
		SELECT $1, $2, $3, 1, $4, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM word_sets WHERE category_id = $1 AND title = $2
		)`,
		categoryID, title, description, sortOrder, preferredPOS)
	return err
}

func ensureWordSet(conn *sql.DB, categoryID int64, title, description string, sortOrder int, preferredPOS *string) (int64, error) {
	var pos any
	if preferredPOS != nil {
		p := strings.TrimSpace(*preferredPOS)
		if p != "" {
			pos = p
		}
	}
	_, err := conn.Exec(`
		INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
		SELECT $1, $2, $3, 1, $4, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM word_sets WHERE category_id = $1 AND title = $2
		)`,
		categoryID, title, description, sortOrder, pos)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := conn.QueryRow(`
		SELECT id FROM word_sets
		WHERE category_id = $1 AND title = $2
		ORDER BY id LIMIT 1`,
		categoryID, title).Scan(&id); err != nil {
		return 0, err
	}
	_, err = conn.Exec(`
		UPDATE word_sets
		SET description = CASE
			WHEN (description IS NULL OR btrim(description) = '') AND btrim($2) <> '' THEN $2
			ELSE description
		END,
		sort_order = $3,
		is_published = 1
		WHERE id = $1`,
		id, strings.TrimSpace(description), sortOrder)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func loadMustHaveBlueprint(path string) (*mustHaveRoot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f mustHaveFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if strings.TrimSpace(f.MustHave.Title) == "" {
		return nil, fmt.Errorf("must_have.title is required")
	}
	return &f.MustHave, nil
}

func normalizeMustHaveWords(words []mustHaveWord, lang string) []string {
	out := make([]string, 0, len(words))
	seen := make(map[string]struct{}, len(words))
	for _, w := range words {
		raw := strings.TrimSpace(w.ES)
		if strings.EqualFold(lang, "en") {
			raw = strings.TrimSpace(w.EN)
		}
		lemma := wordsets.NormalizeLemmaImport(raw)
		if lemma == "" {
			continue
		}
		if _, ok := seen[lemma]; ok {
			continue
		}
		seen[lemma] = struct{}{}
		out = append(out, lemma)
	}
	return out
}

func defaultDescription(kind, title string) string {
	switch kind {
	case "root":
		return "Must-have Spanish vocabulary sets for everyday situations."
	case "subcategory":
		return fmt.Sprintf("Must-have vocabulary for %s.", title)
	default:
		return fmt.Sprintf("Must-have words for %s.", title)
	}
}

func ensureMustHaveBlueprint(ctx context.Context, conn *sql.DB, cfg *config.Config, root *mustHaveRoot, lang string, commit bool, log *zap.Logger) (int, int, error) {
	rootTitle := strings.TrimSpace(root.Title)
	rootDesc := strings.TrimSpace(root.Description)
	if rootDesc == "" {
		rootDesc = defaultDescription("root", rootTitle)
	}
	rootID, err := upsertCategory(conn, nil, rootTitle, rootDesc, 100)
	if err != nil {
		return 0, 0, err
	}
	if err := ensureCategoryMetadata(conn, rootID, rootDesc, 100); err != nil {
		return 0, 0, err
	}

	wordSetRepo := repository.NewWordSetRepository(conn, log)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, log)
	wordRepo := repository.NewWordRepository(conn, log)
	svc := service.NewWordSetService(
		wordSetRepo,
		wordSetCategoryRepo,
		wordRepo,
		nil,
		nil,
		nil,
		nil,
		cfg.Learning,
		"",
		log,
	)

	totalSets := 0
	totalItems := 0
	for subIdx, sub := range root.Subcategories {
		subTitle := strings.TrimSpace(sub.Title)
		if subTitle == "" {
			continue
		}
		subDesc := defaultDescription("subcategory", subTitle)
		subID, err := upsertCategory(conn, &rootID, subTitle, subDesc, subIdx)
		if err != nil {
			return totalSets, totalItems, err
		}
		if err := ensureCategoryMetadata(conn, subID, subDesc, subIdx); err != nil {
			return totalSets, totalItems, err
		}
		for setIdx, set := range sub.Sets {
			setTitle := strings.TrimSpace(set.Title)
			if setTitle == "" {
				continue
			}
			setDesc := defaultDescription("set", setTitle)
			setID, err := ensureWordSet(conn, subID, setTitle, setDesc, setIdx, nil)
			if err != nil {
				return totalSets, totalItems, err
			}
			lemmas := normalizeMustHaveWords(set.Words, lang)
			totalSets++
			totalItems += len(lemmas)
			if !commit || len(lemmas) == 0 {
				continue
			}
			if err := svc.ProcessWordSetItems(ctx, setID, strings.Join(lemmas, ",")); err != nil {
				return totalSets, totalItems, err
			}
		}
	}
	return totalSets, totalItems, nil
}
func ensureRangeSets(conn *sql.DB, categoryID int64, titlePrefix, itemName, preferredPOS string, fromRank, toRank, step int) error {
	idx := 0
	for start := fromRank; start <= toRank; start += step {
		end := start + step - 1
		if end > toRank {
			end = toRank
		}
		rangeStr := strconv.Itoa(start) + "–" + strconv.Itoa(end)
		title := fmt.Sprintf("%s — Top 50 (Ranks %s)", titlePrefix, rangeStr)
		desc := fmt.Sprintf("%s ranked %s by frequency (within %s).", itemName, rangeStr, preferredPOS+"s")
		if preferredPOS == "adverb" {
			desc = fmt.Sprintf("%s ranked %s by frequency (within adverbs).", itemName, rangeStr)
		}
		if err := upsertWordSet(conn, categoryID, title, desc, preferredPOS, idx); err != nil {
			return err
		}
		idx++
	}
	return nil
}

// EnsureEnglishWordSetBlueprint recreates canonical English categories/sets if they were deleted.
func EnsureEnglishWordSetBlueprint(conn *sql.DB) error {
	coreID, err := upsertCategory(conn, nil, "Core Frequency", "High-frequency everyday words used across most conversations and texts.", 0)
	if err != nil {
		return err
	}
	extID, err := upsertCategory(conn, nil, "Extended Frequency", "Mid-frequency words that expand everyday vocabulary beyond the core.", 1)
	if err != nil {
		return err
	}

	coreVerbsID, err := upsertCategory(conn, &coreID, "Top 500 Verbs", "High-frequency verbs used across everyday speech and texts.", 0)
	if err != nil {
		return err
	}
	coreNounsID, err := upsertCategory(conn, &coreID, "Top 500 Nouns", "High-frequency everyday nouns used across most conversations and texts.", 1)
	if err != nil {
		return err
	}
	coreAdjsID, err := upsertCategory(conn, &coreID, "Top 500 Adjectives", "High-frequency adjectives that help you describe people, things, and situations.", 2)
	if err != nil {
		return err
	}
	coreAdvsID, err := upsertCategory(conn, &coreID, "Top Adverbs", "High-frequency adverbs used to express time, frequency, degree, manner, and certainty.", 3)
	if err != nil {
		return err
	}
	coreNouns2ID, err := upsertCategory(conn, &coreID, "Top 500–1000 Nouns", "Mid-frequency nouns that expand core vocabulary and improve precision.", 4)
	if err != nil {
		return err
	}
	extNouns1ID, err := upsertCategory(conn, &extID, "Top 1001–1500 Nouns", "Extended-frequency nouns (ranks 1001–1500 within nouns).", 0)
	if err != nil {
		return err
	}
	extNouns2ID, err := upsertCategory(conn, &extID, "Top 1501–2000 Nouns", "Extended-frequency nouns (ranks 1501–2000 within nouns).", 1)
	if err != nil {
		return err
	}

	if err := ensureRangeSets(conn, coreVerbsID, "Core Verbs", "Verbs", "verb", 1, 500, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, coreNounsID, "Core Nouns", "Nouns", "noun", 1, 500, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, coreAdjsID, "Core Adjectives", "Adjectives", "adjective", 1, 500, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, coreAdvsID, "Core Adverbs", "Adverbs", "adverb", 1, 300, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, coreNouns2ID, "Core Nouns", "Nouns", "noun", 501, 1000, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, extNouns1ID, "Extended Nouns", "Nouns", "noun", 1001, 1500, 50); err != nil {
		return err
	}
	if err := ensureRangeSets(conn, extNouns2ID, "Extended Nouns", "Nouns", "noun", 1501, 2000, 50); err != nil {
		return err
	}
	return nil
}

func shouldForceEnglishImport(conn *sql.DB, currentChecksum, newChecksum string) (bool, string, error) {
	if currentChecksum != newChecksum {
		return true, "csv_checksum_changed", nil
	}
	var setCount int
	if err := conn.QueryRow(`
		SELECT COUNT(*) FROM word_sets
		WHERE title ~* '(ranks?|rangos?)\s+[0-9]+'`).Scan(&setCount); err != nil {
		return false, "", err
	}
	if setCount == 0 {
		return true, "ranked_sets_missing", nil
	}
	var itemCount int
	if err := conn.QueryRow(`
		SELECT COUNT(*) FROM word_set_items wsi
		INNER JOIN word_sets ws ON ws.id = wsi.word_set_id
		WHERE ws.title ~* '(ranks?|rangos?)\s+[0-9]+'`).Scan(&itemCount); err != nil {
		return false, "", err
	}
	if itemCount == 0 {
		return true, "word_set_items_empty", nil
	}
	return false, "up_to_date", nil
}

// EnsureEnglishLabels updates known Spanish labels back to canonical English names for English deployment.
func EnsureEnglishLabels(conn *sql.DB) error {
	statements := []string{
		`UPDATE word_set_categories SET name='Core Frequency', description='High-frequency everyday words used across most conversations and texts.' WHERE parent_id IS NULL AND name='Frecuencia esencial'`,
		`UPDATE word_set_categories SET name='Extended Frequency', description='Mid-frequency words that expand everyday vocabulary beyond the core.' WHERE parent_id IS NULL AND name='Frecuencia ampliada'`,
		`UPDATE word_set_categories SET name='Top 500 Verbs', description='High-frequency verbs used across everyday speech and texts.' WHERE name='Top 500 verbos'`,
		`UPDATE word_set_categories SET name='Top 500 Nouns', description='High-frequency everyday nouns used across most conversations and texts.' WHERE name='Top 500 sustantivos'`,
		`UPDATE word_set_categories SET name='Top 500 Adjectives', description='High-frequency adjectives that help you describe people, things, and situations.' WHERE name='Top 500 adjetivos'`,
		`UPDATE word_set_categories SET name='Top Adverbs', description='High-frequency adverbs used to express time, frequency, degree, manner, and certainty.' WHERE name='Top adverbios'`,
		`UPDATE word_set_categories SET name='Top 500–1000 Nouns', description='Mid-frequency nouns that expand core vocabulary and improve precision.' WHERE name='Top 500-1000 sustantivos'`,
		`UPDATE word_set_categories SET name='Top 1001–1500 Nouns', description='Extended-frequency nouns (ranks 1001–1500 within nouns).' WHERE name='Top 1001-1500 sustantivos'`,
		`UPDATE word_set_categories SET name='Top 1501–2000 Nouns', description='Extended-frequency nouns (ranks 1501–2000 within nouns).' WHERE name='Top 1501-2000 sustantivos'`,
		`UPDATE word_sets SET title='Core Verbs — Top 50 (Ranks ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ')', description='Verbs ranked ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ' by frequency (within verbs).' WHERE title LIKE 'Verbos esenciales — Top 50 (rangos %)'`,
		`UPDATE word_sets SET title='Core Nouns — Top 50 (Ranks ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ')', description='Nouns ranked ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ' by frequency (within nouns).' WHERE title LIKE 'Sustantivos esenciales — Top 50 (rangos %)'`,
		`UPDATE word_sets SET title='Core Adjectives — Top 50 (Ranks ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ')', description='Adjectives ranked ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ' by frequency (within adjectives).' WHERE title LIKE 'Adjetivos esenciales — Top 50 (rangos %)'`,
		`UPDATE word_sets SET title='Core Adverbs — Top 50 (Ranks ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ')', description='Adverbs ranked ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ' by frequency (within adverbs).' WHERE title LIKE 'Adverbios esenciales — Top 50 (rangos %)'`,
		`UPDATE word_sets SET title='Extended Nouns — Top 50 (Ranks ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ')', description='Nouns ranked ' || coalesce(substring(title from 'Ranks ([0-9–-]+)'), substring(title from 'rangos ([0-9–-]+)')) || ' by frequency (within nouns).' WHERE title LIKE 'Sustantivos ampliados — Top 50 (rangos %)'`,
	}
	for _, stmt := range statements {
		if _, err := conn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// AutoSyncEnglishWordSets syncs word_set_items from English frequency CSV once per CSV checksum.
func AutoSyncEnglishWordSets(ctx context.Context, cfg *config.Config, conn *sql.DB, log *zap.Logger) error {
	target := strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	appCode := strings.ToLower(strings.TrimSpace(cfg.Learning.AppCode))
	if target != "en" || appCode != "english" {
		return nil
	}

	csvPath, err := detectEnglishCSVPath()
	if err != nil {
		return err
	}
	checksum, err := fileSHA256(csvPath)
	if err != nil {
		return fmt.Errorf("compute english csv checksum: %w", err)
	}

	if err := EnsureEnglishLabels(conn); err != nil {
		return fmt.Errorf("normalize english labels: %w", err)
	}
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		return fmt.Errorf("ensure english blueprint: %w", err)
	}

	settingsRepo := repository.NewAppSettingsRepository(conn, log)
	currentChecksum, err := settingsRepo.GetSetting(englishCSVChecksumSetting)
	if err != nil {
		return fmt.Errorf("read app setting %q: %w", englishCSVChecksumSetting, err)
	}
	needImport, reason, err := shouldForceEnglishImport(conn, currentChecksum, checksum)
	if err != nil {
		return fmt.Errorf("determine english import necessity: %w", err)
	}
	if !needImport {
		log.Info("english word sets bootstrap skipped: csv checksum unchanged",
			zap.String("csv_path", csvPath),
			zap.String("csv_sha256", checksum),
			zap.String("reason", reason))
		return nil
	}

	res, err := Import(ctx, cfg, conn, log, ImportOptions{
		CSVPath: csvPath,
		Lang:    "en",
		Commit:  true,
	})
	if err != nil {
		return fmt.Errorf("english word sets import failed: %w", err)
	}

	summary := fmt.Sprintf("mode=%s processed=%d sets=%d csv=%s sha256=%s",
		res.Mode, res.Processed, len(res.Sets), csvPath, checksum)
	if err := settingsRepo.SetSetting(englishCSVChecksumSetting, checksum, systemUserID); err != nil {
		return fmt.Errorf("write app setting %q: %w", englishCSVChecksumSetting, err)
	}
	if err := settingsRepo.SetSetting(englishCSVImportedSetting, summary, systemUserID); err != nil {
		return fmt.Errorf("write app setting %q: %w", englishCSVImportedSetting, err)
	}

	log.Info("english word sets bootstrap completed",
		zap.String("csv_path", csvPath),
		zap.String("csv_sha256", checksum),
		zap.String("reason", reason),
		zap.Int("processed_sets", res.Processed),
		zap.Int("selected_sets", len(res.Sets)))
	return nil
}

// AutoSyncSpanishMustHaveWordSets creates/updates "Must Have" hierarchy for Spanish deployment.
func AutoSyncSpanishMustHaveWordSets(ctx context.Context, cfg *config.Config, conn *sql.DB, log *zap.Logger) error {
	target := strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	appCode := strings.ToLower(strings.TrimSpace(cfg.Learning.AppCode))
	if target != "es" || appCode != "spanish" {
		return nil
	}

	yamlPath, err := detectMustHaveYAMLPath([]string{
		defaultSpanishMustHaveYAMLPath,
		fallbackSpanishMustHaveYAMLPath,
	})
	if err != nil {
		return err
	}
	checksum, err := fileSHA256(yamlPath)
	if err != nil {
		return fmt.Errorf("compute must-have yaml checksum: %w", err)
	}
	root, err := loadMustHaveBlueprint(yamlPath)
	if err != nil {
		return fmt.Errorf("parse must-have yaml: %w", err)
	}

	settingsRepo := repository.NewAppSettingsRepository(conn, log)
	currentChecksum, err := settingsRepo.GetSetting(spanishMustHaveChecksum)
	if err != nil {
		return fmt.Errorf("read app setting %q: %w", spanishMustHaveChecksum, err)
	}
	commit := currentChecksum != checksum

	sets, items, err := ensureMustHaveBlueprint(ctx, conn, cfg, root, "es", commit, log)
	if err != nil {
		return fmt.Errorf("sync must-have blueprint: %w", err)
	}

	reason := "up_to_date"
	if commit {
		reason = "yaml_checksum_changed"
		if err := settingsRepo.SetSetting(spanishMustHaveChecksum, checksum, systemUserID); err != nil {
			return fmt.Errorf("write app setting %q: %w", spanishMustHaveChecksum, err)
		}
	}
	summary := fmt.Sprintf("mode=%s sets=%d items=%d yaml=%s sha256=%s",
		map[bool]string{true: "COMMIT", false: "DRY-RUN"}[commit], sets, items, yamlPath, checksum)
	if err := settingsRepo.SetSetting(spanishMustHaveSummary, summary, systemUserID); err != nil {
		return fmt.Errorf("write app setting %q: %w", spanishMustHaveSummary, err)
	}

	log.Info("spanish must-have word sets bootstrap completed",
		zap.String("mode", map[bool]string{true: "COMMIT", false: "DRY-RUN"}[commit]),
		zap.String("reason", reason),
		zap.Int("sets", sets),
		zap.Int("items", items),
		zap.String("yaml_path", yamlPath),
		zap.String("yaml_sha256", checksum))
	return nil
}

// AutoSyncEnglishMustHaveWordSets creates/updates "Must Have" hierarchy for English deployment.
func AutoSyncEnglishMustHaveWordSets(ctx context.Context, cfg *config.Config, conn *sql.DB, log *zap.Logger) error {
	target := strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	appCode := strings.ToLower(strings.TrimSpace(cfg.Learning.AppCode))
	if target != "en" || appCode != "english" {
		return nil
	}

	yamlPath, err := detectMustHaveYAMLPath([]string{
		defaultEnglishMustHaveYAMLPath,
		fallbackEnglishMustHaveYAMLPath,
	})
	if err != nil {
		return err
	}
	checksum, err := fileSHA256(yamlPath)
	if err != nil {
		return fmt.Errorf("compute english must-have yaml checksum: %w", err)
	}
	root, err := loadMustHaveBlueprint(yamlPath)
	if err != nil {
		return fmt.Errorf("parse english must-have yaml: %w", err)
	}

	settingsRepo := repository.NewAppSettingsRepository(conn, log)
	currentChecksum, err := settingsRepo.GetSetting(englishMustHaveChecksum)
	if err != nil {
		return fmt.Errorf("read app setting %q: %w", englishMustHaveChecksum, err)
	}
	commit := currentChecksum != checksum

	sets, items, err := ensureMustHaveBlueprint(ctx, conn, cfg, root, "en", commit, log)
	if err != nil {
		return fmt.Errorf("sync english must-have blueprint: %w", err)
	}

	reason := "up_to_date"
	if commit {
		reason = "yaml_checksum_changed"
		if err := settingsRepo.SetSetting(englishMustHaveChecksum, checksum, systemUserID); err != nil {
			return fmt.Errorf("write app setting %q: %w", englishMustHaveChecksum, err)
		}
	}
	summary := fmt.Sprintf("mode=%s sets=%d items=%d yaml=%s sha256=%s",
		map[bool]string{true: "COMMIT", false: "DRY-RUN"}[commit], sets, items, yamlPath, checksum)
	if err := settingsRepo.SetSetting(englishMustHaveSummary, summary, systemUserID); err != nil {
		return fmt.Errorf("write app setting %q: %w", englishMustHaveSummary, err)
	}

	log.Info("english must-have word sets bootstrap completed",
		zap.String("mode", map[bool]string{true: "COMMIT", false: "DRY-RUN"}[commit]),
		zap.String("reason", reason),
		zap.Int("sets", sets),
		zap.Int("items", items),
		zap.String("yaml_path", yamlPath),
		zap.String("yaml_sha256", checksum))
	return nil
}
