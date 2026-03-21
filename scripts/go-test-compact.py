#!/usr/bin/env python3
"""
Compact formatter for `go test -json` output (Symfony/PHPUnit-style).

- `.` passed subtest
- `E` failed assertion / test error
- `F` panic or build/package fatal
- `S` skipped subtest

Trailing summary: counts, failure details, PASS or FAIL.
"""

from __future__ import annotations

import json
import re
import sys

# Same noise filter as Makefile `check` (testcontainers / docker chatter)
_SUPPRESS_RES = [
    re.compile(
        r"Container (created|started|ready|stopped|terminated)|Creating container|"
        r"Starting container|Terminating container|Waiting for container|Waiting for Reaper|"
        r"Shell not found|Reaper obtained|🐳|✅ Container|🔔 Container|⏳ Waiting|🔥 Reaper|🚫 Container|"
        r"testcontainers-go -|Resolved Docker|Server Version|API Version|Operating System|Total Memory|"
        r"Testcontainers for Go|Test SessionID|Test ProcessID"
    ),
]


def _suppress(line: str) -> bool:
    return any(r.search(line) for r in _SUPPRESS_RES)


def _is_panic(output: str) -> bool:
    o = output.lower()
    return "panic:" in o or "runtime error:" in o


PROG_WIDTH = 100


def main() -> int:
    buf: dict[tuple[str, str], list[str]] = {}
    pkg_out: dict[str, list[str]] = {}
    # Packages that already had a failed subtest (package-level "fail" is then only a summary).
    pkg_with_test_fail: set[str] = set()

    failures: list[tuple[str, str, str, str]] = []  # pkg, name, kind, body
    # kind: "E" | "F" | "pkg"

    passed = 0
    failed = 0
    skipped = 0
    pkg_fail = 0
    prog_col = 0

    def prog_write(ch: str) -> None:
        nonlocal prog_col
        sys.stdout.write(ch)
        prog_col += 1
        if prog_col >= PROG_WIDTH:
            sys.stdout.write("\n")
            prog_col = 0
        sys.stdout.flush()

    for line in sys.stdin:
        raw = line.rstrip("\n")
        if not raw.strip():
            continue
        try:
            ev = json.loads(raw)
        except json.JSONDecodeError:
            if not _suppress(raw):
                print(raw, file=sys.stderr)
            continue

        action = ev.get("Action")
        pkg = ev.get("Package", "")
        test = ev.get("Test")

        if action == "output":
            out = ev.get("Output", "")
            if test:
                key = (pkg, test)
                buf.setdefault(key, []).append(out)
            else:
                pkg_out.setdefault(pkg, []).append(out)
            continue

        if action == "pass" and test:
            passed += 1
            prog_write(".")
            buf.pop((pkg, test), None)
            continue

        if action == "fail" and test:
            failed += 1
            pkg_with_test_fail.add(pkg)
            body = "".join(buf.pop((pkg, test), []))
            kind = "F" if _is_panic(body) else "E"
            prog_write(kind)
            failures.append((pkg, test, kind, body))
            continue

        if action == "skip" and test:
            skipped += 1
            prog_write("S")
            buf.pop((pkg, test), None)
            continue

        if action == "fail" and not test:
            # Package failed: build/init, or summary after tests (latter is redundant).
            if pkg in pkg_with_test_fail:
                continue
            pkg_fail += 1
            body = "".join(pkg_out.pop(pkg, []))
            prog_write("F")
            failures.append((pkg, test or "(package)", "F", body))
            continue

        if action == "skip" and not test:
            # e.g. [no test files] — skip quietly
            continue

    # Завершить последнюю строку прогресса; при отсутствии символов — как раньше одна пустая строка.
    if prog_col > 0:
        sys.stdout.write("\n")
    elif passed + failed + skipped + pkg_fail == 0:
        sys.stdout.write("\n")

    parts = [f"{passed} passed", f"{failed} failed"]
    if skipped:
        parts.append(f"{skipped} skipped")
    if pkg_fail:
        parts.append(f"{pkg_fail} package errors")
    print(f"Tests: {', '.join(parts)}", file=sys.stdout)

    if failures:
        print(f"\n--- FAILURES ({len(failures)}) ---", file=sys.stdout)
        for i, (p, name, kind, body) in enumerate(failures, 1):
            print(f"\n{i}. [{kind}] {p} {name}", file=sys.stdout)
            tail = body.strip()
            if len(tail) > 12000:
                tail = "…\n" + tail[-12000:]
            if tail:
                print(tail, file=sys.stdout)
        print("\nFAIL", file=sys.stdout)
        return 0

    print("PASS", file=sys.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
