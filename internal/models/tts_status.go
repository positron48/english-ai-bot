package models

import "time"

const (
	TTSStatePending         = "pending"
	TTSStateReady           = "ready"
	TTSStateFailedRetryable = "failed_retryable"
	TTSStateFailedTerminal  = "failed_terminal"
)

type TTSGenerationStatus struct {
	Word             string
	WordTarget       string `json:"word_target"` // Neutral alias for Word (same value after sync)
	State            string
	AttemptCount     int
	MaxAttempts      int
	LastErrorCode    *string
	LastErrorMessage *string
	LastProvider     *string
	AudioRelPath     *string
	LastAttemptAt    *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
