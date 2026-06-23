package service

import (
	"testing"

	"tgbot-skeleton/internal/models"
)

func TestNormalizeCyrillicHomoglyphs(t *testing.T) {
	resp := &models.TrainingCardResponse{
		Senses: []models.TrainingCardSense{{
			WordRU:        "травa",                 // trailing Latin 'a'
			WordNative:    "травa",                 // trailing Latin 'a'
			ExampleNative: "Это зелёная травa тут.", // Latin 'a' inside Cyrillic sentence
			DistractorsRU: []string{"дорогa", "ошибка", "software"},
		}},
	}
	normalizeCyrillicHomoglyphs(resp)
	s := resp.Senses[0]
	if containsLatin(s.WordRU) {
		t.Errorf("WordRU still has Latin: %q", s.WordRU)
	}
	if containsLatin(s.WordNative) {
		t.Errorf("WordNative still has Latin: %q", s.WordNative)
	}
	if containsLatin(s.ExampleNative) {
		t.Errorf("ExampleNative still has Latin: %q", s.ExampleNative)
	}
	if s.DistractorsRU[0] != "дорога" {
		t.Errorf("distractor[0] not fixed: %q", s.DistractorsRU[0])
	}
	// Fully-Latin distractor must be left untouched (so R2 still rejects it).
	if s.DistractorsRU[2] != "software" {
		t.Errorf("fully-Latin distractor should be untouched, got %q", s.DistractorsRU[2])
	}
}
