package service

import (
	"bytes"
	"context"
	"testing"
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
