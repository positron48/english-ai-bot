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
	Explanation string               `json:"explanation,omitempty"`
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
		// The native-language prompt must never leak a target-language word. Weaker models
		// occasionally drop a supplied target lemma (or its inflected form) straight into
		// `prompt_ru` instead of translating it, which shows the learner the answer. Detect
		// that and skip the item rather than serve a corrupted exercise.
		if leaked := leakedTargetWord(sentence.PromptRU, words); leaked != "" {
			s.logger.Warn("sentence generation leaked target word into native prompt; dropping",
				zap.String("course", courseCode),
				zap.String("leaked_word", leaked),
				zap.String("prompt_ru", sentence.PromptRU))
			continue
		}
		out = append(out, sentence)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable sentences generated")
	}
	// The model occasionally overshoots the requested count; hold it to exactly `count`.
	if count > 0 && len(out) > count {
		out = out[:count]
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

// leakedTargetWord returns the first supplied target lemma that appears (as a whole word or a
// clearly inflected form) inside the native-language prompt, or "" if none do. The native prompt
// is written in a different script/language than the target lemmas, so a lemma root surfacing
// verbatim in it is an unambiguous leak — never a coincidental substring.
func leakedTargetWord(promptRU string, words []GenSentenceWord) string {
	lower := " " + strings.ToLower(promptRU) + " "
	for _, w := range words {
		lemma := strings.ToLower(strings.TrimSpace(w.Lemma))
		if len([]rune(lemma)) < 3 {
			continue // too short to attribute a Latin-in-Cyrillic match with confidence
		}
		root := lemmaRoot(lemma)
		// Scan every whitespace/punctuation-delimited token of the prompt for one whose
		// letters start with the lemma root. This catches both the bare lemma and inflected
		// forms (e.g. "gato"→"gatos", "comer"→"come") without matching mid-word coincidences.
		for _, tok := range strings.FieldsFunc(lower, func(r rune) bool {
			return r == ' ' || r == ',' || r == '.' || r == '!' || r == '?' || r == ';' || r == ':' || r == '"' || r == '(' || r == ')'
		}) {
			if strings.HasPrefix(tok, root) {
				return w.Lemma
			}
		}
	}
	return ""
}

// lemmaRoot strips a Spanish infinitive ending so a verb lemma matches its conjugated forms,
// then keeps a short root. Mirrors the heuristic used by the sentence LLM test harness.
func lemmaRoot(lemma string) string {
	for _, suf := range []string{"ar", "er", "ir"} {
		if len(lemma) > 4 && strings.HasSuffix(lemma, suf) {
			return lemma[:len(lemma)-len(suf)]
		}
	}
	return lemma
}

func truncateForLog(s string) string {
	const max = 300
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
