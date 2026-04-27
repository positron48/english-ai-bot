package grammartrainingpack

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed en es
var rootFS embed.FS

// FS is the default embedded training pack (English).
var FS fs.FS

func init() {
	var err error
	FS, err = PackFS("en")
	if err != nil {
		panic("grammartrainingpack: embedded en pack: " + err.Error())
	}
}

// PackFS returns embedded grammar training pack FS for packID (en, es, ...).
func PackFS(packID string) (fs.FS, error) {
	id := strings.TrimSpace(strings.ToLower(packID))
	if id == "" {
		id = "en"
	}
	sub, err := fs.Sub(rootFS, id)
	if err != nil {
		return nil, fmt.Errorf("grammar training pack %q: %w", id, err)
	}
	return sub, nil
}

