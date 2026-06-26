package readingcms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScriptEnvLoadsCourseOverlay(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=base\nLLAMACPP_URL=http://127.0.0.1:8090\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.es"), []byte("FOO=es\nLLAMACPP_START_CMD_READING=echo start\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := NewPaths(root)
	env := environMap(p.ScriptEnv("es_ru"))
	if env["FOO"] != "es" {
		t.Fatalf("FOO=%q want es", env["FOO"])
	}
	if env["LLAMACPP_START_CMD_READING"] != "echo start" {
		t.Fatalf("missing LLAMACPP_START_CMD_READING")
	}
	if env["LLAMACPP_URL"] != "http://127.0.0.1:8090" {
		t.Fatalf("LLAMACPP_URL=%q", env["LLAMACPP_URL"])
	}
}
