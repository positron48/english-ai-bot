package service

import (
	"database/sql"
	"fmt"
	"strings"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/grammartrainingpack"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/verbtraining"
)

// UnlockedVerbScopes is shared by ordinary training and daily journeys. It never
// opens tenses from legacy settings or treats theory viewing as chapter passed.
func UnlockedVerbScopes(q database.DBTX, user int64, lc config.LearningConfig) ([]string, error) {
	if !strings.EqualFold(lc.TargetLang, "es") {
		return nil, nil
	}
	fsys, err := grammartrainingpack.PackFS(lc.GrammarBundleID)
	if err != nil {
		return models.DefaultSpanishVerbScopes(), nil
	}
	gates, err := verbtraining.LoadUnlockGates(fsys)
	if err != nil || gates == nil {
		return models.DefaultSpanishVerbScopes(), nil
	}
	allowed := map[string]bool{}
	for id := range gates.Chapters {
		if q == nil || user <= 0 {
			continue
		}
		var passed bool
		err := q.QueryRow(`SELECT passed_at IS NOT NULL FROM grammar_progress WHERE user_id=? AND chapter_id=?`, user, id).Scan(&passed)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("verb scope chapter progress: %w", err)
		}
		allowed[id] = passed
	}
	scopes := gates.EnabledScopes(allowed)
	if len(scopes) == 0 {
		return models.DefaultSpanishVerbScopes(), nil
	}
	return verbtraining.ExpandScopes(scopes), nil
}
