package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
	"testing/fstest"
)

func TestHandleLearningGrammarChapter_SectionFallbackWhenMetadataMissing(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	fs := fstest.MapFS{
		"sections.json": {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":    {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{
			"schema_version":"1",
			"id":"ch1",
			"section_id":"missing-section",
			"title":"Chapter 1",
			"blocks":[],
			"question_bank":{"questions":[]},
			"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":[],"num_questions":1}
		}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(fs, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	svc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, config.DefaultLearningConfig(), logger)
	router.SetGrammarService(svc)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/ch1", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapter(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
