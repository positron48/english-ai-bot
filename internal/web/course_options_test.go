package web

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestOptionsServiceForCourse_ExplicitSpanishCourseStripsToPrefix(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	// Linglow-like instance: server default target is English.
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "en", GrammarBundleID: "en"}}
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	optionsSvc := service.NewOptionsService(trainingCardRepo, logger, "en")
	router := NewRouter(logger, cfg, conn, nil, nil, optionsSvc, nil)

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(88001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.SelectCurrentCourse(context.Background(), user.ID, "en_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var wordCardID int64
	if err := conn.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "hablar", "hablar").Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	displayWord := "to hablar"
	pos := "verbo"
	distractorsEN, _ := json.Marshal([]string{"to comer", "to vivir", "to correr"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "hablar",
		SenseIndex:    0,
		WordRU:        "говорить",
		MeaningEN:     "hablar",
		POS:           &pos,
		DisplayWord:   &displayWord,
		DistractorsEN: string(distractorsEN),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{
			ID: cardID, WordCardID: wordCardID, WordEN: "hablar", WordRU: "говорить",
			POS: &pos, DisplayWord: &displayWord, DistractorsEN: string(distractorsEN),
		},
	}

	// Persisted course is en_ru → optionsServiceForUser would keep English "to " markers.
	legacyOptions, legacyCorrect, err := router.optionsServiceForUser(context.Background(), user.ID).GenerateOptions(userCard, 4, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateOptions legacy: %v", err)
	}
	if !strings.HasPrefix(legacyCorrect, "to ") {
		t.Fatalf("legacy correct answer = %q, want English to-prefix for en_ru user", legacyCorrect)
	}

	// Session/request course es_ru must override and strip "to " for Spanish verbs.
	options, correctAnswer, err := router.optionsServiceForCourse(context.Background(), user.ID, "es_ru").GenerateOptions(userCard, 4, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateOptions es_ru: %v", err)
	}
	if correctAnswer != "hablar" {
		t.Fatalf("correct answer = %q, want hablar", correctAnswer)
	}
	for _, o := range options {
		if strings.HasPrefix(o, "to ") {
			t.Fatalf("option %q must not use English to-prefix for es_ru course", o)
		}
	}
	_ = legacyOptions
}
