package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
)

// jsonMarshalFunc is used for request marshaling; overridable in tests for coverage.
var jsonMarshalFunc = json.Marshal

// stripLLMJSONFences removes markdown code fences often wrapped around JSON chat output.
func stripLLMJSONFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func isSingleWordLookupCandidate(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	if len(strings.Fields(trimmed)) != 1 {
		return false
	}
	hasLatin := false
	for _, r := range trimmed {
		if unicode.IsDigit(r) || unicode.Is(unicode.Cyrillic, r) {
			return false
		}
		if unicode.Is(unicode.Latin, r) {
			hasLatin = true
		}
	}
	return hasLatin
}

// Service handles AI provider interactions
type Service struct {
	client         *http.Client
	url            string
	model          string
	apiKey         string
	prompt         string
	trainingPrompt string
	logger         *zap.Logger
}

// NewService creates a new AI service
func NewService(url, model, apiKey, prompt string, logger *zap.Logger) *Service {
	// Process prompt to handle escaped newlines
	processedPrompt := strings.ReplaceAll(prompt, "\\n", "\n")

	return &Service{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		url:    url,
		model:  model,
		apiKey: apiKey,
		prompt: processedPrompt,
		logger: logger,
	}
}

// SetTrainingPrompt sets the training card generation prompt
func (s *Service) SetTrainingPrompt(prompt string) {
	s.trainingPrompt = prompt
}

// ChatRequest represents the OpenAI-compatible chat request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the OpenAI-compatible chat response
type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Error   *Error   `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// Error represents an API error
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// postChatCompletion sends a chat/completions request and returns the first assistant text.
func (s *Service) postChatCompletion(ctx context.Context, model string, messages []Message, maxTokens int, temperature float64, logFields ...zap.Field) (string, error) {
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	reqBody, err := jsonMarshalFunc(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.url+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	fields := append([]zap.Field{zap.String("url", s.url), zap.String("model", model)}, logFields...)
	s.logger.Debug("sending chat/completions", fields...)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("AI provider returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("AI provider error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}
	out := stripLLMJSONFences(chatResp.Choices[0].Message.Content)
	s.logger.Debug("received chat/completions response", zap.Int("length", len(out)))
	return out, nil
}

// ChatSystemUser sends an explicit system+user chat (no dictionary heuristics on user text).
func (s *Service) ChatSystemUser(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	msgs := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return s.postChatCompletion(ctx, s.model, msgs, 1600, 0.15, zap.String("kind", "system_user"))
}

// GenerateResponse sends a message to the AI provider and returns the response
func (s *Service) GenerateResponse(ctx context.Context, userMessage string) (string, error) {
	if isSingleWordLookupCandidate(userMessage) {
		userMessage = "SINGLE_WORD_LOOKUP_MODE\nReturn ONLY one JSON object for dictionary lookup.\nWord: " + strings.TrimSpace(userMessage)
	}
	messages := []Message{
		{Role: "system", Content: s.prompt},
		{Role: "user", Content: userMessage},
	}
	return s.postChatCompletion(ctx, s.model, messages, 2000, 0.3, zap.String("kind", "dictionary"), zap.String("user_message", userMessage))
}

// GenerateTrainingCard generates a training card for a word using LLM
// If modelOverride is provided, it will be used instead of the default model
func (s *Service) GenerateTrainingCard(ctx context.Context, word string, modelOverride ...string) (string, error) {
	if s.trainingPrompt == "" {
		return "", fmt.Errorf("training prompt not set")
	}

	// Prepare user message with word
	userMessage := strings.TrimSpace(s.trainingPrompt)
	if !strings.HasSuffix(userMessage, "\n") {
		userMessage += "\n"
	}
	userMessage += strings.TrimSpace(word)

	// Prepare messages
	messages := []Message{
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Use model override if provided, otherwise use default
	model := s.model
	if len(modelOverride) > 0 && modelOverride[0] != "" {
		model = modelOverride[0]
	}

	// Create request
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.3, // Balanced determinism; avoids overfitting to wrong reject branch
	}

	// Marshal request
	reqBody, err := jsonMarshalFunc(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.url+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	s.logger.Debug("sending training card generation request",
		zap.String("word", word),
		zap.String("model", model),
	)

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("AI provider returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if chatResp.Error != nil {
		return "", fmt.Errorf("AI provider error: %s", chatResp.Error.Message)
	}

	// Check if we have choices
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}

	response := stripLLMJSONFences(chatResp.Choices[0].Message.Content)

	s.logger.Debug("received training card response",
		zap.String("word", word),
		zap.Int("length", len(response)),
	)

	return response, nil
}

// GenerateAdditionalTrainingCard generates an additional training card for a word with constraints
// constraints can specify things like specific meaning, part of speech, etc.
func (s *Service) GenerateAdditionalTrainingCard(ctx context.Context, word string, constraints string, modelOverride ...string) (string, error) {
	if s.trainingPrompt == "" {
		return "", fmt.Errorf("training prompt not set")
	}

	// Build user message with word and constraints
	var userMessage strings.Builder
	userMessage.WriteString(strings.TrimSpace(s.trainingPrompt))
	userMessage.WriteString("\n")
	userMessage.WriteString(strings.TrimSpace(word))

	if constraints != "" {
		userMessage.WriteString("\n\nAdditional constraints for this card:\n")
		userMessage.WriteString(constraints)
		userMessage.WriteString("\n\nGenerate ONE training card that matches these constraints. Return the same JSON format with exactly ONE sense in the senses array.")
	}

	// Prepare messages
	messages := []Message{
		{
			Role:    "user",
			Content: userMessage.String(),
		},
	}

	// Use model override if provided, otherwise use default
	model := s.model
	if len(modelOverride) > 0 && modelOverride[0] != "" {
		model = modelOverride[0]
	}

	// Create request
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.3, // Balanced determinism; avoids overfitting to wrong reject branch
	}

	// Marshal request
	reqBody, err := jsonMarshalFunc(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.url+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	s.logger.Debug("sending additional training card generation request",
		zap.String("word", word),
		zap.String("constraints", constraints),
		zap.String("model", model),
	)

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("AI provider returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if chatResp.Error != nil {
		return "", fmt.Errorf("AI provider error: %s", chatResp.Error.Message)
	}

	// Check if we have choices
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}

	response := stripLLMJSONFences(chatResp.Choices[0].Message.Content)

	s.logger.Debug("received additional training card response",
		zap.String("word", word),
		zap.Int("length", len(response)),
	)

	return response, nil
}
