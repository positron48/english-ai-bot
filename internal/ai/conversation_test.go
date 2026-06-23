package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
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
