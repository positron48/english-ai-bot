# Standalone placement API

All endpoints require the normal Bearer access token and return `Cache-Control: no-store`. Course codes are `en_ru` and `es_ru`. Questions are from the independent placement banks, not grammar chapter exercises.

## Start or resume

`POST /api/learning/placement/sessions`

```json
{"course_code":"es_ru","idempotency_key":"a-client-generated-uuid","new_attempt":false}
```

A repeated key returns the same attempt, including its completed result. A different key with `new_attempt:false` resumes the active attempt for that user/course. `new_attempt:true` abandons any active attempt and creates a new variant. Expired attempts may be replaced with a new key. A session expires seven days after creation. Every accepted start key is retained even when it resumes a session created with another key. Lookup happens before loading the current bank, so a pinned attempt can resume during a temporary bank outage.

Response shape:

```json
{
  "id":"32-character-session-id",
  "course_code":"es_ru",
  "status":"active",
  "bank_version":"sha256-of-canonical-bank",
  "policy_version":"editorial-v1",
  "questions":[{
    "id":"es.p.example",
    "context":"Короткая ситуация.",
    "instruction":"Дополните сообщение.",
    "prompt":"Voy ___ casa.",
    "choices":[{"id":"a","text":"a"},{"id":"b","text":"en"},{"id":"c","text":"de"}]
  }],
  "answers":{},
  "base_closed":false,
  "clarifying":false
}
```

The initial response contains all 30 public questions. Treat all delivered questions as exposed for future selection, even if the user does not view every card. Keys, explanations, skill names, levels, family IDs and the pinned clarification reserve are not included.

## Resume and answer

`GET /api/learning/placement/sessions/{id}` returns the same shape. Ownership is checked; a different user's session returns 404.

`POST /api/learning/placement/sessions/{id}/answers`

```json
{"question_id":"es.p.example","answer":"a"}
```

The empty string is an explicit “I don't know”. Missing/null answers are invalid. Only an issued question and one of its stable choice IDs (or the empty string) are accepted. Each response returns the updated session.

Answers in the base block can be changed until all 30 have been answered. The thirtieth answer closes that block and, if needed, appends six pinned clarification questions. The base cannot then be changed; replaying the same answer remains idempotent. The client should use `questions.length`, not a hard-coded length. No correctness feedback is returned before completion.

## Finish

`POST /api/learning/placement/sessions/{id}/finish`

Finishing an incomplete session returns 409. Finishing a complete session stores the result and course access atomically. Repeating finish returns the existing result without applying access again.

A completed session also includes `available_chapter_ids`: currently usable links among the result topics, selected with the course access rules. This is not the full accessible catalog. A topic whose lessons are still locked remains visible, but the UI does not link to an inaccessible lesson.

The session's `result` includes:

- `level`, `upper_level`: estimated course level and upper boundary when borderline; initial state is `below_a1`;
- `estimated:true`, `policy_version`;
- `correct`, `total`, `profile:[{level,correct,total,status}]`;
- `review`: public question plus revision, level, skill title, user's answer, correct answer, explanation and published chapter IDs;
- `recommended_skills`: up to five skills worth reviewing and links to published chapters;
- `opened_sections`: sections eligible to open from this attempt. Previously granted access may be broader and is preserved.

`GET /api/learning/placement/results?course_code=es_ru` returns `{sessions:[...]}`, newest completed first, up to ten. Historical results retain their bank and policy versions. Every result response filters course links against the current publication state of both the chapter and its parent section; withdrawing content does not change the recorded score. The current recommendation and maximum course access are intentionally distinct.

## Errors and old clients

- 400 `placement_invalid_request` / `placement_invalid_answer`;
- 401 `unauthorized`;
- 404 `placement_not_found` / `placement_course_not_found`;
- 409 `placement_conflict`;
- 410 `placement_expired`;
- 503 `placement_unavailable` (including a missing active DB bank).

The old `/api/learning/grammar/placement-test` and `/submit` routes return 410 `placement_replaced` with the new UI path. Old clients must reload; they must not submit chapter questions or compute an offline substitute level.

## Editorial policy v1

This is a transparent initial course-placement heuristic, not a calibrated CEFR examination.

- Base form: six questions per level, two at each editorial difficulty 1–3. Selection favours unseen families and different skills while retaining difficulty quotas.
- Recent history: the last 12 sessions, including abandoned sessions. When a stratum has no fresh questions, prefer the least recent exposure; a family is never repeated inside one attempt.
- Profile: at least six observations at a level; >=75% is `secure`, >=50% is `borderline`, otherwise `limited`.
- Clarify one level with six additional questions: first an isolated high success whose preceding level is limited; otherwise the borderline level directly above the highest secure level; otherwise a lower limited gap below a secure level; otherwise the lowest borderline level.
- Grade: choose the highest secure level with sufficient lower evidence (at least 50% across lower levels) and a preceding level that is not limited, or 12 observations confirming the candidate. A1 needs no lower evidence. This allows isolated lower gaps without accepting an isolated high-level guess as a full profile.
- A borderline next level is returned as an upper boundary. Results always say they are estimated. Time is not part of scoring.
- The selected base, unseen clarification reserve, skill map, choice order and keys are snapshotted at start, so changing a live bank does not change an active attempt. New policy versions must retain support for active older versions or explicitly expire those attempts.

Author review, structural validation and code simulations are recorded separately from the still-pending pilot with learners.
