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
