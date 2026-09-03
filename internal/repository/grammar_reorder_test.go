package repository

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"tgbot-skeleton/internal/config"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const reorderChapterJSON = `{
  "id":"ch", "ui_language":"ru",
  "blocks":[{"id":"b1","type":"theory","theory":{"examples":[
    {"text":"I work at home.","translation":"Я работаю дома."},
    {"text":"I stay.","translation":"Я остаюсь."},
    {"text":"I stay.","translation":"Я останусь."},
    {"text":"I stay.","translation":"Я остаюсь."}
  ]}}],
  "question_bank":{"questions":[
    {"id":"exact","type":"reorder","theory_block_id":"b1","correct_answer":" I  work at home. "},
    {"id":"explicit","type":"reorder","theory_block_id":"b1","correct_answer":"I work at home.","translation_ru":"Я работаю из дома."},
    {"id":"partial","type":"reorder","theory_block_id":"b1","correct_answer":"I work."},
    {"id":"wrong-block","type":"reorder","theory_block_id":"b2","correct_answer":"I work at home."},
    {"id":"ambiguous","type":"reorder","theory_block_id":"b1","correct_answer":"I stay."},
    {"id":"choice","type":"mcq_single"}
  ]}
}`

func TestReorderTranslationFromFSAndDB(t *testing.T) {
	for _, source := range []string{"fs", "db"} {
		t.Run(source, func(t *testing.T) {
			var repo *GrammarContentRepository
			if source == "fs" {
				repo = NewGrammarContentRepositoryWithFS(fstest.MapFS{
					"index.json":       {Data: []byte(`{"chapters":{"ch":"ch.json"}}`)},
					"chapters/ch.json": {Data: []byte(reorderChapterJSON)},
				}, zap.NewNop())
			} else {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				mock.ExpectQuery("SELECT raw_json FROM grammar_content_chapters").WithArgs("en", "ch").
					WillReturnRows(sqlmock.NewRows([]string{"raw_json"}).AddRow(reorderChapterJSON))
				repo = NewGrammarContentRepositoryFromDB(db, "en", zap.NewNop())
				t.Cleanup(func() {
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Error(err)
					}
				})
			}
			chapter, err := repo.GetChapter("ch")
			if err != nil {
				t.Fatal(err)
			}
			questions := chapter.QuestionBank["questions"].([]interface{})
			if len(questions) != 6 {
				t.Fatal("must retain original questions for old attempts")
			}
			for _, raw := range questions {
				q := raw.(map[string]interface{})
				want := q["id"] == "exact" || q["id"] == "explicit" || q["id"] == "choice"
				if GrammarQuestionAvailable(q) != want {
					t.Errorf("availability for %s: %v", q["id"], q)
				}
			}
			if got := questions[0].(map[string]interface{})["translation_ru"]; got != "Я работаю дома." {
				t.Errorf("translation: %v", got)
			}
			if got := questions[1].(map[string]interface{})["translation_ru"]; got != "Я работаю из дома." {
				t.Errorf("overwrote explicit translation: %v", got)
			}
		})
	}
}

func TestReorderTranslationRequiresRussianSource(t *testing.T) {
	var chapter Chapter
	if err := json.Unmarshal([]byte(reorderChapterJSON), &chapter); err != nil {
		t.Fatal(err)
	}
	chapter.UILanguage = "en"
	enrichReorderTranslations(&chapter)
	q := chapter.QuestionBank["questions"].([]interface{})[0].(map[string]interface{})
	if GrammarQuestionAvailable(q) {
		t.Fatal("must not use a translation in another UI language")
	}
	for _, value := range []interface{}{nil, "", " \n\t", 42} {
		if GrammarQuestionAvailable(map[string]interface{}{"type": "reorder", "translation_ru": value}) {
			t.Errorf("accepted invalid translation %v", value)
		}
	}
}

func TestReorderBundledCourseCoverage(t *testing.T) {
	for _, bundle := range []string{"en", "es"} {
		t.Run(bundle, func(t *testing.T) {
			lc := config.DefaultLearningConfig()
			lc.GrammarBundleID = bundle
			repo, err := NewGrammarContentRepositoryForLearning(lc, zap.NewNop())
			if err != nil {
				t.Fatal(err)
			}
			index, err := repo.GetIndex()
			if err != nil {
				t.Fatal(err)
			}
			reorder, translated := 0, 0
			for id := range index.Chapters {
				data, err := repo.GetChapterRawJSON(id)
				if err != nil {
					t.Fatal(err)
				}
				var chapter Chapter
				if err := json.Unmarshal(data, &chapter); err != nil {
					t.Fatal(err)
				}
				available := make(map[string]bool)
				for _, raw := range chapter.QuestionBank["questions"].([]interface{}) {
					q := raw.(map[string]interface{})
					qid, _ := q["id"].(string)
					available[qid] = GrammarQuestionAvailable(q)
					if q["type"] == "reorder" {
						reorder++
						if available[qid] {
							translated++
						}
						answer, _ := q["correct_answer"].(string)
						if words := len(strings.Fields(answer)); words < 2 || words > 10 {
							t.Errorf("%s/%s: reorder needs 2-10 words, got %d", id, qid, words)
						}
					}
				}
				pool, _ := chapter.ChapterTest["pool_question_ids"].([]interface{})
				kept := 0
				for _, raw := range pool {
					if id, ok := raw.(string); ok && available[id] {
						kept++
					}
				}
				if len(pool) > 0 && kept == 0 {
					t.Errorf("chapter %s has no available test questions", id)
				}
			}
			if translated == 0 || translated != reorder {
				t.Errorf("every bundled reorder needs its own Russian translation: %d/%d", translated, reorder)
			}
			t.Logf("reorder: %d; with Russian translation: %d; excluded: %d", reorder, translated, reorder-translated)
		})
	}
}
