package webapp

import "embed"

// Embed full dist tree (including nested assets) into the binary.
//go:embed all:dist
var FS embed.FS
