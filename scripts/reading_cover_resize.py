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
    hero = (_env_int("READING_COVER_HERO_W", 1200), _env_int("READING_COVER_HERO_H", 480))
    return thumb, hero


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
    for out_path, size in ((thumb_path, thumb_sz), (hero_path, hero_sz)):
        cropped = _cover_crop(img, size)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        cropped.save(out_path, format="WEBP", quality=quality, method=6)
