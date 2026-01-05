package service

import (
	"encoding/json"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestSRSService_RecordWrongAnswer_NewOption(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	card := &models.UserCard{
		WrongAnswersJSON: "",
	}

	err := service.RecordWrongAnswer(card, "wrong option")
	if err != nil {
		t.Fatalf("RecordWrongAnswer() error = %v", err)
	}

	// Verify wrong answer was recorded
	if card.WrongAnswersJSON == "" {
		t.Error("WrongAnswersJSON should be set after recording wrong answer")
	}

	var wrongAnswers []struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}
	err = json.Unmarshal([]byte(card.WrongAnswersJSON), &wrongAnswers)
	if err != nil {
		t.Fatalf("Failed to unmarshal wrong answers: %v", err)
	}
	if len(wrongAnswers) != 1 {
		t.Errorf("Expected 1 wrong answer, got %d", len(wrongAnswers))
	}
	if wrongAnswers[0].Option != "wrong option" {
		t.Errorf("Expected option 'wrong option', got %q", wrongAnswers[0].Option)
	}
	if wrongAnswers[0].Count != 1 {
		t.Errorf("Expected count 1, got %d", wrongAnswers[0].Count)
	}
}

func TestSRSService_RecordWrongAnswer_ExistingOption(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	// Create card with existing wrong answer
	existingWrongAnswers := []struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}{
		{Option: "wrong option", TS: time.Now(), Count: 2},
	}
	existingJSON, _ := json.Marshal(existingWrongAnswers)

	card := &models.UserCard{
		WrongAnswersJSON: string(existingJSON),
	}

	err := service.RecordWrongAnswer(card, "wrong option")
	if err != nil {
		t.Fatalf("RecordWrongAnswer() error = %v", err)
	}

	// Verify count was incremented
	var wrongAnswers []struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}
	err = json.Unmarshal([]byte(card.WrongAnswersJSON), &wrongAnswers)
	if err != nil {
		t.Fatalf("Failed to unmarshal wrong answers: %v", err)
	}
	if len(wrongAnswers) != 1 {
		t.Errorf("Expected 1 wrong answer, got %d", len(wrongAnswers))
	}
	if wrongAnswers[0].Count != 3 {
		t.Errorf("Expected count 3, got %d", wrongAnswers[0].Count)
	}
}

func TestSRSService_RecordWrongAnswer_Max10(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	// Create card with 10 wrong answers
	existingWrongAnswers := make([]struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}, 10)
	for i := range existingWrongAnswers {
		existingWrongAnswers[i] = struct {
			Option string    `json:"option"`
			TS     time.Time `json:"ts"`
			Count  int       `json:"count"`
		}{Option: "option" + string(rune('0'+i)), TS: time.Now(), Count: 1}
	}
	existingJSON, _ := json.Marshal(existingWrongAnswers)

	card := &models.UserCard{
		WrongAnswersJSON: string(existingJSON),
	}

	// Add one more
	err := service.RecordWrongAnswer(card, "new option")
	if err != nil {
		t.Fatalf("RecordWrongAnswer() error = %v", err)
	}

	// Verify only last 10 are kept
	var wrongAnswers []struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}
	err = json.Unmarshal([]byte(card.WrongAnswersJSON), &wrongAnswers)
	if err != nil {
		t.Fatalf("Failed to unmarshal wrong answers: %v", err)
	}
	if len(wrongAnswers) > 10 {
		t.Errorf("Expected at most 10 wrong answers, got %d", len(wrongAnswers))
	}
}
