"""Shared stdout logging for reading cover pipeline."""
from __future__ import annotations

import sys
import time


def log_stage(stage_id: str, label: str = "") -> None:
    text = label.strip() if label else stage_id
    ts = time.strftime("%H:%M:%S")
    print(f"[reading-cover {ts}] stage {stage_id}|{text}", flush=True)


def log_comfy_prompt(prompt: str) -> None:
    text = (prompt or "").strip()
    log(f"▶ ComfyUI prompt ({len(text)} chars)")
    ts = time.strftime("%H:%M:%S")
    print(f"[reading-cover {ts}] COMFYUI_PROMPT|{text}", flush=True)


def log(msg: str) -> None:
    ts = time.strftime("%H:%M:%S")
    print(f"[reading-cover {ts}] {msg}", flush=True)
