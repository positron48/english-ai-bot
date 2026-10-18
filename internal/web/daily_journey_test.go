package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
)

func TestDailyJourneyDisabledAndAuth(t *testing.T) {
	r := &Router{config: &config.Config{}}
	req := httptest.NewRequest("GET", "/api/linglow/daily-journey", nil)
	response := httptest.NewRecorder()
	r.handleDailyJourney(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET: %d", response.Code)
	}
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	response = httptest.NewRecorder()
	r.handleDailyJourney(response, req)
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"enabled":false`) {
		t.Fatalf("disabled GET: %d %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("personal state may be cached")
	}
	req = httptest.NewRequest("POST", "/api/linglow/daily-journeys", strings.NewReader(`{"course_code":"en_ru"}`)).WithContext(req.Context())
	response = httptest.NewRecorder()
	r.handleDailyJourneys(response, req)
	if response.Code != 503 {
		t.Fatalf("disabled start: %d", response.Code)
	}
	// The flag alone cannot expose a partially implemented itinerary.
	r.config.Linglow.DailyJourneyEnabled = true
	if r.dailyJourneyEnabled() {
		t.Fatal("flag enabled without ready subject adapters")
	}
}
func TestDailyJourneyErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{dailyjourney.ErrNotFound, 404, "daily_journey_not_found"},
		{dailyjourney.ErrAccess, 403, "daily_journey_access_denied"},
		{dailyjourney.ErrConflict, 409, "daily_journey_reload"},
		{dailyjourney.ErrActionKey, 409, "daily_journey_action_key_reused"},
		{dailyjourney.ErrInvalidAction, 422, "daily_journey_invalid_action"},
		{dailyjourney.ErrProcessing, 503, "daily_journey_processing"},
		{dailyjourney.ErrReservationLost, 409, "daily_journey_reload"},
		{fmt.Errorf("%w: private provider diagnostics", dailyjourney.ErrAIProvider), 502, "daily_journey_ai_failed"},
		{errors.New("private provider/DB diagnostics"), 500, "daily_journey_failed"},
	} {
		response := httptest.NewRecorder()
		(&Router{}).journeyError(response, httptest.NewRequest("GET", "/", nil), tc.err)
		if errors.Is(tc.err, dailyjourney.ErrProcessing) && response.Header().Get("Retry-After") != "2" {
			t.Fatal("missing provider processing retry hint")
		}
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.code) || strings.Contains(response.Body.String(), "private provider") {
			t.Fatalf("response %d %s", response.Code, response.Body.String())
		}
	}
}
func TestDailyJourneyStrictBody(t *testing.T) {
	for _, input := range []string{`{"complete":true}`, `{"client_action_id":"a"} {}`, strings.Repeat("x", 33<<10)} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(input))
		var body dailyjourney.Action
		if err := decodeJourneyBody(httptest.NewRecorder(), req, &body); err == nil {
			t.Fatalf("invalid body accepted: %.80s", input)
		}
	}
}

func TestDailyJourneyCompactOverviewOmitsExerciseData(t *testing.T) {
	j := &dailyjourney.Journey{ID: "saved", CourseCode: "es_ru", Status: "active", Steps: []dailyjourney.Step{{ID: "step", Title: "Words", Status: "active", ContentRef: json.RawMessage(`{"private":"reference"}`), State: json.RawMessage(`{"secret":"answer"}`), ResultDetails: json.RawMessage(`{"private":"result"}`)}}}
	data, err := json.Marshal(compactDailyJourney(j))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"contentRef", "contentVersion", "secret", "resultDetails", "launch", "feedback"} {
		if strings.Contains(string(data), key) {
			t.Fatalf("overview contains %s: %s", key, data)
		}
	}
	if !strings.Contains(string(data), `"id":"saved"`) || !strings.Contains(string(data), `"title":"Words"`) {
		t.Fatal(string(data))
	}
	if compactDailyJourney(nil) != nil {
		t.Fatal("missing journey must remain null")
	}
}

func TestDailyJourneyErrorsUseRequestLocale(t *testing.T) {
	for lang, fragment := range map[string]string{"en": "AI reply", "ru": "ответ ИИ", "es": "respuesta de IA"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", lang)
		w := httptest.NewRecorder()
		(&Router{}).journeyError(w, req, dailyjourney.ErrAIProvider)
		if !strings.Contains(w.Body.String(), fragment) {
			t.Fatalf("%s: %s", lang, w.Body.String())
		}
	}
}
