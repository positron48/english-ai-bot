package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// AdminScenarioRow is one scenario as shown in the admin list (joined with district/location codes).
type AdminScenarioRow struct {
	ConversationScenario
	DistrictCode string
	LevelCode    string
	LocationCode string
	Status       string
	SortOrder    int
}

// AdminScenarioInput holds the writable fields of a scenario for create/update.
type AdminScenarioInput struct {
	CourseCode  string
	CEFRLevel   string
	Code        string
	PlaceType   string
	Title       string
	NPCName     string
	NPCPersona  string
	SceneSetup  string
	IsQuest     bool
	MaxTurns    int
	TokenBudget int
	SortOrder   int
	Status      string
}

// AdminTaskInput holds the writable fields of a quest task.
type AdminTaskInput struct {
	Code               string
	Title              string
	CompletionCriteria string
	IsRequired         bool
	SortOrder          int
}

// ListScenariosForCourseAdmin returns every scenario of a course (any status) with district/location codes.
func (r *ConversationRepository) ListScenariosForCourseAdmin(ctx context.Context, courseID int64) ([]AdminScenarioRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cs.id, cs.course_id, cs.district_id, cs.location_id, cs.learning_item_id, cs.code,
		       cs.place_type, cs.cefr_level, cs.title, cs.npc_name, cs.npc_persona, cs.scene_setup,
		       cs.is_quest, cs.max_turns, cs.token_budget, cs.sort_order, cs.status,
		       COALESCE(d.code, ''), COALESCE(d.level_code, ''), COALESCE(l.code, '')
		FROM conversation_scenarios cs
		LEFT JOIN districts d ON d.id = cs.district_id
		LEFT JOIN locations l ON l.id = cs.location_id
		WHERE cs.course_id = ?
		ORDER BY cs.sort_order, cs.id`, courseID)
	if err != nil {
		return nil, fmt.Errorf("admin list scenarios: %w", err)
	}
	defer rows.Close()

	var out []AdminScenarioRow
	for rows.Next() {
		var s AdminScenarioRow
		if err := rows.Scan(
			&s.ID, &s.CourseID, &s.DistrictID, &s.LocationID, &s.LearningItemID, &s.Code,
			&s.PlaceType, &s.CEFRLevel, &s.Title, &s.NPCName, &s.NPCPersona, &s.SceneSetup,
			&s.IsQuest, &s.MaxTurns, &s.TokenBudget, &s.SortOrder, &s.Status,
			&s.DistrictCode, &s.LevelCode, &s.LocationCode,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CourseIDByCode resolves a course id from its code.
func (r *ConversationRepository) CourseIDByCode(ctx context.Context, code string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, code).Scan(&id)
	return id, err
}

// resolveDistrictLocation finds the district (by course + cefr level) and its 'conversation' location.
func (r *ConversationRepository) resolveDistrictLocation(ctx context.Context, q queryRower, courseID int64, cefrLevel string) (districtID, locationID int64, err error) {
	err = q.QueryRowContext(ctx, `SELECT id FROM districts WHERE course_id = ? AND level_code = ?`, courseID, cefrLevel).Scan(&districtID)
	if err != nil {
		return 0, 0, fmt.Errorf("district for level %q: %w", cefrLevel, err)
	}
	err = q.QueryRowContext(ctx, `SELECT id FROM locations WHERE district_id = ? AND code = 'conversation'`, districtID).Scan(&locationID)
	if err != nil {
		return 0, 0, fmt.Errorf("conversation location for level %q: %w", cefrLevel, err)
	}
	return districtID, locationID, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// CreateScenario inserts a new scenario (plus a backing learning_items row) and returns its id.
func (r *ConversationRepository) CreateScenario(ctx context.Context, in AdminScenarioInput) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	courseID, err := r.CourseIDByCode(ctx, in.CourseCode)
	if err != nil {
		return 0, fmt.Errorf("course %q: %w", in.CourseCode, err)
	}
	districtID, locationID, err := r.resolveDistrictLocation(ctx, tx, courseID, in.CEFRLevel)
	if err != nil {
		return 0, err
	}

	// Backing learning_items row so district/location progress aggregation lights up.
	var learningItemID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		VALUES (?, ?, ?, 'speaking_task', 'conversation_scenario', ?, ?, ?, 'published')
		ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET title = excluded.title, cefr_level = excluded.cefr_level
		RETURNING id`,
		courseID, districtID, locationID, in.Code, in.Title, in.CEFRLevel,
	).Scan(&learningItemID); err != nil {
		return 0, fmt.Errorf("backing learning_item: %w", err)
	}

	var id int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO conversation_scenarios
			(course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
			 title, npc_name, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		RETURNING id`,
		courseID, districtID, locationID, learningItemID, in.Code, in.PlaceType, in.CEFRLevel,
		in.Title, in.NPCName, in.NPCPersona, in.SceneSetup, in.IsQuest, in.MaxTurns, in.TokenBudget, in.SortOrder, in.Status,
	).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicateScenarioCode
		}
		return 0, fmt.Errorf("insert scenario: %w", err)
	}
	return id, tx.Commit()
}

// ErrDuplicateScenarioCode is returned when a scenario code already exists for the course.
var ErrDuplicateScenarioCode = errors.New("scenario code already exists for this course")

// UpdateScenario updates the writable fields of a scenario (re-resolving district/location from level).
func (r *ConversationRepository) UpdateScenario(ctx context.Context, id int64, in AdminScenarioInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var courseID int64
	if err := tx.QueryRowContext(ctx, `SELECT course_id FROM conversation_scenarios WHERE id = ?`, id).Scan(&courseID); err != nil {
		return err
	}
	districtID, locationID, err := r.resolveDistrictLocation(ctx, tx, courseID, in.CEFRLevel)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE conversation_scenarios SET
			district_id = ?, location_id = ?, code = ?, place_type = ?, cefr_level = ?,
			title = ?, npc_name = ?, npc_persona = ?, scene_setup = ?, is_quest = ?,
			max_turns = ?, token_budget = ?, sort_order = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		districtID, locationID, in.Code, in.PlaceType, in.CEFRLevel,
		in.Title, in.NPCName, in.NPCPersona, in.SceneSetup, in.IsQuest,
		in.MaxTurns, in.TokenBudget, in.SortOrder, in.Status, id,
	); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateScenarioCode
		}
		return fmt.Errorf("update scenario: %w", err)
	}

	// Keep the backing learning_item in sync (title/level/district/location).
	if _, err := tx.ExecContext(ctx, `
		UPDATE learning_items SET title = ?, cefr_level = ?, district_id = ?, location_id = ?
		WHERE course_id = ? AND source_kind = 'conversation_scenario' AND source_id = ?`,
		in.Title, in.CEFRLevel, districtID, locationID, courseID, in.Code,
	); err != nil {
		return fmt.Errorf("sync learning_item: %w", err)
	}
	return tx.Commit()
}

// DeleteScenario removes a scenario (cascades to tasks/sessions) and its backing learning_item.
func (r *ConversationRepository) DeleteScenario(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var courseID int64
	var code string
	if err := tx.QueryRowContext(ctx, `SELECT course_id, code FROM conversation_scenarios WHERE id = ?`, id).Scan(&courseID, &code); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM conversation_scenarios WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete scenario: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM learning_items
		WHERE course_id = ? AND source_kind = 'conversation_scenario' AND source_id = ?`,
		courseID, code,
	); err != nil {
		return fmt.Errorf("delete backing learning_item: %w", err)
	}
	return tx.Commit()
}

// CreateTask inserts a quest task for a scenario and returns its id.
func (r *ConversationRepository) CreateTask(ctx context.Context, scenarioID int64, in AdminTaskInput) (int64, error) {
	var id int64
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`,
		scenarioID, in.Code, in.SortOrder, in.IsRequired, in.Title, in.CompletionCriteria,
	).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicateTaskCode
		}
		return 0, fmt.Errorf("insert task: %w", err)
	}
	return id, nil
}

// ErrDuplicateTaskCode is returned when a task code already exists in the scenario.
var ErrDuplicateTaskCode = errors.New("task code already exists in this scenario")

// UpdateTask updates the writable fields of a task.
func (r *ConversationRepository) UpdateTask(ctx context.Context, id int64, in AdminTaskInput) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE conversation_tasks SET code = ?, sort_order = ?, is_required = ?, title = ?, completion_criteria = ?
		WHERE id = ?`,
		in.Code, in.SortOrder, in.IsRequired, in.Title, in.CompletionCriteria, id,
	); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateTaskCode
		}
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

// DeleteTask removes a quest task.
func (r *ConversationRepository) DeleteTask(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM conversation_tasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// ScenarioCourseID returns the course id owning a scenario (for ownership checks).
func (r *ConversationRepository) ScenarioCourseID(ctx context.Context, scenarioID int64) (int64, error) {
	var courseID int64
	err := r.db.QueryRowContext(ctx, `SELECT course_id FROM conversation_scenarios WHERE id = ?`, scenarioID).Scan(&courseID)
	return courseID, err
}

// TaskScenarioID returns the scenario id owning a task.
func (r *ConversationRepository) TaskScenarioID(ctx context.Context, taskID int64) (int64, error) {
	var scenarioID int64
	err := r.db.QueryRowContext(ctx, `SELECT scenario_id FROM conversation_tasks WHERE id = ?`, taskID).Scan(&scenarioID)
	return scenarioID, err
}

// CourseLevelOption is one selectable CEFR level (district) for a course.
type CourseLevelOption struct {
	LevelCode string
	Title     string
}

// ListCourseLevels returns the districts (CEFR levels) of a course that have a conversation location.
func (r *ConversationRepository) ListCourseLevels(ctx context.Context, courseID int64) ([]CourseLevelOption, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.level_code, d.title
		FROM districts d
		WHERE d.course_id = ?
		  AND EXISTS (SELECT 1 FROM locations l WHERE l.district_id = d.id AND l.code = 'conversation')
		ORDER BY d.sort_order, d.level_code`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CourseLevelOption
	for rows.Next() {
		var o CourseLevelOption
		if err := rows.Scan(&o.LevelCode, &o.Title); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// isUniqueViolation reports whether err is a Postgres unique-constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
