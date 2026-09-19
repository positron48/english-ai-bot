package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestSentenceGenerationSharedPool(t *testing.T) {
	for _, tc := range []struct {
		name                                       string
		firstAccepted, refillAccepted, want, calls int
		refillInvalidJSON                          bool
	}{
		{name: "full pool in two requests", firstAccepted: 30, want: 20, calls: 2},
		{name: "targeted refill", firstAccepted: 10, refillAccepted: 10, want: 20, calls: 4},
		{name: "keep partial set", firstAccepted: 10, want: 10, calls: 4},
		{name: "keep partial after malformed refill review", firstAccepted: 10, want: 10, calls: 4, refillInvalidJSON: true},
		{name: "empty reviewed pool", want: 0, calls: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService("https://example.test", "default", "test", "", zap.NewNop())
			svc.SetSentenceModel("selected-sentence-model")
			svc.SetSentenceGenPromptForCourse("es_ru", "Generate")
			calls, generations, reviews := 0, 0, 0
			var current []GeneratedSentence
			focus := []GenSentenceWord{{"beber", "пить"}, {"comer", "есть"}}
			svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				var body ChatRequest
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Model != "selected-sentence-model" {
					t.Fatalf("wrong model: %s", body.Model)
				}
				var payload struct {
					Count    int                 `json:"sentence_count"`
					Focus    []GenSentenceWord   `json:"focus_words"`
					Accepted []GeneratedSentence `json:"accepted_sentences"`
					Rejected []sentenceRejection `json:"rejected_sentences"`
				}
				if err := json.Unmarshal([]byte(body.Messages[1].Content), &payload); err != nil {
					t.Fatal(err)
				}
				var response any
				if payload.Count > 0 {
					generations++
					if len(payload.Focus) != len(focus) {
						t.Fatal("focus vocabulary was partitioned")
					}
					if generations == 1 && payload.Count != 30 {
						t.Fatalf("wanted 30 initial candidates, got %d", payload.Count)
					}
					if generations == 2 {
						if len(payload.Accepted) != tc.firstAccepted || len(payload.Rejected) != 30-tc.firstAccepted {
							t.Fatalf("missing refill feedback: %+v", payload)
						}
						if len(payload.Rejected) > 0 && payload.Rejected[0].Reason != "Unnatural sentence; choose an ordinary situation." {
							t.Fatal("review reason not forwarded")
						}
					}
					current = nil
					for i := 0; i < payload.Count; i++ {
						// Text is synthetic: this test verifies orchestration, not linguistics.
						current = append(current, GeneratedSentence{PromptRU: fmt.Sprintf("Я пью воду %d.", generations*100+i), ReferenceES: fmt.Sprintf("Bebo agua %d.", generations*100+i), UsedWords: []string{"beber"}})
					}
					response = generatedSentenceSet{Sentences: current}
				} else {
					reviews++
					if reviews == 2 && len(payload.Accepted) != tc.firstAccepted {
						t.Fatal("refill review is missing accepted sentences for diversity checks")
					}
					if reviews == 2 && tc.refillInvalidJSON {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "invalid JSON"}}}}), nil
					}
					accepted := tc.firstAccepted
					if reviews == 2 {
						accepted = tc.refillAccepted
					}
					var checks []sentenceQualityCheck
					for i := range current {
						checks = append(checks, sentenceQualityCheck{Position: i, Accepted: i < accepted, Reason: "Unnatural sentence; choose an ordinary situation."})
					}
					response = sentenceQualityReview{Checks: checks}
				}
				raw, err := json.Marshal(response)
				if err != nil {
					t.Fatal(err)
				}
				return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: string(raw)}}}}), nil
			})
			got, err := svc.GenerateSentenceSetForCourse(context.Background(), "es_ru", focus, nil, []string{"presente"}, 20)
			if tc.want == 0 {
				if !errors.Is(err, ErrNoReviewedSentences) {
					t.Fatalf("wrong empty result error: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.want || calls != tc.calls {
				t.Fatalf("got %d sentences, %d calls; want %d, %d", len(got), calls, tc.want, tc.calls)
			}
		})
	}
}

func TestSentenceGenerationRepairsRejectedContextWithoutSeparateCall(t *testing.T) {
	svc := NewService("https://example.test", "sentence", "test", "", zap.NewNop())
	svc.SetSentenceGenPromptForCourse("es_ru", "Generate")
	sentence := GeneratedSentence{PromptRU: "Я читаю книгу.", ReferenceES: "Leo el libro.", ClarificationRU: "Собеседники обсуждают конкретную книгу.", UsedWords: []string{"leer"}}
	const good = "Собеседник только что передал вам книгу и спросил, что вы с ней делаете."
	calls := 0
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		var body ChatRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		var response any
		switch calls {
		case 1:
			response = generatedSentenceSet{Sentences: []GeneratedSentence{sentence}}
		case 2:
			if !strings.Contains(body.Messages[1].Content, "Rewrite clarification_ru") || !strings.Contains(body.Messages[1].Content, sentence.ClarificationRU) {
				t.Fatal("local rejection not forwarded to refill")
			}
			sentence.ClarificationRU = good
			duplicate := sentence
			duplicate.ReferenceES = "Yo leo el libro."
			response = generatedSentenceSet{Sentences: []GeneratedSentence{sentence, duplicate}}
		case 3:
			response = sentenceQualityReview{Checks: []sentenceQualityCheck{{Position: 0, Accepted: true, Reason: "Книга передана собеседником.", ContextEvidence: []string{good}}, {Position: 1, Accepted: true, Reason: "Книга передана собеседником.", ContextEvidence: []string{good}}}}
		default:
			t.Fatal("unexpected extra request")
		}
		raw, _ := json.Marshal(response)
		return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: string(raw)}}}}), nil
	})
	got, err := svc.GenerateSentenceSetForCourse(context.Background(), "es_ru", []GenSentenceWord{{"leer", "читать"}}, nil, []string{"presente"}, 2)
	if err != nil || len(got) != 1 || calls != 3 || got[0].ClarificationRU != good {
		t.Fatalf("got=%+v err=%v calls=%d", got, err, calls)
	}
}

func TestSentenceReasoningAndReviewTemperatureAreScoped(t *testing.T) {
	svc := NewService("https://example.test", "sentence", "test", "", zap.NewNop())
	svc.SetSentenceReasoningEffort(" LOW ")
	calls := 0
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if temperature, exists := body["temperature"]; !exists || temperature != float64(0) {
			t.Fatalf("temperature is not explicitly zero: %v", body)
		}
		if calls == 1 {
			reasoning, ok := body["reasoning"].(map[string]any)
			if !ok || reasoning["effort"] != "low" {
				t.Fatalf("missing sentence reasoning: %v", body)
			}
		} else if _, exists := body["reasoning"]; exists {
			t.Fatal("sentence reasoning leaked to general chat")
		}
		return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: `{"checks":[]}`}}}}), nil
	})
	if _, err := svc.reviewGeneratedSentenceQuality(context.Background(), "sentence", "es_ru", nil, nil, nil, []GeneratedSentence{{PromptRU: "Я пью воду.", ReferenceES: "Bebo agua."}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.postChatCompletion(context.Background(), "general", []Message{{Role: "user", Content: "Hello"}}, 100, 0); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}
