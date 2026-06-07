package models

import "time"

// SRS constants based on SM-2 algorithm
const (
	// Initial values
	InitialEF = 2.5
	MinEF     = 1.3
	MaxEF     = 2.5

	// Learning steps in days (defaults, can be overridden by config)
	LearningStep0 = 1
	LearningStep1 = 3
	LearningStep2 = 7
	LearningStep3 = 14 // For RU->EN direction

	// Quality thresholds (in milliseconds). AnswerTimeMS = time from options shown to answer (delay before options not counted).
	// Bands: < Fast → Easy(3), Fast..Slow → Good(2), Slow..VerySlow → Hard(1), > VerySlow → Good (distracted).
	FastThresholdMS     = 2500  // < 2.5 s → Easy
	SlowThresholdMS     = 15000 // > 15 s → Hard (max "slow" threshold)
	VerySlowThresholdMS = 30000 // > 30 s → treat as Good

	// Delay for showing options (in milliseconds); configurable via options_delay_ms. Not used in quality.
	OptionsDelayMS = 3000

	// Session limits
	DefaultMaxCardsPerSession = 30
	DefaultMaxNewPerSession   = 30
	// MaxDuePoolSize limits how many due cards we fetch when building the session pool (random sample is taken from this)
	MaxDuePoolSize = 2000

	// Options
	DefaultOptionCount = 4
	MaxOptionCount     = 6

	// Time multipliers for quality thresholds by answer mode (harder modes get more allowed time)
	// Card (multiple choice) = 1.0. Spell and type use the same formula: base + per-letter, capped.
	SpellTypeTimeMultiplierBase      = 1.2  // spell / type
	SpellTypeTimeMultiplierPerLetter = 0.12 // extra per letter (longer word = more time allowed)
	SpellTypeTimeMultiplierCap       = 2.5  // max multiplier
)

// LearningStepsDays returns the learning steps in days for a given direction
// RU->EN is harder (active recall), so it has more steps
// EN->RU is easier (passive recognition), so it has fewer steps
func LearningStepsDays(direction CardDirection) []int {
	if direction == DirectionRUtoEN {
		// RU->EN: [1, 3, 7, 14] - more gradual progression
		return []int{LearningStep0, LearningStep1, LearningStep2, LearningStep3}
	}
	// EN->RU: [1, 3, 7] - faster progression
	return []int{LearningStep0, LearningStep1, LearningStep2}
}

// Quality represents the internal quality score (0-3)
type Quality int

const (
	QualityWrong Quality = 0 // Incorrect answer
	QualityHard  Quality = 1 // Correct but difficult
	QualityGood  Quality = 2 // Correct, normal
	QualityEasy  Quality = 3 // Correct and easy
)

// ToSM2Quality converts internal quality (0-3) to SM-2 quality (0-5)
func (q Quality) ToSM2Quality() int {
	switch q {
	case QualityWrong:
		return 0
	case QualityHard:
		return 3
	case QualityGood:
		return 4
	case QualityEasy:
		return 5
	default:
		return 0
	}
}

// AttemptData holds data about a user's attempt at answering a card.
// AnswerTimeMS = ms from options shown to answer; delay before options is NOT counted (only time to choose).
type AttemptData struct {
	Correct      bool
	EarlyReveal  bool   // kept for metrics/logging only, not used in quality
	AnswerTimeMS int    // ms from options shown to answer
	TDelayMS     int
	OptionCount  int
	ChosenOption string
	// TimeMultiplier scales quality thresholds: > 1 means more time is allowed (harder mode).
	// 0 or 1 = multiple choice; use TimeMultiplierForMode for spell/type.
	TimeMultiplier float64
	// GradedAt, when set, is used for SRS scheduling timestamps (offline sync).
	GradedAt *time.Time
}

// TimeMultiplierForMode returns the time multiplier for the given answer mode and word length.
// Spell and type use the same formula: base + per-letter, capped. Card = 1.0.
func TimeMultiplierForMode(mode string, wordLen int) float64 {
	switch mode {
	case "spell", "type":
		m := SpellTypeTimeMultiplierBase + float64(wordLen)*SpellTypeTimeMultiplierPerLetter
		if m > SpellTypeTimeMultiplierCap {
			return SpellTypeTimeMultiplierCap
		}
		return m
	default:
		return 1.0
	}
}

// effectiveThreshold returns threshold * multiplier; multiplier 0 or 1 means no scaling
func effectiveThreshold(baseMS int, mult float64) int {
	if mult <= 0 || mult <= 1.0 {
		return baseMS
	}
	return int(float64(baseMS) * mult)
}

// CalculateQuality calculates the quality score from attempt data.
// Only Correct and AnswerTimeMS (and TimeMultiplier) are used. EarlyReveal is not used.
// Time bands (with multiplier 1): < 2.5s Easy, 2.5–15s Good, 15–30s Hard, > 30s Good.
func CalculateQuality(data AttemptData) Quality {
	if !data.Correct {
		return QualityWrong
	}

	mult := data.TimeMultiplier
	if mult < 1.0 {
		mult = 1.0
	}
	fastMS := effectiveThreshold(FastThresholdMS, mult)
	slowMS := effectiveThreshold(SlowThresholdMS, mult)
	verySlowMS := effectiveThreshold(VerySlowThresholdMS, mult)

	if data.AnswerTimeMS > verySlowMS {
		return QualityGood
	}
	if data.AnswerTimeMS > slowMS {
		return QualityHard
	}
	if data.AnswerTimeMS < fastMS {
		return QualityEasy
	}
	return QualityGood
}

// SRSState represents the state of a card for SRS calculations
type SRSState struct {
	State        CardState
	EF           float64
	Reps         int
	IntervalDays int
	LearningStep int
	LapseCount   int
}
