#!/usr/bin/env python3
"""
Parallel orchestrator for `make check`.

Runs the independent CI check stages concurrently, captures each stage's output
to its own buffer, shows a live status board, then prints an ordered summary —
dumping full output for any failed stage. Mirrors the same stages/commands as the
sequential `check` target, so the *result* is identical to CI; only the wall-clock
time and the presentation differ.

Configuration via env (set by the Makefile):
  GO                 go binary (default: go)
  COVER_PKGS         space-separated package list for coverage tests
  GOLANGCI           golangci-lint invocation (binary or `go run ...`)
  CHECK_P            test parallelism (-p), default 8
  CHECK_COUNT        if "1", pass -count=1 (disable Go test cache; CI parity)
  CHECK_SKIP_INTEGRATION  if "1", skip the integration stage (check-quick)
  CHECK_COVERAGE_MIN minimum total coverage %, default 50
"""

from __future__ import annotations

import os
import subprocess
import sys
import threading
import time

GO = os.environ.get("GO", "go")
COVER_PKGS = os.environ.get("COVER_PKGS", "").strip()
# Pinned to match CI (.github/workflows/ci.yml). Lint findings depend on the
# linter version, so we must use exactly this version for CI parity.
GOLANGCI_VERSION = os.environ.get("GOLANGCI_VERSION", "v2.10.1").strip()
CHECK_P = os.environ.get("CHECK_P", "8").strip() or "8"
CHECK_COUNT = os.environ.get("CHECK_COUNT", "0").strip()
SKIP_INTEGRATION = os.environ.get("CHECK_SKIP_INTEGRATION", "0").strip() == "1"
COVERAGE_MIN = int(os.environ.get("CHECK_COVERAGE_MIN", "75"))

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# webapp npm guard: reinstall only when package-lock changed since last install.
WEBAPP_DEPS = r"""
set -e
cd webapp
stamp=node_modules/.check-stamp
if [ ! -d node_modules ] || [ ! -f "$stamp" ] || [ package-lock.json -nt "$stamp" ]; then
  echo "Installing webapp dependencies..."
  npm ci --prefer-offline --no-audit --no-fund || npm install --no-audit --no-fund
  touch "$stamp"
else
  echo "Webapp dependencies up to date (skipped npm install)."
fi
"""

WEBAPP = (
    WEBAPP_DEPS
    + "\necho '--- type-check ---'\ncd webapp && npm run type-check\n"
    + "echo '--- unit tests ---'\nnpm test\n"
    + "echo '--- build ---'\nnpm run build\n"
)

def resolve_golangci() -> str:
    """Use ./bin/golangci-lint only if it matches the pinned version (CI parity);
    otherwise fall back to `go run ...@<pinned>` (slower start, identical findings)."""
    binpath = os.path.join(REPO_ROOT, "bin", "golangci-lint")
    pinned = GOLANGCI_VERSION.lstrip("v")
    if os.path.exists(binpath):
        try:
            out = subprocess.run(
                [binpath, "version"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                text=True, timeout=15,
            ).stdout
            if f"version {pinned} " in out or f"version {pinned}\n" in out:
                return binpath
        except (OSError, subprocess.SubprocessError):
            pass
    return f"{GO} run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{GOLANGCI_VERSION}"


GOLANGCI = resolve_golangci()

_count_flag = "-count=1 " if CHECK_COUNT == "1" else ""
GO_TEST = (
    "set -o pipefail\n"
    "rm -f coverage.out .go-test-output.jsonl\n"
    f"{GO} test -tags=test {_count_flag}-p {CHECK_P} -timeout 30m "
    "-coverprofile=coverage.out -covermode=atomic -json "
    f"{COVER_PKGS} 2>&1 | tee .go-test-output.jsonl | python3 scripts/go-test-compact.py\n"
)


class Stage:
    def __init__(self, name, title, command, depends_on=None):
        self.name = name
        self.title = title
        self.command = command
        self.depends_on = depends_on  # stage name or None
        self.status = "pending"  # pending | running | ok | fail | skip
        self.output = ""
        self.duration = 0.0

    def run(self):
        self.status = "running"
        start = time.time()
        proc = subprocess.run(
            ["/bin/bash", "-c", self.command],
            cwd=REPO_ROOT,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
        )
        self.duration = time.time() - start
        self.output = proc.stdout or ""
        self.status = "ok" if proc.returncode == 0 else "fail"


def coverage_gate(go_test_stage: Stage) -> Stage:
    st = Stage("coverage-gate", "Coverage threshold", "", depends_on="go-test")
    # Wait for the go-test stage to settle.
    while go_test_stage.status in ("pending", "running"):
        time.sleep(0.1)
    if go_test_stage.status != "ok":
        st.status = "skip"
        st.output = "skipped: go-test did not pass"
        return st
    st.status = "running"
    start = time.time()
    proc = subprocess.run(
        ["/bin/bash", "-c", f"{GO} tool cover -func=coverage.out"],
        cwd=REPO_ROOT,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )
    st.duration = time.time() - start
    total = ""
    for line in proc.stdout.splitlines():
        if line.startswith("total:"):
            total = line.split()[-1].rstrip("%")
    if proc.returncode != 0 or not total:
        st.status = "fail"
        st.output = "❌ Failed to get coverage\n" + proc.stdout
        return st
    try:
        cov_int = int(float(total))
    except ValueError:
        st.status = "fail"
        st.output = f"❌ Could not parse coverage: {total!r}\n"
        return st
    if cov_int < COVERAGE_MIN:
        st.status = "fail"
        st.output = f"❌ Test coverage is {total}% (minimum required: {COVERAGE_MIN}%)\n"
    else:
        st.status = "ok"
        st.output = f"✅ Test coverage: {total}% (minimum: {COVERAGE_MIN}%)\n"
    return st


SYMBOL = {
    "pending": "·",
    "running": "▶",
    "ok": "✅",
    "fail": "❌",
    "skip": "⏭",
}


def build_stages():
    stages = [
        Stage("webapp", "Webapp: deps + type-check + build", WEBAPP),
        Stage("go-mod-verify", "Go: mod verify", f"{GO} mod verify"),
        Stage(
            "reading-en",
            "Reading artifacts: english-grammar",
            "$(command -v gmake make | head -1) -C courses/english-grammar reading-validate",
        ),
        Stage(
            "reading-es",
            "Reading artifacts: spanish-grammar",
            "$(command -v gmake make | head -1) -C courses/spanish-grammar reading-validate",
        ),
        Stage("go-test", "Go: tests + coverage profile", GO_TEST),
        Stage("golangci-lint", "Go: golangci-lint", f"{GOLANGCI} run --timeout=3m"),
    ]
    if not SKIP_INTEGRATION:
        stages.append(
            Stage(
                "integration",
                "Integration tests (Testcontainers)",
                "$(command -v gmake make | head -1) test-integration",
            )
        )
    return stages


def render_board(stages, derived, started_at, final=False):
    lines = []
    for st in stages + derived:
        sym = SYMBOL[st.status]
        if st.status in ("ok", "fail", "skip"):
            extra = f"({st.duration:.1f}s)" if st.duration else ""
        elif st.status == "running":
            extra = f"({time.time() - started_at:.0f}s)"
        else:
            extra = ""
        lines.append(f"  {sym} {st.title} {extra}")
    return "\n".join(lines)


def main() -> int:
    if not COVER_PKGS:
        print("run_check.py: COVER_PKGS env is empty (Makefile should set it)", file=sys.stderr)
        return 2

    stages = build_stages()
    started_at = time.time()
    threads = []

    print("=== Running CI checks (parallel) ===\n")

    for st in stages:
        t = threading.Thread(target=st.run, daemon=True)
        t.start()
        threads.append(t)

    # Coverage gate depends on go-test; run it in its own thread.
    go_test = next(s for s in stages if s.name == "go-test")
    cov = {}

    def _cov():
        cov["stage"] = coverage_gate(go_test)

    cov_thread = threading.Thread(target=_cov, daemon=True)
    cov_thread.start()

    is_tty = sys.stdout.isatty()
    prev_lines = 0
    while any(t.is_alive() for t in threads) or cov_thread.is_alive():
        derived = [cov["stage"]] if "stage" in cov else []
        board = render_board(stages, derived, started_at)
        if is_tty:
            if prev_lines:
                sys.stdout.write(f"\033[{prev_lines}A")
            sys.stdout.write("\033[J" + board + "\n")
            sys.stdout.flush()
            prev_lines = board.count("\n") + 1
        time.sleep(0.4)

    cov_thread.join()
    derived = [cov["stage"]]
    board = render_board(stages, derived, started_at, final=True)
    if is_tty and prev_lines:
        sys.stdout.write(f"\033[{prev_lines}A\033[J")
    print(board)

    all_stages = stages + derived
    failed = [s for s in all_stages if s.status == "fail"]
    total_dur = time.time() - started_at

    for st in all_stages:
        # Always show go-test output (compact summary is useful even on pass);
        # show full output for any failed stage.
        if st.status == "fail" or st.name == "go-test":
            print(f"\n─── {st.title} [{SYMBOL[st.status]}] ───")
            print(st.output.rstrip() or "(no output)")

    print(f"\n⏱  Total wall time: {total_dur:.1f}s")

    if failed:
        names = ", ".join(s.name for s in failed)
        print(f"\n❌ FAILED stages: {names}")
        print("ℹ️  Raw Go test JSON stream (if present): .go-test-output.jsonl")
        return 1

    # Success: tidy raw stream, print parity footer.
    try:
        os.remove(os.path.join(REPO_ROOT, ".go-test-output.jsonl"))
    except OSError:
        pass
    print("\n🎉 All CI checks passed!")
    proc = subprocess.run(
        ["/bin/bash", "-c", f"{GO} tool cover -func=coverage.out | awk '/^total:/ {{print $3}}'"],
        cwd=REPO_ROOT,
        stdout=subprocess.PIPE,
        text=True,
    )
    print(f"📊 Total test coverage: {proc.stdout.strip()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
