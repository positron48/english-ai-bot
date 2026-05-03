#!/usr/bin/env python3
"""Prepare English frequency CSV from ODS with rule-based and optional LLM cleanup."""

from __future__ import annotations

import argparse
import csv
import json
import os
import re
import unicodedata
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET
import zipfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any

NS = {
    "table": "urn:oasis:names:tc:opendocument:xmlns:table:1.0",
    "text": "urn:oasis:names:tc:opendocument:xmlns:text:1.0",
}

POS_MAP = {
    "n": "NOUN",
    "v": "VERB",
    "j": "ADJ",
    "r": "ADV",
}

ENGLISH_LEMMA_RE = re.compile(r"^[a-z]+(?:[-'][a-z]+)*$")
ASCII_ALPHA_RE = re.compile(r"^[a-z]+$")
REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_INPUT_ODS = REPO_ROOT / "wordFrequency.ods"
DEFAULT_OUTPUT_CSV = REPO_ROOT / "resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv"
DEFAULT_REPORT_JSON = REPO_ROOT / "resources/wordsets/english_word_freq_pos_ud_top6000.report.json"
DEFAULT_BLOCKLIST = REPO_ROOT / "resources/wordsets/english_lemma_blocklist.txt"


@dataclass(frozen=True)
class LemmaRow:
    rank: int
    lemma: str
    pos: str
    popularity_count: int
    source_pos: str
    caps_ratio: float


def normalize_lemma(lemma: str) -> str:
    return unicodedata.normalize("NFC", lemma.strip().lower())


def parse_float_maybe_comma(value: str) -> float:
    value = (value or "").strip().replace(",", ".")
    if not value:
        return 0.0
    try:
        return float(value)
    except ValueError:
        return 0.0


def parse_int_safe(value: str) -> int:
    value = (value or "").strip()
    if not value:
        return 0
    if "," in value and "." not in value:
        value = value.replace(",", "")
    try:
        return int(float(value))
    except ValueError:
        return 0


def load_blocklist(path: Path) -> set[str]:
    out: set[str] = set()
    if not path.exists():
        return out
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = normalize_lemma(raw)
        if not line or line.startswith("#"):
            continue
        out.add(line)
    return out


def iter_sheet_rows(ods_path: Path, sheet_name: str) -> list[list[str]]:
    with zipfile.ZipFile(ods_path) as zf:
        xml_bytes = zf.read("content.xml")
    root = ET.fromstring(xml_bytes)
    for table in root.findall(".//table:table", NS):
        name = table.get(f"{{{NS['table']}}}name", "")
        if name != sheet_name:
            continue
        out_rows: list[list[str]] = []
        for row in table.findall("table:table-row", NS):
            row_rep = int(row.get(f"{{{NS['table']}}}number-rows-repeated", "1"))
            values: list[str] = []
            for cell in row.findall("table:table-cell", NS):
                col_rep = int(cell.get(f"{{{NS['table']}}}number-columns-repeated", "1"))
                text = "".join((p.text or "") for p in cell.findall(".//text:p", NS))
                for _ in range(col_rep):
                    values.append(text)
            for _ in range(row_rep):
                out_rows.append(values[:])
        return out_rows
    raise ValueError(f"sheet {sheet_name!r} not found in {ods_path}")


def extract_rows_from_ods(ods_path: Path, sheet_name: str) -> list[LemmaRow]:
    rows = iter_sheet_rows(ods_path, sheet_name)
    header_idx = None
    for idx, row in enumerate(rows):
        lowered = [c.strip().lower() for c in row]
        if "rank" in lowered and "lemma" in lowered and "pos" in lowered and "freq" in lowered:
            header_idx = idx
            break
    if header_idx is None:
        raise ValueError("could not find expected header row in ODS sheet")

    header = [c.strip().lower() for c in rows[header_idx]]
    idx_rank = header.index("rank")
    idx_lemma = header.index("lemma")
    idx_pos = header.index("pos")
    idx_freq = header.index("freq")
    idx_caps = header.index("%caps") if "%caps" in header else -1

    out: list[LemmaRow] = []
    for row in rows[header_idx + 1 :]:
        if idx_lemma >= len(row):
            continue
        lemma_raw = (row[idx_lemma] if idx_lemma < len(row) else "").strip()
        if not lemma_raw:
            continue
        source_pos = (row[idx_pos] if idx_pos < len(row) else "").strip().lower()
        ud_pos = POS_MAP.get(source_pos)
        if not ud_pos:
            continue
        out.append(
            LemmaRow(
                rank=parse_int_safe(row[idx_rank] if idx_rank < len(row) else ""),
                lemma=normalize_lemma(lemma_raw),
                pos=ud_pos,
                popularity_count=parse_int_safe(row[idx_freq] if idx_freq < len(row) else ""),
                source_pos=source_pos,
                caps_ratio=parse_float_maybe_comma(row[idx_caps] if idx_caps >= 0 and idx_caps < len(row) else ""),
            )
        )
    return out


def is_vowelless_abbrev_ascii(lemma: str) -> bool:
    if len(lemma) < 2 or len(lemma) > 5:
        return False
    if not lemma.isascii() or not lemma.isalpha():
        return False
    return not any(ch in "aeiouy" for ch in lemma)


def is_rule_valid(lemma: str, blocklist: set[str]) -> bool:
    if not lemma or len(lemma) == 1:
        return False
    if lemma in blocklist:
        return False
    if is_vowelless_abbrev_ascii(lemma):
        return False
    if not ENGLISH_LEMMA_RE.match(lemma):
        return False
    return True


def dedupe_rows(rows: list[LemmaRow]) -> list[LemmaRow]:
    best: dict[tuple[str, str], LemmaRow] = {}
    for row in rows:
        key = (row.lemma, row.pos)
        current = best.get(key)
        if current is None:
            best[key] = row
            continue
        if row.popularity_count > current.popularity_count:
            best[key] = row
    return list(best.values())


def sort_rows(rows: list[LemmaRow]) -> list[LemmaRow]:
    return sorted(rows, key=lambda r: (-r.popularity_count, r.lemma, r.pos))


def suspicious_for_llm(row: LemmaRow) -> bool:
    # Cap-heavy noun candidates are likely to include named entities in COCA.
    if row.pos == "NOUN" and row.caps_ratio >= 0.35:
        return True
    # Very short noun/adj/adverb tokens are often artifacts.
    if row.pos in {"NOUN", "ADJ", "ADV"} and len(row.lemma) <= 2:
        return True
    # Pure ascii but with apostrophe/hyphen may be contractions/odd forms.
    if ("'" in row.lemma or "-" in row.lemma) and row.pos in {"NOUN", "ADJ", "ADV"}:
        return True
    return False


def strip_fenced_json(text: str) -> str:
    txt = text.strip()
    if txt.startswith("```"):
        txt = txt.strip("`")
        if txt.lower().startswith("json"):
            txt = txt[4:].strip()
    return txt


def llm_filter_rows(rows: list[LemmaRow], model: str, base_url: str, api_key: str, batch_size: int) -> tuple[set[tuple[str, str]], list[dict[str, Any]]]:
    if not rows:
        return set(), []
    endpoint = base_url.rstrip("/") + "/chat/completions"
    dropped: set[tuple[str, str]] = set()
    audit: list[dict[str, Any]] = []

    for i in range(0, len(rows), batch_size):
        batch = rows[i : i + batch_size]
        payload_items = [{"lemma": r.lemma, "pos": r.pos, "caps_ratio": r.caps_ratio} for r in batch]
        prompt = (
            "You receive English vocabulary candidates for language learning.\n"
            "Return strict JSON with key decisions: array of objects.\n"
            "Each object: lemma, pos, keep (true/false), reason.\n"
            "Keep normal common words. Drop proper names/entities, artifacts, malformed tokens, obvious noise.\n"
            "Do not hallucinate new lemmas.\n"
            f"Input:\n{json.dumps(payload_items, ensure_ascii=False)}"
        )
        body = {
            "model": model,
            "temperature": 0,
            "messages": [{"role": "user", "content": prompt}],
            "response_format": {"type": "json_object"},
        }
        req = urllib.request.Request(
            endpoint,
            data=json.dumps(body).encode("utf-8"),
            headers={
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=120) as resp:
                raw = resp.read().decode("utf-8")
        except urllib.error.URLError as err:
            raise RuntimeError(f"LLM request failed on batch {i // batch_size + 1}: {err}") from err
        parsed = json.loads(raw)
        content = parsed["choices"][0]["message"]["content"]
        content = strip_fenced_json(content)
        decision_obj = json.loads(content)
        decisions = decision_obj.get("decisions", [])
        index = {(r.lemma, r.pos): r for r in batch}
        for item in decisions:
            lemma = normalize_lemma(str(item.get("lemma", "")))
            pos = str(item.get("pos", "")).strip().upper()
            keep = bool(item.get("keep", True))
            reason = str(item.get("reason", "")).strip()
            key = (lemma, pos)
            if key not in index:
                continue
            if not keep:
                dropped.add(key)
                audit.append({"lemma": lemma, "pos": pos, "reason": reason})
    return dropped, audit


def write_csv(path: Path, rows: list[LemmaRow]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(
            f,
            fieldnames=["lemma", "pos", "popularity_count", "rank", "source_pos", "caps_ratio"],
        )
        writer.writeheader()
        for r in rows:
            writer.writerow(
                {
                    "lemma": r.lemma,
                    "pos": r.pos,
                    "popularity_count": r.popularity_count,
                    "rank": r.rank,
                    "source_pos": r.source_pos,
                    "caps_ratio": f"{r.caps_ratio:.6f}",
                }
            )


def main() -> int:
    parser = argparse.ArgumentParser(description="Prepare English frequency CSV from wordFrequency.ods")
    parser.add_argument("--input-ods", type=Path, default=DEFAULT_INPUT_ODS)
    parser.add_argument("--sheet", default="1 lemmas")
    parser.add_argument("--output-csv", type=Path, default=DEFAULT_OUTPUT_CSV)
    parser.add_argument("--report-json", type=Path, default=DEFAULT_REPORT_JSON)
    parser.add_argument("--blocklist", type=Path, default=DEFAULT_BLOCKLIST)
    parser.add_argument("--llm-batch-size", type=int, default=200)
    parser.add_argument("--skip-llm", action="store_true")
    args = parser.parse_args()

    rows = extract_rows_from_ods(args.input_ods, args.sheet)
    blocklist = load_blocklist(args.blocklist)
    rows = dedupe_rows(rows)

    kept_rule: list[LemmaRow] = []
    removed_rule: list[dict[str, Any]] = []
    for r in rows:
        if is_rule_valid(r.lemma, blocklist):
            kept_rule.append(r)
        else:
            removed_rule.append({"lemma": r.lemma, "pos": r.pos, "reason": "rule_filter"})

    kept_rule = sort_rows(kept_rule)
    llm_candidates = [r for r in kept_rule if suspicious_for_llm(r)]
    llm_dropped: set[tuple[str, str]] = set()
    llm_audit: list[dict[str, Any]] = []

    llm_enabled = not args.skip_llm
    base_url = os.getenv("AI_URL", "").strip()
    api_key = os.getenv("AI_API_KEY", "").strip()
    model = os.getenv("AI_MODEL_HIGH", "").strip() or os.getenv("AI_MODEL", "").strip()
    if llm_enabled and llm_candidates:
        if not base_url or not api_key or not model:
            raise RuntimeError("LLM filtering requested but AI_URL/AI_API_KEY/AI_MODEL(_HIGH) are not set")
        llm_dropped, llm_audit = llm_filter_rows(
            llm_candidates,
            model=model,
            base_url=base_url,
            api_key=api_key,
            batch_size=max(10, args.llm_batch_size),
        )

    final_rows = [r for r in kept_rule if (r.lemma, r.pos) not in llm_dropped]
    final_rows = sort_rows(final_rows)
    write_csv(args.output_csv, final_rows)

    by_pos: dict[str, int] = {}
    for r in final_rows:
        by_pos[r.pos] = by_pos.get(r.pos, 0) + 1
    report = {
        "input_ods": str(args.input_ods),
        "output_csv": str(args.output_csv),
        "sheet": args.sheet,
        "total_extracted_rows": len(rows),
        "rule_filtered_out": len(removed_rule),
        "llm_candidates": len(llm_candidates),
        "llm_filtered_out": len(llm_dropped),
        "final_rows": len(final_rows),
        "final_rows_by_pos": by_pos,
        "llm_enabled": llm_enabled,
        "llm_model": model if llm_enabled else "",
        "removed_samples_rule": removed_rule[:200],
        "removed_samples_llm": llm_audit[:200],
    }
    args.report_json.parent.mkdir(parents=True, exist_ok=True)
    args.report_json.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")

    print(
        f"prepared english csv: final_rows={len(final_rows)} "
        f"rule_removed={len(removed_rule)} llm_removed={len(llm_dropped)}"
    )
    for pos in sorted(by_pos):
        print(f"  {pos}: {by_pos[pos]}")
    print(f"csv: {args.output_csv}")
    print(f"report: {args.report_json}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

