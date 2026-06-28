package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTrainingModelConstants(t *testing.T) {
	if DirectionRUtoEN != "ru_en" || DirectionENtoRU != "en_ru" {
		t.Fatalf("card directions: %q %q", DirectionRUtoEN, DirectionENtoRU)
	}
	if StateNew != "new" || StateLearning != "learning" || StateReview != "review" {
		t.Fatalf("card states: %q %q %q", StateNew, StateLearning, StateReview)
	}
	if SourceNudge != "nudge" || SourceManual != "manual" {
		t.Fatalf("session sources: %q %q", SourceNudge, SourceManual)
	}
}

func TestTrainingCard_JSONRoundTrip(t *testing.T) {
	pos := "verb"
	display := "to run"
	card := TrainingCard{
		ID:            1,
		WordCardID:    2,
		WordEN:        "run",
		WordTarget:    "run",
		Transcription: "rʌn",
		SenseIndex:    0,
		WordRU:        "бежать",
		WordNative:    "бежать",
		MeaningEN:     "run",
		MeaningTarget: "run",
		ExampleEN:     "I run",
		ExampleTarget: "I run",
		ExampleRU:     "Я бегу",
		ExampleNative: "Я бегу",
		DistractorsRU: `["идти","прыгать","плавать"]`,
		DistractorsEN: `["walk","jump","swim"]`,
		Hint:          "",
		POS:           &pos,
		DisplayWord:   &display,
		CreatedAt:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	raw, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded TrainingCard
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.WordTarget != "run" || decoded.WordNative != "бежать" {
		t.Fatalf("decoded: %+v", decoded)
	}
	if decoded.POS == nil || *decoded.POS != "verb" {
		t.Fatalf("POS: %+v", decoded.POS)
	}
}

func TestUserCardAndReviewEventFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	due := now.Add(24 * time.Hour)
	quality := 4
	uc := UserCard{
		ID:             10,
		UserID:         20,
		TrainingCardID: 30,
		Direction:      DirectionENtoRU,
		State:          StateReview,
		EF:             2.5,
		Reps:           3,
		IntervalDays:   7,
		NextDueAt:      &due,
		LastQuality:    &quality,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if uc.Direction != DirectionENtoRU || uc.State != StateReview {
		t.Fatalf("user card enums: %+v", uc)
	}

	ev := ReviewEvent{
		ID:              1,
		UserID:          20,
		UserCardID:      10,
		ClientAttemptID: "attempt-1",
		Direction:       DirectionRUtoEN,
		ShownAt:         now,
		IsCorrect:       true,
		Quality:         5,
	}
	if ev.Direction != DirectionRUtoEN {
		t.Fatalf("review direction: %q", ev.Direction)
	}
}
