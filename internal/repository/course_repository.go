package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

// CourseRepository handles Linglow v2 course and user-course data.
type CourseRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

var ErrCourseNotFound = errors.New("course not found")

// UserCourseBackfillSummary describes one idempotent user_courses bootstrap run.
type UserCourseBackfillSummary struct {
	CourseCode   string
	CourseID     int64
	UsersScanned int64
	Existing     int64
	Created      int64
}

// ContentMappingSummary describes one idempotent legacy-content-to-Linglow mapping run.
type ContentMappingSummary struct {
	CourseCode     string
	CourseID       int64
	ModulesCreated int64
	ItemsCreated   int64
	ModulesTotal   int64
	ItemsTotal     int64
}

// CourseMap is the read model used by the course-aware API.
type CourseMap struct {
	Course     CourseMapCourse      `json:"course"`
	UserCourse *CourseMapUserCourse `json:"user_course,omitempty"`
	Districts  []CourseMapDistrict  `json:"districts"`
	Totals     CourseMapTotals      `json:"totals"`
}

type CourseMapCourse struct {
	ID             int64  `json:"id"`
	Code           string `json:"code"`
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	TargetLanguage string `json:"target_language"`
	NativeLanguage string `json:"native_language"`
	UILocale       string `json:"ui_locale"`
	Status         string `json:"status"`
	CityName       string `json:"city_name"`
}

type CourseMapUserCourse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type CourseMapDistrict struct {
	ID        int64               `json:"id"`
	Code      string              `json:"code"`
	LevelCode string              `json:"level_code"`
	Title     string              `json:"title"`
	Order     int                 `json:"order"`
	Status    string              `json:"status"`
	Locations []CourseMapLocation `json:"locations"`
}

type CourseMapLocation struct {
	ID           int64             `json:"id"`
	Code         string            `json:"code"`
	LocationType string            `json:"location_type"`
	Title        string            `json:"title"`
	Order        int               `json:"order"`
	Status       string            `json:"status"`
	Modules      []CourseMapModule `json:"modules"`
}

type CourseMapModule struct {
	ID         int64           `json:"id"`
	Code       string          `json:"code"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	SourceKind string          `json:"source_kind,omitempty"`
	SourceID   string          `json:"source_id,omitempty"`
	Order      int             `json:"order"`
	Status     string          `json:"status"`
	Items      []CourseMapItem `json:"items"`
}

type CourseMapItem struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	SourceKind string `json:"source_kind"`
	SourceID   string `json:"source_id"`
	Title      string `json:"title,omitempty"`
	CEFRLevel  string `json:"cefr_level,omitempty"`
	Status     string `json:"status"`
}

type CourseMapTotals struct {
	Districts int            `json:"districts"`
	Locations int            `json:"locations"`
	Modules   int            `json:"modules"`
	Items     int            `json:"items"`
	ByType    map[string]int `json:"by_type"`
}

type CourseSummary struct {
	ID             int64  `json:"id"`
	Code           string `json:"code"`
	Title          string `json:"title"`
	CityName       string `json:"city_name"`
	TargetLanguage string `json:"target_language"`
	NativeLanguage string `json:"native_language"`
	UILocale       string `json:"ui_locale"`
	Status         string `json:"status"`
	IsCurrent      bool   `json:"is_current"`
	UserCourseID   *int64 `json:"user_course_id,omitempty"`
	UserStatus     string `json:"user_status,omitempty"`
}

type CurrentCourse struct {
	Course     CourseSummary       `json:"course"`
	UserCourse CourseMapUserCourse `json:"user_course"`
}

type DailyRoute struct {
	Course     CourseMapCourse     `json:"course"`
	UserCourse CourseMapUserCourse `json:"user_course"`
	Summary    DailyRouteSummary   `json:"summary"`
	Review     []DailyRouteItem    `json:"review"`
	NewItems   []DailyRouteItem    `json:"new_items"`
	Generated  string              `json:"generated_at"`
}

type DailyRouteSummary struct {
	DueReviewCount int            `json:"due_review_count"`
	NewItemCount   int            `json:"new_item_count"`
	ByType         map[string]int `json:"by_type"`
	ReadSource     string         `json:"read_source"`
}

type DailyRouteItem struct {
	LearningItemID int64   `json:"learning_item_id"`
	SRSItemID      *int64  `json:"srs_item_id,omitempty"`
	Type           string  `json:"type"`
	SourceKind     string  `json:"source_kind"`
	SourceID       string  `json:"source_id"`
	Title          string  `json:"title,omitempty"`
	CEFRLevel      string  `json:"cefr_level,omitempty"`
	Mode           string  `json:"mode"`
	State          string  `json:"state,omitempty"`
	DueAt          *string `json:"due_at,omitempty"`
	DistrictCode   string  `json:"district_code,omitempty"`
	DistrictTitle  string  `json:"district_title,omitempty"`
	LocationCode   string  `json:"location_code,omitempty"`
	LocationType   string  `json:"location_type,omitempty"`
	LocationTitle  string  `json:"location_title,omitempty"`
	ModuleCode     string  `json:"module_code,omitempty"`
	ModuleTitle    string  `json:"module_title,omitempty"`
}

type ReviewQueue struct {
	Course     CourseMapCourse     `json:"course"`
	UserCourse CourseMapUserCourse `json:"user_course"`
	Summary    ReviewQueueSummary  `json:"summary"`
	Items      []DailyRouteItem    `json:"items"`
	Generated  string              `json:"generated_at"`
}

type ReviewQueueSummary struct {
	DueCount        int            `json:"due_count"`
	LearningCount   int            `json:"learning_count"`
	ReviewCount     int            `json:"review_count"`
	RelearningCount int            `json:"relearning_count"`
	UpcomingCount   int            `json:"upcoming_count"`
	ByType          map[string]int `json:"by_type"`
	ReadSource      string         `json:"read_source"`
}

type CourseProgress struct {
	Course     CourseMapCourse       `json:"course"`
	UserCourse CourseMapUserCourse   `json:"user_course"`
	Summary    CourseProgressSummary `json:"summary"`
	ByType     []CourseProgressType  `json:"by_type"`
	Generated  string                `json:"generated_at"`
}

type CourseProgressSummary struct {
	TotalItems      int     `json:"total_items"`
	AttemptedItems  int     `json:"attempted_items"`
	MasteredItems   int     `json:"mastered_items"`
	DueReviewCount  int     `json:"due_review_count"`
	AttemptCount    int     `json:"attempt_count"`
	CorrectCount    int     `json:"correct_count"`
	ProgressPercent float64 `json:"progress_percent"`
	AccuracyPercent float64 `json:"accuracy_percent"`
}

type CourseProgressType struct {
	Type            string  `json:"type"`
	TotalItems      int     `json:"total_items"`
	AttemptedItems  int     `json:"attempted_items"`
	MasteredItems   int     `json:"mastered_items"`
	ProgressPercent float64 `json:"progress_percent"`
}

type ExerciseAttemptInput struct {
	UserID          int64
	DefaultCourse   string
	ExplicitCourse  string
	LearningItemID  int64
	SRSItemID       int64
	Mode            string
	ClientAttemptID string
	IsCorrect       *bool
	Score           *int
	Quality         *int
	PromptJSON      string
	AnswerJSON      string
	ResultJSON      string
	AnsweredAt      time.Time
	UpdateSRS       bool
}

type ExerciseAttemptResult struct {
	ID              int64               `json:"id"`
	UserCourseID    int64               `json:"user_course_id"`
	LearningItemID  *int64              `json:"learning_item_id,omitempty"`
	SRSItemID       *int64              `json:"srs_item_id,omitempty"`
	ClientAttemptID string              `json:"client_attempt_id,omitempty"`
	Duplicate       bool                `json:"duplicate"`
	EventID         int64               `json:"event_id,omitempty"`
	SRSUpdated      bool                `json:"srs_updated"`
	Course          CourseMapCourse     `json:"course"`
	UserCourse      CourseMapUserCourse `json:"user_course"`
}

type SRSShadowReport struct {
	Course      CourseMapCourse            `json:"course"`
	UserCourse  CourseMapUserCourse        `json:"user_course"`
	Due         SRSShadowDueReport         `json:"due"`
	ReviewQueue SRSShadowReviewQueueReport `json:"review_queue"`
	Mastery     SRSShadowMasteryReport     `json:"mastery"`
	Generated   string                     `json:"generated_at"`
}

type SRSShadowDueReport struct {
	LegacyDueCount   int `json:"legacy_due_count"`
	LinglowDueCount  int `json:"linglow_due_count"`
	OverlapCount     int `json:"overlap_count"`
	LegacyOnlyCount  int `json:"legacy_only_count"`
	LinglowOnlyCount int `json:"linglow_only_count"`
}

type SRSShadowMasteryReport struct {
	ComparedCount     int     `json:"compared_count"`
	AverageLegacy     float64 `json:"average_legacy"`
	AverageLinglow    float64 `json:"average_linglow"`
	AverageDifference float64 `json:"average_difference"`
	MaxDifference     float64 `json:"max_difference"`
}

type SRSShadowReviewQueueReport struct {
	LegacyDueCount        int            `json:"legacy_due_count"`
	CanonicalDueCount     int            `json:"canonical_due_count"`
	OverlapCount          int            `json:"overlap_count"`
	LegacyOnlyCount       int            `json:"legacy_only_count"`
	CanonicalOnlyCount    int            `json:"canonical_only_count"`
	ReadyForCanonicalRead bool           `json:"ready_for_canonical_read"`
	ByType                map[string]int `json:"by_type"`
}

func NewCourseRepository(db *sql.DB, logger *zap.Logger) *CourseRepository {
	return &CourseRepository{db: db, logger: logger}
}

// CourseCodeForLearning returns the target canonical course code for a language pair.
func CourseCodeForLearning(lc config.LearningConfig) string {
	target := strings.TrimSpace(strings.ToLower(lc.TargetLang))
	native := strings.TrimSpace(strings.ToLower(lc.NativeLang))
	if target == "" || native == "" {
		return ""
	}
	return target + "_" + native
}

// BackfillUserCoursesForLearning idempotently attaches all existing users to the current course.
func (r *CourseRepository) BackfillUserCoursesForLearning(ctx context.Context, lc config.LearningConfig) (*UserCourseBackfillSummary, error) {
	return r.BackfillUserCourses(ctx, CourseCodeForLearning(lc))
}

// BackfillUserCourses idempotently attaches all existing users to courseCode.
func (r *CourseRepository) BackfillUserCourses(ctx context.Context, courseCode string) (*UserCourseBackfillSummary, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}

	var courseID int64
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}

	var usersScanned int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&usersScanned); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var existing int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_courses WHERE course_id = ?`, courseID).Scan(&existing); err != nil {
		return nil, fmt.Errorf("count existing user courses: %w", err)
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_courses (user_id, course_id, status, started_at, created_at, updated_at)
		SELECT u.id, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM users u
		WHERE NOT EXISTS (
			SELECT 1 FROM user_courses uc
			WHERE uc.user_id = u.id AND uc.course_id = ?
		)
	`, courseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("insert missing user courses: %w", err)
	}

	created, _ := result.RowsAffected()
	return &UserCourseBackfillSummary{
		CourseCode:   courseCode,
		CourseID:     courseID,
		UsersScanned: usersScanned,
		Existing:     existing,
		Created:      created,
	}, nil
}

func (r *CourseRepository) ListCoursesForUser(ctx context.Context, userID int64, defaultCourseCode string) ([]CourseSummary, error) {
	currentCode, err := r.ResolveCurrentCourseCode(ctx, userID, defaultCourseCode)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.city_name, c.target_lang, c.teaching_locale, c.ui_locale, c.status,
			uc.id, uc.status
		FROM courses c
		LEFT JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?
		WHERE c.status = 'active'
		ORDER BY c.code
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()
	var out []CourseSummary
	for rows.Next() {
		var c CourseSummary
		var userCourseID sql.NullInt64
		var userStatus sql.NullString
		if err := rows.Scan(&c.ID, &c.Code, &c.Title, &c.CityName, &c.TargetLanguage, &c.NativeLanguage, &c.UILocale, &c.Status, &userCourseID, &userStatus); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		if userCourseID.Valid {
			id := userCourseID.Int64
			c.UserCourseID = &id
		}
		if userStatus.Valid {
			c.UserStatus = userStatus.String
		}
		c.IsCurrent = c.Code == currentCode
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CourseRepository) GetCurrentCourse(ctx context.Context, userID int64, defaultCourseCode string) (*CurrentCourse, error) {
	courseCode, err := r.ResolveCurrentCourseCode(ctx, userID, defaultCourseCode)
	if err != nil {
		return nil, err
	}
	if _, err := r.SelectCurrentCourse(ctx, userID, courseCode); err != nil {
		return nil, err
	}
	var current CurrentCourse
	var userCourseID int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.code, c.title, c.city_name, c.target_lang, c.teaching_locale, c.ui_locale, c.status,
			uc.id, uc.status
		FROM courses c
		JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?
		WHERE c.code = ?
	`, userID, courseCode).Scan(
		&current.Course.ID,
		&current.Course.Code,
		&current.Course.Title,
		&current.Course.CityName,
		&current.Course.TargetLanguage,
		&current.Course.NativeLanguage,
		&current.Course.UILocale,
		&current.Course.Status,
		&userCourseID,
		&current.UserCourse.Status,
	); err != nil {
		return nil, fmt.Errorf("get current course: %w", err)
	}
	current.Course.IsCurrent = true
	current.Course.UserCourseID = &userCourseID
	current.Course.UserStatus = current.UserCourse.Status
	current.UserCourse.ID = userCourseID
	return &current, nil
}

func (r *CourseRepository) SelectCurrentCourse(ctx context.Context, userID int64, courseCode string) (*CurrentCourse, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin select current course: %w", err)
	}
	defer tx.Rollback()
	var course CourseSummary
	if err := tx.QueryRowContext(ctx, `
		SELECT id, code, title, city_name, target_lang, teaching_locale, ui_locale, status
		FROM courses
		WHERE code = ? AND status = 'active'
	`, courseCode).Scan(&course.ID, &course.Code, &course.Title, &course.CityName, &course.TargetLanguage, &course.NativeLanguage, &course.UILocale, &course.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("course %q is not active or does not exist", courseCode)
		}
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}
	var userCourseID int64
	var userCourseStatus string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO user_courses (user_id, course_id, status, started_at, created_at, updated_at)
		VALUES (?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, course_id) DO UPDATE SET
			status = CASE WHEN user_courses.status = 'archived' THEN 'active' ELSE user_courses.status END,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, status
	`, userID, course.ID).Scan(&userCourseID, &userCourseStatus); err != nil {
		return nil, fmt.Errorf("upsert user course: %w", err)
	}
	if err := updateCurrentCourseSetting(ctx, tx, userID, course.Code); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit select current course: %w", err)
	}
	course.IsCurrent = true
	course.UserCourseID = &userCourseID
	course.UserStatus = userCourseStatus
	return &CurrentCourse{Course: course, UserCourse: CourseMapUserCourse{ID: userCourseID, Status: course.UserStatus}}, nil
}

func (r *CourseRepository) EnsureUserCourse(ctx context.Context, userID int64, courseCode string) (*CourseMapUserCourse, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	var userCourse CourseMapUserCourse
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO user_courses (user_id, course_id, status, started_at, created_at, updated_at)
		SELECT ?, c.id, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM courses c
		WHERE c.code = ? AND c.status = 'active'
		ON CONFLICT (user_id, course_id) DO UPDATE SET
			status = CASE WHEN user_courses.status = 'archived' THEN 'active' ELSE user_courses.status END,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, status
	`, userID, courseCode).Scan(&userCourse.ID, &userCourse.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", ErrCourseNotFound, courseCode)
		}
		return nil, fmt.Errorf("ensure user course %q: %w", courseCode, err)
	}
	return &userCourse, nil
}

func (r *CourseRepository) ResolveRequestedCourseCode(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string) (string, error) {
	explicitCourseCode = strings.TrimSpace(strings.ToLower(explicitCourseCode))
	if explicitCourseCode != "" {
		if !r.courseIsActive(ctx, explicitCourseCode) {
			return "", fmt.Errorf("%w: %s", ErrCourseNotFound, explicitCourseCode)
		}
		return explicitCourseCode, nil
	}
	return r.ResolveCurrentCourseCode(ctx, userID, defaultCourseCode)
}

func (r *CourseRepository) ResolveCurrentCourseCode(ctx context.Context, userID int64, defaultCourseCode string) (string, error) {
	defaultCourseCode = strings.TrimSpace(strings.ToLower(defaultCourseCode))
	var settingsRaw sql.NullString
	if err := r.db.QueryRowContext(ctx, `SELECT settings_json FROM users WHERE id = ?`, userID).Scan(&settingsRaw); err != nil {
		return "", fmt.Errorf("get user settings: %w", err)
	}
	if settingsRaw.Valid && strings.TrimSpace(settingsRaw.String) != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsRaw.String), &settings); err == nil {
			if value, ok := settings["current_course_code"].(string); ok {
				value = strings.TrimSpace(strings.ToLower(value))
				if value != "" && r.courseIsActive(ctx, value) {
					return value, nil
				}
			}
		}
	}
	if defaultCourseCode != "" && r.courseIsActive(ctx, defaultCourseCode) {
		return defaultCourseCode, nil
	}
	var code string
	if err := r.db.QueryRowContext(ctx, `SELECT code FROM courses WHERE status = 'active' ORDER BY code LIMIT 1`).Scan(&code); err != nil {
		return "", fmt.Errorf("resolve fallback course: %w", err)
	}
	return code, nil
}

func (r *CourseRepository) courseIsActive(ctx context.Context, courseCode string) bool {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM courses WHERE code = ? AND status = 'active'`, courseCode).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func updateCurrentCourseSetting(ctx context.Context, tx *sql.Tx, userID int64, courseCode string) error {
	var raw sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT settings_json FROM users WHERE id = ?`, userID).Scan(&raw); err != nil {
		return fmt.Errorf("get user settings for update: %w", err)
	}
	settings := map[string]interface{}{}
	if raw.Valid && strings.TrimSpace(raw.String) != "" {
		_ = json.Unmarshal([]byte(raw.String), &settings)
	}
	settings["current_course_code"] = courseCode
	encoded, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("encode course setting: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET settings_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, string(encoded), userID); err != nil {
		return fmt.Errorf("update course setting: %w", err)
	}
	return nil
}

// GetCourseMapForLearning returns the Linglow v2 course map for the configured language pair.
func (r *CourseRepository) GetCourseMapForLearning(ctx context.Context, lc config.LearningConfig, userID int64) (*CourseMap, error) {
	return r.GetCourseMap(ctx, CourseCodeForLearning(lc), userID)
}

// GetCourseMapForUser resolves the requested course and returns the Linglow v2 course map.
// An explicit course code is scoped to this read and does not update the user's current course preference.
func (r *CourseRepository) GetCourseMapForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string) (*CourseMap, error) {
	courseCode, err := r.ResolveRequestedCourseCode(ctx, userID, defaultCourseCode, explicitCourseCode)
	if err != nil {
		return nil, err
	}
	if userID > 0 {
		if _, err := r.EnsureUserCourse(ctx, userID, courseCode); err != nil {
			return nil, err
		}
	}
	return r.GetCourseMap(ctx, courseCode, userID)
}

func (r *CourseRepository) GetDailyRouteForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, limit int) (*DailyRoute, error) {
	return r.GetDailyRouteForUserWithSRSRead(ctx, userID, defaultCourseCode, explicitCourseCode, limit, true)
}

func (r *CourseRepository) GetDailyRouteForUserWithSRSRead(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, limit int, canonicalSRSRead bool) (*DailyRoute, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 50 {
		limit = 50
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

	route := &DailyRoute{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Summary:    DailyRouteSummary{ByType: map[string]int{}},
		Generated:  time.Now().UTC().Format(time.RFC3339),
	}
	if courseMap.UserCourse != nil {
		route.UserCourse = *courseMap.UserCourse
	}

	summary, err := r.getDailyRouteSummary(ctx, userCourse.ID, userID, courseMap.Course.Code, canonicalSRSRead)
	if err != nil {
		return nil, err
	}
	route.Summary = summary
	route.Review, err = r.listDailyRouteReviewItems(ctx, userCourse.ID, userID, courseMap.Course.Code, limit, canonicalSRSRead)
	if err != nil {
		return nil, err
	}
	route.NewItems, err = r.listDailyRouteNewItems(ctx, userCourse.ID, courseMap.Course.ID, limit)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (r *CourseRepository) GetReviewQueueForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, limit int) (*ReviewQueue, error) {
	return r.GetReviewQueueForUserWithSRSRead(ctx, userID, defaultCourseCode, explicitCourseCode, limit, true)
}

func (r *CourseRepository) GetReviewQueueForUserWithSRSRead(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, limit int, canonicalSRSRead bool) (*ReviewQueue, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
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
	queue := &ReviewQueue{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Summary:    ReviewQueueSummary{ByType: map[string]int{}},
		Generated:  time.Now().UTC().Format(time.RFC3339),
	}
	if courseMap.UserCourse != nil {
		queue.UserCourse = *courseMap.UserCourse
	}
	queue.Summary, err = r.getReviewQueueSummary(ctx, userCourse.ID, userID, courseMap.Course.Code, canonicalSRSRead)
	if err != nil {
		return nil, err
	}
	queue.Items, err = r.listDailyRouteReviewItems(ctx, userCourse.ID, userID, courseMap.Course.Code, limit, canonicalSRSRead)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func (r *CourseRepository) GetProgressForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string) (*CourseProgress, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
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
	progress := &CourseProgress{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Generated:  time.Now().UTC().Format(time.RFC3339),
	}
	if courseMap.UserCourse != nil {
		progress.UserCourse = *courseMap.UserCourse
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT li.id) AS total_items,
			COUNT(DISTINCT ea.learning_item_id) AS attempted_items,
			COUNT(DISTINCT CASE WHEN si.state = 'mastered' THEN si.learning_item_id END) AS mastered_items,
			COUNT(DISTINCT CASE WHEN si.state IN ('learning', 'review', 'relearning') AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP) THEN si.id END) AS due_review_count,
			COUNT(ea.id) AS attempt_count,
			COUNT(CASE WHEN ea.is_correct THEN 1 END) AS correct_count
		FROM learning_items li
		LEFT JOIN exercise_attempts ea ON ea.learning_item_id = li.id AND ea.user_course_id = ?
		LEFT JOIN srs_items si ON si.learning_item_id = li.id AND si.user_course_id = ?
		WHERE li.course_id = ? AND li.status = 'published'
	`, userCourse.ID, userCourse.ID, courseMap.Course.ID).Scan(
		&progress.Summary.TotalItems,
		&progress.Summary.AttemptedItems,
		&progress.Summary.MasteredItems,
		&progress.Summary.DueReviewCount,
		&progress.Summary.AttemptCount,
		&progress.Summary.CorrectCount,
	); err != nil {
		return nil, fmt.Errorf("get course progress summary: %w", err)
	}
	progress.Summary.ProgressPercent = percent(progress.Summary.AttemptedItems+progress.Summary.MasteredItems, progress.Summary.TotalItems)
	progress.Summary.AccuracyPercent = percent(progress.Summary.CorrectCount, progress.Summary.AttemptCount)

	progress.ByType, err = r.listProgressByType(ctx, userCourse.ID, courseMap.Course.ID)
	if err != nil {
		return nil, err
	}
	return progress, nil
}

func (r *CourseRepository) GetSRSShadowReportForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string) (*SRSShadowReport, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
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
	report := &SRSShadowReport{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Generated:  time.Now().UTC().Format(time.RFC3339),
	}
	if courseMap.UserCourse != nil {
		report.UserCourse = *courseMap.UserCourse
	}
	report.Due, err = r.getSRSShadowDueReport(ctx, userID, userCourse.ID, courseMap.Course.ID)
	if err != nil {
		return nil, err
	}
	report.ReviewQueue, err = r.getSRSShadowReviewQueueReport(ctx, userID, userCourse.ID, courseMap.Course.Code)
	if err != nil {
		return nil, err
	}
	report.Mastery, err = r.getSRSShadowMasteryReport(ctx, userID, userCourse.ID, courseMap.Course.ID)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (r *CourseRepository) getSRSShadowReviewQueueReport(ctx context.Context, userID, userCourseID int64, courseCode string) (SRSShadowReviewQueueReport, error) {
	report := SRSShadowReviewQueueReport{ByType: map[string]int{}}
	if err := r.db.QueryRowContext(ctx, `
		WITH course_scope AS (
			SELECT id, target_lang AS target_language
			FROM courses
			WHERE code = ?
		), legacy_due AS (
			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'word_card'
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
				AND li.status = 'published'
			WHERE uc.user_id = ?
				AND (uc.next_due_at IS NULL OR uc.next_due_at <= CURRENT_TIMESTAMP)
				AND NOT EXISTS (
					SELECT 1 FROM user_word_knowledge uwk
					WHERE uwk.user_id = uc.user_id AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
				)

			UNION

			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM grammar_theory_memory gtm
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'grammar_theory_block'
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
				AND li.status = 'published'
			WHERE gtm.user_id = ?
				AND lower(gtm.language) = lower((SELECT target_language FROM course_scope))
				AND lower(gtm.course_id) = lower((SELECT target_language FROM course_scope))
				AND gtm.next_review_at <= CURRENT_TIMESTAMP
		), canonical_due AS (
			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM srs_items si
			JOIN learning_items li ON li.id = si.learning_item_id
			WHERE si.user_course_id = ?
				AND li.status = 'published'
				AND si.state IN ('learning', 'review', 'relearning')
				AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		)
		SELECT
			(SELECT COUNT(*) FROM legacy_due),
			(SELECT COUNT(*) FROM canonical_due),
			(SELECT COUNT(*) FROM legacy_due l JOIN canonical_due c USING (learning_item_id)),
			(SELECT COUNT(*) FROM legacy_due l WHERE NOT EXISTS (SELECT 1 FROM canonical_due c WHERE c.learning_item_id = l.learning_item_id)),
			(SELECT COUNT(*) FROM canonical_due c WHERE NOT EXISTS (SELECT 1 FROM legacy_due l WHERE l.learning_item_id = c.learning_item_id))
	`, courseCode, userID, userID, userCourseID).Scan(
		&report.LegacyDueCount,
		&report.CanonicalDueCount,
		&report.OverlapCount,
		&report.LegacyOnlyCount,
		&report.CanonicalOnlyCount,
	); err != nil {
		return report, fmt.Errorf("get srs shadow review queue report: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
		WITH course_scope AS (
			SELECT id, target_lang AS target_language
			FROM courses
			WHERE code = ?
		), legacy_due AS (
			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'word_card'
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
				AND li.status = 'published'
			WHERE uc.user_id = ?
				AND (uc.next_due_at IS NULL OR uc.next_due_at <= CURRENT_TIMESTAMP)
				AND NOT EXISTS (
					SELECT 1 FROM user_word_knowledge uwk
					WHERE uwk.user_id = uc.user_id AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
				)

			UNION

			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM grammar_theory_memory gtm
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'grammar_theory_block'
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
				AND li.status = 'published'
			WHERE gtm.user_id = ?
				AND lower(gtm.language) = lower((SELECT target_language FROM course_scope))
				AND lower(gtm.course_id) = lower((SELECT target_language FROM course_scope))
				AND gtm.next_review_at <= CURRENT_TIMESTAMP
		), canonical_due AS (
			SELECT DISTINCT li.id AS learning_item_id, li.item_type
			FROM srs_items si
			JOIN learning_items li ON li.id = si.learning_item_id
			WHERE si.user_course_id = ?
				AND li.status = 'published'
				AND si.state IN ('learning', 'review', 'relearning')
				AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		)
		SELECT COALESCE(l.item_type, c.item_type) AS item_type, COUNT(*)
		FROM legacy_due l
		FULL OUTER JOIN canonical_due c USING (learning_item_id)
		GROUP BY COALESCE(l.item_type, c.item_type)
	`, courseCode, userID, userID, userCourseID)
	if err != nil {
		return report, fmt.Errorf("get srs shadow review queue types: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var itemType string
		var count int
		if err := rows.Scan(&itemType, &count); err != nil {
			return report, fmt.Errorf("scan srs shadow review queue type: %w", err)
		}
		report.ByType[itemType] = count
	}
	if err := rows.Err(); err != nil {
		return report, fmt.Errorf("iterate srs shadow review queue types: %w", err)
	}
	report.ReadyForCanonicalRead = report.LegacyOnlyCount == 0 && report.CanonicalOnlyCount == 0
	return report, nil
}

func (r *CourseRepository) getSRSShadowDueReport(ctx context.Context, userID, userCourseID, courseID int64) (SRSShadowDueReport, error) {
	var report SRSShadowDueReport
	if err := r.db.QueryRowContext(ctx, `
		WITH legacy_due AS (
			SELECT DISTINCT tc.word_card_id
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			WHERE uc.user_id = ?
				AND (uc.next_due_at IS NULL OR uc.next_due_at <= CURRENT_TIMESTAMP)
				AND NOT EXISTS (
					SELECT 1 FROM user_word_knowledge uwk
					WHERE uwk.user_id = uc.user_id AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
				)
		),
		linglow_due AS (
			SELECT DISTINCT li.source_id::BIGINT AS word_card_id
			FROM srs_items si
			JOIN learning_items li ON li.id = si.learning_item_id
			WHERE si.user_course_id = ?
				AND li.course_id = ?
				AND li.source_kind = 'word_card'
				AND li.source_id ~ '^[0-9]+$'
				AND si.state IN ('learning', 'review', 'relearning')
				AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		)
		SELECT
			(SELECT COUNT(*) FROM legacy_due),
			(SELECT COUNT(*) FROM linglow_due),
			(SELECT COUNT(*) FROM legacy_due l JOIN linglow_due n USING (word_card_id)),
			(SELECT COUNT(*) FROM legacy_due l WHERE NOT EXISTS (SELECT 1 FROM linglow_due n WHERE n.word_card_id = l.word_card_id)),
			(SELECT COUNT(*) FROM linglow_due n WHERE NOT EXISTS (SELECT 1 FROM legacy_due l WHERE l.word_card_id = n.word_card_id))
	`, userID, userCourseID, courseID).Scan(
		&report.LegacyDueCount,
		&report.LinglowDueCount,
		&report.OverlapCount,
		&report.LegacyOnlyCount,
		&report.LinglowOnlyCount,
	); err != nil {
		return report, fmt.Errorf("get srs shadow due report: %w", err)
	}
	return report, nil
}

func (r *CourseRepository) getSRSShadowMasteryReport(ctx context.Context, userID, userCourseID, courseID int64) (SRSShadowMasteryReport, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(uwm.mastering_score, 0)::DOUBLE PRECISION AS legacy_score,
			CASE
				WHEN si.stats_json->>'mastery_score' IS NOT NULL THEN COALESCE((si.stats_json->>'mastery_score')::DOUBLE PRECISION, 0)
				WHEN si.state = 'mastered' THEN 100
				WHEN si.state = 'review' THEN LEAST(100, 50 + (si.reps * 10))
				WHEN si.state IN ('learning', 'relearning') THEN 25
				ELSE 0
			END AS linglow_score
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		JOIN user_courses uc ON uc.id = si.user_course_id
		LEFT JOIN user_word_mastering uwm
			ON uwm.user_id = uc.user_id
			AND li.source_kind = 'word_card'
			AND li.source_id ~ '^[0-9]+$'
			AND uwm.word_card_id = li.source_id::BIGINT
		WHERE si.user_course_id = ?
			AND uc.user_id = ?
			AND li.course_id = ?
			AND li.source_kind = 'word_card'
			AND li.source_id ~ '^[0-9]+$'
	`, userCourseID, userID, courseID)
	if err != nil {
		return SRSShadowMasteryReport{}, fmt.Errorf("get srs shadow mastery rows: %w", err)
	}
	defer rows.Close()
	var report SRSShadowMasteryReport
	var legacyTotal, linglowTotal, diffTotal float64
	for rows.Next() {
		var legacy, linglow float64
		if err := rows.Scan(&legacy, &linglow); err != nil {
			return report, fmt.Errorf("scan srs shadow mastery row: %w", err)
		}
		diff := absFloat(legacy - linglow)
		report.ComparedCount++
		legacyTotal += legacy
		linglowTotal += linglow
		diffTotal += diff
		if diff > report.MaxDifference {
			report.MaxDifference = diff
		}
	}
	if err := rows.Err(); err != nil {
		return report, fmt.Errorf("iterate srs shadow mastery rows: %w", err)
	}
	if report.ComparedCount > 0 {
		report.AverageLegacy = legacyTotal / float64(report.ComparedCount)
		report.AverageLinglow = linglowTotal / float64(report.ComparedCount)
		report.AverageDifference = diffTotal / float64(report.ComparedCount)
	}
	return report, nil
}

func (r *CourseRepository) listProgressByType(ctx context.Context, userCourseID, courseID int64) ([]CourseProgressType, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			li.item_type,
			COUNT(DISTINCT li.id) AS total_items,
			COUNT(DISTINCT ea.learning_item_id) AS attempted_items,
			COUNT(DISTINCT CASE WHEN si.state = 'mastered' THEN si.learning_item_id END) AS mastered_items
		FROM learning_items li
		LEFT JOIN exercise_attempts ea ON ea.learning_item_id = li.id AND ea.user_course_id = ?
		LEFT JOIN srs_items si ON si.learning_item_id = li.id AND si.user_course_id = ?
		WHERE li.course_id = ? AND li.status = 'published'
		GROUP BY li.item_type
		ORDER BY li.item_type
	`, userCourseID, userCourseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("list progress by type: %w", err)
	}
	defer rows.Close()
	out := []CourseProgressType{}
	for rows.Next() {
		var row CourseProgressType
		if err := rows.Scan(&row.Type, &row.TotalItems, &row.AttemptedItems, &row.MasteredItems); err != nil {
			return nil, fmt.Errorf("scan progress type: %w", err)
		}
		row.ProgressPercent = percent(row.AttemptedItems+row.MasteredItems, row.TotalItems)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate progress types: %w", err)
	}
	return out, nil
}

func (r *CourseRepository) RecordExerciseAttempt(ctx context.Context, input ExerciseAttemptInput) (*ExerciseAttemptResult, error) {
	if input.UserID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	input.Mode = strings.TrimSpace(input.Mode)
	if input.Mode == "" {
		return nil, fmt.Errorf("mode is required")
	}
	courseCode, err := r.ResolveRequestedCourseCode(ctx, input.UserID, input.DefaultCourse, input.ExplicitCourse)
	if err != nil {
		return nil, err
	}
	userCourse, err := r.EnsureUserCourse(ctx, input.UserID, courseCode)
	if err != nil {
		return nil, err
	}
	courseMap, err := r.GetCourseMap(ctx, courseCode, input.UserID)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin exercise attempt: %w", err)
	}
	defer tx.Rollback()

	var learningItemID interface{}
	var learningItemIDPtr *int64
	if input.LearningItemID > 0 {
		var exists bool
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) > 0
			FROM learning_items
			WHERE id = ? AND course_id = ? AND status = 'published'
		`, input.LearningItemID, courseMap.Course.ID).Scan(&exists); err != nil {
			return nil, fmt.Errorf("check learning item: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("learning item %d does not belong to course %s", input.LearningItemID, courseMap.Course.Code)
		}
		learningItemID = input.LearningItemID
		id := input.LearningItemID
		learningItemIDPtr = &id
	}
	var srsItemID interface{}
	var srsItemIDPtr *int64
	if input.SRSItemID > 0 {
		var linkedLearningItemID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT learning_item_id
			FROM srs_items
			WHERE id = ? AND user_course_id = ?
		`, input.SRSItemID, userCourse.ID).Scan(&linkedLearningItemID); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("srs item %d does not belong to user course", input.SRSItemID)
			}
			return nil, fmt.Errorf("check srs item: %w", err)
		}
		if input.LearningItemID > 0 && linkedLearningItemID != input.LearningItemID {
			return nil, fmt.Errorf("srs item %d does not match learning item %d", input.SRSItemID, input.LearningItemID)
		}
		if input.LearningItemID == 0 {
			learningItemID = linkedLearningItemID
			id := linkedLearningItemID
			learningItemIDPtr = &id
		}
		srsItemID = input.SRSItemID
		id := input.SRSItemID
		srsItemIDPtr = &id
	}
	clientAttemptID := strings.TrimSpace(input.ClientAttemptID)
	if clientAttemptID != "" {
		existing, err := r.getExistingExerciseAttempt(ctx, tx, userCourse.ID, clientAttemptID, courseMap.Course, *userCourse)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, tx.Commit()
		}
	}
	answeredAt := input.AnsweredAt
	if answeredAt.IsZero() {
		answeredAt = time.Now()
	}
	promptJSON := normalizeJSONObject(input.PromptJSON)
	answerJSON := normalizeJSONObject(input.AnswerJSON)
	resultJSON := normalizeJSONObject(input.ResultJSON)
	var clientAttemptValue interface{}
	if clientAttemptID != "" {
		clientAttemptValue = clientAttemptID
	}
	var exerciseID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, srs_item_id, mode, client_attempt_id,
			started_at, answered_at, is_correct, score, quality,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CAST(? AS jsonb), CAST(? AS jsonb), CAST(? AS jsonb), 'linglow_api', ?)
		RETURNING id
	`, userCourse.ID, learningItemID, srsItemID, input.Mode, clientAttemptValue, answeredAt, answeredAt, input.IsCorrect, input.Score, input.Quality, promptJSON, answerJSON, resultJSON, clientAttemptID).Scan(&exerciseID); err != nil {
		return nil, fmt.Errorf("insert exercise attempt: %w", err)
	}
	srsUpdated := false
	if input.UpdateSRS && learningItemIDPtr != nil {
		updatedID, err := r.upsertSRSForExerciseAttempt(ctx, tx, userCourse.ID, *learningItemIDPtr, input.IsCorrect, input.Quality, answeredAt)
		if err != nil {
			return nil, err
		}
		if updatedID > 0 {
			srsUpdated = true
			if srsItemIDPtr == nil {
				srsItemIDPtr = &updatedID
				if _, err := tx.ExecContext(ctx, `UPDATE exercise_attempts SET srs_item_id = ? WHERE id = ?`, updatedID, exerciseID); err != nil {
					return nil, fmt.Errorf("link exercise attempt to srs item: %w", err)
				}
			}
		}
	}
	eventJSON, _ := json.Marshal(map[string]interface{}{
		"mode":              input.Mode,
		"client_attempt_id": clientAttemptID,
		"is_correct":        input.IsCorrect,
		"score":             input.Score,
		"quality":           input.Quality,
		"srs_updated":       srsUpdated,
	})
	var eventID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id,
			event_type, event_time, mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'exercise_attempt_submitted', ?, ?, 'linglow_api', ?, CAST(? AS jsonb))
		RETURNING id
	`, userCourse.ID, learningItemID, exerciseID, answeredAt, input.Mode, clientAttemptID, string(eventJSON)).Scan(&eventID); err != nil {
		return nil, fmt.Errorf("insert learning event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit exercise attempt: %w", err)
	}
	return &ExerciseAttemptResult{
		ID:              exerciseID,
		UserCourseID:    userCourse.ID,
		LearningItemID:  learningItemIDPtr,
		SRSItemID:       srsItemIDPtr,
		ClientAttemptID: clientAttemptID,
		EventID:         eventID,
		SRSUpdated:      srsUpdated,
		Course:          courseMap.Course,
		UserCourse:      *userCourse,
	}, nil
}

func (r *CourseRepository) getExistingExerciseAttempt(ctx context.Context, tx *sql.Tx, userCourseID int64, clientAttemptID string, course CourseMapCourse, userCourse CourseMapUserCourse) (*ExerciseAttemptResult, error) {
	var result ExerciseAttemptResult
	var learningItemID sql.NullInt64
	var srsItemID sql.NullInt64
	err := tx.QueryRowContext(ctx, `
		SELECT id, learning_item_id, srs_item_id
		FROM exercise_attempts
		WHERE user_course_id = ? AND client_attempt_id = ?
	`, userCourseID, clientAttemptID).Scan(&result.ID, &learningItemID, &srsItemID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("check existing exercise attempt: %w", err)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	result.UserCourseID = userCourseID
	result.ClientAttemptID = clientAttemptID
	result.Duplicate = true
	result.Course = course
	result.UserCourse = userCourse
	if learningItemID.Valid {
		id := learningItemID.Int64
		result.LearningItemID = &id
	}
	if srsItemID.Valid {
		id := srsItemID.Int64
		result.SRSItemID = &id
	}
	return &result, nil
}

type linglowSRSState struct {
	ID           int64
	State        string
	EF           float64
	Reps         int
	IntervalDays int
	LearningStep int
	LapseCount   int
	Stats        map[string]interface{}
}

func (r *CourseRepository) upsertSRSForExerciseAttempt(ctx context.Context, tx *sql.Tx, userCourseID, learningItemID int64, isCorrect *bool, quality *int, reviewedAt time.Time) (int64, error) {
	if isCorrect == nil {
		return 0, nil
	}
	current, err := r.getOrCreateSRSState(ctx, tx, userCourseID, learningItemID)
	if err != nil {
		return 0, err
	}
	next := applyLinglowSRS(current, *isCorrect, quality, reviewedAt)
	statsJSON, _ := json.Marshal(next.Stats)
	if _, err := tx.ExecContext(ctx, `
		UPDATE srs_items
		SET state = ?, stability = ?, difficulty = ?, due_at = ?, last_review_at = ?,
			reps = ?, lapse_count = ?, stats_json = CAST(? AS jsonb), updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, next.State, float64(next.IntervalDays), next.EF, next.Stats["next_due_at"], reviewedAt, next.Reps, next.LapseCount, string(statsJSON), next.ID); err != nil {
		return 0, fmt.Errorf("update linglow srs item: %w", err)
	}
	return next.ID, nil
}

func (r *CourseRepository) getOrCreateSRSState(ctx context.Context, tx *sql.Tx, userCourseID, learningItemID int64) (linglowSRSState, error) {
	var state linglowSRSState
	var statsRaw string
	err := tx.QueryRowContext(ctx, `
		SELECT id, state, COALESCE(difficulty, 2.5), reps, lapse_count, COALESCE(stats_json::text, '{}')
		FROM srs_items
		WHERE user_course_id = ? AND learning_item_id = ?
	`, userCourseID, learningItemID).Scan(&state.ID, &state.State, &state.EF, &state.Reps, &state.LapseCount, &statsRaw)
	if err != nil && err != sql.ErrNoRows {
		return state, fmt.Errorf("get linglow srs item: %w", err)
	}
	if err == sql.ErrNoRows {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO srs_items (user_course_id, learning_item_id, state, stability, difficulty, stats_json, created_at, updated_at)
			VALUES (?, ?, 'new', 0, 2.5, '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id
		`, userCourseID, learningItemID).Scan(&state.ID); err != nil {
			return state, fmt.Errorf("create linglow srs item: %w", err)
		}
		state.State = "new"
		state.EF = 2.5
		state.Stats = map[string]interface{}{}
		return state, nil
	}
	state.Stats = map[string]interface{}{}
	_ = json.Unmarshal([]byte(statsRaw), &state.Stats)
	state.IntervalDays = intFromStats(state.Stats, "interval_days", 0)
	state.LearningStep = intFromStats(state.Stats, "learning_step", 0)
	if state.EF <= 0 {
		state.EF = 2.5
	}
	return state, nil
}

func applyLinglowSRS(current linglowSRSState, correct bool, quality *int, now time.Time) linglowSRSState {
	q := normalizeLinglowQuality(correct, quality)
	next := current
	if next.Stats == nil {
		next.Stats = map[string]interface{}{}
	}
	next.Stats["algo_version"] = "linglow_sm2_foundation_v1"
	next.Stats["last_quality"] = q
	next.Stats["last_review_at"] = now.UTC().Format(time.RFC3339)
	if next.EF <= 0 {
		next.EF = 2.5
	}
	if q == 0 {
		next.LapseCount++
		next.EF = maxFloat(1.3, next.EF-0.2)
		if next.State == "review" || next.State == "mastered" {
			next.State = "relearning"
			next.IntervalDays = maxInt(1, next.IntervalDays/2)
		} else {
			next.State = "learning"
			next.IntervalDays = 1
		}
		next.LearningStep = 0
		next.Stats["next_due_at"] = now.Add(time.Duration(next.IntervalDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
		next.Stats["interval_days"] = next.IntervalDays
		next.Stats["learning_step"] = next.LearningStep
		return next
	}
	next.LapseCount = 0
	switch next.State {
	case "new", "learning", "relearning":
		next.LearningStep++
		if q == 1 {
			next.State = "learning"
			next.IntervalDays = 1
		} else if next.LearningStep >= 2 {
			next.State = "review"
			next.Reps = maxInt(1, next.Reps+1)
			next.IntervalDays = 3
		} else {
			next.State = "learning"
			next.IntervalDays = 3
		}
	case "review", "mastered":
		next.EF = maxFloat(1.3, next.EF+(0.1-float64(5-sm2Quality(q))*(0.08+float64(5-sm2Quality(q))*0.02)))
		switch next.Reps {
		case 0:
			next.IntervalDays = 1
		case 1:
			next.IntervalDays = 6
		default:
			next.IntervalDays = maxInt(1, int(float64(maxInt(1, next.IntervalDays))*next.EF+0.999999))
		}
		next.Reps++
		next.State = "review"
	default:
		next.State = "learning"
		next.IntervalDays = 1
	}
	next.Stats["next_due_at"] = now.Add(time.Duration(next.IntervalDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
	next.Stats["interval_days"] = next.IntervalDays
	next.Stats["learning_step"] = next.LearningStep
	return next
}

func normalizeLinglowQuality(correct bool, quality *int) int {
	if !correct {
		return 0
	}
	if quality == nil {
		return 2
	}
	if *quality <= 0 {
		return 0
	}
	if *quality == 1 {
		return 1
	}
	if *quality >= 3 {
		return 3
	}
	return 2
}

func sm2Quality(q int) int {
	switch q {
	case 1:
		return 3
	case 2:
		return 4
	case 3:
		return 5
	default:
		return 0
	}
}

func intFromStats(stats map[string]interface{}, key string, fallback int) int {
	switch value := stats[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	default:
		return fallback
	}
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func (r *CourseRepository) getReviewQueueSummary(ctx context.Context, userCourseID, userID int64, courseCode string, canonicalSRSRead bool) (ReviewQueueSummary, error) {
	summary := ReviewQueueSummary{ByType: map[string]int{}, ReadSource: linglowSRSReadSource(canonicalSRSRead)}
	if !canonicalSRSRead {
		return r.getLegacyReviewQueueSummary(ctx, userID, courseCode)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT state, COUNT(*)
		FROM srs_items
		WHERE user_course_id = ?
			AND state IN ('learning', 'review', 'relearning')
			AND (due_at IS NULL OR due_at <= CURRENT_TIMESTAMP)
		GROUP BY state
	`, userCourseID)
	if err != nil {
		return summary, fmt.Errorf("count review queue states: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return summary, fmt.Errorf("scan review state count: %w", err)
		}
		summary.DueCount += count
		switch state {
		case "learning":
			summary.LearningCount = count
		case "review":
			summary.ReviewCount = count
		case "relearning":
			summary.RelearningCount = count
		}
	}
	if err := rows.Err(); err != nil {
		return summary, fmt.Errorf("iterate review state counts: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM srs_items
		WHERE user_course_id = ?
			AND state IN ('learning', 'review', 'relearning')
			AND due_at > CURRENT_TIMESTAMP
	`, userCourseID).Scan(&summary.UpcomingCount); err != nil {
		return summary, fmt.Errorf("count upcoming review items: %w", err)
	}
	typeRows, err := r.db.QueryContext(ctx, `
		SELECT li.item_type, COUNT(*)
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE si.user_course_id = ?
			AND si.state IN ('learning', 'review', 'relearning')
			AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		GROUP BY li.item_type
	`, userCourseID)
	if err != nil {
		return summary, fmt.Errorf("count review queue types: %w", err)
	}
	defer typeRows.Close()
	for typeRows.Next() {
		var itemType string
		var count int
		if err := typeRows.Scan(&itemType, &count); err != nil {
			return summary, fmt.Errorf("scan review type count: %w", err)
		}
		summary.ByType[itemType] = count
	}
	if err := typeRows.Err(); err != nil {
		return summary, fmt.Errorf("iterate review type counts: %w", err)
	}
	return summary, nil
}

func (r *CourseRepository) getDailyRouteSummary(ctx context.Context, userCourseID, userID int64, courseCode string, canonicalSRSRead bool) (DailyRouteSummary, error) {
	summary := DailyRouteSummary{ByType: map[string]int{}, ReadSource: linglowSRSReadSource(canonicalSRSRead)}
	if !canonicalSRSRead {
		legacy, err := r.getLegacyReviewQueueSummary(ctx, userID, courseCode)
		if err != nil {
			return summary, err
		}
		summary.DueReviewCount = legacy.DueCount
		summary.ByType = legacy.ByType
		summary.ReadSource = legacy.ReadSource
		if err := r.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM learning_items li
			WHERE li.course_id = (
				SELECT course_id FROM user_courses WHERE id = ?
			)
				AND li.status = 'published'
				AND NOT EXISTS (
					SELECT 1 FROM srs_items si
					WHERE si.user_course_id = ? AND si.learning_item_id = li.id
				)
				AND NOT EXISTS (
					SELECT 1 FROM exercise_attempts ea
					WHERE ea.user_course_id = ? AND ea.learning_item_id = li.id
				)
		`, userCourseID, userCourseID, userCourseID).Scan(&summary.NewItemCount); err != nil {
			return summary, fmt.Errorf("count new legacy route items: %w", err)
		}
		return summary, nil
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM srs_items
		WHERE user_course_id = ?
			AND state IN ('learning', 'review', 'relearning')
			AND (due_at IS NULL OR due_at <= CURRENT_TIMESTAMP)
	`, userCourseID).Scan(&summary.DueReviewCount); err != nil {
		return summary, fmt.Errorf("count due route items: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM learning_items li
		WHERE li.course_id = (
			SELECT course_id FROM user_courses WHERE id = ?
		)
			AND li.status = 'published'
			AND NOT EXISTS (
				SELECT 1 FROM srs_items si
				WHERE si.user_course_id = ? AND si.learning_item_id = li.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ? AND ea.learning_item_id = li.id
			)
	`, userCourseID, userCourseID, userCourseID).Scan(&summary.NewItemCount); err != nil {
		return summary, fmt.Errorf("count new route items: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT li.item_type, COUNT(*)
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE si.user_course_id = ?
			AND si.state IN ('learning', 'review', 'relearning')
			AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		GROUP BY li.item_type
	`, userCourseID)
	if err != nil {
		return summary, fmt.Errorf("count route items by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var itemType string
		var count int
		if err := rows.Scan(&itemType, &count); err != nil {
			return summary, fmt.Errorf("scan route type count: %w", err)
		}
		summary.ByType[itemType] = count
	}
	if err := rows.Err(); err != nil {
		return summary, fmt.Errorf("iterate route type counts: %w", err)
	}
	return summary, nil
}

func (r *CourseRepository) listDailyRouteReviewItems(ctx context.Context, userCourseID, userID int64, courseCode string, limit int, canonicalSRSRead bool) ([]DailyRouteItem, error) {
	if !canonicalSRSRead {
		return r.listLegacyDailyRouteReviewItems(ctx, userID, courseCode, limit)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			li.id, si.id, li.item_type, li.source_kind, li.source_id, COALESCE(li.title, ''), COALESCE(li.cefr_level, ''),
			si.state, si.due_at,
			COALESCE(d.code, ''), COALESCE(d.title, ''),
			COALESCE(l.code, ''), COALESCE(l.location_type, ''), COALESCE(l.title, ''),
			COALESCE(m.code, ''), COALESCE(m.title, '')
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		LEFT JOIN districts d ON d.id = li.district_id
		LEFT JOIN locations l ON l.id = li.location_id
		LEFT JOIN modules m ON m.id = li.module_id
		WHERE si.user_course_id = ?
			AND li.status = 'published'
			AND si.state IN ('learning', 'review', 'relearning')
			AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		ORDER BY si.due_at NULLS FIRST, si.last_review_at NULLS FIRST, si.id
		LIMIT ?
	`, userCourseID, limit)
	if err != nil {
		return nil, fmt.Errorf("list due route items: %w", err)
	}
	defer rows.Close()
	return scanDailyRouteItems(rows, true)
}

func (r *CourseRepository) getLegacyReviewQueueSummary(ctx context.Context, userID int64, courseCode string) (ReviewQueueSummary, error) {
	summary := ReviewQueueSummary{ByType: map[string]int{}, ReadSource: "legacy"}
	rows, err := r.db.QueryContext(ctx, `
		WITH course_scope AS (
			SELECT id, target_lang AS target_language
			FROM courses
			WHERE code = ?
		), legacy_due AS (
			SELECT li.id AS learning_item_id, li.item_type, COALESCE(NULLIF(uc.state, ''), 'review') AS state
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'word_card'
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
				AND li.status = 'published'
			WHERE uc.user_id = ?
				AND (uc.next_due_at IS NULL OR uc.next_due_at <= CURRENT_TIMESTAMP)
			GROUP BY li.id, li.item_type, COALESCE(NULLIF(uc.state, ''), 'review')

			UNION ALL

			SELECT li.id AS learning_item_id, li.item_type, COALESCE(NULLIF(gtm.state, ''), 'review') AS state
			FROM grammar_theory_memory gtm
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'grammar_theory_block'
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
				AND li.status = 'published'
			WHERE gtm.user_id = ?
				AND lower(gtm.language) = lower((SELECT target_language FROM course_scope))
				AND lower(gtm.course_id) = lower((SELECT target_language FROM course_scope))
				AND gtm.next_review_at <= CURRENT_TIMESTAMP
		)
		SELECT state, item_type, COUNT(DISTINCT learning_item_id)
		FROM legacy_due
		GROUP BY state, item_type
	`, courseCode, userID, userID)
	if err != nil {
		return summary, fmt.Errorf("count legacy review queue: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		var itemType string
		var count int
		if err := rows.Scan(&state, &itemType, &count); err != nil {
			return summary, fmt.Errorf("scan legacy review count: %w", err)
		}
		summary.DueCount += count
		summary.ByType[itemType] += count
		switch normalizeLinglowSRSState(state) {
		case "learning":
			summary.LearningCount += count
		case "relearning":
			summary.RelearningCount += count
		default:
			summary.ReviewCount += count
		}
	}
	if err := rows.Err(); err != nil {
		return summary, fmt.Errorf("iterate legacy review counts: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `
		WITH course_scope AS (
			SELECT id, target_lang AS target_language
			FROM courses
			WHERE code = ?
		), legacy_upcoming AS (
			SELECT li.id AS learning_item_id
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'word_card'
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
				AND li.status = 'published'
			WHERE uc.user_id = ? AND uc.next_due_at > CURRENT_TIMESTAMP
			GROUP BY li.id

			UNION

			SELECT li.id AS learning_item_id
			FROM grammar_theory_memory gtm
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'grammar_theory_block'
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
				AND li.status = 'published'
			WHERE gtm.user_id = ?
				AND lower(gtm.language) = lower((SELECT target_language FROM course_scope))
				AND lower(gtm.course_id) = lower((SELECT target_language FROM course_scope))
				AND gtm.next_review_at > CURRENT_TIMESTAMP
		)
		SELECT COUNT(*) FROM legacy_upcoming
	`, courseCode, userID, userID).Scan(&summary.UpcomingCount); err != nil {
		return summary, fmt.Errorf("count legacy upcoming review items: %w", err)
	}
	return summary, nil
}

func (r *CourseRepository) listLegacyDailyRouteReviewItems(ctx context.Context, userID int64, courseCode string, limit int) ([]DailyRouteItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH course_scope AS (
			SELECT id, target_lang AS target_language
			FROM courses
			WHERE code = ?
		), legacy_due AS (
			SELECT
				li.id AS learning_item_id,
				li.item_type,
				li.source_kind,
				li.source_id,
				COALESCE(li.title, '') AS title,
				COALESCE(li.cefr_level, '') AS cefr_level,
				COALESCE(NULLIF(uc.state, ''), 'review') AS state,
				MIN(uc.next_due_at) AS due_at,
				COALESCE(d.code, '') AS district_code,
				COALESCE(d.title, '') AS district_title,
				COALESCE(l.code, '') AS location_code,
				COALESCE(l.location_type, '') AS location_type,
				COALESCE(l.title, '') AS location_title,
				COALESCE(m.code, '') AS module_code,
				COALESCE(m.title, '') AS module_title
			FROM user_cards uc
			JOIN training_cards tc ON tc.id = uc.training_card_id
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'word_card'
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
				AND li.status = 'published'
			LEFT JOIN districts d ON d.id = li.district_id
			LEFT JOIN locations l ON l.id = li.location_id
			LEFT JOIN modules m ON m.id = li.module_id
			WHERE uc.user_id = ?
				AND (uc.next_due_at IS NULL OR uc.next_due_at <= CURRENT_TIMESTAMP)
			GROUP BY li.id, li.item_type, li.source_kind, li.source_id, li.title, li.cefr_level, state,
				d.code, d.title, l.code, l.location_type, l.title, m.code, m.title

			UNION ALL

			SELECT
				li.id AS learning_item_id,
				li.item_type,
				li.source_kind,
				li.source_id,
				COALESCE(li.title, '') AS title,
				COALESCE(li.cefr_level, '') AS cefr_level,
				COALESCE(NULLIF(gtm.state, ''), 'review') AS state,
				gtm.next_review_at AS due_at,
				COALESCE(d.code, '') AS district_code,
				COALESCE(d.title, '') AS district_title,
				COALESCE(l.code, '') AS location_code,
				COALESCE(l.location_type, '') AS location_type,
				COALESCE(l.title, '') AS location_title,
				COALESCE(m.code, '') AS module_code,
				COALESCE(m.title, '') AS module_title
			FROM grammar_theory_memory gtm
			JOIN learning_items li ON li.course_id = (SELECT id FROM course_scope)
				AND li.source_kind = 'grammar_theory_block'
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
				AND li.status = 'published'
			LEFT JOIN districts d ON d.id = li.district_id
			LEFT JOIN locations l ON l.id = li.location_id
			LEFT JOIN modules m ON m.id = li.module_id
			WHERE gtm.user_id = ?
				AND lower(gtm.language) = lower((SELECT target_language FROM course_scope))
				AND lower(gtm.course_id) = lower((SELECT target_language FROM course_scope))
				AND gtm.next_review_at <= CURRENT_TIMESTAMP
		)
		SELECT
			learning_item_id, item_type, source_kind, source_id, title, cefr_level, state, due_at,
			district_code, district_title, location_code, location_type, location_title, module_code, module_title
		FROM legacy_due
		ORDER BY due_at NULLS FIRST, learning_item_id
		LIMIT ?
	`, courseCode, userID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list legacy due route items: %w", err)
	}
	defer rows.Close()
	return scanLegacyDailyRouteItems(rows)
}

func (r *CourseRepository) listDailyRouteNewItems(ctx context.Context, userCourseID, courseID int64, limit int) ([]DailyRouteItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			li.id, li.item_type, li.source_kind, li.source_id, COALESCE(li.title, ''), COALESCE(li.cefr_level, ''),
			COALESCE(d.code, ''), COALESCE(d.title, ''),
			COALESCE(l.code, ''), COALESCE(l.location_type, ''), COALESCE(l.title, ''),
			COALESCE(m.code, ''), COALESCE(m.title, '')
		FROM learning_items li
		LEFT JOIN districts d ON d.id = li.district_id
		LEFT JOIN locations l ON l.id = li.location_id
		LEFT JOIN modules m ON m.id = li.module_id
		WHERE li.course_id = ?
			AND li.status = 'published'
			AND NOT EXISTS (
				SELECT 1 FROM srs_items si
				WHERE si.user_course_id = ? AND si.learning_item_id = li.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ? AND ea.learning_item_id = li.id
			)
		ORDER BY d.sort_order NULLS LAST, l.sort_order NULLS LAST, m.sort_order NULLS LAST, li.id
		LIMIT ?
	`, courseID, userCourseID, userCourseID, limit)
	if err != nil {
		return nil, fmt.Errorf("list new route items: %w", err)
	}
	defer rows.Close()
	return scanDailyRouteItems(rows, false)
}

func scanDailyRouteItems(rows *sql.Rows, includeSRS bool) ([]DailyRouteItem, error) {
	out := []DailyRouteItem{}
	for rows.Next() {
		var item DailyRouteItem
		var dueAt sql.NullTime
		if includeSRS {
			var srsItemID int64
			if err := rows.Scan(
				&item.LearningItemID,
				&srsItemID,
				&item.Type,
				&item.SourceKind,
				&item.SourceID,
				&item.Title,
				&item.CEFRLevel,
				&item.State,
				&dueAt,
				&item.DistrictCode,
				&item.DistrictTitle,
				&item.LocationCode,
				&item.LocationType,
				&item.LocationTitle,
				&item.ModuleCode,
				&item.ModuleTitle,
			); err != nil {
				return nil, fmt.Errorf("scan route review item: %w", err)
			}
			item.SRSItemID = &srsItemID
		} else if err := rows.Scan(
			&item.LearningItemID,
			&item.Type,
			&item.SourceKind,
			&item.SourceID,
			&item.Title,
			&item.CEFRLevel,
			&item.DistrictCode,
			&item.DistrictTitle,
			&item.LocationCode,
			&item.LocationType,
			&item.LocationTitle,
			&item.ModuleCode,
			&item.ModuleTitle,
		); err != nil {
			return nil, fmt.Errorf("scan route new item: %w", err)
		}
		item.Mode = dailyRouteMode(item.Type)
		if !includeSRS {
			item.State = "new"
		}
		if dueAt.Valid {
			formatted := dueAt.Time.UTC().Format(time.RFC3339)
			item.DueAt = &formatted
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate route items: %w", err)
	}
	return out, nil
}

func scanLegacyDailyRouteItems(rows *sql.Rows) ([]DailyRouteItem, error) {
	out := []DailyRouteItem{}
	for rows.Next() {
		var item DailyRouteItem
		var dueAt sql.NullTime
		if err := rows.Scan(
			&item.LearningItemID,
			&item.Type,
			&item.SourceKind,
			&item.SourceID,
			&item.Title,
			&item.CEFRLevel,
			&item.State,
			&dueAt,
			&item.DistrictCode,
			&item.DistrictTitle,
			&item.LocationCode,
			&item.LocationType,
			&item.LocationTitle,
			&item.ModuleCode,
			&item.ModuleTitle,
		); err != nil {
			return nil, fmt.Errorf("scan legacy route review item: %w", err)
		}
		item.State = normalizeLinglowSRSState(item.State)
		item.Mode = dailyRouteMode(item.Type)
		if dueAt.Valid {
			formatted := dueAt.Time.UTC().Format(time.RFC3339)
			item.DueAt = &formatted
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy route items: %w", err)
	}
	return out, nil
}

func dailyRouteMode(itemType string) string {
	switch itemType {
	case "word":
		return "word_training"
	case "grammar_chapter", "grammar_concept", "grammar_theory_block", "grammar_question":
		return "grammar"
	case "reading_text", "reading_question":
		return "reading"
	case "speaking_task":
		return "speaking"
	default:
		return itemType
	}
}

func linglowSRSReadSource(canonical bool) string {
	if canonical {
		return "canonical"
	}
	return "legacy"
}

func normalizeJSONObject(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	return raw
}

func percent(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

// GetCourseMap returns the Linglow v2 course map with the current user's course enrollment when present.
func (r *CourseRepository) GetCourseMap(ctx context.Context, courseCode string, userID int64) (*CourseMap, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}

	var result CourseMap
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, code, title, target_lang, teaching_locale, ui_locale, status, city_name
		FROM courses
		WHERE code = ?
	`, courseCode).Scan(
		&result.Course.ID,
		&result.Course.Code,
		&result.Course.Title,
		&result.Course.TargetLanguage,
		&result.Course.NativeLanguage,
		&result.Course.UILocale,
		&result.Course.Status,
		&result.Course.CityName,
	); err != nil {
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}
	result.Course.Slug = result.Course.Code

	if userID > 0 {
		var uc CourseMapUserCourse
		err := r.db.QueryRowContext(ctx, `
			SELECT id, status
			FROM user_courses
			WHERE user_id = ? AND course_id = ?
		`, userID, result.Course.ID).Scan(&uc.ID, &uc.Status)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("get user course: %w", err)
		}
		if err == nil {
			result.UserCourse = &uc
		}
	}

	districts, err := r.listCourseDistricts(ctx, result.Course.ID)
	if err != nil {
		return nil, err
	}
	locations, err := r.listCourseLocations(ctx, result.Course.ID)
	if err != nil {
		return nil, err
	}
	modules, err := r.listCourseModules(ctx, result.Course.ID)
	if err != nil {
		return nil, err
	}
	items, err := r.listCourseItems(ctx, result.Course.ID)
	if err != nil {
		return nil, err
	}

	itemsByModule := make(map[int64][]CourseMapItem)
	for _, item := range items {
		itemsByModule[item.moduleID] = append(itemsByModule[item.moduleID], item.CourseMapItem)
		result.Totals.Items++
		if result.Totals.ByType == nil {
			result.Totals.ByType = make(map[string]int)
		}
		result.Totals.ByType[item.Type]++
	}
	modulesByLocation := make(map[int64][]CourseMapModule)
	for _, module := range modules {
		module.Items = itemsByModule[module.ID]
		if module.Items == nil {
			module.Items = []CourseMapItem{}
		}
		modulesByLocation[module.locationID] = append(modulesByLocation[module.locationID], module.CourseMapModule)
		result.Totals.Modules++
	}
	locationsByDistrict := make(map[int64][]CourseMapLocation)
	for _, location := range locations {
		location.Modules = modulesByLocation[location.ID]
		if location.Modules == nil {
			location.Modules = []CourseMapModule{}
		}
		locationsByDistrict[location.districtID] = append(locationsByDistrict[location.districtID], location.CourseMapLocation)
		result.Totals.Locations++
	}
	result.Districts = make([]CourseMapDistrict, 0, len(districts))
	for _, district := range districts {
		district.Locations = locationsByDistrict[district.ID]
		if district.Locations == nil {
			district.Locations = []CourseMapLocation{}
		}
		result.Districts = append(result.Districts, district)
		result.Totals.Districts++
	}
	if result.Totals.ByType == nil {
		result.Totals.ByType = map[string]int{}
	}
	return &result, nil
}

type courseMapLocationRow struct {
	CourseMapLocation
	districtID int64
}

type courseMapModuleRow struct {
	CourseMapModule
	locationID int64
}

type courseMapItemRow struct {
	CourseMapItem
	moduleID int64
}

func (r *CourseRepository) listCourseDistricts(ctx context.Context, courseID int64) ([]CourseMapDistrict, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, level_code, title, sort_order, status
		FROM districts
		WHERE course_id = ?
		ORDER BY sort_order, id
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list districts: %w", err)
	}
	defer rows.Close()

	var districts []CourseMapDistrict
	for rows.Next() {
		var d CourseMapDistrict
		if err := rows.Scan(&d.ID, &d.Code, &d.LevelCode, &d.Title, &d.Order, &d.Status); err != nil {
			return nil, fmt.Errorf("scan district: %w", err)
		}
		districts = append(districts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate districts: %w", err)
	}
	return districts, nil
}

func (r *CourseRepository) listCourseLocations(ctx context.Context, courseID int64) ([]courseMapLocationRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.district_id, l.code, l.location_type, l.title, l.sort_order, l.status
		FROM locations l
		JOIN districts d ON d.id = l.district_id
		WHERE d.course_id = ?
		ORDER BY d.sort_order, l.sort_order, l.id
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer rows.Close()

	var locations []courseMapLocationRow
	for rows.Next() {
		var l courseMapLocationRow
		if err := rows.Scan(&l.ID, &l.districtID, &l.Code, &l.LocationType, &l.Title, &l.Order, &l.Status); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}
	return locations, nil
}

func (r *CourseRepository) listCourseModules(ctx context.Context, courseID int64) ([]courseMapModuleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, location_id, code, module_type, title, source_kind, source_id, sort_order, status
		FROM modules
		WHERE course_id = ? AND location_id IS NOT NULL
		ORDER BY sort_order, id
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	defer rows.Close()

	var modules []courseMapModuleRow
	for rows.Next() {
		var m courseMapModuleRow
		var sourceKind, sourceID sql.NullString
		if err := rows.Scan(&m.ID, &m.locationID, &m.Code, &m.Type, &m.Title, &sourceKind, &sourceID, &m.Order, &m.Status); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		m.SourceKind = sourceKind.String
		m.SourceID = sourceID.String
		modules = append(modules, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate modules: %w", err)
	}
	return modules, nil
}

func (r *CourseRepository) listCourseItems(ctx context.Context, courseID int64) ([]courseMapItemRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, module_id, item_type, source_kind, source_id, title, cefr_level, status
		FROM learning_items
		WHERE course_id = ? AND module_id IS NOT NULL
		ORDER BY id
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list learning items: %w", err)
	}
	defer rows.Close()

	var items []courseMapItemRow
	for rows.Next() {
		var item courseMapItemRow
		var title, cefrLevel sql.NullString
		if err := rows.Scan(&item.ID, &item.moduleID, &item.Type, &item.SourceKind, &item.SourceID, &title, &cefrLevel, &item.Status); err != nil {
			return nil, fmt.Errorf("scan learning item: %w", err)
		}
		item.Title = title.String
		item.CEFRLevel = cefrLevel.String
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learning items: %w", err)
	}
	return items, nil
}

// MapLegacyContentForLearning idempotently maps existing DB-first/bundle-derived content into Linglow v2 modules and learning_items.
func (r *CourseRepository) MapLegacyContentForLearning(ctx context.Context, lc config.LearningConfig) (*ContentMappingSummary, error) {
	return r.MapLegacyContent(ctx, CourseCodeForLearning(lc), strings.TrimSpace(strings.ToLower(lc.GrammarBundleID)))
}

// MapLegacyContent idempotently maps existing legacy content tables into Linglow v2 content tables.
func (r *CourseRepository) MapLegacyContent(ctx context.Context, courseCode, bundleID string) (*ContentMappingSummary, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	bundleID = strings.TrimSpace(strings.ToLower(bundleID))
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	if bundleID == "" {
		return nil, fmt.Errorf("bundle id is empty")
	}

	var courseID int64
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}
	var modulesBefore, itemsBefore int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM modules WHERE course_id = ?`, courseID).Scan(&modulesBefore); err != nil {
		return nil, fmt.Errorf("count modules before: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_items WHERE course_id = ?`, courseID).Scan(&itemsBefore); err != nil {
		return nil, fmt.Errorf("count learning items before: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin content mapping: %w", err)
	}
	defer tx.Rollback()

	statements := []struct {
		query      string
		usesBundle bool
	}{
		{mapGrammarSectionModulesSQL, true},
		{mapGrammarChapterItemsSQL, true},
		{mapGrammarTheoryBlockItemsSQL, true},
		{mapReadingCategoryModulesSQL, false},
		{mapReadingTextItemsSQL, false},
		{mapSpeakingCategoryModulesSQL, false},
		{mapSpeakingTaskItemsSQL, false},
		{mapWordSetModulesSQL, false},
		{mapWordCardItemsSQL, false},
	}
	for _, stmt := range statements {
		args := []interface{}{courseID}
		if stmt.usesBundle {
			args = []interface{}{bundleID, courseID}
		}
		if _, err := tx.ExecContext(ctx, stmt.query, args...); err != nil {
			return nil, fmt.Errorf("map legacy content: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit content mapping: %w", err)
	}

	var modulesAfter, itemsAfter int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM modules WHERE course_id = ?`, courseID).Scan(&modulesAfter); err != nil {
		return nil, fmt.Errorf("count modules after: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_items WHERE course_id = ?`, courseID).Scan(&itemsAfter); err != nil {
		return nil, fmt.Errorf("count learning items after: %w", err)
	}
	return &ContentMappingSummary{
		CourseCode:     courseCode,
		CourseID:       courseID,
		ModulesCreated: modulesAfter - modulesBefore,
		ItemsCreated:   itemsAfter - itemsBefore,
		ModulesTotal:   modulesAfter,
		ItemsTotal:     itemsAfter,
	}, nil
}

const levelDistrictJoinSQL = `
JOIN districts d ON d.course_id = c.id
    AND d.level_code = CASE
        WHEN upper(coalesce(src.level, '')) IN ('A0', 'A1', 'A2', 'B1', 'B2', 'C1') THEN upper(src.level)
        ELSE 'A0'
    END`

var mapGrammarSectionModulesSQL = `
WITH src AS (
    SELECT section_id AS source_id, title, level, sort_order
    FROM grammar_content_sections
    WHERE bundle_id = ?
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'grammar_section:' || source_id, 'grammar', title, 'grammar_section', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapGrammarChapterItemsSQL = `
WITH src AS (
    SELECT chapter_id AS source_id, section_id, title, level, source_hash, sort_order
    FROM grammar_content_chapters
    WHERE bundle_id = ?
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'grammar_section' AND m.source_id = src.section_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'grammar_chapter', 'grammar_chapter', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapGrammarTheoryBlockItemsSQL = `
WITH src AS (
    SELECT
        q.chapter_id,
        q.theory_block_id,
        MIN(q.concept_id) FILTER (WHERE q.concept_id IS NOT NULL AND q.concept_id <> '') AS concept_id,
        MIN(ch.section_id) AS section_id,
        MIN(ch.level) AS level,
        MIN(ch.sort_order) AS chapter_sort_order,
        COUNT(*) AS question_count,
        md5(string_agg(q.source_hash, ',' ORDER BY q.question_id)) AS source_hash
    FROM grammar_training_content_questions q
    JOIN grammar_content_chapters ch ON ch.bundle_id = q.bundle_id AND ch.chapter_id = q.chapter_id
    WHERE q.bundle_id = ?
    GROUP BY q.chapter_id, q.theory_block_id
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'grammar_section' AND m.source_id = src.section_id
)
INSERT INTO learning_items (
    course_id, module_id, district_id, location_id, item_type, source_kind, source_id,
    title, cefr_level, content_hash, payload_json, status, updated_at
)
SELECT
    course_id,
    module_id,
    district_id,
    location_id,
    'grammar_theory_block',
    'grammar_theory_block',
    chapter_id || ':' || theory_block_id,
    COALESCE(concept_id, theory_block_id),
    level,
    source_hash,
    jsonb_build_object(
        'chapter_id', chapter_id,
        'theory_block_id', theory_block_id,
        'concept_id', concept_id,
        'question_count', question_count,
        'chapter_sort_order', chapter_sort_order
    ),
    'published',
    CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    payload_json = excluded.payload_json,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapReadingCategoryModulesSQL = `
WITH src AS (
    SELECT category_id AS source_id, title, level, sort_order
    FROM reading_categories
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'reading'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'reading_category:' || source_id, 'reading', title, 'reading_category', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapReadingTextItemsSQL = `
WITH src AS (
    SELECT text_id AS source_id, category_id, title, level, md5(reading_passage) AS source_hash
    FROM reading_texts
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'reading'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'reading_category' AND m.source_id = src.category_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'reading_text', 'reading_text', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapSpeakingCategoryModulesSQL = `
WITH src AS (
    SELECT category_id AS source_id, title, level, sort_order
    FROM speaking_categories
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'conversation'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'speaking_category:' || source_id, 'speaking', title, 'speaking_category', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapSpeakingTaskItemsSQL = `
WITH src AS (
    SELECT task_id AS source_id, category_id, title, level, md5(task_json) AS source_hash
    FROM speaking_tasks
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'conversation'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'speaking_category' AND m.source_id = src.category_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'speaking_task', 'speaking_task', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapWordSetModulesSQL = `
WITH src AS (
    SELECT id::text AS source_id, title, sort_order, 'A0' AS level
    FROM word_sets
    WHERE is_published = 1
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'word_set:' || source_id, 'word_set', title, 'word_set', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapWordCardItemsSQL = `
WITH src AS (
    SELECT DISTINCT ON (wc.id)
        wc.id::text AS source_id,
        wsi.word_set_id::text AS module_source_id,
        coalesce(nullif(wc.display_en, ''), wc.word) AS title,
        coalesce(wc.updated_at::text, wc.created_at::text, '') AS source_hash,
        'A0' AS level
    FROM word_cards wc
    JOIN word_set_items wsi ON wsi.word_card_id = wc.id
    JOIN word_sets ws ON ws.id = wsi.word_set_id AND ws.is_published = 1
    ORDER BY wc.id, wsi.sort_order
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'word_set' AND m.source_id = src.module_source_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'word', 'word_card', source_id, title, level, md5(source_hash), 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`
