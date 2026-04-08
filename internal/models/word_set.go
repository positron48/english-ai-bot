package models

import "time"

// WordSetCategory represents a category in the word sets hierarchy
type WordSetCategory struct {
	ID          int64     `json:"id"`
	ParentID    *int64    `json:"parent_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsPublished bool      `json:"is_published"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WordSet represents a collection of words
type WordSet struct {
	ID           int64     `json:"id"`
	CategoryID   *int64    `json:"category_id"`
	Title        string    `json:"title"`
	Description  *string   `json:"description"`
	IsPublished  bool      `json:"is_published"`
	SortOrder    int       `json:"sort_order"`
	PreferredPOS *string   `json:"preferred_pos,omitempty"` // Preferred part of speech (noun, verb, adjective, etc.)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WordSetItem represents a word in a word set
type WordSetItem struct {
	ID         int64     `json:"id"`
	WordSetID  int64     `json:"word_set_id"`
	WordCardID int64     `json:"word_card_id"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserWordKnowledge represents user's knowledge status for a word
type UserWordKnowledge struct {
	UserID     int64     `json:"user_id"`
	WordCardID int64     `json:"word_card_id"`
	Status     string    `json:"status"` // Currently only "known"
	CreatedAt  time.Time `json:"created_at"`
}

// WordSetWithProgress represents a word set with user progress information
type WordSetWithProgress struct {
	WordSet
	TotalWords      int     `json:"total_words"`
	KnownWords      int     `json:"known_words"`
	WordsInVocab    int     `json:"words_in_vocab"`
	UnknownWords    int     `json:"unknown_words"`
	ProgressPercent float64 `json:"progress_percent"`
}

// WordSetWordStatus represents the status of a word in a set for a user
type WordSetWordStatus string

const (
	WordStatusUnknown WordSetWordStatus = "unknown"
	WordStatusInVocab WordSetWordStatus = "in_vocab"
	WordStatusKnown   WordSetWordStatus = "known"
)

// WordSetWordInfo represents information about a word in a set
// If preferred_pos is set and a matching training card exists, it includes data from that card
type WordSetWordInfo struct {
	WordCardID    int64             `json:"word_card_id"`
	Word          string            `json:"word"`
	WordTarget    string            `json:"word_target"`
	DisplayWord   string            `json:"display_word"`
	DisplayTarget string            `json:"display_target"`
	Status        WordSetWordStatus `json:"status"`
	// Data from training card with preferred_pos (if available)
	Transcription *string        `json:"transcription,omitempty"`
	WordRU        *string        `json:"word_ru,omitempty"`
	WordNative    *string        `json:"word_native,omitempty"`
	MeaningEN     *string        `json:"meaning_en,omitempty"`
	MeaningTarget *string        `json:"meaning_target,omitempty"`
	ExampleEN     *string        `json:"example_en,omitempty"`
	ExampleTarget *string        `json:"example_target,omitempty"`
	ExampleRU     *string        `json:"example_ru,omitempty"`
	ExampleNative *string        `json:"example_native,omitempty"`
	Morph         *WordMorphInfo `json:"morph,omitempty"` // Compact morphology payload
}
