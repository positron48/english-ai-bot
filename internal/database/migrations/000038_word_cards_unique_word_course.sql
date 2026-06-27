-- Replace word_cards' UNIQUE(word) with UNIQUE(word, course_code).
--
-- word_cards used to be unique on `word` alone, which meant a word spelled the
-- same in two courses (e.g. "real"/"social"/"hotel" in en_ru and es_ru) was
-- forced to share a single row, even though translation, training cards and
-- pronunciation are entirely course-specific. This was called out as
-- intentionally deferred in 000022_linglow_course_scoped_word_training.sql;
-- this migration completes it now that the per-course code paths
-- (UpsertWordCardLemma, GetWordCardByLemma, etc.) are course-aware.
--
-- Rows with course_code IS NULL keep allowing duplicate `word` values (NULL is
-- never considered equal to NULL by a UNIQUE constraint), matching prior
-- behavior for legacy/untagged rows.

DO $$
DECLARE
    cname text;
BEGIN
    SELECT conname INTO cname
    FROM pg_constraint
    WHERE conrelid = 'word_cards'::regclass
      AND contype = 'u'
      AND pg_get_constraintdef(oid) = 'UNIQUE (word)';

    IF cname IS NOT NULL THEN
        EXECUTE format('ALTER TABLE word_cards DROP CONSTRAINT %I', cname);
    END IF;
END $$;

-- SQLite imports may leave a standalone unique index on word (not a pg_constraint).
DROP INDEX IF EXISTS idx_18611_sqlite_autoindex_word_cards_1;

DO $$
DECLARE
    idx_name text;
BEGIN
    FOR idx_name IN
        SELECT ci.relname
        FROM pg_index i
        JOIN pg_class ct ON ct.oid = i.indrelid
        JOIN pg_class ci ON ci.oid = i.indexrelid
        WHERE ct.relname = 'word_cards'
          AND i.indisunique
          AND NOT i.indisprimary
          AND i.indpred IS NULL
          AND (
            SELECT array_agg(a.attname ORDER BY u.ord)
            FROM unnest(i.indkey) WITH ORDINALITY AS u(attnum, ord)
            JOIN pg_attribute a ON a.attrelid = ct.oid AND a.attnum = u.attnum
          ) = ARRAY['word']::name[]
    LOOP
        EXECUTE format('DROP INDEX IF EXISTS %I', idx_name);
    END LOOP;
END $$;

ALTER TABLE word_cards ADD CONSTRAINT word_cards_word_course_code_key UNIQUE (word, course_code);
