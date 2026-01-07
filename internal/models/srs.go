package models

// SRS constants based on SM-2 algorithm
const (
	// Initial values
	InitialEF          = 2.5
	MinEF              = 1.3
	MaxEF              = 2.5
	
	// Learning steps in days (defaults, can be overridden by config)
	LearningStep0 = 1
	LearningStep1 = 3
	LearningStep2 = 7
	LearningStep3 = 14 // For RU->EN direction
	
	// Quality thresholds (in milliseconds)
	FastThresholdMS = 2500
	SlowThresholdMS = 8000
	
	// Delay for showing options (in milliseconds)
	OptionsDelayMS = 3000
	
	// Session limits
	DefaultMaxCardsPerSession = 1
	DefaultMaxNewPerSession   = 1
	
	// Options
	DefaultOptionCount = 4
	MaxOptionCount     = 6
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
	QualityWrong  Quality = 0 // Incorrect answer
	QualityHard   Quality = 1 // Correct but difficult
	QualityGood   Quality = 2 // Correct, normal
	QualityEasy   Quality = 3 // Correct and easy
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
	Correct       bool
	EarlyReveal   bool
	AnswerTimeMS  int
	TDelayMS      int
	OptionCount   int
	ChosenOption  string
}

// CalculateQuality calculates the quality score from attempt data
func CalculateQuality(data AttemptData) Quality {
	// If incorrect, always quality 0
	if !data.Correct {
		return QualityWrong
	}
	
	// If correct, determine quality based on timing and early reveal
	
	// Quality 1 (Hard) if:
	// - User clicked "show options" early OR
	// - Answer time was very slow
	if data.EarlyReveal || data.AnswerTimeMS > SlowThresholdMS {
		return QualityHard
	}
	
	// Quality 3 (Easy) if:
	// - No early reveal AND
	// - Answer time was fast
	if !data.EarlyReveal && data.AnswerTimeMS < FastThresholdMS {
		return QualityEasy
	}
	
	// Otherwise Quality 2 (Good)
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

