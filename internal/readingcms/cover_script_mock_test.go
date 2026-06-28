package readingcms

import (
	"os"
	"path/filepath"
	"testing"
)

// mockGenerateReadingCoverScript is a PIL-free stub for unit tests (CI has no Pillow).
const mockGenerateReadingCoverScript = `#!/usr/bin/env python3
import argparse
import json
import pathlib
import sys

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--course-root")
    p.add_argument("--draft-dir", default="reading")
    p.add_argument("--text-id", default="")
    p.add_argument("--image-prompt", default="")
    p.add_argument("--prompt-only", action="store_true")
    p.add_argument("--force", action="store_true")
    p.add_argument("--limit", type=int, default=0)
    args = p.parse_args()

    if args.text_id == "__MOCK_COVER_FAIL__" or (args.image_prompt or "").strip() == "__MOCK_COVER_FAIL__":
        print("mock cover script failed", file=sys.stderr)
        sys.exit(1)

    if not args.text_id:
        print("reading-covers-batch completed")
        return

    root = pathlib.Path(args.course_root)
    idx_path = root / args.draft_dir / "index.json"
    if not idx_path.is_file():
        print("missing index", file=sys.stderr)
        sys.exit(1)
    idx = json.loads(idx_path.read_text(encoding="utf-8"))
    rel = (idx.get("texts") or {}).get(args.text_id)
    if not rel:
        print("text not in index", file=sys.stderr)
        sys.exit(1)
    json_path = root / args.draft_dir / rel
    doc = json.loads(json_path.read_text(encoding="utf-8"))
    text_id = str(doc.get("id") or args.text_id).strip()

    if args.prompt_only:
        doc["cover_image_prompt"] = "mock prompt-only watercolor plaza"
        json_path.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print("done (prompt saved, no image)")
        return

    prompt = (args.image_prompt or "mock full cover scene").strip()
    assets = root / "assets" / "reading" / text_id
    assets.mkdir(parents=True, exist_ok=True)
    for name in ("cover_thumb.webp", "cover_hero.webp", "cover_raw.png"):
        (assets / name).write_bytes(b"mock")
    doc["cover_thumb_rel_path"] = f"assets/reading/{text_id}/cover_thumb.webp"
    doc["cover_hero_rel_path"] = f"assets/reading/{text_id}/cover_hero.webp"
    doc["cover_image_prompt"] = prompt
    json_path.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("done")

if __name__ == "__main__":
    main()
`

func installMockCoverScript(t *testing.T, root string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot, err := FindRepoRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	repoScripts := filepath.Join(repoRoot, "scripts")
	localScripts := filepath.Join(root, "scripts")
	if err := os.MkdirAll(localScripts, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(repoScripts)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() == "generate-reading-cover.py" {
			continue
		}
		src := filepath.Join(repoScripts, entry.Name())
		dst := filepath.Join(localScripts, entry.Name())
		if err := os.Symlink(src, dst); err != nil && !os.IsExist(err) {
			t.Fatal(err)
		}
	}
	mockPath := filepath.Join(localScripts, "generate-reading-cover.py")
	if err := os.WriteFile(mockPath, []byte(mockGenerateReadingCoverScript), 0o755); err != nil {
		t.Fatal(err)
	}
}
