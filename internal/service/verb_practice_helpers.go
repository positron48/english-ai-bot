package service

import (
	"encoding/json"
	"golang.org/x/text/unicode/norm"
	"strings"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"
)

func verbPromptString(p map[string]interface{}, key string) string {
	value, _ := p[key].(string)
	return value
}
func VerbScopeAllowed(item repository.VerbQueueCard, scopes []string) bool {
	var p map[string]interface{}
	_ = json.Unmarshal([]byte(item.PromptJSON), &p)
	if p["content_version"] != float64(4) || p["practice_eligible"] != true {
		return false
	}
	scope := verbtraining.CanonicalScope("es." + verbPromptString(p, "tense") + "." + verbPromptString(p, "mood"))
	for _, s := range scopes {
		if scope == verbtraining.CanonicalScope(s) {
			return true
		}
	}
	return false
}

func GradeVerbPractice(item repository.VerbQueueCard, answerText string, skip, assisted bool) repository.VerbPracticeFeedback {
	var p map[string]interface{}
	_ = json.Unmarshal([]byte(item.PromptJSON), &p)
	var answer map[string]string
	_ = json.Unmarshal([]byte(item.AnswerJSON), &answer)
	expected := strings.TrimSpace(answer["surface_form"])
	got := strings.TrimSpace(answerText)
	outcome := "incorrect"
	if skip {
		outcome = "unknown"
		got = ""
	} else if spanishverbs.AcceptedVerbAnswer(expected, got, verbPromptString(p, "mood"), verbPromptString(p, "tense")) {
		outcome = "correct"
	}
	rule, _ := json.Marshal(p["rule"])
	sentence := verbPromptString(p, "question")
	sentence = strings.ReplaceAll(sentence, "____", expected)
	sentence = strings.ReplaceAll(sentence, "…", expected)
	sentence = strings.ReplaceAll(sentence, "...", expected)
	feedback := repository.VerbPracticeFeedback{CardID: item.UserVerbCardID, Outcome: outcome, Assisted: assisted, Chosen: got, Correct: expected, Sentence: sentence, Translation: verbPromptString(p, "example_translation"), Lemma: verbPromptString(p, "lemma"), Tense: verbPromptString(p, "tense"), Mood: verbPromptString(p, "mood"), Person: verbPromptString(p, "person"), Number: verbPromptString(p, "number"), Rule: rule, AccentOnly: outcome == "incorrect" && StripVerbAccents(got) == StripVerbAccents(expected)}
	return feedback
}
func StripVerbAccents(s string) string {
	// Only the stress accent; ñ and ü are different letters/sounds, not missing stress.
	return norm.NFC.String(strings.ReplaceAll(norm.NFD.String(strings.ToLower(strings.TrimSpace(s))), "\u0301", ""))
}

func PublicVerbPractice(state *repository.VerbPractice) map[string]interface{} {
	result := map[string]interface{}{"session_id": state.ID, "card_index": state.Index + 1, "total_cards": len(state.Queue), "completed": state.Completed, "retry": state.Retry, "assisted": state.Assisted, "feedback": state.Feedback, "results": state.Results}
	if !state.Completed && state.Index < len(state.Queue) {
		item := state.Queue[state.Index]
		var p map[string]interface{}
		_ = json.Unmarshal([]byte(item.PromptJSON), &p)
		// Whitelist task data; translation is part of the task; the Spanish answer remains private.
		prompt := map[string]interface{}{}
		for _, key := range []string{"question", "lemma", "ru_gloss", "example_translation", "person", "number"} {
			prompt[key] = p[key]
		}
		if state.Assisted || state.Feedback != nil {
			prompt["rule"] = p["rule"]
			prompt["tense"] = p["tense"]
			prompt["mood"] = p["mood"]
		}
		options := ParseStringJSONArray(item.DistractorsJSON)
		mode := item.InputMode
		if mode == "" {
			mode = "choice"
		}
		if len(options) < 2 {
			mode = "typed"
		}
		if mode == "typed" {
			options = nil
		}
		result["card_id"] = item.UserVerbCardID
		result["prompt"] = prompt
		result["options"] = options
		result["input_mode"] = mode
	}
	return result
}
