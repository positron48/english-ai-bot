package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// LinglowWordItem is a vocabulary entry from the Linglow v2 schema.
type LinglowWordItem struct {
	LearningItemID int64      `json:"learning_item_id"`
	WordCardID     string     `json:"word_card_id"`
	Lemma          string     `json:"lemma"`
	DisplayWord    string     `json:"display_word"`
	Translation    string     `json:"translation"`
	CefrLevel      string     `json:"cefr_level"`
	State          string     `json:"state"`
	DueAt          *time.Time `json:"due_at"`
	Reps           int        `json:"reps"`
	AddedAt        *time.Time `json:"added_at"`
	LastReviewAt   *time.Time `json:"last_review_at"`
	// compat fields mirroring VocabWord for easier frontend reuse
	DueCount   int    `json:"due_count"`
	TotalCards int    `json:"total_cards"`
	TotalReps  int    `json:"total_reps"`
	MasteryLevel string `json:"mastery_level"`
}

// LinglowWordListResult is the paginated word list response.
type LinglowWordListResult struct {
	Course     CourseMapCourse     `json:"course"`
	UserCourse CourseMapUserCourse `json:"user_course"`
	Words      []LinglowWordItem   `json:"words"`
	Total      int                 `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	Generated  string              `json:"generated_at"`
}

// WordListOptions controls filtering/pagination for GetWordListForUser.
type WordListOptions struct {
	Search string // search term (case-insensitive prefix/contains on lemma or translation)
	Status string // srs state filter: new, learning, review, mastered
	Sort   string // word_asc (default), word_desc, added_at
	Limit  int
	Offset int
}

// GetWordListForUser returns vocabulary entries from Linglow v2 tables for words the user has studied.
func (r *CourseRepository) GetWordListForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, opts WordListOptions) (*LinglowWordListResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	courseCode, err := r.ResolveRequestedCourseCode(ctx, userID, defaultCourseCode, explicitCourseCode)
	if err != nil {
		return nil, err
	}
	userCourse, err := r.EnsureUserCourse(ctx, userID, courseCode)
	if err != nil {
		return nil, err
	}
	courseMap, err := r.GetCourseMap(ctx, courseCode, userID)
	if err != nil {
		return nil, err
	}

	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	if opts.Limit > 200 {
		opts.Limit = 200
	}

	whereExtra, whereArgs := buildWordListWhere(opts, userCourse.ID, courseMap.Course.ID)

	total, err := r.countWordList(ctx, whereExtra, whereArgs)
	if err != nil {
		return nil, fmt.Errorf("count word list: %w", err)
	}

	orderBy := wordListOrderBy(opts.Sort)
	items, err := r.queryWordList(ctx, userCourse.ID, courseMap.Course.ID, whereExtra, whereArgs, orderBy, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("query word list: %w", err)
	}

	return &LinglowWordListResult{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Words:      items,
		Total:      total,
		Limit:      opts.Limit,
		Offset:     opts.Offset,
		Generated:  time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func buildWordListWhere(opts WordListOptions, userCourseID, courseID int64) (string, []interface{}) {
	// Base args must match the 3 placeholders in wordListBaseSQL, in order:
	//   1) srs_items LEFT JOIN  -> si.user_course_id
	//   2) WHERE                -> li.course_id
	//   3) EXISTS subquery      -> ea_exists.user_course_id
	args := []interface{}{userCourseID, courseID, userCourseID}
	extra := ""

	if opts.Status != "" {
		state := normalizeWordSRSState(opts.Status)
		if state == "new" {
			extra += " AND si.id IS NULL"
		} else {
			extra += " AND si.state = ?"
			args = append(args, state)
		}
	}

	if s := strings.TrimSpace(opts.Search); s != "" {
		pattern := "%" + strings.ToLower(s) + "%"
		extra += " AND (LOWER(li.title) LIKE ? OR LOWER(wc.definition_ru) LIKE ? OR LOWER(wc.definition) LIKE ?)"
		args = append(args, pattern, pattern, pattern)
	}

	return extra, args
}

func normalizeWordSRSState(status string) string {
	switch strings.ToLower(status) {
	case "new":
		return "new"
	case "learning":
		return "learning"
	case "review":
		return "review"
	case "mastered":
		return "mastered"
	default:
		return ""
	}
}

func wordListOrderBy(sort string) string {
	switch sort {
	case "word_desc":
		return "li.title DESC NULLS LAST"
	case "added_at":
		return "added_at DESC NULLS LAST, li.title ASC"
	default:
		return "li.title ASC"
	}
}

const wordListBaseSQL = `
FROM learning_items li
JOIN word_cards wc ON wc.id = CAST(li.source_id AS BIGINT)
LEFT JOIN srs_items si ON si.learning_item_id = li.id AND si.user_course_id = ?
WHERE li.course_id = ?
  AND li.source_kind = 'word_card'
  AND li.status = 'published'
  AND (si.id IS NOT NULL OR EXISTS (
      SELECT 1 FROM exercise_attempts ea_exists
      WHERE ea_exists.learning_item_id = li.id AND ea_exists.user_course_id = ?
  ))`

func (r *CourseRepository) countWordList(ctx context.Context, whereExtra string, baseArgs []interface{}) (int, error) {
	q := "SELECT COUNT(*) " + wordListBaseSQL + whereExtra
	var count int
	if err := r.db.QueryRowContext(ctx, q, baseArgs...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CourseRepository) queryWordList(ctx context.Context, userCourseID, _ int64, whereExtra string, baseArgs []interface{}, orderBy string, limit, offset int) ([]LinglowWordItem, error) {
	q := `SELECT
		li.id,
		li.source_id,
		COALESCE(li.title, wc.word, '') AS lemma,
		COALESCE(NULLIF(wc.display_en, ''), wc.word, '') AS display_word,
		COALESCE(wc.definition_ru, wc.definition, '') AS translation,
		COALESCE(li.cefr_level, '') AS cefr_level,
		COALESCE(si.state, 'new') AS srs_state,
		si.due_at,
		COALESCE(si.reps, 0) AS reps,
		(SELECT MIN(ea1.created_at) FROM exercise_attempts ea1 WHERE ea1.learning_item_id = li.id AND ea1.user_course_id = ?) AS added_at,
		(SELECT MAX(ea2.created_at) FROM exercise_attempts ea2 WHERE ea2.learning_item_id = li.id AND ea2.user_course_id = ?) AS last_review_at
	` + wordListBaseSQL + whereExtra + `
	ORDER BY ` + orderBy + `
	LIMIT ? OFFSET ?`

	// prepend the two extra userCourseID args for added_at/last_review_at subqueries
	args := append([]interface{}{userCourseID, userCourseID}, baseArgs...)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var items []LinglowWordItem
	for rows.Next() {
		var it LinglowWordItem
		var dueAt, addedAt, lastReviewAt *time.Time
		if err := rows.Scan(
			&it.LearningItemID,
			&it.WordCardID,
			&it.Lemma,
			&it.DisplayWord,
			&it.Translation,
			&it.CefrLevel,
			&it.State,
			&dueAt,
			&it.Reps,
			&addedAt,
			&lastReviewAt,
		); err != nil {
			return nil, err
		}
		it.DueAt = dueAt
		it.AddedAt = addedAt
		it.LastReviewAt = lastReviewAt
		it.TotalCards = 1
		it.TotalReps = it.Reps
		it.MasteryLevel = wordSRSStateToMastery(it.State)
		if dueAt != nil && !dueAt.After(now) && it.State != "mastered" {
			it.DueCount = 1
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func wordSRSStateToMastery(state string) string {
	switch state {
	case "mastered":
		return "mastered"
	case "review":
		return "review"
	case "learning", "relearning":
		return "learning"
	default:
		return "new"
	}
}
