package ai

import (
	"testing"

	"go.uber.org/zap"
)

func TestCourseScopedPromptGetters(t *testing.T) {
	logger := zap.NewNop()
	s := NewService("http://example.com", "model", "key", "default-dict", logger)
	s.SetTrainingPrompt("default-training")

	s.SetDictionaryPromptForCourse("es_ru", "dict-es\\nline2")
	s.SetTrainingPromptForCourse("es_ru", "train-es")
	s.SetConversationPromptForCourse("es_ru", "conv-legacy")
	s.SetConversationQuestPromptForCourse("es_ru", "quest-es")
	s.SetConversationCorrectionPromptForCourse("es_ru", "corr-es")
	s.SetConversationNPCPromptForCourse("es_ru", "npc-es")

	if got := s.trainingPromptForCourse("es_ru"); got != "train-es" {
		t.Fatalf("trainingPromptForCourse: got %q", got)
	}
	if got := s.trainingPromptForCourse("en_ru"); got != "default-training" {
		t.Fatalf("training fallback: got %q", got)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ConversationPromptForCourse", s.ConversationPromptForCourse("es_ru"), "conv-legacy"},
		{"ConversationQuestPromptForCourse", s.ConversationQuestPromptForCourse("es_ru"), "quest-es"},
		{"ConversationCorrectionPromptForCourse", s.ConversationCorrectionPromptForCourse("es_ru"), "corr-es"},
		{"ConversationNPCPromptForCourse", s.ConversationNPCPromptForCourse("es_ru"), "npc-es"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, tt.got, tt.want)
		}
	}

	if !s.HasSplitConversationPrompts("es_ru") {
		t.Fatal("expected split prompts to be complete for es_ru")
	}
	if s.HasSplitConversationPrompts("en_ru") {
		t.Fatal("en_ru should not have split prompts")
	}
	if s.ConversationNPCPromptForCourse("missing") != "" {
		t.Fatal("unknown course should return empty prompt")
	}
}

func TestSetDictionaryPromptForCourse_normalizesNewlines(t *testing.T) {
	s := NewService("http://example.com", "m", "k", "p", zap.NewNop())
	s.SetDictionaryPromptForCourse("es_ru", "line1\\nline2")
	if got := s.dictionaryPrompts["es_ru"]; got != "line1\nline2" {
		t.Fatalf("got %q", got)
	}
}
