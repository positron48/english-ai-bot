package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// QuestListProgress is the latest run and completed tasks for one list item.
type QuestListProgress struct {
	Status    string
	Completed map[int64]bool
}

func questIDArgs(ids []int64) (string, []interface{}) {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(ids)), ","), args
}

// Table and column identifiers come only from the two fixed wrappers below.
func loadQuestListProgress(ctx context.Context, db *sql.DB, prefix, parentColumn string, userCourseID int64, ids []int64) (map[int64]QuestListProgress, error) {
	out := make(map[int64]QuestListProgress)
	if len(ids) == 0 {
		return out, nil
	}
	placeholders, args := questIDArgs(ids)
	args = append([]interface{}{userCourseID}, args...)
	query := fmt.Sprintf(`SELECT latest.parent_id, latest.status, p.task_id FROM (
 SELECT DISTINCT ON (%s) id, %s AS parent_id, status
 FROM %s_sessions WHERE user_course_id = ? AND %s IN (%s)
 ORDER BY %s, started_at DESC, id DESC
 ) latest LEFT JOIN %s_task_progress p ON p.session_id = latest.id AND p.completed = true`, parentColumn, parentColumn, prefix, parentColumn, placeholders, parentColumn, prefix)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var status string
		var task sql.NullInt64
		if err := rows.Scan(&id, &status, &task); err != nil {
			return nil, err
		}
		progress, ok := out[id]
		if !ok {
			progress = QuestListProgress{Status: status, Completed: make(map[int64]bool)}
		}
		if task.Valid {
			progress.Completed[task.Int64] = true
		}
		out[id] = progress
	}
	return out, rows.Err()
}

func (r *ConversationRepository) ListProgress(ctx context.Context, userCourseID int64, ids []int64) (map[int64]QuestListProgress, error) {
	return loadQuestListProgress(ctx, r.db, "conversation", "scenario_id", userCourseID, ids)
}
func (r *PictureQuestRepository) ListProgress(ctx context.Context, userCourseID int64, ids []int64) (map[int64]QuestListProgress, error) {
	return loadQuestListProgress(ctx, r.db, "picture_quest", "quest_id", userCourseID, ids)
}
