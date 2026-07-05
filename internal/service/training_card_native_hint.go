package service

import (
	"fmt"
	"strings"
)

// NativeLookupHintWord returns trimmed input when it is a native-language (Cyrillic) lookup token.
func NativeLookupHintWord(input string) string {
	input = strings.TrimSpace(input)
	if input == "" || !ContainsCyrillic(input) {
		return ""
	}
	return input
}

// BuildTrainingCardNativeHintInstruction adds LLM guidance to prefer the learner's original native query.
func BuildTrainingCardNativeHintInstruction(nativeWord, nativeLang string) string {
	nativeWord = strings.TrimSpace(nativeWord)
	if nativeWord == "" || !ContainsCyrillic(nativeWord) {
		return ""
	}
	if strings.TrimSpace(nativeLang) == "" {
		nativeLang = "native"
	}
	return fmt.Sprintf(
		"NATIVE LOOKUP HINT: The learner searched using the %s word %q. Use this exact Cyrillic form as word_native (and legacy word_ru) for the PRIMARY sense (senses[0]) when it correctly matches the meaning of the target word. Do not transliterate the target spelling into Cyrillic.",
		nativeLang,
		nativeWord,
	)
}
