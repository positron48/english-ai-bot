-- UNIQUE(word, course_code) added in 000038 does not, by itself, stop two
-- legacy/untagged rows (course_code IS NULL) from sharing the same word: a
-- plain UNIQUE constraint never treats NULL as equal to NULL, so it can't
-- catch that case. Add a partial unique index covering only course_code IS
-- NULL rows to restore the old "one row per word" guarantee for legacy data,
-- while still letting different courses each have their own row.

CREATE UNIQUE INDEX IF NOT EXISTS word_cards_word_legacy_key
	ON word_cards (word)
	WHERE course_code IS NULL;
