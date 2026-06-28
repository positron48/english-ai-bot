package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const convLinglowGapUserBase int64 = 900001

func convLinglowGapUser(t *testing.T, conn *sql.DB, telegramID int64) int64 {
	t.Helper()
	user, err := NewUserRepository(conn, zap.NewNop()).GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user %d: %v", telegramID, err)
	}
	return user.ID
}

func convLinglowGapEnsureUserCourse(t *testing.T, conn *sql.DB, userID int64, lc config.LearningConfig) {
	t.Helper()
	if _, err := NewCourseRepository(conn, zap.NewNop()).BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
}

func TestConversationLinglowGap_AdminScenarioQueries(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)
	courseID, err := repo.ScenarioCourseID(ctx, fx.scenarioID)
	if err != nil || courseID == 0 {
		t.Fatalf("ScenarioCourseID: %v", err)
	}
	scenarioID, err := repo.TaskScenarioID(ctx, fx.taskIDByCode["greet"])
	if err != nil || scenarioID != fx.scenarioID {
		t.Fatalf("TaskScenarioID: %v", err)
	}
	levels, err := repo.ListCourseLevels(ctx, courseID)
	if err != nil || len(levels) == 0 {
		t.Fatalf("ListCourseLevels: %v", err)
	}
}

func TestConversationLinglowGap_RecordReadingCompleted(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	userID := convLinglowGapUser(t, conn, convLinglowGapUserBase)
	convLinglowGapEnsureUserCourse(t, conn, userID, lc)
	if _, err := conn.Exec(`
		INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		VALUES ('gap.reading.one', 'gap-cat', 'Gap Reading', 'A0', 'es', 'Hola')
	`); err != nil {
		t.Fatalf("insert reading text: %v", err)
	}
	repo := NewLinglowEventRepository(conn)
	input := ReadingCompletedInput{UserID: userID, CourseCode: "es_ru", ChapterID: "gap.reading.one"}
	id, err := repo.RecordReadingCompleted(ctx, input)
	if err != nil || id == 0 {
		t.Fatalf("RecordReadingCompleted: %v", err)
	}
	second, err := repo.RecordReadingCompleted(ctx, input)
	if err != nil || second != id {
		t.Fatalf("idempotent reading: %v", err)
	}
	if _, err := repo.RecordReadingCompleted(ctx, ReadingCompletedInput{CourseCode: "es_ru", ChapterID: "x"}); err == nil {
		t.Fatal("expected empty user error")
	}
}

func TestConversationLinglowGap_RecordChatMessage(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	userID := convLinglowGapUser(t, conn, convLinglowGapUserBase+1)
	convLinglowGapEnsureUserCourse(t, conn, userID, lc)
	repo := NewLinglowEventRepository(conn)
	if err := repo.RecordChatMessage(ctx, ChatMessageInput{UserID: userID, CourseCode: "es_ru", MessageLen: 42}); err != nil {
		t.Fatalf("RecordChatMessage: %v", err)
	}
	if err := repo.RecordChatMessage(ctx, ChatMessageInput{CourseCode: "es_ru"}); err == nil {
		t.Fatal("expected empty user error")
	}
}

func TestConversationLinglowGap_AdminCRUDBranches(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	courseID, _ := repo.CourseIDByCode(ctx, "es_ru")

	if _, err := conn.Exec(`INSERT INTO districts (course_id, code, level_code, title, sort_order) VALUES (?, 'gap-no-conv', 'GAPNC', 'No Conv', 999)`, courseID); err != nil {
		t.Fatalf("insert district: %v", err)
	}
	if _, err := repo.CreateScenario(ctx, AdminScenarioInput{
		CourseCode: "es_ru", CEFRLevel: "GAPNC", Code: "gap_no_conv", PlaceType: "cafe",
		Title: "X", NPCName: "N", NPCPersona: "p", SceneSetup: "s", Status: "draft",
	}); err == nil {
		t.Fatal("expected missing conversation location")
	}

	id, err := repo.CreateScenario(ctx, AdminScenarioInput{
		CourseCode: "es_ru", CEFRLevel: "A0", Code: "gap_task_dup", PlaceType: "cafe",
		Title: "X", NPCName: "N", NPCPersona: "p", SceneSetup: "s", Status: "active",
	})
	if err != nil {
		t.Fatalf("CreateScenario: %v", err)
	}
	if _, err := repo.CreateTask(ctx, id, AdminTaskInput{Code: "a", Title: "A", CompletionCriteria: "a", SortOrder: 0}); err != nil {
		t.Fatalf("CreateTask a: %v", err)
	}
	taskB, err := repo.CreateTask(ctx, id, AdminTaskInput{Code: "b", Title: "B", CompletionCriteria: "b", SortOrder: 1})
	if err != nil {
		t.Fatalf("CreateTask b: %v", err)
	}
	if err := repo.UpdateTask(ctx, taskB, AdminTaskInput{Code: "a", Title: "dup", CompletionCriteria: "x", SortOrder: 1}); !errors.Is(err, ErrDuplicateTaskCode) {
		t.Fatalf("UpdateTask duplicate: %v", err)
	}
	_ = repo.DeleteScenario(ctx, id)
}

func TestConversationLinglowGap_GrammarSRSSyncAndPrune(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	userID := convLinglowGapUser(t, conn, convLinglowGapUserBase+2)
	courseRepo := NewCourseRepository(conn, zap.NewNop())
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("backfill user courses: %v", err)
	}
	insertGrammarTheoryBlockContent(t, conn, "gap.section.sync", "gap.chapter.sync", "block.sync", "concept.sync")
	insertGrammarTheoryBlockContent(t, conn, "gap.section.prune", "gap.chapter.prune", "block.prune", "concept.prune")
	if _, err := courseRepo.MapLegacyContentForLearning(ctx, lc); err != nil {
		t.Fatalf("map content: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at) VALUES (?, 'es', 'es', 'gap.chapter.sync', 'block.sync', 'concept.sync', 'review', CURRENT_TIMESTAMP)`, userID); err != nil {
		t.Fatalf("insert sync memory: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at) VALUES (?, 'es', 'es', 'gap.chapter.prune', 'block.prune', 'concept.prune', 'review', CURRENT_TIMESTAMP)`, userID); err != nil {
		t.Fatalf("insert prune memory: %v", err)
	}
	repo := NewLinglowGrammarSRSBackfillRepository(conn)
	if err := repo.SyncTheoryBlockForUser(ctx, lc, userID, "gap.chapter.sync", "block.sync"); err != nil {
		t.Fatalf("SyncTheoryBlockForUser: %v", err)
	}
	if err := repo.SyncTheoryBlockForUser(ctx, lc, 0, "ch", "blk"); err != nil {
		t.Fatalf("SyncTheoryBlockForUser empty user: %v", err)
	}
	if _, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true}); err != nil {
		t.Fatalf("backfill commit: %v", err)
	}
	if _, err := conn.Exec(`DELETE FROM grammar_theory_memory WHERE user_id = ? AND chapter_id = 'gap.chapter.prune'`, userID); err != nil {
		t.Fatalf("delete prune memory: %v", err)
	}
	summary, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true, Resync: true, PruneOrphans: true})
	if err != nil {
		t.Fatalf("backfill prune: %v", err)
	}
	if summary.PrunedOrphans < 1 {
		t.Fatalf("expected pruned orphans: %+v", summary)
	}
}

func TestConversationLinglowGap_MockErrors(t *testing.T) {
	ctx := context.Background()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewLinglowGrammarSRSBackfillRepository(db)
	mock.ExpectExec("UPDATE srs_items").WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected failed")))
	if _, err := repo.pruneOrphanGrammarSRSItems(ctx, "es_ru", "es", "es"); err == nil {
		t.Fatal("expected prune error")
	}
	if !isUniqueViolation(fmt.Errorf("duplicate key")) {
		t.Fatal("isUniqueViolation")
	}
	_ = lc
	_ = errors.New
}
