package ai

import (
	"context"
	"encoding/json"
	"errors"
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

type sentenceQualityReview struct {
	Checks []sentenceQualityCheck `json:"checks"`
}

type sentenceQualityCheck struct {
	Position        int      `json:"position"`
	Accepted        bool     `json:"accepted"`
	Reason          string   `json:"reason"`
	ContextEvidence []string `json:"context_evidence"`
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
// score from hiding what was counted. The server checks that these spans explain
// every changed word before deriving the score from the minimal text edit.
type SentenceGradeIssue struct {
	Kind        string `json:"kind"`
	Original    string `json:"original"`
	Corrected   string `json:"corrected"`
	Explanation string `json:"explanation"`
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

// ErrNoReviewedSentences is a content-quality shortfall, not an AI availability failure.
var ErrNoReviewedSentences = errors.New("sentence generation produced no reviewed unique items")

type sentenceRejection struct {
	Sentence GeneratedSentence `json:"sentence"`
	Reason   string            `json:"reason"`
}

type sentenceGenerationResult struct {
	Accepted []GeneratedSentence
	Rejected []sentenceRejection
}

// GenerateSentenceSetForCourse reviews one shared candidate pool and makes at most
// one targeted refill. Only reviewed, unique items are returned, even on a shortfall.
func (s *Service) GenerateSentenceSetForCourse(ctx context.Context, courseCode string, focusWords, supportWords []GenSentenceWord, tenses []string, count int, modelOverride ...string) ([]GeneratedSentence, error) {
	if count <= 0 || len(focusWords) == 0 {
		return nil, fmt.Errorf("sentence generation requires a positive count and focus vocabulary")
	}
	out := make([]GeneratedSentence, 0, count)
	seen := make(map[string]bool, count)
	var rejected []sentenceRejection
	for attempt := 0; attempt < 2 && len(out) < count; attempt++ {
		missing := count - len(out)
		candidateCount := missing + (missing+1)/2
		result, err := s.generateSentenceBatchForCourse(ctx, courseCode, focusWords, supportWords, tenses, candidateCount, out, rejected, modelOverride...)
		if err != nil {
			if len(out) == 0 {
				return nil, err
			}
			s.logger.Warn("sentence refill failed; keeping reviewed items", zap.String("course", courseCode), zap.Error(err), zap.Int("accepted", len(out)))
			break
		}
		rejected = result.Rejected
		for _, sentence := range result.Accepted {
			key := NormalizedSentenceAnswer(sentence.PromptRU)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, sentence)
			if len(out) == count {
				break
			}
		}
	}
	if len(out) == 0 {
		return nil, ErrNoReviewedSentences
	}
	if len(out) < count {
		s.logger.Warn("sentence generation returning partial reviewed set", zap.String("course", courseCode), zap.Int("wanted", count), zap.Int("accepted", len(out)))
	}
	return out, nil
}

func (s *Service) generateSentenceBatchForCourse(ctx context.Context, courseCode string, focusWords, supportWords []GenSentenceWord, tenses []string, count int, accepted []GeneratedSentence, rejected []sentenceRejection, modelOverride ...string) (sentenceGenerationResult, error) {
	prompt := s.sentenceGenPrompts[courseCode]
	if prompt == "" {
		return sentenceGenerationResult{}, fmt.Errorf("sentence generation prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"sentence_count":     count,
		"allowed_tenses":     tenses,
		"focus_words":        focusWords,
		"support_words":      supportWords,
		"accepted_sentences": accepted,
		"rejected_sentences": rejected,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return sentenceGenerationResult{}, fmt.Errorf("marshal generation payload: %w", err)
	}

	messages := []Message{
		{Role: "system", Content: prompt},
		{Role: "user", Content: string(payloadJSON)},
	}
	model := s.sentenceModelOr(modelOverride...)
	// A single request per round; invalid output does not trigger hidden retries.
	raw, err := s.postSentenceChatCompletion(ctx, model, messages, sentenceGenerationTokenLimit(count), 0.6, zap.String("kind", "sentence_gen"), zap.String("course", courseCode))
	if err != nil {
		return sentenceGenerationResult{}, err
	}
	var set generatedSentenceSet
	if err := json.Unmarshal([]byte(raw), &set); err != nil {
		return sentenceGenerationResult{}, fmt.Errorf("parse sentence generation: %w", err)
	}
	// Bound review size even if the generator ignores sentence_count.
	if len(set.Sentences) > count {
		set.Sentences = set.Sentences[:count]
	}
	var result sentenceGenerationResult
	reject := func(sentence GeneratedSentence, reason string) {
		result.Rejected = append(result.Rejected, sentenceRejection{Sentence: sentence, Reason: reason})
		s.logger.Info("sentence rejected by local validation", zap.String("course", courseCode), zap.String("prompt_ru", sentence.PromptRU), zap.String("reason", reason))
	}
	allWords := append(append(make([]GenSentenceWord, 0, len(focusWords)+len(supportWords)), focusWords...), supportWords...)
	focusSet := make(map[string]bool, len(focusWords))
	knownSet := make(map[string]bool, len(allWords))
	canonicalByAlias := make(map[string]string, len(allWords)*2)
	for _, word := range focusWords {
		focusSet[strings.ToLower(strings.TrimSpace(word.Lemma))] = true
	}
	for _, word := range allWords {
		lemma := strings.ToLower(strings.TrimSpace(word.Lemma))
		knownSet[lemma] = true
		canonicalByAlias[lemma] = lemma
		for _, translation := range strings.Split(word.Translation, "/") {
			alias := strings.ToLower(strings.TrimSpace(translation))
			if alias != "" {
				canonicalByAlias[alias] = lemma
			}
		}
	}
	out := make([]GeneratedSentence, 0, len(set.Sentences))
	seenPrompts := make(map[string]bool, len(accepted)+len(set.Sentences))
	for _, sentence := range accepted {
		seenPrompts[NormalizedSentenceAnswer(sentence.PromptRU)] = true
	}
	for _, sentence := range set.Sentences {
		sentence.PromptRU = strings.TrimSpace(sentence.PromptRU)
		sentence.ReferenceES = strings.TrimSpace(sentence.ReferenceES)
		if sentence.PromptRU == "" || sentence.ReferenceES == "" {
			reject(sentence, "Russian prompt and target translation must both be nonempty.")
			continue
		}
		if !usableSentenceClarification(sentence.ClarificationRU) {
			reject(sentence, "Rewrite clarification_ru as observable Russian facts explaining every article and addressee. Do not use конкретн, определённ, определенн, артикл or Latin letters; do not merely label an object as specific.")
			continue
		}
		if strings.TrimSpace(sentence.ClarificationRU) == "" && ((courseCode == "es_ru" && hasRussianAddress(sentence.PromptRU)) || hasSentenceArticle(courseCode, sentence.ReferenceES)) {
			reject(sentence, "Add Russian context resolving every article and, for Spanish вы, number of addressees, formality and region when needed.")
			continue
		}
		if leaked := leakedTargetWord(sentence.PromptRU, allWords); leaked != "" {
			reject(sentence, "Translate the leaked target-language word in prompt_ru: "+leaked)
			continue
		}
		focusUses := 0
		validUsedWords := true
		canonicalUsedWords := make([]string, 0, len(sentence.UsedWords))
		for _, reportedWord := range sentence.UsedWords {
			normalized := strings.ToLower(strings.TrimSpace(reportedWord))
			canonical, ok := canonicalByAlias[normalized]
			if !ok {
				canonical = normalized
			}
			if !knownSet[canonical] {
				validUsedWords = false
				break
			}
			canonicalUsedWords = append(canonicalUsedWords, canonical)
			if focusSet[canonical] {
				focusUses++
			}
		}
		if !validUsedWords || focusUses < 1 {
			reject(sentence, "used_words must list supplied lemmas only and contain at least one focus lemma actually used in the sentence.")
			continue
		}
		key := NormalizedSentenceAnswer(sentence.PromptRU)
		if seenPrompts[key] {
			reject(sentence, "Duplicate Russian prompt. Write a different sentence, not another translation or context for the same prompt.")
			continue
		}
		seenPrompts[key] = true
		sentence.UsedWords = canonicalUsedWords
		out = append(out, sentence)
	}
	if len(out) == 0 {
		return result, nil
	}
	reviewed, err := s.reviewGeneratedSentenceQuality(ctx, model, courseCode, focusWords, supportWords, tenses, out, accepted)
	if err != nil {
		return result, err
	}
	result.Accepted = reviewed.Accepted
	result.Rejected = append(result.Rejected, reviewed.Rejected...)
	return result, nil
}

// Allow enough output for a shared pool including learner-visible context.
func sentenceGenerationTokenLimit(count int) int {
	limit := count * 400
	if limit < 3000 {
		return 3000
	}
	return limit
}

func (s *Service) reviewGeneratedSentenceQuality(ctx context.Context, model, courseCode string, focusWords, supportWords []GenSentenceWord, tenses []string, sentences []GeneratedSentence, accepted []GeneratedSentence) (sentenceGenerationResult, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"allowed_vocabulary": append(append([]GenSentenceWord{}, focusWords...), supportWords...),
		"sentences":          sentences,
		"accepted_sentences": accepted,
		"course":             courseCode,
		"allowed_tenses":     tenses,
	})
	if err != nil {
		return sentenceGenerationResult{}, err
	}
	const reviewPrompt = `Проверь каждое упражнение на перевод. Сначала найди основания для каждого выбора, затем реши, можно ли показывать задание ученику.
Вход: course (es_ru или en_ru), allowed_vocabulary, allowed_tenses, sentences и accepted_sentences (ранее принятые задания того же набора). Проверяй только позиции sentences; accepted_sentences учитывай для разнообразия, не меняй их. reference_es — историческое имя поля эталона, в английском курсе там английский.
Верни ТОЛЬКО JSON: {"checks":[{"position":0,"accepted":true,"reason":"краткое обоснование", "context_evidence":["точная цитата из clarification_ru для первого выбора", "цитата для второго выбора"]}]}.
Одна проверка на каждую позицию. context_evidence — дословные подстроки clarification_ru длиной НЕ МЕНЕЕ 12 символов каждая. Копируй их с тем же регистром, пунктуацией, ё/е и пробелами, не пересказывай. Можно скопировать целое предложение и повторить его, если оно обосновывает несколько выборов. При отказе цитаты не обязательны. Отклоняй при ЛЮБОМ нарушении:
1. Все смысловые слова ОБЕИХ языков должны быть в allowed_vocabulary (с учётом форм и перевода). Проверь каждое существительное, прилагательное, основной глагол, наречие. Новые слова нельзя добавлять ради естественности. used_words должен включать все использованные леммы без лишних. Например, «давать»/dar запрещён, если его нет в списке. Служебные слова и связки разрешены. Tener (включая tiene/tienen) — смысловой глагол, НЕ связка; без tener в allowed_vocabulary отклони «у X есть». Наречия и время «сейчас/right now», «каждый день/every day», «часто/often», «в этом месяце/this month», «около/cerca» тоже требуют словаря, даже если помогают обозначить время глагола или увеличить длину. Лексика только внутри пояснения не ограничена.
2. Русское предложение должно выражать законченную естественную бытовую мысль; эталон должен иметь личную форму глагола, точно передавать смысл и соблюдать allowed_tenses. Русская нулевая связка допустима. Отклоняй приветствия/благодарности, перечни, пары чисел и фрагменты («Привет, спасибо», «Семь или восемь?», «Чай и молоко», «Кофе и сахар»): это не упражнения на составление предложений. Отклоняй бессодержательные «Это момент», «Это политика», «X — это средство» без указания средства чего, «Форма X — это проблема» без сути проблемы, очевидные отрицания «Метро — не отель», «Отель — не библиотека», даже с придуманным контекстом. Полезная мысль должна быть в самом задании, а не только в пояснении. Отклоняй неестественные комбинации ради лексики (любые действия с «чистой книгой», мать ищет окно в комнате, «ты несёшь воду здесь», рука чистая, язык красный, ребёнок ест еду). Не оправдывай странную фразу специально придуманной предысторией.
3. Проверь КАЖДЫЙ артикль в эталоне отдельно, включая подлежащее и объект, включая al/del. Найди в clarification_ru точную цитату, объясняющую, почему ОБА собеседника знают именно этого человека/предмет, или почему он впервые вводится/ещё не выбран, или что речь обо всём классе. Добавь по одной такой цитате в context_evidence для каждого артикля по порядку.
«Друг открывает дверь» + пояснение только о двери НЕ обосновывает el amigo: отклони. «Врач открывает то самое окно в кабинете» не объясняет, как оба знают врача: отклони. Само наличие существительного в prompt_ru не делает его известным. Простого «мы его видим» тоже недостаточно, если непонятно, какой из многих предметов. Для неопределённого артикля контекст не должен одновременно указывать на уже опознанный предмет.
4. Для испанского «вы/вам/вас/ваш» в русском задании добавь ЕЩЁ ОДНУ цитату в context_evidence, которая задаёт число адресатов, вежливость, при необходимости регион и пол. Если такой цитаты нет — отклони. Usted: один человек вежливо, глагол в 3-м лице ед. числа. Vosotros/vosotras: несколько человек на ты в Испании, 2-е лицо мн. числа. Ustedes: несколько человек вежливо или в Латинской Америке, 3-е лицо мн. числа. Проверь также глагол при опущенном местоимении. «Одному другу на вы» с abres — ОШИБКА, это форма tú. Не принимай контекст с развилками «вежливо ИЛИ в Латинской Америке»: нужна одна ясная ситуация.
5. Контекст не должен противоречить переводу, добавлять смысл к самому предложению, выдавать испанские/английские слова, рассказывать абстрактные правила. Он должен помогать выбрать ответ. Если речь о поиске в контексте, а предложение говорит о находке — отклони как несогласованное.
6. После проверок 1–5 оцени разнообразие прошедших заданий вместе с accepted_sentences. Оставляй не более ДВУХ заданий одного каркаса; у лишних accepted=false с причиной, называющей повтор и предлагающей другое действие/отношение. Замена только существительного, лица, числа, отрицания или вопросительного знака не создаёт нового каркаса («Где отель?», «Где метро?», «Где терраса?»). Для «X находится у Y» и «X у Y» это тоже один каркас; «X — проблема» и «Форма X — не проблема» не становятся разными конструкциями из-за добавленного слова. Разные смысловые глаголы или отношения могут создавать разные каркасы, даже при одинаковом начале «Я ...». Не принимай серию голых «Это X» / «Он X». Короткая полезная фраза сама по себе допустима; длина не заменяет содержательность. Для большинства набора ожидаются развёрнутые мысли с деталями из словаря, без искусственного удлинения. Не считай отклонённые по 1–5 задания при ограничении повторов.
Если основания для хотя бы одного выбора нет, accepted=false. Не домысливай недостающие факты. В reason кратко назови факты или причину отказа. Ничего не исправляй.`
	messages := []Message{{Role: "system", Content: reviewPrompt}, {Role: "user", Content: string(payload)}}
	raw, err := s.postSentenceChatCompletion(ctx, model, messages, sentenceGenerationTokenLimit(len(sentences)), 0, zap.String("kind", "sentence_quality_review"), zap.String("course", courseCode))
	if err != nil {
		s.logger.Warn("sentence quality review failed; dropping unreviewed candidates", zap.Error(err))
		return sentenceGenerationResult{}, fmt.Errorf("sentence quality review: %w", err)
	}
	var review sentenceQualityReview
	if err := json.Unmarshal([]byte(raw), &review); err != nil {
		s.logger.Warn("sentence quality review returned invalid JSON; dropping unreviewed candidates", zap.Error(err))
		return sentenceGenerationResult{}, fmt.Errorf("sentence quality review: %w", err)
	}
	checks := make(map[int][]sentenceQualityCheck, len(review.Checks))
	for _, check := range review.Checks {
		checks[check.Position] = append(checks[check.Position], check)
	}
	var result sentenceGenerationResult
	for i, sentence := range sentences {
		reason := "Quality review must return exactly one check per candidate with a nonempty reason."
		passed := false
		if entries := checks[i]; len(entries) == 1 {
			check := entries[0]
			if strings.TrimSpace(check.Reason) != "" {
				reason = check.Reason
				if check.Accepted {
					required := sentenceArticleCount(courseCode, sentence.ReferenceES)
					if courseCode == "es_ru" && hasRussianAddress(sentence.PromptRU) {
						required++
					}
					passed = len(check.ContextEvidence) >= required
					if !passed {
						reason = "Missing context evidence for one or more articles or the addressee: " + check.Reason
					}
					for _, evidence := range check.ContextEvidence {
						if len([]rune(strings.TrimSpace(evidence))) < 12 || !strings.Contains(sentence.ClarificationRU, evidence) {
							passed = false
							reason = "Context evidence must be an exact quote of at least 12 characters from clarification_ru: " + check.Reason
							break
						}
					}
				}
			}
		}
		if passed {
			result.Accepted = append(result.Accepted, sentence)
		} else {
			result.Rejected = append(result.Rejected, sentenceRejection{Sentence: sentence, Reason: reason})
			s.logger.Info("sentence rejected by quality review", zap.String("course", courseCode), zap.String("prompt_ru", sentence.PromptRU), zap.String("reason", reason))
		}
	}
	return result, nil
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

func hasSpanishArticle(sentence string) bool {
	for _, token := range strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	}) {
		switch token {
		case "el", "la", "los", "las", "un", "una", "unos", "unas", "al", "del":
			return true
		}
	}
	return false
}

func hasSentenceArticle(courseCode, sentence string) bool {
	return sentenceArticleCount(courseCode, sentence) > 0
}

func sentenceArticleCount(courseCode, sentence string) int {
	count := 0
	for _, word := range strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool { return !unicode.IsLetter(r) }) {
		if courseCode == "en_ru" {
			if word == "a" || word == "an" || word == "the" {
				count++
			}
		} else {
			switch word {
			case "el", "la", "los", "las", "un", "una", "unos", "unas", "al", "del":
				count++
			}
		}
	}
	return count
}

// GradeSentenceForCourse checks exact reference matches locally, then grades other
// translations from learner-visible facts only, avoiding reference-answer bias.
func (s *Service) GradeSentenceForCourse(ctx context.Context, courseCode, promptRU, clarificationRU, referenceES, userInput string, modelOverride ...string) (*SentenceGrade, error) {
	prompt := s.sentenceGradePrompts[courseCode]
	if prompt == "" {
		return nil, fmt.Errorf("sentence grading prompt not set for course %q", courseCode)
	}

	payload := map[string]interface{}{
		"prompt_ru":        promptRU,
		"clarification_ru": clarificationRU,
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
	if NormalizedSentenceAnswer(userInput) != "" && NormalizedSentenceAnswer(userInput) == NormalizedSentenceAnswer(referenceES) {
		return NewExactSentenceGrade(userInput), nil
	}
	model := s.sentenceModelOr(modelOverride...)
	var validationErr error
	for attempt := 0; attempt < 2; attempt++ {
		maxTokens := 2500
		if attempt > 0 {
			maxTokens = 3500
		}
		raw, err := s.postSentenceChatCompletion(ctx, model, messages, maxTokens, 0.1, zap.String("kind", "sentence_grade"), zap.String("course", courseCode))
		if err != nil {
			return nil, err
		}
		// Always decode into a fresh value: fields omitted on retry must not survive.
		var grade SentenceGrade
		if err := json.Unmarshal([]byte(raw), &grade); err != nil {
			validationErr = fmt.Errorf("invalid grading JSON: %w", err)
		} else {
			validationErr = normalizeSentenceGrade(&grade, promptRU, userInput)
			if validationErr == nil {
				if attempt == 0 && sentenceGradeNeedsMeaningAudit(grade) {
					messages = append(messages, Message{Role: "assistant", Content: raw}, Message{Role: "user", Content: sentenceMeaningAuditPrompt})
					continue
				}
				return &grade, nil
			}
		}
		s.logger.Warn("inconsistent sentence grade", zap.Int("attempt", attempt+1), zap.Error(validationErr))
		messages = append(messages, Message{Role: "assistant", Content: raw}, Message{Role: "user", Content: "Re-grade the original learner input. The previous response was inconsistent: " + validationErr.Error() + ". Return complete JSON with a minimal corrected_es and exactly one anchored issue (original, corrected, Russian explanation) per necessary correction. Cosmetic punctuation/case/spacing and optional style do not count. If there are no real errors, keep user_input unchanged, issues=[], explanation=\"\". Do not invent context absent from the input."})
	}
	return nil, fmt.Errorf("sentence grader failed consistency check: %w", validationErr)
}

// Context-dependent penalties and added Spanish subjects are the most common
// sources of false negatives. Audit these selectively rather than doubling every call.
const sentenceMeaningAuditPrompt = `Проведи независимую проверку предложенных исправлений и верни полный JSON заново.
Проверь каждую замену артикля/обращения: какой ИМЕННО факт из показанного контекста или исходного русского текста её требует? Отсутствие факта о книге НЕ доказывает, что книга неизвестная; контекст про стол НЕ задаёт известность книги. Если оба варианта естественны, восстанови выбор ученика и убери ложную ошибку, сохранив остальные реальные исправления.
Проверь род/число в ИСПРАВЛЕННОМ ответе: нельзя исправить el libro на una libro.
Не вставляй и не удаляй необязательные испанские местоимения подлежащего. Если ученик написал неверную форму глагола без местоимения, исправь только форму, сохранив отсутствие местоимения (например, для обращения к группе в Мексике достаточно поменять только форму глагола). Не увеличивай число ошибок за добавленное тобой местоимение.
Явные факты контекста обязательны: вежливое обращение к одному человеку не может оставаться множественным; явно выбранный общий предмет нельзя заменить любым случайным. Не принимай предыдущую оценку на веру. Объяснения — по-русски.`

func sentenceGradeNeedsMeaningAudit(grade SentenceGrade) bool {
	for _, issue := range grade.Issues {
		if issue.Kind == "article" || issue.Kind == "pronoun" {
			return true
		}
		if len(strings.Fields(issue.Corrected)) > len(strings.Fields(issue.Original)) {
			for _, word := range strings.Fields(NormalizedSentenceAnswer(issue.Corrected)) {
				switch word {
				case "yo", "tú", "él", "ella", "usted", "nosotros", "nosotras", "vosotros", "vosotras", "ellos", "ellas", "ustedes":
					return true
				}
			}
		}
	}
	return false
}

// normalizeSentenceGrade enforces the contract shared by score, correction markup
// and explanations. Invalid output is re-graded before an attempt can be recorded.
func normalizeSentenceGrade(grade *SentenceGrade, promptRU, userInput string) error {
	if strings.TrimSpace(grade.CorrectedES) == "" {
		return fmt.Errorf("corrected_es is empty")
	}
	if grade.Issues == nil {
		return fmt.Errorf("issues array is missing")
	}
	issues := make([]SentenceGradeIssue, 0, len(grade.Issues))
	for _, issue := range grade.Issues {
		if strings.TrimSpace(issue.Original) == "" && strings.TrimSpace(issue.Corrected) == "" {
			return fmt.Errorf("issue has no original/corrected spans")
		}
		if NormalizedSentenceAnswer(issue.Original) == NormalizedSentenceAnswer(issue.Corrected) {
			continue
		}
		switch issue.Kind {
		case "spelling", "article", "pronoun", "verb_form", "missing_word", "extra_word", "agreement", "preposition", "meaning":
		default:
			return fmt.Errorf("unsupported issue kind %q", issue.Kind)
		}
		if !sentenceContainsSpan(userInput, issue.Original) || !sentenceContainsSpan(grade.CorrectedES, issue.Corrected) {
			return fmt.Errorf("issue spans do not match the input and correction")
		}
		if strings.TrimSpace(issue.Explanation) == "" || !sentenceExplanationLanguageMatches(promptRU, issue.Explanation) {
			return fmt.Errorf("each issue needs a Russian explanation")
		}
		issues = append(issues, issue)
	}
	unchanged := NormalizedSentenceAnswer(userInput) == NormalizedSentenceAnswer(grade.CorrectedES)
	if len(issues) == 0 {
		if !unchanged {
			return fmt.Errorf("corrected_es changes words without issues")
		}
		*grade = *NewExactSentenceGrade(userInput)
		return nil
	}
	if unchanged {
		return fmt.Errorf("issues claim errors but corrected_es has no visible correction")
	}
	if !sentenceIssuesCoverCorrection(userInput, grade.CorrectedES, issues) {
		return fmt.Errorf("issues do not account for all corrected words, or count the same edit twice")
	}
	// Build feedback from exactly the issues counted. The model's summary can omit
	// an error or introduce an unsupported suggestion, so it is not authoritative.
	explanations := make([]string, 0, len(issues))
	for _, issue := range issues {
		explanations = append(explanations, strings.TrimSpace(issue.Explanation))
	}
	grade.Explanation = strings.Join(explanations, " ")
	grade.Issues = issues
	grade.ErrorCount = sentenceWordEditDistance(userInput, grade.CorrectedES)
	grade.Outcome = "passed"
	if grade.ErrorCount > 1 {
		grade.Outcome = "failed"
	}
	grade.Tokens = nil // The UI derives markup from the validated minimal correction.
	return nil
}

// Account for every inserted/deleted/replaced word, including repeated words.
// A model may otherwise explain only one edit while silently rewriting others.
func sentenceIssuesCoverCorrection(input, corrected string, issues []SentenceGradeIssue) bool {
	balance := map[string]int{}
	add := func(text string, sign int) {
		for _, word := range strings.Fields(NormalizedSentenceAnswer(text)) {
			balance[word] += sign
		}
	}
	add(input, 1)
	add(corrected, -1)
	coveredEdits := 0
	for _, issue := range issues {
		add(issue.Original, -1)
		add(issue.Corrected, 1)
		coveredEdits += sentenceWordEditDistance(issue.Original, issue.Corrected)
	}
	for _, count := range balance {
		if count != 0 {
			return false
		}
	}
	edits := sentenceWordEditDistance(input, corrected)
	return coveredEdits == edits && len(issues) <= edits
}

func sentenceWordEditDistance(a, b string) int {
	left, right := strings.Fields(NormalizedSentenceAnswer(a)), strings.Fields(NormalizedSentenceAnswer(b))
	row := make([]int, len(right)+1)
	for j := range row {
		row[j] = j
	}
	for i, word := range left {
		previous := row[0]
		row[0] = i + 1
		for j, other := range right {
			old := row[j+1]
			cost := 0
			if word != other {
				cost = 1
			}
			row[j+1] = min(row[j]+1, row[j+1]+1, previous+cost)
			previous = old
		}
	}
	return row[len(right)]
}

func sentenceContainsSpan(sentence, span string) bool {
	span = NormalizedSentenceAnswer(span)
	return span == "" || strings.Contains(" "+NormalizedSentenceAnswer(sentence)+" ", " "+span+" ")
}

func hasRussianAddress(sentence string) bool {
	for _, word := range strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool { return !unicode.IsLetter(r) }) {
		switch word {
		case "вы", "вас", "вам", "вами", "ваш", "ваша", "ваше", "ваши", "вашего", "вашей", "ваших", "вашему", "вашим", "вашими", "вашу", "вашем":
			return true
		}
	}
	return false
}

// Sentence models are configured independently of dictionary and chat models.
func (s *Service) SetSentenceReasoningEffort(effort string) {
	s.sentenceReasoningEffort = strings.ToLower(strings.TrimSpace(effort))
}

func (s *Service) SetSentenceModel(model string) { s.sentenceModel = strings.TrimSpace(model) }

func (s *Service) sentenceModelOr(overrides ...string) string {
	if len(overrides) > 0 && strings.TrimSpace(overrides[0]) != "" {
		return strings.TrimSpace(overrides[0])
	}
	if s.sentenceModel != "" {
		return s.sentenceModel
	}
	return s.model
}

func sentenceExplanationLanguageMatches(prompt, explanation string) bool {
	if strings.TrimSpace(explanation) == "" || !strings.ContainsFunc(prompt, func(r rune) bool { return unicode.Is(unicode.Cyrillic, r) }) {
		return true
	}
	return strings.ContainsFunc(explanation, func(r rune) bool { return unicode.Is(unicode.Cyrillic, r) })
}

// NormalizedSentenceAnswer ignores cosmetic punctuation, case and spacing, as
// does the correction UI. Keep lexical accents, apostrophes and hyphens: they
// can distinguish words (te/té, we're/were, re-sign/resign).
func NormalizedSentenceAnswer(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(".,!?¡¿:;\"«»“”()[]{}…", r) {
			return ' '
		}
		if r == '’' {
			return '\''
		}
		return unicode.ToLower(r)
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

// NewExactSentenceGrade is used for an answer equal to the stored reference
// after harmless normalization. It guarantees that a missing final dot can
// never be turned into a lost star by a model.
func NewExactSentenceGrade(userInput string) *SentenceGrade {
	return &SentenceGrade{ErrorCount: 0, Outcome: "star", CorrectedES: strings.TrimSpace(userInput), Issues: []SentenceGradeIssue{}}
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
