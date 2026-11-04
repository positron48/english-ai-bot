-- Read-only production verification after rollout and ES content import.
-- Exactly two rows must have matches_expected = true before resolving #71/#73.
WITH expected(report_id, chapter_id, question_id, expected_question) AS (
 VALUES
  (71, 'es.grammar.noun_phrase_upgrade_precision.neuter_lo_lo_bueno_lo_importante', 'q18', '{"id": "q18", "type": "fill_blank", "theory_block_id": "b2_theory_lo_plus_adjective_form", "difficulty": 2, "prompt": "Вставьте одно слово, чтобы передать смысл «Странное в этом деле — молчание»: ___ raro del caso es el silencio.", "correct_answer": "lo", "explanation": "Lo + прилагательное обозначает абстрактный аспект: lo raro — «то, что странно». Здесь не называется конкретный предмет мужского или женского рода.", "tags": ["lo_adjective", "abstract"]}'::jsonb),
  (73, 'es.grammar.noun_phrase_upgrade_precision.build_speak_compare_cities_people_habits_options', 'q23', '{"id": "q23", "type": "fill_blank", "theory_block_id": "b2_theory_comparing_cities_and_places", "difficulty": 2, "prompt": "Вставьте одно слово, чтобы сказать: «Этот город такой же зелёный, как мой родной». Esta ciudad es ___ verde como mi ciudad natal.", "correct_answer": "tan", "explanation": "Для сравнения равных по степени качеств используется tan + прилагательное + como: tan verde como — «такой же зелёный, как». Перед прилагательным verde нужна форма tan, а не tanto.", "tags": []}'::jsonb)
)
SELECT e.report_id, e.question_id,
       COALESCE(q.value = e.expected_question, false) AS matches_expected,
       q.value->>'prompt' AS current_prompt,
       q.value->>'correct_answer' AS current_answer
FROM expected e
LEFT JOIN grammar_content_chapters c
  ON c.bundle_id = 'es' AND c.chapter_id = e.chapter_id
LEFT JOIN LATERAL jsonb_array_elements(c.raw_json::jsonb->'question_bank'->'questions') q(value)
  ON q.value->>'id' = e.question_id
ORDER BY e.report_id;
