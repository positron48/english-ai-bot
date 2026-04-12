package spanishverbs

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// JehleForm is one finite conjugated form from the Jehle CSV row (six slots).
type JehleForm struct {
	Mood        string
	Tense       string
	Person      string
	Number      string
	SurfaceForm string
}

// JehleLemma groups all parsed forms for one Spanish infinitive.
type JehleLemma struct {
	Lemma string
	Forms []JehleForm
}

// ParseJehleVerbDatabaseCSV reads Fred Jehle / Ghidinelli `jehle_verb_database.csv`
// and returns one entry per infinitive with finite forms (1s–3p) per mood/tense row.
func ParseJehleVerbDatabaseCSV(r io.Reader) ([]JehleLemma, error) {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.LazyQuotes = true

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	idx := map[string]int{}
	for i, h := range header {
		k := strings.Trim(strings.TrimSpace(h), `"`)
		idx[k] = i
	}
	required := []string{
		"infinitive", "mood", "tense",
		"form_1s", "form_2s", "form_3s", "form_1p", "form_2p", "form_3p",
	}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("csv missing required column %q", col)
		}
	}

	type acc struct {
		forms []JehleForm
	}
	byLemma := map[string]*acc{}

	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}
		get := func(col string) string {
			i := idx[col]
			if i < 0 || i >= len(rec) {
				return ""
			}
			return strings.TrimSpace(rec[i])
		}
		lemma := strings.ToLower(strings.TrimSpace(get("infinitive")))
		if lemma == "" {
			continue
		}
		mood := strings.ToLower(strings.TrimSpace(get("mood")))
		tense := strings.ToLower(strings.TrimSpace(get("tense")))
		if mood == "" || tense == "" {
			continue
		}
		a := byLemma[lemma]
		if a == nil {
			a = &acc{}
			byLemma[lemma] = a
		}
		slots := []struct {
			form   string
			person string
			number string
		}{
			{get("form_1s"), "1", "singular"},
			{get("form_2s"), "2", "singular"},
			{get("form_3s"), "3", "singular"},
			{get("form_1p"), "1", "plural"},
			{get("form_2p"), "2", "plural"},
			{get("form_3p"), "3", "plural"},
		}
		for _, s := range slots {
			form := strings.TrimSpace(s.form)
			if form == "" {
				continue
			}
			a.forms = append(a.forms, JehleForm{
				Mood:        mood,
				Tense:       tense,
				Person:      s.person,
				Number:      s.number,
				SurfaceForm: form,
			})
		}
	}

	out := make([]JehleLemma, 0, len(byLemma))
	for lemma, a := range byLemma {
		if len(a.forms) == 0 {
			continue
		}
		out = append(out, JehleLemma{Lemma: lemma, Forms: a.forms})
	}
	return out, nil
}
