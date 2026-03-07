package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newJSONResponse(status int, payload any) *http.Response {
	body, _ := json.Marshal(payload)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func newTextResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
	}
}

// failingReadBody is an io.ReadCloser that fails on Read (to test "failed to read response" path).
type failingReadBody struct{}

func (failingReadBody) Read(p []byte) (n int, err error) { return 0, errors.New("read failed") }
func (failingReadBody) Close() error                    { return nil }

func TestGenerateResponse_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
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

		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Failed to parse request: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("Expected model test-model, got %s", req.Model)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "Test response"}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

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

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Error: &Error{
				Message: "API error occurred",
				Type:    "api_error",
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for API error response")
	}
}

func TestGenerateResponse_HTTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusInternalServerError, "Internal Server Error"), nil
	})

	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestGenerateResponse_NoChoices(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{Choices: []Choice{}}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for empty choices")
	}
}

func TestGenerateResponse_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusOK, "invalid json"), nil
	})

	ctx := context.Background()
	_, err := service.GenerateResponse(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGenerateResponse_ContextCanceled(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-time.After(100 * time.Millisecond):
			resp := ChatResponse{
				Choices: []Choice{{Message: Message{Content: "response"}}},
			}
			return newJSONResponse(http.StatusOK, resp), nil
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GenerateResponse(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for canceled context")
	}
}

func TestGenerateTrainingCard_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test", "meaning": "тест"}`}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})
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

func TestGenerateTrainingCard_ModelOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate card for: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}
		if req.Model != "override-model" {
			t.Errorf("Expected override model, got %s", req.Model)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test", "meaning": "override"}`}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	ctx := context.Background()
	_, err := service.GenerateTrainingCard(ctx, "test", "override-model")
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
}

func TestGenerateTrainingCard_PromptNotSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error when training prompt is not set")
	}
}

func TestGenerateTrainingCard_StripsMarkdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate card for: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: "```json\n{\"word\":\"test\"}\n```"}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	ctx := context.Background()
	response, err := service.GenerateTrainingCard(ctx, "test")
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
	if response != `{"word":"test"}` {
		t.Errorf("Unexpected response: %s", response)
	}
}

func TestGenerateAdditionalTrainingCard_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate card for: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Fatalf("Expected single user message")
		}
		if !strings.Contains(req.Messages[0].Content, "Additional constraints") {
			t.Errorf("Expected constraints to be included")
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test", "meaning": "extra"}`}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	ctx := context.Background()
	response, err := service.GenerateAdditionalTrainingCard(ctx, "test", "Use as a verb")
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
	if response != `{"word": "test", "meaning": "extra"}` {
		t.Errorf("Unexpected response: %s", response)
	}
}

func TestGenerateAdditionalTrainingCard_PromptNotSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err == nil {
		t.Error("Expected error when training prompt is not set")
	}
}

func TestGenerateResponse_RequestBuildError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("://bad", "test-model", "test-key", "test prompt", logger)
	_, err := service.GenerateResponse(context.Background(), "Hello")
	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
}

func TestGenerateResponse_ReadBodyError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       failingReadBody{},
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	_, err := service.GenerateResponse(context.Background(), "Hello")
	if err == nil {
		t.Error("Expected error when response body read fails")
	}
}

func TestGenerateResponse_SendError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	_, err := service.GenerateResponse(context.Background(), "Hello")
	if err == nil {
		t.Error("Expected error when transport fails")
	}
}

func TestGenerateTrainingCard_HTTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusServiceUnavailable, "Service Unavailable"), nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for HTTP 503")
	}
}

func TestGenerateTrainingCard_ModelOverrideEmptyString(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}
		if req.Model != "default-model" {
			t.Errorf("Expected default model when override is empty string, got %s", req.Model)
		}
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: `{"word": "test"}`}}},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("GenerateTrainingCard() error = %v", err)
	}
}

func TestGenerateTrainingCard_SendError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error when transport fails")
	}
}

func TestGenerateTrainingCard_ReadBodyError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       failingReadBody{},
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error when response body read fails")
	}
}

func TestGenerateTrainingCard_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusOK, "not valid json"), nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestGenerateTrainingCard_APIError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Error: &Error{Message: "rate limit exceeded", Type: "rate_limit"},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for API error in response")
	}
}

func TestGenerateTrainingCard_NoChoices(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{Choices: []Choice{}}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateTrainingCard(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for empty choices")
	}
}

func TestGenerateAdditionalTrainingCard_WithModelOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("Expected gpt-4, got %s", req.Model)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test"}`}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "constraints", "gpt-4")
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
}

func TestGenerateAdditionalTrainingCard_EmptyConstraints(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: `{"word": "test"}`}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
}

func TestGenerateAdditionalTrainingCard_CleansMarkdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Content: "```\n{\"word\": \"test\"}\n```"}},
			},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	response, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
	if response != `{"word": "test"}` {
		t.Errorf("Expected cleaned JSON, got: %s", response)
	}
}

func TestGenerateAdditionalTrainingCard_ModelOverrideEmptyString(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "default-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}
		if req.Model != "default-model" {
			t.Errorf("Expected default model when override is empty string, got %s", req.Model)
		}
		resp := ChatResponse{
			Choices: []Choice{{Message: Message{Content: `{"word": "test"}`}}},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "", "")
	if err != nil {
		t.Fatalf("GenerateAdditionalTrainingCard() error = %v", err)
	}
}

func TestGenerateAdditionalTrainingCard_SendError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "constraints")
	if err == nil {
		t.Error("Expected error when transport fails")
	}
}

func TestGenerateAdditionalTrainingCard_ReadBodyError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       failingReadBody{},
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err == nil {
		t.Error("Expected error when response body read fails")
	}
}

func TestGenerateAdditionalTrainingCard_HTTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusBadRequest, "Bad Request"), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "constraints")
	if err == nil {
		t.Error("Expected error for HTTP 400")
	}
}

func TestGenerateAdditionalTrainingCard_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return newTextResponse(http.StatusOK, "not valid json"), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestGenerateAdditionalTrainingCard_APIError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{
			Error: &Error{Message: "quota exceeded", Type: "insufficient_quota"},
		}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "constraints")
	if err == nil {
		t.Error("Expected error for API error in response")
	}
}

func TestGenerateAdditionalTrainingCard_NoChoices(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://example.com", "test-model", "test-key", "test prompt", logger)
	service.SetTrainingPrompt("Generate: ")
	service.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := ChatResponse{Choices: []Choice{}}
		return newJSONResponse(http.StatusOK, resp), nil
	})

	_, err := service.GenerateAdditionalTrainingCard(context.Background(), "test", "")
	if err == nil {
		t.Error("Expected error for empty choices")
	}
}

func TestNewService_ProcessesEscapedNewlines(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://test", "model", "key", "Hello\\nWorld", logger)

	if service.prompt != "Hello\nWorld" {
		t.Errorf("Expected processed newlines, got %q", service.prompt)
	}
}
