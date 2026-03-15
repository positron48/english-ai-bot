package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setUserIDInContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func setupOrphanedTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {} // shared db, do not close

	return router, db, cleanup
}

func TestHandleAdminOrphanedCards_Get(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-cards", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCards(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminOrphanedCards_GetWithPagination(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-cards?limit=10&offset=5", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCards(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminOrphanedCards_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/orphaned-cards", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCards(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedCard_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-cards/123", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCard(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedCard_InvalidID(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-cards/invalid", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedCard_NotFound(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-cards/99999", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedUserCards_Get(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-user-cards", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCards(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminOrphanedUserCards_GetWithPagination(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-user-cards?limit=20&offset=10", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCards(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminOrphanedUserCards_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/orphaned-user-cards", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCards(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedUserCard_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-user-cards/123", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCard(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedUserCard_InvalidID(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-user-cards/invalid", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedUserCard_NotFound(t *testing.T) {
	router, _, cleanup := setupOrphanedTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-user-cards/99999", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestHandleAdminOrphanedCard_DeleteSuccess(t *testing.T) {
	router, db, cleanup := setupOrphanedTest(t)
	defer cleanup()

	conn := db.GetConnection()
	_, err := conn.Exec(`INSERT INTO word_cards (id, word, definition) VALUES (9001, 'orphanword', 'def') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("Setup word_cards: %v", err)
	}
	_, err = conn.Exec(`INSERT INTO training_cards (id, word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (9001, 9001, 'orphanword', 0, 'слово', 'meaning') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("Setup training_cards: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-cards/9001", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminOrphanedUserCard_DeleteSuccess(t *testing.T) {
	router, db, cleanup := setupOrphanedTest(t)
	defer cleanup()

	conn := db.GetConnection()
	_, err := conn.Exec(`INSERT INTO word_cards (id, word, definition) VALUES (9002, 'ucword', 'def') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("Setup word_cards: %v", err)
	}
	_, err = conn.Exec(`INSERT INTO training_cards (id, word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (9002, 9002, 'ucword', 0, 'слово', 'meaning') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("Setup training_cards: %v", err)
	}
	userRepo := repository.NewUserRepository(conn, router.logger)
	u, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, err = conn.Exec(`INSERT INTO user_cards (id, user_id, training_card_id, direction, state, ef) VALUES (9002, $1, 9002, 'en_ru', 'new', 2.5)`, u.ID)
	if err != nil {
		t.Fatalf("Setup user_cards: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/orphaned-user-cards/9002", nil)
	req = setUserIDInContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminOrphanedUserCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
