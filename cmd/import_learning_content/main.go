package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/grammartrainingpack"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/readingsync"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/speakingsync"

	"go.uber.org/zap"
)

type importPlan struct {
	BundleID          string
	AppCode           string
	NativeLang        string
	TargetLang        string
	SourceHash        string
	SectionsData      *repository.SectionsData
	IndexData         *repository.IndexData
	SectionsRaw       []byte
	IndexRaw          []byte
	ChaptersRaw       map[string][]byte
	TrainingIndexRaw  []byte
	TrainingIndex     *repository.GrammarTrainingPackIndex
	TrainingQuestions []map[string]interface{}
}

func main() {
	os.Exit(run())
}

func run() int {
	commit := flag.Bool("commit", false, "write imported content to DB")
	flag.Parse()
	if !*commit && strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
		_ = os.Setenv("DATABASE_URL", "postgres://dry-run:dry-run@localhost:5432/dry-run?sslmode=disable")
	}
	if !*commit && strings.TrimSpace(os.Getenv("AI_PROMPT")) == "" && strings.TrimSpace(os.Getenv("AI_PROMPT_FILE")) == "" {
		_ = os.Setenv("AI_PROMPT", "dry-run")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		return 1
	}
	plan, err := buildPlan(cfg, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build import plan: %v\n", err)
		return 1
	}
	mode := "dry-run"
	if *commit {
		mode = "commit"
	}
	fmt.Printf("[%s] bundle=%s app=%s target=%s sections=%d chapters=%d training_questions=%d hash=%s\n",
		mode, plan.BundleID, plan.AppCode, plan.TargetLang, len(plan.SectionsData.Sections), len(plan.ChaptersRaw), len(plan.TrainingQuestions), plan.SourceHash)
	if !*commit {
		return 0
	}

	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		return 1
	}
	defer db.Close()
	ctx := context.Background()
	conn := db.GetConnection()
	if err := writePlan(ctx, conn, plan); err != nil {
		fmt.Fprintf(os.Stderr, "commit import: %v\n", err)
		return 1
	}
	if err := readingsync.SyncFromBundle(ctx, cfg, repository.NewReadingCatalogRepository(conn), log); err != nil {
		fmt.Fprintf(os.Stderr, "commit reading import: %v\n", err)
		return 1
	}
	if err := speakingsync.SyncFromBundle(ctx, cfg, repository.NewSpeakingCatalogRepository(conn), log); err != nil {
		fmt.Fprintf(os.Stderr, "commit speaking import: %v\n", err)
		return 1
	}
	fmt.Println("Import committed.")
	return 0
}

func buildPlan(cfg *config.Config, log *zap.Logger) (*importPlan, error) {
	contentRepo, err := repository.NewGrammarContentRepositoryForLearning(cfg.Learning, log)
	if err != nil {
		return nil, err
	}
	sections, err := contentRepo.GetSections()
	if err != nil {
		return nil, err
	}
	index, err := contentRepo.GetIndex()
	if err != nil {
		return nil, err
	}
	sectionsRaw, err := json.Marshal(sections)
	if err != nil {
		return nil, err
	}
	indexRaw, err := json.Marshal(index)
	if err != nil {
		return nil, err
	}
	chapterIDs := make([]string, 0, len(index.Chapters))
	for id := range index.Chapters {
		chapterIDs = append(chapterIDs, id)
	}
	sort.Strings(chapterIDs)
	chaptersRaw := make(map[string][]byte, len(chapterIDs))
	hash := sha256.New()
	_, _ = hash.Write(sectionsRaw)
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(indexRaw)
	for _, id := range chapterIDs {
		raw, err := contentRepo.GetChapterRawJSON(id)
		if err != nil {
			return nil, err
		}
		chaptersRaw[id] = raw
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(id))
		_, _ = hash.Write(raw)
	}

	packFS, err := grammartrainingpack.PackFS(cfg.Learning.GrammarBundleID)
	if err != nil {
		return nil, err
	}
	trainingIndexRaw, err := fs.ReadFile(packFS, "index.json")
	if err != nil {
		return nil, err
	}
	packRepo := repository.NewGrammarTrainingPackRepositoryWithFS(packFS, log)
	trainingIndex, err := packRepo.GetIndex()
	if err != nil {
		return nil, err
	}
	trainingQuestions, err := packRepo.GetAllQuestions()
	if err != nil {
		return nil, err
	}
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(trainingIndexRaw)
	for _, q := range trainingQuestions {
		raw, _ := json.Marshal(q)
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(raw)
	}

	return &importPlan{
		BundleID:          strings.ToLower(strings.TrimSpace(cfg.Learning.GrammarBundleID)),
		AppCode:           cfg.Learning.AppCode,
		NativeLang:        cfg.Learning.NativeLang,
		TargetLang:        cfg.Learning.TargetLang,
		SourceHash:        hex.EncodeToString(hash.Sum(nil)),
		SectionsData:      sections,
		IndexData:         index,
		SectionsRaw:       sectionsRaw,
		IndexRaw:          indexRaw,
		ChaptersRaw:       chaptersRaw,
		TrainingIndexRaw:  trainingIndexRaw,
		TrainingIndex:     trainingIndex,
		TrainingQuestions: trainingQuestions,
	}, nil
}

func writePlan(ctx context.Context, db *sql.DB, plan *importPlan) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	runID, err := insertImportRun(tx, plan, "commit")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM grammar_training_content_questions WHERE bundle_id = ?`, plan.BundleID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM grammar_content_chapters WHERE bundle_id = ?`, plan.BundleID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM grammar_content_sections WHERE bundle_id = ?`, plan.BundleID); err != nil {
		return err
	}

	for _, section := range plan.SectionsData.Sections {
		raw, _ := json.Marshal(section)
		titleTranslations, _ := json.Marshal(section.TitleTranslations)
		chapterIDs, _ := json.Marshal(section.ChapterIDs)
		if _, err := tx.ExecContext(ctx, `
INSERT INTO grammar_content_sections (bundle_id, section_id, title, title_translations_json, level, sort_order, chapter_ids_json, raw_json, source_hash, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			plan.BundleID, section.SectionID, section.Title, string(titleTranslations), section.Level, section.Order, string(chapterIDs), string(raw), contentHash(raw)); err != nil {
			return fmt.Errorf("section %s: %w", section.SectionID, err)
		}
	}

	chapterIDs := make([]string, 0, len(plan.ChaptersRaw))
	for id := range plan.ChaptersRaw {
		chapterIDs = append(chapterIDs, id)
	}
	sort.Strings(chapterIDs)
	for _, chapterID := range chapterIDs {
		raw := plan.ChaptersRaw[chapterID]
		var chapter repository.Chapter
		if err := json.Unmarshal(raw, &chapter); err != nil {
			return fmt.Errorf("chapter %s parse: %w", chapterID, err)
		}
		titleTranslations, _ := json.Marshal(chapter.TitleTranslations)
		if _, err := tx.ExecContext(ctx, `
INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, title_translations_json, title_short, description, ui_language, target_language, level, sort_order, estimated_minutes, raw_json, source_hash, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			plan.BundleID, chapter.ID, chapter.SectionID, chapter.Title, string(titleTranslations), nullableString(chapter.TitleShort), nullableString(chapter.Description), chapter.UILanguage, chapter.TargetLanguage, nullableString(chapter.Level), chapter.Order, chapter.EstimatedMinutes, string(raw), contentHash(raw)); err != nil {
			return fmt.Errorf("chapter %s: %w", chapter.ID, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO grammar_content_bundle_meta (bundle_id, app_code, native_lang, target_lang, version, generated_at, source_hash, sections_json, index_json, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(bundle_id) DO UPDATE SET app_code=excluded.app_code, native_lang=excluded.native_lang, target_lang=excluded.target_lang, version=excluded.version, generated_at=excluded.generated_at, source_hash=excluded.source_hash, sections_json=excluded.sections_json, index_json=excluded.index_json, updated_at=CURRENT_TIMESTAMP`,
		plan.BundleID, plan.AppCode, plan.NativeLang, plan.TargetLang, plan.IndexData.Version, plan.IndexData.GeneratedAt, plan.SourceHash, string(plan.SectionsRaw), string(plan.IndexRaw)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO grammar_training_content_meta (bundle_id, language, course_id, version, generated_at, index_json, source_hash, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(bundle_id) DO UPDATE SET language=excluded.language, course_id=excluded.course_id, version=excluded.version, generated_at=excluded.generated_at, index_json=excluded.index_json, source_hash=excluded.source_hash, updated_at=CURRENT_TIMESTAMP`,
		plan.BundleID, plan.TrainingIndex.Language, plan.TrainingIndex.CourseID, plan.TrainingIndex.Version, plan.TrainingIndex.GeneratedAt, string(plan.TrainingIndexRaw), plan.SourceHash); err != nil {
		return err
	}
	for _, q := range plan.TrainingQuestions {
		raw, _ := json.Marshal(q)
		qid := fmt.Sprint(q["id"])
		chapterID, _ := q["chapter_id"].(string)
		blockID, _ := q["theory_block_id"].(string)
		conceptID, _ := q["concept_id"].(string)
		difficulty := intFromAny(q["difficulty"])
		if qid == "" || chapterID == "" || blockID == "" {
			return fmt.Errorf("training question missing stable ids: %s", string(raw))
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO grammar_training_content_questions (bundle_id, question_id, chapter_id, theory_block_id, concept_id, difficulty, raw_json, source_hash, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			plan.BundleID, qid, chapterID, blockID, nullableString(conceptID), difficulty, string(raw), contentHash(raw)); err != nil {
			return fmt.Errorf("training question %s: %w", qid, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE learning_content_import_runs
SET finished_at = CURRENT_TIMESTAMP, sections_count = ?, chapters_count = ?, training_questions_count = ?
WHERE id = ?`, len(plan.SectionsData.Sections), len(plan.ChaptersRaw), len(plan.TrainingQuestions), runID); err != nil {
		return err
	}
	return tx.Commit()
}

func insertImportRun(tx *sql.Tx, plan *importPlan, mode string) (int64, error) {
	var id int64
	err := tx.QueryRow(`
INSERT INTO learning_content_import_runs (app_code, bundle_id, target_lang, source, source_hash, mode, started_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id`, plan.AppCode, plan.BundleID, plan.TargetLang, "embedded-bundle", plan.SourceHash, mode, time.Now()).Scan(&id)
	return id, err
}

func contentHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func nullableString(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func intFromAny(v interface{}) interface{} {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	default:
		return nil
	}
}
