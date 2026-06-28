package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const gapUserBase int64 = 900001

func gapRepo(t *testing.T) (*CourseRepository, *sql.DB) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	return NewCourseRepository(conn, zap.NewNop()), conn
}

func gapUser(t *testing.T, conn *sql.DB, telegramID int64) int64 {
	t.Helper()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user %d: %v", telegramID, err)
	}
	return user.ID
}

func gapInsertGrammarItem(t *testing.T, conn *sql.DB, courseCode, moduleCode, sourceID string) int64 {
	t.Helper()
	var itemID int64
	err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = ?
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, ?, 'grammar', 'Gap Grammar', 'grammar_category', ?, 1, 'published'
			FROM target
			ON CONFLICT (course_id, code) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', ?, 'Gap Block', 'A0', 'published'
		FROM module
		ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, courseCode, moduleCode, moduleCode, sourceID).Scan(&itemID)
	if err != nil {
		t.Fatalf("insert grammar item: %v", err)
	}
	return itemID
}

func gapInsertWordItem(t *testing.T, conn *sql.DB, courseCode, word string) (wordCardID, itemID int64) {
	t.Helper()
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES (?, 'gap') RETURNING id`, word).Scan(&wordCardID); err != nil {
		t.Fatalf("word card: %v", err)
	}
	err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = ?
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, ?, 'word_set', 'Gap Words', 'word_set', ?, 1, 'published'
			FROM target ON CONFLICT (course_id, code) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', CAST(? AS text), ?, 'published'
		FROM module ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, courseCode, "word:gap-"+word, "word:gap-"+word, strconv.FormatInt(wordCardID, 10), word).Scan(&itemID)
	if err != nil {
		t.Fatalf("word item: %v", err)
	}
	return wordCardID, itemID
}


func TestCourseRepositoryGap_PureHelpers(t *testing.T) {
	now := time.Now().UTC()
	q0, q1, q2, q3, q4, qNeg := 0, 1, 2, 3, 4, -1
	if maxInt(3, 5) != 5 || maxInt(7, 2) != 7 || maxFloat(1.1, 2.2) != 2.2 || maxFloat(9.9, 1.0) != 9.9 {
		t.Fatal("max helpers")
	}
	stats := map[string]interface{}{"interval_days": float64(4), "step": 2}
	if intFromStats(stats, "interval_days", 0) != 4 || intFromStats(stats, "step", 0) != 2 || intFromStats(stats, "missing", 9) != 9 {
		t.Fatal("intFromStats")
	}
	if sm2Quality(1) != 3 || sm2Quality(2) != 4 || sm2Quality(3) != 5 || sm2Quality(0) != 0 {
		t.Fatal("sm2Quality")
	}
	if normalizeLinglowQuality(false, &q3) != 0 || normalizeLinglowQuality(true, nil) != 2 || normalizeLinglowQuality(true, &q0) != 0 || normalizeLinglowQuality(true, &q1) != 1 || normalizeLinglowQuality(true, &q3) != 3 || normalizeLinglowQuality(true, &q4) != 3 || normalizeLinglowQuality(true, &qNeg) != 0 {
		t.Fatal("normalizeLinglowQuality")
	}
	for itemType, want := range map[string]string{"word": "word_training", "grammar_chapter": "grammar", "grammar_theory_block": "grammar", "reading_text": "reading", "speaking_task": "speaking", "custom_type": "custom_type"} {
		if dailyRouteMode(itemType) != want {
			t.Fatalf("dailyRouteMode(%q)", itemType)
		}
	}
	failReview := applyLinglowSRS(linglowSRSState{State: "review", EF: 2.5, Reps: 2, IntervalDays: 6, Stats: map[string]interface{}{}}, false, nil, now)
	if failReview.State != "relearning" {
		t.Fatalf("fail review = %+v", failReview)
	}
	failLearning := applyLinglowSRS(linglowSRSState{State: "learning", EF: 2.5, Stats: map[string]interface{}{}}, false, &q0, now)
	if failLearning.State != "learning" {
		t.Fatalf("fail learning = %+v", failLearning)
	}
	failMastered := applyLinglowSRS(linglowSRSState{State: "mastered", EF: 2.5, IntervalDays: 10, Stats: map[string]interface{}{}}, false, nil, now)
	if failMastered.State != "relearning" {
		t.Fatalf("fail mastered = %+v", failMastered)
	}
	hard := applyLinglowSRS(linglowSRSState{State: "new", EF: 0, Stats: nil}, true, &q1, now)
	if hard.State != "learning" || hard.EF != 2.5 {
		t.Fatalf("new hard = %+v", hard)
	}
	step1 := applyLinglowSRS(linglowSRSState{State: "learning", EF: 2.5, LearningStep: 0, Stats: map[string]interface{}{}}, true, &q2, now)
	if step1.LearningStep != 1 || step1.State != "learning" {
		t.Fatalf("step1 = %+v", step1)
	}
	graduate := applyLinglowSRS(linglowSRSState{State: "learning", EF: 2.5, LearningStep: 1, Stats: map[string]interface{}{}}, true, &q2, now)
	if graduate.State != "review" {
		t.Fatalf("graduate = %+v", graduate)
	}
	rep0 := applyLinglowSRS(linglowSRSState{State: "review", EF: 2.5, Reps: 0, Stats: map[string]interface{}{}}, true, &q2, now)
	if rep0.IntervalDays != 1 {
		t.Fatalf("rep0 = %+v", rep0)
	}
	rep1 := applyLinglowSRS(linglowSRSState{State: "review", EF: 2.5, Reps: 1, IntervalDays: 3, Stats: map[string]interface{}{"interval_days": 3}}, true, &q2, now)
	if rep1.IntervalDays != 6 {
		t.Fatalf("rep1 = %+v", rep1)
	}
	repMany := applyLinglowSRS(linglowSRSState{State: "review", EF: 2.5, Reps: 3, IntervalDays: 6, Stats: map[string]interface{}{"interval_days": 6}}, true, &q2, now)
	if repMany.IntervalDays < 6 {
		t.Fatalf("repMany = %+v", repMany)
	}
	weird := applyLinglowSRS(linglowSRSState{State: "paused", EF: 2.5, Stats: map[string]interface{}{}}, true, &q2, now)
	if weird.State != "learning" {
		t.Fatalf("default = %+v", weird)
	}
}


func TestCourseRepositoryGap_ResolveSelectEnsure(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	if _, err := repo.SelectCurrentCourse(ctx, 0, "es_ru"); err == nil {
		t.Fatal("user 0")
	}
	if _, err := repo.SelectCurrentCourse(ctx, gapUserBase+1, "missing-gap-course"); err == nil {
		t.Fatal("invalid course")
	}
	userID := gapUser(t, conn, gapUserBase+2)
	if _, err := repo.SelectCurrentCourse(ctx, userID, "en_ru"); err != nil {
		t.Fatalf("select en_ru: %v", err)
	}
	onlyEN := gapUser(t, conn, gapUserBase+10)
	if _, err := repo.SelectCurrentCourse(ctx, onlyEN, "en_ru"); err != nil {
		t.Fatalf("only en: %v", err)
	}
	code, err := repo.ResolveCurrentCourseCode(ctx, onlyEN, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("enrolled only = %q", code)
	}
	newUser := gapUser(t, conn, gapUserBase+11)
	code, err = repo.ResolveCurrentCourseCode(ctx, newUser, "es_ru")
	if err != nil || code != "es_ru" {
		t.Fatalf("new user = %q", code)
	}
	if _, err := conn.Exec(`UPDATE users SET settings_json = '{bad' WHERE id = ?`, userID); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	code, err = repo.ResolveCurrentCourseCode(ctx, userID, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("bad json fallback = %q", code)
	}
	if _, err := conn.Exec(`UPDATE users SET settings_json = '{"current_course_code":"ghost-course"}' WHERE id = ?`, userID); err != nil {
		t.Fatalf("ghost settings: %v", err)
	}
	code, err = repo.ResolveCurrentCourseCode(ctx, userID, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("ghost fallback = %q", code)
	}
	if !repo.courseIsActive(ctx, "es_ru") || repo.courseIsActive(ctx, "ghost-course") {
		t.Fatal("courseIsActive")
	}
	current, err := repo.GetCurrentCourse(ctx, userID, "es_ru")
	if err != nil || current.Course.Code != "en_ru" {
		t.Fatalf("GetCurrentCourse = %+v", current)
	}
	if _, err := repo.ListCoursesForUser(ctx, userID, "es_ru"); err != nil {
		t.Fatalf("ListCoursesForUser: %v", err)
	}
}

func TestCourseRepositoryGap_ReviewQueueAndDailyRoute(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	if _, err := repo.GetReviewQueueForUserWithSRSRead(ctx, 0, "es_ru", "", 10, true); err == nil {
		t.Fatal("review user 0")
	}
	if _, err := repo.GetDailyRouteForUserWithSRSRead(ctx, 0, "es_ru", "", 10, true); err == nil {
		t.Fatal("route user 0")
	}
	userID := gapUser(t, conn, gapUserBase+20)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	relearnItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-relearn", "gap.relearn")
	learnItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-learn", "gap.learn")
	upcomingItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-upcoming", "gap.upcoming")
	for _, row := range []struct {
		itemID int64
		state  string
		due    string
	}{
		{relearnItem, "relearning", "CURRENT_TIMESTAMP - INTERVAL '1 hour'"},
		{learnItem, "learning", "CURRENT_TIMESTAMP - INTERVAL '1 hour'"},
		{upcomingItem, "learning", "CURRENT_TIMESTAMP + INTERVAL '2 hours'"},
	} {
		if _, err := conn.Exec(fmt.Sprintf(`INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps) VALUES (?, ?, ?, %s, 1)`, row.due), current.UserCourse.ID, row.itemID, row.state); err != nil {
			t.Fatalf("srs: %v", err)
		}
	}
	queue, err := repo.GetReviewQueueForUserWithSRSRead(ctx, userID, "es_ru", "", 0, true)
	if err != nil || queue.Summary.DueCount < 2 || queue.Summary.RelearningCount < 1 || queue.Summary.UpcomingCount < 1 {
		t.Fatalf("canonical queue = %+v %v", queue.Summary, err)
	}
	bigQueue, err := repo.GetReviewQueueForUserWithSRSRead(ctx, userID, "es_ru", "", 500, true)
	if err != nil || len(bigQueue.Items) > 100 {
		t.Fatalf("limit cap = %d", len(bigQueue.Items))
	}
	route, err := repo.GetDailyRouteForUserWithSRSRead(ctx, userID, "es_ru", "", 500, true)
	if err != nil || len(route.Review) > 50 {
		t.Fatalf("route cap = %d", len(route.Review))
	}
	wcID, wordItemID := gapInsertWordItem(t, conn, "es_ru", fmt.Sprintf("gap-known-%d", gapUserBase))
	var tcID int64
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'k', 0, 'k', 'k') RETURNING id`, wcID).Scan(&tcID); err != nil {
		t.Fatalf("tc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at) VALUES (?, ?, 'es_ru', 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour')`, userID, tcID); err != nil {
		t.Fatalf("uc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, userID, wcID); err != nil {
		t.Fatalf("known: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at) VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour')`, current.UserCourse.ID, wordItemID); err != nil {
		t.Fatalf("known srs: %v", err)
	}
	knownQueue, err := repo.GetReviewQueueForUserWithSRSRead(ctx, userID, "es_ru", "", 50, true)
	if err != nil {
		t.Fatalf("known queue: %v", err)
	}
	for _, item := range knownQueue.Items {
		if item.LearningItemID == wordItemID {
			t.Fatal("known word in queue")
		}
	}
	legacyUser := gapUser(t, conn, gapUserBase+21)
	legacyCurrent, err := repo.SelectCurrentCourse(ctx, legacyUser, "es_ru")
	if err != nil {
		t.Fatalf("legacy select: %v", err)
	}
	lwc, _ := gapInsertWordItem(t, conn, "es_ru", "gap-legacy-relearn")
	var ltc int64
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'lr', 0, 'lr', 'lr') RETURNING id`, lwc).Scan(&ltc); err != nil {
		t.Fatalf("ltc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at) VALUES (?, ?, 'es_ru', 'relearning', CURRENT_TIMESTAMP - INTERVAL '1 hour')`, legacyUser, ltc); err != nil {
		t.Fatalf("legacy card: %v", err)
	}
	_ = gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-legacy-up", "gap.legacy.up:b1")
	if _, err := conn.Exec(`INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at) VALUES (?, 'es', 'es', 'gap.legacy.up', 'b1', 'c1', 'review', CURRENT_TIMESTAMP + INTERVAL '3 hours')`, legacyUser); err != nil {
		t.Fatalf("grammar upcoming: %v", err)
	}
	legacyQueue, err := repo.GetReviewQueueForUserWithSRSRead(ctx, legacyUser, "es_ru", "", 20, false)
	if err != nil || legacyQueue.Summary.ReadSource != "legacy" || legacyQueue.Summary.RelearningCount < 1 || legacyQueue.Summary.UpcomingCount < 1 {
		t.Fatalf("legacy queue = %+v %v", legacyQueue.Summary, err)
	}
	legacyRoute, err := repo.GetDailyRouteForUserWithSRSRead(ctx, legacyUser, "es_ru", "", 10, false)
	if err != nil || legacyRoute.UserCourse.ID != legacyCurrent.UserCourse.ID {
		t.Fatalf("legacy route = %+v", legacyRoute.UserCourse)
	}
}

func TestCourseRepositoryGap_RecordExerciseAttemptSRS(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	userID := gapUser(t, conn, gapUserBase+30)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	itemID := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-at", "gap.at.block")
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID}); err == nil {
		t.Fatal("empty mode")
	}
	if _, err := repo.GetProgressForUser(ctx, 0, "es_ru", ""); err == nil {
		t.Fatal("progress user 0")
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID, Mode: "grammar", ClientAttemptID: "gap-nil", UpdateSRS: true}); err != nil {
		t.Fatalf("nil isCorrect: %v", err)
	}
	masteredItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-mastered", "gap.mastered")
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, difficulty, reps) VALUES (?, ?, 'mastered', 2.5, 5)`, current.UserCourse.ID, masteredItem); err != nil {
		t.Fatalf("mastered srs: %v", err)
	}
	wrong, correct := false, true
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: masteredItem, Mode: "grammar", ClientAttemptID: "gap-mastered-fail", IsCorrect: &wrong, UpdateSRS: true}); err != nil {
		t.Fatalf("mastered fail: %v", err)
	}
	var srsID int64
	if err := conn.QueryRow(`INSERT INTO srs_items (user_course_id, learning_item_id, state) VALUES (?, ?, 'review') RETURNING id`, current.UserCourse.ID, itemID).Scan(&srsID); err != nil {
		t.Fatalf("srs: %v", err)
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID, SRSItemID: srsID + 9999, Mode: "grammar", ClientAttemptID: "gap-missing-srs", IsCorrect: &correct, UpdateSRS: true}); err == nil {
		t.Fatal("missing srs")
	}
	otherItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-other", "gap.other")
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: otherItem, SRSItemID: srsID, Mode: "grammar", ClientAttemptID: "gap-mismatch", IsCorrect: &correct, UpdateSRS: true}); err == nil {
		t.Fatal("srs mismatch")
	}
}

func TestCourseRepositoryGap_SRSShadowAndAggregate(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	if _, err := repo.GetSRSReadinessAggregate(ctx, "", 10); err != ErrCourseNotFound {
		t.Fatalf("aggregate empty = %v", err)
	}
	if _, err := repo.GetSRSShadowReportForUser(ctx, 0, "es_ru", ""); err == nil {
		t.Fatal("shadow user 0")
	}
	legacyOnlyUser := gapUser(t, conn, gapUserBase+50)
	if _, err := repo.SelectCurrentCourse(ctx, legacyOnlyUser, "es_ru"); err != nil {
		t.Fatalf("select: %v", err)
	}
	wc, _ := gapInsertWordItem(t, conn, "es_ru", "gap-shadow-legacy")
	var tcID int64
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 's', 0, 's', 's') RETURNING id`, wc).Scan(&tcID); err != nil {
		t.Fatalf("tc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at) VALUES (?, ?, 'es_ru', 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour')`, legacyOnlyUser, tcID); err != nil {
		t.Fatalf("uc: %v", err)
	}
	legacyReport, err := repo.GetSRSShadowReportForUser(ctx, legacyOnlyUser, "es_ru", "")
	if err != nil || legacyReport.Due.LegacyOnlyCount < 1 {
		t.Fatalf("legacy-only = %+v", legacyReport.Due)
	}
	canonUser := gapUser(t, conn, gapUserBase+51)
	canonCurrent, err := repo.SelectCurrentCourse(ctx, canonUser, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	_, canonItem := gapInsertWordItem(t, conn, "es_ru", "gap-shadow-canon")
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps) VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', 1)`, canonCurrent.UserCourse.ID, canonItem); err != nil {
		t.Fatalf("srs: %v", err)
	}
	canonReport, err := repo.GetSRSShadowReportForUser(ctx, canonUser, "es_ru", "")
	if err != nil || canonReport.Due.LinglowOnlyCount < 1 {
		t.Fatalf("canonical-only = %+v", canonReport.Due)
	}
	overlapUser := gapUser(t, conn, gapUserBase+52)
	overlapCurrent, err := repo.SelectCurrentCourse(ctx, overlapUser, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	owc, oli := gapInsertWordItem(t, conn, "es_ru", "gap-shadow-overlap")
	var otc int64
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'o', 0, 'o', 'o') RETURNING id`, owc).Scan(&otc); err != nil {
		t.Fatalf("otc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at, reps) VALUES (?, ?, 'es_ru', 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', 2)`, overlapUser, otc); err != nil {
		t.Fatalf("overlap legacy: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps) VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', 2)`, overlapCurrent.UserCourse.ID, oli); err != nil {
		t.Fatalf("overlap srs: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES (?, ?, 80)`, overlapUser, owc); err != nil {
		t.Fatalf("mastering: %v", err)
	}
	overlapReport, err := repo.GetSRSShadowReportForUser(ctx, overlapUser, "es_ru", "")
	if err != nil || overlapReport.Mastery.ComparedCount < 1 {
		t.Fatalf("overlap mastery = %+v", overlapReport.Mastery)
	}
	agg, err := repo.GetSRSReadinessAggregate(ctx, "es_ru", 20)
	if err != nil || agg.Course.Code != "es_ru" {
		t.Fatalf("aggregate = %+v %v", agg, err)
	}
}

func TestCourseRepositoryGap_CourseMapAndLegacyMap(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	userID := gapUser(t, conn, gapUserBase+40)
	if _, err := repo.GetCourseMap(ctx, "missing-gap-course", userID); err == nil {
		t.Fatal("missing course")
	}
	if _, err := conn.Exec(`UPDATE districts SET description = 'Gap district', metadata_json = '{"gap":true}'::jsonb WHERE course_id = (SELECT id FROM courses WHERE code = 'es_ru') AND level_code = 'A0'`); err != nil {
		t.Fatalf("district: %v", err)
	}
	courseMap, err := repo.GetCourseMap(ctx, "es_ru", userID)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	for _, d := range courseMap.Districts {
		if d.LevelCode == "A0" && d.Description == "Gap district" && len(d.Metadata) == 0 {
			t.Fatalf("metadata missing: %+v", d)
		}
	}
	if _, err := conn.Exec(`INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash) VALUES ('es', 'es.section.gap', 'Gap', 'B1', 1, '["es.chapter.gap"]', '{}', 'sec-gap')`); err != nil {
		t.Fatalf("section: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, ui_language, target_language, level, sort_order, raw_json, source_hash) VALUES ('es', 'es.chapter.gap', 'es.section.gap', 'Gap Chapter', 'ru', 'es', 'B1', 1, '{}', 'ch-gap')`); err != nil {
		t.Fatalf("chapter: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids) VALUES ('gap.read.cat', 'Gap Read', 'A0', 1, '["gap.read.text"]') ON CONFLICT (category_id) DO NOTHING`); err != nil {
		t.Fatalf("read cat: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage) VALUES ('gap.read.text', 'gap.read.cat', 'Gap Text', 'A0', 'es', 'Hola gap') ON CONFLICT (text_id) DO NOTHING`); err != nil {
		t.Fatalf("read text: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO speaking_categories (category_id, title, level, sort_order, task_ids) VALUES ('gap.speak.cat', 'Gap Speak', 'A0', 1, '["gap.speak.task"]') ON CONFLICT (category_id) DO NOTHING`); err != nil {
		t.Fatalf("speak cat: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO speaking_tasks (task_id, category_id, title, level, task_type, target_language, task_json) VALUES ('gap.speak.task', 'gap.speak.cat', 'Gap Task', 'A0', 'answer', 'es', '{}') ON CONFLICT (task_id) DO NOTHING`); err != nil {
		t.Fatalf("speak task: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO word_sets (title, description, is_published, sort_order, course_code) VALUES ('Gap Set', 'Gap', 1, 99, 'es_ru')`); err != nil {
		t.Fatalf("word set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO word_cards (word, definition) VALUES ('gapmapword', 'gap')`); err != nil {
		t.Fatalf("word card: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order) SELECT ws.id, wc.id, 1 FROM word_sets ws CROSS JOIN word_cards wc WHERE ws.title = 'Gap Set' AND wc.word = 'gapmapword'`); err != nil {
		t.Fatalf("word set item: %v", err)
	}
	summary, err := repo.MapLegacyContent(ctx, "es_ru", "es")
	if err != nil || summary.ModulesCreated < 1 {
		t.Fatalf("MapLegacyContent = %+v err=%v", summary, err)
	}
	if _, err := repo.MapLegacyContentForLearning(ctx, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
}

func TestCourseRepositoryGap_ProgressRoutesAndAttempts(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()
	courseMap, err := repo.GetCourseMap(ctx, "es_ru", 0)
	if err != nil || courseMap.UserCourse != nil {
		t.Fatalf("GetCourseMap user 0 = %+v", courseMap.UserCourse)
	}
	fallbackUser := gapUser(t, conn, gapUserBase+60)
	if _, err := repo.SelectCurrentCourse(ctx, fallbackUser, "en_ru"); err != nil {
		t.Fatalf("select en: %v", err)
	}
	code, err := repo.ResolveCurrentCourseCode(ctx, fallbackUser, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("fallback = %q", code)
	}
	userID := gapUser(t, conn, gapUserBase+61)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	dueItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-prog-due", "gap.prog.due")
	masteredItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-prog-mastered", "gap.prog.mastered")
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps) VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', 2)`, current.UserCourse.ID, dueItem); err != nil {
		t.Fatalf("due srs: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, reps) VALUES (?, ?, 'mastered', 4)`, current.UserCourse.ID, masteredItem); err != nil {
		t.Fatalf("mastered srs: %v", err)
	}
	ok := true
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", ExplicitCourse: "es_ru", LearningItemID: masteredItem, Mode: "grammar", ClientAttemptID: "gap-prog", IsCorrect: &ok}); err != nil {
		t.Fatalf("attempt: %v", err)
	}
	progress, err := repo.GetProgressForUser(ctx, userID, "en_ru", "es_ru")
	if err != nil || progress.Summary.MasteredItems < 1 || progress.Summary.DueReviewCount < 1 {
		t.Fatalf("progress = %+v", progress.Summary)
	}
	route, err := repo.GetDailyRouteForUserWithSRSRead(ctx, userID, "es_ru", "", 10, true)
	if err != nil || len(route.Review) == 0 {
		t.Fatalf("route = %+v", route)
	}
	efZeroItem := gapInsertGrammarItem(t, conn, "es_ru", "grammar:gap-efzero", "gap.efzero")
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state, difficulty, reps, stats_json) VALUES (?, ?, 'review', 0, 1, '{}'::jsonb)`, current.UserCourse.ID, efZeroItem); err != nil {
		t.Fatalf("ef zero: %v", err)
	}
	q2 := 2
	res, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: userID, DefaultCourse: "es_ru", LearningItemID: efZeroItem, Mode: "grammar", ClientAttemptID: fmt.Sprintf("gap-ef-%d", gapUserBase), IsCorrect: &ok, Quality: &q2, UpdateSRS: true})
	if err != nil || !res.SRSUpdated {
		t.Fatalf("ef zero attempt = %+v", res)
	}
	var ef float64
	if err := conn.QueryRow(`SELECT difficulty FROM srs_items WHERE user_course_id = ? AND learning_item_id = ?`, current.UserCourse.ID, efZeroItem).Scan(&ef); err != nil || ef <= 0 {
		t.Fatalf("ef = %v", ef)
	}
}

func TestCourseRepositoryGap_ExtraBranches(t *testing.T) {
	repo, conn := gapRepo(t)
	ctx := context.Background()

	if _, err := repo.SelectCurrentCourse(ctx, 0, "es_ru"); err == nil {
		t.Fatal("select user 0")
	}
	if _, err := repo.SelectCurrentCourse(ctx, gapUserBase+70, ""); err == nil {
		t.Fatal("select empty code")
	}
	if _, err := repo.EnsureUserCourse(ctx, gapUserBase+70, "ghost-gap"); err == nil {
		t.Fatal("ensure ghost")
	}

	userID := gapUser(t, conn, gapUserBase+70)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	route := &DailyRoute{
		Course:     CourseMapCourse{TargetLanguage: "es"},
		UserCourse: current.UserCourse,
		Summary:    DailyRouteSummary{DueReviewCount: 2},
	}
	repo.FillDailyRouteToday(ctx, route, userID)
	if route.Today == nil {
		t.Fatal("FillDailyRouteToday")
	}

	legacyUser := gapUser(t, conn, gapUserBase+71)
	if _, err := repo.SelectCurrentCourse(ctx, legacyUser, "es_ru"); err != nil {
		t.Fatalf("legacy select: %v", err)
	}
	wc, _ := gapInsertWordItem(t, conn, "es_ru", "gap-legacy-default-state")
	var tc int64
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'gap-legacy-default-state', 0, 'd', 'd') RETURNING id`, wc).Scan(&tc); err != nil {
		t.Fatalf("tc: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at) VALUES (?, ?, 'es_ru', '', CURRENT_TIMESTAMP - INTERVAL '1 hour')`, legacyUser, tc); err != nil {
		t.Fatalf("user card: %v", err)
	}
	q, err := repo.GetReviewQueueForUserWithSRSRead(ctx, legacyUser, "es_ru", "", 10, false)
	if err != nil || q.Summary.ReviewCount < 1 {
		t.Fatalf("legacy default review = %+v %v", q.Summary, err)
	}

	if _, err := repo.GetCourseMapForUser(ctx, userID, "es_ru", "es_ru"); err != nil {
		t.Fatalf("GetCourseMapForUser explicit: %v", err)
	}
	if _, err := repo.ResolveRequestedCourseCode(ctx, userID, "es_ru", "es_ru"); err != nil {
		t.Fatalf("ResolveRequestedCourseCode: %v", err)
	}
}
