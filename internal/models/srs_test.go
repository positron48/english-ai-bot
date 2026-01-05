package models

import (
	"testing"
)

func TestToSM2Quality(t *testing.T) {
	tests := []struct {
		name     string
		quality  Quality
		expected int
	}{
		{"QualityWrong", QualityWrong, 0},
		{"QualityHard", QualityHard, 3},
		{"QualityGood", QualityGood, 4},
		{"QualityEasy", QualityEasy, 5},
		{"Invalid", Quality(99), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.quality.ToSM2Quality()
			if result != tt.expected {
				t.Errorf("ToSM2Quality() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateQuality(t *testing.T) {
	tests := []struct {
		name     string
		data     AttemptData
		expected Quality
	}{
		{
			name: "Incorrect answer",
			data: AttemptData{
				Correct:      false,
				AnswerTimeMS: 1000,
			},
			expected: QualityWrong,
		},
		{
			name: "Correct with early reveal",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  true,
				AnswerTimeMS: 1000,
			},
			expected: QualityHard,
		},
		{
			name: "Correct but slow",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: 9000, // > SlowThresholdMS (8000)
			},
			expected: QualityHard,
		},
		{
			name: "Correct and fast",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: 2000, // < FastThresholdMS (2500)
			},
			expected: QualityEasy,
		},
		{
			name: "Correct normal speed",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: 5000, // Between FastThresholdMS and SlowThresholdMS
			},
			expected: QualityGood,
		},
		{
			name: "Correct exactly at fast threshold",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: FastThresholdMS,
			},
			expected: QualityGood,
		},
		{
			name: "Correct exactly at slow threshold",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: SlowThresholdMS,
			},
			expected: QualityGood,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateQuality(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateQuality() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLearningStepsDays(t *testing.T) {
	tests := []struct {
		name      string
		direction CardDirection
		expected  []int
	}{
		{
			name:      "RU to EN",
			direction: DirectionRUtoEN,
			expected:  []int{LearningStep0, LearningStep1, LearningStep2, LearningStep3},
		},
		{
			name:      "EN to RU",
			direction: DirectionENtoRU,
			expected:  []int{LearningStep0, LearningStep1, LearningStep2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LearningStepsDays(tt.direction)
			if len(result) != len(tt.expected) {
				t.Errorf("LearningStepsDays() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("LearningStepsDays()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
