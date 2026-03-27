//go:build integration

package user_flows

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/integration/testkit"
)

// TestSpanishGrammarBundle_Sections lists grammar categories using embedded es bundle (GRAMMAR_BUNDLE_ID=es).
func TestSpanishGrammarBundle_Sections(t *testing.T) {
	t.Parallel()

	lc := config.LearningConfig{
		Pair:            "ru-es",
		NativeLang:      "ru",
		TargetLang:      "es",
		AppCode:         "spanish",
		GrammarBundleID: "es",
	}
	if err := config.ValidateLearningConfig(lc); err != nil {
		t.Fatalf("learning config: %v", err)
	}

	h := testkit.NewHarness(t, testkit.WithLearning(lc))
	token := h.AuthAsUser(424242)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("categories: %d %s", w.Code, w.Body.String())
	}

	var resp struct {
		Categories []struct {
			SectionID string `json:"section_id"`
		} `json:"categories"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	wantSection := "es.grammar.orientation_alphabet_sounds"
	var found bool
	for _, c := range resp.Categories {
		if c.SectionID == wantSection {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected section %s in categories, got %+v", wantSection, resp.Categories)
	}
}
