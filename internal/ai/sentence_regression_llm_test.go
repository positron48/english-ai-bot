package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type sentenceRegressionCase struct {
	Name, Course, Prompt, Context, Reference, Input string
	Errors                                          int
	Corrections                                     []string
}

// These are authored expectations, not another model's verdict. Live calls are opt-in.
var sentenceRegressionCases = []sentenceRegressionCase{
	{"ambiguous usted", "es_ru", "Вы пьёте воду.", "", "Vosotros bebéis agua.", "Usted bebe agua", 0, nil},
	{"ambiguous ustedes", "es_ru", "Вы пьёте воду.", "На улице жарко.", "Vosotros bebéis agua.", "Ustedes beben agua", 0, nil},
	{"ambiguous vosotros", "es_ru", "Вы пьёте воду.", "", "Usted bebe agua.", "Bebéis agua", 0, nil},
	{"formal singular", "es_ru", "Вы пьёте воду.", "Вы вежливо обращаетесь к одному незнакомому мужчине.", "Usted bebe agua.", "Vosotros bebéis agua", 2, []string{"Usted bebe agua", "Bebe agua"}},
	{"spain friends", "es_ru", "Вы пьёте воду.", "Разговор в Испании: вы обращаетесь к нескольким близким друзьям на ты.", "Vosotros bebéis agua.", "Bebéis agua", 0, nil},
	{"latin america friends", "es_ru", "Вы пьёте воду.", "Разговор в Мексике: вы обращаетесь к нескольким друзьям.", "Ustedes beben agua.", "Ustedes beben agua", 0, nil},
	{"they are not you", "es_ru", "Они едят апельсин.", "", "Ellos comen una naranja.", "Vosotros comen la naranja", 1, []string{"Ellos comen la naranja", "Comen la naranja"}},
	{"article no context", "es_ru", "Я читаю книгу.", "", "Leo un libro.", "Leo el libro", 0, nil},
	{"article irrelevant context", "es_ru", "Я читаю книгу.", "На улице идёт дождь.", "Leo un libro.", "Leo el libro", 0, nil},
	{"known book", "es_ru", "Я читаю книгу.", "Собеседник только что передал вам книгу и спросил, что вы с ней делаете.", "Leo el libro.", "Leo un libro", 1, []string{"Leo el libro"}},
	{"new book", "es_ru", "Я покупаю книгу.", "Вы ещё не выбрали книгу и хотите купить любую из тех, что есть в магазине.", "Compro un libro.", "Compro el libro", 1, []string{"Compro un libro"}},
	{"article gender", "es_ru", "Я читаю книгу.", "", "Leo un libro.", "Leo una libro", 1, []string{"Leo un libro"}},
	{"generic article", "es_ru", "Мне нравится музыка.", "", "Me gusta la música.", "Me gusta música", 1, []string{"Me gusta la música"}},
	{"comma optional pronoun", "es_ru", "Да, я пью воду.", "", "Sí, bebo agua.", "Sí yo bebo agua", 0, nil},
	{"question marks optional pronoun", "es_ru", "Ты пьёшь воду?", "", "¿Bebes agua?", "tú bebes agua", 0, nil},
	{"valid order", "es_ru", "Я покупаю красивый дом.", "Дом ещё не выбран: вы ищете любой красивый дом.", "Compro una casa bonita.", "Yo compro una bonita casa", 0, nil},
	{"verb agreement", "es_ru", "Мы пьём воду.", "", "Bebemos agua.", "Nosotros bebe agua", 1, []string{"Nosotros bebemos agua"}},
	{"meaning omission", "es_ru", "Я не пью молоко.", "", "No bebo leche.", "Bebo leche", 1, []string{"No bebo leche"}},
	{"meaning accent", "es_ru", "Он пьёт чай.", "", "Él bebe té.", "Él bebe te", 1, []string{"Él bebe té"}},
	{"spelling", "es_ru", "Я ем яблоко.", "Вы взяли первое попавшееся яблоко, раньше о нём не говорили.", "Como una manzana.", "Como una mansana", 1, []string{"Como una manzana"}},
	{"two errors", "es_ru", "Мы не пьём молоко.", "", "No bebemos leche.", "Nosotros bebe leche", 2, []string{"Nosotros no bebemos leche", "No nosotros bebemos leche"}},
	{"en contraction comma", "en_ru", "Нет, я не пью молоко.", "", "No, I do not drink milk.", "No I don't drink milk", 0, nil},
	{"en article ambiguous", "en_ru", "Я читаю книгу.", "На улице идёт дождь.", "I am reading a book.", "I am reading the book", 0, nil},
	{"en agreement", "en_ru", "Она пьёт воду.", "", "She drinks water.", "She drink water", 1, []string{"She drinks water"}},
}

// Held-out examples use different vocabulary and combinations from prompt tuning.
var sentenceHoldoutCases = []sentenceRegressionCase{
	{"house context not addressee", "es_ru", "Вы продаёте дом.", "Вы с собеседником говорите о доме, в котором вместе жили раньше.", "Vosotros vendéis la casa.", "Usted vende la casa", 0, nil},
	{"latin american plural form", "es_ru", "Вы читаете газету.", "В Мексике вы обращаетесь к нескольким друзьям на ты. Газету никто из собеседников прежде не видел и не обсуждал.", "Ustedes leen un periódico.", "Leéis un periódico", 1, []string{"Leen un periódico"}},
	{"women in spain", "es_ru", "Вы работаете.", "В Испании вы обращаетесь к двум подругам на ты; обе собеседницы — женщины.", "Vosotras trabajáis.", "Vosotros trabajáis", 1, []string{"Vosotras trabajáis"}},
	{"known key", "es_ru", "Я ищу ключ.", "Собеседник дал вам единственный ключ от комнаты; вы ищете именно его, чтобы вернуть владельцу.", "Busco la llave.", "Busco una llave", 1, []string{"Busco la llave"}},
	{"new dog", "es_ru", "Я хочу купить собаку.", "Вы ещё не выбрали животное, готовы купить любую собаку и впервые рассказываете о своём желании.", "Quiero comprar un perro.", "Quiero comprar el perro", 1, []string{"Quiero comprar un perro"}},
	{"partial context", "es_ru", "Я кладу книгу на стол.", "Стол — тот, который вы с собеседником только что собрали.", "Pongo un libro en la mesa.", "Pongo el libro en la mesa", 0, nil},
	{"accent changes noun", "es_ru", "Моя мама читает книги.", "", "Mi madre lee libros.", "Mi mama lee libros", 1, []string{"Mi mamá lee libros"}},
	{"correct synonym", "es_ru", "Моя мама читает книги.", "", "Mi madre lee libros.", "Mi mamá lee libros", 0, nil},
	{"two independent mistakes", "es_ru", "Мы не покупаем хлеб.", "", "No compramos pan.", "Nosotros compra pan", 2, []string{"Nosotros no compramos pan"}},
	{"en contraction", "en_ru", "Она здесь не работает.", "", "She does not work here.", "She doesn't work here", 0, nil},
	{"en known car", "en_ru", "Я вижу машину.", "Собеседник спрашивает, видите ли вы его машину, которую он только что показал вам; речь именно о ней.", "I see the car.", "I see a car", 1, []string{"I see the car"}},
	{"en article grammar", "en_ru", "Я ем яблоко.", "", "I am eating an apple.", "I am eating a apple", 1, []string{"I am eating an apple"}},
	{"en valid word order", "en_ru", "Я обычно пью чай утром.", "", "I usually drink tea in the morning.", "In the morning I usually drink tea", 0, nil},
	{"en negation present", "en_ru", "Они не читают книги.", "", "They do not read books.", "They don't read books", 0, nil},
}

func TestSentenceRegressionLLM(t *testing.T) {
	if os.Getenv("RUN_SENTENCE_REGRESSION_LLM") != "1" {
		t.Skip("opt-in paid cloud benchmark")
	}
	env, _ := godotenv.Read("../../.env")
	key := os.Getenv("POLZA_AI_API_KEY")
	if key == "" {
		key = env["POLZA_AI_API_KEY"]
	}
	if key == "" {
		t.Fatal("POLZA_AI_API_KEY required")
	}
	models := strings.Split(os.Getenv("SENTENCE_TEST_MODELS"), ",")
	if models[0] == "" {
		models = []string{"openai/gpt-4o-mini", "openai/gpt-5.4-nano", "openai/gpt-5.4-mini"}
	}
	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			t.Parallel()
			svc := NewServiceWithTimeout("https://polza.ai/api/v1", model, key, "", 90*time.Second, zap.NewNop())
			svc.SetSentenceReasoningEffort(os.Getenv("SENTENCE_TEST_REASONING_EFFORT"))
			for _, lang := range []string{"es", "en"} {
				path := "../../prompts/sentence-grade-ru-" + lang + ".txt"
				if dir := os.Getenv("SENTENCE_TEST_PROMPT_DIR"); dir != "" {
					path = filepath.Join(dir, filepath.Base(path))
				}
				prompt, err := LoadRenderedPromptFile(path, "ru", lang, "ru-"+lang)
				if err != nil {
					t.Fatal(err)
				}
				svc.SetSentenceGradePromptForCourse(lang+"_ru", prompt)
			}
			usage := observeSentenceUsage(t, svc)
			cases := sentenceRegressionCases
			if os.Getenv("SENTENCE_TEST_HOLDOUT") == "1" {
				cases = sentenceHoldoutCases
			}
			passed := 0
			start := time.Now()
			for _, tc := range cases {
				g, err := svc.GradeSentenceForCourse(context.Background(), tc.Course, tc.Prompt, tc.Context, tc.Reference, tc.Input)
				if err != nil {
					t.Errorf("%s: %v", tc.Name, err)
					if strings.Contains(err.Error(), "status 402") {
						break
					}
					continue
				}
				ok := g.ErrorCount == tc.Errors
				if tc.Errors == 0 {
					ok = ok && NormalizedSentenceAnswer(g.CorrectedES) == NormalizedSentenceAnswer(tc.Input) && g.Explanation == ""
				}
				if len(tc.Corrections) > 0 {
					found := false
					for _, c := range tc.Corrections {
						found = found || NormalizedSentenceAnswer(g.CorrectedES) == NormalizedSentenceAnswer(c)
					}
					ok = ok && found && sentenceExplanationLanguageMatches(tc.Prompt, g.Explanation) && g.Explanation != ""
				}
				raw, _ := json.Marshal(g)
				if ok {
					passed++
				} else {
					t.Errorf("case %s expected %d errors: %s", tc.Name, tc.Errors, raw)
				}
				t.Logf("CASE %s ok=%v %s", tc.Name, ok, raw)
			}
			t.Logf("SUMMARY model=%s pass=%d/%d calls=%d input=%d output=%d provider_cost=%f seconds=%.1f", model, passed, len(cases), usage.Calls, usage.Input, usage.Output, usage.Cost, time.Since(start).Seconds())
		})
	}
}

type sentenceCloudUsage struct {
	Calls, Input, Output int
	Cost                 float64
}

func observeSentenceUsage(t *testing.T, svc *Service) *sentenceCloudUsage {
	usage := &sentenceCloudUsage{}
	transport := svc.client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		usage.Calls++
		var requestBody []byte
		if os.Getenv("RUN_SENTENCE_GENERATION_LLM") == "1" && req.GetBody != nil {
			reader, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			requestBody, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				return nil, err
			}
		}
		resp, err := transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(strings.NewReader(string(body)))
		var accounting struct {
			Usage struct {
				PromptTokens     int     `json:"prompt_tokens"`
				CompletionTokens int     `json:"completion_tokens"`
				Cost             float64 `json:"cost"`
			} `json:"usage"`
		}
		_ = json.Unmarshal(body, &accounting)
		var metadata struct {
			Model    string `json:"model"`
			Provider string `json:"provider"`
			Choices  []struct {
				FinishReason string `json:"finish_reason"`
				Message      struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		_ = json.Unmarshal(body, &metadata)
		if len(metadata.Choices) > 0 && strings.TrimSpace(metadata.Choices[0].Message.Content) == "" {
			t.Logf("EMPTY model=%s provider=%s finish=%s usage=%+v", metadata.Model, metadata.Provider, metadata.Choices[0].FinishReason, accounting.Usage)
		}

		t.Logf("LLM_RESPONSE model=%s provider=%s status=%d input=%d output=%d cost=%f", metadata.Model, metadata.Provider, resp.StatusCode, accounting.Usage.PromptTokens, accounting.Usage.CompletionTokens, accounting.Usage.Cost)
		if dir := os.Getenv("SENTENCE_TEST_OUTPUT_DIR"); dir != "" && os.Getenv("RUN_SENTENCE_GENERATION_LLM") == "1" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			// Test payloads contain authored vocabulary, never production learner data or headers.
			snapshot, err := json.MarshalIndent(map[string]any{"request": json.RawMessage(requestBody), "response": json.RawMessage(body)}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("call-%02d.json", usage.Calls)), snapshot, 0644); err != nil {
				t.Fatal(err)
			}
		}
		usage.Input += accounting.Usage.PromptTokens
		usage.Output += accounting.Usage.CompletionTokens
		usage.Cost += accounting.Usage.Cost
		return resp, nil
	})
	return usage
}

func TestSentenceGenerationContextLLM(t *testing.T) {
	if os.Getenv("RUN_SENTENCE_GENERATION_LLM") != "1" {
		t.Skip("opt-in paid generation smoke")
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
		model = "google/gemini-3.8-flash"
	}
	course, lang := "es_ru", "es"
	if os.Getenv("SENTENCE_TEST_COURSE") == "en_ru" {
		course, lang = "en_ru", "en"
	}
	svc := NewServiceWithTimeout("https://polza.ai/api/v1", "unused-default", key, "", 90*time.Second, zap.NewNop())
	svc.SetSentenceModel(model)
	effort, effortSet := os.LookupEnv("SENTENCE_TEST_REASONING_EFFORT")
	if !effortSet {
		effort = "low"
	}
	svc.SetSentenceReasoningEffort(effort)
	prompt, err := LoadRenderedPromptFile("../../prompts/sentence-gen-ru-"+lang+".txt", "ru", lang, "ru-"+lang)
	if err != nil {
		t.Fatal(err)
	}
	// Exercise the reported ambiguity rather than hoping random generation uses вы.
	if os.Getenv("SENTENCE_TEST_REALISTIC") != "1" {
		prompt += "\nFor this validation set, make at least half the items address the learner as Russian вы; vary singular polite, plural polite and plural informal. Keep all the usual context requirements."
	}
	svc.SetSentenceGenPromptForCourse(course, prompt)
	gradePrompt, err := LoadRenderedPromptFile("../../prompts/sentence-grade-ru-"+lang+".txt", "ru", lang, "ru-"+lang)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetSentenceGradePromptForCourse(course, gradePrompt)
	words := sentenceTestWords("adversarial")
	tenses := []string{"presente (indicativo)"}
	if course == "en_ru" {
		words = []GenSentenceWord{{"book", "книга"}, {"friend", "друг"}, {"house", "дом"}, {"door", "дверь"}, {"apple", "яблоко"}, {"read", "читать"}, {"help", "помогать"}, {"buy", "покупать"}, {"open", "открывать"}, {"eat", "есть"}, {"water", "вода"}, {"drink", "пить"}, {"new", "новый"}, {"big", "большой"}, {"small", "маленький"}, {"close", "закрывать"}}
		tenses = []string{"present simple", "present continuous"}
	}
	count := atoiDefault(os.Getenv("SENTENCE_TEST_COUNT"), 20)
	usage := observeSentenceUsage(t, svc)
	start := time.Now()
	// Safe focus words still share a vocabulary pool with the adversarial body-part words.
	focus := []GenSentenceWord{{"libro", "книга"}, {"amigo", "друг"}, {"puerta", "дверь"}, {"ventana", "окно"}, {"agua", "вода"}}
	if course == "en_ru" {
		focus = words[:5]
	}
	if course == "es_ru" && os.Getenv("SENTENCE_TEST_WORD_PROFILE") == "holdout" {
		words = []GenSentenceWord{{"leche", "молоко"}, {"café", "кофе"}, {"pan", "хлеб"}, {"coche", "машина"}, {"tienda", "магазин"}, {"profesor", "учитель"}, {"mujer", "женщина"}, {"hermano", "брат"}, {"comprar", "покупать"}, {"vender", "продавать"}, {"beber", "пить"}, {"comer", "есть"}, {"trabajar", "работать"}, {"esperar", "ждать"}, {"necesitar", "нуждаться / требоваться"}, {"buscar", "искать"}, {"nuevo", "новый"}, {"viejo", "старый"}, {"caro", "дорогой"}, {"barato", "дешёвый"}}
		focus = words[:5]
	}
	sentences, err := svc.GenerateSentenceSetForCourse(context.Background(), course, focus, words, tenses, count)
	t.Logf("GENERATION model=%s course=%s count=%d calls=%d input=%d output=%d provider_cost=%f seconds=%.1f", model, course, len(sentences), usage.Calls, usage.Input, usage.Output, usage.Cost, time.Since(start).Seconds())
	if dir := os.Getenv("SENTENCE_TEST_OUTPUT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		report, _ := json.MarshalIndent(map[string]any{"model": model, "reasoning_effort": effort, "course": course, "wanted": count, "accepted": len(sentences), "calls": usage.Calls, "input_tokens": usage.Input, "output_tokens": usage.Output, "provider_cost": usage.Cost, "seconds": time.Since(start).Seconds()}, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, strings.ReplaceAll(model, "/", "-")+"-"+course+"-metrics.json"), report, 0644); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.MarshalIndent(sentences, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, strings.ReplaceAll(model, "/", "-")+"-"+course+".json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(sentences) == 0 || len(sentences) > count {
		t.Fatalf("got %d/%d", len(sentences), count)
	}
	if len(sentences) < count {
		t.Logf("PARTIAL reviewed set retained: %d/%d", len(sentences), count)
	}
	if usage.Calls > 4 {
		t.Errorf("shared-pool smoke exceeded four HTTP requests: %d", usage.Calls)
	}
	seen := map[string]bool{}
	for _, sentence := range sentences {
		key := NormalizedSentenceAnswer(sentence.PromptRU)
		if seen[key] {
			t.Errorf("duplicate Russian prompt: %s", sentence.PromptRU)
		}
		seen[key] = true
		if !usableSentenceClarification(sentence.ClarificationRU) {
			t.Errorf("unusable context: %s", sentence.ClarificationRU)
		}
	}
	for i, sentence := range sentences {
		raw, _ := json.Marshal(sentence)
		t.Logf("GENERATED %d %s", i+1, raw)
		if course == "es_ru" && hasRussianAddress(sentence.PromptRU) && strings.TrimSpace(sentence.ClarificationRU) == "" {
			t.Errorf("%d: missing address context", i+1)
		}
	}
	// New exercises must accept their reference, with presentation omitted.
	for i, sentence := range sentences {
		grade, err := svc.GradeSentenceForCourse(context.Background(), course, sentence.PromptRU, sentence.ClarificationRU, sentence.ReferenceES, NormalizedSentenceAnswer(sentence.ReferenceES))
		if err != nil || grade.Outcome != "star" {
			t.Fatalf("%d: reference grade=%+v err=%v", i+1, grade, err)
		}
	}
}
