package models

import "strings"

const (
	VerbCardTypeRecall   = "form_recall"
	VerbCardTypeCloze    = "cloze_form"
	VerbCardTypeContrast = "contrast_pick"
)

// VerbLemma describes one lemma in deterministic conjugation dictionary.
type VerbLemma struct {
	ID            int64
	Lemma         string
	Language      string
	Source        string
	SourceVersion string
	Checksum      string
	MetadataJSON  string
}

// VerbFormDict row in deterministic verb form dictionary.
type VerbFormDict struct {
	ID          int64
	VerbLemmaID int64
	Mood        string
	Tense       string
	Person      string
	Number      string
	SurfaceForm string
	IsIrregular bool
	TagsJSON    string
}

// VerbFormExample stores pre-generated sentence examples for one exact form.
type VerbFormExample struct {
	ID             int64
	VerbFormDictID int64
	ExampleTarget  string
	GlossNative    string
	Source         string
	QualityScore   int
}

// VerbTrainingCard is a deterministic card generated from dictionary + examples.
type VerbTrainingCard struct {
	ID              int64
	WordCardID      int64
	VerbFormDictID  int64
	CardType        string
	PromptJSON      string
	AnswerJSON      string
	DistractorsJSON string
	ExampleID       *int64
}

// UserVerbCard keeps SRS state for verb-form training separately from word training.
type UserVerbCard struct {
	ID                 int64
	UserID             int64
	VerbTrainingCardID int64
	State              string
	EF                 float64
	Reps               int
	IntervalDays       int
	LearningStep       int
	LapseCount         int
	NextDueAt          string
	LastReviewAt       string
	LastQuality        *int
	StatsJSON          string
}

// VerbScope is a normalized scope key used for gating queue by learner level.
// Canonical format: "<target_lang>.<tense>.<mood>".
func VerbScope(targetLang, mood, tense string) string {
	return strings.ToLower(strings.TrimSpace(targetLang)) + "." +
		strings.ToLower(strings.TrimSpace(tense)) + "." +
		strings.ToLower(strings.TrimSpace(mood))
}

// DefaultSpanishVerbScopes returns default beginner scope for Spanish deployments.
func DefaultSpanishVerbScopes() []string {
	return []string{"es.presente.indicativo"}
}
