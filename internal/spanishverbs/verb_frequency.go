package spanishverbs

import (
	_ "embed"
	"encoding/json"
)

// Generated from resources/wordsets/spanish_word_freq_pos_ud_top6000.csv (VERB/AUX).
//
//go:embed verb_frequency.json
var verbFrequencyJSON []byte

var OrderedVerbLemmas []string
var verbFrequencyRanks map[string]int

func init() {
	if err := json.Unmarshal(verbFrequencyJSON, &OrderedVerbLemmas); err != nil {
		panic(err)
	}
	verbFrequencyRanks = map[string]int{}
	for rank, lemma := range OrderedVerbLemmas {
		verbFrequencyRanks[lemma] = rank
	}
}
func VerbFrequencyRank(lemma string) int {
	if rank, ok := verbFrequencyRanks[lemma]; ok {
		return rank
	}
	return len(OrderedVerbLemmas) + 100
}
