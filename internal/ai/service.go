package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"tgbot-skeleton/internal/netproxy"

	"go.uber.org/zap"
)

// DefaultHTTPTimeout is the AI HTTP client timeout when RequestTimeout is unset or invalid.
const DefaultHTTPTimeout = 30 * time.Second

// ParseHTTPTimeout parses a duration string (e.g. "120s", "3m", "2h"). Empty returns 0 (use default in NewServiceWithTimeout). Invalid returns 0.
func ParseHTTPTimeout(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

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
	client                        *http.Client
	url                           string
	model                         string
	apiKey                        string
	prompt                        string
	trainingPrompt                string
	dictionaryPrompts             map[string]string // course_code -> dictionary lookup prompt override
	trainingPrompts               map[string]string // course_code -> training card generation prompt override
	conversationPrompts           map[string]string // course_code -> legacy combined NPC prompt (deprecated)
	conversationQuestPrompts      map[string]string // course_code -> quest task evaluation prompt (Prompt A)
	conversationCorrectionPrompts map[string]string // course_code -> error correction prompt (Prompt B)
	conversationNPCPrompts        map[string]string // course_code -> in-character NPC reply prompt (Prompt C)
	conversationModel             string            // optional model override for NPC conversations ("" = use model)
	pictureQuestPrompts           map[string]string // course_code -> picture quest task evaluation prompt
	pictureLumiPrompts            map[string]string // course_code -> Lumi reply prompt for picture quests
	sentenceGenPrompts            map[string]string // course_code -> daily sentence-set generation prompt
	sentenceGradePrompts          map[string]string // course_code -> per-sentence grading prompt
	logger                        *zap.Logger
}

func (s *Service) openRouterHeaders(accept string) map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + s.apiKey,
	}
	if strings.TrimSpace(accept) != "" {
		headers["Accept"] = accept
	}
	return headers
}

func (s *Service) doOpenRouterJSON(ctx context.Context, endpoint string, body []byte, accept string) (*http.Response, error) {
	if _, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := netproxy.DoJSONWithRetry(ctx, s.client, http.MethodPost, endpoint, body, s.openRouterHeaders(accept))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

// SetConversationModel sets an optional model override used only by ConversationTurn (NPC
// role-play). Empty keeps the default model. Conversations benefit from a stronger
// instruction-following model than dictionary/training generation.
func (s *Service) SetConversationModel(model string) {
	s.conversationModel = strings.TrimSpace(model)
}

// NewService creates a new AI service with the default HTTP client timeout (30s).
func NewService(url, model, apiKey, prompt string, logger *zap.Logger) *Service {
	return NewServiceWithTimeout(url, model, apiKey, prompt, 0, logger)
}

// NewServiceWithTimeout creates a new AI service. httpTimeout <= 0 means DefaultHTTPTimeout.
func NewServiceWithTimeout(url, model, apiKey, prompt string, httpTimeout time.Duration, logger *zap.Logger) *Service {
	return NewServiceWithTimeoutAndSocks5Proxy(url, model, apiKey, prompt, httpTimeout, "", logger)
}

// NewServiceWithTimeoutAndSocks5Proxy creates a new AI service and optionally routes
// chat/completions requests through a SOCKS5 proxy.
func NewServiceWithTimeoutAndSocks5Proxy(url, model, apiKey, prompt string, httpTimeout time.Duration, socks5Proxy string, logger *zap.Logger) *Service {
	if httpTimeout <= 0 {
		httpTimeout = DefaultHTTPTimeout
	}
	// Process prompt to handle escaped newlines
	processedPrompt := strings.ReplaceAll(prompt, "\\n", "\n")
	httpClient, err := netproxy.NewHTTPClient(httpTimeout, socks5Proxy)
	if err != nil && logger != nil {
		logger.Warn("invalid AI SOCKS5 proxy, using direct OpenRouter connection", zap.String("proxy", socks5Proxy), zap.Error(err))
	}

	return &Service{
		client: httpClient,
		url:    url,
		model:  model,
		apiKey: apiKey,
		prompt: processedPrompt,
		logger: logger,
	}
}

// SetDictionaryPromptForCourse registers a dictionary-lookup system prompt override for a given course code.
// GenerateResponseForCourse uses it instead of the default prompt when the course matches.
func (s *Service) SetDictionaryPromptForCourse(courseCode, prompt string) {
	if s.dictionaryPrompts == nil {
		s.dictionaryPrompts = make(map[string]string)
	}
	s.dictionaryPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// SetTrainingPrompt sets the default training card generation prompt
func (s *Service) SetTrainingPrompt(prompt string) {
	s.trainingPrompt = prompt
}

// SetTrainingPromptForCourse registers a training card generation prompt override for a given course code.
// GenerateTrainingCardForCourse/GenerateAdditionalTrainingCardForCourse use it instead of the default
// training prompt when the course matches, so e.g. es_ru training cards aren't generated against the
// default English-only training prompt.
func (s *Service) SetTrainingPromptForCourse(courseCode, prompt string) {
	if s.trainingPrompts == nil {
		s.trainingPrompts = make(map[string]string)
	}
	s.trainingPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

func (s *Service) trainingPromptForCourse(courseCode string) string {
	if p, ok := s.trainingPrompts[courseCode]; ok && p != "" {
		return p
	}
	return s.trainingPrompt
}

// SetConversationPromptForCourse registers an NPC role-play system prompt for a given course code.
// ConversationTurn uses it (with runtime scenario details appended by the caller) so each course
// role-plays in its own target language.
func (s *Service) SetConversationPromptForCourse(courseCode, prompt string) {
	if s.conversationPrompts == nil {
		s.conversationPrompts = make(map[string]string)
	}
	s.conversationPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// ConversationPromptForCourse returns the registered legacy combined conversation prompt for a course.
func (s *Service) ConversationPromptForCourse(courseCode string) string {
	if p, ok := s.conversationPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// SetConversationQuestPromptForCourse registers the quest task evaluation prompt (Prompt A).
func (s *Service) SetConversationQuestPromptForCourse(courseCode, prompt string) {
	if s.conversationQuestPrompts == nil {
		s.conversationQuestPrompts = make(map[string]string)
	}
	s.conversationQuestPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// ConversationQuestPromptForCourse returns the quest evaluation base prompt for a course.
func (s *Service) ConversationQuestPromptForCourse(courseCode string) string {
	if p, ok := s.conversationQuestPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// SetConversationCorrectionPromptForCourse registers the error correction prompt (Prompt B).
func (s *Service) SetConversationCorrectionPromptForCourse(courseCode, prompt string) {
	if s.conversationCorrectionPrompts == nil {
		s.conversationCorrectionPrompts = make(map[string]string)
	}
	s.conversationCorrectionPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// ConversationCorrectionPromptForCourse returns the error correction base prompt for a course.
func (s *Service) ConversationCorrectionPromptForCourse(courseCode string) string {
	if p, ok := s.conversationCorrectionPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// SetConversationNPCPromptForCourse registers the in-character NPC reply prompt (Prompt C).
func (s *Service) SetConversationNPCPromptForCourse(courseCode, prompt string) {
	if s.conversationNPCPrompts == nil {
		s.conversationNPCPrompts = make(map[string]string)
	}
	s.conversationNPCPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// ConversationNPCPromptForCourse returns the NPC reply base prompt for a course.
func (s *Service) ConversationNPCPromptForCourse(courseCode string) string {
	if p, ok := s.conversationNPCPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// SetPictureQuestPromptForCourse registers the picture-quest task evaluation prompt.
func (s *Service) SetPictureQuestPromptForCourse(courseCode, prompt string) {
	if s.pictureQuestPrompts == nil {
		s.pictureQuestPrompts = make(map[string]string)
	}
	s.pictureQuestPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// PictureQuestPromptForCourse returns the picture-quest evaluation base prompt for a course.
func (s *Service) PictureQuestPromptForCourse(courseCode string) string {
	if p, ok := s.pictureQuestPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// SetPictureLumiPromptForCourse registers the Lumi reply prompt for picture description quests.
func (s *Service) SetPictureLumiPromptForCourse(courseCode, prompt string) {
	if s.pictureLumiPrompts == nil {
		s.pictureLumiPrompts = make(map[string]string)
	}
	s.pictureLumiPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// PictureLumiPromptForCourse returns the Lumi reply base prompt for a course.
func (s *Service) PictureLumiPromptForCourse(courseCode string) string {
	if p, ok := s.pictureLumiPrompts[courseCode]; ok && p != "" {
		return p
	}
	return ""
}

// HasPictureQuestPrompts reports whether the picture-quest prompts (plus the shared correction
// prompt) are registered for a course.
func (s *Service) HasPictureQuestPrompts(courseCode string) bool {
	return s.PictureQuestPromptForCourse(courseCode) != "" &&
		s.PictureLumiPromptForCourse(courseCode) != "" &&
		s.ConversationCorrectionPromptForCourse(courseCode) != ""
}

// HasSplitConversationPrompts reports whether all three split prompts are registered for a course.
func (s *Service) HasSplitConversationPrompts(courseCode string) bool {
	return s.ConversationQuestPromptForCourse(courseCode) != "" &&
		s.ConversationCorrectionPromptForCourse(courseCode) != "" &&
		s.ConversationNPCPromptForCourse(courseCode) != ""
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
	Usage   *Usage   `json:"usage,omitempty"`
	Error   *Error   `json:"error,omitempty"`
}

// Usage represents token accounting returned by the provider.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
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
	fields := append([]zap.Field{zap.String("url", s.url), zap.String("model", model)}, logFields...)
	s.logger.Debug("sending chat/completions", fields...)

	resp, err := s.doOpenRouterJSON(ctx, s.url+"/chat/completions", reqBody, "")
	if err != nil {
		return "", err
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

// ControlSentinel separates the visible NPC reply from the trailing task-completion control block.
const ControlSentinel = "###CONTROL###"

// ChatCorrection is a single mistake the model spotted in the learner's latest message,
// with a corrected version and a short explanation (in the learner's native language).
type ChatCorrection struct {
	Original    string `json:"original"`
	Corrected   string `json:"corrected"`
	Explanation string `json:"explanation"`
}

// chatControlSignal is the JSON the model appends after ControlSentinel on each NPC turn.
type chatControlSignal struct {
	CompletedTaskCodes []string         `json:"completed_task_codes"`
	AllDone            bool             `json:"all_done"`
	Corrections        []ChatCorrection `json:"corrections"`
}

// ChatTurnResult is one parsed NPC turn: the visible reply plus the (advisory) task-completion signal.
type ChatTurnResult struct {
	VisibleContent     string
	CompletedTaskCodes []string
	AllDoneSignal      bool
	Corrections        []ChatCorrection
	PromptTokens       int
	CompletionTokens   int
	Raw                string
}

// postChatCompletionRaw sends a chat/completions request and returns the assistant text UNMODIFIED
// (unlike postChatCompletion it does not strip markdown/JSON fences, which would corrupt a chat
// reply) along with token usage when the provider reports it.
func (s *Service) postChatCompletionRaw(ctx context.Context, model string, messages []Message, maxTokens int, temperature float64, logFields ...zap.Field) (string, *Usage, error) {
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	reqBody, err := jsonMarshalFunc(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	fields := append([]zap.Field{zap.String("url", s.url), zap.String("model", model)}, logFields...)
	s.logger.Debug("sending chat/completions (raw)", fields...)

	resp, err := s.doOpenRouterJSON(ctx, s.url+"/chat/completions", reqBody, "")
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("AI provider returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", nil, fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if chatResp.Error != nil {
		return "", nil, fmt.Errorf("AI provider error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response choices received")
	}
	return chatResp.Choices[0].Message.Content, chatResp.Usage, nil
}

// ConversationTurn sends the full running history plus the scenario system prompt and returns the
// parsed NPC turn. history holds prior user/assistant turns (oldest first); the system prompt is
// prepended and the new userMessage appended here. The trailing control block is split off so the
// visible reply never contains it.
func (s *Service) ConversationTurn(ctx context.Context, systemPrompt string, history []Message, userMessage string, maxTokens int, modelOverride ...string) (*ChatTurnResult, error) {
	model := s.model
	if s.conversationModel != "" {
		model = s.conversationModel
	}
	if len(modelOverride) > 0 && strings.TrimSpace(modelOverride[0]) != "" {
		model = modelOverride[0]
	}
	if maxTokens <= 0 {
		maxTokens = 600
	}

	msgs := make([]Message, 0, len(history)+2)
	msgs = append(msgs, Message{Role: "system", Content: systemPrompt})
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: userMessage})

	raw, usage, err := s.postChatCompletionRaw(ctx, model, msgs, maxTokens, 0.6, zap.String("kind", "conversation"))
	if err != nil {
		return nil, err
	}

	visible, signal := parseChatControl(raw)
	res := &ChatTurnResult{
		VisibleContent:     visible,
		CompletedTaskCodes: signal.CompletedTaskCodes,
		AllDoneSignal:      signal.AllDone,
		Corrections:        sanitizeCorrections(signal.Corrections),
		Raw:                raw,
	}
	if usage != nil {
		res.PromptTokens = usage.PromptTokens
		res.CompletionTokens = usage.CompletionTokens
	}
	return res, nil
}

// parseChatControl splits an NPC completion into the visible reply and the parsed control signal.
// A missing or malformed control block yields the full text as visible and an empty signal.
func parseChatControl(raw string) (string, chatControlSignal) {
	idx := strings.Index(raw, ControlSentinel)
	if idx < 0 {
		return strings.TrimSpace(raw), chatControlSignal{}
	}
	visible := strings.TrimSpace(raw[:idx])
	tail := stripLLMJSONFences(raw[idx+len(ControlSentinel):])
	var sig chatControlSignal
	if tail != "" {
		if err := json.Unmarshal([]byte(tail), &sig); err != nil {
			// Defensive: best-effort extraction of the first JSON object in the tail.
			if start := strings.Index(tail, "{"); start >= 0 {
				if end := strings.LastIndex(tail, "}"); end > start {
					_ = json.Unmarshal([]byte(tail[start:end+1]), &sig)
				}
			}
		}
	}
	return visible, sig
}

// sanitizeCorrections trims and drops empty/degenerate corrections so the client only ever
// receives meaningful entries (a corrected form that actually differs from the original).
func sanitizeCorrections(in []ChatCorrection) []ChatCorrection {
	if len(in) == 0 {
		return nil
	}
	out := make([]ChatCorrection, 0, len(in))
	for _, c := range in {
		c.Original = strings.TrimSpace(c.Original)
		c.Corrected = strings.TrimSpace(c.Corrected)
		c.Explanation = strings.TrimSpace(c.Explanation)
		if c.Corrected == "" {
			continue
		}
		if c.Original != "" && strings.EqualFold(c.Original, c.Corrected) {
			continue
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
	return s.generateResponseWithPrompt(ctx, userMessage, s.prompt)
}

// GenerateResponseForCourse behaves like GenerateResponse but uses the dictionary prompt
// registered for courseCode via SetDictionaryPromptForCourse, if any, falling back to the
// default prompt otherwise. Use this for word-card generation so courses in a different
// target language (e.g. Spanish) aren't validated against the default (English) prompt.
func (s *Service) GenerateResponseForCourse(ctx context.Context, userMessage, courseCode string) (string, error) {
	prompt := s.prompt
	if p, ok := s.dictionaryPrompts[courseCode]; ok && p != "" {
		prompt = p
	}
	return s.generateResponseWithPrompt(ctx, userMessage, prompt)
}

func (s *Service) generateResponseWithPrompt(ctx context.Context, userMessage, systemPrompt string) (string, error) {
	if isSingleWordLookupCandidate(userMessage) {
		userMessage = "SINGLE_WORD_LOOKUP_MODE\nReturn ONLY one JSON object for dictionary lookup.\nWord: " + strings.TrimSpace(userMessage)
	}
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return s.postChatCompletion(ctx, s.model, messages, 2000, 0.3, zap.String("kind", "dictionary"), zap.String("user_message", userMessage))
}

// GenerateTrainingCard generates a training card for a word using LLM, using the default training prompt.
// If modelOverride is provided, it will be used instead of the default model
func (s *Service) GenerateTrainingCard(ctx context.Context, word string, modelOverride ...string) (string, error) {
	return s.GenerateTrainingCardForCourse(ctx, word, "", modelOverride...)
}

// GenerateTrainingCardForCourse behaves like GenerateTrainingCard but uses the training prompt
// registered for courseCode via SetTrainingPromptForCourse, if any, falling back to the default
// training prompt otherwise. Use this so courses in a different target language (e.g. Spanish)
// aren't validated against the default (English) training prompt.
func (s *Service) GenerateTrainingCardForCourse(ctx context.Context, word, courseCode string, modelOverride ...string) (string, error) {
	return s.generateTrainingCardForCourse(ctx, word, courseCode, "", modelOverride...)
}

// GenerateTrainingCardForCourseWithContext behaves like GenerateTrainingCardForCourse but inserts
// contextBlock (e.g. native lookup hint) between the training prompt and the target word token.
func (s *Service) GenerateTrainingCardForCourseWithContext(ctx context.Context, word, courseCode, contextBlock string, modelOverride ...string) (string, error) {
	return s.generateTrainingCardForCourse(ctx, word, courseCode, contextBlock, modelOverride...)
}

func (s *Service) generateTrainingCardForCourse(ctx context.Context, word, courseCode, contextBlock string, modelOverride ...string) (string, error) {
	trainingPrompt := s.trainingPromptForCourse(courseCode)
	if trainingPrompt == "" {
		return "", fmt.Errorf("training prompt not set")
	}

	// Prepare user message with word
	userMessage := strings.TrimSpace(trainingPrompt)
	if !strings.HasSuffix(userMessage, "\n") {
		userMessage += "\n"
	}
	if strings.TrimSpace(contextBlock) != "" {
		userMessage += strings.TrimSpace(contextBlock) + "\n\n"
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

	s.logger.Debug("sending training card generation request",
		zap.String("word", word),
		zap.String("model", model),
	)

	// Send request
	resp, err := s.doOpenRouterJSON(ctx, s.url+"/chat/completions", reqBody, "")
	if err != nil {
		return "", err
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

// GenerateAdditionalTrainingCard generates an additional training card for a word with constraints,
// using the default training prompt. constraints can specify things like specific meaning, part of speech, etc.
func (s *Service) GenerateAdditionalTrainingCard(ctx context.Context, word string, constraints string, modelOverride ...string) (string, error) {
	return s.GenerateAdditionalTrainingCardForCourse(ctx, word, "", constraints, modelOverride...)
}

// GenerateAdditionalTrainingCardForCourse behaves like GenerateAdditionalTrainingCard but uses the
// training prompt registered for courseCode via SetTrainingPromptForCourse, if any.
func (s *Service) GenerateAdditionalTrainingCardForCourse(ctx context.Context, word, courseCode string, constraints string, modelOverride ...string) (string, error) {
	trainingPrompt := s.trainingPromptForCourse(courseCode)
	if trainingPrompt == "" {
		return "", fmt.Errorf("training prompt not set")
	}

	// Build user message with word and constraints
	var userMessage strings.Builder
	userMessage.WriteString(strings.TrimSpace(trainingPrompt))
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

	s.logger.Debug("sending additional training card generation request",
		zap.String("word", word),
		zap.String("constraints", constraints),
		zap.String("model", model),
	)

	// Send request
	resp, err := s.doOpenRouterJSON(ctx, s.url+"/chat/completions", reqBody, "")
	if err != nil {
		return "", err
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
