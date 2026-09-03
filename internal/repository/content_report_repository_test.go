package repository

import (
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestContentReportRepository_ListActiveGrammarReports_AndResolveBulk(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(db.GetConnection(), logger)
	user, err := userRepo.GetOrCreateUser(777001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	repo := NewContentReportRepository(db.GetConnection(), logger)

	id1, err := repo.Create(CreateContentReportInput{
		UserID:            user.ID,
		SourceType:        "grammar_training",
		GrammarChapterID:  "es.chapter.1",
		TheoryBlockID:     "b1",
		GrammarQuestionID: "es.chapter.1::b1::q1",
	})
	if err != nil {
		t.Fatalf("create #1: %v", err)
	}
	id2, err := repo.Create(CreateContentReportInput{
		UserID:            user.ID,
		SourceType:        "grammar_training",
		GrammarChapterID:  "en.chapter.2",
		TheoryBlockID:     "b2",
		GrammarQuestionID: "en.chapter.2::b2::q1",
	})
	if err != nil {
		t.Fatalf("create #2: %v", err)
	}
	_, err = repo.Create(CreateContentReportInput{
		UserID:     user.ID,
		SourceType: "word_training",
		Word:       "hello",
	})
	if err != nil {
		t.Fatalf("create word report: %v", err)
	}

	list, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports all: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 grammar active reports, got %d", len(list))
	}

	all, err := repo.ListActiveReports(ListActiveReportsFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListActiveReports: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 active reports total, got %d", len(all))
	}
	summary, err := repo.SummaryActiveReports("")
	if err != nil {
		t.Fatalf("SummaryActiveReports: %v", err)
	}
	if len(summary) == 0 {
		t.Fatal("expected summary rows")
	}
	if list[0].ID <= list[1].ID {
		t.Fatalf("expected DESC order by id")
	}

	esOnly, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Course: "es", Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports es: %v", err)
	}
	if len(esOnly) != 1 || esOnly[0].ID != id1 {
		t.Fatalf("expected only es report id=%d, got %+v", id1, esOnly)
	}

	if _, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Course: "xx"}); err == nil {
		t.Fatal("expected unsupported course filter error")
	}

	cursorList, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{CursorID: id2, Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports cursor: %v", err)
	}
	if len(cursorList) != 1 || cursorList[0].ID != id1 {
		t.Fatalf("expected cursor to return id1 only, got %+v", cursorList)
	}

	affected, err := repo.ResolveBulk([]int64{id1, id2}, nil)
	if err != nil {
		t.Fatalf("ResolveBulk: %v", err)
	}
	if affected != 2 {
		t.Fatalf("expected affected=2, got %d", affected)
	}

	after, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports after resolve: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected 0 active grammar reports after resolve, got %d", len(after))
	}
}

func TestContentReportRepository_ReadingCourseFilterAndSummary(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	repo := NewContentReportRepository(conn, zap.NewNop())
	user, err := NewUserRepository(conn, zap.NewNop()).GetOrCreateUser(777002)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO reading_texts (text_id,category_id,title,level,target_language,reading_passage) VALUES ('cms-rain','es_a1','Rain','A1','es','{}'),('cms-sun','en_a1','Sun','A1','en','{}')`); err != nil {
		t.Fatal(err)
	}
	var esIDs = map[int64]bool{}
	for _, item := range []struct {
		id, kind string
		es       bool
	}{
		{"free_es_a1_rain", "reading_text", true},
		{"es_a1_retired", "reading_text", true},
		{"cms-rain", "reading_text", true},
		{"es.chapter", "grammar_chapter", true},
		{"es.chapter", "grammar_test", true},
		{"free_en_a1_rain", "reading_text", false},
		{"cms-sun", "reading_text", false},
		{"freeXesYa1_wrong", "reading_text", false},
	} {
		id, err := repo.Create(CreateContentReportInput{UserID: user.ID, SourceType: item.kind, GrammarChapterID: item.id})
		if err != nil {
			t.Fatal(err)
		}
		if item.es {
			esIDs[id] = true
		}
	}
	for _, course := range []string{"es", "spanish"} {
		rows, err := repo.ListActiveReports(ListActiveReportsFilter{Course: course})
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(esIDs) {
			t.Fatalf("%s: got %d reports, want %d", course, len(rows), len(esIDs))
		}
		for _, row := range rows {
			if !esIDs[row.ID] {
				t.Fatalf("unexpected report %+v", row)
			}
		}
		summary, err := repo.SummaryActiveReports(course)
		if err != nil {
			t.Fatal(err)
		}
		var count int64
		for _, row := range summary {
			count += row.Count
		}
		if count != int64(len(rows)) {
			t.Fatalf("list/summary mismatch: %d/%d", len(rows), count)
		}
	}
	enRows, err := repo.ListActiveReports(ListActiveReportsFilter{Course: "en", SourceType: "reading_text"})
	if err != nil {
		t.Fatal(err)
	}
	if len(enRows) != 2 {
		t.Fatalf("expected 2 English reading reports, got %d", len(enRows))
	}
	page, err := repo.ListActiveReports(ListActiveReportsFilter{Course: "es", SourceType: "reading_text", Limit: 1})
	if err != nil || len(page) != 1 {
		t.Fatalf("first page: %v, %v", page, err)
	}
	rest, err := repo.ListActiveReports(ListActiveReportsFilter{Course: "es", SourceType: "reading_text", CursorID: page[0].ID})
	if err != nil || len(rest) != 2 {
		t.Fatalf("remaining page: %v, %v", rest, err)
	}
}
