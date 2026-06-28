package repository

import (
	"context"
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// ============================================================
// conversation_repository.go — integration coverage
// ============================================================

func TestConversationRepository_ScenarioAndSessionQueries(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	var courseID, districtID int64
	if err := conn.QueryRow(`SELECT course_id, district_id FROM conversation_scenarios WHERE id = ?`, fx.scenarioID).Scan(&courseID, &districtID); err != nil {
		t.Fatalf("read scenario meta: %v", err)
	}

	scenarios, err := repo.ListScenariosForDistrict(ctx, courseID, districtID)
	if err != nil {
		t.Fatalf("ListScenariosForDistrict: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatal("expected at least one scenario in district")
	}

	byID, err := repo.GetScenarioByID(ctx, fx.scenarioID)
	if err != nil || byID == nil || byID.Code != "test_cafe" {
		t.Fatalf("GetScenarioByID: %+v err=%v", byID, err)
	}
	missing, err := repo.GetScenarioByID(ctx, 999999999)
	if err != nil || missing != nil {
		t.Fatalf("GetScenarioByID missing: %+v err=%v", missing, err)
	}

	byCode, err := repo.GetScenarioByCode(ctx, courseID, "test_cafe")
	if err != nil || byCode == nil {
		t.Fatalf("GetScenarioByCode: %+v err=%v", byCode, err)
	}
	noCode, err := repo.GetScenarioByCode(ctx, courseID, "no_such_scenario")
	if err != nil || noCode != nil {
		t.Fatalf("GetScenarioByCode missing: %+v err=%v", noCode, err)
	}

	user900001, err := NewUserRepository(conn, zap.NewNop()).GetOrCreateUser(900001)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	ownerID, err := userIDForUserCourse(conn, fx.userCourseID)
	if err != nil {
		t.Fatalf("owner user: %v", err)
	}
	var otherUC int64
	if err := conn.QueryRow(`INSERT INTO user_courses (user_id, course_id, status) VALUES (?, ?, 'active') RETURNING id`, user900001.ID, courseID).Scan(&otherUC); err != nil {
		t.Fatalf("other user_course: %v", err)
	}

	session, created, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil || !created {
		t.Fatalf("StartSession: created=%v err=%v", created, err)
	}

	open, err := repo.GetOpenSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil || open == nil || open.ID != session.ID {
		t.Fatalf("GetOpenSession: %+v err=%v", open, err)
	}
	got, err := repo.GetSession(ctx, session.ID, fx.userCourseID)
	if err != nil || got == nil || got.ID != session.ID {
		t.Fatalf("GetSession: %+v err=%v", got, err)
	}
	wrongUC, err := repo.GetSession(ctx, session.ID, otherUC)
	if err != nil || wrongUC != nil {
		t.Fatalf("GetSession wrong user_course: %+v err=%v", wrongUC, err)
	}

	forUser, err := repo.GetSessionForUser(ctx, session.ID, ownerID)
	if err != nil || forUser == nil {
		t.Fatalf("GetSessionForUser: %+v err=%v", forUser, err)
	}
	noUser, err := repo.GetSessionForUser(ctx, session.ID, user900001.ID)
	if err != nil || noUser != nil {
		t.Fatalf("GetSessionForUser missing: %+v err=%v", noUser, err)
	}
}

func TestConversationRepository_BumpAndCloseSession(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	session, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if err := repo.BumpSessionCounters(ctx, session.ID, 2, 100); err != nil {
		t.Fatalf("BumpSessionCounters: %v", err)
	}
	got, err := repo.GetSession(ctx, session.ID, fx.userCourseID)
	if err != nil || got.TurnCount != 2 || got.TokensUsed != 100 {
		t.Fatalf("counters not bumped: %+v err=%v", got, err)
	}

	if err := repo.CloseSession(ctx, session.ID, fx.userCourseID, "abandoned"); err != nil {
		t.Fatalf("CloseSession abandoned: %v", err)
	}
	closed, _ := repo.GetSession(ctx, session.ID, fx.userCourseID)
	if closed == nil || closed.Status != "abandoned" {
		t.Fatalf("expected abandoned session, got %+v", closed)
	}
}

func TestConversationRepository_ProgressAndNPC(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	var courseID int64
	if err := conn.QueryRow(`SELECT course_id FROM conversation_scenarios WHERE id = ?`, fx.scenarioID).Scan(&courseID); err != nil {
		t.Fatalf("course id: %v", err)
	}

	session, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	if err := repo.RecordQuestCompletion(ctx, fx.userCourseID, sql.NullInt64{}, "test_cafe", session.ID); err != nil {
		t.Fatalf("RecordQuestCompletion null learning item: %v", err)
	}

	passed, err := repo.LatestPassedScenarioCodes(ctx, fx.userCourseID, courseID)
	if err != nil || !passed["test_cafe"] {
		t.Fatalf("LatestPassedScenarioCodes after quest: %v err=%v", passed, err)
	}
	passedAt, err := repo.PassedAtByScenarioCode(ctx, fx.userCourseID, courseID)
	if err != nil || passedAt["test_cafe"].IsZero() {
		t.Fatalf("PassedAtByScenarioCode: %v err=%v", passedAt, err)
	}
	ever, err := repo.ScenarioEverPassed(ctx, fx.userCourseID, "test_cafe")
	if err != nil || !ever {
		t.Fatalf("ScenarioEverPassed: ever=%v err=%v", ever, err)
	}

	s2, _, _ := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err := repo.CloseSession(ctx, s2.ID, fx.userCourseID, "completed"); err != nil {
		t.Fatalf("CloseSession completed: %v", err)
	}
	completedAt, err := repo.CompletedAtByScenarioCode(ctx, fx.userCourseID, courseID)
	if err != nil || completedAt["test_cafe"].IsZero() {
		t.Fatalf("CompletedAtByScenarioCode: %v err=%v", completedAt, err)
	}

	if err := repo.UpsertNPCImage(ctx, courseID, "mara", "https://example.com/mara.png"); err != nil {
		t.Fatalf("UpsertNPCImage set: %v", err)
	}
	images, err := repo.GetNPCImages(ctx, courseID)
	if err != nil || images["mara"] != "https://example.com/mara.png" {
		t.Fatalf("GetNPCImages: %v err=%v", images, err)
	}
	if err := repo.UpsertNPCImage(ctx, courseID, "mara", ""); err != nil {
		t.Fatalf("UpsertNPCImage clear: %v", err)
	}
}

func TestConversationRepository_AppendMessageWhitespaceCorrections(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewConversationRepository(conn, zap.NewNop())
	fx := setupConversationFixture(t, conn)

	s, _, err := repo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if err := repo.AppendMessageWithCorrections(ctx, s.ID, 1, "assistant", "Hola", 5, 3, "   "); err != nil {
		t.Fatalf("AppendMessageWithCorrections: %v", err)
	}
	msgs, err := repo.ListMessages(ctx, s.ID)
	if err != nil || len(msgs) != 1 || msgs[0].CorrectionsJSON != "[]" {
		t.Fatalf("whitespace corrections should become []: %+v err=%v", msgs, err)
	}
}

// ============================================================
// conversation_repository.go — DB error paths
// ============================================================

func TestConversationRepository_DBErrors(t *testing.T) {
	ctx := context.Background()
	db := newBrokenDB(t)
	repo := NewConversationRepository(db, zap.NewNop())

	cases := []struct {
		name string
		fn   func() error
	}{
		{"ListScenariosForDistrict", func() error { _, err := repo.ListScenariosForDistrict(ctx, 1, 1); return err }},
		{"GetScenarioByID", func() error { _, err := repo.GetScenarioByID(ctx, 1); return err }},
		{"GetScenarioByCode", func() error { _, err := repo.GetScenarioByCode(ctx, 1, "x"); return err }},
		{"ListTasks", func() error { _, err := repo.ListTasks(ctx, 1); return err }},
		{"GetOpenSession", func() error { _, err := repo.GetOpenSession(ctx, 1, 1); return err }},
		{"GetSession", func() error { _, err := repo.GetSession(ctx, 1, 1); return err }},
		{"GetSessionForUser", func() error { _, err := repo.GetSessionForUser(ctx, 1, 1); return err }},
		{"StartSession", func() error { _, _, err := repo.StartSession(ctx, 1, 1); return err }},
		{"NextSeq", func() error { _, err := repo.NextSeq(ctx, 1); return err }},
		{"AppendMessage", func() error { return repo.AppendMessage(ctx, 1, 1, "user", "hi", 0, 0) }},
		{"AppendMessageWithCorrections", func() error {
			return repo.AppendMessageWithCorrections(ctx, 1, 1, "user", "hi", 0, 0, "[]")
		}},
		{"ListMessages", func() error { _, err := repo.ListMessages(ctx, 1); return err }},
		{"BumpSessionCounters", func() error { return repo.BumpSessionCounters(ctx, 1, 1, 1) }},
		{"MarkTasksCompleted", func() error {
			return repo.MarkTasksCompleted(ctx, 1, map[string]int64{"x": 1}, []string{"x"}, 1)
		}},
		{"GetCompletedTaskIDs", func() error { _, err := repo.GetCompletedTaskIDs(ctx, 1); return err }},
		{"RecordQuestCompletion", func() error {
			return repo.RecordQuestCompletion(ctx, 1, sql.NullInt64{Int64: 1, Valid: true}, "code", 1)
		}},
		{"LatestPassedScenarioCodes", func() error { _, err := repo.LatestPassedScenarioCodes(ctx, 1, 1); return err }},
		{"PassedAtByScenarioCode", func() error { _, err := repo.PassedAtByScenarioCode(ctx, 1, 1); return err }},
		{"ScenarioEverPassed", func() error { _, err := repo.ScenarioEverPassed(ctx, 1, "x"); return err }},
		{"LatestCompletedScenarioCodes", func() error { _, err := repo.LatestCompletedScenarioCodes(ctx, 1, 1); return err }},
		{"CloseSession", func() error { return repo.CloseSession(ctx, 1, 1, "completed") }},
		{"GetNPCImages", func() error { _, err := repo.GetNPCImages(ctx, 1); return err }},
		{"CompletedAtByScenarioCode", func() error { _, err := repo.CompletedAtByScenarioCode(ctx, 1, 1); return err }},
		{"UpsertNPCImage", func() error { return repo.UpsertNPCImage(ctx, 1, "npc", "url") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err == nil {
				t.Fatalf("%s: expected error on broken DB", tc.name)
			}
		})
	}
}

// ============================================================
// content_report_repository.go — integration coverage
// ============================================================

func setupContentReportUser(t *testing.T, telegramID int64) (*models.User, *ContentReportRepository) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	user, err := NewUserRepository(conn, logger).GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	return user, NewContentReportRepository(conn, logger)
}

func userIDForUserCourse(conn *sql.DB, userCourseID int64) (int64, error) {
	var userID int64
	err := conn.QueryRow(`SELECT user_id FROM user_courses WHERE id = ?`, userCourseID).Scan(&userID)
	return userID, err
}

func TestContentReportRepository_CreateAndGetByID(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewContentReportRepository(conn, logger)
	user, err := NewUserRepository(conn, logger).GetOrCreateUser(900010)
	if err != nil {
		t.Fatalf("user: %v", err)
	}

	var wordCardID, trainingCardID, userCardID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en) VALUES ('gato', 'cat', 'gato') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("word card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos) VALUES (?, 'gato', 0, 'кот', 'cat', 'noun') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("training card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'new', 2.5) RETURNING id`, user.ID, trainingCardID).Scan(&userCardID); err != nil {
		t.Fatalf("user card: %v", err)
	}
	resolver, err := NewUserRepository(conn, logger).GetOrCreateUser(900011)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}

	id, err := repo.Create(CreateContentReportInput{
		UserID:               user.ID,
		SourceType:           "word_training",
		ClientReportID:       "  client-900010  ",
		Word:                 "hola",
		TranslationDirection: "es_ru",
		WordCardID:           &wordCardID,
		TrainingCardID:       &trainingCardID,
		UserCardID:           &userCardID,
		WordCategory:         "noun",
		ReportCategory:       "bad_audio",
		CommentText:          "  sounds wrong  ",
		Payload:              map[string]interface{}{"hint": "test"},
	})
	if err != nil {
		t.Fatalf("Create full: %v", err)
	}

	got, err := repo.GetByID(id)
	if err != nil || got == nil {
		t.Fatalf("GetByID: %+v err=%v", got, err)
	}
	if got.WordCardID == nil || *got.WordCardID != wordCardID {
		t.Fatalf("WordCardID not scanned: %+v", got)
	}
	if got.TrainingCardID == nil || got.UserCardID == nil {
		t.Fatalf("TrainingCardID/UserCardID not scanned: %+v", got)
	}
	if got.CommentText != "sounds wrong" {
		t.Fatalf("CommentText trimmed: %q", got.CommentText)
	}

	if err := repo.Resolve(id, resolver.ID); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	resolved, err := repo.GetByID(id)
	if err != nil || resolved == nil || resolved.Status != "resolved" {
		t.Fatalf("GetByID resolved: %+v err=%v", resolved, err)
	}
	if resolved.ResolvedAt == nil || resolved.ResolvedByUserID == nil || *resolved.ResolvedByUserID != resolver.ID {
		t.Fatalf("resolved fields: %+v", resolved)
	}

	missing, err := repo.GetByID(999999999)
	if err != nil || missing != nil {
		t.Fatalf("GetByID missing: %+v err=%v", missing, err)
	}
}

func TestContentReportRepository_CreateMarshalError(t *testing.T) {
	_, repo := setupContentReportUser(t, 900012)
	_, err := repo.Create(CreateContentReportInput{
		UserID:     1,
		SourceType: "word_training",
		Payload:    map[string]interface{}{"bad": make(chan int)},
	})
	if err == nil {
		t.Fatal("expected marshal payload error")
	}
}

func TestContentReportRepository_HasClientReportEdgeCases(t *testing.T) {
	user, repo := setupContentReportUser(t, 900013)

	for _, tc := range []struct {
		uid    int64
		client string
	}{
		{0, "x"},
		{user.ID, ""},
		{user.ID, "   "},
	} {
		exists, err := repo.HasClientReport(tc.uid, tc.client)
		if err != nil || exists {
			t.Fatalf("HasClientReport(%d, %q) = %v err=%v", tc.uid, tc.client, exists, err)
		}
	}
}

func TestContentReportRepository_ResolveNotFound(t *testing.T) {
	_, repo := setupContentReportUser(t, 900014)
	if err := repo.Resolve(999999999, 1); err == nil {
		t.Fatal("expected resolve not found error")
	}
}

func TestContentReportRepository_ListAndFilters(t *testing.T) {
	user, repo := setupContentReportUser(t, 900015)

	idEN, err := repo.Create(CreateContentReportInput{
		UserID: user.ID, SourceType: "grammar_training",
		GrammarChapterID: "en.chapter.filters", TheoryBlockID: "tb1", ReportCategory: "typo",
	})
	if err != nil {
		t.Fatalf("create en: %v", err)
	}
	idES, err := repo.Create(CreateContentReportInput{
		UserID: user.ID, SourceType: "grammar_training",
		GrammarChapterID: "es.chapter.filters", TheoryBlockID: "tb2", ReportCategory: "wrong_answer",
	})
	if err != nil {
		t.Fatalf("create es: %v", err)
	}

	allList, err := repo.List("", 0)
	if err != nil || len(allList) < 2 {
		t.Fatalf("List default limit: n=%d err=%v", len(allList), err)
	}
	activeOnly, err := repo.List(string(models.ContentReportStatusActive), 2000)
	if err != nil {
		t.Fatalf("List active capped: %v", err)
	}
	if len(activeOnly) > 1000 {
		t.Fatalf("List should cap at 1000, got %d", len(activeOnly))
	}

	filtered, err := repo.ListActiveReports(ListActiveReportsFilter{
		Status:      string(models.ContentReportStatusActive),
		SourceType:  "grammar_training",
		ChapterID:   "en.chapter.filters",
		TheoryBlock: "tb1",
		Category:    "typo",
		Course:      "english",
		CursorID:    idES + 1,
		Limit:       5,
	})
	if err != nil || len(filtered) != 1 || filtered[0].ID != idEN {
		t.Fatalf("ListActiveReports filtered: %+v err=%v", filtered, err)
	}

	esSummary, err := repo.SummaryActiveReports("spanish")
	if err != nil {
		t.Fatalf("SummaryActiveReports es: %v", err)
	}
	foundES := false
	for _, row := range esSummary {
		if row.GrammarChapterID == "es.chapter.filters" {
			foundES = true
		}
	}
	if !foundES {
		t.Fatalf("expected es summary row, got %+v", esSummary)
	}

	if _, err := repo.SummaryActiveReports("bad-course"); err == nil {
		t.Fatal("expected unsupported course in SummaryActiveReports")
	}
	if _, err := repo.ListActiveReports(ListActiveReportsFilter{Course: "bad"}); err == nil {
		t.Fatal("expected unsupported course in ListActiveReports")
	}

	enSummary, err := repo.SummaryActiveReports("en")
	if err != nil {
		t.Fatalf("SummaryActiveReports en: %v", err)
	}
	_ = enSummary

	capped, err := repo.ListActiveReports(ListActiveReportsFilter{Limit: 5000})
	if err != nil {
		t.Fatalf("ListActiveReports limit cap: %v", err)
	}
	_ = capped
}

func TestContentReportRepository_ResolveBulkEmpty(t *testing.T) {
	_, repo := setupContentReportUser(t, 900016)
	affected, err := repo.ResolveBulk(nil, nil)
	if err != nil || affected != 0 {
		t.Fatalf("ResolveBulk empty: affected=%d err=%v", affected, err)
	}
}

func TestContentReportRepository_GetWordTrainingContext(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	user, err := NewUserRepository(conn, logger).GetOrCreateUser(900017)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	repo := NewContentReportRepository(conn, logger)

	var wordCardID, trainingCardID, userCardID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en) VALUES ('perro', 'dog', 'perro') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("word card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos) VALUES (?, 'perro', 0, 'собака', 'dog', 'noun') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("training card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'new', 2.5) RETURNING id`, user.ID, trainingCardID).Scan(&userCardID); err != nil {
		t.Fatalf("user card: %v", err)
	}

	ctx, err := repo.GetWordTrainingContextByUserCard(userCardID)
	if err != nil || ctx == nil || ctx.Word != "perro" {
		t.Fatalf("GetWordTrainingContextByUserCard: %+v err=%v", ctx, err)
	}
	if ctx.WordCardID == nil || ctx.TrainingCardID == nil || ctx.UserCardID == nil {
		t.Fatalf("expected ids populated: %+v", ctx)
	}

	none, err := repo.GetWordTrainingContextByUserCard(999999999)
	if err != nil || none != nil {
		t.Fatalf("missing user card: %+v err=%v", none, err)
	}
}

func TestContentReportRepository_ParsePayload(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := NewContentReportRepository(nil, logger)

	if len(repo.ParsePayload("")) != 0 {
		t.Fatal("empty payload should return empty map")
	}
	got := repo.ParsePayload(`{"k":"v"}`)
	if got["k"] != "v" {
		t.Fatalf("valid payload: %+v", got)
	}
	if len(repo.ParsePayload("{invalid")) != 0 {
		t.Fatal("invalid payload should return empty map")
	}
	repoNoLogger := NewContentReportRepository(nil, nil)
	if len(repoNoLogger.ParsePayload("{invalid")) != 0 {
		t.Fatal("invalid payload without logger")
	}
}

// ============================================================
// content_report_repository.go — DB error paths
// ============================================================

func TestContentReportRepository_DBErrors(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewContentReportRepository(db, zap.NewNop())

	if _, err := repo.Create(CreateContentReportInput{UserID: 1, SourceType: "word_training"}); err == nil {
		t.Fatal("Create expected error")
	}
	if _, err := repo.HasClientReport(1, "x"); err == nil {
		t.Fatal("HasClientReport expected error")
	}
	if err := repo.Resolve(1, 1); err == nil {
		t.Fatal("Resolve expected error")
	}
	if _, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{}); err == nil {
		t.Fatal("ListActiveGrammarReports expected error")
	}
	if _, err := repo.ListActiveReports(ListActiveReportsFilter{}); err == nil {
		t.Fatal("ListActiveReports expected error")
	}
	if _, err := repo.SummaryActiveReports(""); err == nil {
		t.Fatal("SummaryActiveReports expected error")
	}
	if _, err := repo.ResolveBulk([]int64{1}, nil); err == nil {
		t.Fatal("ResolveBulk expected error")
	}
	if _, err := repo.List("", 10); err == nil {
		t.Fatal("List expected error")
	}
	if _, err := repo.GetByID(1); err == nil {
		t.Fatal("GetByID expected error")
	}
	if _, err := repo.GetWordTrainingContextByUserCard(1); err == nil {
		t.Fatal("GetWordTrainingContextByUserCard expected error")
	}
}

func TestContentReportRepository_scanContentReportError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	repo := NewContentReportRepository(db, zap.NewNop())
	if _, err := repo.GetByID(1); err == nil {
		t.Fatal("expected scan error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestContentReportRepository_SummaryScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT source_type").
		WillReturnRows(sqlmock.NewRows([]string{"source_type"}).AddRow("grammar_training"))

	repo := NewContentReportRepository(db, zap.NewNop())
	if _, err := repo.SummaryActiveReports(""); err == nil {
		t.Fatal("expected summary scan error")
	}
}

func TestContentReportRepository_ResolveRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE content_reports").
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	repo := NewContentReportRepository(db, zap.NewNop())
	if err := repo.Resolve(1, 1); err == nil {
		t.Fatal("expected rows affected error")
	}
}

func TestContentReportRepository_ResolveBulkRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE content_reports").
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	repo := NewContentReportRepository(db, zap.NewNop())
	if _, err := repo.ResolveBulk([]int64{1}, nil); err == nil {
		t.Fatal("expected bulk rows affected error")
	}
}

func TestContentReportRepository_ListActiveReportsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	repo := NewContentReportRepository(db, zap.NewNop())
	if _, err := repo.ListActiveReports(ListActiveReportsFilter{}); err == nil {
		t.Fatal("expected list active scan error")
	}
}

func TestContentReportRepository_ListScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	repo := NewContentReportRepository(db, zap.NewNop())
	if _, err := repo.List("", 10); err == nil {
		t.Fatal("expected list scan error")
	}
}

// ============================================================
// conversation_repository.go — sqlmock edge paths
// ============================================================

var (
	conversationSessionCols = []string{"id", "user_course_id", "scenario_id", "status", "turn_count", "tokens_used", "started_at", "completed_at"}
)

func TestConversationRepository_StartSessionErrors(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewConversationRepository(db, zap.NewNop())

	t.Run("create session fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM conversation_sessions").
			WillReturnRows(sqlmock.NewRows(conversationSessionCols))
		mock.ExpectQuery("INSERT INTO conversation_sessions").
			WillReturnError(sql.ErrConnDone)
		if _, _, err := repo.StartSession(ctx, 1, 1); err == nil {
			t.Fatal("expected create session error")
		}
	})

	t.Run("get session after create fails", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM conversation_sessions").
			WillReturnRows(sqlmock.NewRows(conversationSessionCols))
		mock.ExpectQuery("INSERT INTO conversation_sessions").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
		mock.ExpectQuery("SELECT .+ FROM conversation_sessions").
			WillReturnError(sql.ErrConnDone)
		if _, _, err := repo.StartSession(ctx, 1, 1); err == nil {
			t.Fatal("expected get session after create error")
		}
	})
}

func TestConversationRepository_RecordQuestCompletionEventError(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewConversationRepository(db, zap.NewNop())

	mock.ExpectExec("INSERT INTO exercise_attempts").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO learning_events").
		WillReturnError(sql.ErrConnDone)

	if err := repo.RecordQuestCompletion(ctx, 1, sql.NullInt64{}, "code", 1); err == nil {
		t.Fatal("expected learning event insert error")
	}
}

func TestConversationRepository_RowScanErrors(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewConversationRepository(db, zap.NewNop())

	cases := []struct {
		name string
		fn   func() error
	}{
		{"ListScenariosForDistrict", func() error {
			mock.ExpectQuery("SELECT .+ FROM conversation_scenarios").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			_, err := repo.ListScenariosForDistrict(ctx, 1, 1)
			return err
		}},
		{"ListTasks", func() error {
			mock.ExpectQuery("SELECT id, scenario_id").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			_, err := repo.ListTasks(ctx, 1)
			return err
		}},
		{"ListMessages", func() error {
			mock.ExpectQuery("SELECT id, seq").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			_, err := repo.ListMessages(ctx, 1)
			return err
		}},
		{"GetCompletedTaskIDs", func() error {
			mock.ExpectQuery("SELECT task_id").
				WillReturnRows(sqlmock.NewRows([]string{"task_id"}).AddRow("not-int"))
			_, err := repo.GetCompletedTaskIDs(ctx, 1)
			return err
		}},
		{"LatestPassedScenarioCodes", func() error {
			mock.ExpectQuery("SELECT DISTINCT sc.code").
				WillReturnRows(sqlmock.NewRows([]string{"code", "extra"}).AddRow("x", "y"))
			_, err := repo.LatestPassedScenarioCodes(ctx, 1, 1)
			return err
		}},
		{"PassedAtByScenarioCode", func() error {
			mock.ExpectQuery("SELECT code, MAX").
				WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("x"))
			_, err := repo.PassedAtByScenarioCode(ctx, 1, 1)
			return err
		}},
		{"LatestCompletedScenarioCodes", func() error {
			mock.ExpectQuery("SELECT DISTINCT sc.code").
				WillReturnRows(sqlmock.NewRows([]string{"code", "extra"}).AddRow("x", "y"))
			_, err := repo.LatestCompletedScenarioCodes(ctx, 1, 1)
			return err
		}},
		{"GetNPCImages", func() error {
			mock.ExpectQuery("SELECT npc_code, image_url").
				WillReturnRows(sqlmock.NewRows([]string{"npc_code"}).AddRow("mara"))
			_, err := repo.GetNPCImages(ctx, 1)
			return err
		}},
		{"CompletedAtByScenarioCode", func() error {
			mock.ExpectQuery("SELECT sc.code, MAX").
				WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("x"))
			_, err := repo.CompletedAtByScenarioCode(ctx, 1, 1)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err == nil {
				t.Fatalf("%s: expected scan error", tc.name)
			}
		})
	}
}
