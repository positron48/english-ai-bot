package service

import (
	"strings"

	"tgbot-skeleton/internal/config"
)

func requireNativeCyrillic(learning config.LearningConfig) bool {
	return strings.EqualFold(learning.NativeLang, "ru") && strings.EqualFold(learning.TargetLang, "es")
}

func validDefinitionNativeForLearning(definitionRU string, learning config.LearningConfig) bool {
	if !requireNativeCyrillic(learning) {
		return true
	}
	return ContainsCyrillic(strings.TrimSpace(definitionRU))
}

