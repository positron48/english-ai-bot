package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func createNonAdminUser(t *testing.T, dbRepo *repository.UserRepository, telegramID int64) int64 {
	t.Helper()
	user, err := dbRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}
	return user.ID
}

func decodeJSONMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func TestHandleAdminWordSetCategories_CRUDAndPermissions(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	nonAdminUserID := createNonAdminUser(t, userRepo, 987654321)

	t.Run("post forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/word-set-categories", bytes.NewBufferString(`{"name":"forbidden"}`))
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	postReq := httptest.NewRequest(http.MethodPost, "/api/admin/word-set-categories", bytes.NewBufferString(`{"name":"verbs","sort_order":1}`))
	postReq = setUserIDInContextWordSets(postReq, adminUserID)
	postRR := httptest.NewRecorder()
	router.handleAdminWordSetCategories(postRR, postReq)
	if postRR.Code != http.StatusOK {
		t.Fatalf("create category: expected 200, got %d (%s)", postRR.Code, postRR.Body.String())
	}
	postPayload := decodeJSONMap(t, postRR)
	idFloat, ok := postPayload["id"].(float64)
	if !ok {
		t.Fatalf("expected id in response, got %v", postPayload["id"])
	}
	categoryID := int64(idFloat)

	t.Run("put forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBufferString(`{"name":"x"}`))
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("put invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBufferString(`{invalid json`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("put invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-set-categories/bad", bytes.NewBufferString(`{"name":"x"}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("put success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBufferString(`{"name":"verbs-updated","sort_order":2}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})

	categoryRepo := repository.NewWordSetCategoryRepository(router.db, router.logger)
	childID, err := categoryRepo.CreateCategory(&models.WordSetCategory{
		ParentID:  &categoryID,
		Name:      "child",
		SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("CreateCategory child error: %v", err)
	}

	t.Run("delete forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), nil)
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-set-categories/xyz", nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("delete with children returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	if err := categoryRepo.DeleteCategory(childID); err != nil {
		t.Fatalf("DeleteCategory child error: %v", err)
	}

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-set-categories/"+strconv.FormatInt(categoryID, 10), nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetCategories(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})
}

func TestHandleAdminWordSets_CRUDAndItemsUpdate(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	nonAdminUserID := createNonAdminUser(t, userRepo, 222333444)

	t.Run("post forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets", bytes.NewBufferString(`{"title":"forbidden"}`))
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("title required", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets", bytes.NewBufferString(`{"title":""}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	postReq := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets", bytes.NewBufferString(`{"title":"starter set","sort_order":1}`))
	postReq = setUserIDInContextWordSets(postReq, adminUserID)
	postRR := httptest.NewRecorder()
	router.handleAdminWordSets(postRR, postReq)
	if postRR.Code != http.StatusOK {
		t.Fatalf("create set: expected 200, got %d (%s)", postRR.Code, postRR.Body.String())
	}
	postPayload := decodeJSONMap(t, postRR)
	idFloat, ok := postPayload["id"].(float64)
	if !ok {
		t.Fatalf("expected id in response, got %v", postPayload["id"])
	}
	setID := int64(idFloat)

	t.Run("put forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10), bytes.NewBufferString(`{"title":"x"}`))
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("put invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/bad", bytes.NewBufferString(`{"title":"x"}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("put success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10), bytes.NewBufferString(`{"title":"starter set updated","sort_order":2}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})

	t.Run("put items success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10)+"/items", bytes.NewBufferString(`{"words":"  "}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})

	t.Run("put items invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10)+"/items", bytes.NewBufferString(`{not json`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("delete forbidden for non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10), nil)
		req = setUserIDInContextWordSets(req, nonAdminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-sets/bad", nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/word-sets/"+strconv.FormatInt(setID, 10), nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})
}

func TestHandleAdminWordSetDetailOrSets_RoutesToDetailAndMutationHandlers(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	// Create a set
	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets", bytes.NewBufferString(`{"title":"detail set"}`))
	createReq = setUserIDInContextWordSets(createReq, adminUserID)
	createRR := httptest.NewRecorder()
	router.handleAdminWordSets(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create set: expected 200, got %d (%s)", createRR.Code, createRR.Body.String())
	}
	payload := decodeJSONMap(t, createRR)
	setID := int64(payload["id"].(float64))
	setIDStr := strconv.FormatInt(setID, 10)

	t.Run("get detail success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets/"+setIDStr, nil)
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetDetailOrSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
		resp := decodeJSONMap(t, rr)
		if _, ok := resp["word_set"]; !ok {
			t.Fatal("expected word_set in detail response")
		}
	})

	t.Run("put routed to handleAdminWordSets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+setIDStr, bytes.NewBufferString(`{"title":"detail set updated"}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetDetailOrSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})

	t.Run("put items routed to handleAdminWordSets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/admin/word-sets/"+setIDStr+"/items", bytes.NewBufferString(`{"words":""}`))
		req = setUserIDInContextWordSets(req, adminUserID)
		rr := httptest.NewRecorder()
		router.handleAdminWordSetDetailOrSets(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
		}
	})
}
