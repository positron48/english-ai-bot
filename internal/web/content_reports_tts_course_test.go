package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

type complaintPronunciationSpy struct {
	mockPronunciationService
	regenerated []string
	checked     []string
	scheduled   []string
}

func (s *complaintPronunciationSpy) ForceRegenerate(word string) (service.TTSStatusResult, error) {
	s.regenerated = append(s.regenerated, word)
	return service.TTSStatusResult{Word: word}, nil
}
func (s *complaintPronunciationSpy) GetStatus(word string) (service.TTSStatusResult, error) {
	s.checked = append(s.checked, word)
	return service.TTSStatusResult{Word: word}, nil
}
func (s *complaintPronunciationSpy) ScheduleWord(word string) bool {
	s.scheduled = append(s.scheduled, word)
	return true
}
func TestInternalComplaintTTSUsesSelectedCourse(t *testing.T) {
	r, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()
	en := &complaintPronunciationSpy{mockPronunciationService: mockPronunciationService{enabled: true}}
	es := &complaintPronunciationSpy{mockPronunciationService: mockPronunciationService{enabled: true}}
	r.pronunciationService = en
	r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en, "es": es}
	r.internalServiceTokens = map[string]string{"default": "tok"}
	req := httptest.NewRequest(http.MethodPost, "/api/internal/tts/regenerate?course_code=es_ru", bytes.NewBufferString(`{"word":"chat"}`))
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	r.handleInternalTTSRegenerate(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("regenerate: %d %s", w.Code, w.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/api/internal/tts/status?course_code=es_ru&word=chat", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	r.handleInternalTTSStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d %s", w.Code, w.Body.String())
	}
	if len(es.regenerated) != 1 || len(es.checked) != 1 || len(en.regenerated) != 0 || len(en.checked) != 0 {
		t.Fatalf("wrong TTS service: EN=%+v ES=%+v", en, es)
	}
}

func TestInternalComplaintDetailAndUpdateUseWordCourse(t *testing.T) {
	r, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	en := &complaintPronunciationSpy{mockPronunciationService: mockPronunciationService{enabled: true}}
	es := &complaintPronunciationSpy{mockPronunciationService: mockPronunciationService{enabled: true}}
	r.pronunciationService = en
	r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en, "es": es}
	r.internalServiceTokens = map[string]string{"default": "tok"}
	_, cardID := seedWordTrainingReportCard(t, r, userID)
	var wordID int64
	if err := r.db.QueryRow("SELECT word_card_id FROM training_cards WHERE id=?", cardID).Scan(&wordID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec("UPDATE word_cards SET course_code='es_ru' WHERE id=?", wordID); err != nil {
		t.Fatal(err)
	}
	if err := r.applyTrainingCardUpdate(cardID, map[string]interface{}{"word_ru": "слово"}); err != nil {
		t.Fatal(err)
	}
	id, err := repository.NewContentReportRepository(r.db, r.logger).Create(repository.CreateContentReportInput{UserID: userID, SourceType: "word_training", Word: "reportword", WordCardID: &wordID, TrainingCardID: &cardID, ReportCategory: "bad_audio"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/"+strconv.FormatInt(id, 10), nil)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	r.handleInternalContentReportByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("detail: %d %s", w.Code, w.Body.String())
	}
	if len(es.scheduled) != 1 || len(es.checked) != 1 || len(en.scheduled) != 0 || len(en.checked) != 0 {
		t.Fatalf("wrong TTS service: EN=%+v ES=%+v", en, es)
	}
}
