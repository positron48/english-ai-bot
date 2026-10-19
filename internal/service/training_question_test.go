package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/i18n"
)

func TestDailyJourneyWordQuestionUsesCurrentLocale(t *testing.T) {
	adapter := &DailyJourneyWordsAdapter{}
	for _, mode := range []string{"card", "spell", "type"} {
		for _, direction := range []string{"ru_en", "en_ru"} {
			raw, err := json.Marshal(journeyWordsState{Phase: "train", Card: map[string]any{"type": mode, "direction": direction, "question": "Что означает: old snapshot", "word_ru": "дом", "word_en": "casa", "display_word": "casa", "transcription": "kasa"}})
			if err != nil {
				t.Fatal(err)
			}
			step := &dailyjourney.Step{State: raw}
			for _, lang := range []string{"es", "en", "ru"} {
				view, err := adapter.Exercise(i18n.WithLanguage(context.Background(), lang), 1, &dailyjourney.Journey{CourseCode: "es_ru"}, step)
				if err != nil {
					t.Fatal(err)
				}
				question := view.(JourneyWordsExercise).Card["question"].(string)
				want := TrainingQuestion(lang, "es", mode, direction, "дом", "casa", "kasa")
				if question != want || strings.Contains(question, "old snapshot") || strings.Contains(question, "%!") {
					t.Fatalf("%s/%s/%s: %q", lang, mode, direction, question)
				}
				if lang == "es" && !strings.Contains(question, map[string]string{"card": "palabra", "spell": "Forma", "type": "Escribe"}[mode]) && direction != "ru_en" {
					t.Fatal(question)
				}
			}
		}
	}
}
