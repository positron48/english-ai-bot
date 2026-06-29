package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// GenSentenceWord is one well-learned word offered to the generator.
type GenSentenceWord struct {
	Lemma       string `json:"lemma"`
	Translation string `json:"translation"`
}

// GeneratedSentence is one Russian sentence the model produced, with the suggested
// correct target translation and the lemmas it actually used.
type GeneratedSentence struct {
	PromptRU    string   `json:"prompt_ru"`
	ReferenceES string   `json:"reference_es"`
	UsedWords   []string `json:"used_words"`
}

type generatedSentenceSet struct {
	Sentences []GeneratedSentence `json:"sentences"`
}

// SentenceGradeToken is one rendered token of teacher-style markup.
// Status: "ok" (correct), "wrong" (struck out, Correction shown above),
// "insert" (a missing word the learner should have written, shown as Correction).
type SentenceGradeToken struct {
	Text       string `json:"text"`
	Status     string `json:"status"`
	Correction string `json:"correction,omitempty"`
}

// SentenceGrade is the parsed grading of a single learner submission.
type SentenceGrade struct {
	ErrorCount  int                  `json:"error_count"`
	Outcome     string               `json:"outcome"`
	CorrectedES string               `json:"corrected_es"`
	Tokens      []SentenceGradeToken `json:"tokens"`
}

// SetSentenceGenPromptForCourse registers the daily sentence-set generation prompt for a course.
func (s *Service) SetSentenceGenPromptForCourse(courseCode, prompt string) {
	if s.sentenceGenPrompts == nil {
		s.sentenceGenPrompts = make(map[string]string)
	}
	s.sentenceGenPrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// SetSentenceGradePromptForCourse registers the per-sentence grading prompt for a course.
func (s *Service) SetSentenceGradePromptForCourse(courseCode, prompt string) {
	if s.sentenceGradePrompts == nil {
		s.sentenceGradePrompts = make(map[string]string)
	}
	s.sentenceGradePrompts[courseCode] = strings.ReplaceAll(prompt, "\\n", "\n")
}

// HasSentencePromptsForCourse reports whether both sentence prompts are registered for a course.
func (s *Service) HasSentencePromptsForCourse(courseCode string) bool {
	return s.sentenceGenPrompts[courseCode] != "" && s.sentenceGradePrompts[courseCode] != ""
}

// GenerateSentenceSetForCourse asks the model to build `count` Russian sentences from the
// given well-learned words, constrained to the provided grammar tenses (human-readable).
func (s *Service) GenerateSentenceSetForCourse(ctx context.Context, courseCode string, words []GenSentenceWord, tenses []string, count int, modelOverride ...string) ([]GeneratedSentence, error) {
	prompt := s.sentenceGenPrompts[courseCode]
	if prompt == "" {
		return nil, fmt.Errorf("sentence generation prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"sentence_count": count,
		"allowed_tenses": tenses,
		"words":          words,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal generation payload: %w", err)
	}

	messages := []Message{
		{Role: "system", Content: prompt},
		{Role: "user", Content: string(payloadJSON)},
	}
	model := s.modelOr(modelOverride...)
	raw, err := s.postChatCompletion(ctx, model, messages, 3000, 0.6, zap.String("kind", "sentence_gen"), zap.String("course", courseCode))
	if err != nil {
		return nil, err
	}
	var set generatedSentenceSet
	if err := json.Unmarshal([]byte(raw), &set); err != nil {
		return nil, fmt.Errorf("parse generated sentences: %w (raw: %s)", err, truncateForLog(raw))
	}
	out := make([]GeneratedSentence, 0, len(set.Sentences))
	for _, sentence := range set.Sentences {
		sentence.PromptRU = strings.TrimSpace(sentence.PromptRU)
		sentence.ReferenceES = strings.TrimSpace(sentence.ReferenceES)
		if sentence.PromptRU == "" || sentence.ReferenceES == "" {
			continue
		}
		out = append(out, sentence)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable sentences generated")
	}
	return out, nil
}

// GradeSentenceForCourse grades one learner submission against the prompt and reference,
// returning teacher-style markup tokens and an error count.
func (s *Service) GradeSentenceForCourse(ctx context.Context, courseCode, promptRU, referenceES, userInput string, modelOverride ...string) (*SentenceGrade, error) {
	prompt := s.sentenceGradePrompts[courseCode]
	if prompt == "" {
		return nil, fmt.Errorf("sentence grading prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"prompt_ru":    promptRU,
		"reference_es": referenceES,
		"user_input":   userInput,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal grading payload: %w", err)
	}

	messages := []Message{
		{Role: "system", Content: prompt},
		{Role: "user", Content: string(payloadJSON)},
	}
	model := s.modelOr(modelOverride...)
	raw, err := s.postChatCompletion(ctx, model, messages, 1500, 0.2, zap.String("kind", "sentence_grade"), zap.String("course", courseCode))
	if err != nil {
		return nil, err
	}
	var grade SentenceGrade
	if err := json.Unmarshal([]byte(raw), &grade); err != nil {
		return nil, fmt.Errorf("parse sentence grade: %w (raw: %s)", err, truncateForLog(raw))
	}
	if grade.ErrorCount < 0 {
		grade.ErrorCount = 0
	}
	grade.Outcome = strings.TrimSpace(grade.Outcome)
	return &grade, nil
}

// modelOr returns the first non-empty model override, else the default model.
func (s *Service) modelOr(modelOverride ...string) string {
	if len(modelOverride) > 0 && strings.TrimSpace(modelOverride[0]) != "" {
		return modelOverride[0]
	}
	return s.model
}

func truncateForLog(s string) string {
	const max = 300
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
