package web

import "testing"

func TestUserCardStateRank(t *testing.T) {
	tests := []struct {
		state string
		want  int
	}{
		{"new", 0},
		{"learning", 1},
		{"review", 2},
		{"unknown", 1},
	}
	for _, tc := range tests {
		if got := userCardStateRank(tc.state); got != tc.want {
			t.Fatalf("userCardStateRank(%q) = %d, want %d", tc.state, got, tc.want)
		}
	}
}

func TestVocabWordSummaryBucket(t *testing.T) {
	tests := []struct {
		name        string
		minCardRank int
		isKnown     bool
		want        int
	}{
		{"new only", 0, false, 0},
		{"learning only", 1, false, 1},
		{"review only", 2, false, 2},
		{"known with new card", 0, true, 0},
		{"known with learning card", 1, true, 1},
		{"known all review", 2, true, 3},
		{"known without cards defaults review bucket", 2, true, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := vocabWordSummaryBucket(tc.minCardRank, tc.isKnown); got != tc.want {
				t.Fatalf("vocabWordSummaryBucket(%d, %v) = %d, want %d", tc.minCardRank, tc.isKnown, got, tc.want)
			}
		})
	}
}
