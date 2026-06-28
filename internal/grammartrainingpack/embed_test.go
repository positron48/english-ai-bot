package grammartrainingpack

import (
	"io/fs"
	"testing"
)

func TestPackFS_EnglishAndSpanish(t *testing.T) {
	t.Parallel()
	for _, pack := range []string{"en", "es", "EN", " ES "} {
		sub, err := PackFS(pack)
		if err != nil {
			t.Fatalf("PackFS(%q): %v", pack, err)
		}
		entries, err := fs.ReadDir(sub, ".")
		if err != nil {
			t.Fatalf("ReadDir(%q): %v", pack, err)
		}
		if len(entries) == 0 {
			t.Fatalf("PackFS(%q): empty dir", pack)
		}
	}
}

func TestPackFS_DefaultsToEnglish(t *testing.T) {
	t.Parallel()
	sub, err := PackFS("")
	if err != nil {
		t.Fatal(err)
	}
	en, err := PackFS("en")
	if err != nil {
		t.Fatal(err)
	}
	enEntries, err := fs.ReadDir(en, ".")
	if err != nil {
		t.Fatal(err)
	}
	defEntries, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(enEntries) != len(defEntries) {
		t.Fatalf("default pack len=%d en len=%d", len(defEntries), len(enEntries))
	}
}

func TestPackFS_InvalidPack(t *testing.T) {
	t.Parallel()
	sub, err := PackFS("zz-not-a-pack")
	if err != nil {
		return
	}
	// fs.Sub may succeed with an empty subtree when the embed path is missing.
	if _, err := fs.ReadDir(sub, "."); err != nil {
		return
	}
	t.Fatal("expected error or unreadable pack for invalid id")
}

func TestDefaultFSReadable(t *testing.T) {
	t.Parallel()
	if FS == nil {
		t.Fatal("FS is nil")
	}
	if _, err := fs.ReadDir(FS, "."); err != nil {
		t.Fatalf("ReadDir default FS: %v", err)
	}
}
