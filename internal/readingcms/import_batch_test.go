package readingcms

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestParseImportJSONDocuments(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single object",
			input:     `{"title_short":"Uno","level":"A1","segments":[]}`,
			wantCount: 1,
		},
		{
			name:      "array",
			input:     `[{"title_short":"Uno","level":"A1","segments":[]},{"title_short":"Dos","level":"A2","segments":[]}]`,
			wantCount: 2,
		},
		{
			name:      "consecutive objects",
			input:     `{"title_short":"Uno","level":"A1","segments":[]} {"title_short":"Dos","level":"A2","segments":[]}`,
			wantCount: 2,
		},
		{
			name:      "malformed second object",
			input:     `{"title_short":"Uno","level":"A1","segments":[]} {bad`,
			wantCount: 1,
			wantErr:   true,
		},
		{
			name:    "empty input",
			input:   " \n\t ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := parseImportJSONDocuments(tt.input)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(docs) != tt.wantCount {
				t.Fatalf("docs=%d want=%d", len(docs), tt.wantCount)
			}
			for i, doc := range docs {
				if !json.Valid(doc) {
					t.Fatalf("doc %d is invalid: %s", i, string(doc))
				}
			}
		})
	}
}

func TestImportJSONBatchEndpoint(t *testing.T) {
	f := setupGenFixture(t)
	handler := NewServer(f.svc, "").Handler()

	body := map[string]any{
		"course_code": "en_ru",
		"level":       "A2",
		"format":      "dialogue",
		"documents_text": `[{
			"title_short":"Batch One",
			"level":"A2",
			"segments":[{"segment_id":"s1","speaker_id":"speaker_a","text":"Hi","text_translation_ru":"Привет"}]
		},{
			"title_short":"Batch Two",
			"level":"A2",
			"segments":[{"segment_id":"s1","speaker_id":"speaker_a","text":"Hello","text_translation_ru":"Здравствуйте"}]
		}]`,
		"with_audio":   false,
		"auto_publish": true,
		"sync_bundle":  true,
	}
	rec := doRequest(t, handler, http.MethodPost, "/api/drafts/import-json-batch", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp ImportJSONBatchResponse
	decodeJSONBody(t, rec, &resp)
	if resp.Total != 2 || resp.Succeeded != 2 || resp.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", resp)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("results=%d", len(resp.Results))
	}
	for _, result := range resp.Results {
		if result.Error != "" {
			t.Fatalf("unexpected item error: %+v", result)
		}
		if result.Draft == nil || result.Draft.LastJobLog != "JSON imported; run Generate audio before publish" {
			t.Fatalf("with_audio flag not reflected in draft: %+v", result.Draft)
		}
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/import-json-batch", map[string]any{
		"course_code": "en_ru",
		"level":       "A2",
		"documents_text": `{"title_short":"Good","level":"A2","segments":[{"segment_id":"s1","text":"Hi"}]}
{bad`,
		"with_audio": false,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("partial status=%d body=%s", rec.Code, rec.Body.String())
	}
	decodeJSONBody(t, rec, &resp)
	if resp.Total != 2 || resp.Succeeded != 1 || resp.Failed != 1 {
		t.Fatalf("partial summary: %+v", resp)
	}
	if !strings.Contains(resp.Results[1].Error, "invalid character") {
		t.Fatalf("parse error not reported as item failure: %+v", resp.Results[1])
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/import-json-batch", map[string]any{
		"course_code":    "en_ru",
		"level":          "A2",
		"documents_text": " ",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, handler, http.MethodGet, "/api/drafts/import-json-batch", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status=%d", rec.Code)
	}
}
