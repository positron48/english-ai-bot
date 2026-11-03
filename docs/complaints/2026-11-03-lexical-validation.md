# Lexical validation follow-up, 2026-11-03

Root cause in current sources: both Spanish dictionary and training prompts forced acceptance of alphabetic tokens. The dictionary demanded a plausible entry for alphabetic tokens of length >=4; training allowed rejection only for digits and three exact policy exclusions. This explains how the generation contract can admit a misspelling and invented translation; the original request/log/model that created word card 298537 is still not established.

Changes:
- Both prompts now reject misspellings/nonwords rather than inventing a definition or silently substituting a real word. Genuine inflections, rare words and established borrowings remain valid.
- `apatecer`/`apetecer` are explicit regression examples; natural translation is хотеться, not апатировать.
- Runtime validation rejects an inconsistent generated canonical word/lemma and contradictory accepted/rejected output. It does not compare the original inflected input with its normalized lemma, so valid morphology remains supported.
- No claim of complete dictionary verification: self-consistent LLM hallucinations can still evade these checks. No new paid call was added to normal card generation.

Validation:
- Full service package tests PASS, including new spelling-consistency and legitimate-inflection regressions.
- Full golangci-lint PASS (0 issues).
- Earlier API release checks: 5292 Go tests PASS, integration/frontend/coverage PASS; final lint PASS after HTTP constants correction.
- Opt-in `scripts/word-card-lexical-smoke.py`: fixed four cases per invocation, max 700 output tokens each, no automatic retries, no database writes. Default invocation prints the test plan and makes no API calls.
- Cheap cloud model used for this limited smoke: `openai/gpt-4o-mini` via OpenRouter, not a claim about production model equivalence.
- First training smoke: 3/4, software incorrectly rejected; $0.001059. Prompt clarified for established borrowings.
- Final training smoke: 4/4 PASS; $0.0013638.
- Final dictionary smoke (`--kind dictionary`): 4/4 PASS; $0.0015648.
- Total: 12 calls, provider-reported cost **$0.0039876**. No more paid runs.

Run after exporting the API key: `python3 scripts/word-card-lexical-smoke.py --live` and `python3 scripts/word-card-lexical-smoke.py --live --kind dictionary`. Local results: logs/complaints/fixes-20261103/*smoke*.log.

API rename commit bff720e6 was pushed and tagged 0.12.237. Deletion of incorrect word card 298537 is left to the user as requested; the rename API remains available after rollout. No production word deletion was executed by the agent.
