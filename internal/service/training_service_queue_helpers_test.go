package service

import (
	"testing"

	"tgbot-skeleton/internal/repository"
)

func Test_computeMasteringScore(t *testing.T) {
	tests := []struct {
		name   string
		stats  *repository.WordMasteringStats
		expect int
	}{
		{"known", &repository.WordMasteringStats{IsKnown: true}, 100},
		{"review_all_reps", &repository.WordMasteringStats{
			TotalCards:        2,
			ReviewStateCount:  2,
			LearningStateCount: 0,
			NewStateCount:     0,
			TotalReps:         50,
			IsKnown:           false,
		}, 95}, // 75 + min(20, 25) = 95
		{"review_all_reps_cap", &repository.WordMasteringStats{
			TotalCards:        2,
			ReviewStateCount:  2,
			LearningStateCount: 0,
			NewStateCount:     0,
			TotalReps:         100,
			IsKnown:           false,
		}, 95}, // 75 + 20 cap
		{"learning_and_review", &repository.WordMasteringStats{
			TotalCards:        4,
			ReviewStateCount:  2,
			LearningStateCount: 2,
			NewStateCount:     0,
			TotalReps:         0,
			IsKnown:           false,
		}, 50}, // 25 + 25
		{"only_new", &repository.WordMasteringStats{
			TotalCards:        2,
			ReviewStateCount:  0,
			LearningStateCount: 0,
			NewStateCount:     2,
			TotalReps:         0,
			IsKnown:           false,
		}, 0},
		{"total_zero_but_has_learning", &repository.WordMasteringStats{
			TotalCards:        0,
			ReviewStateCount:  0,
			LearningStateCount: 1,
			NewStateCount:     0,
			TotalReps:         0,
			IsKnown:           false,
		}, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeMasteringScore(tt.stats)
			if got != tt.expect {
				t.Errorf("computeMasteringScore() = %d, want %d", got, tt.expect)
			}
		})
	}
}

func Test_spellPrefixAndLetters(t *testing.T) {
	tests := []struct {
		word       string
		wantPrefix string
		wantLen    int // length of letters slice
	}{
		{"to spy", "to ", 3},
		{"to run", "to ", 3},
		{"hello", "", 5},
		{"to ", "", 3}, // len("to ") is 3, not > 3, so no prefix; letters from full string
		{"a", "", 0},
	}
	for _, tt := range tests {
		prefix, letters := spellPrefixAndLetters(tt.word)
		if prefix != tt.wantPrefix {
			t.Errorf("spellPrefixAndLetters(%q) prefix = %q, want %q", tt.word, prefix, tt.wantPrefix)
		}
		if len(letters) != tt.wantLen {
			t.Errorf("spellPrefixAndLetters(%q) letters len = %d, want %d", tt.word, len(letters), tt.wantLen)
		}
	}
}

func Test_shuffleLetters(t *testing.T) {
	// shuffleLetters is non-deterministic; we only check length and that runes are preserved
	word := "hello"
	letters := shuffleLetters(word)
	if len(letters) != 5 {
		t.Errorf("shuffleLetters(%q) len = %d, want 5", word, len(letters))
	}
	runes := make(map[rune]int)
	for _, r := range word {
		runes[r]++
	}
	for _, s := range letters {
		for _, r := range s {
			runes[r]--
		}
	}
	for r, c := range runes {
		if c != 0 {
			t.Errorf("shuffleLetters: rune %q count diff %d", r, c)
		}
	}

	// short word returns nil
	if got := shuffleLetters("a"); got != nil {
		t.Errorf("shuffleLetters(\"a\") = %v, want nil", got)
	}
	if got := shuffleLetters(""); got != nil {
		t.Errorf("shuffleLetters(\"\") = %v, want nil", got)
	}
}
