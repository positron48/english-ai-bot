#!/usr/bin/env python3
"""Apply word-training fixes on prod ES via internal API (no grammar deploy)."""
from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request


def http(method: str, url: str, token: str, body: dict | None = None) -> dict:
    data = None
    headers = {"X-Service-Token": token, "Content-Type": "application/json"}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            raw = resp.read().decode("utf-8")
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        raise RuntimeError(f"{method} {url} -> {e.code}: {e.read().decode()}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"Network {url}: {e}") from e


def main() -> int:
    base = os.environ.get("COMPLAINTS_SERVICE_URL_ES", "").rstrip("/")
    token = os.environ.get("COMPLAINTS_SERVICE_TOKEN_ES") or os.environ.get(
        "COMPLAINTS_SERVICE_TOKEN_EN", ""
    )
    if not base or not token:
        print("Set COMPLAINTS_SERVICE_URL_ES and token in secrets/complaints-prod.env", file=sys.stderr)
        return 1

    # aunque: empty distractors in RU array
    http(
        "PUT",
        f"{base}/api/internal/training/card/8641",
        token,
        {
            "distractors_ru": json.dumps(
                ["потому", "однако", "если", "тогда"], ensure_ascii=False
            )
        },
    )
    print("✓ fixed training card 8641 (aunque distractors_ru)")

    tts_words = ["niño", "pueblo", "producto", "perdonar", "reír"]
    for word in tts_words:
        for attempt in range(3):
            try:
                out = http("POST", f"{base}/api/internal/tts/regenerate", token, {"word": word})
                print(f"✓ tts regenerate {word}: state={out.get('state', out)}")
                break
            except RuntimeError as e:
                if attempt == 2:
                    raise
                print(f"  retry {word}: {e}")
                import time

                time.sleep(2)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
