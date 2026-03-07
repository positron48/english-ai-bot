package testkit

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDefaultWordJSON(t *testing.T) {
	tests := []struct {
		name        string
		lemma       string
		definitionRU string
	}{
		{"empty_lemma", "", ""},
		{"empty_definition", "word", ""},
		{"both_set", "hello", "привет"},
		{"defaults", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultWordJSON(tt.lemma, tt.definitionRU)
			if got == "" {
				t.Error("DefaultWordJSON returned empty string")
			}
			var m MockLLMResponse
			if err := json.Unmarshal([]byte(got), &m); err != nil {
				t.Errorf("DefaultWordJSON returned invalid JSON: %v", err)
			}
			expLemma := tt.lemma
			if expLemma == "" {
				expLemma = "testword"
			}
			if m.Lemma != expLemma {
				t.Errorf("lemma = %q, want %q", m.Lemma, expLemma)
			}
			expDef := tt.definitionRU
			if expDef == "" {
				expDef = "тестовое слово"
			}
			if m.DefinitionRU != expDef {
				t.Errorf("definition_ru = %q, want %q", m.DefinitionRU, expDef)
			}
		})
	}
}

func TestStartMockLLMServer(t *testing.T) {
	wordJSON := DefaultWordJSON("test", "тест")
	srv := StartMockLLMServer(t, wordJSON)

	t.Run("post_chat_completions_ok", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/chat/completions", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(parsed.Choices) == 0 {
			t.Fatal("choices empty")
		}
		content := strings.TrimSpace(parsed.Choices[0].Message.Content)
		if content != wordJSON {
			t.Errorf("content = %q, want %q", content, wordJSON)
		}
	})

	t.Run("wrong_path_404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/other", nil)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("get_method_not_allowed", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/v1/chat/completions", nil)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", resp.StatusCode)
		}
	})
}

func TestStartMockLLMServer_EmptyWordJSON(t *testing.T) {
	srv := StartMockLLMServer(t, "")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/chat/completions", strings.NewReader("{}"))
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(parsed.Choices) == 0 {
		t.Fatal("choices empty")
	}
	// Empty input should yield default JSON with lemma "test"
	if !strings.Contains(parsed.Choices[0].Message.Content, `"lemma":"test"`) {
		t.Errorf("expected default lemma in content: %s", parsed.Choices[0].Message.Content)
	}
}
