package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func (r *Router) handleInternalTrainingCard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/internal/training/card/")
	cardID, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
	if err != nil || cardID <= 0 {
		http.Error(w, "Invalid card id", http.StatusBadRequest)
		return
	}
	var updateData map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	if err := r.applyTrainingCardUpdate(cardID, updateData); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Training card not found", http.StatusNotFound)
			return
		}
		r.logger.Error("internal training card update failed", zap.Error(err), zap.Int64("card_id", cardID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "card_id": cardID})
}

func (r *Router) applyTrainingCardUpdate(cardID int64, updateData map[string]interface{}) error {
	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	card, err := trainingCardRepo.GetTrainingCard(cardID)
	if err != nil {
		return err
	}
	if card == nil {
		return fmt.Errorf("training card not found")
	}
	stringField := func(key string) string {
		if val, ok := updateData[key].(string); ok {
			return val
		}
		return ""
	}
	wordEN := stringField("word_en")
	wordRU := stringField("word_ru")
	meaningEN := stringField("meaning_en")
	exampleEN := stringField("example_en")
	exampleRU := stringField("example_ru")
	transcription := stringField("transcription")
	distractorsRU := stringField("distractors_ru")
	distractorsEN := stringField("distractors_en")
	hint := stringField("hint")
	pos := stringField("pos")
	displayWord := stringField("display_word")
	posProvided := false
	if _, ok := updateData["pos"]; ok {
		posProvided = true
	}
	if wordEN != "" {
		card.WordEN = wordEN
	}
	if posProvided {
		if pos != "" {
			card.POS = &pos
		} else {
			card.POS = nil
		}
	}
	if displayWord != "" {
		card.DisplayWord = &displayWord
	} else if _, ok := updateData["display_word"]; ok {
		card.DisplayWord = nil
	}
	if _, ok := updateData["word_ru"]; ok {
		card.WordRU = wordRU
	}
	if _, ok := updateData["meaning_en"]; ok {
		card.MeaningEN = meaningEN
	}
	if _, ok := updateData["example_en"]; ok {
		card.ExampleEN = exampleEN
	}
	if _, ok := updateData["example_ru"]; ok {
		card.ExampleRU = exampleRU
	}
	if _, ok := updateData["transcription"]; ok {
		card.Transcription = transcription
	}
	if _, ok := updateData["distractors_ru"]; ok {
		card.DistractorsRU = distractorsRU
	}
	if _, ok := updateData["distractors_en"]; ok {
		card.DistractorsEN = distractorsEN
	}
	if _, ok := updateData["hint"]; ok {
		card.Hint = hint
	}
	if err := trainingCardRepo.UpdateTrainingCard(card); err != nil {
		return err
	}
	if r.pronunciationService != nil {
		wordRepo := repository.NewWordRepository(r.db, r.logger)
		canonicalWord := card.WordEN
		if wc, err := wordRepo.GetWordCardByID(card.WordCardID); err == nil && wc != nil {
			canonicalWord = wc.Word
		}
		r.pronunciationService.ScheduleWord(canonicalWord)
	}
	return nil
}

type internalTTSRegenerateRequest struct {
	Word string `json:"word"`
}

func (r *Router) handleInternalTTSRegenerate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.pronunciationService == nil || !r.pronunciationService.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}
	var body internalTTSRegenerateRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	word := strings.TrimSpace(body.Word)
	if word == "" {
		http.Error(w, "word required", http.StatusBadRequest)
		return
	}
	status, err := r.pronunciationService.ForceRegenerate(word)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeTTSStatusResponse(w, status)
}

func (r *Router) handleInternalTTSStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.pronunciationService == nil || !r.pronunciationService.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}
	word := strings.TrimSpace(req.URL.Query().Get("word"))
	if word == "" {
		http.Error(w, "word query param required", http.StatusBadRequest)
		return
	}
	status, err := r.pronunciationService.GetStatus(word)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeTTSStatusResponse(w, status)
}
