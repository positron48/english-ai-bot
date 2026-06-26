#!/usr/bin/env python3
"""Generate reading cover images: LLM prompt -> ComfyUI -> WebP thumb/hero -> update text JSON."""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import subprocess
import sys
import time

REPO_ROOT = pathlib.Path(__file__).resolve().parents[1]
SCRIPTS = REPO_ROOT / "scripts"
if str(SCRIPTS) not in sys.path:
    sys.path.insert(0, str(SCRIPTS))

from reading_cover_log import log, log_comfy_prompt, log_stage  # noqa: E402
from reading_cover_prompt import build_cover_prompt  # noqa: E402
from reading_cover_resize import resize_cover_assets  # noqa: E402


def load_index(course_root: pathlib.Path, draft_dir: str) -> dict:
    path = course_root / draft_dir / "index.json"
    if not path.exists():
        raise FileNotFoundError(f"missing {path}")
    return json.loads(path.read_text(encoding="utf-8"))


def text_json_path(course_root: pathlib.Path, draft_dir: str, text_id: str, rel: str) -> pathlib.Path:
    if ".." in rel or rel.startswith("/"):
        raise ValueError(f"invalid rel path: {rel!r}")
    return course_root / draft_dir / pathlib.Path(rel)


def assets_dir(course_root: pathlib.Path, text_id: str) -> pathlib.Path:
    return course_root / "assets" / "reading" / text_id


def cover_paths(text_id: str) -> tuple[str, str]:
    base = f"assets/reading/{text_id}"
    return f"{base}/cover_thumb.webp", f"{base}/cover_hero.webp"


def has_cover_files(course_root: pathlib.Path, doc: dict) -> bool:
    thumb = str(doc.get("cover_thumb_rel_path") or "").strip()
    hero = str(doc.get("cover_hero_rel_path") or "").strip()
    if not thumb or not hero:
        return False
    return (course_root / thumb).is_file() and (course_root / hero).is_file()


def update_doc_cover(doc: dict, text_id: str, image_prompt: str) -> None:
    thumb_rel, hero_rel = cover_paths(text_id)
    doc["cover_thumb_rel_path"] = thumb_rel
    doc["cover_hero_rel_path"] = hero_rel
    doc["cover_image_prompt"] = image_prompt


def write_json_atomic(path: pathlib.Path, data: dict) -> None:
    tmp = path.with_suffix(path.suffix + ".tmp")
    tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    tmp.replace(path)


def render_cover_for_doc(
    course_root: pathlib.Path,
    json_path: pathlib.Path,
    image_prompt: str,
    *,
    force: bool = True,
) -> str:
    """ComfyUI + resize + JSON update using a ready-made image prompt (no LLM)."""
    image_prompt = " ".join((image_prompt or "").split()).strip()
    if not image_prompt:
        raise ValueError("image_prompt is required")
    doc = json.loads(json_path.read_text(encoding="utf-8"))
    text_id = str(doc.get("id") or json_path.stem).strip()
    title = str(doc.get("title") or text_id).strip()
    log(f"=== {text_id} — {title} (prompt only, no LLM) ===")
    if not force and has_cover_files(course_root, doc):
        log("skip: cover files already exist (use --force to regenerate)")
        return "skip"
    log_stage("prepare", "Подготовка")
    log("step 1/4: skipped LLM — using provided image prompt")
    out_dir = assets_dir(course_root, text_id)
    out_dir.mkdir(parents=True, exist_ok=True)
    raw_png = out_dir / "cover_raw.png"
    thumb_path = out_dir / "cover_thumb.webp"
    hero_path = out_dir / "cover_hero.webp"

    log_stage("comfyui", "Картинка (ComfyUI)")
    log_comfy_prompt(image_prompt)
    log(f"step 2/4: ComfyUI txt2img -> {raw_png.name}")
    script = SCRIPTS / "generate-reading-cover.sh"
    subprocess.run(
        ["bash", str(script), "--prompt", image_prompt, "--output", str(raw_png)],
        check=True,
        cwd=str(REPO_ROOT),
    )
    log_stage("resize", "WebP thumb + hero")
    log(f"step 3/4: resize WebP thumb + hero")
    resize_cover_assets(raw_png, thumb_path, hero_path)
    log(f"  thumb: {thumb_path} ({thumb_path.stat().st_size if thumb_path.is_file() else 0} bytes)")
    log(f"  hero:  {hero_path} ({hero_path.stat().st_size if hero_path.is_file() else 0} bytes)")
    log_stage("save", "Сохранение")
    log("step 4/4: update text JSON")
    update_doc_cover(doc, text_id, image_prompt)
    write_json_atomic(json_path, doc)
    log_stage("done", "Готово")
    log("done")
    return "ok"


def save_prompt_for_doc(course_root: pathlib.Path, json_path: pathlib.Path, force: bool = False) -> str:
    """LLM cover prompt only — saves cover_image_prompt without ComfyUI."""
    doc = json.loads(json_path.read_text(encoding="utf-8"))
    text_id = str(doc.get("id") or json_path.stem).strip()
    title = str(doc.get("title") or text_id).strip()
    log(f"=== {text_id} — {title} (prompt only) ===")
    if not force and str(doc.get("cover_image_prompt") or "").strip():
        log("skip: cover_image_prompt already set (use --force to regenerate)")
        return "skip"
    log_stage("prepare", "Подготовка")
    log_stage("llm", "Промпт (LLM)")
    log("step 1/2: LLM cover scene prompt (local llama.cpp)")
    image_prompt = build_cover_prompt(course_root, doc)
    log_stage("save", "Сохранение промпта")
    log("step 2/2: update text JSON (cover_image_prompt)")
    doc["cover_image_prompt"] = image_prompt
    write_json_atomic(json_path, doc)
    log_stage("done", "Готово")
    log("done (prompt saved, no image)")
    return "ok"


def generate_for_doc(course_root: pathlib.Path, json_path: pathlib.Path, force: bool = False) -> str:
    doc = json.loads(json_path.read_text(encoding="utf-8"))
    text_id = str(doc.get("id") or json_path.stem).strip()
    title = str(doc.get("title") or text_id).strip()
    log(f"=== {text_id} — {title} ===")
    if not force and has_cover_files(course_root, doc):
        log("skip: cover files already exist (use --force to regenerate)")
        return "skip"
    log_stage("prepare", "Подготовка")
    log_stage("llm", "Промпт (LLM)")
    log("step 1/4: LLM cover scene prompt (local llama.cpp)")
    image_prompt = build_cover_prompt(course_root, doc)
    out_dir = assets_dir(course_root, text_id)
    out_dir.mkdir(parents=True, exist_ok=True)
    raw_png = out_dir / "cover_raw.png"
    thumb_path = out_dir / "cover_thumb.webp"
    hero_path = out_dir / "cover_hero.webp"

    log_stage("comfyui", "Картинка (ComfyUI)")
    log_comfy_prompt(image_prompt)
    log(f"step 2/4: ComfyUI txt2img -> {raw_png.name}")
    script = SCRIPTS / "generate-reading-cover.sh"
    subprocess.run(
        ["bash", str(script), "--prompt", image_prompt, "--output", str(raw_png)],
        check=True,
        cwd=str(REPO_ROOT),
    )
    log_stage("resize", "WebP thumb + hero")
    log(f"step 3/4: resize WebP thumb + hero")
    resize_cover_assets(raw_png, thumb_path, hero_path)
    log(f"  thumb: {thumb_path} ({thumb_path.stat().st_size if thumb_path.is_file() else 0} bytes)")
    log(f"  hero:  {hero_path} ({hero_path.stat().st_size if hero_path.is_file() else 0} bytes)")
    log_stage("save", "Сохранение")
    log("step 4/4: update text JSON")
    update_doc_cover(doc, text_id, image_prompt)
    write_json_atomic(json_path, doc)
    log_stage("done", "Готово")
    log("done")
    return "ok"


def batch(course_root: pathlib.Path, draft_dir: str, force: bool, limit: int) -> int:
    idx = load_index(course_root, draft_dir)
    text_ids = sorted((idx.get("texts") or {}).keys())
    if limit > 0:
        text_ids = text_ids[:limit]
    failed = 0
    for i, text_id in enumerate(text_ids, start=1):
        rel = idx["texts"][text_id]
        json_path = text_json_path(course_root, draft_dir, text_id, rel)
        print(f"reading-cover {i}/{len(text_ids)} {text_id}", flush=True)
        try:
            status = generate_for_doc(course_root, json_path, force=force)
            print(f"  -> {status}", flush=True)
        except Exception as exc:
            failed += 1
            print(f"  !! failed: {exc}", flush=True)
        if i < len(text_ids):
            time.sleep(float(os.getenv("READING_COVER_BATCH_SLEEP_SEC", "2")))
    if failed:
        print(f"reading-covers-batch completed with {failed} failure(s)", flush=True)
        return 1
    print("reading-covers-batch completed", flush=True)
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Generate reading cover images")
    parser.add_argument("--course-root", type=pathlib.Path, default=pathlib.Path(".").resolve())
    parser.add_argument("--draft-dir", default="reading")
    parser.add_argument("--text-id", default="", help="single text id from index")
    parser.add_argument("--image-prompt", default="", help="ready prompt: ComfyUI+resize only, skip LLM")
    parser.add_argument("--prompt-only", action="store_true", help="LLM prompt only, save cover_image_prompt")
    parser.add_argument("--force", action="store_true")
    parser.add_argument("--limit", type=int, default=0)
    args = parser.parse_args(argv)
    course_root = args.course_root.resolve()

    if args.text_id:
        idx = load_index(course_root, args.draft_dir)
        rel = (idx.get("texts") or {}).get(args.text_id)
        if not rel:
            print(f"text_id not in index: {args.text_id}", file=sys.stderr)
            return 1
        json_path = text_json_path(course_root, args.draft_dir, args.text_id, rel)
        prompt = (args.image_prompt or "").strip()
        if args.prompt_only:
            save_prompt_for_doc(course_root, json_path, force=args.force)
        elif prompt:
            render_cover_for_doc(course_root, json_path, prompt, force=args.force)
        else:
            generate_for_doc(course_root, json_path, force=args.force)
        return 0
    return batch(course_root, args.draft_dir, args.force, args.limit)


if __name__ == "__main__":
    raise SystemExit(main())
