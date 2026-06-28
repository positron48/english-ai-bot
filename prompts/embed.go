// Package prompts embeds the prompt template files so they are always available to the
// binary regardless of the process working directory. Loading a course's prompt must never
// silently fail (which would make word-card generation fall back to the default English
// prompt and produce wrong-language cards), so the files are compiled into the binary.
package prompts

import "embed"

//go:embed *.txt
var FS embed.FS
