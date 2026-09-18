package repository

import (
	"context"
	"fmt"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestQuestListQueriesAreBatched(t *testing.T) {
	for _, kind := range []string{"conversation", "picture_quest"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			ids := make([]int64, 30)
			for i := range ids {
				ids[i] = int64(i + 1)
			}
			// Regardless of list size: one task query and one latest-progress query.
			mock.ExpectQuery("SELECT id, .* FROM " + kind + "_tasks WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "parent", "code", "sort_order", "is_required", "title", "completion_criteria"}).AddRow(10, 1, "a", 0, true, "Task", "done"))
			mock.ExpectQuery("SELECT latest.parent_id, latest.status, p.task_id FROM").WillReturnRows(sqlmock.NewRows([]string{"parent_id", "status", "task_id"}).AddRow(1, "active", 10).AddRow(1, "active", 11).AddRow(2, "active", nil))
			var progress map[int64]QuestListProgress
			if kind == "conversation" {
				repo := NewConversationRepository(db, zap.NewNop())
				tasks, e := repo.ListTasksForList(context.Background(), ids)
				if e != nil || len(tasks[1]) != 1 {
					t.Fatalf("%v %v", tasks, e)
				}
				progress, err = repo.ListProgress(context.Background(), 77, ids)
			} else {
				repo := NewPictureQuestRepository(db, zap.NewNop())
				tasks, e := repo.ListTasksForList(context.Background(), ids)
				if e != nil || len(tasks[1]) != 1 {
					t.Fatalf("%v %v", tasks, e)
				}
				progress, err = repo.ListProgress(context.Background(), 77, ids)
			}
			if err != nil || len(progress[1].Completed) != 2 || len(progress[2].Completed) != 0 {
				t.Fatalf("%v %v", progress, err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQuestListProgressUsesLatestSessionAndUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	fx := setupConversationFixture(t, db)
	ctx := context.Background()
	repo := NewConversationRepository(db, zap.NewNop())
	first, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.MarkTasksCompleted(ctx, first.ID, fx.taskIDByCode, []string{"greet"}, 1); err != nil {
		t.Fatal(err)
	}
	// A newer run must not inherit completed tasks from an older run.
	if _, err = db.Exec(`UPDATE conversation_sessions SET status = 'completed' WHERE id = ?`, first.ID); err != nil {
		t.Fatal(err)
	}
	second, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.MarkTasksCompleted(ctx, second.ID, fx.taskIDByCode, []string{"order"}, 1); err != nil {
		t.Fatal(err)
	}
	progress, err := repo.ListProgress(ctx, fx.userCourseID, []int64{fx.scenarioID})
	if err != nil {
		t.Fatal(err)
	}
	got := progress[fx.scenarioID]
	if got.Status != "open" || got.Completed[fx.taskIDByCode["greet"]] || !got.Completed[fx.taskIDByCode["order"]] {
		t.Fatal(fmt.Sprint(got))
	}
	other, err := repo.ListProgress(ctx, fx.userCourseID+1000, []int64{fx.scenarioID})
	if err != nil || len(other) != 0 {
		t.Fatalf("other user: %v %v", other, err)
	}
}
