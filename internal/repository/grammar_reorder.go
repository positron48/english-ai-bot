package repository

import "strings"

// GrammarQuestionAvailable prevents word-order questions without a Russian
// sentence from being selected. Keep the original bank for grading old attempts.
func GrammarQuestionAvailable(question map[string]interface{}) bool {
	if question["type"] != "reorder" {
		return true
	}
	translation, _ := question["translation_ru"].(string)
	return strings.TrimSpace(translation) != ""
}

func normalizeGrammarSentence(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// enrichReorderTranslations reuses only an exact example from the question's
// own theory block. Explanations, partial matches and other languages are not
// translations of the target sentence.
func enrichReorderTranslations(chapter *Chapter) {
	if chapter.UILanguage != "ru" {
		return
	}
	translations := make(map[string]map[string]string)
	for _, raw := range chapter.Blocks {
		block, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := block["id"].(string)
		theory, _ := block["theory"].(map[string]interface{})
		examples, _ := theory["examples"].([]interface{})
		byText := make(map[string]string)
		for _, rawExample := range examples {
			example, ok := rawExample.(map[string]interface{})
			if !ok {
				continue
			}
			text, _ := example["text"].(string)
			translation, _ := example["translation"].(string)
			text = normalizeGrammarSentence(text)
			translation = strings.TrimSpace(translation)
			if text == "" || translation == "" {
				continue
			}
			if previous, exists := byText[text]; exists && previous != translation {
				byText[text] = "" // Conflicting translations: do not guess.
			} else {
				byText[text] = translation
			}
		}
		translations[id] = byText
	}
	questions, _ := chapter.QuestionBank["questions"].([]interface{})
	for _, raw := range questions {
		question, ok := raw.(map[string]interface{})
		if !ok || GrammarQuestionAvailable(question) {
			continue
		}
		blockID, _ := question["theory_block_id"].(string)
		answer, _ := question["correct_answer"].(string)
		if translation := translations[blockID][normalizeGrammarSentence(answer)]; translation != "" {
			question["translation_ru"] = translation
		}
	}
}
