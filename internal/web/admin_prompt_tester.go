package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"

	"go.uber.org/zap"
)

// handleAdminPromptTesterDefaultPrompts returns default prompts from config
// @Summary      Получить дефолтные промпты
// @Description  Возвращает дефолтные промпты для word data и training cards из конфигурации
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Дефолтные промпты"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Router       /api/admin/prompt-tester/default-prompts [get]
func (r *Router) handleAdminPromptTesterDefaultPrompts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get word prompt (rendered the same way as at startup)
	var wordPrompt string
	var wordPromptSource string
	if r.config.AI.PromptFile != "" {
		rendered, err := ai.LoadRenderedPromptFile(r.config.AI.PromptFile, r.config.Learning.NativeLang, r.config.Learning.TargetLang, r.config.Learning.Pair)
		if err != nil {
			r.logger.Warn("failed to read AI prompt file, falling back to configured AI_PROMPT",
				zap.String("file", r.config.AI.PromptFile),
				zap.Error(err),
			)
			wordPrompt = r.config.AI.Prompt
			wordPromptSource = "env"
		} else {
			wordPrompt = rendered
			wordPromptSource = r.config.AI.PromptFile
		}
	} else {
		wordPrompt = r.config.AI.Prompt
		wordPromptSource = "env"
	}

	// Get training prompt
	var trainingPrompt string
	var trainingPromptSource string
	trainingPromptFile := r.config.Training.PromptFile
	if trainingPromptFile == "" {
		trainingPromptFile = "prompts/training-card-ru-en.txt"
	}

	content, err := ai.LoadRenderedPromptFile(trainingPromptFile, r.config.Learning.NativeLang, r.config.Learning.TargetLang, r.config.Learning.Pair)
	if err != nil {
		r.logger.Error("failed to read training prompt file",
			zap.String("file", trainingPromptFile),
			zap.Error(err),
		)
		trainingPrompt = ""
		trainingPromptSource = trainingPromptFile + " (not found)"
	} else {
		trainingPrompt = content
		trainingPromptSource = trainingPromptFile
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word_prompt":            wordPrompt,
		"word_prompt_source":     wordPromptSource,
		"training_prompt":        trainingPrompt,
		"training_prompt_source": trainingPromptSource,
	})
}

// PromptTesterRunRequest represents the request body for running prompt tests
type PromptTesterRunRequest struct {
	Words          []string `json:"words"`
	WordPrompt     string   `json:"word_prompt"`
	TrainingPrompt string   `json:"training_prompt"`
}

// PromptTesterEvent represents a single event in the NDJSON stream
type PromptTesterEvent struct {
	Word       string                 `json:"word"`
	Step       string                 `json:"step"` // "word" or "cards"
	OK         bool                   `json:"ok"`
	Raw        string                 `json:"raw,omitempty"`
	Parsed     map[string]interface{} `json:"parsed,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMS int64                  `json:"duration_ms"`
}

// handleAdminPromptTesterRun runs prompt tests and streams results as NDJSON
// @Summary      Запустить тестирование промптов
// @Description  Прогоняет список слов через LLM с указанными промптами и возвращает результаты в формате NDJSON stream
// @Tags         Admin
// @Accept       json
// @Produce      application/x-ndjson
// @Security     ApiKeyAuth
// @Param        request  body  PromptTesterRunRequest  true  "Запрос с словами и промптами"
// @Success      200  {string}  string  "NDJSON stream событий"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Router       /api/admin/prompt-tester/run [post]
func (r *Router) handleAdminPromptTesterRun(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var runReq PromptTesterRunRequest
	if err := json.NewDecoder(req.Body).Decode(&runReq); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(runReq.Words) == 0 {
		http.Error(w, "words array is required and must not be empty", http.StatusBadRequest)
		return
	}

	if runReq.WordPrompt == "" {
		http.Error(w, "word_prompt is required", http.StatusBadRequest)
		return
	}

	if runReq.TrainingPrompt == "" {
		http.Error(w, "training_prompt is required", http.StatusBadRequest)
		return
	}

	// Set up streaming response
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Flush headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	ctx := req.Context()

	// Create temporary AI service with custom prompts
	aiService := ai.NewServiceWithTimeout(
		r.config.AI.URL,
		r.config.AI.Model,
		r.config.AI.APIKey,
		runReq.WordPrompt,
		ai.ParseHTTPTimeout(r.config.AI.RequestTimeout),
		r.logger,
	)
	aiService.SetTrainingPrompt(runReq.TrainingPrompt)

	// Process each word
	for _, word := range runReq.Words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Step 1: Generate word data
		wordStart := time.Now()
		wordResponse, wordErr := r.generateWordData(ctx, aiService, word)
		wordDuration := time.Since(wordStart).Milliseconds()

		wordEvent := PromptTesterEvent{
			Word:       word,
			Step:       "word",
			OK:         wordErr == nil,
			DurationMS: wordDuration,
		}

		if wordErr != nil {
			wordEvent.Error = wordErr.Error()
			wordEvent.Raw = ""
		} else {
			// Try to parse as JSON map; fall back to raw string if not a map.
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(wordResponse), &parsed); err == nil {
				wordEvent.Parsed = parsed
				formattedJSON, _ := json.MarshalIndent(parsed, "", "  ")
				wordEvent.Raw = string(formattedJSON)
			} else {
				wordEvent.Raw = wordResponse
			}
		}

		// Send word event
		if err := r.sendNDJSONEvent(w, wordEvent); err != nil {
			r.logger.Error("failed to send word event", zap.Error(err))
			return
		}

		// Step 2: Generate training cards
		cardsStart := time.Now()
		cardsResponse, cardsErr := r.generateTrainingCards(ctx, aiService, word)
		cardsDuration := time.Since(cardsStart).Milliseconds()

		cardsEvent := PromptTesterEvent{
			Word:       word,
			Step:       "cards",
			OK:         cardsErr == nil,
			DurationMS: cardsDuration,
		}

		if cardsErr != nil {
			cardsEvent.Error = cardsErr.Error()
			cardsEvent.Raw = ""
		} else {
			// Try to parse as JSON map; fall back to raw string if not a map.
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(cardsResponse), &parsed); err == nil {
				cardsEvent.Parsed = parsed
				formattedJSON, _ := json.MarshalIndent(parsed, "", "  ")
				cardsEvent.Raw = string(formattedJSON)
			} else {
				cardsEvent.Raw = cardsResponse
			}
		}

		// Send cards event
		if err := r.sendNDJSONEvent(w, cardsEvent); err != nil {
			r.logger.Error("failed to send cards event", zap.Error(err))
			return
		}
	}
}

// generateWordData generates word data using AI service
func (r *Router) generateWordData(ctx context.Context, aiService *ai.Service, word string) (string, error) {
	response, err := aiService.GenerateResponse(ctx, word)
	if err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}

	// Clean up response - remove markdown code blocks if present (similar to GenerateTrainingCard)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	return response, nil
}

// generateTrainingCards generates training cards using AI service
func (r *Router) generateTrainingCards(ctx context.Context, aiService *ai.Service, word string) (string, error) {
	response, err := aiService.GenerateTrainingCard(ctx, word)
	if err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}
	return response, nil
}

// sendNDJSONEvent sends a single NDJSON event to the response writer
func (r *Router) sendNDJSONEvent(w http.ResponseWriter, event PromptTesterEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = w.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// Flush if possible
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}
