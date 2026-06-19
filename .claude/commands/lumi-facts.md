# lumi-facts skill

Generates and inserts 10 new Lumi facts into the next available migration file.
Facts are added for ALL contexts × ALL courses.

## What to do

1. Find the highest-numbered migration in `internal/database/migrations/` to determine the next number (pad to 6 digits).
2. Generate 10 new, unique, **interesting** facts in Russian (`locale = 'ru'`) for **each combination** of:
   - Courses: `en_ru`, `es_ru`
   - Contexts: `general`, `grammar`, `reading`, `practice`, `progress`, `city`
   - That's 10 × 2 × 6 = **120 facts total**.
3. Facts must be specific, engaging, and non-trivial — avoid generic motivational phrases.
   - `general`: curious facts about the language itself (etymology, history, records)
   - `grammar`: specific grammar rules, quirks, or comparisons with Russian
   - `reading`: tips, genre facts, author trivia, text-level knowledge
   - `practice`: evidence-based learning techniques, study habits, memory science
   - `progress`: milestones, CEFR levels, motivation backed by research
   - `city`: geography, history, architecture, culture of cities where the language is spoken
4. Each INSERT block must have a `WHERE NOT EXISTS` guard that is **unique per context+course**
   (e.g., check for a distinctive substring from one fact in that batch using `AND body LIKE '%...%'`).
5. Write the migration file to `internal/database/migrations/NNNNNN_lumi_facts_batch_<timestamp>.sql`
   following the exact same SQL style as `000029_lumi_facts_all_contexts_seed.sql`.
6. After creating the file, print a summary table: course × context → count of facts added.

## Example INSERT block structure

```sql
-- ENGLISH (en_ru) — GENERAL context
INSERT INTO lumi_facts (course_code, context, locale, body)
SELECT 'en_ru', 'general', 'ru', f.body
FROM (VALUES
    ('Fact text here...'),
    ...
) AS f(body)
WHERE NOT EXISTS (
    SELECT 1 FROM lumi_facts WHERE course_code = 'en_ru' AND context = 'general' AND body LIKE '%unique phrase%'
);
```

Do not add facts that duplicate any of the following existing batches:
- `000026_lumi_facts.sql` — general context seed (10 facts each)
- `000028_lumi_facts_grammar_seed.sql` — grammar context (50 facts each)
- `000029_lumi_facts_all_contexts_seed.sql` — all other contexts (15 facts each)
