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

	"go.uber.org/zap"
)

const (
	englishCSVChecksumSetting = "word_sets.english.csv_sha256"
	englishCSVImportedSetting = "word_sets.english.last_import_summary"
	defaultEnglishCSVPath     = "/app/data/english_word_freq_pos_ud_top6000.filtered.csv"
	fallbackEnglishCSVPath    = "resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv"
	systemUserID              = int64(0)
)

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
