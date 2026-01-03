package models

import "time"

// WordCard represents a vocabulary card stored in the database
type WordCard struct {
	ID              int64
	Word            string
	Definition      string
	ProcessedAt     *time.Time // NULL if not processed yet, set when processing completes (success or error)
	ProcessingError *string    // NULL if no error, contains error message if processing failed
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WordRequestHistory represents a history entry of word requests
type WordRequestHistory struct {
	ID          int64
	UserID      int64
	Word        string
	RequestedAt time.Time
}
