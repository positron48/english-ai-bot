package service

import (
	"context"
	"database/sql"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

// Reads enrollment and the canonical city mastery without enrolling or opening content.
func dailyJourneyOpenLevels(ctx context.Context, db *sql.DB, q database.ContextDBTX, user int64, course string) (map[string]bool, error) {
	var courseID, userCourseID int64
	var language, tier string
	err := q.QueryRowContext(ctx, `SELECT c.id,uc.id,c.target_lang,COALESCE(u.subscription_tier,'free') FROM user_courses uc JOIN courses c ON c.id=uc.course_id JOIN users u ON u.id=uc.user_id WHERE uc.user_id=? AND c.code=? AND uc.status='active' AND c.status='active'`, user, course).Scan(&courseID, &userCourseID, &language, &tier)
	if err == sql.ErrNoRows {
		return nil, dailyjourney.ErrAccess
	}
	if err != nil {
		return nil, err
	}
	repo := repository.NewCourseRepository(db, nil)
	var mastery *repository.CourseMastery
	if tx, ok := q.(*sql.Tx); ok {
		mastery, err = repo.BuildCourseMasteryTx(ctx, tx, user, courseID, userCourseID, course, language, models.ParseUserTier(tier))
	} else {
		mastery, err = repo.BuildCourseMastery(ctx, user, courseID, userCourseID, course, language, models.ParseUserTier(tier))
	}
	if err != nil {
		return nil, err
	}
	levels := map[string]bool{}
	for _, level := range mastery.Levels {
		levels[level.LevelCode] = level.Unlocked
	}
	return levels, nil
}

// DailyJourneyOpenLevels shares canonical enrollment/mastery guards with web-owned chat engines.
func DailyJourneyOpenLevels(ctx context.Context, db *sql.DB, q database.ContextDBTX, user int64, course string) (map[string]bool, error) {
	return dailyJourneyOpenLevels(ctx, db, q, user, course)
}
