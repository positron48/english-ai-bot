#!/usr/bin/env python3
import importlib.util
import tempfile
from pathlib import Path


def load_module(path: Path):
    spec = importlib.util.spec_from_file_location("improve_loop", str(path))
    mod = importlib.util.module_from_spec(spec)
    assert spec and spec.loader
    spec.loader.exec_module(mod)
    return mod


def assert_true(cond: bool, msg: str):
    if not cond:
        raise AssertionError(msg)


def test_collect_affected_chapters(mod):
    changed = {
        "blocks": [
            {"course": "english", "chapter_id": "en.a"},
            {"course": "english", "chapter_id": "en.a"},
            {"course": "spanish", "chapter_id": "es.b"},
            {"course": "spanish", "chapter_id": "es.c"},
        ]
    }
    out = mod.collect_affected_chapters(changed)
    assert_true(out["english"] == ["en.a"], "english chapter dedupe failed")
    assert_true(out["spanish"] == ["es.b", "es.c"], "spanish chapter order/dedupe failed")


def test_chapter_number_map(mod):
    with tempfile.TemporaryDirectory() as td:
        root = Path(td)
        chapters = root / "chapters"
        chapters.mkdir(parents=True, exist_ok=True)
        (chapters / "001.es.one").mkdir()
        (chapters / "010.es.ten").mkdir()
        (chapters / "abc.withoutprefix").mkdir()
        mapping = mod.chapter_number_map(root)
        assert_true(mapping["es.one"] == 1, "prefixed chapter map failed")
        assert_true(mapping["es.ten"] == 2, "sorted chapter map failed")
        assert_true(mapping["abc.withoutprefix"] == 3, "non-prefixed chapter map failed")


def test_strict_iteration_ok(mod):
    ok = mod.strict_iteration_ok(True, {"english": True, "spanish": True}, {"english": True, "spanish": True})
    fail = mod.strict_iteration_ok(True, {"english": True, "spanish": False}, {"english": True, "spanish": True})
    assert_true(ok is True, "strict gate should pass for all-true")
    assert_true(fail is False, "strict gate should fail when any sub-check is false")


def test_loop_outcomes(mod):
    assert_true(mod.next_action(1, 3, True) == "accept", "success-path outcome mismatch")
    assert_true(mod.next_action(2, 3, False) == "retry", "retry-path outcome mismatch")
    assert_true(mod.next_action(3, 3, False) == "stop_failed", "fail-after-3 outcome mismatch")


def main() -> int:
    root = Path(__file__).resolve().parents[3]
    mod = load_module(root / "tools-local" / "complaints-worker" / "improve-prompt-loop.py")
    test_collect_affected_chapters(mod)
    test_chapter_number_map(mod)
    test_strict_iteration_ok(mod)
    test_loop_outcomes(mod)
    print('{"status":"ok","tests":4}')
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

