package models

import "strings"

// Report categories submitted from the webapp (stored in content_reports.report_category).

var WordTrainingReportCategories = []string{
	"wrong_translation",
	"wrong_example",
	"wrong_distractors",
	"typo",
	"bad_audio",
	"unclear_question",
	"other",
}

var GrammarTrainingReportCategories = []string{
	"wrong_answer",
	"ambiguous",
	"wrong_explanation",
	"theory_mismatch",
	"typo",
	"too_hard",
	"other",
}

func IsValidReportCategory(sourceType, category string) bool {
	category = trimCategory(category)
	if category == "" {
		return false
	}
	switch sourceType {
	case "word_training":
		for _, c := range WordTrainingReportCategories {
			if c == category {
				return true
			}
		}
	case "grammar_training":
		for _, c := range GrammarTrainingReportCategories {
			if c == category {
				return true
			}
		}
	}
	return false
}

func NormalizeReportCategory(sourceType, category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return "other"
	}
	if IsValidReportCategory(sourceType, category) {
		return category
	}
	return "other"
}

func trimCategory(s string) string {
	return strings.TrimSpace(s)
}
