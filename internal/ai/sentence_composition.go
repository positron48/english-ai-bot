package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

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
	PromptRU        string   `json:"prompt_ru"`
	ClarificationRU string   `json:"clarification_ru"`
	ReferenceES     string   `json:"reference_es"`
	UsedWords       []string `json:"used_words"`
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
	Issues      []SentenceGradeIssue `json:"issues,omitempty"`
}

// SentenceGradeIssue makes model output auditable and prevents a vague numeric
// score from hiding what was counted. The server derives the final count from it.
type SentenceGradeIssue struct {
	Kind string `json:"kind"`
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
func (s *Service) GenerateSentenceSetForCourse(ctx context.Context, courseCode string, focusWords, supportWords []GenSentenceWord, tenses []string, count int, modelOverride ...string) ([]GeneratedSentence, error) {
	if count <= 0 || len(focusWords) == 0 {
		return nil, fmt.Errorf("sentence generation requires a positive count and focus vocabulary")
	}
	const batchSize = 5
	out := make([]GeneratedSentence, 0, count)
	for offset := 0; offset < count; offset += batchSize {
		batchCount := batchSize
		if remaining := count - offset; remaining < batchCount {
			batchCount = remaining
		}
		batchFocus := make([]GenSentenceWord, 0, batchCount)
		for i := 0; i < batchCount; i++ {
			batchFocus = append(batchFocus, focusWords[(offset+i)%len(focusWords)])
		}
		batch := make([]GeneratedSentence, 0, batchCount)
		for attempt := 0; attempt < 3 && len(batch) < batchCount; attempt++ {
			generated, err := s.generateSentenceBatchForCourse(ctx, courseCode, batchFocus, supportWords, tenses, batchCount-len(batch), modelOverride...)
			if err != nil {
				return nil, err
			}
			batch = append(batch, generated...)
		}
		if len(batch) < batchCount {
			s.logger.Warn("sentence generation batch remained short after quality retries",
				zap.String("course", courseCode), zap.Int("wanted", batchCount), zap.Int("generated", len(batch)))
		}
		out = append(out, batch...)
	}
	return out, nil
}

func (s *Service) generateSentenceBatchForCourse(ctx context.Context, courseCode string, focusWords, supportWords []GenSentenceWord, tenses []string, count int, modelOverride ...string) ([]GeneratedSentence, error) {
	prompt := s.sentenceGenPrompts[courseCode]
	if prompt == "" {
		return nil, fmt.Errorf("sentence generation prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"sentence_count": count,
		"allowed_tenses": tenses,
		"focus_words":    focusWords,
		"support_words":  supportWords,
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
	if strings.TrimSpace(raw) == "" {
		raw, err = s.postChatCompletion(ctx, model, messages, 4500, 0.4, zap.String("kind", "sentence_gen_retry"), zap.String("course", courseCode))
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("sentence generator returned an empty response")
	}
	var set generatedSentenceSet
	if err := json.Unmarshal([]byte(raw), &set); err != nil {
		s.logger.Warn("sentence generator returned malformed JSON; retrying",
			zap.String("course", courseCode), zap.Error(err))
		raw, retryErr := s.postChatCompletion(ctx, model, messages, 8000, 0.2, zap.String("kind", "sentence_gen_json_retry"), zap.String("course", courseCode))
		if retryErr != nil {
			return nil, retryErr
		}
		if err := json.Unmarshal([]byte(raw), &set); err != nil {
			return nil, fmt.Errorf("parse sentence generation retry: %w (raw: %s)", err, truncateForLog(raw))
		}
	}
	allWords := append(append(make([]GenSentenceWord, 0, len(focusWords)+len(supportWords)), focusWords...), supportWords...)
	focusSet := make(map[string]bool, len(focusWords))
	knownSet := make(map[string]bool, len(allWords))
	for _, word := range focusWords {
		focusSet[strings.ToLower(strings.TrimSpace(word.Lemma))] = true
	}
	for _, word := range allWords {
		knownSet[strings.ToLower(strings.TrimSpace(word.Lemma))] = true
	}
	out := make([]GeneratedSentence, 0, len(set.Sentences))
	for _, sentence := range set.Sentences {
		sentence.PromptRU = strings.TrimSpace(sentence.PromptRU)
		sentence.ReferenceES = strings.TrimSpace(sentence.ReferenceES)
		if sentence.PromptRU == "" || sentence.ReferenceES == "" {
			continue
		}
		if !usableSentenceClarification(sentence.ClarificationRU) {
			s.logger.Warn("sentence generation returned an unusable learner context; dropping",
				zap.String("course", courseCode), zap.String("clarification_ru", sentence.ClarificationRU))
			continue
		}
		// The native-language prompt must never leak a target-language word. Weaker models
		// occasionally drop a supplied target lemma (or its inflected form) straight into
		// `prompt_ru` instead of translating it, which shows the learner the answer. Detect
		// that and skip the item rather than serve a corrupted exercise.
		if leaked := leakedTargetWord(sentence.PromptRU, allWords); leaked != "" {
			s.logger.Warn("sentence generation leaked target word into native prompt; dropping",
				zap.String("course", courseCode),
				zap.String("leaked_word", leaked),
				zap.String("prompt_ru", sentence.PromptRU))
			continue
		}
		focusUses := 0
		validUsedWords := true
		for _, lemma := range sentence.UsedWords {
			normalized := strings.ToLower(strings.TrimSpace(lemma))
			if !knownSet[normalized] {
				validUsedWords = false
				break
			}
			if focusSet[normalized] {
				focusUses++
			}
		}
		if !validUsedWords || focusUses < 1 {
			s.logger.Warn("sentence generation did not use a valid focus vocabulary word; dropping",
				zap.String("course", courseCode),
				zap.String("prompt_ru", sentence.PromptRU),
				zap.Strings("used_words", sentence.UsedWords))
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

func usableSentenceClarification(context string) bool {
	lower := strings.ToLower(strings.TrimSpace(context))
	if lower == "" {
		return true
	}
	for _, forbidden := range []string{
		"конкретн", "неконкретн", "определённ", "определенн",
		"неопределённ", "неопределенн", "артикл", "используй el", "используй un",
	} {
		if strings.Contains(lower, forbidden) {
			return false
		}
	}
	return !strings.ContainsFunc(context, func(r rune) bool { return unicode.Is(unicode.Latin, r) })
}

// GradeSentenceForCourse grades one learner submission against the prompt and reference,
// returning teacher-style markup tokens and an error count.
func (s *Service) GradeSentenceForCourse(ctx context.Context, courseCode, promptRU, clarificationRU, referenceES, userInput string, modelOverride ...string) (*SentenceGrade, error) {
	prompt := s.sentenceGradePrompts[courseCode]
	if prompt == "" {
		return nil, fmt.Errorf("sentence grading prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"prompt_ru":        promptRU,
		"clarification_ru": clarificationRU,
		"reference_es":     referenceES,
		"user_input":       userInput,
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
	raw, err := s.postChatCompletion(ctx, model, messages, 2500, 0.2, zap.String("kind", "sentence_grade"), zap.String("course", courseCode))
	if err != nil {
		return nil, err
	}
	// Some reasoning-capable local models can spend their initial completion
	// budget on hidden reasoning and return an empty visible answer. Retry once
	// with more room so a transient model-format failure never becomes a lost
	// exercise attempt or a 502 for the learner.
	if strings.TrimSpace(raw) == "" {
		raw, err = s.postChatCompletion(ctx, model, messages, 3500, 0.2, zap.String("kind", "sentence_grade_retry"), zap.String("course", courseCode))
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("sentence grader returned an empty response")
	}
	var grade SentenceGrade
	if err := json.Unmarshal([]byte(raw), &grade); err != nil {
		return nil, fmt.Errorf("parse sentence grade: %w (raw: %s)", err, truncateForLog(raw))
	}
	if grade.ErrorCount < 0 {
		grade.ErrorCount = 0
	}
	if grade.Issues != nil {
		grade.ErrorCount = len(grade.Issues)
	}
	grade.Outcome = strings.TrimSpace(grade.Outcome)
	return &grade, nil
}

// NormalizedSentenceAnswer ignores the presentation differences explicitly not
// assessed by this exercise: leading capitalization and terminal punctuation.
func NormalizedSentenceAnswer(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRightFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(".!?¡¿", r)
	})
	s = strings.TrimLeftFunc(s, func(r rune) bool { return strings.ContainsRune("¡¿", r) })
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToLower(s)
}

// NewExactSentenceGrade is used for an answer equal to the stored reference
// after harmless normalization. It guarantees that a missing final dot can
// never be turned into a lost star by a model.
func NewExactSentenceGrade(userInput string) *SentenceGrade {
	return &SentenceGrade{ErrorCount: 0, Outcome: "star", CorrectedES: strings.TrimSpace(userInput)}
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
