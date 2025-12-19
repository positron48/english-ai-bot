package models

import "time"

// TrainingCard represents a training card with one sense of a word
type TrainingCard struct {
	ID            int64
	WordCardID    int64
	WordEN        string
	Transcription string
	SenseIndex    int
	WordRU        string
	MeaningEN     string
	ExampleEN     string
	ExampleRU     string
	DistractorsRU string // JSON array
	DistractorsEN string // JSON array
	Hint          string
	CreatedAt     time.Time
}

// CardDirection represents the direction of a card (RU→EN or EN→RU)
type CardDirection string

const (
	DirectionRUtoEN CardDirection = "ru_en"
	DirectionENtoRU CardDirection = "en_ru"
)

// CardState represents the current state of a user card in SRS
type CardState string

const (
	StateNew      CardState = "new"
	StateLearning CardState = "learning"
	StateReview   CardState = "review"
)

// UserCard represents a user's progress on a training card
type UserCard struct {
	ID              int64
	UserID          int64
	TrainingCardID  int64
	Direction       CardDirection
	State           CardState
	EF              float64 // Easiness Factor
	Reps            int     // Number of successful reviews
	IntervalDays    int
	LearningStep    int
	LapseCount      int
	NextDueAt       *time.Time
	LastReviewAt    *time.Time
	LastQuality     *int
	LastOptionsJSON string // JSON array of last shown options
	WrongAnswersJSON string // JSON array of wrong answers history
	StatsJSON       string // JSON object with stats
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SessionSource represents how a training session was started
type SessionSource string

const (
	SourceNudge  SessionSource = "nudge"
	SourceManual SessionSource = "manual"
)

// TrainingSession represents a training session
type TrainingSession struct {
	ID           int64
	UserID       int64
	StartedAt    time.Time
	EndedAt      *time.Time
	Source       SessionSource
	PlannedCount int
	DoneCount    int
	SessionJSON  string // JSON with session config
}

// ReviewEvent represents a single card review attempt
type ReviewEvent struct {
	ID             int64
	SessionID      *int64
	UserID         int64
	UserCardID     int64
	Direction      CardDirection
	ShownAt        time.Time
	OptionsShownAt *time.Time
	AnsweredAt     *time.Time
	TDelayMS       int
	EarlyReveal    bool
	OptionCount    int
	OptionsJSON    string
	ChosenOption   string
	IsCorrect      bool
	Quality        int
	MetricsJSON    string
	SRSBeforeJSON  string
	SRSAfterJSON   string
}

// TrainingNudge represents a daily training notification
type TrainingNudge struct {
	ID             int64
	UserID         int64
	LocalDate      string
	SentAt         time.Time
	ConsumedAt     *time.Time
	DueCountAtSend int
	MessageID      *int
}

// CircuitBreakerState represents the circuit breaker state
type CircuitBreakerState struct {
	ID                 int64
	IsOpen             bool
	FailureCount       int
	LastFailureAt      *time.Time
	LastFailureMessage string
	LastResetAt        *time.Time
	UpdatedAt          time.Time
}

// TrainingCardSense represents one sense/meaning in LLM response
type TrainingCardSense struct {
	Index         int      `json:"index"`
	WordRU        string   `json:"word_ru"`
	MeaningEN     string   `json:"meaning_en"`
	ExampleEN     string   `json:"example_en"`
	ExampleRU     string   `json:"example_ru"`
	DistractorsRU []string `json:"distractors_ru"`
	DistractorsEN []string `json:"distractors_en"`
	Hint          string   `json:"hint"`
}

// TrainingCardResponse represents LLM response for training card generation
type TrainingCardResponse struct {
	WordEN        string                `json:"word_en"`
	Transcription string                `json:"transcription"`
	Senses        []TrainingCardSense   `json:"senses"`
}

// UserCardWithTraining combines UserCard with TrainingCard data for display
type UserCardWithTraining struct {
	UserCard     UserCard
	TrainingCard TrainingCard
}

