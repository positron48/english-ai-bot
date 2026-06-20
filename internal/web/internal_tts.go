package web

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"tgbot-skeleton/internal/models"
)

type internalTTSAudioRequest struct {
	Word         string `json:"word"`
	TargetLang   string `json:"target_lang"`
	Format       string `json:"format"`
	Provider     string `json:"provider"`
	Engine       string `json:"engine"`
	Voice        string `json:"voice"`
	SampleRateHz int    `json:"sample_rate_hz"`
	DurationMS   int    `json:"duration_ms"`
	AudioBase64  string `json:"audio_base64"`
}

type internalTTSFailRequest struct {
	Word         string `json:"word"`
	TargetLang   string `json:"target_lang"`
	State        string `json:"state"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Provider     string `json:"provider"`
	Engine       string `json:"engine"`
	Voice        string `json:"voice"`
}

func (r *Router) authorizeInternalTTS(req *http.Request) bool {
	if !r.internalTTSEnabled {
		return false
	}
	token := strings.TrimSpace(req.Header.Get("X-Service-Token"))
	if token == "" {
		return false
	}
	target := strings.ToLower(strings.TrimSpace(r.config.Learning.TargetLang))
	if target != "" {
		if expected, ok := r.internalTTSTokens[target]; ok && expected == token {
			return true
		}
	}
	if expected, ok := r.internalTTSTokens["default"]; ok && expected == token {
		return true
	}
	for _, expected := range r.internalTTSTokens {
		if expected == token {
			return true
		}
	}
	return false
}

func (r *Router) handleInternalTTSPending(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalTTS(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	targetLang := req.URL.Query().Get("target_lang")
	if targetLang == "" {
		targetLang = req.URL.Query().Get("course_code")
	}
	svc := r.pronunciationServiceForLang(targetLang)
	if svc == nil || !svc.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}

	limit := 100
	if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > r.internalTTSMaxPendingLimit {
		limit = r.internalTTSMaxPendingLimit
	}

	items, err := svc.ListPendingExternal(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"count": len(items),
		"limit": limit,
	})
}

func (r *Router) handleInternalTTSAudio(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalTTS(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	payload, audio, err := r.decodeInternalTTSAudio(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(audio) > r.internalTTSMaxUploadBytes {
		http.Error(w, "audio payload is too large", http.StatusRequestEntityTooLarge)
		return
	}

	svc := r.pronunciationServiceForLang(payload.TargetLang)
	if svc == nil || !svc.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}

	provider := strings.TrimSpace(payload.Provider)
	if provider == "" {
		provider = "external"
	}
	if strings.TrimSpace(payload.Engine) != "" {
		provider = provider + ":" + strings.TrimSpace(payload.Engine)
	}
	status, err := svc.StoreExternalAudio(payload.Word, provider, payload.Format, audio)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetLang := strings.ToLower(strings.TrimSpace(payload.TargetLang))
	if targetLang == "" {
		targetLang = strings.ToLower(strings.TrimSpace(r.config.Learning.TargetLang))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"word":        status.Word,
		"target_lang": targetLang,
		"status":      status,
	})
}

func (r *Router) handleInternalTTSFail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalTTS(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload internalTTSFailRequest
	if err := json.NewDecoder(io.LimitReader(req.Body, 64*1024)).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	svc := r.pronunciationServiceForLang(payload.TargetLang)
	if svc == nil || !svc.IsEnabled() {
		http.Error(w, "TTS service is unavailable", http.StatusServiceUnavailable)
		return
	}

	state := strings.TrimSpace(payload.State)
	if state == "" {
		state = models.TTSStateFailedRetryable
	}
	if state != models.TTSStateFailedRetryable && state != models.TTSStateFailedTerminal {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	msg := strings.TrimSpace(payload.ErrorMessage)
	if utf8.RuneCountInString(msg) > 1000 {
		msg = string([]rune(msg)[:1000])
	}
	provider := strings.TrimSpace(payload.Provider)
	if provider == "" {
		provider = "external"
	}
	if strings.TrimSpace(payload.Engine) != "" {
		provider = provider + ":" + strings.TrimSpace(payload.Engine)
	}
	status, err := svc.MarkExternalFailure(payload.Word, provider, state, strings.TrimSpace(payload.ErrorCode), msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	targetLang := strings.ToLower(strings.TrimSpace(payload.TargetLang))
	if targetLang == "" {
		targetLang = strings.ToLower(strings.TrimSpace(r.config.Learning.TargetLang))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"word":        status.Word,
		"target_lang": targetLang,
		"status":      status,
	})
}

func (r *Router) decodeInternalTTSAudio(req *http.Request) (internalTTSAudioRequest, []byte, error) {
	var payload internalTTSAudioRequest
	ct := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := req.ParseMultipartForm(int64(r.internalTTSMaxUploadBytes) + 1024); err != nil {
			return payload, nil, err
		}
		payload.Word = strings.TrimSpace(req.FormValue("word"))
		payload.TargetLang = strings.TrimSpace(req.FormValue("target_lang"))
		payload.Format = strings.TrimSpace(req.FormValue("format"))
		payload.Provider = strings.TrimSpace(req.FormValue("provider"))
		payload.Engine = strings.TrimSpace(req.FormValue("engine"))
		payload.Voice = strings.TrimSpace(req.FormValue("voice"))

		file, _, err := req.FormFile("audio")
		if err != nil {
			return payload, nil, err
		}
		defer file.Close()
		audio, err := io.ReadAll(io.LimitReader(file, int64(r.internalTTSMaxUploadBytes)+1))
		if err != nil {
			return payload, nil, err
		}
		if payload.Format == "" {
			payload.Format = "mp3"
		}
		return payload, audio, nil
	}

	if err := json.NewDecoder(io.LimitReader(req.Body, int64(r.internalTTSMaxUploadBytes)*2)).Decode(&payload); err != nil {
		return payload, nil, err
	}
	if payload.Format == "" {
		payload.Format = "mp3"
	}
	audio, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload.AudioBase64))
	if err != nil {
		audio, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(payload.AudioBase64))
		if err != nil {
			return payload, nil, err
		}
	}
	return payload, audio, nil
}
