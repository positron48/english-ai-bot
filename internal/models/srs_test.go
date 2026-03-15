package models

import (
	"fmt"
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
				AnswerTimeMS: 16000, // > SlowThresholdMS (15s)
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
		{
			name: "Correct but very slow (> 30 seconds) - should be average",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: 35000, // > VerySlowThresholdMS (30000)
			},
			expected: QualityGood, // Should be average, not hard
		},
		{
			name: "Correct with early reveal but very slow (> 30 seconds) - should be average",
			data: AttemptData{
				Correct:      true,
				EarlyReveal:  true,
				AnswerTimeMS: 35000, // > VerySlowThresholdMS (30000)
			},
			expected: QualityGood, // Should be average, not hard (very slow overrides early reveal)
		},
		// Time multiplier: spell/type get scaled thresholds (same real time → less likely Hard)
		{
			name: "Correct 25s with multiplier 1.5 (spell) - scaled slow 22.5s, so Hard",
			data: AttemptData{
				Correct:        true,
				EarlyReveal:    false,
				AnswerTimeMS:   25000,
				TimeMultiplier: SpellTimeMultiplier,
			},
			expected: QualityHard,
		},
		{
			name: "Correct 25s with multiplier 2 (type long word) - scaled slow 30s, so Good",
			data: AttemptData{
				Correct:        true,
				EarlyReveal:    false,
				AnswerTimeMS:   25000,
				TimeMultiplier: 2.0,
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

func TestTimeMultiplierForMode(t *testing.T) {
	tests := []struct {
		mode     string
		wordLen  int
		expected float64
	}{
		{"", 0, 1.0},
		{"card", 5, 1.0},
		{"spell", 0, SpellTimeMultiplier},
		{"spell", 10, SpellTimeMultiplier},
		{"type", 0, TypeTimeMultiplierBase},
		{"type", 5, 1.6},                    // TypeTimeMultiplierBase + 5*TypeTimeMultiplierPerLetter
		{"type", 20, TypeTimeMultiplierCap}, // capped
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_len%d", tt.mode, tt.wordLen), func(t *testing.T) {
			got := TimeMultiplierForMode(tt.mode, tt.wordLen)
			diff := got - tt.expected
			if diff < -1e-9 || diff > 1e-9 {
				t.Errorf("TimeMultiplierForMode(%q, %d) = %v, want %v", tt.mode, tt.wordLen, got, tt.expected)
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
