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
