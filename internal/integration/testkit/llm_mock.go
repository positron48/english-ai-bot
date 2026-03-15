package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockLLMResponse is the raw content (JSON) the LLM returns for word definition
type MockLLMResponse struct {
	Lemma         string `json:"lemma"`
	POS           string `json:"pos"`
	Transcription string `json:"transcription"`
	DefinitionRU  string `json:"definition_ru"`
	InputWord     string `json:"input_word"`
	Examples      []struct {
		ExampleEN string `json:"example_en"`
		GlossRU   string `json:"gloss_ru"`
	} `json:"examples,omitempty"`
}

// openAIChatResponse mimics OpenAI API response
type openAIChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}
type openAIChatResponse struct {
	Choices []openAIChoice `json:"choices"`
}

// StartMockLLMServer starts an httptest.Server that returns fixed word JSON for /chat/completions
func StartMockLLMServer(t *testing.T, wordJSON string) *httptest.Server {
	t.Helper()
	content := strings.TrimSpace(wordJSON)
	if content == "" {
		content = `{"lemma":"test","pos":"noun","transcription":"test","definition_ru":"тест","input_word":"test"}`
	}
	body := openAIChatResponse{
		Choices: []openAIChoice{
			{Message: struct {
				Content string `json:"content"`
			}{Content: content}},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("mock LLM marshal: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))

	t.Cleanup(srv.Close)
	return srv
}

// DefaultWordJSON returns a valid WordInfoResponse JSON for testing
func DefaultWordJSON(lemma, definitionRU string) string {
	if lemma == "" {
		lemma = "testword"
	}
	if definitionRU == "" {
		definitionRU = "тестовое слово"
	}
	m := MockLLMResponse{
		Lemma:         lemma,
		POS:           "noun",
		Transcription: "ˈtest",
		DefinitionRU:  definitionRU,
		InputWord:     lemma,
		Examples: []struct {
			ExampleEN string `json:"example_en"`
			GlossRU   string `json:"gloss_ru"`
		}{
			{ExampleEN: "This is a test.", GlossRU: "Это тест."},
		},
	}
	b, _ := json.Marshal(m)
	return string(b)
}
