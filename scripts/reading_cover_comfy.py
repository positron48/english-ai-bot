"""ComfyUI txt2img client for reading covers."""
from __future__ import annotations

import copy
import json
import os
import pathlib
import random
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid


def comfyui_url() -> str:
    return os.getenv("COMFYUI_URL", "http://127.0.0.1:8188").rstrip("/")


def workflow_path() -> pathlib.Path:
    custom = os.getenv("COMFYUI_WORKFLOW", "").strip()
    if custom:
        return pathlib.Path(custom).expanduser().resolve()
    return pathlib.Path(__file__).resolve().parent / "comfyui" / "reading-cover.workflow.json"


def load_workflow() -> dict:
    data = json.loads(workflow_path().read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError("invalid ComfyUI workflow JSON")
    return data


def inject_prompt(workflow: dict, prompt: str, checkpoint: str) -> dict:
    wf = copy.deepcopy(workflow)
    ckpt = checkpoint or os.getenv("COMFYUI_CHECKPOINT", "").strip() or "v1-5-pruned-emaonly.safetensors"
    raw = json.dumps(wf)
    raw = raw.replace("__PROMPT__", json.dumps(prompt)[1:-1])
    raw = raw.replace("__CHECKPOINT__", json.dumps(ckpt)[1:-1])
    out = json.loads(raw)
    # Randomize KSampler seed when present.
    for node in out.values():
        if isinstance(node, dict) and node.get("class_type") == "KSampler":
            inputs = node.setdefault("inputs", {})
            inputs["seed"] = random.randint(0, 2**31 - 1)
    return out


def _http_json(url: str, payload: dict | None = None, timeout: int = 120) -> dict:
    data = None
    headers = {"Content-Type": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method="POST" if payload else "GET")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode("utf-8"))


def queue_prompt(workflow: dict, client_id: str | None = None) -> str:
    client_id = client_id or str(uuid.uuid4())
    base = comfyui_url()
    body = {"prompt": workflow, "client_id": client_id}
    res = _http_json(f"{base}/prompt", body, timeout=60)
    prompt_id = str(res.get("prompt_id") or "").strip()
    if not prompt_id:
        raise RuntimeError(f"ComfyUI /prompt failed: {res}")
    return prompt_id


def wait_for_image(prompt_id: str, poll_sec: float = 1.0, timeout_sec: int = 600) -> dict:
    base = comfyui_url()
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        try:
            hist = _http_json(f"{base}/history/{prompt_id}", timeout=30)
        except urllib.error.URLError:
            time.sleep(poll_sec)
            continue
        entry = hist.get(prompt_id) if isinstance(hist, dict) else None
        if not entry:
            time.sleep(poll_sec)
            continue
        outputs = entry.get("outputs") or {}
        for node_out in outputs.values():
            images = node_out.get("images") or []
            if images:
                return images[0]
        status = entry.get("status") or {}
        if status.get("status_str") == "error":
            raise RuntimeError(f"ComfyUI generation error: {status}")
        time.sleep(poll_sec)
    raise TimeoutError(f"ComfyUI timed out waiting for prompt {prompt_id}")


def download_image(image_meta: dict, output_path: pathlib.Path) -> None:
    base = comfyui_url()
    params = urllib.parse.urlencode(
        {
            "filename": image_meta.get("filename", ""),
            "subfolder": image_meta.get("subfolder", ""),
            "type": image_meta.get("type", "output"),
        }
    )
    url = f"{base}/view?{params}"
    req = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = resp.read()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_bytes(data)


def generate_png(prompt: str, output_path: pathlib.Path, checkpoint: str = "") -> None:
    wf = inject_prompt(load_workflow(), prompt, checkpoint)
    prompt_id = queue_prompt(wf)
    image_meta = wait_for_image(prompt_id)
    download_image(image_meta, output_path)
