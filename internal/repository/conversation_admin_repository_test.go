package repository

import (
	"context"
	"errors"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestConversationRepository_AdminScenarioCRUD(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())

	courseID, err := repo.CourseIDByCode(ctx, "es_ru")
	if err != nil {
		t.Fatalf("course id: %v", err)
	}

	in := AdminScenarioInput{
		CourseCode: "es_ru", CEFRLevel: "A0", Code: "admin_test_cafe", PlaceType: "cafe",
		Title: "Админ кафе", NPCName: "Mara", NPCPersona: "barista", SceneSetup: "cafe",
		IsQuest: true, MaxTurns: 12, TokenBudget: 5000, SortOrder: 7, Status: "active",
	}
	id, err := repo.CreateScenario(ctx, in)
	if err != nil {
		t.Fatalf("CreateScenario: %v", err)
	}

	// Duplicate code rejected.
	if _, err := repo.CreateScenario(ctx, in); !errors.Is(err, ErrDuplicateScenarioCode) {
		t.Fatalf("duplicate scenario: want ErrDuplicateScenarioCode, got %v", err)
	}

	// Backing learning_item created.
	var liCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_items WHERE course_id = ? AND source_kind = 'conversation_scenario' AND source_id = ?`, courseID, in.Code).Scan(&liCount); err != nil || liCount != 1 {
		t.Fatalf("backing learning_item count = %d err=%v", liCount, err)
	}

	rows, err := repo.ListScenariosForCourseAdmin(ctx, courseID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found *AdminScenarioRow
	for i := range rows {
		if rows[i].ID == id {
			found = &rows[i]
		}
	}
	if found == nil {
		t.Fatal("created scenario not in admin list")
	}
	if found.LocationCode != "conversation" || found.LevelCode != "A0" {
		t.Fatalf("joined codes wrong: level=%q loc=%q", found.LevelCode, found.LocationCode)
	}

	// Update.
	in.Title = "Обновлено"
	in.MaxTurns = 30
	if err := repo.UpdateScenario(ctx, id, in); err != nil {
		t.Fatalf("UpdateScenario: %v", err)
	}
	sc, err := repo.GetScenarioByCode(ctx, courseID, in.Code)
	if err != nil || sc == nil || sc.Title != "Обновлено" || sc.MaxTurns != 30 {
		t.Fatalf("update not applied: %+v err=%v", sc, err)
	}

	// Tasks CRUD.
	taskID, err := repo.CreateTask(ctx, id, AdminTaskInput{Code: "greet", Title: "Поздороваться", CompletionCriteria: "greets", IsRequired: true, SortOrder: 0})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := repo.CreateTask(ctx, id, AdminTaskInput{Code: "greet", Title: "x", CompletionCriteria: "y"}); !errors.Is(err, ErrDuplicateTaskCode) {
		t.Fatalf("duplicate task: want ErrDuplicateTaskCode, got %v", err)
	}
	if err := repo.UpdateTask(ctx, taskID, AdminTaskInput{Code: "greet", Title: "Привет", CompletionCriteria: "greets warmly", IsRequired: false, SortOrder: 1}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	tasks, _ := repo.ListTasks(ctx, id)
	if len(tasks) != 1 || tasks[0].Title != "Привет" || tasks[0].IsRequired {
		t.Fatalf("task update not applied: %+v", tasks)
	}
	if err := repo.DeleteTask(ctx, taskID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	// Delete scenario also removes backing learning_item.
	if err := repo.DeleteScenario(ctx, id); err != nil {
		t.Fatalf("DeleteScenario: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_items WHERE course_id = ? AND source_kind = 'conversation_scenario' AND source_id = ?`, courseID, in.Code).Scan(&liCount); err != nil || liCount != 0 {
		t.Fatalf("learning_item not deleted: count=%d err=%v", liCount, err)
	}
}

func TestConversationRepository_ImportScenario(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	courseID, _ := repo.CourseIDByCode(ctx, "es_ru")

	in := AdminScenarioInput{
		CourseCode: "es_ru", CEFRLevel: "A0", Code: "import_cafe", PlaceType: "cafe",
		Title: "Импорт кафе", NPCName: "Mara", NPCPersona: "barista", SceneSetup: "cafe",
		IsQuest: true, MaxTurns: 16, TokenBudget: 6000, SortOrder: 0, Status: "active",
		NPCCode: "mara", PrerequisiteCode: "",
	}
	tasks := []AdminTaskInput{
		{Code: "greet", Title: "Поздороваться", CompletionCriteria: "greets", IsRequired: true, SortOrder: 0},
		{Code: "order", Title: "Заказать", CompletionCriteria: "orders coffee", IsRequired: true, SortOrder: 1},
	}

	// First import creates the scenario + tasks and the npc_code/prerequisite columns persist.
	id, created, err := repo.ImportScenario(ctx, in, tasks)
	if err != nil || !created {
		t.Fatalf("import create: id=%d created=%v err=%v", id, created, err)
	}
	sc, err := repo.GetScenarioByCode(ctx, courseID, in.Code)
	if err != nil || sc == nil || sc.NPCCode != "mara" {
		t.Fatalf("scenario not imported with npc_code: %+v err=%v", sc, err)
	}
	if got, _ := repo.ListTasks(ctx, id); len(got) != 2 {
		t.Fatalf("want 2 tasks, got %d", len(got))
	}

	// Re-import with fewer tasks updates in place and REPLACES the task set.
	in.Title = "Импорт кафе v2"
	in.PrerequisiteCode = "import_intro"
	id2, created2, err := repo.ImportScenario(ctx, in, tasks[:1])
	if err != nil || created2 || id2 != id {
		t.Fatalf("import update: id=%d (want %d) created=%v err=%v", id2, id, created2, err)
	}
	sc, _ = repo.GetScenarioByCode(ctx, courseID, in.Code)
	if sc == nil || sc.Title != "Импорт кафе v2" || sc.PrerequisiteCode != "import_intro" {
		t.Fatalf("re-import did not update fields: %+v", sc)
	}
	if got, _ := repo.ListTasks(ctx, id); len(got) != 1 || got[0].Code != "greet" {
		t.Fatalf("tasks not replaced on re-import: %+v", got)
	}
}

func TestConversationRepository_LatestCompletedScenarioCodes(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)
	courseID, _ := repo.CourseIDByCode(ctx, "es_ru")

	// No completed sessions yet.
	done, err := repo.LatestCompletedScenarioCodes(ctx, fx.userCourseID, courseID)
	if err != nil || len(done) != 0 {
		t.Fatalf("expected none completed: %v err=%v", done, err)
	}

	s, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if err := repo.CloseSession(ctx, s.ID, fx.userCourseID, "completed"); err != nil {
		t.Fatalf("CloseSession: %v", err)
	}

	done, err = repo.LatestCompletedScenarioCodes(ctx, fx.userCourseID, courseID)
	if err != nil || !done["test_cafe"] {
		t.Fatalf("completed code not reported: %v err=%v", done, err)
	}
}

func TestConversationRepository_MessageCorrections(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	s, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	corrections := `[{"original":"yo gusta","corrected":"me gusta","explanation":"возвратное местоимение"}]`
	if err := repo.AppendMessageWithCorrections(ctx, s.ID, 1, "assistant", "¡Hola!", 0, 0, corrections); err != nil {
		t.Fatalf("append with corrections: %v", err)
	}
	// Plain AppendMessage defaults to "[]".
	if err := repo.AppendMessage(ctx, s.ID, 2, "user", "yo gusta", 0, 0); err != nil {
		t.Fatalf("append user: %v", err)
	}

	msgs, err := repo.ListMessages(ctx, s.ID)
	if err != nil || len(msgs) != 2 {
		t.Fatalf("list messages: n=%d err=%v", len(msgs), err)
	}
	if msgs[0].CorrectionsJSON == "" || msgs[0].CorrectionsJSON == "[]" {
		t.Fatalf("assistant corrections lost: %q", msgs[0].CorrectionsJSON)
	}
	if msgs[1].CorrectionsJSON != "[]" {
		t.Fatalf("user corrections should be empty, got %q", msgs[1].CorrectionsJSON)
	}
}
