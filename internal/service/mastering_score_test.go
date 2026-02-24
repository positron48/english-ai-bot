package service

import (
	"testing"
)

func TestComputeMasteringScore(t *testing.T) {
	tests := []struct {
		name          string
		total         int64
		correct       int64
		recentTotal   int64
		recentCorrect int64
		isKnown       bool
		wantMin       int
		wantMax       int
	}{
		{"known", 0, 0, 0, 0, true, 100, 100},
		{"no_reviews", 0, 0, 0, 0, false, 0, 0},
		{"all_correct_10", 10, 10, 10, 10, false, 95, 100},
		{"half_correct_20", 20, 10, 10, 10, false, 70, 80},
		{"one_correct_low_trust", 1, 1, 1, 1, false, 15, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeMasteringScore(tt.total, tt.correct, tt.recentTotal, tt.recentCorrect, tt.isKnown)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ComputeMasteringScore() = %d, want in [%d, %d]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
