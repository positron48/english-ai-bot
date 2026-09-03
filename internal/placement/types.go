// Package placement defines the versioned, course-independent diagnostic engine.
// It never reads chapter exercise banks and never calls a language model.
package placement

import "fmt"

const PolicyVersion = "editorial-v1"

var Levels = []string{"A1", "A2", "B1", "B2", "C1"}

type Skill struct {
	ID          string   `json:"id"`
	Level       string   `json:"level"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ChapterIDs  []string `json:"chapter_ids"`
	SectionID   string   `json:"section_id"`
}
type Choice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type Item struct {
	ID            string   `json:"id"`
	Revision      int      `json:"revision"`
	SkillID       string   `json:"skill_id"`
	FamilyID      string   `json:"family_id"`
	Level         string   `json:"level"`
	Difficulty    int      `json:"difficulty"`
	Context       string   `json:"context"`
	Instruction   string   `json:"instruction"`
	Prompt        string   `json:"prompt"`
	Choices       []Choice `json:"choices"`
	CorrectAnswer string   `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
	Status        string   `json:"status"`
}
type Bank struct {
	Version    string  `json:"version"`
	CourseCode string  `json:"course_code"`
	Language   string  `json:"language"`
	Skills     []Skill `json:"skills"`
	Items      []Item  `json:"items"`
}

// Question intentionally has no key, skill/level metadata or explanations.
type Question struct {
	ID          string   `json:"id"`
	Context     string   `json:"context"`
	Instruction string   `json:"instruction"`
	Prompt      string   `json:"prompt"`
	Choices     []Choice `json:"choices"`
}

func (i Item) Public() Question { return Question{i.ID, i.Context, i.Instruction, i.Prompt, i.Choices} }

type Snapshot struct {
	Items              []Item            `json:"items"`
	Reserve            []Item            `json:"reserve"` // unseen clarification items, pinned at start
	Skills             []Skill           `json:"skills"`
	Answers            map[string]string `json:"answers"` // empty string means explicit "I don't know"
	BaseClosed         bool              `json:"base_closed"`
	ClarificationLevel string            `json:"clarification_level,omitempty"`
	RepeatedFamilies   int               `json:"repeated_families"`
}
type LevelScore struct {
	Level   string `json:"level"`
	Correct int    `json:"correct"`
	Total   int    `json:"total"`
	Status  string `json:"status"` // secure, borderline, limited
}
type Review struct {
	Question
	Level         string   `json:"level"`
	SkillID       string   `json:"skill_id"`
	SkillTitle    string   `json:"skill_title"`
	Revision      int      `json:"revision"`
	Answer        string   `json:"answer"`
	CorrectAnswer string   `json:"correct_answer"`
	Correct       bool     `json:"correct"`
	Explanation   string   `json:"explanation"`
	ChapterIDs    []string `json:"chapter_ids"`
}
type Result struct {
	Level             string       `json:"level"`
	UpperLevel        string       `json:"upper_level"`
	Estimated         bool         `json:"estimated"`
	PolicyVersion     string       `json:"policy_version"`
	Correct           int          `json:"correct"`
	Total             int          `json:"total"`
	Profile           []LevelScore `json:"profile"`
	Review            []Review     `json:"review"`
	RecommendedSkills []Skill      `json:"recommended_skills"`
	OpenedSections    []string     `json:"opened_sections"`
}

func LevelIndex(level string) int {
	for i, l := range Levels {
		if l == level {
			return i
		}
	}
	return -1
}
func (b *Bank) Validate() error {
	if b.Version == "" || (b.CourseCode != "en_ru" && b.CourseCode != "es_ru") {
		return fmt.Errorf("invalid placement bank identity")
	}
	skills := map[string]Skill{}
	ids := map[string]bool{}
	counts := map[string]map[int]int{}
	for _, s := range b.Skills {
		if s.ID == "" || LevelIndex(s.Level) < 0 || len(s.ChapterIDs) == 0 {
			return fmt.Errorf("invalid skill %s", s.ID)
		}
		if _, ok := skills[s.ID]; ok {
			return fmt.Errorf("duplicate skill %s", s.ID)
		}
		skills[s.ID] = s
	}
	for _, q := range b.Items {
		s, ok := skills[q.SkillID]
		if !ok || s.Level != q.Level || q.ID == "" || q.FamilyID == "" || q.Revision < 1 || q.Status != "approved" || q.Prompt == "" || q.Instruction == "" || q.Explanation == "" || q.Difficulty < 1 || q.Difficulty > 3 || len(q.Choices) < 3 || len(q.Choices) > 4 {
			return fmt.Errorf("invalid item %s", q.ID)
		}
		if ids[q.ID] {
			return fmt.Errorf("duplicate item %s", q.ID)
		}
		ids[q.ID] = true
		keys := map[string]bool{}
		texts := map[string]bool{}
		for _, c := range q.Choices {
			if c.ID == "" || c.Text == "" || keys[c.ID] || texts[c.Text] {
				return fmt.Errorf("invalid choices %s", q.ID)
			}
			keys[c.ID] = true
			texts[c.Text] = true
		}
		if !keys[q.CorrectAnswer] {
			return fmt.Errorf("invalid key %s", q.ID)
		}
		if counts[q.Level] == nil {
			counts[q.Level] = map[int]int{}
		}
		counts[q.Level][q.Difficulty]++
	}
	for _, l := range Levels {
		for d := 1; d <= 3; d++ {
			if counts[l][d] < 8 {
				return fmt.Errorf("insufficient reserve at %s difficulty %d", l, d)
			}
		}
	}
	return nil
}
