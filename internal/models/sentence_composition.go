package models

import "time"

// Sentence set status values.
const (
	SentenceSetReady     = "ready"     // generated, never opened
	SentenceSetStarted   = "started"   // user attempted at least one sentence
	SentenceSetCompleted = "completed" // every sentence attempted
)

// Sentence item outcomes (0 errors = star, 1 error = passed, 2+ = failed).
const (
	SentenceOutcomeStar   = "star"
	SentenceOutcomePassed = "passed"
	SentenceOutcomeFailed = "failed"
)

// OutcomeForErrorCount maps an LLM-reported error count to the scoring outcome.
func OutcomeForErrorCount(errors int) string {
	switch {
	case errors <= 0:
		return SentenceOutcomeStar
	case errors == 1:
		return SentenceOutcomePassed
	default:
		return SentenceOutcomeFailed
	}
}

// SentenceSet is one generated set of sentences for a user/course/date.
type SentenceSet struct {
	ID             int64
	UserID         int64
	CourseCode     string
	GenerationDate string // YYYY-MM-DD
	Scopes         []string
	Status         string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	StarCount      int
	PassedCount    int
	FailedCount    int
	CreatedAt      time.Time
}

// SentenceItem is a single sentence to translate, plus its single-shot attempt result.
type SentenceItem struct {
	ID              int64
	SetID           int64
	Position        int
	PromptRU        string
	ClarificationRU string // short, optional Russian context shown before answering
	ReferenceES     string
	WordCardIDs     []int64
	AttemptedAt     *time.Time
	UserInput       *string
	ErrorCount      *int
	Outcome         *string
	GradingJSON     *string // raw teacher-markup JSON, passed through to the client
}

// SentenceWordCandidate is a well-learned word eligible to seed sentence generation,
// ordered by least participation in this feature so far.
type SentenceWordCandidate struct {
	WordCardID  int64
	Lemma       string
	Translation string // native-language gloss
	UsedCount   int
}
