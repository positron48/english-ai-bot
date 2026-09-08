package repository

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"
)

type practiceCandidate struct {
	card               VerbQueueCard
	state              string
	lemma, scope, rule string
	regular            bool
	rank               int
}

// GetPracticeQueue prioritizes due work and representative rules over collecting
// every regular lemma. A rule transfers only after success on three different verbs.
func (r *VerbFormsRepository) GetPracticeQueue(userID int64, scopes []string, maxCards, maxNew int) ([]VerbQueueCard, error) {
	scopes = verbtraining.ExpandScopes(scopes)
	if len(scopes) == 0 || maxCards <= 0 {
		return nil, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(scopes)), ",")
	args := []interface{}{userID}
	for _, scope := range scopes {
		args = append(args, scope)
	}
	args = append(args, userID, userID, time.Now())
	freqSQL := "'" + strings.Join(spanishverbs.OrderedVerbLemmas, "','") + "'"
	rows, err := r.db.Query(`SELECT u.id,c.word_card_id,c.card_type,c.prompt_json,c.answer_json,COALESCE(c.distractors_json,''),u.state
 FROM user_verb_cards u JOIN verb_training_cards c ON c.id=u.verb_training_card_id
 JOIN verb_forms_dict d ON d.id=c.verb_form_dict_id JOIN word_cards w ON w.id=c.word_card_id
 WHERE u.user_id=? AND ('es.'||d.tense||'.'||d.mood) IN (`+marks+`)
 AND (LOWER(w.word) IN (`+verbCoreSQL()+`) OR w.id IN (
 SELECT tc.word_card_id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=?
 UNION SELECT word_card_id FROM user_word_knowledge WHERE user_id=? AND status='known'))
 AND (u.state='new' OR u.next_due_at IS NULL OR u.next_due_at<=?)
 AND c.prompt_json::jsonb->>'content_version'='2'
 ORDER BY CASE WHEN u.state='learning' THEN 0 WHEN u.state='review' THEN 1 ELSE 2 END,
 COALESCE(array_position(ARRAY[`+freqSQL+`]::text[],LOWER(w.word)),1000),u.next_due_at NULLS FIRST,u.id LIMIT 3000`, args...)
	if err != nil {
		return nil, err
	}
	candidates := []practiceCandidate{}
	for rows.Next() {
		var c practiceCandidate
		if err = rows.Scan(&c.card.UserVerbCardID, &c.card.WordCardID, &c.card.CardType, &c.card.PromptJSON, &c.card.AnswerJSON, &c.card.DistractorsJSON, &c.state); err != nil {
			rows.Close()
			return nil, err
		}
		var p struct {
			Lemma string            `json:"lemma"`
			Mood  string            `json:"mood"`
			Tense string            `json:"tense"`
			Rule  spanishverbs.Rule `json:"rule"`
		}
		if json.Unmarshal([]byte(c.card.PromptJSON), &p) != nil {
			continue
		}
		c.lemma = p.Lemma
		c.scope = p.Tense + "|" + p.Mood
		c.rule = p.Rule.ID
		c.regular = p.Rule.Regular
		c.rank = spanishverbs.VerbFrequencyRank(p.Lemma)
		candidates = append(candidates, c)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	mastered, err := r.MasteredVerbRules(userID)
	if err != nil {
		return nil, err
	}
	// Keep SQL due ordering. Within new work prefer frequent verbs.
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.state != b.state {
			rank := map[string]int{"learning": 0, "review": 1, "new": 2}
			return rank[a.state] < rank[b.state]
		}
		return a.rank < b.rank
	})
	out := []VerbQueueCard{}
	used := map[int64]bool{}
	lemmas := map[string]int{}
	rules := map[string]int{}
	scopeCounts := map[string]int{}
	slots := map[string]bool{}
	newCount, regularCount := 0, 0
	for len(out) < maxCards {
		best := -1
		bestScore := -1e9
		for i, c := range candidates {
			if used[c.card.UserVerbCardID] || slots[c.lemma+"|"+c.rule] {
				continue
			}
			if c.state == "new" && (newCount >= maxNew || (c.regular && mastered[c.rule])) {
				continue
			}
			score := 100.0 - float64(c.rank)/100
			if c.state == "learning" {
				score += 55
			} else if c.state == "review" {
				score += 35
			}
			score -= float64(lemmas[c.lemma]*35 + rules[c.rule]*25 + scopeCounts[c.scope]*12)
			if c.regular && regularCount*2 <= len(out) {
				score += 45
			}
			if !c.regular && regularCount*2 > len(out) {
				score += 30
			}
			if score > bestScore {
				best = i
				bestScore = score
			}
		}
		if best < 0 {
			break
		}
		c := candidates[best]
		out = append(out, c.card)
		used[c.card.UserVerbCardID] = true
		slots[c.lemma+"|"+c.rule] = true
		lemmas[c.lemma]++
		rules[c.rule]++
		scopeCounts[c.scope]++
		if c.state == "new" {
			newCount++
		}
		if c.regular {
			regularCount++
		}
	}
	return out, nil
}

func (r *VerbFormsRepository) MasteredVerbRules(userID int64) (map[string]bool, error) {
	rows, err := r.db.Query(`WITH attempts AS (
 SELECT answered_at,metrics_json::jsonb->'feedback' AS f FROM verb_review_events
 WHERE user_id=? AND metrics_json IS NOT NULL AND metrics_json::jsonb->>'version'='2' AND metrics_json::jsonb->>'retry'='false'
 ), failures AS (SELECT f->'rule'->>'id' AS rule,MAX(answered_at) AS last_failure FROM attempts WHERE f->>'outcome'!='correct' OR f->>'assisted'='true' GROUP BY 1)
 SELECT a.f->'rule'->>'id',COUNT(DISTINCT a.f->>'lemma') FROM attempts a
 LEFT JOIN failures b ON b.rule=a.f->'rule'->>'id'
 WHERE a.f->>'outcome'='correct' AND a.f->>'assisted'='false' AND a.f->'rule'->>'regular'='true'
 AND (b.last_failure IS NULL OR a.answered_at>b.last_failure) GROUP BY 1 HAVING COUNT(DISTINCT a.f->>'lemma')>=3`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var rule string
		var count int
		if err = rows.Scan(&rule, &count); err != nil {
			return nil, err
		}
		out[rule] = true
	}
	return out, rows.Err()
}
