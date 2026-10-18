// Package dailyjourney contains the persisted itinerary, independent of exercise engines.
package dailyjourney

import (
	"encoding/json"
	"errors"
	"math/rand"
)

const AlgorithmVersion = 1

var (
	ErrProcessing      = errors.New("daily journey action is processing")
	ErrReservationLost = errors.New("daily journey reservation was superseded")
	ErrAIProvider      = errors.New("daily journey AI provider failed")
	ErrNotFound        = errors.New("daily journey not found")
	ErrConflict        = errors.New("daily journey revision changed")
	ErrActionKey       = errors.New("daily journey action key reused")
	ErrInvalidAction   = errors.New("invalid daily journey action")
	ErrAccess          = errors.New("daily journey content access denied")
)

type Mode string

const (
	Words        Mode = "words"
	Reading      Mode = "reading"
	Grammar      Mode = "grammar"
	GrammarSRS   Mode = "grammar_srs"
	VerbForms    Mode = "verb_forms"
	Picture      Mode = "picture"
	Conversation Mode = "conversation"
)

type Launch struct {
	RouteName string            `json:"routeName"`
	Params    map[string]string `json:"params,omitempty"`
}
type Step struct {
	ID             string          `json:"id"`
	Mode           Mode            `json:"mode"`
	Title          string          `json:"title"`
	Goal           string          `json:"goal"`
	Required       bool            `json:"required"`
	Status         string          `json:"status"`
	ContentRef     json.RawMessage `json:"contentRef"`
	ContentVersion string          `json:"contentVersion"`
	TargetCount    int             `json:"targetCount"`
	CompletedCount int             `json:"completedCount"`
	Feedback       *string         `json:"feedback"`
	Result         *string         `json:"result"`
	Launch         *Launch         `json:"launch,omitempty"`
	ResultDetails  json.RawMessage `json:"resultDetails,omitempty"`
	State          json.RawMessage `json:"-"`
}
type Journey struct {
	ID               string `json:"id"`
	CourseCode       string `json:"courseCode"`
	Date             string `json:"date"`
	DayTimezone      string `json:"dayTimezone"`
	Status           string `json:"status"`
	AlgorithmVersion int    `json:"algorithmVersion"`
	Revision         int64  `json:"revision"`
	Steps            []Step `json:"steps"`
}
type Action struct {
	ClientActionID   string          `json:"client_action_id"`
	ExpectedRevision int64           `json:"expected_revision"`
	Type             string          `json:"action"`
	Payload          json.RawMessage `json:"payload"`
}

func (j *Journey) Current() *Step {
	for i := range j.Steps {
		if j.Steps[i].Status == "pending" || j.Steps[i].Status == "active" {
			return &j.Steps[i]
		}
	}
	return nil
}
func (j *Journey) CanSkip(step *Step) bool {
	if j.Status != "active" || step == nil || step.Required {
		return false
	}
	for _, s := range j.Steps {
		if s.Status == "skipped" {
			return false
		}
	}
	return true
}

// Resolve never grants learning credit; subject adapters alone record learning results.
func (j *Journey) Resolve() {
	if j.Current() != nil {
		return
	}
	j.Status = "unavailable"
	for _, s := range j.Steps {
		if s.Status == "completed" {
			j.Status = "completed"
			return
		}
	}
}

// Select consumes already-authorized candidates. Each optional MODE has equal weight,
// regardless of how many content items its catalog contains. It performs no exercise writes.
func Select(candidates []Step, rng *rand.Rand) []Step {
	groups := map[Mode][]Step{}
	for _, candidate := range candidates {
		switch candidate.Mode {
		case Words, Reading, Grammar, GrammarSRS, VerbForms, Picture, Conversation:
			groups[candidate.Mode] = append(groups[candidate.Mode], candidate)
		}
	}
	out := make([]Step, 0, 5)
	pick := func(mode Mode, required bool) {
		list := groups[mode]
		step := Step{Mode: mode, Title: string(mode), Required: required, Status: "unavailable", ContentRef: json.RawMessage(`{}`)}
		if len(list) > 0 {
			step = list[rng.Intn(len(list))]
			step.Required = required
			step.Status = "pending"
		}
		out = append(out, step)
	}
	pick(Words, true)
	pick(Reading, true)
	optional := []Mode{}
	for _, m := range []Mode{Grammar, GrammarSRS, VerbForms, Picture, Conversation} {
		if len(groups[m]) > 0 {
			optional = append(optional, m)
		}
	}
	rng.Shuffle(len(optional), func(i, k int) { optional[i], optional[k] = optional[k], optional[i] })
	if len(optional) > 3 {
		optional = optional[:3]
	}
	for _, mode := range optional {
		pick(mode, false)
	}
	return out
}

// dueWords is the number of unique words, never the number of directional cards.
func WordAdditionTarget(dueWords int) int {
	if dueWords >= 20 {
		return 0
	}
	n := 20 - dueWords
	if n < 10 {
		n = 10
	}
	if n > 15 {
		n = 15
	}
	return n
}
