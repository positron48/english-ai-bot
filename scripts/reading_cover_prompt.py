"""LLM prompt builder for reading cover images."""
from __future__ import annotations

import json
import os
import pathlib
import re
import sys

_REPO_SCRIPTS = pathlib.Path(__file__).resolve().parent
if str(_REPO_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_REPO_SCRIPTS))

import reading_llm_client as rlc  # noqa: E402

MAX_SOURCE_CHARS = 1500


def plain_text_from_doc(doc: dict) -> str:
    title = str(doc.get("title") or "").strip()
    rp = doc.get("reading_passage") or {}
    if not title:
        title = str(rp.get("title") or "").strip()
    parts: list[str] = []
    if title:
        parts.append(title)
    for seg in rp.get("segments") or []:
        if isinstance(seg, dict):
            text = str(seg.get("text") or "").strip()
            if text:
                parts.append(text)
    joined = "\n".join(parts)
    if len(joined) > MAX_SOURCE_CHARS:
        joined = joined[:MAX_SOURCE_CHARS].rstrip() + "…"
    return joined


def build_cover_prompt_messages(source_text: str, target_lang: str) -> list[dict]:
    lang = (target_lang or "en").strip().lower()
    return [
        {
            "role": "user",
            "content": (
                "You write short English prompts for Stable Diffusion illustration covers.\n"
                "Given a reading text (title + dialogue/narrative), output ONE image prompt in English only.\n"
                "Style: flat modern illustration, warm colors, scene from the story, no text, no watermark, no logos.\n"
                f"The source text is in {lang}; describe the visual scene, not the language lesson.\n"
                "Reply with JSON only: {\"image_prompt\": \"...\"}\n\n"
                f"SOURCE TEXT:\n{source_text}"
            ),
        }
    ]


def parse_image_prompt(raw: str) -> str:
    text = (raw or "").strip()
    if not text:
        return ""
    fence = re.search(r"```(?:json)?\s*([\s\S]*?)```", text, flags=re.I)
    if fence:
        text = fence.group(1).strip()
    try:
        data = json.loads(text)
        if isinstance(data, dict):
            for key in ("image_prompt", "prompt", "cover_image_prompt"):
                val = str(data.get(key) or "").strip()
                if val:
                    return val
    except json.JSONDecodeError:
        pass
    return text.splitlines()[0].strip() if text else ""


def build_cover_prompt(course_root: pathlib.Path, doc: dict) -> str:
    source = plain_text_from_doc(doc)
    if not source.strip():
        raise ValueError("reading text has no content for cover prompt")
    target_lang = str(doc.get("target_language") or "en").strip().lower()
    if os.getenv("READING_COVER_PROMPT_MOCK", "").strip():
        return "flat illustration of two people talking in a cozy cafe, warm colors, no text"
    prompt = build_cover_prompt_messages(source, target_lang)[0]["content"]
    raw = rlc.chat_completion(prompt, course_root, temperature=0.4)
    image_prompt = parse_image_prompt(raw)
    if not image_prompt:
        raise ValueError("LLM returned empty cover image prompt")
    return image_prompt
