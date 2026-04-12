package service

import (
	"encoding/json"
	"math/rand"
	"strings"

	"tgbot-skeleton/internal/repository"
)

// ShuffleVerbQueue randomizes card order for a session (same multiset of cards).
func ShuffleVerbQueue(queue []repository.VerbQueueCard, seed int64) {
	if len(queue) < 2 {
		return
	}
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(queue), func(i, j int) {
		queue[i], queue[j] = queue[j], queue[i]
	})
}

func verbQueuePromptAnswerKey(c repository.VerbQueueCard) string {
	var prompt map[string]interface{}
	_ = json.Unmarshal([]byte(c.PromptJSON), &prompt)
	q, _ := prompt["question"].(string)
	var answer map[string]string
	_ = json.Unmarshal([]byte(c.AnswerJSON), &answer)
	surface := strings.TrimSpace(answer["surface_form"])
	return strings.TrimSpace(q) + "\x00" + strings.ToLower(surface)
}

// SpreadAdjacentDuplicateVerbPromptKeys reorders the queue so identical (question + answer) cards are not shown back-to-back when possible.
func SpreadAdjacentDuplicateVerbPromptKeys(queue []repository.VerbQueueCard) []repository.VerbQueueCard {
	if len(queue) < 2 {
		return queue
	}
	out := append([]repository.VerbQueueCard(nil), queue...)
	keys := make([]string, len(out))
	refresh := func() {
		for i := range out {
			keys[i] = verbQueuePromptAnswerKey(out[i])
		}
	}
	refresh()
	changed := true
	for pass := 0; changed && pass < len(out)+5; pass++ {
		changed = false
		for i := 0; i < len(out)-1; i++ {
			if keys[i] != keys[i+1] {
				continue
			}
			j := -1
			for k := i + 2; k < len(out); k++ {
				if keys[k] != keys[i] {
					j = k
					break
				}
			}
			if j < 0 {
				for k := 0; k < i; k++ {
					if keys[k] != keys[i] {
						j = k
						break
					}
				}
			}
			if j >= 0 {
				out[i+1], out[j] = out[j], out[i+1]
				keys[i+1], keys[j] = keys[j], keys[i+1]
				changed = true
			}
		}
	}
	return out
}
