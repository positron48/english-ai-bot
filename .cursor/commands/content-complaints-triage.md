# Content complaints triage

Запусти навык **`content-complaints-triage`**.

1. Спроси: `course=en|es` (или оба в одном журнале), `dry-run|apply`.
2. `secrets/complaints-prod.env` из `env.example.complaints-prod`.
3. При **apply**: `make complaints-journal-new` → вести `docs/complaints/journal-YYYY-MM-DD-*.md`.
4. `make complaints-triage-dry-en` / `dry-es` или `fetch_reports.py`.
5. Фазы A–F в `.cursor/skills/content-complaints-triage/SKILL.md`.
6. Закоммитить `docs/complaints/journal-*.md`.

Справка: `docs/complaints/README.md`.
