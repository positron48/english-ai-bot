package ai

import (
	"os"
	"strings"

	"tgbot-skeleton/prompts"
)

// RenderLearningPromptTemplate substitutes {{native_lang}}, {{target_lang}}, and {{pair}}.
func RenderLearningPromptTemplate(template string, nativeLang, targetLang, pair string) string {
	s := template
	s = strings.ReplaceAll(s, "{{native_lang}}", nativeLang)
	s = strings.ReplaceAll(s, "{{target_lang}}", targetLang)
	s = strings.ReplaceAll(s, "{{pair}}", pair)
	return s
}

// PreparePrompt applies escaped-newline processing and learning-template substitution.
// This is the single entry point used after loading AI_PROMPT / file or TRAINING_PROMPT_FILE.
func PreparePrompt(prompt string, nativeLang, targetLang, pair string) string {
	s := strings.ReplaceAll(prompt, "\\n", "\n")
	return RenderLearningPromptTemplate(s, nativeLang, targetLang, pair)
}

// LoadRenderedPromptFile reads a prompt file and runs PreparePrompt. An on-disk file (e.g. a
// deployment override) takes precedence; if it is not present, the prompt embedded into the
// binary at build time is used, so prompt loading never depends on the working directory.
func LoadRenderedPromptFile(path string, nativeLang, targetLang, pair string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		embedded, embErr := prompts.FS.ReadFile(strings.TrimPrefix(path, "prompts/"))
		if embErr != nil {
			return "", err
		}
		content = embedded
	}
	raw := strings.TrimSpace(string(content))
	return PreparePrompt(raw, nativeLang, targetLang, pair), nil
}
