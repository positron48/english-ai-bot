"""LLM prompt builder for reading cover images."""
from __future__ import annotations

import hashlib
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

# Read the whole passage: appearance and scene constraints can occur after
# the opening lines (for example the mother's green coat in a family scene).
_SENTENCE_SPLIT_RE = re.compile(r"(?<=[.!?…])\s+")

DEFAULT_COVER_STYLE_PREFIX = "casual watercolor illustration, "
NO_TEXT_DIRECTIVE = (
    "no text, no words, no letters, no numbers, no signs, no labels, "
    "no captions, no logos, no watermark"
)

MOCK_SCENE_PROMPT = "two people at a small cafe table on a quiet street"

_IMAGE_PROMPT_KEYS = ("image_prompt", "prompt", "cover_image_prompt")

_COVER_FEWSHOT_EXAMPLE = (
    "Example:\n"
    "Title: Tapas\n"
    'Output: {"image_prompt": "Two friends sharing tapas at a Spanish bar, no text, no signs, no labels."}'
)


def cover_style_prefix() -> str:
    v = os.getenv("READING_COVER_STYLE_PREFIX")
    return DEFAULT_COVER_STYLE_PREFIX if v is None else v


def cover_style_suffix() -> str:
    return os.getenv("READING_COVER_STYLE_SUFFIX", "").strip()


def normalize_scene_text(scene: str) -> str:
    return " ".join((scene or "").split()).strip().rstrip(".,;")


def ensure_no_text_directive(prompt: str) -> str:
    text = normalize_scene_text(prompt)
    if not text:
        return ""
    if "no text" in text.lower():
        return text
    return text + ", " + NO_TEXT_DIRECTIVE


def finalize_cover_prompt(scene_prompt: str, *, people: str = "", source_context: str = "") -> str:
    """Keep exact source facts through to the image model, not only the scene LLM."""
    scene = normalize_scene_text(scene_prompt)
    if not scene:
        return ""
    prefix = cover_style_prefix()
    suffix = cover_style_suffix()
    if prefix and not prefix.endswith((" ", ",")):
        prefix += " "
    parts = [prefix + scene]
    if people:
        parts.append(people)
    if source_context:
        parts.append(
            "Passage context for factual accuracy only; never render its words as writing. "
            "Preserve the passage's explicit people, ages, counts, appearance, clothing, positions, "
            "objects, setting and event order. These source facts override optional casting suggestions. "
            "Choose one faithful scene rather than depicting every event together. "
            "Source passage:\n" + source_context
        )
    if suffix:
        parts.append(suffix.lstrip(", "))
    parts.append(NO_TEXT_DIRECTIVE)
    return ". ".join(part.rstrip("., ") for part in parts if part)


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


# Gender-neutral traits; the passage determines each person's gender and age.
# Most optional appearance palettes are European, with an occasional Afro
# character. Never replace explicit traits in the passage with these defaults.
_EUROPEAN_APPEARANCES = (
    "fair freckled skin, straight ash-blond hair, oval face, blue cardigan",
    "light olive skin, short wavy chestnut hair, square face, natural average build, rust sweater",
    "fair skin, cropped copper-red hair, round face, natural average build, forest-green jacket",
    "lightly tanned skin, straight dark hair, long face, natural proportions, burgundy shirt",
    "rosy fair skin, straight dark-blond hair, heart-shaped face, round glasses, cream knitwear",
    "light olive skin, dark-brown curls, broad face, natural average build, ochre sweater",
    "fair skin, short brown side-parted hair, angular face, navy jacket",
    "medium olive skin, long wavy black hair, rounded face, natural proportions, teal overshirt",
)
_AFRO_APPEARANCE = "dark-brown skin, short tightly coiled black hair, rounded face, rectangular glasses, plum shirt"


def appearance_guidance(seed: str) -> str:
    """Stable per text, varied between texts; only unspecified human traits."""
    digest = hashlib.sha256(seed.encode("utf-8")).digest()
    first = digest[1] % len(_EUROPEAN_APPEARANCES)
    second = (first + 1 + digest[2] % (len(_EUROPEAN_APPEARANCES) - 1)) % len(_EUROPEAN_APPEARANCES)
    primary = _AFRO_APPEARANCE if digest[0] % 10 == 0 else _EUROPEAN_APPEARANCES[first]
    ages = ("young adult around 20–30", "young adult around 25–35", "adult around 30–40")
    casting = (
        "primary person with Afro appearance and second person with European appearance"
        if digest[0] % 10 == 0
        else "people with European appearance"
    )
    return (
        "Human appearance: do not add people absent from the scene. Preserve ALL passage-specified "
        "gender, age, complexion, hair and clothing; those facts override every suggestion below. "
        "Only for unspecified traits, use " + casting + ". Describe each visible person individually, avoiding the same "
        "generic brunette couple across covers. Suggested primary person: " + primary + ". "
        "Suggested second person, only if present: " + _EUROPEAN_APPEARANCES[second] + ". "
        "For adults with unspecified age, suggested ages are " + ages[digest[3] % 3] + " and "
        + ages[digest[4] % 3] + "; never age up a child or contradict the story. "
        "Unless the story requires otherwise, prioritize ordinary young adults with natural "
        "proportions and everyday contemporary European clothing suitable for the scene. "
        "Vary natural hair colors, hairstyles, facial features and small clothing details, "
        "rather than advanced age, exaggerated muscular physiques, elaborate costumes or stylized exotic looks. "
        "Keep people natural, without caricatures or stereotypes."
    )


def plain_text_from_doc(doc: dict, *, for_cover: bool = False, compact: bool = False) -> str:
    title = str(doc.get("title") or "").strip()
    rp = doc.get("reading_passage") or {}
    if not title:
        title = str(rp.get("title") or "").strip()
    parts: list[str] = [f"TITLE: {title}"] if title else []
    for seg in rp.get("segments") or []:
        if not isinstance(seg, dict):
            continue
        text = str(seg.get("text") or "").strip()
        if not text:
            continue
        ru = str(seg.get("text_translation_ru") or "").strip()
        parts.append(f"{text} — {ru}" if ru and not (for_cover or compact) else text)
    return "\n".join(parts)


def build_cover_prompt_messages(
    source_text: str, target_lang: str, title: str = "", *, text_id: str = ""
) -> list[dict]:
    """Build one faithful scene, with explicit appearance for existing people."""
    lang = (target_lang or "en").strip().lower()
    topic = (title or "").strip() or "reading text"
    people = appearance_guidance(text_id or f"{lang}|{topic}|{source_text}")
    return [
        {
            "role": "user",
            "content": (
                f"{_COVER_FEWSHOT_EXAMPLE}\n\n"
                f"Title: {topic}\n"
                f"Passage language: {lang}\n"
                f"Passage:\n{source_text}\n"
                f"{people}\n"
                "Rules: Return a visual scene only. Do not include any visible text, words, letters, "
                "numbers, signs, labels, captions, logos or readable markings in the image. "
                "When required by the passage, books, documents, computer screens, maps and forms "
                "may appear as props, but their contents must be abstract or unlettered. "
                "Depict the main action and setting faithfully, not an unrelated generic cafe scene.\n"
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
            label = "cover scene" if attempt == 0 else "cover scene (retry)"
            log(
                f"LLM: text_id={text_id or '?'} title={title or '?'} "
                f"source={len(source)} chars lang={target_lang} compact={compact}"
            )
            log("LLM: requesting scene description from local llama.cpp…")
            try:
                messages = build_cover_prompt_messages(source, target_lang, title, text_id=text_id)
                if compact:
                    messages[0]["content"] += (
                        '\nReturn only valid JSON {"image_prompt": "visual scene"}; no reasoning or markdown.'
                    )
                scene = _request_cover_scene(
                    messages,
                    course_root,
                    label=label,
                )
                break
            except ValueError:
                if attempt == 1:
                    raise
                log("LLM: no JSON in response — retry without dropping passage facts")
        if not scene:
            raise ValueError("LLM returned empty cover image prompt")
        preview = scene if len(scene) <= 160 else scene[:157] + "…"
        log(f"LLM: scene parsed ({len(scene)} chars): {preview}")
    source = plain_text_from_doc(doc, for_cover=True)
    people = appearance_guidance(text_id or f"{target_lang}|{title}|{source}")
    full = finalize_cover_prompt(scene, people=people, source_context=source)
    if not full:
        raise ValueError("failed to build cover image prompt")
    log(f"LLM: final prompt {len(full)} chars (style prefix applied)")
    return full
