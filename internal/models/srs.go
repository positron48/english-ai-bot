package models

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

	// Quality thresholds (in milliseconds)
	FastThresholdMS     = 2500
	SlowThresholdMS     = 15000 // 15 seconds - only answers longer than this count as "Hard" (was 8s)
	VerySlowThresholdMS = 30000 // 30 seconds - if answer time is longer, consider it as average (user was distracted)

	// Delay for showing options (in milliseconds)
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
	// Card (multiple choice) = 1.0. Spell (compose from letters) and type (full word) are harder.
	SpellTimeMultiplier         = 1.5  // compose word from letters
	TypeTimeMultiplierBase      = 1.2  // type full word
	TypeTimeMultiplierPerLetter = 0.08 // extra per letter (longer word = more time allowed)
	TypeTimeMultiplierCap       = 2.5  // max multiplier for type
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

// AttemptData holds data about a user's attempt at answering a card
type AttemptData struct {
	Correct      bool
	EarlyReveal  bool
	AnswerTimeMS int
	TDelayMS     int
	OptionCount  int
	ChosenOption string
	// TimeMultiplier scales quality thresholds: > 1 means more time is allowed (harder mode).
	// 0 or 1 = multiple choice; use TimeMultiplierForMode for spell/type.
	TimeMultiplier float64
}

// TimeMultiplierForMode returns the time multiplier for the given answer mode and word length.
// Spell = fixed 1.5; type = base + per-letter, capped.
func TimeMultiplierForMode(mode string, wordLen int) float64 {
	switch mode {
	case "spell":
		return SpellTimeMultiplier
	case "type":
		m := TypeTimeMultiplierBase + float64(wordLen)*TypeTimeMultiplierPerLetter
		if m > TypeTimeMultiplierCap {
			return TypeTimeMultiplierCap
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
// TimeMultiplier > 1 scales thresholds (harder modes: spell/type get more allowed time).
func CalculateQuality(data AttemptData) Quality {
	// If incorrect, always quality 0
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

	// If answer time is very large (scaled), consider it as average (QualityGood)
	if data.AnswerTimeMS > verySlowMS {
		return QualityGood
	}

	// Quality 1 (Hard) if: early reveal OR answer time was slow (scaled)
	if data.EarlyReveal || data.AnswerTimeMS > slowMS {
		return QualityHard
	}

	// Quality 3 (Easy) if: no early reveal AND answer was fast (scaled)
	if !data.EarlyReveal && data.AnswerTimeMS < fastMS {
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
