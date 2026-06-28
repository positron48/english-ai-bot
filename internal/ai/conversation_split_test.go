package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

func TestConversationTurnSplit_NPCPromptRequired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		NPCPrompt:   "  ",
		UserMessage: "hello",
	})
	if err == nil {
		t.Fatal("expected error for empty NPC prompt")
	}
	if !strings.Contains(err.Error(), "npc prompt is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConversationTurnSplit_SkipsQuestAndCorrection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	var callCount int32
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&callCount, 1)
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("parse request: %v", err)
		}
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[0].Content != "npc only" {
			t.Fatalf("expected single NPC system+user call, got %+v", req.Messages)
		}
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: "Hello there!"}}},
			Usage:   &Usage{PromptTokens: 7, CompletionTokens: 3},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	res, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		QuestPrompt:         "quest evaluator",
		CorrectionPrompt:    "correction evaluator",
		NPCPrompt:           "npc only",
		UserMessage:         "hi",
		EvaluateQuest:       false,
		EvaluateCorrections: false,
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("expected 1 LLM call (NPC only), got %d", callCount)
	}
	if res.VisibleContent != "Hello there!" {
		t.Errorf("VisibleContent = %q", res.VisibleContent)
	}
	if len(res.CompletedTaskCodes) != 0 || len(res.Corrections) != 0 {
		t.Errorf("expected no quest/correction merge, got codes=%v corrections=%v", res.CompletedTaskCodes, res.Corrections)
	}
	if res.PromptTokens != 7 || res.CompletionTokens != 3 {
		t.Errorf("tokens = %d/%d, want 7/3", res.PromptTokens, res.CompletionTokens)
	}
}

func TestConversationTurnSplit_SkipsWhenPromptsEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	var callCount int32
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&callCount, 1)
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: "NPC reply"}}},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		NPCPrompt:           "npc",
		UserMessage:         "hello",
		EvaluateQuest:       true,
		QuestPrompt:         "   ",
		EvaluateCorrections: true,
		CorrectionPrompt:    "",
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("expected 1 LLM call when quest/correction prompts empty, got %d", callCount)
	}
}

func TestConversationTurnSplit_TokenAggregation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		_ = json.Unmarshal(body, &req)

		kind := "npc"
		if len(req.Messages) > 0 {
			switch {
			case strings.Contains(req.Messages[0].Content, "EVAL_QUEST"):
				kind = "quest"
			case strings.Contains(req.Messages[0].Content, "EVAL_CORRECTION"):
				kind = "correction"
			}
		}

		var content string
		switch kind {
		case "quest":
			content = `{"completed_task_codes":["greet"],"all_done":false}`
		case "correction":
			content = `{"corrections":[{"original":"helo","corrected":"hello","explanation":"typo"}]}`
		default:
			content = "Welcome!"
		}

		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: content}}},
			Usage:   &Usage{PromptTokens: 11, CompletionTokens: 4},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	res, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		QuestPrompt:         "EVAL_QUEST_TASKS",
		CorrectionPrompt:    "EVAL_CORRECTION_ERRORS",
		NPCPrompt:           "NPC_CHARACTER_REPLY",
		UserMessage:         "helo",
		EvaluateQuest:       true,
		EvaluateCorrections: true,
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
	if res.PromptTokens != 33 || res.CompletionTokens != 12 {
		t.Errorf("aggregated tokens = %d/%d, want 33/12", res.PromptTokens, res.CompletionTokens)
	}
	if !strings.Contains(res.Raw, "quest:") || !strings.Contains(res.Raw, "correction:") || !strings.Contains(res.Raw, "npc:") {
		t.Errorf("Raw should contain all parallel parts, got %q", res.Raw)
	}
}

func TestConversationTurnSplit_ModelOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("conversation model", func(t *testing.T) {
		service := NewService("http://example.com", "default-model", "test-key", "p", logger)
		service.SetConversationModel("conv-model")
		service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(r.Body)
			var req ChatRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("parse request: %v", err)
			}
			if req.Model != "conv-model" {
				t.Errorf("model = %q, want conv-model", req.Model)
			}
			resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
			return newJSONResponse(http.StatusOK, resp), nil
		})

		_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
			NPCPrompt:   "npc",
			UserMessage: "hello",
		})
		if err != nil {
			t.Fatalf("ConversationTurnSplit() error = %v", err)
		}
	})

	t.Run("explicit override wins", func(t *testing.T) {
		service := NewService("http://example.com", "default-model", "test-key", "p", logger)
		service.SetConversationModel("conv-model")
		service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(r.Body)
			var req ChatRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("parse request: %v", err)
			}
			if req.Model != "override-model" {
				t.Errorf("model = %q, want override-model", req.Model)
			}
			resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
			return newJSONResponse(http.StatusOK, resp), nil
		})

		_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
			NPCPrompt:   "npc",
			UserMessage: "hello",
		}, "override-model")
		if err != nil {
			t.Fatalf("ConversationTurnSplit() error = %v", err)
		}
	})

	t.Run("empty override keeps conversation model", func(t *testing.T) {
		service := NewService("http://example.com", "default-model", "test-key", "p", logger)
		service.SetConversationModel("conv-model")
		service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(r.Body)
			var req ChatRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("parse request: %v", err)
			}
			if req.Model != "conv-model" {
				t.Errorf("model = %q, want conv-model", req.Model)
			}
			resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
			return newJSONResponse(http.StatusOK, resp), nil
		})

		_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
			NPCPrompt:   "npc",
			UserMessage: "hello",
		}, "  ")
		if err != nil {
			t.Fatalf("ConversationTurnSplit() error = %v", err)
		}
	})
}

func TestConversationTurnSplit_DefaultMaxTokens(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("parse request: %v", err)
		}
		if req.MaxTokens != 600 {
			t.Errorf("MaxTokens = %d, want 600 default", req.MaxTokens)
		}
		resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		NPCPrompt:   "npc",
		UserMessage: "hello",
		MaxTokens:   0,
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
}

func TestConversationTurnSplit_ParallelErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name    string
		setup   func(*Service)
		wantErr string
	}{
		{
			name: "quest HTTP error",
			setup: func(s *Service) {
				s.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					body, _ := io.ReadAll(r.Body)
					var req ChatRequest
					_ = json.Unmarshal(body, &req)
					if len(req.Messages) == 0 {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}), nil
					}
					sys := req.Messages[0].Content
					switch {
					case strings.Contains(sys, "EVAL_QUEST"):
						return nil, errors.New("quest provider down")
					case strings.Contains(sys, "EVAL_CORRECTION"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"corrections":[]}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					default:
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					}
				})
			},
			wantErr: "quest evaluation:",
		},
		{
			name: "quest parse error",
			setup: func(s *Service) {
				s.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					body, _ := io.ReadAll(r.Body)
					var req ChatRequest
					_ = json.Unmarshal(body, &req)
					if len(req.Messages) == 0 {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}), nil
					}
					sys := req.Messages[0].Content
					switch {
					case strings.Contains(sys, "EVAL_QUEST"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "not json"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					case strings.Contains(sys, "EVAL_CORRECTION"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"corrections":[]}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					default:
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					}
				})
			},
			wantErr: "quest evaluation parse:",
		},
		{
			name: "correction HTTP error",
			setup: func(s *Service) {
				s.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					body, _ := io.ReadAll(r.Body)
					var req ChatRequest
					_ = json.Unmarshal(body, &req)
					if len(req.Messages) == 0 {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}), nil
					}
					sys := req.Messages[0].Content
					switch {
					case strings.Contains(sys, "EVAL_QUEST"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"completed_task_codes":[],"all_done":false}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					case strings.Contains(sys, "EVAL_CORRECTION"):
						return nil, errors.New("correction provider down")
					default:
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					}
				})
			},
			wantErr: "correction evaluation:",
		},
		{
			name: "correction parse error",
			setup: func(s *Service) {
				s.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					body, _ := io.ReadAll(r.Body)
					var req ChatRequest
					_ = json.Unmarshal(body, &req)
					if len(req.Messages) == 0 {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}), nil
					}
					sys := req.Messages[0].Content
					switch {
					case strings.Contains(sys, "EVAL_QUEST"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"completed_task_codes":[],"all_done":false}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					case strings.Contains(sys, "EVAL_CORRECTION"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "no json here"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					default:
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					}
				})
			},
			wantErr: "correction evaluation parse:",
		},
		{
			name: "npc HTTP error",
			setup: func(s *Service) {
				s.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					body, _ := io.ReadAll(r.Body)
					var req ChatRequest
					_ = json.Unmarshal(body, &req)
					if len(req.Messages) == 0 {
						return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}), nil
					}
					sys := req.Messages[0].Content
					switch {
					case strings.Contains(sys, "EVAL_QUEST"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"completed_task_codes":[],"all_done":false}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					case strings.Contains(sys, "EVAL_CORRECTION"):
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: `{"corrections":[]}`}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					case strings.Contains(sys, "NPC_CHARACTER"):
						return nil, errors.New("npc provider down")
					default:
						resp := ChatResponse{Choices: []Choice{{Message: Message{Content: "Hi"}}}}
						return newJSONResponse(http.StatusOK, resp), nil
					}
				})
			},
			wantErr: "npc reply:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService("http://example.com", "test-model", "test-key", "p", logger)
			tt.setup(service)

			_, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
				QuestPrompt:         "EVAL_QUEST_TASKS",
				CorrectionPrompt:    "EVAL_CORRECTION_ERRORS",
				NPCPrompt:           "NPC_CHARACTER_REPLY",
				UserMessage:         "hello",
				EvaluateQuest:       true,
				EvaluateCorrections: true,
			})
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestConversationTurnSplit_SuccessAllParallel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		_ = json.Unmarshal(body, &req)

		kind := "npc"
		if len(req.Messages) > 0 {
			switch {
			case strings.Contains(req.Messages[0].Content, "EVAL_QUEST"):
				kind = "quest"
			case strings.Contains(req.Messages[0].Content, "EVAL_CORRECTION"):
				kind = "correction"
			}
		}

		var content string
		switch kind {
		case "quest":
			content = `{"completed_task_codes":["greet"],"all_done":false}`
		case "correction":
			content = `{"corrections":[{"original":"helo","corrected":"hello","explanation":"typo"}]}`
		default:
			content = "Hello! Welcome to the cafe."
		}

		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: content}}},
			Usage:   &Usage{PromptTokens: 10, CompletionTokens: 5},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	res, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		QuestPrompt:         "EVAL_QUEST_TASKS",
		CorrectionPrompt:    "EVAL_CORRECTION_ERRORS",
		NPCPrompt:           "NPC_CHARACTER_REPLY",
		History:             []Message{{Role: "assistant", Content: "Welcome"}},
		UserMessage:         "helo",
		MaxTokens:           200,
		EvaluateQuest:       true,
		EvaluateCorrections: true,
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
	if res.VisibleContent != "Hello! Welcome to the cafe." {
		t.Errorf("VisibleContent = %q", res.VisibleContent)
	}
	if !reflect.DeepEqual(res.CompletedTaskCodes, []string{"greet"}) {
		t.Errorf("CompletedTaskCodes = %v", res.CompletedTaskCodes)
	}
	if len(res.Corrections) != 1 || res.Corrections[0].Corrected != "hello" {
		t.Errorf("Corrections = %v", res.Corrections)
	}
}

func TestUnmarshalLLMJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{name: "empty", raw: "   ", wantErr: "empty response"},
		{name: "no json object", raw: "just text", wantErr: "no JSON object found"},
		{
			name: "embedded object",
			raw:  "Result: {\"completed_task_codes\":[\"pay\"],\"all_done\":false}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sig questEvalSignal
			err := unmarshalLLMJSON(tt.raw, &sig)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshalLLMJSON() error = %v", err)
			}
			if !reflect.DeepEqual(sig.CompletedTaskCodes, []string{"pay"}) {
				t.Errorf("codes = %v", sig.CompletedTaskCodes)
			}
		})
	}
}

func TestConversationTurnSplit_NilUsageIgnored(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: "Hi there"}}},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	res, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		NPCPrompt:   "npc",
		UserMessage: "hello",
	})
	if err != nil {
		t.Fatalf("ConversationTurnSplit() error = %v", err)
	}
	if res.PromptTokens != 0 || res.CompletionTokens != 0 {
		t.Errorf("expected zero tokens when usage missing, got %d/%d", res.PromptTokens, res.CompletionTokens)
	}
}

func TestParseQuestEvalJSON_Error(t *testing.T) {
	_, err := parseQuestEvalJSON("not json at all")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseCorrectionEvalJSON_Error(t *testing.T) {
	_, err := parseCorrectionEvalJSON("")
	if err == nil {
		t.Fatal("expected parse error")
	}
}
