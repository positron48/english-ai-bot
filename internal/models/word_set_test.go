package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidWordSetLevelCodes(t *testing.T) {
	for _, level := range []string{"A0", "A1", "A2", "B1", "B2", "C1"} {
		if !ValidWordSetLevelCodes[level] {
			t.Fatalf("expected valid level %q", level)
		}
	}
	if ValidWordSetLevelCodes["X9"] {
		t.Fatal("X9 should not be valid")
	}
}

func TestWordSetWordStatusConstants(t *testing.T) {
	if WordStatusUnknown != "unknown" || WordStatusInVocab != "in_vocab" || WordStatusKnown != "known" {
		t.Fatalf("status constants: %q %q %q", WordStatusUnknown, WordStatusInVocab, WordStatusKnown)
	}
}

func TestWordSetCategory_JSONRoundTrip(t *testing.T) {
	desc := "desc"
	level := "A1"
	parent := int64(5)
	cat := WordSetCategory{
		ID:          1,
		CourseCode:  "es_ru",
		ParentID:    &parent,
		Name:        "Basics",
		Description: &desc,
		IsPublished: true,
		SortOrder:   10,
		LevelCode:   &level,
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
	}
	raw, err := json.Marshal(cat)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded WordSetCategory
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.CourseCode != "es_ru" || decoded.Name != "Basics" || !decoded.IsPublished {
		t.Fatalf("decoded: %+v", decoded)
	}
}

func TestWordSetWithProgressAndWordInfo(t *testing.T) {
	ws := WordSetWithProgress{
		WordSet: WordSet{
			ID:         2,
			CourseCode: "en_ru",
			Title:      "Top words",
		},
		TotalWords:      100,
		KnownWords:      40,
		WordsInVocab:    30,
		UnknownWords:    30,
		ProgressPercent: 40,
	}
	if ws.ProgressPercent != 40 || ws.UnknownWords != 30 {
		t.Fatalf("progress: %+v", ws)
	}

	info := WordSetWordInfo{
		WordCardID:    1,
		Word:          "spy",
		WordTarget:    "spy",
		DisplayWord:   "spy",
		DisplayTarget: "spy",
		Status:        WordStatusInVocab,
	}
	if info.Status != WordStatusInVocab {
		t.Fatalf("status: %q", info.Status)
	}
}
