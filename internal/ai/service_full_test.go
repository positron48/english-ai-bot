package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestGenerateResponse_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Bearer test-key authorization")
		}
		
		// Parse request body
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Failed to parse request: %v", err)
		}
		
		if req.Model != "test-model" {
			t.Errorf("Expected model test-model, got %s", req.Model)
		}
		
		// Send response
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "Test response"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx := context.Background()
	response, err := service.GenerateResponse(ctx, "Hello")
	
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}
	
	if response != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", response)
	}
}

func TestGenerateResponse_APIError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Error: &Error{
				Message: "API error occurred",
				Type:    "api_error",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")
	
	if err == nil {
		t.Error("Expected error for API error response")
	}
}

func TestGenerateResponse_HTTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")
	
	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestGenerateResponse_NoChoices(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")
	
	if err == nil {
		t.Error("Expected error for empty choices")
	}
}

func TestGenerateResponse_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")
	
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGenerateResponse_ContextCanceled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(ChatResponse{
			Choices: []Choice{{Message: Message{Content: "response"}}},
		})
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err := service.GenerateResponse(ctx, "Hello")
	
	if err == nil {
		t.Error("Expected error for canceled context")
	}
}

func TestGenerateTrainingCard_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test", "meaning": "тест"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate card for: ")
	
	ctx := context.Background()
	response, err := service.GenerateTrainingCard(ctx, "test")
	
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
	
	if response != `{"word": "test", "meaning": "тест"}` {
		t.Errorf("Unexpected response: %s", response)
	}
}

func TestGenerateTrainingCard_NoPrompt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	service := NewService("http://test", "test-model", "test-key", "test prompt", logger)
	// Don't set training prompt
	
	ctx := context.Background()
	_, err := service.GenerateTrainingCard(ctx, "test")
	
	if err == nil {
		t.Error("Expected error for missing training prompt")
	}
}

func TestGenerateTrainingCard_WithModelOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	var usedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		json.Unmarshal(body, &req)
		usedModel = req.Model
		
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	_, err := service.GenerateTrainingCard(ctx, "test", "override-model")
	
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
	
	if usedModel != "override-model" {
		t.Errorf("Expected override-model, got %s", usedModel)
	}
}

func TestGenerateTrainingCard_CleansMarkdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: "```json\n{\"word\": \"test\"}\n```"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	response, err := service.GenerateTrainingCard(ctx, "test")
	
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
	
	if response != `{"word": "test"}` {
		t.Errorf("Expected cleaned JSON, got: %s", response)
	}
}

func TestGenerateTrainingCard_HTTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Unavailable"))
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	_, err := service.GenerateTrainingCard(ctx, "test")
	
	if err == nil {
		t.Error("Expected error for HTTP 503")
	}
}

func TestGenerateAdditionalTrainingCard_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	var receivedContent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		json.Unmarshal(body, &req)
		if len(req.Messages) > 0 {
			receivedContent = req.Messages[0].Content
		}
		
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "run", "pos": "verb"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate card: ")
	
	ctx := context.Background()
	response, err := service.GenerateAdditionalTrainingCard(ctx, "run", "Part of speech: verb")
	
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
	
	if response != `{"word": "run", "pos": "verb"}` {
		t.Errorf("Unexpected response: %s", response)
	}
	
	// Check that constraints were included
	if receivedContent == "" {
		t.Error("Expected content to be sent")
	}
}

func TestGenerateAdditionalTrainingCard_NoPrompt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	service := NewService("http://test", "test-model", "test-key", "test prompt", logger)
	// Don't set training prompt
	
	ctx := context.Background()
	_, err := service.GenerateAdditionalTrainingCard(ctx, "test", "constraints")
	
	if err == nil {
		t.Error("Expected error for missing training prompt")
	}
}

func TestGenerateAdditionalTrainingCard_WithModelOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	var usedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		json.Unmarshal(body, &req)
		usedModel = req.Model
		
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	_, err := service.GenerateAdditionalTrainingCard(ctx, "test", "constraints", "gpt-4")
	
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
	
	if usedModel != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", usedModel)
	}
}

func TestGenerateAdditionalTrainingCard_EmptyConstraints(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	_, err := service.GenerateAdditionalTrainingCard(ctx, "test", "")
	
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
}

func TestGenerateAdditionalTrainingCard_CleansMarkdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: "```\n{\"word\": \"test\"}\n```"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	service := NewService(server.URL, "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	
	ctx := context.Background()
	response, err := service.GenerateAdditionalTrainingCard(ctx, "test", "")
	
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
	
	if response != `{"word": "test"}` {
		t.Errorf("Expected cleaned JSON, got: %s", response)
	}
}

func TestNewService_ProcessesEscapedNewlines(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	service := NewService("http://test", "model", "key", "Hello\\nWorld", logger)
	
	if service.prompt != "Hello\nWorld" {
		t.Errorf("Expected processed newlines, got %q", service.prompt)
	}
}
