package repository

import (
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestGrammarQuestionLookupUsesPrimaryKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT raw_json FROM grammar_training_content_questions WHERE bundle_id = \? AND question_id = \?`).WithArgs("es", "ch::b::q").WillReturnRows(sqlmock.NewRows([]string{"raw_json"}).AddRow(`{"id":"ch::b::q","correct_answer":"a"}`))
	repo := NewGrammarTrainingPackRepositoryFromDB(db, "es", zap.NewNop())
	q, err := repo.GetQuestion("ch::b::q")
	if err != nil || q["correct_answer"] != "a" {
		t.Fatalf("%v %v", q, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestGrammarQuestionLookupDoesNotExposeCachedAnswers(t *testing.T) {
	files := fstest.MapFS{"index.json": {Data: []byte(`{"chapters":{"ch":"one.json"}}`)}, "chapters/one.json": {Data: []byte(`{"questions":[{"id":"q","chapter_id":"ch","theory_block_id":"b","correct_answer":["a"]}]}`)}}
	repo := NewGrammarTrainingPackRepositoryWithFS(files, zap.NewNop())
	q, err := repo.GetQuestion("ch::b::q")
	if err != nil {
		t.Fatal(err)
	}
	q["correct_answer"].([]interface{})[0] = "changed"
	delete(files, "chapters/one.json") // subsequent lookups use the immutable index
	q, err = repo.GetQuestion("ch::b::q")
	if err != nil {
		t.Fatal(err)
	}
	if q["correct_answer"].([]interface{})[0] != "a" {
		t.Fatal("caller mutated shared answer")
	}
}
func TestGrammarUnsupportedQuestionsAreNotSelected(t *testing.T) {
	for _, kind := range []string{"matching", "mcq_multi", "cloze_mixed", "transform", "explain_choice", "production_hint", "short_answer", "unknown"} {
		if GrammarQuestionAvailable(map[string]interface{}{"type": kind}) {
			t.Errorf("unsupported type %s", kind)
		}
	}
}
