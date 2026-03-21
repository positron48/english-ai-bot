package models

import (
	"encoding/json"
	"testing"
)

func TestTrainingCardResponse_UnmarshalJSON_neutralOnly(t *testing.T) {
	const payload = `{
  "word_target": "run",
  "lemma": "run",
  "transcription": "rʌn",
  "senses": [{
    "pos": "verb",
    "display_word": "to run",
    "word_native": "бежать",
    "meaning_target": "run",
    "example_target": "I run",
    "example_native": "Я бегу",
    "distractors_ru": ["a","b","c"],
    "distractors_en": ["to walk","to jump","to swim"],
    "hint": ""
  }],
  "error": ""
}`
	var r TrainingCardResponse
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.WordEN != "run" || r.WordTarget != "run" {
		t.Fatalf("word fields: WordEN=%q WordTarget=%q", r.WordEN, r.WordTarget)
	}
	if len(r.Senses) != 1 {
		t.Fatalf("senses: %d", len(r.Senses))
	}
	s := r.Senses[0]
	if s.WordRU != "бежать" || s.MeaningEN != "run" {
		t.Fatalf("sense legacy fields: WordRU=%q MeaningEN=%q", s.WordRU, s.MeaningEN)
	}
}

func TestTrainingCardResponse_UnmarshalJSON_legacyOnly(t *testing.T) {
	const payload = `{"word_en":"x","lemma":"x","transcription":"","senses":[{"pos":"n","word_ru":"икс","meaning_en":"x","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`
	var r TrainingCardResponse
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.WordTarget != "x" {
		t.Fatalf("WordTarget=%q", r.WordTarget)
	}
}
