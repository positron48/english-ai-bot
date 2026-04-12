package service

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"strings"
)

// VerbChoiceOptionCount is the number of alternatives shown in verb-form multiple choice (one correct + distractors).
const VerbChoiceOptionCount = 4

// BuildVerbFormMultipleChoiceOptions returns exactly VerbChoiceOptionCount strings when enough distractors can be built:
// the correct surface form plus distractors, shuffled. seed stabilizes shuffle/distractor picks.
func BuildVerbFormMultipleChoiceOptions(surfaceForm, lemma string, seed int64) []string {
	surfaceForm = strings.TrimSpace(surfaceForm)
	lemma = strings.TrimSpace(lemma)
	correct := surfaceForm

	var distractors []string
	addDistractor := func(v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" || strings.EqualFold(v, correct) {
			return
		}
		for _, existing := range distractors {
			if strings.EqualFold(existing, v) {
				return
			}
		}
		distractors = append(distractors, v)
	}
	addDistractor(lemma)
	if strings.HasSuffix(surfaceForm, "o") {
		addDistractor(strings.TrimSuffix(surfaceForm, "o") + "a")
	}
	if strings.HasSuffix(surfaceForm, "s") {
		addDistractor(strings.TrimSuffix(surfaceForm, "s"))
	} else {
		addDistractor(surfaceForm + "s")
	}
	ru := []rune(surfaceForm)
	if len(ru) >= 2 {
		cp := make([]rune, len(ru))
		copy(cp, ru)
		cp[len(cp)-2], cp[len(cp)-1] = cp[len(cp)-1], cp[len(cp)-2]
		addDistractor(string(cp))
	}
	if len(ru) > 1 {
		addDistractor(string(ru[:len(ru)-1]))
	}
	pads := []string{"a", "e", "o", "as", "es", "os", "an", "en", "na", "te"}
	for i := 0; len(distractors) < VerbChoiceOptionCount-1 && i < len(pads)*4; i++ {
		addDistractor(surfaceForm + pads[i%len(pads)])
	}
	for i := 0; len(distractors) < VerbChoiceOptionCount-1 && i < len(pads)*4; i++ {
		if lemma != "" && !strings.EqualFold(lemma, surfaceForm) {
			addDistractor(lemma + pads[i%len(pads)])
		}
	}

	needWrong := VerbChoiceOptionCount - 1
	for len(distractors) < needWrong {
		added := false
		for _, p := range pads {
			cand := strings.ToLower(strings.TrimSpace(surfaceForm)) + p
			before := len(distractors)
			addDistractor(cand)
			if len(distractors) > before {
				added = true
				break
			}
		}
		if !added && lemma != "" {
			for _, p := range pads {
				cand := strings.ToLower(strings.TrimSpace(lemma)) + p
				before := len(distractors)
				addDistractor(cand)
				if len(distractors) > before {
					added = true
					break
				}
			}
		}
		if !added {
			break
		}
	}

	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(distractors), func(i, j int) { distractors[i], distractors[j] = distractors[j], distractors[i] })
	if len(distractors) > needWrong {
		distractors = distractors[:needWrong]
	}

	out := make([]string, 0, VerbChoiceOptionCount)
	out = append(out, correct)
	out = append(out, distractors...)

	// If still short (pathological input), pad onto out until we have VerbChoiceOptionCount unique strings.
	padIdx := 0
	for len(out) < VerbChoiceOptionCount && padIdx < len(pads)*16 {
		cand := strings.ToLower(strings.TrimSpace(surfaceForm)) + pads[padIdx%len(pads)]
		padIdx++
		if strings.EqualFold(cand, correct) {
			continue
		}
		dup := false
		for _, x := range out {
			if strings.EqualFold(x, cand) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, cand)
		}
	}

	rng2 := rand.New(rand.NewSource(seed ^ int64(0x6a09e667f3bcc909)))
	rng2.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// CapVerbMultipleChoiceOptions trims stored options to VerbChoiceOptionCount while preserving the correct answer when present.
func CapVerbMultipleChoiceOptions(correct string, options []string, seed int64) []string {
	correct = strings.TrimSpace(correct)
	if len(options) <= VerbChoiceOptionCount {
		return options
	}
	var wrong []string
	var correctCanon string
	for _, o := range options {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if strings.EqualFold(o, correct) {
			correctCanon = o
			continue
		}
		dup := false
		for _, w := range wrong {
			if strings.EqualFold(w, o) {
				dup = true
				break
			}
		}
		if !dup {
			wrong = append(wrong, o)
		}
	}
	if correctCanon == "" {
		correctCanon = correct
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(wrong), func(i, j int) { wrong[i], wrong[j] = wrong[j], wrong[i] })
	needWrong := VerbChoiceOptionCount - 1
	if len(wrong) > needWrong {
		wrong = wrong[:needWrong]
	}
	out := append([]string{correctCanon}, wrong...)
	pads := []string{"a", "e", "o", "as", "es", "os", "an", "en", "na", "te"}
	pi := 0
	for len(out) < VerbChoiceOptionCount && pi < len(pads)*16 {
		cand := strings.ToLower(strings.TrimSpace(correct)) + pads[pi%len(pads)]
		pi++
		dup := false
		for _, x := range out {
			if strings.EqualFold(x, cand) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, cand)
		}
	}
	rng2 := rand.New(rand.NewSource(int64(uint64(seed) ^ uint64(0x9e3779b97f4a7c15))))
	rng2.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// ClozeBlankPlaceholder is shown instead of the target surface form in cloze-style prompts.
const ClozeBlankPlaceholder = "____"

// MaskClozeVerbSurfaceInQuestion replaces every whole-word occurrence of surface (case-insensitive) with a blank.
func MaskClozeVerbSurfaceInQuestion(question, surface string) string {
	q := strings.TrimSpace(question)
	s := strings.TrimSpace(surface)
	if q == "" || s == "" {
		return question
	}
	re, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(s) + `\b`)
	if err != nil {
		return question
	}
	return re.ReplaceAllString(q, ClozeBlankPlaceholder)
}

// ParseStringJSONArray parses JSON array of strings; returns nil on empty/invalid.
func ParseStringJSONArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
