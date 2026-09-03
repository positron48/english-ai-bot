package service

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"testing/fstest"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const reorderSelectionChapter = `{
 "id":"ch", "ui_language":"ru", "section_id":"sec",
 "blocks":[
   {"id":"b1","type":"theory"},
   {"id":"quiz","type":"quiz_inline","quiz_inline":{"question_ids":["translated","missing","choice"]}},
   {"id":"empty","type":"quiz_inline","quiz_inline":{"question_ids":["missing"]}}
 ],
 "question_bank":{"questions":[
   {"id":"translated","type":"reorder","theory_block_id":"b1","correct_answer":"I work.","translation_ru":"Я работаю."},
   {"id":"missing","type":"reorder","theory_block_id":"b1","correct_answer":"I work here."},
   {"id":"choice","type":"mcq_single","theory_block_id":"b1","correct_answer":"a"}
 ]},
 "chapter_test":{"num_questions":10,"selection_strategy":{"type":"random"},"pool_question_ids":["translated","missing","choice"]}
}`

func TestReorderFilteredFromNewTests(t *testing.T) {
	logger := zap.NewNop()
	content := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{
		"index.json":       {Data: []byte(`{"chapters":{"ch":"ch.json"}}`)},
		"sections.json":    {Data: []byte(`{"sections":[{"section_id":"sec","chapter_ids":["ch"]}]}`)},
		"chapters/ch.json": {Data: []byte(reorderSelectionChapter)},
	}, logger)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := NewGrammarService(content, repository.NewGrammarPublishRepository(db, logger), repository.NewGrammarAttemptRepository(db, logger), config.DefaultLearningConfig(), logger)
	for _, scope := range []string{"chapter", "category", "placement"} {
		t.Run(scope, func(t *testing.T) {
			if scope != "chapter" {
				mock.ExpectQuery("SELECT id, item_type, item_id, is_published").WithArgs("chapter").
					WillReturnRows(sqlmock.NewRows([]string{"id", "item_type", "item_id", "is_published", "name", "updated_at", "updated_by_user_id"}).
						AddRow(1, "chapter", "ch", true, nil, "2026-09-03T00:00:00Z", nil))
			}
			var result *TestQuestions
			var err error
			switch scope {
			case "chapter":
				result, err = svc.GenerateChapterTest(context.Background(), "ch")
			case "category":
				result, err = svc.GenerateCategoryTest(context.Background(), "sec")
			case "placement":
				result, err = svc.GeneratePlacementTest(context.Background())
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Questions) != 2 || result.Total != 2 {
				t.Fatalf("wrong total: %+v", result)
			}
			for _, raw := range result.Questions {
				q := raw.(map[string]interface{})
				if !repository.GrammarQuestionAvailable(q) {
					t.Errorf("unsafe question: %v", q)
				}
				if q["type"] == "reorder" && q["correct_answer"] != "I work." {
					t.Error("word tiles require the answer")
				}
				if q["type"] != "reorder" && q["correct_answer"] != nil {
					t.Error("leaked choice answer")
				}
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestReorderInlineAndTrainingFilteringPreservesBank(t *testing.T) {
	var chapter repository.Chapter
	if err := json.Unmarshal([]byte(reorderSelectionChapter), &chapter); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(chapter)
	svc := &GrammarService{}
	filtered := svc.filterQuestionBankForQuizzes(&chapter)
	if len(filtered.Blocks) != 2 {
		t.Fatalf("empty quiz should disappear: %v", filtered.Blocks)
	}
	quiz := filtered.Blocks[1].(map[string]interface{})["quiz_inline"].(map[string]interface{})
	if !reflect.DeepEqual(quiz["question_ids"], []interface{}{"translated", "choice"}) {
		t.Errorf("quiz references: %v", quiz)
	}
	if len(filtered.QuestionBank["questions"].([]interface{})) != 2 {
		t.Error("untranslated inline question retained")
	}
	after, _ := json.Marshal(chapter)
	if string(before) != string(after) {
		t.Fatal("display filtering mutated source bank or quiz references")
	}
	byBlock := map[string][]map[string]interface{}{}
	for _, raw := range chapter.QuestionBank["questions"].([]interface{}) {
		q := raw.(map[string]interface{})
		q["chapter_id"] = "ch"
		byBlock["b1"] = append(byBlock["b1"], q)
	}
	if got := svc.filterBlocksByAllowedChapters(byBlock, map[string]bool{"ch": true}); len(got["b1"]) != 2 {
		t.Errorf("training filter: %v", got)
	}
}
