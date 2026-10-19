package service

import (
	"fmt"
	"strings"
	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/learning"
)

// TrainingQuestion is shared by ordinary sessions and daily journey snapshots.
// Resolve UI language on every read so persisted sessions follow locale changes.
func TrainingQuestion(lang, target, mode, direction, native, display, transcription string) string {
	var name string
	switch lang {
	case "ru":
		name = learning.TargetLangNameRUAccusative(target)
	case "es":
		name = learning.TargetLangNameES(target)
	default:
		name = learning.TargetLangNameEN(target)
	}
	if mode == "spell" || mode == "type" {
		if lang == "ru" {
			name = learning.TargetLangNameRUPrepositional(target)
		}
		return fmt.Sprintf(i18n.T(lang, "training."+mode+"Word"), name, native)
	}
	if direction == "ru_en" {
		return fmt.Sprintf(i18n.T(lang, "training.translateTo"), name, native)
	}
	if target != "en" {
		display = strings.TrimPrefix(display, "to ")
	}
	suffix := ""
	if transcription != "" {
		suffix = fmt.Sprintf(` <span class="transcription">%s</span>`, transcription)
	}
	return fmt.Sprintf(i18n.T(lang, "training.whatMeansWord"), display, suffix)
}
