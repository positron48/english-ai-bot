package service

import (
	"math"
)

// Mastering score formula constants (see docs/MASTERING_SCORE.md).
const (
	// MasteringScoreM is the minimum number of reviews for the word to trust the score (plan constant M).
	MasteringScoreM = 5
	// MasteringScoreW is the weight of recent accuracy in combined score, 0..1 (plan constant W).
	MasteringScoreW = 0.5
)

// ComputeMasteringScore returns 0–100 using the transparent formula:
// - isKnown => 100
// - total == 0 => 0
// - else: A = (1-W)*A_overall + W*A_recent, trust = min(1, total/M), score = round(100 * A * trust)
func ComputeMasteringScore(total, correct, recentTotal, recentCorrect int64, isKnown bool) int {
	if isKnown {
		return 100
	}
	if total <= 0 {
		return 0
	}
	aOverall := float64(correct) / float64(total)
	rt := recentTotal
	if rt <= 0 {
		rt = 1
	}
	aRecent := float64(recentCorrect) / float64(rt)
	trust := float64(total) / float64(MasteringScoreM)
	if trust > 1 {
		trust = 1
	}
	a := (1-MasteringScoreW)*aOverall + MasteringScoreW*aRecent
	score := 100 * a * trust
	return int(math.Round(score))
}
