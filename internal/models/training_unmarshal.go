package models

import (
	"encoding/json"
)

// UnmarshalJSON accepts neutral keys (word_target, word_native, meaning_target, …) and legacy wire aliases (word_en, word_ru, meaning_en, …).
func (r *TrainingCardResponse) UnmarshalJSON(data []byte) error {
	type wire struct {
		Error         string            `json:"error"`
		WordEN        string            `json:"word_en"`
		WordTarget    string            `json:"word_target"`
		Lemma         string            `json:"lemma"`
		Transcription string            `json:"transcription"`
		Senses        []json.RawMessage `json:"senses"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	r.Error = w.Error
	r.WordEN = firstNonEmpty(w.WordEN, w.WordTarget)
	r.WordTarget = firstNonEmpty(w.WordTarget, w.WordEN)
	r.Lemma = w.Lemma
	r.Transcription = w.Transcription
	if len(w.Senses) == 0 {
		r.Senses = nil
		SyncTrainingCardResponseNeutralAliases(r)
		return nil
	}
	r.Senses = make([]TrainingCardSense, len(w.Senses))
	for i, raw := range w.Senses {
		if err := json.Unmarshal(raw, &r.Senses[i]); err != nil {
			return err
		}
	}
	SyncTrainingCardResponseNeutralAliases(r)
	return nil
}

// UnmarshalJSON accepts neutral keys and legacy wire aliases for one training-card sense.
func (s *TrainingCardSense) UnmarshalJSON(data []byte) error {
	type wire struct {
		POS           string   `json:"pos"`
		DisplayWord   string   `json:"display_word"`
		WordRU        string   `json:"word_ru"`
		WordNative    string   `json:"word_native"`
		MeaningEN     string   `json:"meaning_en"`
		MeaningTarget string   `json:"meaning_target"`
		ExampleEN     string   `json:"example_en"`
		ExampleTarget string   `json:"example_target"`
		ExampleRU     string   `json:"example_ru"`
		ExampleNative string   `json:"example_native"`
		DistractorsRU []string `json:"distractors_ru"`
		DistractorsEN []string `json:"distractors_en"`
		Hint          string   `json:"hint"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	s.POS = w.POS
	s.DisplayWord = w.DisplayWord
	s.WordRU = firstNonEmpty(w.WordRU, w.WordNative)
	s.WordNative = firstNonEmpty(w.WordNative, w.WordRU)
	s.MeaningEN = firstNonEmpty(w.MeaningEN, w.MeaningTarget)
	s.MeaningTarget = firstNonEmpty(w.MeaningTarget, w.MeaningEN)
	s.ExampleEN = firstNonEmpty(w.ExampleEN, w.ExampleTarget)
	s.ExampleTarget = firstNonEmpty(w.ExampleTarget, w.ExampleEN)
	s.ExampleRU = firstNonEmpty(w.ExampleRU, w.ExampleNative)
	s.ExampleNative = firstNonEmpty(w.ExampleNative, w.ExampleRU)
	s.DistractorsRU = w.DistractorsRU
	s.DistractorsEN = w.DistractorsEN
	s.Hint = w.Hint
	SyncTrainingCardSenseNeutralAliases(s)
	return nil
}
