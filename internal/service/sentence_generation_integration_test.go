package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestSentenceGenerationPersistsPartialAndSeparatesQualityFailures(t *testing.T) {
	for _, scenario := range []string{"partial", "empty", "provider-invalid-json"} {
		t.Run(scenario, func(t *testing.T) {
			logger := zap.NewNop()
			db := testutil.SetupTestDB(t)
			repo := repository.NewSentenceCompositionRepository(db, logger)
			users := repository.NewUserRepository(db, logger)
			user, err := users.GetOrCreateUser(7821)
			if err != nil {
				t.Fatal(err)
			}
			var wordID int64
			if err := db.QueryRow(`INSERT INTO word_cards (word, definition, definition_ru, course_code) VALUES ('beber', 'пить', 'пить', 'es_ru') RETURNING id`).Scan(&wordID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status, course_code) VALUES (?, ?, 'known', 'es_ru')`, user.ID, wordID); err != nil {
				t.Fatal(err)
			}
			breaker := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 1, logger)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request ai.ChatRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(500)
					return
				}
				var payload struct {
					Count int `json:"sentence_count"`
				}
				_ = json.Unmarshal([]byte(request.Messages[1].Content), &payload)
				content := "not JSON"
				if scenario != "provider-invalid-json" {
					if payload.Count > 0 {
						sentences := []ai.GeneratedSentence{}
						if scenario == "partial" {
							for i := 0; i < 10; i++ {
								sentences = append(sentences, ai.GeneratedSentence{PromptRU: fmt.Sprintf("Я пью воду %d.", i), ReferenceES: fmt.Sprintf("Bebo agua %d.", i), UsedWords: []string{"beber"}})
							}
						}
						raw, _ := json.Marshal(map[string]any{"sentences": sentences})
						content = string(raw)
					} else {
						checks := []map[string]any{}
						for i := 0; i < 10; i++ {
							checks = append(checks, map[string]any{"position": i, "accepted": true, "reason": "Checked.", "context_evidence": []string{}})
						}
						raw, _ := json.Marshal(map[string]any{"checks": checks})
						content = string(raw)
					}
				}
				_ = json.NewEncoder(w).Encode(ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}})
			}))
			defer server.Close()
			aiService := ai.NewService(server.URL, "sentence", "test", "", logger)
			aiService.SetSentenceGenPromptForCourse("es_ru", "Generate")
			worker := NewSentenceCompositionWorker(aiService, repo, users, nil, breaker, config.SentenceCompositionConfig{MinWords: 1, SentencesPerSet: 20}, config.LearningConfig{TargetLang: "es"}, nil, nil, "es_ru", logger)
			id, err := worker.generateSet(context.Background(), user, "es_ru", "2026-09-19")
			latest, loadErr := repo.LatestSet(user.ID, "es_ru")
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			isOpen, failures, _, stateErr := breaker.GetState()
			if stateErr != nil {
				t.Fatal(stateErr)
			}
			if scenario == "partial" {
				if err != nil || id == 0 || latest == nil || latest.Status != models.SentenceSetReady {
					t.Fatalf("id=%d latest=%+v err=%v", id, latest, err)
				}
				items, err := repo.GetItems(id)
				if err != nil || len(items) != 10 {
					t.Fatalf("items=%d err=%v", len(items), err)
				}
				uses, err := repo.ParticipationCount(user.ID, wordID, "es_ru")
				if err != nil || uses != 1 {
					t.Fatalf("usage=%d err=%v", uses, err)
				}
			} else {
				if err == nil || id != 0 || latest != nil {
					t.Fatalf("id=%d latest=%+v err=%v", id, latest, err)
				}
				if scenario == "empty" && !errors.Is(err, ai.ErrNoReviewedSentences) {
					t.Fatalf("unexpected quality error: %v", err)
				}
			}
			wantOpen := scenario == "provider-invalid-json"
			if isOpen != wantOpen || (failures > 0) != wantOpen {
				t.Fatalf("breaker open=%v failures=%d", isOpen, failures)
			}
		})
	}
}
