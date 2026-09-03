package placement

import (
	"fmt"
	"math/rand"
	"sort"
)

// Select preserves difficulty quotas and skill variety before considering history.
// recent maps family IDs to recency rank: 1 is the immediately preceding session.
// History is a soft constraint only when a stratum's fresh reserve is exhausted.
func Select(bank *Bank, levels []string, recent map[string]int, existing []Item, rng *rand.Rand) ([]Item, int, error) {
	usedFamilies := map[string]bool{}
	usedIDs := map[string]bool{}
	for _, q := range existing {
		usedFamilies[q.FamilyID] = true
		usedIDs[q.ID] = true
	}
	selected := []Item{}
	repeated := 0
	for _, level := range levels {
		usedSkills := map[string]int{}
		for slot := 0; slot < 6; slot++ {
			difficulty := slot%3 + 1
			pool := []Item{}
			for _, q := range bank.Items {
				if q.Level == level && q.Difficulty == difficulty && !usedFamilies[q.FamilyID] && !usedIDs[q.ID] {
					pool = append(pool, q)
				}
			}
			if len(pool) == 0 {
				return nil, 0, fmt.Errorf("placement reserve exhausted: %s/%d", level, difficulty)
			}
			rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
			// A fresh family beats a recently exposed family; within that band, favour new skills.
			penalty := func(q Item) int {
				r := recent[q.FamilyID]
				if r == 0 {
					return 0
				}
				return 1000 - r
			}
			sort.SliceStable(pool, func(i, j int) bool {
				a, b := penalty(pool[i]), penalty(pool[j])
				if a != b {
					return a < b
				}
				return usedSkills[pool[i].SkillID] < usedSkills[pool[j].SkillID]
			})
			q := pool[0]
			q.Choices = append([]Choice(nil), q.Choices...)
			rng.Shuffle(len(q.Choices), func(i, j int) { q.Choices[i], q.Choices[j] = q.Choices[j], q.Choices[i] })
			selected = append(selected, q)
			usedIDs[q.ID] = true
			usedFamilies[q.FamilyID] = true
			usedSkills[q.SkillID]++
			if recent[q.FamilyID] > 0 {
				repeated++
			}
		}
	}
	return selected, repeated, nil
}

func Profile(s *Snapshot) []LevelScore {
	out := make([]LevelScore, len(Levels))
	for i, l := range Levels {
		out[i] = LevelScore{Level: l, Status: "limited"}
	}
	for _, q := range s.Items {
		a, ok := s.Answers[q.ID]
		if !ok {
			continue
		}
		i := LevelIndex(q.Level)
		if i < 0 {
			continue
		}
		out[i].Total++
		if a == q.CorrectAnswer {
			out[i].Correct++
		}
	}
	for i := range out {
		p := &out[i]
		if p.Total >= 6 && p.Correct*4 >= p.Total*3 {
			p.Status = "secure"
		} else if p.Total >= 6 && p.Correct*2 >= p.Total {
			p.Status = "borderline"
		}
	}
	return out
}

// Clarify picks one level, never stops assessment at the first low-level error.
func Clarify(s *Snapshot) string {
	p := Profile(s)
	highest := -1
	for i, v := range p {
		if v.Status == "secure" {
			highest = i
		}
	}
	// Isolated high successes need confirmation before opening higher course levels.
	if highest > 0 && p[highest-1].Status == "limited" {
		return p[highest].Level
	}
	if highest+1 < len(p) && p[highest+1].Status == "borderline" {
		return p[highest+1].Level
	}
	if highest >= 0 {
		for i := 0; i < highest; i++ {
			if p[i].Status == "limited" {
				return p[i].Level
			}
		}
	}
	if highest < 0 {
		for _, v := range p {
			if v.Status == "borderline" {
				return v.Level
			}
		}
	}
	return ""
}

func Grade(s *Snapshot) Result {
	profile := Profile(s)
	r := Result{Level: "below_a1", UpperLevel: "below_a1", Estimated: true, PolicyVersion: PolicyVersion, Profile: profile, Review: []Review{}, RecommendedSkills: []Skill{}, OpenedSections: []string{}}
	highest := -1
	for i, p := range profile {
		if p.Status != "secure" {
			continue
		}
		// A stronger level plus broad lower evidence can survive individual gaps.
		lowerCorrect, lowerTotal := 0, 0
		for j := 0; j < i; j++ {
			lowerCorrect += profile[j].Correct
			lowerTotal += profile[j].Total
		}
		if i == 0 || (lowerCorrect*2 >= lowerTotal && (profile[i-1].Status != "limited" || p.Total >= 12)) {
			highest = i
		}
	}
	if highest >= 0 {
		r.Level = Levels[highest]
		r.UpperLevel = r.Level
	}
	if highest+1 < len(profile) && profile[highest+1].Status == "borderline" {
		r.UpperLevel = profile[highest+1].Level
	}
	skills := map[string]Skill{}
	for _, sk := range s.Skills {
		skills[sk.ID] = sk
	}
	totals := map[string]int{}
	errors := map[string]int{}
	for _, q := range s.Items {
		a := s.Answers[q.ID]
		correct := a == q.CorrectAnswer
		sk := skills[q.SkillID]
		r.Total++
		if correct {
			r.Correct++
		} else {
			errors[q.SkillID]++
		}
		totals[q.SkillID]++
		r.Review = append(r.Review, Review{Question: q.Public(), Level: q.Level, SkillID: q.SkillID, SkillTitle: sk.Title, Revision: q.Revision, Answer: a, CorrectAnswer: q.CorrectAnswer, Correct: correct, Explanation: q.Explanation, ChapterIDs: append([]string(nil), sk.ChapterIDs...)})
	}
	// Recommendations say "worth reviewing", not "failed topic"; one response is not mastery evidence.
	for _, sk := range s.Skills {
		if errors[sk.ID] > 0 && LevelIndex(sk.Level) <= highest+1 {
			r.RecommendedSkills = append(r.RecommendedSkills, sk)
		}
	}
	sort.SliceStable(r.RecommendedSkills, func(i, j int) bool {
		a, b := r.RecommendedSkills[i], r.RecommendedSkills[j]
		if errors[a.ID] != errors[b.ID] {
			return errors[a.ID] > errors[b.ID]
		}
		return LevelIndex(a.Level) < LevelIndex(b.Level)
	})
	if len(r.RecommendedSkills) > 5 {
		r.RecommendedSkills = r.RecommendedSkills[:5]
	}
	return r
}
