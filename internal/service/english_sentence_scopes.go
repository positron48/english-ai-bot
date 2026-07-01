package service

import (
	"sort"
	"strings"

	"tgbot-skeleton/internal/repository"
)

type EnglishSentenceScopeRule struct {
	Scope      string
	Label      string
	SectionID  string
	ChapterIDs []string
}

func EnglishSentenceScopeRules() []EnglishSentenceScopeRule {
	return []EnglishSentenceScopeRule{
		{
			Scope:     "en.be_present",
			Label:     "Present forms of be",
			SectionID: "en.grammar.first_sentences_be_as",
			ChapterIDs: []string{
				"en.grammar.first_sentences_be_as.personal_pronouns_am_is",
				"en.grammar.first_sentences_be_as.statements_with_be_identity",
				"en.grammar.first_sentences_be_as.negatives_with_be_not",
				"en.grammar.first_sentences_be_as.questions_with_be_are",
			},
		},
		{
			Scope:     "en.present_simple",
			Label:     "Present Simple",
			SectionID: "en.grammar.first_actions_present_simple",
			ChapterIDs: []string{
				"en.grammar.first_actions_present_simple.verbs_as_actions_base",
				"en.grammar.first_actions_present_simple.present_simple_statements_i",
				"en.grammar.first_actions_present_simple.negatives_don_t_doesn",
				"en.grammar.first_actions_present_simple.questions_with_do_does",
			},
		},
		{
			Scope:     "en.present_continuous",
			Label:     "Present Continuous",
			SectionID: "en.grammar.now_present_continuous_contrast",
			ChapterIDs: []string{
				"en.grammar.now_present_continuous_contrast.form_am_is_are",
				"en.grammar.now_present_continuous_contrast.meanings_right_now_temporary",
				"en.grammar.now_present_continuous_contrast.simple_vs_continuous_clear",
			},
		},
		{
			Scope:     "en.past_simple",
			Label:     "Past Simple",
			SectionID: "en.grammar.past_1_past_simple",
			ChapterIDs: []string{
				"en.grammar.past_1_past_simple.past_of_be_was",
				"en.grammar.past_1_past_simple.regular_verbs_v2_basic",
				"en.grammar.past_1_past_simple.negatives_didn_t_v1",
				"en.grammar.past_1_past_simple.questions_with_did_did",
			},
		},
		{
			Scope:     "en.past_continuous",
			Label:     "Past Continuous",
			SectionID: "en.grammar.past_2_background_process",
			ChapterIDs: []string{
				"en.grammar.past_2_background_process.form_was_were_v",
				"en.grammar.past_2_background_process.past_simple_vs_past",
				"en.grammar.past_2_background_process.two_parallel_actions_while",
			},
		},
		{
			Scope:     "en.future_will",
			Label:     "Future with will",
			SectionID: "en.grammar.future_plans",
			ChapterIDs: []string{
				"en.grammar.future_plans.will_instant_decisions_promises",
			},
		},
		{
			Scope:     "en.going_to",
			Label:     "Future with going to",
			SectionID: "en.grammar.future_plans",
			ChapterIDs: []string{
				"en.grammar.future_plans.be_going_to_intention",
			},
		},
		{
			Scope:     "en.future_arrangements",
			Label:     "Present Continuous for arrangements",
			SectionID: "en.grammar.future_plans",
			ChapterIDs: []string{
				"en.grammar.future_plans.present_continuous_for_arrangements",
			},
		},
		{
			Scope:     "en.modals",
			Label:     "Core modals",
			SectionID: "en.grammar.core_modality_ability_permission",
			ChapterIDs: []string{
				"en.grammar.core_modality_ability_permission.can_can_t_ability",
				"en.grammar.core_modality_ability_permission.must_vs_have_to",
				"en.grammar.core_modality_ability_permission.should_advice_expectation",
			},
		},
		{
			Scope:     "en.present_perfect",
			Label:     "Present Perfect",
			SectionID: "en.grammar.perfect_aspect_experience_result",
			ChapterIDs: []string{
				"en.grammar.perfect_aspect_experience_result.present_perfect_form_have",
				"en.grammar.perfect_aspect_experience_result.experience_have_you_ever",
				"en.grammar.perfect_aspect_experience_result.result_i_ve_lost",
			},
		},
		{
			Scope:     "en.past_perfect",
			Label:     "Past Perfect",
			SectionID: "en.grammar.perfect_aspect_experience_result",
			ChapterIDs: []string{
				"en.grammar.perfect_aspect_experience_result.past_perfect_past_before",
			},
		},
		{
			Scope:     "en.future_perfect",
			Label:     "Future Perfect",
			SectionID: "en.grammar.perfect_aspect_experience_result",
			ChapterIDs: []string{
				"en.grammar.perfect_aspect_experience_result.future_perfect_future_continuous",
			},
		},
		{
			Scope:     "en.future_continuous",
			Label:     "Future Continuous",
			SectionID: "en.grammar.perfect_aspect_experience_result",
			ChapterIDs: []string{
				"en.grammar.perfect_aspect_experience_result.future_perfect_future_continuous",
			},
		},
		{
			Scope:     "en.passive",
			Label:     "Passive voice",
			SectionID: "en.grammar.voice_focus_passive_and",
			ChapterIDs: []string{
				"en.grammar.voice_focus_passive_and.passive_basics_be_v3",
				"en.grammar.voice_focus_passive_and.passive_across_tenses",
			},
		},
		{
			Scope:     "en.conditionals",
			Label:     "Conditionals",
			SectionID: "en.grammar.conditionals_hypotheticals",
			ChapterIDs: []string{
				"en.grammar.conditionals_hypotheticals.zero_conditional",
				"en.grammar.conditionals_hypotheticals.first_conditional",
				"en.grammar.conditionals_hypotheticals.second_conditional",
				"en.grammar.conditionals_hypotheticals.third_conditional",
			},
		},
	}
}

func DefaultEnglishSentenceScopes() []string {
	return []string{"en.be_present", "en.present_simple"}
}

func ResolveEnglishSentenceScopes(
	sectionsData *repository.SectionsData,
	placement *repository.PlacementTestResult,
	progressByChapter map[string]*repository.ChapterProgress,
) []string {
	if sectionsData == nil {
		return DefaultEnglishSentenceScopes()
	}

	sectionByID := make(map[string]repository.Section, len(sectionsData.Sections))
	maxPlacementLevel := -1
	opened := map[string]bool{}
	for _, section := range sectionsData.Sections {
		sectionByID[section.SectionID] = section
	}
	if placement != nil {
		for _, sectionID := range placement.OpenedSections {
			sectionID = strings.TrimSpace(sectionID)
			if sectionID == "" {
				continue
			}
			opened[sectionID] = true
			if section, ok := sectionByID[sectionID]; ok {
				if ord := englishLevelOrder(section.Level); ord > maxPlacementLevel {
					maxPlacementLevel = ord
				}
			}
		}
	}

	out := make([]string, 0, len(EnglishSentenceScopeRules()))
	seen := map[string]bool{}
	for _, rule := range EnglishSentenceScopeRules() {
		if englishRuleUnlocked(rule, sectionByID, opened, maxPlacementLevel, progressByChapter) {
			if !seen[rule.Scope] {
				out = append(out, rule.Scope)
				seen[rule.Scope] = true
			}
		}
	}
	if len(out) == 0 {
		return DefaultEnglishSentenceScopes()
	}
	return out
}

func EnglishSentenceScopeLabels(scopes []string) []string {
	labelByScope := map[string]string{}
	for _, rule := range EnglishSentenceScopeRules() {
		labelByScope[rule.Scope] = rule.Label
	}
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if scope == "" {
			continue
		}
		if label := labelByScope[scope]; label != "" {
			out = append(out, label)
		} else {
			out = append(out, strings.ReplaceAll(strings.TrimPrefix(scope, "en."), "_", " "))
		}
	}
	return out
}

func englishRuleUnlocked(
	rule EnglishSentenceScopeRule,
	sectionByID map[string]repository.Section,
	opened map[string]bool,
	maxPlacementLevel int,
	progressByChapter map[string]*repository.ChapterProgress,
) bool {
	if opened[rule.SectionID] {
		return true
	}
	if maxPlacementLevel >= 0 {
		if section, ok := sectionByID[rule.SectionID]; ok {
			ord := englishLevelOrder(section.Level)
			if ord >= 0 && ord <= maxPlacementLevel {
				return true
			}
		}
	}
	for _, chapterID := range rule.ChapterIDs {
		if progress := progressByChapter[chapterID]; progress != nil && progress.Passed {
			return true
		}
	}
	return false
}

func englishLevelOrder(level string) int {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "A0":
		return 0
	case "A1":
		return 1
	case "A2":
		return 2
	case "B1":
		return 3
	case "B2":
		return 4
	case "C1":
		return 5
	case "C2":
		return 6
	default:
		return -1
	}
}

func sortedEnglishScopesForTest(scopes []string) []string {
	out := append([]string(nil), scopes...)
	sort.Strings(out)
	return out
}
