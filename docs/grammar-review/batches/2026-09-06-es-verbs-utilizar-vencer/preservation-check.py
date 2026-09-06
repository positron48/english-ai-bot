#!/usr/bin/env python3
"""Verify files outside the utilizar/vaciar/valer/vencer batch are unchanged."""

import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BATCH = Path(__file__).resolve().parent
SNAPSHOT = json.loads(Path("/tmp/grammar-es-verbs-utilizar-vencer-preserve.json").read_text())
META = json.loads((BATCH / "preservation-before.json").read_text())
allowed = set(META["allowed_paths"])
batch_rel = BATCH.relative_to(ROOT).as_posix()


def status_paths(repo, prefix=""):
    raw = subprocess.check_output([
        "git", "-C", str(repo), "status", "--porcelain=v1", "-z", "--untracked-files=all"
    ]).decode("utf-8", "surrogateescape").split("\0")
    result = []
    index = 0
    while index < len(raw) - 1:
        item = raw[index]
        index += 1
        if not item:
            continue
        code, path = item[:2], item[3:]
        if code[0] in "RC" and index < len(raw) - 1:
            index += 1
        result.append((Path(prefix) / path).as_posix() if prefix else path)
    return result


current_paths = set()
for repo, prefix in [
    (ROOT, ""),
    (ROOT / "courses/english-grammar", "courses/english-grammar"),
    (ROOT / "courses/spanish-grammar", "courses/spanish-grammar"),
]:
    current_paths.update(status_paths(repo, prefix))

preserved = {rel: expected for rel, expected in SNAPSHOT.items() if rel not in allowed}
changed = []
missing = []
for rel, expected in preserved.items():
    path = ROOT / rel
    if expected is None:
        if path.exists():
            changed.append(rel)
    elif not path.exists():
        missing.append(rel)
    elif hashlib.sha256(path.read_bytes()).hexdigest() != expected:
        changed.append(rel)

unexpected = sorted(
    rel for rel in current_paths
    if rel not in SNAPSHOT
    and rel not in allowed
    and rel != batch_rel
    and not rel.startswith(batch_rel + "/")
    and not (ROOT / rel).is_dir()
)
print("preserved_total", len(preserved))
print("changed", len(changed))
print("missing", len(missing))
print("unexpected_new", len(unexpected))
if changed:
    print("changed_paths", *changed, sep="\n")
if missing:
    print("missing_paths", *missing, sep="\n")
if unexpected:
    print("unexpected_paths", *unexpected, sep="\n")
raise SystemExit(1 if changed or missing or unexpected else 0)
