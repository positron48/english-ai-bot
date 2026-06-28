package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestNormalizeSpeakingResult(t *testing.T) {
	r := &SpeakingEvaluationResult{
		MeaningScore:       6,
		GrammarScore:       0,
		PronunciationScore: 3,
		FluencyScore:       3,
		AudioQuality:       "clear",
	}
	normalizeSpeakingResult(r, 3)
	if r.MeaningScore != 5 || r.GrammarScore != 1 {
		t.Fatalf("clamp failed: %+v", r)
	}
	if !r.IsAcceptable {
		t.Fatal("expected acceptable")
	}
}

func TestNormalizeSpeakingResult_Unclear(t *testing.T) {
	r := &SpeakingEvaluationResult{
		MeaningScore: 5,
		AudioQuality: "unclear",
		IsAcceptable: true,
	}
	normalizeSpeakingResult(r, 3)
	if r.IsAcceptable {
		t.Fatal("unclear audio should not be acceptable")
	}
}

func TestStripSpeakingJSON(t *testing.T) {
	raw := "```json\n{\"meaning_score\":3}\n```"
	got := stripSpeakingJSON(raw)
	if got != `{"meaning_score":3}` {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeAudioFormat(t *testing.T) {
	if normalizeAudioFormat("audio/webm") != "webm" {
		t.Fatal("expected webm default parsing")
	}
	if normalizeAudioFormat("wav") != "wav" {
		t.Fatal("expected wav")
	}
}

func TestPrepareSpeakingAudioForModel_Passthrough(t *testing.T) {
	ctx := context.Background()
	raw := []byte{1, 2, 3}
	for _, format := range []string{"mp3", "wav"} {
		out, apiFormat, err := prepareSpeakingAudioForModel(ctx, raw, format, nil)
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		if apiFormat != format || !bytes.Equal(out, raw) {
			t.Fatalf("%s: got format=%s len=%d", format, apiFormat, len(out))
		}
	}
}

func TestNewSpeakingEvaluatorService(t *testing.T) {
	logger := zap.NewNop()
	if got := NewSpeakingEvaluatorService(nil, logger); got == nil {
		t.Fatal("nil config should still return service")
	}

	cfg := &config.Config{
		Speaking: config.SpeakingConfig{EvalTimeout: "45s"},
	}
	svc := NewSpeakingEvaluatorService(cfg, logger)
	if svc.client.Timeout != 45*time.Second {
		t.Fatalf("timeout: got %v want 45s", svc.client.Timeout)
	}
}

func TestSpeakingEvaluatorService_Enabled(t *testing.T) {
	logger := zap.NewNop()
	var nilSvc *SpeakingEvaluatorService
	if nilSvc.Enabled() {
		t.Fatal("nil receiver should not be enabled")
	}

	disabled := NewSpeakingEvaluatorService(&config.Config{}, logger)
	if disabled.Enabled() {
		t.Fatal("missing key should be disabled")
	}

	enabled := NewSpeakingEvaluatorService(&config.Config{
		Speaking: config.SpeakingConfig{Enabled: true, EvalAPIKey: "k"},
	}, logger)
	if !enabled.Enabled() {
		t.Fatal("expected enabled")
	}
}

func TestSpeakingEvaluatorService_AcceptMeaningScore(t *testing.T) {
	var nilSvc *SpeakingEvaluatorService
	if got := nilSvc.AcceptMeaningScore(); got != 3 {
		t.Fatalf("nil default: got %d", got)
	}

	svc := NewSpeakingEvaluatorService(&config.Config{
		Speaking: config.SpeakingConfig{AcceptMeaningScore: 4},
	}, zap.NewNop())
	if got := svc.AcceptMeaningScore(); got != 4 {
		t.Fatalf("got %d want 4", got)
	}

	zero := NewSpeakingEvaluatorService(&config.Config{
		Speaking: config.SpeakingConfig{AcceptMeaningScore: 0},
	}, zap.NewNop())
	if got := zero.AcceptMeaningScore(); got != 3 {
		t.Fatalf("zero should fall back to 3, got %d", got)
	}
}

func TestBuildSpeakingEvalPrompt(t *testing.T) {
	task := &repository.SpeakingTaskFull{
		SpeakingTaskDocument: repository.SpeakingTaskDocument{
			Type:           "repeat",
			Level:          "A1",
			TargetLanguage: "es",
			PromptRU:       "Скажи фразу",
			DisplayText:    "Hola",
		},
		ExpectedMeaningRU: "привет",
		AcceptableAnswers: []string{"hola", "buenos días"},
		EvaluationNotes:   "allow informal greeting",
	}
	prompt := buildSpeakingEvalPrompt(task, 2, "practice", 4)
	for _, want := range []string{
		"meaning_score>=4",
		"Attempt: 2, mode: practice",
		"Instruction RU: Скажи фразу",
		"Expected phrase to say: Hola",
		"Expected meaning RU: привет",
		"Acceptable variants: hola | buenos días",
		"Evaluator notes: allow informal greeting",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestNormalizeSpeakingResult_nil(t *testing.T) {
	normalizeSpeakingResult(nil, 3)
}

func TestNormalizeSpeakingResult_BelowAcceptScore(t *testing.T) {
	r := &SpeakingEvaluationResult{
		MeaningScore: 2,
		AudioQuality: "clear",
		IsAcceptable: false,
	}
	normalizeSpeakingResult(r, 3)
	if r.IsAcceptable {
		t.Fatal("meaning below threshold should not become acceptable")
	}
}

func TestNormalizeAudioFormat_allKnown(t *testing.T) {
	for _, format := range []string{"mp3", "webm", "ogg", "m4a"} {
		if got := normalizeAudioFormat(format); got != format {
			t.Fatalf("%s: got %q", format, got)
		}
	}
}

func TestSpeakingEvaluatorService_Evaluate(t *testing.T) {
	ctx := context.Background()
	task := &repository.SpeakingTaskFull{
		SpeakingTaskDocument: repository.SpeakingTaskDocument{
			Type:           "repeat",
			Level:          "A1",
			TargetLanguage: "es",
			DisplayText:    "Hola",
		},
	}
	logger := zap.NewNop()

	t.Run("disabled", func(t *testing.T) {
		svc := NewSpeakingEvaluatorService(&config.Config{}, logger)
		_, err := svc.Evaluate(ctx, task, []byte{1}, "wav", 1, "practice")
		if err == nil || !strings.Contains(err.Error(), "not enabled") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("empty audio", func(t *testing.T) {
		svc := NewSpeakingEvaluatorService(&config.Config{
			Speaking: config.SpeakingConfig{Enabled: true, EvalAPIKey: "k"},
		}, logger)
		_, err := svc.Evaluate(ctx, task, nil, "wav", 1, "practice")
		if err == nil || !strings.Contains(err.Error(), "empty audio") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("too large", func(t *testing.T) {
		svc := NewSpeakingEvaluatorService(&config.Config{
			Speaking: config.SpeakingConfig{Enabled: true, EvalAPIKey: "k", MaxAudioMB: 1},
		}, logger)
		_, err := svc.Evaluate(ctx, task, make([]byte, 2*1024*1024), "wav", 1, "practice")
		if err == nil || !strings.Contains(err.Error(), "too large") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		evalJSON := `{"understood_answer":"hola","meaning_score":4,"grammar_score":4,"pronunciation_score":3,"fluency_score":3,"is_acceptable":false,"audio_quality":"clear","short_feedback_ru":"ok","better_version":"Hola","repeat_task":""}`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
				t.Fatalf("auth: %q", auth)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "```json\n" + evalJSON + "\n```"}},
				},
			})
		}))
		t.Cleanup(srv.Close)

		svc := NewSpeakingEvaluatorService(&config.Config{
			Speaking: config.SpeakingConfig{
				Enabled:            true,
				EvalAPIKey:         "test-key",
				EvalBaseURL:        srv.URL,
				EvalModel:          "test-model",
				AcceptMeaningScore: 3,
			},
		}, logger)

		result, err := svc.Evaluate(ctx, task, []byte{1, 2, 3}, "wav", 1, "practice")
		if err != nil {
			t.Fatalf("Evaluate: %v", err)
		}
		if result.UnderstoodAnswer != "hola" || !result.IsAcceptable {
			t.Fatalf("result: %+v", result)
		}
	})

	t.Run("model error status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("bad gateway"))
		}))
		t.Cleanup(srv.Close)

		svc := NewSpeakingEvaluatorService(&config.Config{
			Speaking: config.SpeakingConfig{
				Enabled:     true,
				EvalAPIKey:  "k",
				EvalBaseURL: srv.URL,
			},
		}, logger)
		_, err := svc.Evaluate(ctx, task, []byte{1}, "wav", 1, "practice")
		if err == nil || !strings.Contains(err.Error(), "status 502") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("no choices", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
		}))
		t.Cleanup(srv.Close)

		svc := NewSpeakingEvaluatorService(&config.Config{
			Speaking: config.SpeakingConfig{Enabled: true, EvalAPIKey: "k", EvalBaseURL: srv.URL},
		}, logger)
		_, err := svc.Evaluate(ctx, task, []byte{1}, "wav", 1, "practice")
		if err == nil || !strings.Contains(err.Error(), "no choices") {
			t.Fatalf("got %v", err)
		}
	})
}

func TestPrepareSpeakingAudioForModel_ConvertsWebm(t *testing.T) {
	ctx := context.Background()
	called := false
	convert := func(ctx context.Context, audio []byte, format string) ([]byte, string, error) {
		called = true
		if format != "webm" {
			t.Fatalf("expected webm, got %s", format)
		}
		return []byte("mp3-bytes"), "mp3", nil
	}
	out, apiFormat, err := prepareSpeakingAudioForModel(ctx, []byte("webm"), "webm", convert)
	if err != nil || !called || apiFormat != "mp3" || string(out) != "mp3-bytes" {
		t.Fatalf("convert webm: called=%v format=%s out=%q err=%v", called, apiFormat, out, err)
	}
}
