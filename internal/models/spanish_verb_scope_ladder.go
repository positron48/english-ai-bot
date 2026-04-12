package models

import "strings"

// SpanishVerbScopeLadderStep is one rung of indicative-first verb-form training (keys match
// verb_forms_dict after Jehle import: 'es.' + lower(tense) + '.' + lower(mood)).
type SpanishVerbScopeLadderStep struct {
	Scope   string `json:"scope"`
	LabelRU string `json:"label_ru"`
	LabelEN string `json:"label_en"`
}

// SpanishVerbScopeLadder returns ordered progression (easiest → harder) for UI and scope expansion.
// Cumulative: level N means scopes ladder[0]…ladder[N] are all active in training.
func SpanishVerbScopeLadder() []SpanishVerbScopeLadderStep {
	return []SpanishVerbScopeLadderStep{
		{Scope: "es.presente.indicativo", LabelRU: "Настоящее (indicativo)", LabelEN: "Present (indicative)"},
		{Scope: "es.pretérito.indicativo", LabelRU: "Претерит (соверш. прошл.)", LabelEN: "Preterite (indicative)"},
		{Scope: "es.imperfecto.indicativo", LabelRU: "Имперфект (несоверш. прошл.)", LabelEN: "Imperfect (indicative)"},
		{Scope: "es.futuro.indicativo", LabelRU: "Будущее", LabelEN: "Future (indicative)"},
		{Scope: "es.condicional.indicativo", LabelRU: "Условное настоящее", LabelEN: "Conditional (indicative)"},
		{Scope: "es.presente.subjuntivo", LabelRU: "Настоящее сослагательное", LabelEN: "Present subjunctive"},
		{Scope: "es.imperfecto.subjuntivo", LabelRU: "Прошедшее сослагательное", LabelEN: "Imperfect subjunctive"},
	}
}

// SpanishVerbScopesThroughIndex returns cumulative scopes for ladder indices [0..idx] inclusive.
func SpanishVerbScopesThroughIndex(idx int) []string {
	steps := SpanishVerbScopeLadder()
	if len(steps) == 0 {
		return nil
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(steps) {
		idx = len(steps) - 1
	}
	out := make([]string, 0, idx+1)
	for i := 0; i <= idx; i++ {
		out = append(out, steps[i].Scope)
	}
	return out
}

// InferSpanishVerbProgressionIndex finds the largest k such that every ladder scope [0..k]
// appears in enabled (cumulative preset). If enabled is empty or no prefix match, returns 0.
func InferSpanishVerbProgressionIndex(enabled []string) int {
	steps := SpanishVerbScopeLadder()
	if len(steps) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(enabled))
	for _, s := range enabled {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			set[s] = struct{}{}
		}
	}
	best := 0
	for i := 0; i < len(steps); i++ {
		key := strings.ToLower(strings.TrimSpace(steps[i].Scope))
		if _, ok := set[key]; !ok {
			break
		}
		best = i
	}
	return best
}
