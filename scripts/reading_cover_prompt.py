"""LLM prompt builder for reading cover images."""
from __future__ import annotations

import json
import os
import pathlib
import re
import sys
import time

_REPO_SCRIPTS = pathlib.Path(__file__).resolve().parent
if str(_REPO_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_REPO_SCRIPTS))

import reading_llm_client as rlc  # noqa: E402
from reading_cover_log import log  # noqa: E402

# qwen3:30b needs headroom: reasoning can eat 2k+ tokens before JSON lands in content.
# Cover LLM input: opening sentences only (scene is set in the first lines).
MAX_COVER_SENTENCES = 3
MAX_COVER_SENTENCES_COMPACT = 2
MAX_SOURCE_CHARS = 400
MAX_COVER_SOURCE_LINES = 6  # non-cover plain_text (RU pairs)
MAX_COVER_SOURCE_LINES_COMPACT = 4
MAX_SOURCE_CHARS_COMPACT = 260

_SENTENCE_SPLIT_RE = re.compile(r"(?<=[.!?…])\s+")

DEFAULT_COVER_STYLE_PREFIX = "casual watercolor illustration, "

MOCK_SCENE_PROMPT = "two people at a small cafe table on a quiet street"

_IMAGE_PROMPT_KEYS = ("image_prompt", "prompt", "cover_image_prompt")

_COVER_FEWSHOT_EXAMPLE = (
    "Example:\n"
    "Title: Tapas\n"
    'Output: {"image_prompt": "Two friends sharing tapas at a Spanish bar."}'
)


def cover_style_prefix() -> str:
    v = os.getenv("READING_COVER_STYLE_PREFIX")
    return DEFAULT_COVER_STYLE_PREFIX if v is None else v


def cover_style_suffix() -> str:
    return os.getenv("READING_COVER_STYLE_SUFFIX", "").strip()


def normalize_scene_text(scene: str) -> str:
    return " ".join((scene or "").split()).strip().rstrip(".,;")


def finalize_cover_prompt(scene_prompt: str) -> str:
    """Prepend ComfyUI style prefix; optional READING_COVER_STYLE_SUFFIX env only."""
    scene = normalize_scene_text(scene_prompt)
    if not scene:
        return ""
    prefix = cover_style_prefix()
    suffix = cover_style_suffix()
    if prefix and not prefix.endswith((" ", ",")):
        prefix = prefix + " "
    parts = [prefix, scene]
    if suffix:
        suf = suffix if suffix.startswith(",") else ", " + suffix
        parts.append(suf)
    return "".join(parts)


def split_sentences(text: str) -> list[str]:
    """Split passage into sentences (Spanish/English punctuation)."""
    raw = " ".join((text or "").split()).strip()
    if not raw:
        return []
    parts = [p.strip() for p in _SENTENCE_SPLIT_RE.split(raw) if p.strip()]
    return parts or [raw]


def extract_leading_sentences(text: str, *, max_sentences: int, max_chars: int) -> str:
    """First N complete sentences that fit the char budget."""
    sentences = split_sentences(text)
    if not sentences:
        return ""
    picked: list[str] = []
    total = 0
    for sentence in sentences:
        if len(picked) >= max_sentences:
            break
        extra = len(sentence) + (1 if picked else 0)
        if picked and total + extra > max_chars:
            break
        if not picked and len(sentence) > max_chars:
            return sentence[: max_chars - 1].rstrip() + "…"
        picked.append(sentence)
        total += extra
    return " ".join(picked)


def plain_text_from_doc(doc: dict, *, for_cover: bool = False, compact: bool = False) -> str:
    title = str(doc.get("title") or "").strip()
    rp = doc.get("reading_passage") or {}
    if not title:
        title = str(rp.get("title") or "").strip()
    parts: list[str] = []
    if title:
        parts.append(f"TITLE: {title}")
    if for_cover or compact:
        body_parts: list[str] = []
        for seg in rp.get("segments") or []:
            if isinstance(seg, dict):
                text = str(seg.get("text") or "").strip()
                if text:
                    body_parts.append(text)
        max_sentences = MAX_COVER_SENTENCES_COMPACT if compact else MAX_COVER_SENTENCES
        max_chars = MAX_SOURCE_CHARS_COMPACT if compact else MAX_SOURCE_CHARS
        excerpt = extract_leading_sentences(
            " ".join(body_parts),
            max_sentences=max_sentences,
            max_chars=max_chars,
        )
        if excerpt:
            parts.append(excerpt)
        return "\n".join(parts)
    max_segments = None
    seg_count = 0
    for seg in rp.get("segments") or []:
        if max_segments is not None and seg_count >= max_segments:
            break
        if isinstance(seg, dict):
            text = str(seg.get("text") or "").strip()
            if not text:
                continue
            seg_count += 1
            ru = str(seg.get("text_translation_ru") or "").strip()
            if ru:
                parts.append(f"{text} — {ru}")
            else:
                parts.append(text)
    return "\n".join(parts)


def build_cover_prompt_messages(source_text: str, target_lang: str, title: str = "") -> list[dict]:
    """Few-shot JSON task — opening sentences are enough for a cover scene."""
    lang = (target_lang or "en").strip().lower()
    topic = (title or "").strip() or "reading text"
    return [
        {
            "role": "user",
            "content": (
                f"{_COVER_FEWSHOT_EXAMPLE}\n\n"
                f"Title: {topic}\n"
                f"Passage language: {lang}\n"
                f"Passage:\n{source_text}\n"
                "Output:"
            ),
        }
    ]


def _image_prompt_field_values(text: str) -> list[str]:
    vals: list[str] = []
    for key in _IMAGE_PROMPT_KEYS:
        for m in re.finditer(rf'"{key}"\s*:\s*"([^"]*)"', text, flags=re.I):
            val = m.group(1).strip()
            if val:
                vals.append(val)
    return vals


def _salvage_image_prompt_value(text: str) -> str:
    """Truncated stream JSON: {"image_prompt": "scene without closing quote."""
    m = re.search(r'"image_prompt"\s*:\s*"((?:[^"\\]|\\.)*)', text, flags=re.I | re.S)
    if not m:
        return ""
    val = m.group(1).replace('\\"', '"').strip()
    return val


def parse_image_prompt_json_only(text: str) -> str:
    if not text:
        return ""
    vals = _image_prompt_field_values(text)
    if vals:
        return vals[-1]
    blob = rlc.extract_json_object(text)
    if blob:
        try:
            data = json.loads(blob)
            if isinstance(data, dict):
                for key in _IMAGE_PROMPT_KEYS:
                    val = str(data.get(key) or "").strip()
                    if val:
                        return val
        except json.JSONDecodeError:
            salvaged = _salvage_image_prompt_value(blob)
            if salvaged:
                return salvaged
    salvaged = _salvage_image_prompt_value(text)
    if salvaged:
        return salvaged
    return ""


def parse_image_prompt(raw: str, title: str = "") -> str:
    """Extract image_prompt from JSON in LLM output (never the whole reasoning blob)."""
    _ = title
    text = rlc.strip_qwen_thinking(raw or "")
    parsed = parse_image_prompt_json_only(text)
    if parsed:
        return normalize_scene_text(parsed)
    preview = text[:400].replace("\n", " ")
    raise ValueError(
        "LLM не вернул JSON image_prompt (только рассуждение без JSON). "
        f"Длина ответа={len(text)}. Начало: {preview!r}. "
        "Для qwen3: /no_think, enable_thinking=false, READING_COVER_MAX_TOKENS>=4096."
    )


def extract_scene_candidates(raw: str) -> list[str]:
    """All image_prompt values found in output (for tests/diagnostics)."""
    text = rlc.strip_qwen_thinking(raw or "")
    seen: set[str] = set()
    ordered: list[str] = []

    def add(cand: str) -> None:
        cand = normalize_scene_text(cand)
        if not cand or cand in seen:
            return
        seen.add(cand)
        ordered.append(cand)

    for val in _image_prompt_field_values(text):
        add(val)
    blob = rlc.extract_json_object(text)
    if blob:
        try:
            data = json.loads(blob)
            if isinstance(data, dict):
                for key in _IMAGE_PROMPT_KEYS:
                    val = str(data.get(key) or "").strip()
                    if val:
                        add(val)
        except json.JSONDecodeError:
            salvaged = _salvage_image_prompt_value(blob)
            if salvaged:
                add(salvaged)
    salvaged = _salvage_image_prompt_value(text)
    if salvaged:
        add(salvaged)
    return ordered


def _request_cover_scene(
    messages: list[dict],
    course_root: pathlib.Path,
    *,
    label: str,
) -> str:
    prompt = messages[0]["content"]
    log(f"LLM: {label}")
    t0 = time.time()
    raw = rlc.chat_completion(
        prompt,
        course_root,
        temperature=0.1,
        prompt_profile="cover",
    )
    elapsed = time.time() - t0
    log(f"LLM: response in {elapsed:.1f}s ({len(raw)} chars)")
    return parse_image_prompt(raw)


def build_cover_prompt(course_root: pathlib.Path, doc: dict) -> str:
    target_lang = str(doc.get("target_language") or "en").strip().lower()
    title = str(doc.get("title") or "").strip()
    if not title:
        title = str((doc.get("reading_passage") or {}).get("title") or "").strip()
    text_id = str(doc.get("id") or "").strip()
    if os.getenv("READING_COVER_PROMPT_MOCK", "").strip():
        log(f"LLM: text_id={text_id or '?'} title={title or '?'} READING_COVER_PROMPT_MOCK set — using mock scene")
        scene = MOCK_SCENE_PROMPT
    else:
        scene = ""
        for attempt, compact in enumerate((False, True)):
            source = plain_text_from_doc(doc, for_cover=True, compact=compact)
            if not source.strip():
                raise ValueError("reading text has no content for cover prompt")
            label = "cover scene" if attempt == 0 else "cover scene (fewer sentences)"
            log(
                f"LLM: text_id={text_id or '?'} title={title or '?'} "
                f"source={len(source)} chars lang={target_lang} compact={compact}"
            )
            log("LLM: requesting scene description from local llama.cpp…")
            try:
                scene = _request_cover_scene(
                    build_cover_prompt_messages(source, target_lang, title),
                    course_root,
                    label=label,
                )
                break
            except ValueError:
                if attempt == 1:
                    raise
                log("LLM: no JSON in response — retry with fewer opening sentences")
        if not scene:
            raise ValueError("LLM returned empty cover image prompt")
        preview = scene if len(scene) <= 160 else scene[:157] + "…"
        log(f"LLM: scene parsed ({len(scene)} chars): {preview}")
    full = finalize_cover_prompt(scene)
    if not full:
        raise ValueError("failed to build cover image prompt")
    log(f"LLM: final prompt {len(full)} chars (style prefix applied)")
    return full
