# Журнал триажа content reports

**Прогон:** YYYY-MM-DD (prod EN `qantrix.ru`, ES `es.qantrix.ru`)  
**Run ID:** `triage-YYYY-MM-DD`  
**Снимок:** `logs/complaints/snapshot-en-*.json`, `snapshot-es-*.json` (локально, не в git)  
**Тег / коммит:** (заполнить после релиза, например `0.11.13`)  
**Resolve на prod:** (дата, EN N + ES M, `resolve_all_active.py` или resolve-bulk)

Формат блока: **дата жалобы** → **суть** → **что изменено**.

---

## English (grammar_training | word_training)

### #<id> — YYYY-MM-DD

**Жалоба:** …

**Изменено:** …

---

## Spanish (grammar_training | word_training)

### #<id> — YYYY-MM-DD

**Жалоба:** …

**Изменено:** …

---

## Правила / pattern sweep

(если добавлялись guardrails или массовые grep-фиксы)

---

## Артефакты прогона

| Файл | Назначение |
|------|------------|
| `docs/complaints/journal-YYYY-MM-DD-<slug>.md` | этот журнал (в git) |
| `logs/complaints/triage-YYYY-MM.jsonl` | машинный журнал (локально) |
| `logs/complaints/snapshot-*.json` | снимок жалоб с prod |
