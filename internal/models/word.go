package models

import "time"

// WordCard represents a vocabulary card (lemma) stored in the database
type WordCard struct {
	ID              int64
	Word            string // Lemma (base form)
	Definition      string // Legacy field, kept for compatibility
	POS             *string // Part of speech
	Transcription   *string // IPA transcription
	DefinitionRU   *string // Russian definition
	ExamplesJSON    *string // JSON array of examples
	VerbFormsJSON   *string // JSON object with verb forms (v1, v2, v3, etc.)
	DisplayEN       *string // Display form (e.g., "spy" or "to spy" for verbs)
	ProcessedAt     *time.Time // NULL if not processed yet, set when processing completes (success or error)
	ProcessingError *string    // NULL if no error, contains error message if processing failed
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WordRequestHistory represents a history entry of word requests
type WordRequestHistory struct {
	ID          int64
	UserID      int64
	Word        *string // Legacy field, kept for compatibility
	WordCardID  *int64  // Reference to word_cards.id (lemma)
	InputWord   *string // The word as entered by user (word form)
	RequestedAt time.Time
}

// WordForm represents a mapping from word form to lemma
type WordForm struct {
	Form      string
	WordCardID int64
}

// WordInfoExample represents an example sentence with translation
type WordInfoExample struct {
	ExampleEN string `json:"example_en"`
	GlossRU   string `json:"gloss_ru"`
}

// WordInfoVerbForms represents verb forms (for verbs only)
type WordInfoVerbForms struct {
	V1         string `json:"v1"` // Base form (infinitive)
	V2         string `json:"v2"` // Past simple
	V3         string `json:"v3"` // Past participle
	Gerund     string `json:"gerund,omitempty"` // -ing form
	ThirdPerson string `json:"third_person,omitempty"` // Third person singular
}

// WordInfoResponse represents LLM response for word information (JSON format)
type WordInfoResponse struct {
	Error         string              `json:"error,omitempty"` // Error if word is not English/proper noun/etc
	InputWord     string              `json:"input_word"` // Word as entered by user
	Lemma         string              `json:"lemma"` // Base form (lemma)
	POS           string              `json:"pos"` // Part of speech
	Transcription string              `json:"transcription"` // IPA transcription
	DefinitionRU  string              `json:"definition_ru"` // Russian definition
	Examples      []WordInfoExample   `json:"examples"` // 2-3 examples
	VerbForms     *WordInfoVerbForms `json:"verb_forms,omitempty"` // Verb forms (if verb)
}
