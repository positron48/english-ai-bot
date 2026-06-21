package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// UnmarshalJSON accepts legacy keys (definition_ru) and neutral definition_native.
func (w *WordInfoResponse) UnmarshalJSON(data []byte) error {
	type wire struct {
		Error            ErrorField         `json:"error"`
		Hint             string             `json:"hint"`
		InputWord        string             `json:"input_word"`
		Lemma            string             `json:"lemma"`
		POS              string             `json:"pos"`
		NounGender       string             `json:"noun_gender"`
		OppositeGender   string             `json:"opposite_gender_word"`
		Transcription    string             `json:"transcription"`
		DefinitionRU     string             `json:"definition_ru"`
		DefinitionNative string             `json:"definition_native"`
		Examples         []WordInfoExample  `json:"examples"`
		VerbForms        *WordInfoVerbForms `json:"verb_forms"`
	}
	var x wire
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	w.Error = x.Error
	w.Hint = x.Hint
	w.InputWord = x.InputWord
	w.Lemma = x.Lemma
	w.POS = x.POS
	w.NounGender = x.NounGender
	w.OppositeGenderWord = x.OppositeGender
	w.Transcription = x.Transcription
	w.DefinitionRU = firstNonEmpty(x.DefinitionRU, x.DefinitionNative)
	w.DefinitionNative = firstNonEmpty(x.DefinitionNative, x.DefinitionRU)
	w.Examples = x.Examples
	w.VerbForms = x.VerbForms
	SyncWordInfoResponseNeutralAliases(w)
	return nil
}

// WordCard represents a vocabulary card (lemma) stored in the database
type WordCard struct {
	ID                 int64
	Word               string     // Lemma (base form) — DB column `word` (target language)
	WordTarget         string     `json:"word_target"` // Neutral alias for Word (same value after SyncWordCardNeutralAliases)
	Definition         string     // Legacy field, kept for compatibility
	POS                *string    // Part of speech
	NounGender         *string    `json:"noun_gender,omitempty"`          // Noun gender (mainly for Spanish nouns): m|f|mf|n
	OppositeGenderWord *string    `json:"opposite_gender_word,omitempty"` // Opposite-gender noun form (e.g. hermano -> hermana), if exists
	Transcription      *string    // IPA transcription
	DefinitionRU       *string    // Russian definition — DB column `definition_ru`
	DefinitionNative   *string    `json:"definition_native,omitempty"` // Neutral alias; shares pointer with DefinitionRU after sync
	ExamplesJSON       *string    // JSON array of examples
	VerbFormsJSON      *string    // JSON object with verb forms (v1, v2, v3, etc.)
	DisplayEN          *string    // Display form (e.g., "spy" or "to spy" for verbs) — DB `display_en`
	DisplayTarget      *string    `json:"display_target,omitempty"` // Neutral alias; shares pointer with DisplayEN after sync
	ProcessedAt        *time.Time // NULL if not processed yet, set when processing completes (success or error)
	ProcessingError    *string    // NULL if no error, contains error message if processing failed
	CourseCode         string     // DB column `course_code`; empty for legacy/untagged rows
	CreatedAt          time.Time
	UpdatedAt          time.Time
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
	Form       string
	WordCardID int64
}

// WordInfoExample represents an example sentence with translation
type WordInfoExample struct {
	ExampleEN     string `json:"example_en"`
	ExampleTarget string `json:"example_target"` // Neutral alias; not persisted on MarshalJSON (see below)
	GlossRU       string `json:"gloss_ru"`
	GlossNative   string `json:"gloss_native"` // Neutral alias; emitted alongside gloss_ru in JSON
}

// MarshalJSON emits legacy keys and neutral aliases (API dual contract). DB rows may store either shape; Unmarshal accepts both.
func (e WordInfoExample) MarshalJSON() ([]byte, error) {
	ex := e
	SyncWordInfoExampleNeutralAliases(&ex)
	return json.Marshal(map[string]string{
		"example_en":     ex.ExampleEN,
		"example_target": ex.ExampleTarget,
		"gloss_ru":       ex.GlossRU,
		"gloss_native":   ex.GlossNative,
	})
}

// UnmarshalJSON accepts legacy and neutral keys; neutral fills legacy when missing.
func (e *WordInfoExample) UnmarshalJSON(data []byte) error {
	type aux struct {
		ExampleEN     string `json:"example_en"`
		ExampleTarget string `json:"example_target"`
		GlossRU       string `json:"gloss_ru"`
		GlossNative   string `json:"gloss_native"`
	}
	var a aux
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	e.ExampleEN = a.ExampleEN
	e.ExampleTarget = a.ExampleTarget
	e.GlossRU = a.GlossRU
	e.GlossNative = a.GlossNative
	SyncWordInfoExampleNeutralAliases(e)
	return nil
}

// WordInfoVerbForms represents verb forms (for verbs only)
type WordInfoVerbForms struct {
	V1          string `json:"v1"`                     // Base form (infinitive)
	V2          string `json:"v2"`                     // Past simple
	V3          string `json:"v3"`                     // Past participle
	Gerund      string `json:"gerund,omitempty"`       // -ing form
	ThirdPerson string `json:"third_person,omitempty"` // Third person singular
}

// WordMorphInfo is a compact morphology payload for UI cards/training.
type WordMorphInfo struct {
	POS                string             `json:"pos,omitempty"`
	NounGender         string             `json:"noun_gender,omitempty"`
	Article            string             `json:"article,omitempty"`
	OppositeGenderWord string             `json:"opposite_gender_word,omitempty"`
	VerbForms          *WordInfoVerbForms `json:"verb_forms,omitempty"`
}

// ErrorField represents error field that can be either bool or string
// This allows parsing both "error": true and "error": "some message"
type ErrorField struct {
	IsError bool
	Message string
}

// UnmarshalJSON implements custom JSON unmarshaling for ErrorField
// It handles both bool (true/false) and string values
func (e *ErrorField) UnmarshalJSON(data []byte) error {
	// Try to parse as bool first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		e.IsError = b
		if b {
			e.Message = "true"
		}
		return nil
	}

	// Try to parse as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Message = s
		// Check if string is "true" (case-insensitive)
		if val, err := strconv.ParseBool(s); err == nil {
			e.IsError = val
		} else {
			// If it's not a boolean string, treat non-empty as error
			e.IsError = s != "" && s != "false" && s != "null"
		}
		return nil
	}

	// If neither bool nor string, default to false
	e.IsError = false
	e.Message = ""
	return nil
}

// IsTrue checks if error is true (either bool true or string "true")
func (e *ErrorField) IsTrue() bool {
	if e.IsError {
		return true
	}
	msgLower := strings.ToLower(strings.TrimSpace(e.Message))
	return msgLower == "true"
}

// WordInfoResponse represents LLM response for word information (JSON format)
type WordInfoResponse struct {
	Error              ErrorField         `json:"error,omitempty"`                // Error if word is not English/proper noun/etc (can be bool or string)
	Hint               string             `json:"hint,omitempty"`                 // User-friendly hint/suggestion when word is not found
	InputWord          string             `json:"input_word"`                     // Word as entered by user
	Lemma              string             `json:"lemma"`                          // Base form (lemma)
	POS                string             `json:"pos"`                            // Part of speech
	NounGender         string             `json:"noun_gender,omitempty"`          // Noun gender (mainly for Spanish nouns)
	OppositeGenderWord string             `json:"opposite_gender_word,omitempty"` // Opposite-gender noun form (if exists)
	Transcription      string             `json:"transcription"`                  // IPA transcription
	DefinitionRU       string             `json:"definition_ru"`                  // Russian definition
	DefinitionNative   string             `json:"definition_native"`              // Neutral alias (filled from definition_ru when missing)
	Examples           []WordInfoExample  `json:"examples"`                       // 2-3 examples
	VerbForms          *WordInfoVerbForms `json:"verb_forms,omitempty"`           // Verb forms (if verb)
}
