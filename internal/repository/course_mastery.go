package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/models"
)

// CEFR levels tracked by Linglow city districts and mastery progression.
var masteryLevelOrder = []string{"A0", "A1", "A2", "B1", "B2", "C1"}

// Cumulative mastered/known word counts required to complete each CEFR band.
var masteryWordCumulativeThreshold = map[string]int{
	"A0": 150,
	"A1": 350,
	"A2": 700,
	"B1": 1300,
	"B2": 2500,
	"C1": 5000,
}

var masteryReadingTarget = map[string]int{
	"A0": 10,
	"A1": 20,
	"A2": 30,
	"B1": 40,
	"B2": 40,
	"C1": 40,
}

var masteryConversationTarget = map[string]int{
	"A0": 5,
	"A1": 20,
	"A2": 20,
	"B1": 20,
	"B2": 20,
	"C1": 20,
}

var masteryPictureTarget = map[string]int{
	"A0": 5,
	"A1": 20,
	"A2": 20,
	"B1": 20,
	"B2": 20,
	"C1": 20,
}

// CourseMastery is the unified course progress model shared by dashboard, progress,
// city map and district screens.
type CourseMastery struct {
	CurrentLevelCode string                  `json:"current_level_code"`
	NextLevelCode    string                  `json:"next_level_code,omitempty"`
	ProgressPercent  float64                 `json:"progress_percent"`
	Levels           []CourseMasteryLevel    `json:"levels"`
	Remainder        *CourseMasteryRemainder `json:"remainder,omitempty"`
}

// CourseMasteryLevel is per-CEFR district mastery with activity metrics.
type CourseMasteryLevel struct {
	LevelCode      string                         `json:"level_code"`
	Unlocked       bool                           `json:"unlocked"`
	Current        bool                           `json:"current"`
	MasteryPercent float64                        `json:"mastery_percent"`
	CanOpenNext    bool                           `json:"can_open_next"`
	Metrics        map[string]CourseMasteryMetric `json:"metrics"`
}

// CourseMasteryMetric tracks one activity type for a level.
type CourseMasteryMetric struct {
	Done     int     `json:"done"`
	Total    int     `json:"total"`
	Target   int     `json:"target"`
	Percent  float64 `json:"percent"`
	Included bool    `json:"included"`
}

// CourseMasteryRemainder describes what is left on the grammar vs. vocabulary unlock paths.
type CourseMasteryRemainder struct {
	GrammarRemaining int `json:"grammar_remaining"`
	WordsRemaining   int `json:"words_remaining"`
}

type masteryRawData struct {
	grammarTarget     map[string]int
	grammarDone       map[string]int
	wordMasteredTotal int
	readingTotal      map[string]int
	readingDone   map[string]int
	convTotal     map[string]int
	convDone      map[string]int
	pictureTotal  map[string]int
	pictureDone   map[string]int
	placementMax  int
}

func masteryLevelIndex(code string) int {
	code = strings.ToUpper(strings.TrimSpace(code))
	for i, lv := range masteryLevelOrder {
		if lv == code {
			return i
		}
	}
	return -1
}

func nextMasteryLevel(code string) string {
	idx := masteryLevelIndex(code)
	if idx < 0 || idx >= len(masteryLevelOrder)-1 {
		return ""
	}
	return masteryLevelOrder[idx+1]
}

func wordPrevCumulativeThreshold(level string) int {
	idx := masteryLevelIndex(level)
	if idx <= 0 {
		return 0
	}
	return masteryWordCumulativeThreshold[masteryLevelOrder[idx-1]]
}

func wordCumulativeThreshold(level string) int {
	return masteryWordCumulativeThreshold[strings.ToUpper(strings.TrimSpace(level))]
}

// wordBandMetrics maps total mastered/known words to progress within one CEFR band.
// Example: 400 total → A0/A1 at 100%, A2 at (400-350)/(700-350).
func wordBandMetrics(level string, totalMastered int) (done, bandSize int, percent float64) {
	prev := wordPrevCumulativeThreshold(level)
	cum := wordCumulativeThreshold(level)
	bandSize = cum - prev
	if bandSize <= 0 {
		return 0, 0, 0
	}
	progress := totalMastered - prev
	if progress < 0 {
		progress = 0
	}
	if progress > bandSize {
		progress = bandSize
	}
	done = progress
	percent = float64(done) * 100 / float64(bandSize)
	if percent > 100 {
		percent = 100
	}
	return done, bandSize, percent
}

func wordUnlockCumulative(totalMastered int, level string) bool {
	cum := wordCumulativeThreshold(level)
	return cum > 0 && totalMastered >= cum
}

func cappedTarget(configured, available int) int {
	if available <= 0 {
		return 0
	}
	if configured <= 0 {
		return available
	}
	if configured > available {
		return available
	}
	return configured
}

func metricPercent(done, target int) float64 {
	if target <= 0 {
		if done > 0 {
			return 100
		}
		return 0
	}
	p := float64(done) * 100 / float64(target)
	if p > 100 {
		return 100
	}
	if p < 0 {
		return 0
	}
	return p
}

func buildMasteryMetric(done, total, target int, included bool) CourseMasteryMetric {
	if target <= 0 && total > 0 {
		target = total
	}
	return CourseMasteryMetric{
		Done:     done,
		Total:    total,
		Target:   target,
		Percent:  metricPercent(done, target),
		Included: included,
	}
}

func weightedMasteryPercent(metrics map[string]CourseMasteryMetric, proFeatures bool, hasConversation, hasPicture bool) float64 {
	type weightDef struct {
		key    string
		weight float64
	}
	weights := []weightDef{
		{"grammar", 0.45},
		{"words", 0.35},
		{"reading", 0.20},
	}
	if proFeatures {
		weights = []weightDef{
			{"grammar", 0.35},
			{"words", 0.30},
			{"reading", 0.15},
		}
		if hasConversation {
			weights = append(weights, weightDef{"conversation", 0.10})
		}
		if hasPicture {
			weights = append(weights, weightDef{"picture", 0.10})
		}
	}

	var sumWeight float64
	var sum float64
	for _, w := range weights {
		m, ok := metrics[w.key]
		if !ok || !m.Included || (m.Target <= 0 && m.Total <= 0) {
			continue
		}
		sumWeight += w.weight
		sum += w.weight * m.Percent
	}
	if sumWeight <= 0 {
		return 0
	}
	return sum / sumWeight
}

func levelUnlockedViaPlacement(level string, placementMax int) bool {
	idx := masteryLevelIndex(level)
	return idx >= 0 && placementMax >= 0 && idx <= placementMax
}

func grammarUnlock(done, target int) bool {
	return target > 0 && done >= target
}

func unlockProgressPercent(grammarDone, grammarTarget, wordMastered, wordTarget int) float64 {
	g := metricPercent(grammarDone, grammarTarget)
	w := metricPercent(wordMastered, wordTarget)
	if g > w {
		return g
	}
	return w
}

// BuildCourseMastery aggregates deterministic mastery for a user/course pair.
func (r *CourseRepository) BuildCourseMastery(ctx context.Context, userID, courseID, userCourseID int64, courseCode, targetLang string, tier models.UserTier) (*CourseMastery, error) {
	if userID == 0 || courseID == 0 {
		return &CourseMastery{Levels: []CourseMasteryLevel{}}, nil
	}
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	bundleID := GrammarBundleIDForCourse(courseCode)
	raw, err := r.loadMasteryRawData(ctx, userID, courseID, userCourseID, courseCode, bundleID, targetLang)
	if err != nil {
		return nil, err
	}

	pro := models.TierAllowsFeature(tier, "conversation") || models.TierAllowsFeature(tier, "picture_description")
	hasConversation := sumMap(raw.convTotal) > 0
	hasPicture := sumMap(raw.pictureTotal) > 0

	levels := make([]CourseMasteryLevel, 0, len(masteryLevelOrder))
	var currentLevel string
	var nextLevel string
	var headerProgress float64
	var remainder *CourseMasteryRemainder

	for i, level := range masteryLevelOrder {
		gTarget := raw.grammarTarget[level]
		gDone := raw.grammarDone[level]
		wDone, wBand, wPct := wordBandMetrics(level, raw.wordMasteredTotal)

		readTotal := raw.readingTotal[level]
		readTarget := cappedTarget(masteryReadingTarget[level], readTotal)
		readDone := raw.readingDone[level]
		if readDone > readTarget && readTarget > 0 {
			readDone = readTarget
		}

		convTotal := raw.convTotal[level]
		convTarget := cappedTarget(masteryConversationTarget[level], convTotal)
		convDone := raw.convDone[level]
		if convDone > convTarget && convTarget > 0 {
			convDone = convTarget
		}

		picTotal := raw.pictureTotal[level]
		picTarget := cappedTarget(masteryPictureTarget[level], picTotal)
		picDone := raw.pictureDone[level]
		if picDone > picTarget && picTarget > 0 {
			picDone = picTarget
		}

		metrics := map[string]CourseMasteryMetric{
			"grammar": buildMasteryMetric(gDone, gTarget, gTarget, gTarget > 0),
			"words": {
				Done:     wDone,
				Total:    wBand,
				Target:   wBand,
				Percent:  wPct,
				Included: wBand > 0,
			},
			"reading": buildMasteryMetric(readDone, readTotal, readTarget, readTotal > 0),
		}
		if pro && hasConversation {
			metrics["conversation"] = buildMasteryMetric(convDone, convTotal, convTarget, convTotal > 0)
		} else {
			metrics["conversation"] = buildMasteryMetric(convDone, convTotal, convTarget, false)
		}
		if pro && hasPicture {
			metrics["picture"] = buildMasteryMetric(picDone, picTotal, picTarget, picTotal > 0)
		} else {
			metrics["picture"] = buildMasteryMetric(picDone, picTotal, picTarget, false)
		}

		canOpenNext := grammarUnlock(gDone, gTarget) || wordUnlockCumulative(raw.wordMasteredTotal, level)
		unlocked := i == 0
		if !unlocked {
			prev := masteryLevelOrder[i-1]
			prevTarget := raw.grammarTarget[prev]
			prevDone := raw.grammarDone[prev]
			prevCanOpen := grammarUnlock(prevDone, prevTarget) || wordUnlockCumulative(raw.wordMasteredTotal, prev)
			unlocked = prevCanOpen || levelUnlockedViaPlacement(level, raw.placementMax)
		}

		masteryPct := weightedMasteryPercent(metrics, pro, hasConversation, hasPicture)
		levels = append(levels, CourseMasteryLevel{
			LevelCode:      level,
			Unlocked:       unlocked,
			Current:        false,
			MasteryPercent: masteryPct,
			CanOpenNext:    canOpenNext,
			Metrics:        metrics,
		})
	}

	for i := len(levels) - 1; i >= 0; i-- {
		if !levels[i].Unlocked {
			continue
		}
		if !levels[i].CanOpenNext || i == len(levels)-1 {
			currentLevel = levels[i].LevelCode
			for j := range levels {
				levels[j].Current = levels[j].LevelCode == currentLevel
			}
			nextLevel = nextMasteryLevel(currentLevel)
			lv := &levels[i]
			headerProgress = unlockProgressPercent(
				lv.Metrics["grammar"].Done, lv.Metrics["grammar"].Target,
				lv.Metrics["words"].Done, lv.Metrics["words"].Target,
			)
			gRem := lv.Metrics["grammar"].Target - lv.Metrics["grammar"].Done
			if gRem < 0 {
				gRem = 0
			}
			wRem := lv.Metrics["words"].Target - lv.Metrics["words"].Done
			if wRem < 0 {
				wRem = 0
			}
			remainder = &CourseMasteryRemainder{
				GrammarRemaining: gRem,
				WordsRemaining:   wRem,
			}
			break
		}
	}
	if currentLevel == "" {
		currentLevel = "A0"
		for j := range levels {
			levels[j].Current = levels[j].LevelCode == currentLevel
		}
		nextLevel = nextMasteryLevel(currentLevel)
		if len(levels) > 0 {
			lv := &levels[0]
			headerProgress = unlockProgressPercent(
				lv.Metrics["grammar"].Done, lv.Metrics["grammar"].Target,
				lv.Metrics["words"].Done, lv.Metrics["words"].Target,
			)
		}
	}

	return &CourseMastery{
		CurrentLevelCode: currentLevel,
		NextLevelCode:    nextLevel,
		ProgressPercent:  headerProgress,
		Levels:           levels,
		Remainder:        remainder,
	}, nil
}

func sumMap(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

func (r *CourseRepository) scanLevelCounts(ctx context.Context, query string, args []interface{}, dest map[string]int) error {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var level sql.NullString
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return err
		}
		if !level.Valid {
			continue
		}
		lv := strings.ToUpper(strings.TrimSpace(level.String))
		if lv == "" {
			continue
		}
		dest[lv] += count
	}
	return rows.Err()
}

func (r *CourseRepository) loadPlacementMaxOrder(ctx context.Context, userID int64, bundleID string, out *masteryRawData) error {
	var openedJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT opened_sections_json
		FROM grammar_placement_test
		WHERE user_id = ?`, userID).Scan(&openedJSON)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("placement test: %w", err)
	}
	if !openedJSON.Valid || strings.TrimSpace(openedJSON.String) == "" {
		return nil
	}
	var opened []string
	if err := json.Unmarshal([]byte(openedJSON.String), &opened); err != nil {
		return fmt.Errorf("parse placement sections: %w", err)
	}
	if len(opened) == 0 {
		return nil
	}
	placeholders := make([]string, len(opened))
	args := make([]interface{}, 0, len(opened)+1)
	args = append(args, bundleID)
	for i, id := range opened {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := fmt.Sprintf(`
		SELECT upper(level)
		FROM grammar_content_sections
		WHERE bundle_id = ? AND section_id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("placement section levels: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var level sql.NullString
		if err := rows.Scan(&level); err != nil {
			return err
		}
		if idx := masteryLevelIndex(level.String); idx > out.placementMax {
			out.placementMax = idx
		}
	}
	return rows.Err()
}

func (r *CourseRepository) loadMasteryRawData(ctx context.Context, userID, courseID, userCourseID int64, courseCode, bundleID, targetLang string) (*masteryRawData, error) {
	out := &masteryRawData{
		grammarTarget: make(map[string]int),
		grammarDone:   make(map[string]int),
		readingTotal:  make(map[string]int),
		readingDone:   make(map[string]int),
		convTotal:     make(map[string]int),
		convDone:      make(map[string]int),
		pictureTotal:  make(map[string]int),
		pictureDone:   make(map[string]int),
		placementMax:  -1,
	}

	if bundleID != "" {
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(ch.level), COUNT(*)
			FROM grammar_content_chapters ch
			JOIN grammar_published_items gpi ON gpi.item_type = 'chapter' AND gpi.item_id = ch.chapter_id AND gpi.is_published = 1
			WHERE ch.bundle_id = ?
			GROUP BY upper(ch.level)`, []interface{}{bundleID}, out.grammarTarget); err != nil {
			return nil, fmt.Errorf("grammar chapter targets: %w", err)
		}
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(s.level), COUNT(*)
			FROM grammar_content_sections s
			JOIN grammar_published_items gpi ON gpi.item_type = 'section' AND gpi.item_id = s.section_id AND gpi.is_published = 1
			WHERE s.bundle_id = ?
			GROUP BY upper(s.level)`, []interface{}{bundleID}, out.grammarTarget); err != nil {
			return nil, fmt.Errorf("grammar section targets: %w", err)
		}
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(ch.level), COUNT(*)
			FROM grammar_content_chapters ch
			JOIN grammar_published_items gpi ON gpi.item_type = 'chapter' AND gpi.item_id = ch.chapter_id AND gpi.is_published = 1
			JOIN grammar_progress gp ON gp.chapter_id = ch.chapter_id AND gp.user_id = ? AND gp.passed_at IS NOT NULL
			WHERE ch.bundle_id = ?
			GROUP BY upper(ch.level)`, []interface{}{userID, bundleID}, out.grammarDone); err != nil {
			return nil, fmt.Errorf("grammar chapter done: %w", err)
		}
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(s.level), COUNT(*)
			FROM grammar_content_sections s
			JOIN grammar_published_items gpi ON gpi.item_type = 'section' AND gpi.item_id = s.section_id AND gpi.is_published = 1
			JOIN (
				SELECT scope_id, MAX(score) AS best_score
				FROM grammar_test_attempts
				WHERE user_id = ? AND scope_type = 'category'
				GROUP BY scope_id
			) cat ON cat.scope_id = s.section_id AND cat.best_score >= 50
			WHERE s.bundle_id = ?
			GROUP BY upper(s.level)`, []interface{}{userID, bundleID}, out.grammarDone); err != nil {
			return nil, fmt.Errorf("grammar category done: %w", err)
		}
		if err := r.loadPlacementMaxOrder(ctx, userID, bundleID, out); err != nil {
			return nil, err
		}
	}

	if err := r.loadTotalMasteredWords(ctx, userID, courseCode, out); err != nil {
		return nil, err
	}

	targetLang = strings.TrimSpace(strings.ToLower(targetLang))
	if targetLang != "" {
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(rt.level), COUNT(*)
			FROM reading_texts rt
			WHERE LOWER(rt.target_language) = ?
			GROUP BY upper(rt.level)`, []interface{}{targetLang}, out.readingTotal); err != nil {
			return nil, fmt.Errorf("reading totals: %w", err)
		}
		if err := r.scanLevelCounts(ctx, `
			SELECT upper(rt.level), COUNT(DISTINCT rtp.chapter_id)
			FROM reading_texts rt
			JOIN reading_text_progress rtp ON rtp.chapter_id = rt.text_id AND rtp.user_id = ?
			WHERE LOWER(rt.target_language) = ?
			GROUP BY upper(rt.level)`, []interface{}{userID, targetLang}, out.readingDone); err != nil {
			return nil, fmt.Errorf("reading done: %w", err)
		}
	}

	if err := r.scanLevelCounts(ctx, `
		SELECT upper(sc.cefr_level), COUNT(*)
		FROM conversation_scenarios sc
		WHERE sc.course_id = ? AND sc.status = 'active'
		GROUP BY upper(sc.cefr_level)`, []interface{}{courseID}, out.convTotal); err != nil {
		return nil, fmt.Errorf("conversation totals: %w", err)
	}
	if err := r.scanLevelCounts(ctx, `
		SELECT upper(sc.cefr_level), COUNT(DISTINCT sc.id)
		FROM conversation_scenarios sc
		WHERE sc.course_id = ? AND sc.status = 'active'
		AND (
			EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ? AND ea.mode = 'chat'
					AND ea.result_json->>'scenario_code' = sc.code
			)
			OR EXISTS (
				SELECT 1 FROM conversation_sessions sess
				WHERE sess.scenario_id = sc.id AND sess.user_course_id = ? AND sess.status = 'completed'
			)
		)
		GROUP BY upper(sc.cefr_level)`, []interface{}{courseID, userCourseID, userCourseID}, out.convDone); err != nil {
		return nil, fmt.Errorf("conversation done: %w", err)
	}

	if err := r.scanLevelCounts(ctx, `
		SELECT upper(d.level_code), COUNT(*)
		FROM picture_quests q
		JOIN districts d ON d.id = q.district_id
		WHERE q.course_id = ? AND q.status = 'active'
		GROUP BY upper(d.level_code)`, []interface{}{courseID}, out.pictureTotal); err != nil {
		return nil, fmt.Errorf("picture totals: %w", err)
	}
	if err := r.scanLevelCounts(ctx, `
		SELECT upper(d.level_code), COUNT(DISTINCT q.id)
		FROM picture_quests q
		JOIN districts d ON d.id = q.district_id
		WHERE q.course_id = ? AND q.status = 'active'
		AND (
			EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ? AND ea.mode = 'chat'
					AND ea.result_json->>'picture_quest_code' = q.code
			)
			OR EXISTS (
				SELECT 1 FROM picture_quest_sessions sess
				WHERE sess.quest_id = q.id AND sess.user_course_id = ? AND sess.status = 'completed'
			)
		)
		GROUP BY upper(d.level_code)`, []interface{}{courseID, userCourseID, userCourseID}, out.pictureDone); err != nil {
		return nil, fmt.Errorf("picture done: %w", err)
	}

	return out, nil
}

func (r *CourseRepository) loadTotalMasteredWords(ctx context.Context, userID int64, courseCode string, out *masteryRawData) error {
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	if userID == 0 || courseCode == "" {
		return nil
	}
	const q = `
SELECT COUNT(DISTINCT word_card_id) FROM (
    SELECT uwk.word_card_id
    FROM user_word_knowledge uwk
    JOIN word_cards wc ON wc.id = uwk.word_card_id
    WHERE uwk.user_id = ? AND uwk.status = 'known'
      AND LOWER(COALESCE(wc.course_code, '')) = ?
    UNION
    SELECT uwm.word_card_id
    FROM user_word_mastering uwm
    WHERE uwm.user_id = ? AND COALESCE(uwm.mastering_score, 0) >= 80
      AND LOWER(COALESCE(uwm.course_code, '')) = ?
) mastered_words`
	if err := r.db.QueryRowContext(ctx, q, userID, courseCode, userID, courseCode).Scan(&out.wordMasteredTotal); err != nil {
		return fmt.Errorf("total mastered words: %w", err)
	}
	return nil
}

// GetUserTier loads the subscription tier for mastery feature weighting.
func (r *CourseRepository) GetUserTier(ctx context.Context, userID int64) models.UserTier {
	if userID == 0 {
		return models.TierFree
	}
	var tier sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(subscription_tier, 'free') FROM users WHERE id = ?`, userID).Scan(&tier); err != nil {
		return models.TierFree
	}
	return models.ParseUserTier(tier.String)
}

// MasteryLevelByCode returns one level entry from mastery, if present.
func MasteryLevelByCode(m *CourseMastery, levelCode string) *CourseMasteryLevel {
	if m == nil {
		return nil
	}
	code := strings.ToUpper(strings.TrimSpace(levelCode))
	for i := range m.Levels {
		if m.Levels[i].LevelCode == code {
			return &m.Levels[i]
		}
	}
	return nil
}
