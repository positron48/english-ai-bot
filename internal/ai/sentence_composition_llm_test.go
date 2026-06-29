package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	aiprompts "tgbot-skeleton/prompts"

	"go.uber.org/zap"
)

// TestSentenceLLMHarness is a MANUAL prompt-tuning harness for the daily sentence-composition
// feature. It is skipped in normal runs and only executes when RUN_SENTENCE_LLM=1, calling the
// real model so you can eyeball generation quality, grading behaviour and metrics, then tweak
// prompts/sentence-gen-ru-es.txt and prompts/sentence-grade-ru-es.txt until the numbers are good.
//
// Run (against the default nano model in your .env chain):
//
//	set -a; source .env; source .env.es; set +a   # load AI_URL / AI_API_KEY / AI_MODEL
//	RUN_SENTENCE_LLM=1 go test ./internal/ai -run TestSentenceLLMHarness -v -count=1 -timeout 600s
//
// Optional overrides:
//
//	SENTENCE_TEST_MODEL=openai/gpt-5.4-nano   # model id (default: $AI_MODEL)
//	SENTENCE_TEST_COUNT=8                      # how many sentences to generate
//	SENTENCE_TEST_INPUT="Tengo un gato"        # grade ONE custom answer against sentence #1 and stop
func TestSentenceLLMHarness(t *testing.T) {
	if os.Getenv("RUN_SENTENCE_LLM") != "1" {
		t.Skip("manual harness; set RUN_SENTENCE_LLM=1 (and AI_URL/AI_API_KEY/AI_MODEL) to run")
	}
	url := strings.TrimSpace(os.Getenv("AI_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	model := firstNonEmptyEnv("SENTENCE_TEST_MODEL", "AI_MODEL")
	if url == "" || model == "" {
		t.Fatalf("need AI_URL and a model (SENTENCE_TEST_MODEL or AI_MODEL); url=%q model=%q", url, model)
	}
	t.Logf("harness: url=%s model=%s", url, model)

	logger, _ := zap.NewDevelopment()
	svc := NewServiceWithTimeout(url, model, apiKey, "", 120*time.Second, logger)

	const course = "es_ru"
	genPrompt := mustLoadPrompt(t, "prompts/sentence-gen-ru-es.txt")
	gradePrompt := mustLoadPrompt(t, "prompts/sentence-grade-ru-es.txt")
	svc.SetSentenceGenPromptForCourse(course, genPrompt)
	svc.SetSentenceGradePromptForCourse(course, gradePrompt)

	// Deterministic grading-only mode: grade ONE fully-specified triple (no generation).
	// Useful for pinning down grader behaviour on a known prompt/reference/answer.
	//   SENTENCE_TEST_PROMPT_RU="Кот пьёт воду в доме." \
	//   SENTENCE_TEST_REF_ES="El gato bebe agua en la casa." \
	//   SENTENCE_TEST_INPUT="El gato bebe la agua"
	if fr := strings.TrimSpace(os.Getenv("SENTENCE_TEST_REF_ES")); fr != "" {
		fp := strings.TrimSpace(os.Getenv("SENTENCE_TEST_PROMPT_RU"))
		fin := os.Getenv("SENTENCE_TEST_INPUT")
		g, err := svc.GradeSentenceForCourse(context.Background(), course, fp, fr, fin, model)
		if err != nil {
			t.Fatalf("fixed grade: %v", err)
		}
		t.Logf("\nFIXED GRADE:\n  prompt_ru=%q\n  reference=%q\n  input=%q\n  errors=%d outcome=%s corrected=%q\n  tokens=%+v",
			fp, fr, fin, g.ErrorCount, g.Outcome, g.CorrectedES, g.Tokens)
		return
	}

	// A realistic pool of "well-learned" Spanish words (lemma + RU gloss).
	words := []GenSentenceWord{
		{"gato", "кот"}, {"perro", "собака"}, {"casa", "дом"}, {"comer", "есть"},
		{"beber", "пить"}, {"agua", "вода"}, {"libro", "книга"}, {"leer", "читать"},
		{"amigo", "друг"}, {"grande", "большой"}, {"pequeño", "маленький"}, {"hablar", "говорить"},
		{"trabajar", "работать"}, {"ciudad", "город"}, {"comprar", "покупать"}, {"feliz", "счастливый"},
	}
	tenses := []string{"presente (indicativo)"}
	count := atoiDefault(os.Getenv("SENTENCE_TEST_COUNT"), 8)

	ctx := context.Background()
	sentences, err := svc.GenerateSentenceSetForCourse(ctx, course, words, tenses, count, model)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}
	t.Logf("generated %d/%d sentences", len(sentences), count)

	allowed := map[string]bool{}
	for _, w := range words {
		allowed[strings.ToLower(w.Lemma)] = true
	}

	// --- Generation quality report ---
	usedWordsValid, usedWordsTotal := 0, 0
	for i, s := range sentences {
		t.Logf("\n[%d] RU: %s\n    ES: %s\n    used: %v", i+1, s.PromptRU, s.ReferenceES, s.UsedWords)
		refLower := strings.ToLower(s.ReferenceES)
		for _, uw := range s.UsedWords {
			usedWordsTotal++
			uwl := strings.ToLower(strings.TrimSpace(uw))
			// used_words must come from the provided pool AND appear (stem) in the reference.
			inPool := allowed[uwl]
			inRef := strings.Contains(refLower, stem(uwl))
			if inPool && inRef {
				usedWordsValid++
			} else {
				t.Logf("    ⚠ used word %q: inPool=%v inRef=%v", uw, inPool, inRef)
			}
		}
	}

	// Single custom-answer debug mode.
	if custom := strings.TrimSpace(os.Getenv("SENTENCE_TEST_INPUT")); custom != "" && len(sentences) > 0 {
		g, err := svc.GradeSentenceForCourse(ctx, course, sentences[0].PromptRU, sentences[0].ReferenceES, custom, model)
		if err != nil {
			t.Fatalf("grade custom: %v", err)
		}
		t.Logf("\nCUSTOM GRADE for #1 (%q):\n  input=%q\n  errors=%d outcome=%s corrected=%q\n  tokens=%+v",
			sentences[0].PromptRU, custom, g.ErrorCount, g.Outcome, g.CorrectedES, g.Tokens)
		return
	}

	// --- Grading quality report ---
	// Scenario A: feed the reference answer back -> a good grader must score 0 errors (star).
	// Scenario B: drop the last word -> a good grader must find >= 1 error (not star).
	// Scenario C: empty answer -> must be "failed".
	var aStar, bNonStar, cFailed int
	for i, s := range sentences {
		ga, err := svc.GradeSentenceForCourse(ctx, course, s.PromptRU, s.ReferenceES, s.ReferenceES, model)
		if err != nil {
			t.Errorf("[%d] grade(reference) error: %v", i+1, err)
			continue
		}
		if ga.ErrorCount == 0 && ga.Outcome == "star" {
			aStar++
		} else {
			t.Logf("[%d] FALSE POSITIVE on correct answer: errors=%d outcome=%s tokens=%+v", i+1, ga.ErrorCount, ga.Outcome, ga.Tokens)
		}

		corrupted := dropLastWord(s.ReferenceES)
		if corrupted != s.ReferenceES {
			gb, err := svc.GradeSentenceForCourse(ctx, course, s.PromptRU, s.ReferenceES, corrupted, model)
			if err == nil {
				if gb.ErrorCount >= 1 && gb.Outcome != "star" {
					bNonStar++
				} else {
					t.Logf("[%d] MISSED error on corrupted %q: errors=%d outcome=%s", i+1, corrupted, gb.ErrorCount, gb.Outcome)
				}
			}
		} else {
			bNonStar++ // single-word sentence; skip but don't penalize
		}

		gc, err := svc.GradeSentenceForCourse(ctx, course, s.PromptRU, s.ReferenceES, "", model)
		if err == nil {
			if gc.Outcome == "failed" {
				cFailed++
			} else {
				t.Logf("[%d] empty answer not failed: outcome=%s errors=%d", i+1, gc.Outcome, gc.ErrorCount)
			}
		}
	}

	n := len(sentences)
	t.Logf("\n================ METRICS ================")
	t.Logf("generation: %d sentences (asked %d)", n, count)
	t.Logf("used_words validity: %s", pct(usedWordsValid, usedWordsTotal))
	t.Logf("grader: correct→star (no false positives): %s", pct(aStar, n))
	t.Logf("grader: corrupted→detected error:          %s", pct(bNonStar, n))
	t.Logf("grader: empty→failed:                      %s", pct(cFailed, n))
	t.Logf("========================================")

	// Soft quality gates (logged, not hard-failed, so you can iterate on prompts).
	if n > 0 {
		if aStar*100/n < 90 {
			t.Logf("⚠ grader flags correct answers too often (want >=90%% star on references)")
		}
		if bNonStar*100/n < 80 {
			t.Logf("⚠ grader misses introduced errors too often (want >=80%% detection)")
		}
	}
}

func mustLoadPrompt(t *testing.T, path string) string {
	t.Helper()
	p, err := LoadRenderedPromptFile(path, "ru", "es", "ru-es")
	if err != nil {
		// Fall back to embedded FS directly (cwd-independent).
		b, e2 := aiprompts.FS.ReadFile(strings.TrimPrefix(path, "prompts/"))
		if e2 != nil {
			t.Fatalf("load prompt %s: %v / %v", path, err, e2)
		}
		return PreparePrompt(string(b), "ru", "es", "ru-es")
	}
	return p
}

func firstNonEmptyEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func atoiDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

// stem returns a crude prefix so a lemma matches its inflected forms in the reference
// (e.g. "comer"→"com" matches "come", "beber"→"beb" matches "bebo"). Spanish verb/noun
// roots are short, so 3 chars keeps false negatives low without over-matching.
func stem(w string) string {
	// Drop common infinitive endings before taking the root.
	for _, suf := range []string{"ar", "er", "ir"} {
		if len(w) > 3 && strings.HasSuffix(w, suf) {
			w = w[:len(w)-len(suf)]
			break
		}
	}
	if len(w) > 3 {
		return w[:3]
	}
	return w
}

func dropLastWord(s string) string {
	fields := strings.Fields(strings.TrimRight(s, ".!?"))
	if len(fields) < 2 {
		return s
	}
	return strings.Join(fields[:len(fields)-1], " ")
}

func pct(n, total int) string {
	if total == 0 {
		return "n/a (0)"
	}
	return strings.TrimSpace(itoa(n*100/total)) + "% (" + itoa(n) + "/" + itoa(total) + ")"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
