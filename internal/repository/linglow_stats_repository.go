package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// StatsDay is one day in the weekly rhythm.
type StatsDay struct {
	Date          string `json:"date"`
	ActiveSeconds int    `json:"active_seconds"`
	AttemptCount  int    `json:"attempt_count"`
	Status        string `json:"status"` // done|today|empty
}

// SkillStat aggregates per-mode accuracy.
type SkillStat struct {
	Mode            string  `json:"mode"`
	AttemptCount    int     `json:"attempt_count"`
	CorrectCount    int     `json:"correct_count"`
	AccuracyPercent float64 `json:"accuracy_percent"`
}

// FavoriteDistrict is the most practiced district over the month.
type FavoriteDistrict struct {
	DistrictCode    string  `json:"district_code"`
	Title           string  `json:"title"`
	AttemptCount    int     `json:"attempt_count"`
	ProgressPercent float64 `json:"progress_percent"`
}

// LinglowStats is the aggregate response for the progress screen and headers.
type LinglowStats struct {
	Course     CourseMapCourse     `json:"course"`
	UserCourse CourseMapUserCourse `json:"user_course"`
	Streak     struct {
		CurrentDays int  `json:"current_days"`
		BestDays    int  `json:"best_days"`
		TodayActive bool `json:"today_active"`
	} `json:"streak"`
	Today struct {
		ActiveSeconds int `json:"active_seconds"`
		AttemptCount  int `json:"attempt_count"`
	} `json:"today"`
	Week  []StatsDay `json:"week"`
	Month struct {
		Month         string `json:"month"`
		ActiveMinutes int    `json:"active_minutes"`
		WordsLearned  int    `json:"words_learned"`
		TextsRead     int    `json:"texts_read"`
		ChatMessages  int    `json:"chat_messages"`
		ActiveDays    int    `json:"active_days"`
	} `json:"month"`
	Skills           []SkillStat       `json:"skills"`
	SkillsPeriod     string            `json:"skills_period"` // month|all
	FavoriteDistrict *FavoriteDistrict `json:"favorite_district,omitempty"`
	Achievements     []AchievementStat `json:"achievements"`
	Improvements     []ImprovementStat `json:"improvements"`
	Generated        string            `json:"generated_at"`
}

// AchievementStat is a computed (not persisted) achievement state.
type AchievementStat struct {
	Code     string `json:"code"`
	Value    int    `json:"value"`
	Unlocked bool   `json:"unlocked"`
}

// ImprovementStat is one rule-based "what to improve" suggestion.
type ImprovementStat struct {
	Kind         string  `json:"kind"`
	Mode         string  `json:"mode,omitempty"`
	Count        int     `json:"count,omitempty"`
	DistrictCode string  `json:"district_code,omitempty"`
	Title        string  `json:"title,omitempty"`
	Accuracy     float64 `json:"accuracy,omitempty"`
}

// dayActive reports whether a day counts towards the streak.
// A day counts only when the user actually completed training (any mode), not
// merely spent time on the site — time-only heartbeats record attempts == 0.
func dayActive(attempts, seconds int) bool {
	return attempts > 0
}

// GetStatsForUser builds the stats payload from daily aggregates and canonical tables.
func (r *CourseRepository) GetStatsForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode, month string) (*LinglowStats, error) {
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

	now := time.Now()
	today := now.Format("2006-01-02")
	if !validStatsMonth(month) {
		month = now.Format("2006-01")
	}
	monthStart := month + "-01"

	stats := &LinglowStats{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Generated:  now.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if courseMap.UserCourse != nil {
		stats.UserCourse = *courseMap.UserCourse
	}
	stats.Month.Month = month
	stats.Week = []StatsDay{}
	stats.Skills = []SkillStat{}

	// All aggregate days (newest first) — drives streak, today, week and month sums.
	rows, err := r.db.QueryContext(ctx, `
		SELECT CAST(local_date AS text), attempt_count, active_seconds
		FROM daily_course_stats
		WHERE user_course_id = ?
		ORDER BY local_date DESC
		LIMIT 1100
	`, userCourse.ID)
	if err != nil {
		return nil, fmt.Errorf("stats days: %w", err)
	}
	defer rows.Close()
	type dayRow struct {
		date     string
		attempts int
		seconds  int
	}
	byDate := map[string]dayRow{}
	var dates []string
	for rows.Next() {
		var d dayRow
		if err := rows.Scan(&d.date, &d.attempts, &d.seconds); err != nil {
			return nil, fmt.Errorf("scan stats day: %w", err)
		}
		if len(d.date) > 10 {
			d.date = d.date[:10]
		}
		byDate[d.date] = d
		dates = append(dates, d.date)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Today.
	if d, ok := byDate[today]; ok {
		stats.Today.ActiveSeconds = d.seconds
		stats.Today.AttemptCount = d.attempts
	}

	// Streak: walk back from today; an empty today does not break the streak yet.
	cursor := now
	if d, ok := byDate[today]; !ok || !dayActive(d.attempts, d.seconds) {
		cursor = cursor.AddDate(0, 0, -1)
	} else {
		stats.Streak.TodayActive = true
	}
	for {
		key := cursor.Format("2006-01-02")
		d, ok := byDate[key]
		if !ok || !dayActive(d.attempts, d.seconds) {
			break
		}
		stats.Streak.CurrentDays++
		cursor = cursor.AddDate(0, 0, -1)
	}
	// Best streak across all loaded days.
	best, run := 0, 0
	for i := len(dates) - 1; i >= 0; i-- { // oldest → newest
		d := byDate[dates[i]]
		if !dayActive(d.attempts, d.seconds) {
			run = 0
			continue
		}
		if i < len(dates)-1 {
			prev, _ := time.Parse("2006-01-02", dates[i+1])
			cur, _ := time.Parse("2006-01-02", dates[i])
			if cur.Sub(prev) != 24*time.Hour {
				run = 0
			}
		}
		run++
		if run > best {
			best = run
		}
	}
	if stats.Streak.CurrentDays > best {
		best = stats.Streak.CurrentDays
	}
	stats.Streak.BestDays = best

	// Week: last 7 calendar days, oldest first.
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		d := byDate[date]
		status := "empty"
		if dayActive(d.attempts, d.seconds) {
			status = "done"
		} else if date == today {
			status = "today"
		}
		stats.Week = append(stats.Week, StatsDay{Date: date, ActiveSeconds: d.seconds, AttemptCount: d.attempts, Status: status})
	}

	// Month sums from aggregates.
	for date, d := range byDate {
		if !strings.HasPrefix(date, month) {
			continue
		}
		stats.Month.ActiveMinutes += d.seconds / 60
		if dayActive(d.attempts, d.seconds) {
			stats.Month.ActiveDays++
		}
	}

	// Words learned this month: word SRS items that reached review/mastered.
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE si.user_course_id = ? AND li.item_type = 'word'
			AND si.state IN ('review', 'mastered')
			AND si.updated_at >= CAST(? AS date)
	`, userCourse.ID, monthStart).Scan(&stats.Month.WordsLearned); err != nil {
		return nil, fmt.Errorf("stats words learned: %w", err)
	}

	// Texts read and chat messages this month from learning events.
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE event_type = 'reading_text_completed'),
			COUNT(*) FILTER (WHERE event_type = 'chat_message_sent')
		FROM learning_events
		WHERE user_course_id = ? AND event_time >= CAST(? AS date)
	`, userCourse.ID, monthStart).Scan(&stats.Month.TextsRead, &stats.Month.ChatMessages); err != nil {
		return nil, fmt.Errorf("stats events: %w", err)
	}

	// Skills by mode: month first, all-time fallback.
	stats.SkillsPeriod = "month"
	skills, err := r.queryStatsSkills(ctx, userCourse.ID, monthStart)
	if err != nil {
		return nil, err
	}
	if len(skills) == 0 {
		stats.SkillsPeriod = "all"
		skills, err = r.queryStatsSkills(ctx, userCourse.ID, "")
		if err != nil {
			return nil, err
		}
	}
	stats.Skills = skills

	// Favorite district: most attempts this month (all-time fallback).
	fav, err := r.queryFavoriteDistrict(ctx, userCourse.ID, monthStart)
	if err != nil {
		return nil, err
	}
	if fav == nil {
		fav, err = r.queryFavoriteDistrict(ctx, userCourse.ID, "")
		if err != nil {
			return nil, err
		}
	}
	stats.FavoriteDistrict = fav

	// Achievements (computed, all-time).
	var textsTotal int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM learning_events
		WHERE user_course_id = ? AND event_type = 'reading_text_completed'
	`, userCourse.ID).Scan(&textsTotal); err != nil {
		return nil, fmt.Errorf("stats texts total: %w", err)
	}
	var wordsTotal int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE si.user_course_id = ? AND li.item_type = 'word' AND si.state IN ('review', 'mastered')
	`, userCourse.ID).Scan(&wordsTotal); err != nil {
		return nil, fmt.Errorf("stats words total: %w", err)
	}
	districtRows, err := r.listProgressByDistrict(ctx, userCourse.ID, courseMap.Course.ID)
	if err != nil {
		return nil, err
	}
	attemptedDistricts, expertDistricts, activeDistricts := 0, 0, 0
	var weakest *CourseProgressDistrict
	dueBacklog := 0
	for i := range districtRows {
		d := districtRows[i]
		if d.TotalItems == 0 {
			continue
		}
		activeDistricts++
		dueBacklog += d.DueReviewCount
		if d.AttemptedItems > 0 {
			attemptedDistricts++
			if weakest == nil || d.Weakness > weakest.Weakness {
				weakest = &districtRows[i]
			}
		}
		if d.ProgressPercent >= 80 {
			expertDistricts++
		}
	}
	stats.Achievements = []AchievementStat{
		{Code: "streak", Value: stats.Streak.CurrentDays, Unlocked: stats.Streak.CurrentDays >= 1},
		{Code: "reader", Value: textsTotal, Unlocked: textsTotal >= 1},
		{Code: "collector", Value: wordsTotal, Unlocked: wordsTotal >= 10},
		{Code: "explorer", Value: attemptedDistricts, Unlocked: attemptedDistricts >= 1},
		{Code: "expert", Value: expertDistricts, Unlocked: activeDistricts > 0 && expertDistricts == activeDistricts},
	}

	// Improvements: rule-based, top 3 by priority.
	stats.Improvements = []ImprovementStat{}
	if dueBacklog > 20 {
		stats.Improvements = append(stats.Improvements, ImprovementStat{Kind: "due_backlog", Count: dueBacklog})
	}
	var weakSkill *SkillStat
	for i := range stats.Skills {
		s := stats.Skills[i]
		if s.AttemptCount >= 20 && s.AccuracyPercent < 80 {
			if weakSkill == nil || s.AccuracyPercent < weakSkill.AccuracyPercent {
				weakSkill = &stats.Skills[i]
			}
		}
	}
	if weakSkill != nil {
		stats.Improvements = append(stats.Improvements, ImprovementStat{Kind: "mode_accuracy", Mode: weakSkill.Mode, Accuracy: weakSkill.AccuracyPercent})
	}
	if weakest != nil && weakest.Weakness > 0 {
		stats.Improvements = append(stats.Improvements, ImprovementStat{Kind: "weak_district", DistrictCode: weakest.DistrictCode, Title: weakest.Title})
	}
	if len(stats.Improvements) < 3 {
		weekAgo := now.AddDate(0, 0, -7)
		var readWeek, chatWeek int
		if err := r.db.QueryRowContext(ctx, `
			SELECT
				COUNT(*) FILTER (WHERE event_type = 'reading_text_completed'),
				COUNT(*) FILTER (WHERE event_type = 'chat_message_sent')
			FROM learning_events
			WHERE user_course_id = ? AND event_time >= ?
		`, userCourse.ID, weekAgo).Scan(&readWeek, &chatWeek); err == nil {
			if readWeek == 0 {
				stats.Improvements = append(stats.Improvements, ImprovementStat{Kind: "no_reading"})
			}
			if chatWeek == 0 && len(stats.Improvements) < 3 {
				stats.Improvements = append(stats.Improvements, ImprovementStat{Kind: "no_chat"})
			}
		}
	}
	if len(stats.Improvements) > 3 {
		stats.Improvements = stats.Improvements[:3]
	}

	return stats, nil
}

func validStatsMonth(month string) bool {
	if len(month) != 7 {
		return false
	}
	_, err := time.Parse("2006-01", month)
	return err == nil
}

func (r *CourseRepository) queryStatsSkills(ctx context.Context, userCourseID int64, monthStart string) ([]SkillStat, error) {
	query := `
		SELECT mode, SUM(attempt_count), SUM(correct_count)
		FROM mode_daily_stats
		WHERE user_course_id = ?`
	args := []interface{}{userCourseID}
	if monthStart != "" {
		query += ` AND local_date >= CAST(? AS date)`
		args = append(args, monthStart)
	}
	query += ` GROUP BY mode HAVING SUM(attempt_count) > 0 ORDER BY SUM(attempt_count) DESC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("stats skills: %w", err)
	}
	defer rows.Close()
	out := []SkillStat{}
	for rows.Next() {
		var s SkillStat
		if err := rows.Scan(&s.Mode, &s.AttemptCount, &s.CorrectCount); err != nil {
			return nil, err
		}
		if s.AttemptCount > 0 {
			s.AccuracyPercent = float64(s.CorrectCount) / float64(s.AttemptCount) * 100
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *CourseRepository) queryFavoriteDistrict(ctx context.Context, userCourseID int64, monthStart string) (*FavoriteDistrict, error) {
	query := `
		SELECT d.code, d.title, COUNT(*) AS attempts,
			COALESCE((
				SELECT 100.0 * COUNT(DISTINCT ea2.learning_item_id) / NULLIF(COUNT(DISTINCT li2.id), 0)
				FROM learning_items li2
				LEFT JOIN exercise_attempts ea2
					ON ea2.learning_item_id = li2.id AND ea2.user_course_id = ?
				WHERE li2.district_id = d.id AND li2.status = 'published'
			), 0)
		FROM exercise_attempts ea
		JOIN learning_items li ON li.id = ea.learning_item_id
		JOIN districts d ON d.id = li.district_id
		WHERE ea.user_course_id = ?`
	args := []interface{}{userCourseID, userCourseID}
	if monthStart != "" {
		query += ` AND ea.answered_at >= CAST(? AS date)`
		args = append(args, monthStart)
	}
	query += ` GROUP BY d.id, d.code, d.title ORDER BY attempts DESC LIMIT 1`
	var fav FavoriteDistrict
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&fav.DistrictCode, &fav.Title, &fav.AttemptCount, &fav.ProgressPercent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("stats favorite district: %w", err)
	}
	return &fav, nil
}
