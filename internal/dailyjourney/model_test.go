package dailyjourney

import (
	"math/rand"
	"testing"
)

func TestDailyJourneySelection(t *testing.T) {
	candidates := []Step{{Mode: Words}, {Mode: Reading}}
	for _, m := range []Mode{Grammar, GrammarSRS, VerbForms, Picture, Conversation} {
		candidates = append(candidates, Step{Mode: m})
	}
	// A much larger picture catalog must not bias mode selection.
	for i := 0; i < 99; i++ {
		candidates = append(candidates, Step{Mode: Picture})
	}
	seen := map[Mode]int{}
	for seed := int64(0); seed < 2000; seed++ {
		selected := Select(candidates, rand.New(rand.NewSource(seed)))
		if len(selected) != 5 || selected[0].Mode != Words || selected[1].Mode != Reading || !selected[0].Required || !selected[1].Required {
			t.Fatalf("invalid route: %+v", selected)
		}
		unique := map[Mode]bool{}
		for _, s := range selected {
			if unique[s.Mode] {
				t.Fatalf("duplicate mode %s", s.Mode)
			}
			unique[s.Mode] = true
			if !s.Required {
				seen[s.Mode]++
			}
		}
	}
	for _, m := range []Mode{Grammar, GrammarSRS, VerbForms, Picture, Conversation} {
		if seen[m] < 1050 || seen[m] > 1350 {
			t.Fatalf("biased/missing mode %s: %d", m, seen[m])
		}
	}
	short := Select([]Step{{Mode: Words}, {Mode: Grammar}}, rand.New(rand.NewSource(1)))
	if len(short) != 3 || short[1].Status != "unavailable" {
		t.Fatalf("short route: %+v", short)
	}
}
func TestDailyJourneyWordAdditionTarget(t *testing.T) {
	for _, tc := range []struct{ due, want int }{{0, 15}, {5, 15}, {19, 10}, {20, 0}, {21, 0}} {
		if got := WordAdditionTarget(tc.due); got != tc.want {
			t.Errorf("due=%d got=%d want=%d", tc.due, got, tc.want)
		}
	}
}
func TestDailyJourneyResolutionAndSkip(t *testing.T) {
	j := Journey{Status: "active", Steps: []Step{{Required: true, Status: "pending"}, {Status: "pending"}, {Status: "pending"}}}
	if j.CanSkip(j.Current()) {
		t.Fatal("mandatory skip allowed")
	}
	j.Steps[0].Status = "completed"
	if !j.CanSkip(j.Current()) {
		t.Fatal("optional skip denied")
	}
	j.Steps[1].Status = "skipped"
	if j.CanSkip(j.Current()) {
		t.Fatal("second skip allowed")
	}
	j.Steps[2].Status = "unavailable"
	j.Resolve()
	if j.Status != "completed" {
		t.Fatalf("got %s", j.Status)
	}
	j = Journey{Status: "active", Steps: []Step{{Status: "unavailable"}}}
	j.Resolve()
	if j.Status != "unavailable" {
		t.Fatal("empty itinerary granted success")
	}
}
