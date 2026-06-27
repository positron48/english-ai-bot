package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestParseChatControl(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantVisible string
		wantCodes   []string
		wantAllDone bool
	}{
		{
			name:        "visible plus control",
			raw:         "¡Hola! ¿Qué desea?\n###CONTROL###\n{\"completed_task_codes\":[\"greet\"],\"all_done\":false}",
			wantVisible: "¡Hola! ¿Qué desea?",
			wantCodes:   []string{"greet"},
			wantAllDone: false,
		},
		{
			name:        "no control block",
			raw:         "Just a plain reply",
			wantVisible: "Just a plain reply",
			wantCodes:   nil,
			wantAllDone: false,
		},
		{
			name:        "all done true",
			raw:         "Perfecto, aquí tiene.###CONTROL### {\"completed_task_codes\":[\"order\",\"sugar\"],\"all_done\":true}",
			wantVisible: "Perfecto, aquí tiene.",
			wantCodes:   []string{"order", "sugar"},
			wantAllDone: true,
		},
		{
			name:        "fenced control json",
			raw:         "Vale.\n###CONTROL###\n```json\n{\"completed_task_codes\":[\"pay\"],\"all_done\":false}\n```",
			wantVisible: "Vale.",
			wantCodes:   []string{"pay"},
			wantAllDone: false,
		},
		{
			name:        "garbled control yields empty signal",
			raw:         "Hola.###CONTROL### not json at all",
			wantVisible: "Hola.",
			wantCodes:   nil,
			wantAllDone: false,
		},
		{
			name:        "control with surrounding noise",
			raw:         "Bien.###CONTROL### here you go: {\"completed_task_codes\":[\"greet\"],\"all_done\":false} thanks",
			wantVisible: "Bien.",
			wantCodes:   []string{"greet"},
			wantAllDone: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visible, sig := parseChatControl(tt.raw)
			if visible != tt.wantVisible {
				t.Errorf("visible = %q, want %q", visible, tt.wantVisible)
			}
			if !reflect.DeepEqual(sig.CompletedTaskCodes, tt.wantCodes) {
				t.Errorf("codes = %v, want %v", sig.CompletedTaskCodes, tt.wantCodes)
			}
			if sig.AllDone != tt.wantAllDone {
				t.Errorf("allDone = %v, want %v", sig.AllDone, tt.wantAllDone)
			}
		})
	}
}

func TestConversationTurn(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	var captured ChatRequest
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Role: "assistant", Content: "Hola.\n###CONTROL###\n{\"completed_task_codes\":[\"greet\"],\"all_done\":false}"}}},
			Usage:   &Usage{PromptTokens: 30, CompletionTokens: 12, TotalTokens: 42},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	history := []Message{
		{Role: "assistant", Content: "Welcome line"},
	}
	res, err := service.ConversationTurn(context.Background(), "system prompt", history, "Hola", 200)
	if err != nil {
		t.Fatalf("ConversationTurn() error = %v", err)
	}
	if res.VisibleContent != "Hola." {
		t.Errorf("VisibleContent = %q, want %q", res.VisibleContent, "Hola.")
	}
	if !reflect.DeepEqual(res.CompletedTaskCodes, []string{"greet"}) {
		t.Errorf("CompletedTaskCodes = %v", res.CompletedTaskCodes)
	}
	if res.PromptTokens != 30 || res.CompletionTokens != 12 {
		t.Errorf("tokens = %d/%d, want 30/12", res.PromptTokens, res.CompletionTokens)
	}

	// Messages must be system, history..., user.
	if len(captured.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(captured.Messages))
	}
	if captured.Messages[0].Role != "system" || captured.Messages[0].Content != "system prompt" {
		t.Errorf("first message not system prompt: %+v", captured.Messages[0])
	}
	if captured.Messages[1].Role != "assistant" || captured.Messages[1].Content != "Welcome line" {
		t.Errorf("history message wrong: %+v", captured.Messages[1])
	}
	if captured.Messages[2].Role != "user" || captured.Messages[2].Content != "Hola" {
		t.Errorf("last message not user: %+v", captured.Messages[2])
	}
}

func TestParseQuestEvalJSON(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantCodes   []string
		wantAllDone bool
	}{
		{
			name:        "plain json",
			raw:         `{"completed_task_codes":["greet","ask_price"],"all_done":false}`,
			wantCodes:   []string{"greet", "ask_price"},
			wantAllDone: false,
		},
		{
			name:        "fenced json",
			raw:         "```json\n{\"completed_task_codes\":[\"order\"],\"all_done\":true}\n```",
			wantCodes:   []string{"order"},
			wantAllDone: true,
		},
		{
			name:        "surrounding text",
			raw:         "Here is the result: {\"completed_task_codes\":[\"pay\"],\"all_done\":false}",
			wantCodes:   []string{"pay"},
			wantAllDone: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := parseQuestEvalJSON(tt.raw)
			if err != nil {
				t.Fatalf("parseQuestEvalJSON() error = %v", err)
			}
			if !reflect.DeepEqual(sig.CompletedTaskCodes, tt.wantCodes) {
				t.Errorf("codes = %v, want %v", sig.CompletedTaskCodes, tt.wantCodes)
			}
			if sig.AllDone != tt.wantAllDone {
				t.Errorf("allDone = %v, want %v", sig.AllDone, tt.wantAllDone)
			}
		})
	}
}

func TestParseCorrectionEvalJSON(t *testing.T) {
	raw := `{"corrections":[{"original":"cafe","corrected":"café","explanation":"Ударение"}]}`
	sig, err := parseCorrectionEvalJSON(raw)
	if err != nil {
		t.Fatalf("parseCorrectionEvalJSON() error = %v", err)
	}
	if len(sig.Corrections) != 1 {
		t.Fatalf("expected 1 correction, got %d", len(sig.Corrections))
	}
	if sig.Corrections[0].Original != "cafe" || sig.Corrections[0].Corrected != "café" {
		t.Errorf("correction = %+v", sig.Corrections[0])
	}
}

func TestSanitizeNPCReply(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain", raw: "Hello! How can I help?", want: "Hello! How can I help?"},
		{name: "strips control", raw: "Hi there.\n###CONTROL###\n{\"completed_task_codes\":[]}", want: "Hi there."},
		{name: "strips fences", raw: "```\nHola\n```", want: "Hola"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeNPCReply(tt.raw); got != tt.want {
				t.Errorf("sanitizeNPCReply() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConversationTurnSplit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://example.com", "test-model", "test-key", "p", logger)

	callKinds := make([]string, 0, 3)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		_ = json.Unmarshal(body, &req)
		kind := "unknown"
		if len(req.Messages) > 0 {
			switch {
			case strings.Contains(req.Messages[0].Content, "quest"):
				kind = "quest"
			case strings.Contains(req.Messages[0].Content, "correction"):
				kind = "correction"
			default:
				kind = "npc"
			}
		}
		callKinds = append(callKinds, kind)

		var content string
		switch kind {
		case "quest":
			content = `{"completed_task_codes":["greet"],"all_done":false}`
		case "correction":
			content = `{"corrections":[{"original":"helo","corrected":"hello","explanation":"опечатка"}]}`
		default:
			content = "Hello! Welcome to the cafe."
		}
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Role: "assistant", Content: content}}},
			Usage:   &Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	res, err := service.ConversationTurnSplit(context.Background(), ConversationTurnSplitInput{
		QuestPrompt:         "quest evaluator",
		CorrectionPrompt:    "correction evaluator",
		NPCPrompt:           "npc character",
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
	if res.PromptTokens != 30 || res.CompletionTokens != 15 {
		t.Errorf("tokens = %d/%d, want 30/15", res.PromptTokens, res.CompletionTokens)
	}
	if len(callKinds) != 3 {
		t.Fatalf("expected 3 LLM calls, got %d: %v", len(callKinds), callKinds)
	}
}
