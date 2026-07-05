package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// AdminPictureQuestRow is one quest as shown in the admin list (joined with district codes).
type AdminPictureQuestRow struct {
	PictureQuest
	DistrictCode string
	LevelCode    string
	Status       string
	SortOrder    int
}

// AdminPictureQuestInput holds the writable fields of a picture quest for create/update.
type AdminPictureQuestInput struct {
	CourseCode       string
	CEFRLevel        string
	Code             string
	Title            string
	ImageURL         string
	ImageDescription string
	MaxTurns         int
	TokenBudget      int
	SortOrder        int
	Status           string
}

// AdminPictureTaskInput holds the writable fields of a picture quest task.
type AdminPictureTaskInput struct {
	Code               string
	Title              string
	CompletionCriteria string
	IsRequired         bool
	SortOrder          int
}

// ErrDuplicatePictureQuestCode is returned when a quest code already exists for the course.
var ErrDuplicatePictureQuestCode = errors.New("picture quest code already exists for this course")

// ErrDuplicatePictureTaskCode is returned when a task code already exists in the quest.
var ErrDuplicatePictureTaskCode = errors.New("task code already exists in this picture quest")

// ListQuestsForCourseAdmin returns every picture quest of a course (any status) with district codes.
func (r *PictureQuestRepository) ListQuestsForCourseAdmin(ctx context.Context, courseID int64) ([]AdminPictureQuestRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT q.id, q.course_id, q.district_id, q.location_id, q.learning_item_id, q.code,
		       q.cefr_level, q.title, q.image_url, q.image_description, q.max_turns, q.token_budget,
		       q.sort_order, q.status,
		       COALESCE(d.code, ''), COALESCE(d.level_code, '')
		FROM picture_quests q
		LEFT JOIN districts d ON d.id = q.district_id
		WHERE q.course_id = ?
		ORDER BY q.sort_order, q.id`, courseID)
	if err != nil {
		return nil, fmt.Errorf("admin list picture quests: %w", err)
	}
	defer rows.Close()

	var out []AdminPictureQuestRow
	for rows.Next() {
		var q AdminPictureQuestRow
		if err := rows.Scan(
			&q.ID, &q.CourseID, &q.DistrictID, &q.LocationID, &q.LearningItemID, &q.Code,
			&q.CEFRLevel, &q.Title, &q.ImageURL, &q.ImageDescription, &q.MaxTurns, &q.TokenBudget,
			&q.SortOrder, &q.Status,
			&q.DistrictCode, &q.LevelCode,
		); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// CourseIDByCode resolves a course id from its code.
func (r *PictureQuestRepository) CourseIDByCode(ctx context.Context, code string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, code).Scan(&id)
	return id, err
}

// resolveDistrictLocation finds the district (by course + cefr level) and its 'conversation' location.
func (r *PictureQuestRepository) resolveDistrictLocation(ctx context.Context, q queryRower, courseID int64, cefrLevel string) (districtID, locationID int64, err error) {
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

// CreateQuest inserts a new picture quest (plus a backing learning_items row) and returns its id.
func (r *PictureQuestRepository) CreateQuest(ctx context.Context, in AdminPictureQuestInput) (int64, error) {
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
		VALUES (?, ?, ?, 'speaking_task', 'picture_quest', ?, ?, ?, 'published')
		ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET title = excluded.title, cefr_level = excluded.cefr_level
		RETURNING id`,
		courseID, districtID, locationID, in.Code, in.Title, in.CEFRLevel,
	).Scan(&learningItemID); err != nil {
		return 0, fmt.Errorf("backing learning_item: %w", err)
	}

	var id int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO picture_quests
			(course_id, district_id, location_id, learning_item_id, code, cefr_level,
			 title, image_url, image_description, max_turns, token_budget, sort_order, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		RETURNING id`,
		courseID, districtID, locationID, learningItemID, in.Code, in.CEFRLevel,
		in.Title, in.ImageURL, in.ImageDescription, in.MaxTurns, in.TokenBudget, in.SortOrder, in.Status,
	).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicatePictureQuestCode
		}
		return 0, fmt.Errorf("insert picture quest: %w", err)
	}
	return id, tx.Commit()
}

// UpdateQuest updates the writable fields of a quest (re-resolving district/location from level).
func (r *PictureQuestRepository) UpdateQuest(ctx context.Context, id int64, in AdminPictureQuestInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var courseID int64
	if err := tx.QueryRowContext(ctx, `SELECT course_id FROM picture_quests WHERE id = ?`, id).Scan(&courseID); err != nil {
		return err
	}
	districtID, locationID, err := r.resolveDistrictLocation(ctx, tx, courseID, in.CEFRLevel)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE picture_quests SET
			district_id = ?, location_id = ?, code = ?, cefr_level = ?,
			title = ?, image_url = ?, image_description = ?,
			max_turns = ?, token_budget = ?, sort_order = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		districtID, locationID, in.Code, in.CEFRLevel,
		in.Title, in.ImageURL, in.ImageDescription,
		in.MaxTurns, in.TokenBudget, in.SortOrder, in.Status, id,
	); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicatePictureQuestCode
		}
		return fmt.Errorf("update picture quest: %w", err)
	}

	// Keep the backing learning_item in sync (title/level/district/location).
	if _, err := tx.ExecContext(ctx, `
		UPDATE learning_items SET title = ?, cefr_level = ?, district_id = ?, location_id = ?
		WHERE course_id = ? AND source_kind = 'picture_quest' AND source_id = ?`,
		in.Title, in.CEFRLevel, districtID, locationID, courseID, in.Code,
	); err != nil {
		return fmt.Errorf("sync learning_item: %w", err)
	}
	return tx.Commit()
}

// DeleteQuest removes a quest (cascades to tasks/sessions) and its backing learning_item.
func (r *PictureQuestRepository) DeleteQuest(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var courseID int64
	var code string
	if err := tx.QueryRowContext(ctx, `SELECT course_id, code FROM picture_quests WHERE id = ?`, id).Scan(&courseID, &code); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM picture_quests WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete picture quest: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM learning_items
		WHERE course_id = ? AND source_kind = 'picture_quest' AND source_id = ?`,
		courseID, code,
	); err != nil {
		return fmt.Errorf("delete backing learning_item: %w", err)
	}
	return tx.Commit()
}

// ImportQuest upserts a quest (by course+code) together with its full task list. Existing tasks of
// the quest are replaced by the provided set. Returns the quest id and whether it was newly created.
// Used by the admin JSON import.
func (r *PictureQuestRepository) ImportQuest(ctx context.Context, in AdminPictureQuestInput, tasks []AdminPictureTaskInput) (id int64, created bool, err error) {
	courseID, err := r.CourseIDByCode(ctx, in.CourseCode)
	if err != nil {
		return 0, false, fmt.Errorf("course %q: %w", in.CourseCode, err)
	}

	var existingID int64
	scanErr := r.db.QueryRowContext(ctx,
		`SELECT id FROM picture_quests WHERE course_id = ? AND code = ?`, courseID, in.Code).Scan(&existingID)
	switch {
	case scanErr == nil:
		in = r.mergeImportQuestFields(ctx, existingID, in)
		if err := r.UpdateQuest(ctx, existingID, in); err != nil {
			return 0, false, err
		}
		id = existingID
	case errors.Is(scanErr, sql.ErrNoRows):
		newID, err := r.CreateQuest(ctx, in)
		if err != nil {
			return 0, false, err
		}
		id, created = newID, true
	default:
		return 0, false, fmt.Errorf("lookup picture quest: %w", scanErr)
	}

	if err := r.replaceQuestTasks(ctx, id, tasks); err != nil {
		return 0, false, err
	}
	return id, created, nil
}

// mergeImportQuestFields keeps admin-uploaded media when JSON import leaves image_url empty.
func (r *PictureQuestRepository) mergeImportQuestFields(ctx context.Context, existingID int64, in AdminPictureQuestInput) AdminPictureQuestInput {
	if strings.TrimSpace(in.ImageURL) != "" {
		return in
	}
	var existingURL string
	if err := r.db.QueryRowContext(ctx, `SELECT image_url FROM picture_quests WHERE id = ?`, existingID).Scan(&existingURL); err != nil {
		return in
	}
	if strings.TrimSpace(existingURL) == "" {
		return in
	}
	in.ImageURL = existingURL
	return in
}

// replaceQuestTasks deletes the quest's existing tasks and inserts the provided set in one tx.
func (r *PictureQuestRepository) replaceQuestTasks(ctx context.Context, questID int64, tasks []AdminPictureTaskInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM picture_quest_tasks WHERE quest_id = ?`, questID); err != nil {
		return fmt.Errorf("clear tasks: %w", err)
	}
	for _, t := range tasks {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO picture_quest_tasks (quest_id, code, sort_order, is_required, title, completion_criteria)
			VALUES (?, ?, ?, ?, ?, ?)`,
			questID, t.Code, t.SortOrder, t.IsRequired, t.Title, t.CompletionCriteria,
		); err != nil {
			if isUniqueViolation(err) {
				return ErrDuplicatePictureTaskCode
			}
			return fmt.Errorf("insert task %q: %w", t.Code, err)
		}
	}
	return tx.Commit()
}

// CreateTask inserts a quest task and returns its id.
func (r *PictureQuestRepository) CreateTask(ctx context.Context, questID int64, in AdminPictureTaskInput) (int64, error) {
	var id int64
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO picture_quest_tasks (quest_id, code, sort_order, is_required, title, completion_criteria)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`,
		questID, in.Code, in.SortOrder, in.IsRequired, in.Title, in.CompletionCriteria,
	).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDuplicatePictureTaskCode
		}
		return 0, fmt.Errorf("insert task: %w", err)
	}
	return id, nil
}

// UpdateTask updates the writable fields of a task.
func (r *PictureQuestRepository) UpdateTask(ctx context.Context, id int64, in AdminPictureTaskInput) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE picture_quest_tasks SET code = ?, sort_order = ?, is_required = ?, title = ?, completion_criteria = ?
		WHERE id = ?`,
		in.Code, in.SortOrder, in.IsRequired, in.Title, in.CompletionCriteria, id,
	); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicatePictureTaskCode
		}
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

// DeleteTask removes a quest task.
func (r *PictureQuestRepository) DeleteTask(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM picture_quest_tasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// QuestCourseID returns the course id owning a quest (for ownership checks).
func (r *PictureQuestRepository) QuestCourseID(ctx context.Context, questID int64) (int64, error) {
	var courseID int64
	err := r.db.QueryRowContext(ctx, `SELECT course_id FROM picture_quests WHERE id = ?`, questID).Scan(&courseID)
	return courseID, err
}

// TaskQuestID returns the quest id owning a task.
func (r *PictureQuestRepository) TaskQuestID(ctx context.Context, taskID int64) (int64, error) {
	var questID int64
	err := r.db.QueryRowContext(ctx, `SELECT quest_id FROM picture_quest_tasks WHERE id = ?`, taskID).Scan(&questID)
	return questID, err
}

// ListCourseLevels returns the districts (CEFR levels) of a course that have a conversation location.
func (r *PictureQuestRepository) ListCourseLevels(ctx context.Context, courseID int64) ([]CourseLevelOption, error) {
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
