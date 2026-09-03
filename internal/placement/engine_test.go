package placement

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func fixtureBank() *Bank {
	b := &Bank{Version: "fixture", CourseCode: "es_ru", Language: "es"}
	for _, l := range Levels {
		for sk := 0; sk < 8; sk++ {
			sid := fmt.Sprintf("es.%s.%d", l, sk)
			b.Skills = append(b.Skills, Skill{ID: sid, Level: l, Title: sid, Description: sid, SectionID: sid, ChapterIDs: []string{sid}})
			for n := 0; n < 10; n++ {
				id := fmt.Sprintf("%s.%d", sid, n)
				b.Items = append(b.Items, Item{ID: id, Revision: 1, SkillID: sid, FamilyID: id, Level: l, Difficulty: n%3 + 1, Context: "Context", Instruction: "Choose", Prompt: id, Choices: []Choice{{"a", "one"}, {"b", "two"}, {"c", "three"}}, CorrectAnswer: "a", Explanation: "Because", Status: "approved"})
			}
		}
	}
	return b
}
func TestPlacementBalancedRetakesAndFiniteReserve(t *testing.T) {
	b := fixtureBank()
	if err := b.Validate(); err != nil {
		t.Fatal(err)
	}
	recent := map[string]int{}
	for attempt := 1; attempt <= 3; attempt++ {
		items, repeats, err := Select(b, Levels, recent, nil, rand.New(rand.NewSource(int64(attempt))))
		if err != nil || repeats != 0 {
			t.Fatalf("attempt%d repeats=%d err=%v", attempt, repeats, err)
		}
		counts := map[string]int{}
		skills := map[string]map[string]bool{}
		families := map[string]bool{}
		for _, q := range items {
			counts[fmt.Sprintf("%s/%d", q.Level, q.Difficulty)]++
			if families[q.FamilyID] || recent[q.FamilyID] != 0 {
				t.Fatal("repeated family")
			}
			families[q.FamilyID] = true
			if skills[q.Level] == nil {
				skills[q.Level] = map[string]bool{}
			}
			skills[q.Level][q.SkillID] = true
			recent[q.FamilyID] = attempt
		}
		for _, l := range Levels {
			for d := 1; d <= 3; d++ {
				if counts[fmt.Sprintf("%s/%d", l, d)] != 2 {
					t.Fatal("lost quota")
				}
			}
			if len(skills[l]) < 4 {
				t.Fatal("poor skill coverage")
			}
		}
	}
	for _, q := range b.Items {
		recent[q.FamilyID] = 1
	}
	items, repeats, err := Select(b, Levels, recent, nil, rand.New(rand.NewSource(1)))
	if err != nil || len(items) != 30 || repeats != 30 {
		t.Fatal("finite reserve must retain balanced form")
	}
	// Selection shuffles a copy, not shared content or another user's snapshot.
	if b.Items[0].Choices[0].ID != "a" {
		t.Fatal("mutated bank")
	}
}
func TestPlacementPublicQuestionHasNoAnswerMetadata(t *testing.T) {
	q := fixtureBank().Items[0]
	raw, err := json.Marshal(q.Public())
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"correct_answer", "explanation", "feedback", "skill_id", "family_id", "level", "difficulty"} {
		if strings.Contains(string(raw), `"`+private+`"`) {
			t.Fatalf("leaked %s", private)
		}
	}
}
func scoredSnapshot(correct map[string]int) *Snapshot {
	b := fixtureBank()
	items, _, _ := Select(b, Levels, nil, nil, rand.New(rand.NewSource(2)))
	s := &Snapshot{Items: items, Skills: b.Skills, Answers: map[string]string{}}
	seen := map[string]int{}
	for _, q := range items {
		if seen[q.Level] < correct[q.Level] {
			s.Answers[q.ID] = "a"
		} else {
			s.Answers[q.ID] = ""
		}
		seen[q.Level]++
	}
	return s
}
func TestPlacementScoringProfiles(t *testing.T) {
	tests := []struct {
		name           string
		correct        map[string]int
		level, clarify string
	}{
		{"all unknown", map[string]int{}, "below_a1", ""},
		{"all correct", map[string]int{"A1": 6, "A2": 6, "B1": 6, "B2": 6, "C1": 6}, "C1", ""},
		{"one beginner error", map[string]int{"A1": 5, "A2": 6, "B1": 5}, "B1", ""},
		{"borderline B1", map[string]int{"A1": 6, "A2": 6, "B1": 4}, "A2", "B1"},
		{"isolated advanced guessing", map[string]int{"C1": 5}, "below_a1", "C1"},
		{"lower gap", map[string]int{"A1": 2, "A2": 6, "B1": 6}, "B1", "A1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scoredSnapshot(tt.correct)
			r := Grade(s)
			if r.Level != tt.level || Clarify(s) != tt.clarify {
				t.Fatalf("level=%s clarify=%s", r.Level, Clarify(s))
			}
			if r.Total != 30 || len(r.Review) != 30 || !r.Estimated {
				t.Fatal("invalid result")
			}
		})
	}
}
