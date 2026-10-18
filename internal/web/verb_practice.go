package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

func verbScopeAllowed(item repository.VerbQueueCard, scopes []string) bool {
	return service.VerbScopeAllowed(item, scopes)
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
		feedback := service.GradeVerbPractice(item, body.Answer, body.Skip, state.Assisted)
		outcome := feedback.Outcome
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

func writeJSONVerb(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func (r *Router) writeVerbPractice(w http.ResponseWriter, state *repository.VerbPractice) {
	writeJSONVerb(w, service.PublicVerbPractice(state))
}
