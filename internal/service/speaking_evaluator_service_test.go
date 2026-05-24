package service

import "testing"

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
