//go:build integration

package testkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHarness_Smoke(t *testing.T) {
	h := NewHarness(t)
	token := h.AuthAsUser(12345)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("dashboard: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
