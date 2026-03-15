package service

import (
	"testing"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupOptionsServiceTest(t *testing.T) (*OptionsService, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	svc := NewOptionsService(tcRepo, logger)

	return svc, db
}

func TestGetFallbackDistractors_RUtoEN_Verb(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionRUtoEN, "verb")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for verb")
	}

	// Check that all are verbs (common English verbs)
	expectedVerbs := []string{"make", "take", "get", "give", "come", "go"}
	found := false
	for _, d := range distractors {
		for _, ev := range expectedVerbs {
			if d == ev {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common verbs in distractors")
	}
}

func TestGetFallbackDistractors_RUtoEN_Noun(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionRUtoEN, "noun")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for noun")
	}

	// Check that all are nouns (common English nouns)
	expectedNouns := []string{"time", "person", "year", "way", "day"}
	found := false
	for _, d := range distractors {
		for _, en := range expectedNouns {
			if d == en {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common nouns in distractors")
	}
}

func TestGetFallbackDistractors_RUtoEN_Adjective(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionRUtoEN, "adjective")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for adjective")
	}

	// Check that all are adjectives (common English adjectives)
	expectedAdjs := []string{"good", "new", "first", "last", "long"}
	found := false
	for _, d := range distractors {
		for _, ea := range expectedAdjs {
			if d == ea {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common adjectives in distractors")
	}
}

func TestGetFallbackDistractors_RUtoEN_Default(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionRUtoEN, "unknown")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for unknown POS")
	}

	// Should contain all types
	hasVerb := false
	hasNoun := false
	hasAdj := false
	for _, d := range distractors {
		if d == "make" || d == "take" {
			hasVerb = true
		}
		if d == "time" || d == "person" {
			hasNoun = true
		}
		if d == "good" || d == "new" {
			hasAdj = true
		}
	}

	if !hasVerb || !hasNoun || !hasAdj {
		t.Error("Expected distractors to contain verbs, nouns, and adjectives for unknown POS")
	}
}

func TestGetFallbackDistractors_ENtoRU_Verb(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionENtoRU, "verb")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for verb")
	}

	// Check that all are Russian verbs
	expectedVerbs := []string{"делать", "брать", "получать", "давать"}
	found := false
	for _, d := range distractors {
		for _, ev := range expectedVerbs {
			if d == ev {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common Russian verbs in distractors")
	}
}

func TestGetFallbackDistractors_ENtoRU_Noun(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionENtoRU, "noun")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for noun")
	}

	// Check that all are Russian nouns
	expectedNouns := []string{"время", "человек", "год", "путь"}
	found := false
	for _, d := range distractors {
		for _, en := range expectedNouns {
			if d == en {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common Russian nouns in distractors")
	}
}

func TestGetFallbackDistractors_ENtoRU_Adjective(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionENtoRU, "adjective")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for adjective")
	}

	// Check that all are Russian adjectives
	expectedAdjs := []string{"хороший", "новый", "первый", "последний"}
	found := false
	for _, d := range distractors {
		for _, ea := range expectedAdjs {
			if d == ea {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected to find common Russian adjectives in distractors")
	}
}

func TestGetFallbackDistractors_ENtoRU_Default(t *testing.T) {
	service, _ := setupOptionsServiceTest(t)

	distractors := service.getFallbackDistractors(models.DirectionENtoRU, "")

	if len(distractors) == 0 {
		t.Error("Expected non-empty distractors for empty POS")
	}

	// Should contain all types
	hasVerb := false
	hasNoun := false
	hasAdj := false
	for _, d := range distractors {
		if d == "делать" || d == "брать" {
			hasVerb = true
		}
		if d == "время" || d == "человек" {
			hasNoun = true
		}
		if d == "хороший" || d == "новый" {
			hasAdj = true
		}
	}

	if !hasVerb || !hasNoun || !hasAdj {
		t.Error("Expected distractors to contain verbs, nouns, and adjectives for empty POS")
	}
}
