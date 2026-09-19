package ai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Authored expectations check the reviewer independently of the model's own generation.
func TestSentenceReviewRegressionLLM(t *testing.T) {
	if os.Getenv("RUN_SENTENCE_REVIEW_LLM") != "1" {
		t.Skip("opt-in paid review regression")
	}
	env, _ := godotenv.Read("../../.env")
	key := os.Getenv("POLZA_AI_API_KEY")
	if key == "" {
		key = env["POLZA_AI_API_KEY"]
	}
	if key == "" {
		t.Fatal("POLZA_AI_API_KEY required")
	}
	model := os.Getenv("SENTENCE_TEST_MODEL")
	if model == "" {
		t.Fatal("SENTENCE_TEST_MODEL required")
	}
	cases := []struct {
		name, prompt, context, reference string
		good                             bool
	}{
		{"formal singular valid", "Вы пьёте воду.", "Вы вежливо обращаетесь к одному незнакомому человеку, который пьёт обычную питьевую воду.", "Usted bebe agua.", true},
		{"formal singular wrong verb", "Вы пьёте воду?", "Вы вежливо обращаетесь к одному незнакомому человеку, который пьёт обычную питьевую воду.", "¿Vosotros bebéis agua?", false},
		{"Russian vy cannot be tu", "Вы покупаете книги.", "Вы обращаетесь к одному человеку на ты. Он выбирает любые книги в магазине.", "Compras libros.", false},
		{"unselected book definite", "Я покупаю книгу.", "Вы ещё не выбрали книгу и собираетесь купить любую подходящую в магазине.", "Compro el libro.", false},
		{"missing subject identification", "Врач открывает окно.", "Окно — то самое, которое собеседники только что попросили открыть. О враче ничего не сообщается.", "El médico abre la ventana.", false},
		{"big is not tall", "Она ищет большого друга.", "Она ищет своего высокого приятеля среди людей. Речь только о его росте.", "Busca a un amigo grande.", false},
		{"identified book indefinite", "Я читаю книгу.", "Собеседник только что передал вам единственную книгу и спросил, что вы делаете с ней.", "Leo un libro.", false},
		{"identified book valid", "Я читаю новую книгу.", "Собеседник только что передал вам новую книгу и спросил, что вы делаете с ней.", "Leo el libro nuevo.", true},
		{"unselected book valid", "Я покупаю новую книгу.", "Вы ещё не выбрали книгу и хотите купить любое новое издание из магазина.", "Compro un libro nuevo.", true},
		{"Spain informal valid", "Вы помогаете другу.", "В Испании вы обращаетесь к двум близким друзьям на ты. Вы все хорошо знаете товарища, которому они помогают.", "Ayudáis al amigo.", true},
		{"artificial clean book", "Мать находит чистую книгу.", "Собеседники обсуждают свою маму, которая нашла одну любую незагрязнённую книгу.", "La madre encuentra un libro limpio.", false},
		{"unknown content verb", "Я стираю книгу.", "Вы взяли любую книгу и стираете её в воде.", "Lavo un libro.", false},
	}
	svc := NewServiceWithTimeout("https://polza.ai/api/v1", model, key, "", 90*time.Second, zap.NewNop())
	svc.SetSentenceReasoningEffort(os.Getenv("SENTENCE_TEST_REASONING_EFFORT"))
	usage := observeSentenceUsage(t, svc)
	sentences := make([]GeneratedSentence, 0, len(cases))
	// Derive complete supplied lemma claims for this fixed corpus explicitly.
	used := [][]string{{"beber", "agua"}, {"beber", "agua"}, {"comprar", "libro"}, {"comprar", "libro"}, {"médico", "abrir", "ventana"}, {"buscar", "grande", "amigo"}, {"leer", "libro"}, {"leer", "libro", "nuevo"}, {"comprar", "libro", "nuevo"}, {"ayudar", "amigo"}, {"madre", "encontrar", "libro", "limpio"}, {"libro"}}
	for i, c := range cases {
		sentences = append(sentences, GeneratedSentence{PromptRU: c.prompt, ClarificationRU: c.context, ReferenceES: c.reference, UsedWords: used[i]})
	}
	start := time.Now()
	result, err := svc.reviewGeneratedSentenceQuality(context.Background(), model, "es_ru", sentenceTestWords("adversarial"), nil, []string{"presente (indicativo)"}, sentences, nil)
	if err != nil {
		t.Fatal(err)
	}
	accepted := map[string]bool{}
	for _, s := range result.Accepted {
		accepted[s.PromptRU] = true
	}
	passed := 0
	for _, c := range cases {
		got := accepted[c.prompt]
		if got == c.good {
			passed++
		} else {
			t.Errorf("%s: accepted=%v want=%v", c.name, got, c.good)
		}
		t.Logf("REVIEW_CASE %s accepted=%v want=%v", c.name, got, c.good)
	}
	for _, r := range result.Rejected {
		t.Logf("REVIEW_REJECTION %s: %s", r.Sentence.PromptRU, r.Reason)
	}
	t.Logf("REVIEW_SUMMARY model=%s passed=%d/%d calls=%d cost=%f seconds=%.1f", model, passed, len(cases), usage.Calls, usage.Cost, time.Since(start).Seconds())
}

// Review must reject screenshot-style fragments and consider templates already
// accepted before a refill, while preserving useful short complete sentences.
func TestSentenceDiversityReviewLLM(t *testing.T) {
	if os.Getenv("RUN_SENTENCE_REVIEW_LLM") != "1" {
		t.Skip("opt-in paid diversity review regression")
	}
	env, _ := godotenv.Read("../../.env")
	key := os.Getenv("POLZA_AI_API_KEY")
	if key == "" {
		key = env["POLZA_AI_API_KEY"]
	}
	if key == "" {
		t.Fatal("POLZA_AI_API_KEY required")
	}
	model := "google/gemini-3.8-flash"
	svc := NewServiceWithTimeout("https://polza.ai/api/v1", model, key, "", 90*time.Second, zap.NewNop())
	svc.SetSentenceReasoningEffort("low")
	usage := observeSentenceUsage(t, svc)
	words := []GenSentenceWord{{"hola", "привет"}, {"gracias", "спасибо"}, {"té", "чай"}, {"café", "кофе"}, {"siete", "семь"}, {"ocho", "восемь"}, {"hotel", "отель"}, {"metro", "метро"}, {"terraza", "терраса"}, {"agua", "вода"}, {"leche", "молоко"}, {"beber", "пить"}, {"leer", "читать"}}
	previous := []GeneratedSentence{
		{PromptRU: "Где отель?", ReferenceES: "¿Dónde está el hotel?"},
		{PromptRU: "Где метро?", ReferenceES: "¿Dónde está el metro?"},
	}
	cases := []struct {
		sentence GeneratedSentence
		good     bool
	}{
		{GeneratedSentence{PromptRU: "Привет, спасибо.", ReferenceES: "Hola, gracias.", UsedWords: []string{"hola", "gracias"}}, false},
		{GeneratedSentence{PromptRU: "Чай или кофе?", ReferenceES: "¿Té o café?", UsedWords: []string{"té", "café"}}, false},
		{GeneratedSentence{PromptRU: "Семь или восемь?", ReferenceES: "¿Siete u ocho?", UsedWords: []string{"siete", "ocho"}}, false},
		{GeneratedSentence{PromptRU: "Где терраса?", ClarificationRU: "Вы спрашиваете официанта о единственной террасе ресторана, в котором вы оба находитесь.", ReferenceES: "¿Dónde está la terraza?", UsedWords: []string{"terraza"}}, false},
		{GeneratedSentence{PromptRU: "Ты пьёшь воду или молоко?", ReferenceES: "¿Bebes agua o leche?", UsedWords: []string{"beber", "agua", "leche"}}, true},
		{GeneratedSentence{PromptRU: "Она читает.", ReferenceES: "Ella lee.", UsedWords: []string{"leer"}}, true},
	}
	sentences := make([]GeneratedSentence, 0, len(cases))
	for _, c := range cases {
		sentences = append(sentences, c.sentence)
	}
	result, err := svc.reviewGeneratedSentenceQuality(context.Background(), model, "es_ru", words, nil, []string{"presente (indicativo)"}, sentences, previous)
	if err != nil {
		t.Fatal(err)
	}
	accepted := map[string]bool{}
	for _, s := range result.Accepted {
		accepted[s.PromptRU] = true
	}
	for _, c := range cases {
		got := accepted[c.sentence.PromptRU]
		if got != c.good {
			t.Errorf("%s: accepted=%v want=%v", c.sentence.PromptRU, got, c.good)
		}
		t.Logf("DIVERSITY_CASE %s accepted=%v want=%v", c.sentence.PromptRU, got, c.good)
	}
	for _, r := range result.Rejected {
		t.Logf("REJECTION %s: %s", r.Sentence.PromptRU, r.Reason)
	}
	t.Logf("DIVERSITY_REVIEW calls=%d cost=%f", usage.Calls, usage.Cost)
}
