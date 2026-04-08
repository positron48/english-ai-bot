package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type SpanishGenderLexiconEntry struct {
	Lemma              string
	Gender             string
	Article            string
	OppositeGenderWord string
	Source             string
	Notes              string
}

func defaultSpanishGenderLexiconPaths() []string {
	return []string{
		"resources/wordsets/spanish_gender_lexicon.tsv",
		"data/spanish_gender_lexicon.tsv",
	}
}

func LoadSpanishGenderLexicon(path string) (map[string]SpanishGenderLexiconEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open lexicon: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	index := map[string]int{}
	for i, h := range header {
		index[strings.TrimSpace(strings.ToLower(h))] = i
	}
	required := []string{"lemma", "gender", "article", "opposite_gender_word", "source", "notes"}
	for _, key := range required {
		if _, ok := index[key]; !ok {
			return nil, fmt.Errorf("missing required column %q", key)
		}
	}

	out := make(map[string]SpanishGenderLexiconEntry, 65536)
	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read row: %w", err)
		}
		lemma := strings.ToLower(strings.TrimSpace(rec[index["lemma"]]))
		if lemma == "" {
			continue
		}
		gender := NormalizeNounGenderValue(rec[index["gender"]])
		if gender == "" {
			continue
		}
		entry := SpanishGenderLexiconEntry{
			Lemma:              lemma,
			Gender:             gender,
			Article:            strings.TrimSpace(rec[index["article"]]),
			OppositeGenderWord: strings.ToLower(strings.TrimSpace(rec[index["opposite_gender_word"]])),
			Source:             strings.TrimSpace(rec[index["source"]]),
			Notes:              strings.TrimSpace(rec[index["notes"]]),
		}
		out[lemma] = entry
	}
	return out, nil
}

func LoadSpanishGenderLexiconDefault() (map[string]SpanishGenderLexiconEntry, string, error) {
	if custom := strings.TrimSpace(os.Getenv("SPANISH_GENDER_LEXICON_PATH")); custom != "" {
		lex, err := LoadSpanishGenderLexicon(custom)
		return lex, custom, err
	}
	for _, p := range defaultSpanishGenderLexiconPaths() {
		abs, _ := filepath.Abs(p)
		lex, err := LoadSpanishGenderLexicon(p)
		if err == nil {
			return lex, abs, nil
		}
	}
	return nil, "", fmt.Errorf("lexicon file not found in default paths")
}
