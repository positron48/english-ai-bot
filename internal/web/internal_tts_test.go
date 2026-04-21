package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

type internalTTSMock struct {
	mockPronunciationService
	pending []service.ExternalPendingWord
}

func (m *internalTTSMock) ListPendingExternal(limit int) ([]service.ExternalPendingWord, error) {
	return m.pending, nil
}

func (m *internalTTSMock) StoreExternalAudio(word, provider, format string, audio []byte) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{Word: strings.ToLower(strings.TrimSpace(word)), State: "ready"}, nil
}

func (m *internalTTSMock) MarkExternalFailure(word, provider, state, errorCode, errorMessage string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{Word: strings.ToLower(strings.TrimSpace(word)), State: state}, nil
}

func TestInternalTTSPending_AuthRequired(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.LearningConfig{TargetLang: "en"},
		TTS:      config.TTSConfig{InternalEnabled: true},
	}, nil, nil, nil, nil, nil)
	r.pronunciationService = &internalTTSMock{
		mockPronunciationService: mockPronunciationService{enabled: true},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/internal/tts/pending?limit=5", nil)
	w := httptest.NewRecorder()
	r.handleInternalTTSPending(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestInternalTTSPending_OK(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.LearningConfig{TargetLang: "en"},
		TTS:      config.TTSConfig{InternalEnabled: true, InternalMaxPendingLimit: 100},
	}, nil, nil, nil, nil, nil)
	r.internalTTSTokens = map[string]string{"en": "tok"}
	r.pronunciationService = &internalTTSMock{
		mockPronunciationService: mockPronunciationService{enabled: true},
		pending:                  []service.ExternalPendingWord{{Word: "hello", TargetLang: "en"}},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/internal/tts/pending?limit=5", nil)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	r.handleInternalTTSPending(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if int(body["count"].(float64)) != 1 {
		t.Fatalf("unexpected count: %v", body["count"])
	}
}

func TestInternalTTSAudio_Base64(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.LearningConfig{TargetLang: "es"},
		TTS:      config.TTSConfig{InternalEnabled: true, InternalMaxUploadMB: 1},
	}, nil, nil, nil, nil, nil)
	r.internalTTSTokens = map[string]string{"es": "tok-es"}
	r.pronunciationService = &internalTTSMock{
		mockPronunciationService: mockPronunciationService{enabled: true},
	}
	// Minimal valid mp3 header bytes.
	mp3 := []byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00}
	payload := `{"word":"hola","format":"mp3","audio_base64":"` + base64.StdEncoding.EncodeToString(mp3) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/internal/tts/audio", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok-es")
	w := httptest.NewRecorder()
	r.handleInternalTTSAudio(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}
