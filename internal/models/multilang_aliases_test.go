package models

import (
	"testing"
)

func TestSyncWordCardNeutralAliases(t *testing.T) {
	dr := "ru def"
	de := "to run"
	c := &WordCard{
		Word:         "run",
		DefinitionRU: &dr,
		DisplayEN:    &de,
	}
	SyncWordCardNeutralAliases(c)
	if c.WordTarget != "run" {
		t.Errorf("WordTarget = %q", c.WordTarget)
	}
	if c.DefinitionNative != c.DefinitionRU || c.DefinitionNative == nil || *c.DefinitionNative != dr {
		t.Errorf("DefinitionNative pointer/value mismatch")
	}
	if c.DisplayTarget != c.DisplayEN {
		t.Errorf("DisplayTarget pointer mismatch")
	}
}

func TestNormalizeWordCardLegacyBeforeWrite(t *testing.T) {
	dn := "native def"
	dt := "disp"
	c := &WordCard{
		WordTarget:       "lemma",
		DefinitionNative: &dn,
		DisplayTarget:    &dt,
	}
	NormalizeWordCardLegacyBeforeWrite(c)
	if c.Word != "lemma" {
		t.Errorf("Word = %q", c.Word)
	}
	if c.DefinitionRU == nil || *c.DefinitionRU != dn {
		t.Errorf("DefinitionRU not filled from DefinitionNative")
	}
	if c.DisplayEN == nil || *c.DisplayEN != dt {
		t.Errorf("DisplayEN not filled from DisplayTarget")
	}
}

func TestSyncTrainingCardNeutralAliases(t *testing.T) {
	c := &TrainingCard{
		WordEN:    "go",
		WordRU:    "идти",
		MeaningEN: "move",
		ExampleEN: "I go.",
		ExampleRU: "Я иду.",
	}
	SyncTrainingCardNeutralAliases(c)
	if c.WordTarget != "go" || c.WordNative != "идти" || c.MeaningTarget != "move" || c.ExampleTarget != "I go." || c.ExampleNative != "Я иду." {
		t.Fatalf("neutral fields: %+v", c)
	}
}

func TestSyncTrainingCardResponseNeutralAliases(t *testing.T) {
	r := &TrainingCardResponse{
		WordEN: "test",
		Senses: []TrainingCardSense{
			{WordRU: "т", MeaningEN: "m", ExampleEN: "e", ExampleRU: "р"},
		},
	}
	SyncTrainingCardResponseNeutralAliases(r)
	if r.WordTarget != "test" {
		t.Errorf("WordTarget = %q", r.WordTarget)
	}
	s := r.Senses[0]
	if s.WordNative != "т" || s.MeaningTarget != "m" || s.ExampleTarget != "e" || s.ExampleNative != "р" {
		t.Fatalf("sense neutral: %+v", s)
	}
}

func TestSyncWordInfoResponseNeutralAliases(t *testing.T) {
	w := &WordInfoResponse{
		DefinitionRU: "old",
		Examples:     []WordInfoExample{{ExampleEN: "en", GlossRU: "ru"}},
	}
	SyncWordInfoResponseNeutralAliases(w)
	if w.DefinitionNative != "old" {
		t.Errorf("DefinitionNative = %q", w.DefinitionNative)
	}
	ex := w.Examples[0]
	if ex.ExampleTarget != "en" || ex.GlossNative != "ru" {
		t.Fatalf("example aliases: %+v", ex)
	}
}

func TestSyncTTSStatusNeutralAliases(t *testing.T) {
	s := &TTSGenerationStatus{Word: "hello"}
	SyncTTSStatusNeutralAliases(s)
	if s.WordTarget != "hello" {
		t.Errorf("WordTarget = %q", s.WordTarget)
	}
}

func TestSyncWordSetWordInfoNeutralAliases(t *testing.T) {
	ru := "ру"
	m := "mean"
	w := &WordSetWordInfo{
		Word:        "w",
		DisplayWord: "dw",
		WordRU:      &ru,
		MeaningEN:   &m,
	}
	SyncWordSetWordInfoNeutralAliases(w)
	if w.WordTarget != "w" || w.DisplayTarget != "dw" {
		t.Fatalf("word/display target")
	}
	if w.WordNative != w.WordRU || w.MeaningTarget != w.MeaningEN {
		t.Fatalf("pointer aliases")
	}
}
