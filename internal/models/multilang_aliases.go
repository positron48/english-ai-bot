package models

// SyncWordCardNeutralAliases fills neutral alias fields from legacy DB-mapped fields (stage B).
func SyncWordCardNeutralAliases(c *WordCard) {
	if c == nil {
		return
	}
	c.WordTarget = c.Word
	c.DefinitionNative = c.DefinitionRU
	c.DisplayTarget = c.DisplayEN
}

// NormalizeWordCardLegacyBeforeWrite copies neutral fields into legacy fields when legacy is empty.
func NormalizeWordCardLegacyBeforeWrite(c *WordCard) {
	if c == nil {
		return
	}
	if c.Word == "" && c.WordTarget != "" {
		c.Word = c.WordTarget
	}
	if c.DefinitionRU == nil && c.DefinitionNative != nil {
		c.DefinitionRU = c.DefinitionNative
	}
	if c.DisplayEN == nil && c.DisplayTarget != nil {
		c.DisplayEN = c.DisplayTarget
	}
}

// SyncTrainingCardNeutralAliases fills neutral alias fields from legacy column-mapped fields.
func SyncTrainingCardNeutralAliases(c *TrainingCard) {
	if c == nil {
		return
	}
	c.WordTarget = c.WordEN
	c.WordNative = c.WordRU
	c.MeaningTarget = c.MeaningEN
	c.ExampleTarget = c.ExampleEN
	c.ExampleNative = c.ExampleRU
}

// NormalizeTrainingCardLegacyBeforeWrite copies neutral fields into legacy when legacy is empty.
func NormalizeTrainingCardLegacyBeforeWrite(c *TrainingCard) {
	if c == nil {
		return
	}
	if c.WordEN == "" {
		c.WordEN = c.WordTarget
	}
	if c.WordRU == "" {
		c.WordRU = c.WordNative
	}
	if c.MeaningEN == "" {
		c.MeaningEN = c.MeaningTarget
	}
	if c.ExampleEN == "" {
		c.ExampleEN = c.ExampleTarget
	}
	if c.ExampleRU == "" {
		c.ExampleRU = c.ExampleNative
	}
}

// SyncTrainingCardSenseNeutralAliases fills neutral aliases on a sense from legacy LLM keys.
func SyncTrainingCardSenseNeutralAliases(s *TrainingCardSense) {
	if s == nil {
		return
	}
	if s.WordNative == "" {
		s.WordNative = s.WordRU
	}
	if s.MeaningTarget == "" {
		s.MeaningTarget = s.MeaningEN
	}
	if s.ExampleTarget == "" {
		s.ExampleTarget = s.ExampleEN
	}
	if s.ExampleNative == "" {
		s.ExampleNative = s.ExampleRU
	}
}

// SyncTrainingCardResponseNeutralAliases fills neutral aliases on the LLM training-card response.
func SyncTrainingCardResponseNeutralAliases(r *TrainingCardResponse) {
	if r == nil {
		return
	}
	if r.WordTarget == "" {
		r.WordTarget = r.WordEN
	}
	if r.WordEN == "" {
		r.WordEN = r.WordTarget
	}
	for i := range r.Senses {
		SyncTrainingCardSenseNeutralAliases(&r.Senses[i])
	}
}

// SyncWordInfoResponseNeutralAliases fills neutral keys from legacy LLM keys and vice versa.
func SyncWordInfoResponseNeutralAliases(w *WordInfoResponse) {
	if w == nil {
		return
	}
	if w.DefinitionRU == "" && w.DefinitionNative != "" {
		w.DefinitionRU = w.DefinitionNative
	}
	if w.DefinitionNative == "" {
		w.DefinitionNative = w.DefinitionRU
	}
	for i := range w.Examples {
		SyncWordInfoExampleNeutralAliases(&w.Examples[i])
	}
}

// SyncWordInfoExampleNeutralAliases mirrors legacy ↔ neutral for in-memory use (serialization uses MarshalJSON on WordInfoExample).
func SyncWordInfoExampleNeutralAliases(e *WordInfoExample) {
	if e == nil {
		return
	}
	if e.ExampleEN == "" && e.ExampleTarget != "" {
		e.ExampleEN = e.ExampleTarget
	}
	if e.GlossRU == "" && e.GlossNative != "" {
		e.GlossRU = e.GlossNative
	}
	if e.ExampleTarget == "" {
		e.ExampleTarget = e.ExampleEN
	}
	if e.GlossNative == "" {
		e.GlossNative = e.GlossRU
	}
}

// SyncTTSStatusNeutralAliases sets WordTarget from Word (TTS row key is target-language surface form).
func SyncTTSStatusNeutralAliases(s *TTSGenerationStatus) {
	if s == nil {
		return
	}
	s.WordTarget = s.Word
}

// SyncWordSetWordInfoNeutralAliases fills neutral aliases for word-set study API payloads.
func SyncWordSetWordInfoNeutralAliases(w *WordSetWordInfo) {
	if w == nil {
		return
	}
	w.WordTarget = w.Word
	w.DisplayTarget = w.DisplayWord
	w.WordNative = w.WordRU
	w.MeaningTarget = w.MeaningEN
	w.ExampleTarget = w.ExampleEN
	w.ExampleNative = w.ExampleRU
}
