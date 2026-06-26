"""Resize reading cover PNG to thumb + hero WebP."""
from __future__ import annotations

import os
import pathlib

from PIL import Image


def _env_int(name: str, default: int) -> int:
    raw = os.getenv(name, "").strip()
    if not raw:
        return default
    try:
        return int(raw)
    except ValueError:
        return default


def cover_sizes() -> tuple[tuple[int, int], tuple[int, int]]:
    thumb = (_env_int("READING_COVER_THUMB_W", 400), _env_int("READING_COVER_THUMB_H", 300))
    hero = (_env_int("READING_COVER_HERO_W", 1024), _env_int("READING_COVER_HERO_H", 768))
    return thumb, hero


def _cover_fit(img: Image.Image, max_size: tuple[int, int]) -> Image.Image:
    """Scale down to fit inside max_size; never crop."""
    max_w, max_h = max_size
    src_w, src_h = img.size
    if src_w <= max_w and src_h <= max_h:
        return img.copy()
    scale = min(max_w / src_w, max_h / src_h)
    return img.resize((max(1, int(src_w * scale)), max(1, int(src_h * scale))), Image.Resampling.LANCZOS)


def _cover_crop(img: Image.Image, target_size: tuple[int, int]) -> Image.Image:
    tw, th = target_size
    src_w, src_h = img.size
    scale = max(tw / src_w, th / src_h)
    resized = img.resize((int(src_w * scale), int(src_h * scale)), Image.Resampling.LANCZOS)
    left = (resized.width - tw) // 2
    top = (resized.height - th) // 2
    return resized.crop((left, top, left + tw, top + th))


def resize_cover_assets(raw_png: pathlib.Path, thumb_path: pathlib.Path, hero_path: pathlib.Path) -> None:
    thumb_sz, hero_sz = cover_sizes()
    quality = _env_int("READING_COVER_WEBP_QUALITY", 85)
    img = Image.open(raw_png).convert("RGB")

    thumb_path.parent.mkdir(parents=True, exist_ok=True)
    _cover_crop(img, thumb_sz).save(thumb_path, format="WEBP", quality=quality, method=6)

    hero_path.parent.mkdir(parents=True, exist_ok=True)
    _cover_fit(img, hero_sz).save(hero_path, format="WEBP", quality=quality, method=6)
