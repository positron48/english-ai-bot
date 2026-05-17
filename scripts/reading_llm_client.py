"""LLM client for reading text generation — llama.cpp first (training-pack parity)."""
from __future__ import annotations

import json
import os
import pathlib
import re
import subprocess
import time
import urllib.error
import urllib.request


def parse_duration_seconds(raw: str, default: int = 0) -> int:
    s = (raw or "").strip()
    if not s:
        return default
    if s.isdigit():
        return int(s)
    m = re.fullmatch(r"(\d+(?:\.\d+)?)(s|m|h)", s, flags=re.I)
    if m:
        val = float(m.group(1))
        mult = {"s": 1, "m": 60, "h": 3600}[m.group(2).lower()]
        return int(val * mult)
    try:
        return int(float(s))
    except ValueError:
        return default


def is_local_ollama_url(ai_url: str) -> bool:
    u = (ai_url or "").strip().lower()
    return ":11434" in u or "ollama" in u


def is_local_llamacpp_url(ai_url: str) -> bool:
    u = (ai_url or "").strip().lower()
    if is_local_ollama_url(u):
        return False
    if any(p in u for p in (":8080", ":8090", "llama")):
        return True
    return u.startswith("http://127.0.0.1") or u.startswith("http://localhost")


def load_training_pack_defaults(course_root: pathlib.Path) -> dict:
    cfg = course_root / "config" / "training-pack.json"
    if not cfg.exists():
        return {}
    try:
        data = json.loads(cfg.read_text(encoding="utf-8"))
        defaults = data.get("defaults")
        return defaults if isinstance(defaults, dict) else {}
    except Exception:
        return {}


def _normalize_base_url(raw: str) -> str:
    base = (raw or "").strip().rstrip("/")
    if base.endswith("/v1"):
        base = base[:-3].rstrip("/")
    return base


def resolve_llm_base_url(course_root: pathlib.Path) -> str:
    for key in ("LLAMACPP_URL", "LLM_BASE_URL"):
        v = os.getenv(key, "").strip()
        if v:
            return _normalize_base_url(v)
    defaults = load_training_pack_defaults(course_root)
    v = (defaults.get("llm_base_url") or "").strip()
    if v:
        return _normalize_base_url(v)
    ai = os.getenv("AI_URL", "").strip()
    if ai and not is_local_ollama_url(ai):
        return _normalize_base_url(ai)
    return "http://127.0.0.1:8090"


def resolve_llm_model(course_root: pathlib.Path) -> str:
    for key in ("READING_AI_MODEL", "LLAMACPP_MODEL", "TRAINING_PACK_MODEL"):
        v = os.getenv(key, "").strip()
        if v:
            return v
    defaults = load_training_pack_defaults(course_root)
    v = (defaults.get("llm_model") or "").strip()
    if v:
        return v
    return os.getenv("AI_MODEL", "").strip() or "gpt-4o-mini"


def resolve_llm_api_key() -> str:
    return (
        os.getenv("LOCAL_LLM_API_KEY", "").strip()
        or os.getenv("OPENAI_API_KEY", "").strip()
        or os.getenv("AI_API_KEY", "").strip()
        or "local"
    )


def openai_chat_completions_url(base_url: str) -> str:
    base = base_url.strip().rstrip("/")
    if base.endswith("/chat/completions"):
        return base
    if re.search(r"/v\d+$", base):
        return base + "/chat/completions"
    return base + "/v1/chat/completions"


def llm_timeout_seconds(base_url: str) -> int:
    explicit = parse_duration_seconds(os.getenv("READING_LLM_TIMEOUT", ""), 0)
    if explicit:
        return explicit
    timeout = parse_duration_seconds(os.getenv("AI_TIMEOUT", ""), 0)
    if not timeout:
        timeout = parse_duration_seconds(os.getenv("AI_REQUEST_TIMEOUT", ""), 0)
    if not timeout:
        timeout = 600 if is_local_llamacpp_url(base_url) or is_local_ollama_url(base_url) else 120
    if is_local_llamacpp_url(base_url) or is_local_ollama_url(base_url):
        timeout = max(timeout, 600)
    return timeout


_CTX_PROBE_CACHE: dict[str, int] = {}


def _probe_llama_server_ctx(base_url: str) -> int | None:
    """Read n_ctx from GET /props (falls back if server is down or old build)."""
    base = (base_url or "").strip().rstrip("/")
    if not base or base in _CTX_PROBE_CACHE:
        return _CTX_PROBE_CACHE.get(base)
    for path in ("/props",):
        try:
            req = urllib.request.Request(base + path, method="GET")
            with urllib.request.urlopen(req, timeout=5) as resp:
                data = json.loads(resp.read().decode("utf-8"))
            gen = data.get("default_generation_settings") or {}
            n_ctx = gen.get("n_ctx") or data.get("n_ctx")
            if n_ctx is not None:
                n = max(2048, int(n_ctx))
                _CTX_PROBE_CACHE[base] = n
                return n
        except (OSError, urllib.error.URLError, urllib.error.HTTPError, ValueError, TypeError):
            continue
    return None


def _llama_context_tokens(base_url: str = "") -> int:
    """Prefer live /props n_ctx; else READING_CTX_TOKENS / LLAMACPP_CTX."""
    if base_url and is_local_llamacpp_url(base_url):
        probed = _probe_llama_server_ctx(base_url)
        if probed:
            return probed
    for key in ("READING_CTX_TOKENS", "LLAMACPP_CTX", "LLAMA_CTX"):
        raw = os.getenv(key, "").strip()
        if raw:
            try:
                return max(2048, int(raw))
            except ValueError:
                pass
    return 8192


def _estimate_prompt_tokens(text: str) -> int:
    return max(1, len(text) // 4)


def _max_output_tokens(level: str | None = None) -> int:
    """Per-level caps; READING_MAX_TOKENS is an optional ceiling (not a floor)."""
    by_level = {"A0": 1024, "A1": 1152, "A2": 1280, "B1": 1536, "B2": 1792, "C1": 1792}
    cap = by_level.get((level or "").upper(), 1024)
    raw = os.getenv("READING_MAX_TOKENS", "").strip()
    if raw:
        try:
            cap = min(cap, int(raw))
        except ValueError:
            pass
    return max(384, min(cap, 2048))


def _local_llama_omit_max_tokens(base_url: str) -> bool:
    """Training-pack parity: local llama.cpp without READING_MAX_TOKENS → server default (no cap)."""
    return is_local_llamacpp_url(base_url) and not os.getenv("READING_MAX_TOKENS", "").strip()


def _resolve_max_tokens(prompt: str, level: str | None, base_url: str) -> int | None:
    """Return max_tokens for the request, or None to omit (llama.cpp default, as training-pack)."""
    if _local_llama_omit_max_tokens(base_url):
        return None
    return _cap_output_for_context(prompt, level, base_url)


def _cap_output_for_context(prompt: str, level: str | None, base_url: str) -> int:
    """Ensure prompt_tokens + max_tokens fits in ctx — prevents llama.cpp HTTP 500 Compute error."""
    want = _max_output_tokens(level)
    if not is_local_llamacpp_url(base_url):
        return want
    ctx = _llama_context_tokens(base_url)
    reserve = int(os.getenv("READING_CTX_RESERVE", "512").strip() or "512")
    prompt_tok = _estimate_prompt_tokens(prompt)
    available = ctx - prompt_tok - reserve
    if available < 256:
        raise RuntimeError(
            f"reading prompt слишком длинный (~{prompt_tok} токенов при ctx={ctx}, reserve={reserve}). "
            "Сократите каталог в промпте, уменьшите COUNT или поднимите -c на llama-server "
            "(и READING_CTX_TOKENS в .env)."
        )
    capped = min(want, available)
    if capped < want:
        print(
            f"[reading-llm] max_tokens {want} → {capped} (ctx={ctx}, prompt≈{prompt_tok} tok, reserve={reserve})",
            flush=True,
        )
    return capped


def _local_cooldown_seconds() -> float:
    raw = os.getenv("READING_LLM_COOLDOWN_SEC", "3").strip()
    try:
        return max(0.0, float(raw))
    except ValueError:
        return 3.0


def _reading_disable_stream() -> bool:
    return os.getenv("READING_DISABLE_STREAM", "").strip().lower() in ("1", "true", "yes")


def _is_qwen_thinking_model(model: str) -> bool:
    return "qwen3" in (model or "").lower()


def _prepare_reading_prompt(prompt: str, model: str, base_url: str) -> str:
    if not is_local_llamacpp_url(base_url) or not _is_qwen_thinking_model(model):
        return prompt
    body = prompt
    if not body.lstrip().startswith("/no_think"):
        body = "/no_think\n" + body
    if "First non-whitespace character must be {" not in body:
        body += (
            "\n\nOutput ONLY one JSON object. "
            "First non-whitespace character must be {. "
            "No English planning, no markdown fences, no XML think blocks."
        )
    return body


def _text_from_parts(part: dict) -> str:
    if not isinstance(part, dict):
        return ""
    chunks: list[str] = []
    for key in ("content", "reasoning_content", "reasoning"):
        v = part.get(key)
        if v:
            chunks.append(str(v))
    return "".join(chunks)


_THINK_BLOCK_RE = re.compile(r"<think(?:ing)?>.*?</think(?:ing)?>", re.I | re.S)


def strip_qwen_thinking(text: str) -> str:
    """Remove Qwen3 thinking blocks; leave JSON/prose outside tags."""
    s = str(text or "")
    s = _THINK_BLOCK_RE.sub("", s)
    s = re.sub(r"</?think(?:ing)?>", "", s, flags=re.I)
    return s.strip()


def extract_json_object(raw: str) -> str:
    """Pull one JSON object from model output (fenced or brace-balanced)."""
    cleaned = strip_qwen_thinking(raw)
    if not cleaned:
        return ""
    fenced = re.search(r"```(?:json)?\s*(\{.*\})\s*```", cleaned, flags=re.S)
    if fenced:
        return fenced.group(1).strip()
    start = cleaned.find("{")
    if start < 0:
        return ""
    try:
        _, end = json.JSONDecoder().raw_decode(cleaned, start)
        return cleaned[start:end]
    except json.JSONDecodeError:
        return cleaned[start:].strip()


def parse_reading_json_response(raw: str) -> dict:
    blob = extract_json_object(raw)
    if not blob:
        preview = strip_qwen_thinking(raw or "")[:600].replace("\n", " ")
        raise ValueError(
            "LLM не вернул JSON (пустой ответ или только thinking без JSON). "
            f"Длина сырого ответа={len(raw or '')}. Начало: {preview!r}. "
            "Для qwen3: /no_think, enable_thinking=false; при обрезке — поднять max_tokens для уровня."
        )
    try:
        data = json.loads(blob)
    except json.JSONDecodeError as e:
        raise ValueError(
            f"Невалидный JSON от LLM ({e.msg}), фрагмент: {blob[:500]!r}"
        ) from e
    if not isinstance(data, dict):
        raise ValueError(f"Ожидался JSON-объект, получен {type(data).__name__}")
    return data


def _message_text_variants(message: dict) -> list[tuple[str, str]]:
    variants: list[tuple[str, str]] = []
    content = message.get("content")
    if content is not None:
        cleaned = strip_qwen_thinking(str(content))
        if cleaned:
            variants.append(("content", cleaned))
    for key in ("reasoning_content", "reasoning"):
        v = message.get(key)
        if not v:
            continue
        cleaned = strip_qwen_thinking(str(v))
        if cleaned and all(cleaned != text for _, text in variants):
            variants.append((key, cleaned))
    return variants


def _pick_response_text(message: dict, model: str = "") -> tuple[str, str]:
    """Parseable JSON may be in content; qwen3 planning often lands in reasoning — ignore it."""
    if _is_qwen_thinking_model(model):
        content = strip_qwen_thinking(str(message.get("content") or ""))
        if content:
            blob = extract_json_object(content)
            if blob:
                try:
                    json.loads(blob)
                    return "content", content
                except json.JSONDecodeError:
                    pass
        return "", ""

    variants = _message_text_variants(message)
    if not variants:
        return "", ""
    order = sorted(variants, key=lambda kv: (0 if kv[0] == "content" else 1, kv[0]))
    candidates: list[tuple[str, str]] = list(order)
    merged = "\n".join(text for _, text in order)
    if merged and all(merged != text for _, text in order):
        candidates.append(("merged", merged))
    for key, text in candidates:
        blob = extract_json_object(text)
        if not blob:
            continue
        try:
            json.loads(blob)
            return key, text
        except json.JSONDecodeError:
            continue
    return "", ""


def _reading_use_stream(base_url: str, model: str) -> bool:
    if _reading_disable_stream():
        return False
    if os.getenv("READING_ENABLE_STREAM", "").strip().lower() in ("1", "true", "yes"):
        return True
    # Training-pack parity: always stream on local llama (incl. qwen3).
    return is_local_llamacpp_url(base_url)


def _host_port_from_base_url(http_base: str) -> tuple[str, int]:
    from urllib.parse import urlparse

    raw = (http_base or "").strip()
    if not raw.startswith(("http://", "https://")):
        raw = "http://" + raw
    parsed = urlparse(raw)
    host = parsed.hostname or "127.0.0.1"
    if parsed.port is not None:
        port = parsed.port
    else:
        port = 443 if parsed.scheme == "https" else 80
    return host, port


def _tcp_port_open(host: str, port: int, timeout: float = 1.0) -> bool:
    import socket

    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except OSError:
        return False


def _llamacpp_probe_ready(http_base: str) -> bool:
    if not http_base:
        return False
    base = http_base.rstrip("/")
    for path in ("/v1/models", "/health"):
        try:
            req = urllib.request.Request(base + path, method="GET")
            with urllib.request.urlopen(req, timeout=3) as resp:
                if resp.status == 200:
                    return True
        except (OSError, urllib.error.URLError, urllib.error.HTTPError):
            continue
    return False


def resolve_llamacpp_start_cmd() -> str:
    return (
        os.getenv("LLAMACPP_START_CMD_READING", "").strip()
        or os.getenv("LLAMACPP_START_CMD", "").strip()
        or os.getenv("LLAMACPP_START_CMD_VERB", "").strip()
    )


def _ensure_llama_disabled() -> bool:
    mode = os.getenv("READING_ENSURE_LLAMA", os.getenv("VERB_FORMS_ENSURE_LLAMA", "auto")).strip().lower()
    return mode in ("0", "false", "no", "off")


_LLAMA_ENSURED_BASES: set[str] = set()


def ensure_llamacpp_server(course_root: pathlib.Path, label: str = "reading-llm") -> bool:
    http_base = resolve_llm_base_url(course_root)
    if http_base in _LLAMA_ENSURED_BASES:
        return True

    if _ensure_llama_disabled():
        ok = _llamacpp_probe_ready(http_base)
        if ok:
            _LLAMA_ENSURED_BASES.add(http_base)
        return ok

    start_cmd = resolve_llamacpp_start_cmd()
    wait_raw = os.getenv("LLAMACPP_START_MAX_WAIT_SEC", "").strip()
    try:
        max_wait = int(wait_raw) if wait_raw else 120
    except ValueError:
        max_wait = 120
    max_wait = max(5, min(max_wait, 3600))
    busy_wait_raw = os.getenv("LLAMACPP_PORT_BUSY_GRACE_SEC", "").strip()
    try:
        busy_grace = int(busy_wait_raw) if busy_wait_raw else 20
    except ValueError:
        busy_grace = 20
    busy_grace = max(1, min(busy_grace, max_wait))

    if _llamacpp_probe_ready(http_base):
        print(f"[{label}] llama.cpp ready at {http_base}", flush=True)
        _LLAMA_ENSURED_BASES.add(http_base)
        return True

    host, port = _host_port_from_base_url(http_base)
    if _tcp_port_open(host, port):
        print(
            f"[{label}] port {port} already in use — ждём готовности llama.cpp",
            flush=True,
        )
        for i in range(busy_grace):
            if _llamacpp_probe_ready(http_base):
                print(f"[{label}] llama.cpp ready after {i + 1}s", flush=True)
                _LLAMA_ENSURED_BASES.add(http_base)
                return True
            time.sleep(1)
        if start_cmd:
            print(
                f"[{label}] порт {port} занят, но API не отвечает {busy_grace}s — запускаем START_CMD",
                flush=True,
            )
            subprocess.run(["bash", "-lc", start_cmd], check=False)
            for i in range(max_wait):
                if _llamacpp_probe_ready(http_base):
                    print(f"[{label}] llama.cpp ready after START_CMD ({i + 1}s)", flush=True)
                    _LLAMA_ENSURED_BASES.add(http_base)
                    return True
                time.sleep(1)
        print(
            f"[{label}] порт {port} занят, но API не отвечает за {max_wait}s — "
            "проверьте процесс на порту и LLAMACPP_START_CMD",
            flush=True,
        )
        return False

    if not start_cmd:
        print(
            f"[{label}] {http_base} not reachable; set LLAMACPP_START_CMD_VERB in ../../.env.es",
            flush=True,
        )
        return False

    print(f"[{label}] llama.cpp not running; autostart (START_CMD)...", flush=True)
    subprocess.run(["bash", "-lc", start_cmd], check=False)
    for i in range(max_wait):
        if _llamacpp_probe_ready(http_base):
            print(f"[{label}] llama.cpp ready after {i + 1}s", flush=True)
            _LLAMA_ENSURED_BASES.add(http_base)
            return True
        time.sleep(1)

    print(
        f"[{label}] llama.cpp not ready after {max_wait}s — check LLAMACPP_URL={http_base} "
        "matches --port in START_CMD",
        flush=True,
    )
    return False


def ensure_llamacpp_for_reading(course_root: pathlib.Path) -> None:
    ensure_llamacpp_server(course_root, label="reading-llm")


def _stream_chat_completion(chat_url: str, payload: dict, headers: dict, timeout_s: int) -> str:
    body = dict(payload)
    body["stream"] = True
    req = urllib.request.Request(
        chat_url,
        data=json.dumps(body).encode("utf-8"),
        headers=headers,
        method="POST",
    )
    parts: list[str] = []
    with urllib.request.urlopen(req, timeout=timeout_s) as resp:
        for raw_line in resp:
            line = raw_line.decode("utf-8", errors="ignore").strip()
            if not line:
                continue
            if line.startswith("data:"):
                line = line[5:].strip()
            if line == "[DONE]":
                break
            try:
                data = json.loads(line)
            except json.JSONDecodeError:
                continue
            choices = data.get("choices", [])
            if not choices:
                continue
            delta = choices[0].get("delta", {})
            # Training-pack parity: only `content` — qwen3 reasoning stream is English planning, not JSON.
            chunk = delta.get("content") or ""
            if not chunk and isinstance(choices[0].get("message"), dict):
                chunk = choices[0]["message"].get("content") or ""
            if chunk:
                parts.append(str(chunk))
    return "".join(parts)


def _blocking_chat_completion(chat_url: str, payload: dict, headers: dict, timeout_s: int) -> str:
    req = urllib.request.Request(
        chat_url,
        data=json.dumps(payload).encode("utf-8"),
        headers=headers,
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=timeout_s) as resp:
        body = json.loads(resp.read().decode("utf-8"))
    choice = body["choices"][0]
    message = choice.get("message", {})
    finish = choice.get("finish_reason")
    field, text = _pick_response_text(message, str(payload.get("model", "")))
    if text:
        if field != "content":
            print(f"[reading-llm] using message.{field} ({len(text)} chars)", flush=True)
        if finish == "length":
            print(
                "[reading-llm] finish_reason=length — ответ обрезан; JSON может быть битым",
                flush=True,
            )
        return text
    parts = _message_text_variants(message)
    detail = ", ".join(f"{k}={len(t)}ch" for k, t in parts) or "no fields"
    raise RuntimeError(
        f"LLM did not return parseable JSON (finish_reason={finish!r}; {detail}). "
        "Для qwen3: response_format=json_object (по умолчанию), /no_think, enable_thinking=false; "
        "при length — увеличить max_tokens или сократить промпт."
    )


def _reading_temperature(base_url: str) -> float:
    """Local llama: match training-pack (0.1); remote APIs keep slightly higher default."""
    if is_local_llamacpp_url(base_url) or is_local_ollama_url(base_url):
        return 0.1
    return 0.4


def chat_completion(
    prompt: str,
    course_root: pathlib.Path,
    temperature: float | None = None,
    level: str | None = None,
) -> str:
    base_url = resolve_llm_base_url(course_root)
    if temperature is None:
        temperature = _reading_temperature(base_url)
    if is_local_llamacpp_url(base_url) and not ensure_llamacpp_server(course_root, label="reading-llm"):
        raise RuntimeError(
            f"llama.cpp недоступен на {base_url} — задайте LLAMACPP_START_CMD_VERB в ../../.env.es"
        )

    model = resolve_llm_model(course_root)
    api_key = resolve_llm_api_key()
    chat_url = openai_chat_completions_url(base_url)
    timeout_s = llm_timeout_seconds(base_url)

    user_prompt = _prepare_reading_prompt(prompt, model, base_url)
    max_out = _resolve_max_tokens(user_prompt, level, base_url)

    payload: dict = {
        "model": model,
        "messages": [{"role": "user", "content": user_prompt}],
        "temperature": temperature,
    }
    if max_out is not None:
        payload["max_tokens"] = max_out
    if is_local_ollama_url(base_url):
        payload["think"] = False
    if is_local_llamacpp_url(base_url) and _is_qwen_thinking_model(model):
        payload["chat_template_kwargs"] = {"enable_thinking": False}
        payload["reasoning_format"] = "none"
        rb = os.getenv("READING_REASONING_BUDGET", "0").strip()
        if rb:
            try:
                payload["reasoning_budget"] = int(rb)
            except ValueError:
                pass
    # Training-pack does not set response_format; qwen3 + json_object often leaks planning into content.
    if (
        is_local_llamacpp_url(base_url)
        and os.getenv("READING_ENABLE_JSON_FORMAT", "").strip().lower() in ("1", "true", "yes")
    ):
        payload["response_format"] = {"type": "json_object"}

    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"

    use_stream = _reading_use_stream(base_url, model)
    ctx = _llama_context_tokens(base_url) if is_local_llamacpp_url(base_url) else 0
    ctx_from_props = base_url.rstrip("/") in _CTX_PROBE_CACHE if ctx else False
    max_tok_label = "server-default" if max_out is None else str(max_out)
    print(
        f"[reading-llm] → {chat_url} model={model} stream={use_stream} "
        f"prompt_chars={len(user_prompt)} max_tokens={max_tok_label}"
        + (f" ctx={ctx}" + (" (from /props)" if ctx_from_props else " (from env)") if ctx else ""),
        flush=True,
    )

    try:
        if use_stream:
            content = _stream_chat_completion(chat_url, payload, headers, timeout_s)
        else:
            content = _blocking_chat_completion(chat_url, payload, headers, timeout_s)
    except urllib.error.HTTPError as e:
        detail = ""
        try:
            detail = e.read().decode("utf-8", errors="replace")[:2000]
        except Exception:
            pass
        hint = ""
        if e.code == 500 and "Compute error" in detail:
            hint = (
                f" — ctx={ctx}, prompt≈{_estimate_prompt_tokens(user_prompt)} tok, max_tokens={max_tok_label}. "
                "Обычно: max_tokens слишком велик для 30B или ctx на сервере меньше READING_CTX_TOKENS — "
                "проверьте GET /props (n_ctx) и -c/-n в LLAMACPP_START_CMD_VERB; "
                "для жёсткого лимита задайте READING_MAX_TOKENS."
            )
        raise RuntimeError(f"LLM HTTP {e.code} for {chat_url}. Body: {detail}{hint}") from e

    content = strip_qwen_thinking(content)
    if not content.strip():
        raise RuntimeError("LLM returned empty content")

    if is_local_llamacpp_url(base_url):
        cool = _local_cooldown_seconds()
        if cool > 0:
            time.sleep(cool)

    return content
