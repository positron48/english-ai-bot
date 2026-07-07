package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/netproxy"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// SpeakingEvaluationResult is structured feedback from the multimodal evaluator.
type SpeakingEvaluationResult struct {
	UnderstoodAnswer   string `json:"understood_answer"`
	MeaningScore       int    `json:"meaning_score"`
	GrammarScore       int    `json:"grammar_score"`
	PronunciationScore int    `json:"pronunciation_score"`
	FluencyScore       int    `json:"fluency_score"`
	IsAcceptable       bool   `json:"is_acceptable"`
	AudioQuality       string `json:"audio_quality"`
	ShortFeedbackRU    string `json:"short_feedback_ru"`
	BetterVersion      string `json:"better_version"`
	RepeatTask         string `json:"repeat_task"`
}

// SpeakingEvaluatorService calls OpenRouter with audio input for speaking assessment.
type SpeakingEvaluatorService struct {
	cfg    config.SpeakingConfig
	client *http.Client
	logger *zap.Logger
}

func NewSpeakingEvaluatorService(cfg *config.Config, logger *zap.Logger) *SpeakingEvaluatorService {
	if cfg == nil {
		return &SpeakingEvaluatorService{logger: logger}
	}
	timeout := 60 * time.Second
	if d, err := time.ParseDuration(strings.TrimSpace(cfg.Speaking.EvalTimeout)); err == nil && d > 0 {
		timeout = d
	}
	httpClient, proxyErr := netproxy.NewHTTPClient(timeout, cfg.Speaking.Socks5Proxy)
	if proxyErr != nil && logger != nil {
		logger.Warn("invalid speaking/OpenRouter SOCKS5 proxy, using direct connection",
			zap.String("proxy", cfg.Speaking.Socks5Proxy),
			zap.Error(proxyErr),
		)
	}
	return &SpeakingEvaluatorService{
		cfg:    cfg.Speaking,
		client: httpClient,
		logger: logger,
	}
}

func (s *SpeakingEvaluatorService) Enabled() bool {
	return s != nil && s.cfg.Enabled && strings.TrimSpace(s.cfg.EvalAPIKey) != ""
}

func (s *SpeakingEvaluatorService) AcceptMeaningScore() int {
	if s == nil || s.cfg.AcceptMeaningScore <= 0 {
		return 3
	}
	return s.cfg.AcceptMeaningScore
}

func (s *SpeakingEvaluatorService) Evaluate(
	ctx context.Context,
	task *repository.SpeakingTaskFull,
	audio []byte,
	audioFormat string,
	attemptNo int,
	mode string,
) (*SpeakingEvaluationResult, error) {
	if s == nil || !s.Enabled() {
		return nil, fmt.Errorf("speaking evaluator is not enabled")
	}
	if len(audio) == 0 {
		return nil, fmt.Errorf("empty audio")
	}
	maxBytes := s.cfg.MaxAudioMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024
	}
	if len(audio) > maxBytes {
		return nil, fmt.Errorf("audio too large")
	}

	format := normalizeAudioFormat(audioFormat)
	prepared, apiFormat, err := prepareSpeakingAudioForModel(ctx, audio, format, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare audio: %w", err)
	}
	if len(prepared) > maxBytes {
		return nil, fmt.Errorf("audio too large after conversion")
	}

	prompt := buildSpeakingEvalPrompt(task, attemptNo, mode, s.AcceptMeaningScore())

	result, err := s.callModel(ctx, prompt, prepared, apiFormat)
	if err != nil {
		return nil, err
	}
	normalizeSpeakingResult(result, s.AcceptMeaningScore())
	return result, nil
}

func normalizeAudioFormat(f string) string {
	f = strings.ToLower(strings.TrimSpace(f))
	switch f {
	case "wav", "mp3", "webm", "ogg", "m4a":
		return f
	default:
		return "webm"
	}
}

func buildSpeakingEvalPrompt(task *repository.SpeakingTaskFull, attemptNo int, mode string, acceptScore int) string {
	var b strings.Builder
	b.WriteString("You are a Spanish speaking tutor for Russian-speaking learners.\n")
	b.WriteString("Listen to the user's audio and evaluate their spoken response.\n")
	b.WriteString("Return ONLY valid JSON with keys: understood_answer, meaning_score, grammar_score, pronunciation_score, fluency_score, is_acceptable, audio_quality, short_feedback_ru, better_version, repeat_task.\n")
	b.WriteString("Scores are integers 1-5. audio_quality is clear or unclear.\n")
	b.WriteString("Evaluate MEANING, not exact string match. Always fill understood_answer with what you heard.\n")
	fmt.Fprintf(&b, "Mark is_acceptable true when meaning_score>=%d and audio is clear enough.\n", acceptScore)
	b.WriteString("short_feedback_ru must be in Russian, concise and encouraging.\n")
	fmt.Fprintf(&b, "Attempt: %d, mode: %s\n", attemptNo, mode)
	fmt.Fprintf(&b, "Task type: %s, level: %s, target language: %s\n", task.Type, task.Level, task.TargetLanguage)
	if task.PromptRU != "" {
		b.WriteString("Instruction RU: " + task.PromptRU + "\n")
	}
	if task.DisplayText != "" {
		b.WriteString("Expected phrase to say: " + task.DisplayText + "\n")
	}
	if task.ExpectedMeaningRU != "" {
		b.WriteString("Expected meaning RU: " + task.ExpectedMeaningRU + "\n")
	}
	if len(task.AcceptableAnswers) > 0 {
		b.WriteString("Acceptable variants: " + strings.Join(task.AcceptableAnswers, " | ") + "\n")
	}
	if task.EvaluationNotes != "" {
		b.WriteString("Evaluator notes: " + task.EvaluationNotes + "\n")
	}
	return b.String()
}

func (s *SpeakingEvaluatorService) callModel(ctx context.Context, prompt string, audio []byte, format string) (*SpeakingEvaluationResult, error) {
	b64 := base64.StdEncoding.EncodeToString(audio)
	payload := map[string]interface{}{
		"model": s.cfg.EvalModel,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": prompt},
					{
						"type": "input_audio",
						"input_audio": map[string]string{
							"data":   b64,
							"format": format,
						},
					},
				},
			},
		},
		"temperature": 0.2,
		"max_tokens":  800,
	}
	body, _ := json.Marshal(payload)
	endpoint := strings.TrimRight(s.cfg.EvalBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(s.cfg.EvalAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("speaking eval request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("speaking eval status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("speaking eval decode: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("speaking eval: no choices")
	}
	raw := stripSpeakingJSON(chatResp.Choices[0].Message.Content)
	var result SpeakingEvaluationResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("speaking eval json parse: %w (raw=%q)", err, raw)
	}
	return &result, nil
}

func stripSpeakingJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 {
			start := 1
			end := len(lines)
			if strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				end = len(lines) - 1
			}
			s = strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		}
	}
	return s
}

func normalizeSpeakingResult(r *SpeakingEvaluationResult, acceptScore int) {
	if r == nil {
		return
	}
	clamp := func(v int) int {
		if v < 1 {
			return 1
		}
		if v > 5 {
			return 5
		}
		return v
	}
	r.MeaningScore = clamp(r.MeaningScore)
	r.GrammarScore = clamp(r.GrammarScore)
	r.PronunciationScore = clamp(r.PronunciationScore)
	r.FluencyScore = clamp(r.FluencyScore)
	if strings.EqualFold(strings.TrimSpace(r.AudioQuality), "unclear") {
		r.IsAcceptable = false
		return
	}
	if r.MeaningScore >= acceptScore {
		r.IsAcceptable = true
	}
}
