#!/usr/bin/env python3
"""Score the canonical Spanish reading catalog with GPT-6 Luna.

Usage: python3 tools/reading-quality/score_spanish_reading.py pilot|full|export
The raw batch responses are checkpoints, so `full` can be restarted safely.
"""

import csv
import hashlib
import json
import subprocess
import sys
import tempfile
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "courses/spanish-grammar/reading/texts"
OUTPUT = ROOT / "docs/reading-quality/2026-09-27"
MODEL = "gpt-6-luna"
DIMENSIONS = {
    "content": 3,
    "spanish": 2,
    "level_fit": 1,
    "translation": 1,
    "questions": 2,
    "learning_value": 1,
}
RUBRIC = """Оцени КАЖДЫЙ учебный текст для чтения на испанском по шкале 0–10, сумма шести целых баллов:
- content 0–3: цельность, логика, содержательность и отсутствие явных фактических ошибок. 0 = бессвязен/существенно неверен; 1 = искусственный перечень связанных темой фактов или реплик, между которыми нет мотивированной связи; 2 = есть содержательное развитие или внятная жизненная ситуация, но поверхностно; 3 = цельный и интересный относительно уровня.
- spanish 0–2: грамматика и естественность испанского. 0 = явные/частые ошибки, 1 = понятный, но заметно неестественный (например, реплики отвечают друг другу формально, связки не выражают реального отношения), 2 = естественный.
- level_fit 0–1: соответствует указанному CEFR уровню; 0 при существенном расхождении, 1 если соответствует.
- translation 0–1: русский перевод передает испанский смысл естественно и точно; 0 при заметной кальке или смысловой ошибке.
- questions 0–2: верны ли ответы и проверяется ли понимание; 0 = неверный ключ/формулировка или вопросы практически не проверяют понимание, 1 = ключи верны, но вопросы в основном буквальные/однообразные, 2 = точные и разнообразные вопросы по смыслу.
- learning_value 0–1: текст дает полезный контекст, лексику, культурное знание или маленький сюжет для своего уровня; 0 если это просто заучивание разрозненных фактов без объяснения/контекста.

Правила: оценивай относительно заявленного уровня. Короткий A0/A1 текст не штрафуй только за краткость. Не завышай балл за формальную грамматическую правильность, если это механический список фактов или искусственный диалог. Пример: в диалоге о рождественском ужине после 24 декабря без причины перечисляются елка, вертеп, 31 декабря и 6 января; формальные «Entonces», «Sí» и «Después» не создают связного разговора. Такому тексту нельзя давать content выше 1 и spanish выше 1, даже если грамматика правильная. Не переноси эту оценку автоматически на другие тексты: оцени каждый по его содержимому.
Проверь каждую пару ES/RU и все ключи true/false. Перед замечанием о неверном переводе местоимения проверь ближайший возможный антецедент в полном контексте, в том числе предыдущего предложения; не выдумывай ошибок там, где перевод корректен. Не утверждай внешнюю фактическую ошибку без надежных оснований. Укажи только реальные конкретные недостатки с короткими цитатами. Если проблем нет, issues может быть пустым. Для каждого id верни ровно одну запись. Ответ строго по JSON-схеме. Никаких инструментов или работы с файлами.
"""


def source_docs():
    docs = {}
    for path in sorted(SOURCE.glob("*.json")):
        raw = path.read_bytes()
        doc = json.loads(raw)
        doc_id = doc["id"]
        if doc_id in docs:
            raise ValueError(f"duplicate ID: {doc_id}")
        passage = doc["reading_passage"]
        docs[doc_id] = {
            "id": doc_id,
            "title": doc["title"],
            "level": doc["level"],
            "segments": [
                {
                    "speaker": s.get("speaker_id", ""),
                    "es": s["text"],
                    "ru": s["text_translation_ru"],
                }
                for s in passage["segments"]
            ],
            "questions": [
                {
                    "prompt": q["prompt"],
                    "answer": q["correct_answer"],
                    "explanation": q.get("explanation", ""),
                }
                for q in passage["comprehension_questions"]
            ],
            "source_path": str(path.relative_to(ROOT)),
            "sha256": hashlib.sha256(raw).hexdigest(),
        }
    return docs


def schema():
    props = {"id": {"type": "string"}}
    props.update({k: {"type": "integer", "minimum": 0, "maximum": v} for k, v in DIMENSIONS.items()})
    props["reason"] = {"type": "string"}
    props["issues"] = {"type": "array", "items": {"type": "string"}}
    return {
        "type": "object",
        "properties": {
            "scores": {
                "type": "array",
                "items": {"type": "object", "properties": props, "required": list(props), "additionalProperties": False},
            }
        },
        "required": ["scores"],
        "additionalProperties": False,
    }


def compact(doc):
    return {k: doc[k] for k in ("id", "title", "level", "segments", "questions")}


def call_luna(batch, result_path):
    prompt = RUBRIC + "\nДАННЫЕ:\n" + json.dumps([compact(d) for d in batch], ensure_ascii=False, separators=(",", ":"))
    with tempfile.TemporaryDirectory(prefix="reading-luna-") as tmp:
        schema_path = Path(tmp) / "schema.json"
        response_path = Path(tmp) / "response.json"
        schema_path.write_text(json.dumps(schema(), ensure_ascii=False))
        command = [
            "/Applications/ChatGPT.app/Contents/Resources/codex", "exec",
            "-C", "/tmp", "--skip-git-repo-check", "--ephemeral",
            "--ignore-user-config", "-s", "read-only", "-m", MODEL,
            "-c", 'model_reasoning_effort="medium"',
            "--output-schema", str(schema_path), "--output-last-message", str(response_path), "-",
        ]
        run = subprocess.run(command, input=prompt, text=True, capture_output=True, timeout=300)
        if run.returncode or not response_path.exists():
            raise RuntimeError(f"Luna failed ({run.returncode}): {run.stderr[-3000:]}")
        response = json.loads(response_path.read_text())
        validate(response, batch)
        result_path.parent.mkdir(parents=True, exist_ok=True)
        result_path.write_text(json.dumps(response, ensure_ascii=False, indent=2) + "\n")
        return response


def validate(response, batch):
    scores = response["scores"]
    ids = [s["id"] for s in scores]
    expected = [d["id"] for d in batch]
    if Counter(ids) != Counter(expected):
        raise ValueError(f"response IDs differ: expected {expected}, got {ids}")
    for score in scores:
        for key, maximum in DIMENSIONS.items():
            value = score[key]
            if type(value) is not int or not 0 <= value <= maximum:
                raise ValueError(f"invalid {key} for {score['id']}: {value}")
        if not score["reason"].strip():
            raise ValueError(f"missing reason for {score['id']}")


def pilot(docs):
    examples = [
        "free_es_a1_cena_de_nochebuena_1782924757",
        "free_es_a1_cava_catal_n_1782924852",
        "free_es_a1_la_tienda_nueva_1782925044",
        "free_es_b2_el_tribunal_del_agua_1782407229",
        "free_es_a0_mi_nombre_completo_1782924463",
    ]
    batch = [docs[i] for i in examples]
    corrupted = json.loads(json.dumps(docs[examples[2]]))
    corrupted["id"] = "pilot_intentionally_corrupted_tienda"
    corrupted["segments"][0]["ru"] = "Новый магазин закрывается пятого июня в девять."
    corrupted["questions"][0]["answer"] = "false"
    batch.append(corrupted)
    result = call_luna(batch, OUTPUT / "pilot.json")
    by_id = {s["id"]: s for s in result["scores"]}
    for doc in batch:
        score = by_id[doc["id"]]
        total = sum(score[k] for k in DIMENSIONS)
        print(f"{total:2d}/10 {doc['id']}: {score['reason']}", flush=True)
        for issue in score["issues"]:
            print(f"   - {issue}", flush=True)
    bad = by_id[examples[0]]
    good = by_id[examples[1]]
    clean = by_id[examples[2]]
    broken = by_id[corrupted["id"]]
    assert sum(good[k] for k in DIMENSIONS) - sum(bad[k] for k in DIMENSIONS) >= 3, "weak/strong separation failed"
    assert broken["translation"] == 0 and broken["questions"] == 0, "injected defects missed"
    assert sum(clean[k] for k in DIMENSIONS) - sum(broken[k] for k in DIMENSIONS) >= 2, "injected defects not reflected"
    assert not any("местоимени" in issue or "принадлежит собору" in issue for issue in by_id[examples[3]]["issues"]), "fabricated pronoun error"
    print("PILOT PASS", flush=True)


def full(docs):
    keys = sorted(docs)
    size = 10
    pending = []
    for start in range(0, len(keys), size):
        batch = [docs[k] for k in keys[start:start + size]]
        path = OUTPUT / "batches" / f"{start//size+1:03d}.json"
        if path.exists():
            validate(json.loads(path.read_text()), batch)
            print(f"batch {start//size+1:03d}: checkpoint", flush=True)
            continue
        pending.append((start, batch, path))
    with ThreadPoolExecutor(max_workers=3) as pool:
        futures = {pool.submit(call_luna, batch, path): (start, batch, path) for start, batch, path in pending}
        for future in as_completed(futures):
            start, batch, path = futures[future]
            try:
                future.result()
            except Exception as error:
                print(f"batch {start//size+1:03d}: FAILED: {error}", flush=True)
                raise
            print(f"batch {start//size+1:03d}: {min(start+size,len(keys))}/{len(keys)}", flush=True)
    export(docs)


def export(docs):
    batch_files = sorted((OUTPUT / "batches").glob("*.json"))
    records = {}
    for path in batch_files:
        for score in json.loads(path.read_text())["scores"]:
            key = score["id"]
            if key in records:
                raise ValueError(f"duplicate result: {key}")
            records[key] = score
    if set(records) != set(docs):
        raise ValueError(f"missing {len(set(docs)-set(records))}, extra {len(set(records)-set(docs))}")
    fields = ["id", "title", "level", "score_10", *DIMENSIONS, "reason", "issues", "source_path", "source_sha256", "model"]
    OUTPUT.mkdir(parents=True, exist_ok=True)
    with (OUTPUT / "scores.csv").open("w", encoding="utf-8", newline="") as stream:
        writer = csv.DictWriter(stream, fieldnames=fields, lineterminator="\n")
        writer.writeheader()
        for key in sorted(docs):
            score = records[key]
            doc = docs[key]
            writer.writerow({
                "id": key, "title": doc["title"], "level": doc["level"],
                "score_10": sum(score[k] for k in DIMENSIONS),
                **{k: score[k] for k in DIMENSIONS},
                "reason": score["reason"], "issues": " | ".join(score["issues"]),
                "source_path": doc["source_path"], "source_sha256": doc["sha256"], "model": MODEL,
            })
    print(f"exported {len(records)} scores to {OUTPUT / 'scores.csv'}", flush=True)


if __name__ == "__main__":
    if len(sys.argv) != 2 or sys.argv[1] not in {"pilot", "full", "export"}:
        sys.exit("usage: score_spanish_reading.py pilot|full|export")
    documents = source_docs()
    {"pilot": pilot, "full": full, "export": export}[sys.argv[1]](documents)
