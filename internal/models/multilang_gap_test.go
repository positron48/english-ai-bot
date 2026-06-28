package models

import "testing"

func strPtr(s string) *string { return &s }

func TestMultilangGap_CanonicalWordPOS(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"  ", ""},
		{"Noun", "noun"},
		{"sustantivo masculino", "noun"},
		{"VERB", "verb"},
		{"verbo", "verb"},
		{"Adjective", "adjective"},
		{"adjetivo", "adjective"},
		{"Adverb", "adverb"},
		{"adverbio", "adverb"},
		{"Pronoun", "pronoun"},
		{"pronombre", "pronoun"},
		{"Preposition", "preposition"},
		{"preposición", "preposition"},
		{"preposicion", "preposition"},
		{"Conjunction", "conjunction"},
		{"conjunción", "conjunction"},
		{"conjuncion", "conjunction"},
		{"Interjection", "interjection"},
		{"interjección", "interjection"},
		{"interjeccion", "interjection"},
		{"Article", "article"},
		{"artículo", "article"},
		{"articulo", "article"},
		{"auxiliary", "auxiliary"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := CanonicalWordPOS(tt.raw); got != tt.want {
				t.Fatalf("CanonicalWordPOS(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestMultilangGap_IsNounPOS(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{"noun", true},
		{"NOUN", true},
		{"sustantivo", true},
		{"verb", false},
		{"", false},
		{"adjective", false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := IsNounPOS(tt.raw); got != tt.want {
				t.Fatalf("IsNounPOS(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestMultilangGap_InferNounGenderFromPOSText(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"  ", ""},
		{"sustantivo femenino", "f"},
		{"feminine noun", "f"},
		{"masculino", "m"},
		{"masculine", "m"},
		{"neutro", "n"},
		{"neutral gender", "n"},
		{"género común", "mf"},
		{"comun", "mf"},
		{"common gender", "mf"},
		{"m/f", "mf"},
		{"verb", ""},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := InferNounGenderFromPOSText(tt.raw); got != tt.want {
				t.Fatalf("InferNounGenderFromPOSText(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestMultilangGap_IsLikelySimpleSpanishGenderPair(t *testing.T) {
	tests := []struct {
		lemma    string
		opposite string
		want     bool
	}{
		{"", "hermana", false},
		{"hermano", "", false},
		{"hermano", "hermano", false},
		{"hermano", "hermana", true},
		{"hermana", "hermano", true},
		{"actor", "actriz", false},
		{"perro", "gato", false},
		{"  Hermano ", " hermana ", true},
	}
	for _, tt := range tests {
		t.Run(tt.lemma+"_"+tt.opposite, func(t *testing.T) {
			if got := IsLikelySimpleSpanishGenderPair(tt.lemma, tt.opposite); got != tt.want {
				t.Fatalf("IsLikelySimpleSpanishGenderPair(%q, %q) = %v, want %v", tt.lemma, tt.opposite, got, tt.want)
			}
		})
	}
}

func TestMultilangGap_SyncWordCardNeutralAliases_nil(t *testing.T) {
	SyncWordCardNeutralAliases(nil)
}

func TestMultilangGap_NormalizeWordCardLegacyBeforeWrite(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		NormalizeWordCardLegacyBeforeWrite(nil)
	})

	t.Run("infers gender from POS and normalizes fields", func(t *testing.T) {
		pos := "  Sustantivo femenino  "
		c := &WordCard{
			WordTarget:       "hermana",
			DefinitionNative: strPtr("сестра"),
			DisplayTarget:    strPtr("hermana"),
			POS:              &pos,
		}
		NormalizeWordCardLegacyBeforeWrite(c)
		if c.Word != "hermana" {
			t.Fatalf("Word = %q", c.Word)
		}
		if c.NounGender == nil || *c.NounGender != "f" {
			t.Fatalf("NounGender = %v, want f", c.NounGender)
		}
		if c.POS == nil || *c.POS != "noun" {
			t.Fatalf("POS = %v, want noun", c.POS)
		}
	})

	t.Run("clears invalid gender and empty POS", func(t *testing.T) {
		invalidGender := "unknown"
		emptyPOS := "   "
		c := &WordCard{
			NounGender: &invalidGender,
			POS:        &emptyPOS,
		}
		NormalizeWordCardLegacyBeforeWrite(c)
		if c.NounGender != nil {
			t.Fatalf("NounGender = %v, want nil", c.NounGender)
		}
		if c.POS != nil {
			t.Fatalf("POS = %v, want nil", c.POS)
		}
	})

	t.Run("preserves explicit valid gender", func(t *testing.T) {
		g := "M"
		pos := "noun"
		c := &WordCard{NounGender: &g, POS: &pos}
		NormalizeWordCardLegacyBeforeWrite(c)
		if c.NounGender == nil || *c.NounGender != "m" {
			t.Fatalf("NounGender = %v, want m", c.NounGender)
		}
	})
}

func TestMultilangGap_SyncTrainingCardNeutralAliases_nil(t *testing.T) {
	SyncTrainingCardNeutralAliases(nil)
}

func TestMultilangGap_NormalizeTrainingCardLegacyBeforeWrite(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		NormalizeTrainingCardLegacyBeforeWrite(nil)
	})

	t.Run("copies neutral into legacy", func(t *testing.T) {
		c := &TrainingCard{
			WordTarget:    "go",
			WordNative:    "идти",
			MeaningTarget: "move",
			ExampleTarget: "I go.",
			ExampleNative: "Я иду.",
		}
		NormalizeTrainingCardLegacyBeforeWrite(c)
		if c.WordEN != "go" || c.WordRU != "идти" || c.MeaningEN != "move" || c.ExampleEN != "I go." || c.ExampleRU != "Я иду." {
			t.Fatalf("legacy fields not filled: %+v", c)
		}
	})

	t.Run("does not overwrite legacy", func(t *testing.T) {
		c := &TrainingCard{
			WordEN:        "keep",
			WordRU:        "ru",
			MeaningEN:     "m",
			ExampleEN:     "e",
			ExampleRU:     "r",
			WordTarget:    "other",
			WordNative:    "other",
			MeaningTarget: "other",
			ExampleTarget: "other",
			ExampleNative: "other",
		}
		NormalizeTrainingCardLegacyBeforeWrite(c)
		if c.WordEN != "keep" || c.WordRU != "ru" || c.MeaningEN != "m" || c.ExampleEN != "e" || c.ExampleRU != "r" {
			t.Fatalf("legacy fields overwritten: %+v", c)
		}
	})
}

func TestMultilangGap_SyncTrainingCardSenseNeutralAliases(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		SyncTrainingCardSenseNeutralAliases(nil)
	})

	t.Run("fills from legacy", func(t *testing.T) {
		s := &TrainingCardSense{WordRU: "т", MeaningEN: "m", ExampleEN: "e", ExampleRU: "р"}
		SyncTrainingCardSenseNeutralAliases(s)
		if s.WordNative != "т" || s.MeaningTarget != "m" || s.ExampleTarget != "e" || s.ExampleNative != "р" {
			t.Fatalf("sense neutral: %+v", s)
		}
	})

	t.Run("preserves WordNative", func(t *testing.T) {
		s := &TrainingCardSense{WordNative: "native", WordRU: "ru"}
		SyncTrainingCardSenseNeutralAliases(s)
		if s.WordNative != "native" {
			t.Fatalf("WordNative = %q", s.WordNative)
		}
	})
}

func TestMultilangGap_SyncTrainingCardResponseNeutralAliases(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		SyncTrainingCardResponseNeutralAliases(nil)
	})

	t.Run("WordTarget from WordEN", func(t *testing.T) {
		r := &TrainingCardResponse{WordEN: "lemma"}
		SyncTrainingCardResponseNeutralAliases(r)
		if r.WordTarget != "lemma" {
			t.Fatalf("WordTarget = %q", r.WordTarget)
		}
	})

	t.Run("WordEN from WordTarget", func(t *testing.T) {
		r := &TrainingCardResponse{WordTarget: "lemma"}
		SyncTrainingCardResponseNeutralAliases(r)
		if r.WordEN != "lemma" {
			t.Fatalf("WordEN = %q", r.WordEN)
		}
	})
}

func TestMultilangGap_SyncWordInfoResponseNeutralAliases(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		SyncWordInfoResponseNeutralAliases(nil)
	})

	t.Run("DefinitionRU from DefinitionNative", func(t *testing.T) {
		w := &WordInfoResponse{DefinitionNative: "native only"}
		SyncWordInfoResponseNeutralAliases(w)
		if w.DefinitionRU != "native only" {
			t.Fatalf("DefinitionRU = %q", w.DefinitionRU)
		}
	})
}

func TestMultilangGap_SyncWordInfoExampleNeutralAliases(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		SyncWordInfoExampleNeutralAliases(nil)
	})

	t.Run("fills legacy from neutral", func(t *testing.T) {
		e := &WordInfoExample{ExampleTarget: "target", GlossNative: "gloss"}
		SyncWordInfoExampleNeutralAliases(e)
		if e.ExampleEN != "target" || e.GlossRU != "gloss" {
			t.Fatalf("legacy not filled: %+v", e)
		}
	})

	t.Run("fills neutral from legacy", func(t *testing.T) {
		e := &WordInfoExample{ExampleEN: "en", GlossRU: "ru"}
		SyncWordInfoExampleNeutralAliases(e)
		if e.ExampleTarget != "en" || e.GlossNative != "ru" {
			t.Fatalf("neutral not filled: %+v", e)
		}
	})
}

func TestMultilangGap_SyncTTSStatusNeutralAliases_nil(t *testing.T) {
	SyncTTSStatusNeutralAliases(nil)
}

func TestMultilangGap_SyncWordSetWordInfoNeutralAliases_nil(t *testing.T) {
	SyncWordSetWordInfoNeutralAliases(nil)
}
