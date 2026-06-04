# Content complaints triage

Запусти навык `content-complaints-triage`.

1. Спроси: `course=en|es`, `dry-run|apply`.
2. Создай `secrets/complaints-prod.env` из `env.example.complaints-prod` и загрузи env.
3. `python3 tools-local/complaints-triage/fetch_reports.py --course <course>`.
4. Следуй фазам A–F из `.cursor/skills/content-complaints-triage/SKILL.md`.
