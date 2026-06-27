package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// ConversationTurnSplitInput configures the three dedicated LLM calls for one NPC turn.
type ConversationTurnSplitInput struct {
	QuestPrompt         string // system prompt for quest evaluation (Prompt A); empty skips call
	CorrectionPrompt    string // system prompt for error correction (Prompt B); empty skips call
	NPCPrompt           string // system prompt for in-character reply (Prompt C)
	History             []Message
	UserMessage         string
	MaxTokens           int // max tokens for NPC reply
	EvaluateQuest       bool
	EvaluateCorrections bool
}

type questEvalSignal struct {
	CompletedTaskCodes []string `json:"completed_task_codes"`
	AllDone            bool     `json:"all_done"`
}

type correctionEvalSignal struct {
	Corrections []ChatCorrection `json:"corrections"`
}

// ConversationTurnSplit runs up to three parallel LLM calls (quest eval, correction, NPC reply)
// and merges the results into a single ChatTurnResult.
func (s *Service) ConversationTurnSplit(ctx context.Context, in ConversationTurnSplitInput, modelOverride ...string) (*ChatTurnResult, error) {
	if strings.TrimSpace(in.NPCPrompt) == "" {
		return nil, fmt.Errorf("npc prompt is required")
	}
	model := s.model
	if s.conversationModel != "" {
		model = s.conversationModel
	}
	if len(modelOverride) > 0 && strings.TrimSpace(modelOverride[0]) != "" {
		model = modelOverride[0]
	}
	maxTokens := in.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 600
	}

	var (
		wg              sync.WaitGroup
		mu              sync.Mutex
		firstErr        error
		questSig        questEvalSignal
		corrections     []ChatCorrection
		npcReply        string
		promptTokens    int
		completionTokens int
		rawParts        []string
	)

	recordErr := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}
	addUsage := func(usage *Usage) {
		if usage == nil {
			return
		}
		mu.Lock()
		promptTokens += usage.PromptTokens
		completionTokens += usage.CompletionTokens
		mu.Unlock()
	}
	appendRaw := func(label, raw string) {
		mu.Lock()
		rawParts = append(rawParts, label+":\n"+raw)
		mu.Unlock()
	}

	if in.EvaluateQuest && strings.TrimSpace(in.QuestPrompt) != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgs := []Message{
				{Role: "system", Content: in.QuestPrompt},
				{Role: "user", Content: in.UserMessage},
			}
			raw, usage, err := s.postChatCompletionRaw(ctx, model, msgs, 200, 0.2, zap.String("kind", "conversation-quest"))
			addUsage(usage)
			appendRaw("quest", raw)
			if err != nil {
				recordErr(fmt.Errorf("quest evaluation: %w", err))
				return
			}
			sig, err := parseQuestEvalJSON(raw)
			if err != nil {
				recordErr(fmt.Errorf("quest evaluation parse: %w", err))
				return
			}
			mu.Lock()
			questSig = sig
			mu.Unlock()
		}()
	}

	if in.EvaluateCorrections && strings.TrimSpace(in.CorrectionPrompt) != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgs := []Message{
				{Role: "system", Content: in.CorrectionPrompt},
				{Role: "user", Content: in.UserMessage},
			}
			raw, usage, err := s.postChatCompletionRaw(ctx, model, msgs, 200, 0.2, zap.String("kind", "conversation-correction"))
			addUsage(usage)
			appendRaw("correction", raw)
			if err != nil {
				recordErr(fmt.Errorf("correction evaluation: %w", err))
				return
			}
			sig, err := parseCorrectionEvalJSON(raw)
			if err != nil {
				recordErr(fmt.Errorf("correction evaluation parse: %w", err))
				return
			}
			mu.Lock()
			corrections = sanitizeCorrections(sig.Corrections)
			mu.Unlock()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		msgs := make([]Message, 0, len(in.History)+2)
		msgs = append(msgs, Message{Role: "system", Content: in.NPCPrompt})
		msgs = append(msgs, in.History...)
		msgs = append(msgs, Message{Role: "user", Content: in.UserMessage})
		raw, usage, err := s.postChatCompletionRaw(ctx, model, msgs, maxTokens, 0.6, zap.String("kind", "conversation-npc"))
		addUsage(usage)
		appendRaw("npc", raw)
		if err != nil {
			recordErr(fmt.Errorf("npc reply: %w", err))
			return
		}
		mu.Lock()
		npcReply = sanitizeNPCReply(raw)
		mu.Unlock()
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return &ChatTurnResult{
		VisibleContent:     npcReply,
		CompletedTaskCodes: questSig.CompletedTaskCodes,
		AllDoneSignal:      questSig.AllDone,
		Corrections:        corrections,
		PromptTokens:       promptTokens,
		CompletionTokens:   completionTokens,
		Raw:                strings.Join(rawParts, "\n---\n"),
	}, nil
}

func parseQuestEvalJSON(raw string) (questEvalSignal, error) {
	var sig questEvalSignal
	if err := unmarshalLLMJSON(raw, &sig); err != nil {
		return sig, err
	}
	return sig, nil
}

func parseCorrectionEvalJSON(raw string) (correctionEvalSignal, error) {
	var sig correctionEvalSignal
	if err := unmarshalLLMJSON(raw, &sig); err != nil {
		return sig, err
	}
	return sig, nil
}

// unmarshalLLMJSON parses JSON from a model response, tolerating markdown fences and surrounding text.
func unmarshalLLMJSON(raw string, dest any) error {
	raw = stripLLMJSONFences(strings.TrimSpace(raw))
	if raw == "" {
		return fmt.Errorf("empty response")
	}
	if err := json.Unmarshal([]byte(raw), dest); err == nil {
		return nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("no JSON object found")
	}
	return json.Unmarshal([]byte(raw[start:end+1]), dest)
}

// sanitizeNPCReply strips accidental control blocks or JSON from NPC-only responses.
func sanitizeNPCReply(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, ControlSentinel); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}
	return stripLLMJSONFences(raw)
}
