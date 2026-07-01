-- Case-insensitive word/course_code lookups (reading word-click, vocab word-card detail,
-- dictionary course filter) all wrap columns in LOWER(...), which a plain B-tree index on
-- word_cards(word) / word_cards(course_code) / word_forms(form) cannot serve — forcing a
-- sequential scan on every lookup. Add matching functional indexes.
CREATE INDEX IF NOT EXISTS idx_word_cards_lower_word_course ON word_cards (LOWER(word), course_code);
CREATE INDEX IF NOT EXISTS idx_word_cards_lower_course_code ON word_cards (LOWER(course_code));
CREATE INDEX IF NOT EXISTS idx_word_forms_lower_form ON word_forms (LOWER(form));
