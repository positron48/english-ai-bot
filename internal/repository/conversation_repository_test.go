package repository

import (
	"context"
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// convFixture builds a user_course + scenario (with two required tasks and one optional) for es_ru.
type convFixture struct {
	userCourseID   int64
	scenarioID     int64
	learningItemID int64
	taskIDByCode   map[string]int64
}

func setupConversationFixture(t *testing.T, conn *sql.DB) convFixture {
	t.Helper()
	ctx := context.Background()
	logger := zap.NewNop()

	user, err := NewUserRepository(conn, logger).GetOrCreateUser(770001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	var courseID, districtID, locationID int64
	if err := conn.QueryRow(`SELECT id FROM courses WHERE code = 'es_ru'`).Scan(&courseID); err != nil {
		t.Fatalf("get course: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM districts WHERE course_id = ? AND level_code = 'A0'`, courseID).Scan(&districtID); err != nil {
		t.Fatalf("get district: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM locations WHERE district_id = ? AND code = 'conversation'`, districtID).Scan(&locationID); err != nil {
		t.Fatalf("get location: %v", err)
	}

	var userCourseID int64
	if err := conn.QueryRow(`
		INSERT INTO user_courses (user_id, course_id, status) VALUES (?, ?, 'active') RETURNING id`,
		user.ID, courseID).Scan(&userCourseID); err != nil {
		t.Fatalf("insert user_course: %v", err)
	}

	var learningItemID int64
	if err := conn.QueryRow(`
		INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		VALUES (?, ?, ?, 'speaking_task', 'conversation_scenario', 'test_cafe', 'Test Cafe', 'A0', 'published')
		RETURNING id`, courseID, districtID, locationID).Scan(&learningItemID); err != nil {
		t.Fatalf("insert learning_item: %v", err)
	}

	var scenarioID int64
	if err := conn.QueryRow(`
		INSERT INTO conversation_scenarios
			(course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
			 title, npc_name, npc_persona, scene_setup, is_quest, max_turns, status)
		VALUES (?, ?, ?, ?, 'test_cafe', 'cafe', 'A0', 'Тест кафе', 'Mara', 'barista', 'cafe', true, 16, 'active')
		RETURNING id`, courseID, districtID, locationID, learningItemID).Scan(&scenarioID); err != nil {
		t.Fatalf("insert scenario: %v", err)
	}

	taskIDByCode := map[string]int64{}
	for _, tk := range []struct {
		code     string
		order    int
		required bool
	}{
		{"greet", 0, true},
		{"order", 1, true},
		{"thank", 2, false},
	} {
		var id int64
		if err := conn.QueryRow(`
			INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
			VALUES (?, ?, ?, ?, ?, 'criteria')
			RETURNING id`, scenarioID, tk.code, tk.order, tk.required, tk.code).Scan(&id); err != nil {
			t.Fatalf("insert task %s: %v", tk.code, err)
		}
		taskIDByCode[tk.code] = id
	}

	_ = ctx
	return convFixture{userCourseID: userCourseID, scenarioID: scenarioID, learningItemID: learningItemID, taskIDByCode: taskIDByCode}
}

func TestConversationRepository_SessionLifecycle(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	// StartSession creates, then resumes the same open session.
	s1, created1, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil || !created1 {
		t.Fatalf("StartSession create: created=%v err=%v", created1, err)
	}
	s2, created2, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil || created2 {
		t.Fatalf("StartSession resume: created=%v err=%v", created2, err)
	}
	if s1.ID != s2.ID {
		t.Fatalf("resume returned different session: %d vs %d", s1.ID, s2.ID)
	}

	// Messages: seq is monotonic, ListMessages ordered.
	seq, _ := repo.NextSeq(ctx, s1.ID)
	if seq != 1 {
		t.Fatalf("first seq = %d, want 1", seq)
	}
	if err := repo.AppendMessage(ctx, s1.ID, 1, "assistant", "¡Hola!", 10, 5); err != nil {
		t.Fatalf("append assistant: %v", err)
	}
	if err := repo.AppendMessage(ctx, s1.ID, 2, "user", "Hola", 0, 0); err != nil {
		t.Fatalf("append user: %v", err)
	}
	seq, _ = repo.NextSeq(ctx, s1.ID)
	if seq != 3 {
		t.Fatalf("next seq = %d, want 3", seq)
	}
	msgs, err := repo.ListMessages(ctx, s1.ID)
	if err != nil || len(msgs) != 2 || msgs[0].Role != "assistant" || msgs[1].Content != "Hola" {
		t.Fatalf("ListMessages = %+v err=%v", msgs, err)
	}
}

func TestConversationRepository_MonotonicTaskCompletion(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	session, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Mark greet at seq 2; unknown codes ignored.
	if err := repo.MarkTasksCompleted(ctx, session.ID, fx.taskIDByCode, []string{"greet", "bogus"}, 2); err != nil {
		t.Fatalf("mark greet: %v", err)
	}
	completed, _ := repo.GetCompletedTaskIDs(ctx, session.ID)
	if !completed[fx.taskIDByCode["greet"]] || len(completed) != 1 {
		t.Fatalf("after greet completed = %v", completed)
	}

	// Re-marking greet at a later seq keeps the original completed_in_seq (monotonic).
	if err := repo.MarkTasksCompleted(ctx, session.ID, fx.taskIDByCode, []string{"greet", "order"}, 6); err != nil {
		t.Fatalf("mark again: %v", err)
	}
	var greetSeq int
	if err := conn.QueryRow(`SELECT completed_in_seq FROM conversation_task_progress WHERE session_id = ? AND task_id = ?`,
		session.ID, fx.taskIDByCode["greet"]).Scan(&greetSeq); err != nil {
		t.Fatalf("read greet seq: %v", err)
	}
	if greetSeq != 2 {
		t.Fatalf("greet completed_in_seq = %d, want 2 (monotonic)", greetSeq)
	}
	completed, _ = repo.GetCompletedTaskIDs(ctx, session.ID)
	if len(completed) != 2 {
		t.Fatalf("expected 2 completed, got %v", completed)
	}
}

func TestConversationRepository_RecordQuestCompletionIdempotent(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)
	session, _, _ := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)

	li := sql.NullInt64{Int64: fx.learningItemID, Valid: true}
	if err := repo.RecordQuestCompletion(ctx, fx.userCourseID, li, "test_cafe", session.ID); err != nil {
		t.Fatalf("record 1: %v", err)
	}
	if err := repo.RecordQuestCompletion(ctx, fx.userCourseID, li, "test_cafe", session.ID); err != nil {
		t.Fatalf("record 2: %v", err)
	}

	var attempts int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE user_course_id = ? AND learning_item_id = ?`,
		fx.userCourseID, fx.learningItemID).Scan(&attempts); err != nil {
		t.Fatalf("count attempts: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("exercise_attempts = %d, want 1 (idempotent)", attempts)
	}
}
