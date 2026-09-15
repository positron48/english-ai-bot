package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/text/unicode/norm"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"
)

func verbScopeAllowed(item repository.VerbQueueCard, scopes []string) bool {
	var p map[string]interface{}
	_ = json.Unmarshal([]byte(item.PromptJSON), &p)
	if p["content_version"] != float64(4) || p["practice_eligible"] != true {
		return false
	}
	scope := verbtraining.CanonicalScope("es." + promptString(p, "tense") + "." + promptString(p, "mood"))
	for _, s := range scopes {
		if scope == verbtraining.CanonicalScope(s) {
			return true
		}
	}
	return false
}

func (r *Router) handleVerbPractice(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !r.verbFormsEnabledForUser(req.Context(), userID) {
		r.writeVerbTrainingDisabled(w)
		return
	}
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	action := strings.TrimPrefix(req.URL.Path, "/api/verb-training/v2/")
	if action == "current" && req.Method == http.MethodGet {
		state, err := repo.LoadVerbPractice(userID)
		if err != nil {
			http.Error(w, "Unable to load practice", 500)
			return
		}
		if state == nil {
			writeJSONVerb(w, map[string]interface{}{"idle": true})
			return
		}
		if !state.Completed && state.Index < len(state.Queue) && !verbScopeAllowed(state.Queue[state.Index], r.getUserVerbScopes(req.Context(), userID)) {
			writeJSONVerb(w, map[string]interface{}{"idle": true})
			return
		}
		r.writeVerbPractice(w, state)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		SessionID int64  `json:"session_id"`
		CardID    int64  `json:"card_id"`
		Answer    string `json:"answer"`
		Skip      bool   `json:"skip"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	scopes := r.getUserVerbScopes(req.Context(), userID)
	if action == "start" || action == "repeat" {
		previous, err := repo.LoadVerbPractice(userID)
		if err != nil {
			http.Error(w, "Unable to load practice", 500)
			return
		}
		if action == "start" && previous != nil && !previous.Completed && previous.Index < len(previous.Queue) && verbScopeAllowed(previous.Queue[previous.Index], scopes) {
			r.writeVerbPractice(w, previous)
			return
		}
		state := &repository.VerbPractice{Version: 2, Retry: action == "repeat"}
		if state.Retry {
			if previous == nil || !previous.Completed || previous.ID != body.SessionID {
				http.Error(w, "Finish the session first", http.StatusConflict)
				return
			}
			wrong := map[int64]bool{}
			for _, result := range previous.Results {
				if result.Outcome != "correct" {
					wrong[result.CardID] = true
				}
			}
			for _, item := range previous.Queue {
				if wrong[item.UserVerbCardID] && verbScopeAllowed(item, scopes) {
					state.Queue = append(state.Queue, item)
					delete(wrong, item.UserVerbCardID)
				}
			}
		} else {
			svc := r.newVerbTrainingServiceForUser(req.Context(), userID)
			state.Queue, err = svc.PracticeQueue(userID, scopes)
			if err != nil && err.Error() != "no cards available for training" {
				http.Error(w, "Unable to prepare practice", 500)
				return
			}
		}
		if len(state.Queue) == 0 {
			writeJSONVerb(w, map[string]interface{}{"idle": true, "empty": true})
			return
		}
		for i := range state.Queue {
			item := &state.Queue[i]
			item.InputMode = "choice"
			srs, lookupErr := repo.GetVerbUserCardSRS(item.UserVerbCardID)
			if lookupErr != nil {
				http.Error(w, "Unable to load practice", 500)
				return
			}
			if len(service.ParseStringJSONArray(item.DistractorsJSON)) < 2 || (srs != nil && srs.Reps >= max(2, r.config.Training.VerbFormsTypedMinReps) && srs.State != "learning") {
				item.InputMode = "typed"
			}
		}
		if err = repo.CreateVerbPractice(userID, state, scopes); err != nil {
			http.Error(w, "Unable to start practice", 500)
			return
		}
		r.writeVerbPractice(w, state)
		return
	}
	if action != "answer" && action != "advance" && action != "help" {
		http.NotFound(w, req)
		return
	}
	state, err := repo.ChangeVerbPractice(req.Context(), userID, body.SessionID, func(state *repository.VerbPractice, tx *sql.Tx) error {
		// Repeated delivery of the same answer returns its existing result.
		if state.Feedback != nil && state.Feedback.CardID == body.CardID && action == "answer" {
			return nil
		}
		if state.Completed {
			if action == "advance" {
				return nil
			}
			return fmt.Errorf("session complete")
		}
		if state.Index >= len(state.Queue) {
			return fmt.Errorf("invalid position")
		}
		item := state.Queue[state.Index]
		if body.CardID != item.UserVerbCardID {
			if action == "advance" {
				for _, result := range state.Results {
					if result.CardID == body.CardID {
						return nil
					}
				}
			}
			return fmt.Errorf("card changed")
		}
		if !verbScopeAllowed(item, scopes) {
			return fmt.Errorf("time is not unlocked")
		}
		switch action {
		case "help":
			if state.Feedback == nil {
				state.Assisted = true
			}
			return nil
		case "advance":
			if state.Feedback == nil {
				return fmt.Errorf("answer first")
			}
			state.Index++
			state.Feedback = nil
			state.Assisted = false
			state.Completed = state.Index >= len(state.Queue)
			return nil
		}
		var p map[string]interface{}
		_ = json.Unmarshal([]byte(item.PromptJSON), &p)
		var answer map[string]string
		_ = json.Unmarshal([]byte(item.AnswerJSON), &answer)
		expected := strings.TrimSpace(answer["surface_form"])
		got := strings.TrimSpace(body.Answer)
		outcome := "incorrect"
		if body.Skip {
			outcome = "unknown"
			got = ""
		} else if spanishverbs.AcceptedVerbAnswer(expected, got, promptString(p, "mood"), promptString(p, "tense")) {
			outcome = "correct"
		}
		rule, _ := json.Marshal(p["rule"])
		sentence := promptString(p, "question")
		sentence = strings.ReplaceAll(sentence, "____", expected)
		sentence = strings.ReplaceAll(sentence, "…", expected)
		sentence = strings.ReplaceAll(sentence, "...", expected)
		feedback := repository.VerbPracticeFeedback{CardID: item.UserVerbCardID, Outcome: outcome, Assisted: state.Assisted, Chosen: got, Correct: expected, Sentence: sentence, Translation: promptString(p, "example_translation"), Lemma: promptString(p, "lemma"), Tense: promptString(p, "tense"), Mood: promptString(p, "mood"), Person: promptString(p, "person"), Number: promptString(p, "number"), Rule: rule, AccentOnly: outcome == "incorrect" && stripVerbAccents(got) == stripVerbAccents(expected)}
		card, err := repository.ReadVerbSRSTx(tx, userID, item.UserVerbCardID)
		if err != nil {
			return err
		}
		due, quality := service.AdvanceVerbSRS(card, outcome == "correct", state.Assisted)
		if err = repository.SaveVerbAttemptTx(tx, userID, state, card, due, quality, feedback); err != nil {
			return err
		}
		state.Feedback = &feedback
		state.Results = append(state.Results, feedback)
		return nil
	})
	if err != nil {
		r.logger.Warn("verb practice transition failed")
		http.Error(w, "Unable to update practice. Reload the current session.", http.StatusConflict)
		return
	}
	r.BumpUserCache(userID)
	r.writeVerbPractice(w, state)
}

func stripVerbAccents(s string) string {
	// Only the stress accent; ñ and ü are different letters/sounds, not missing stress.
	return norm.NFC.String(strings.ReplaceAll(norm.NFD.String(strings.ToLower(strings.TrimSpace(s))), "\u0301", ""))
}

func writeJSONVerb(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func (r *Router) writeVerbPractice(w http.ResponseWriter, state *repository.VerbPractice) {
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
		options := service.ParseStringJSONArray(item.DistractorsJSON)
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
	writeJSONVerb(w, result)
}
