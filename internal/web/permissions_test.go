package web

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestAllPermissions(t *testing.T) {
	perms := AllPermissions()
	if len(perms) == 0 {
		t.Error("AllPermissions() should return at least one permission")
	}

	expectedPerms := []Permission{
		PermissionFullAccess,
		PermissionWordsReadAll,
		PermissionWordsEditAll,
		PermissionWordSetsRead,
		PermissionWordSetsEdit,
		PermissionUsersReadAll,
	}

	for _, expected := range expectedPerms {
		found := false
		for _, perm := range perms {
			if perm == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Permission %s not found in AllPermissions()", expected)
		}
	}
}

func TestAllPermissionStrings(t *testing.T) {
	perms := AllPermissionStrings()
	if len(perms) == 0 {
		t.Error("AllPermissionStrings() should return at least one permission")
	}

	for _, perm := range perms {
		if perm == "" {
			t.Error("AllPermissionStrings() should not return empty strings")
		}
	}
}

func TestIsValidPermission(t *testing.T) {
	tests := []struct {
		name     string
		perm     string
		expected bool
	}{
		{"valid full_access", "full_access", true},
		{"valid words.read_all", "words.read_all", true},
		{"valid words.edit_all", "words.edit_all", true},
		{"valid word_sets.read", "word_sets.read", true},
		{"valid word_sets.edit", "word_sets.edit", true},
		{"valid users.read_all", "users.read_all", true},
		{"invalid permission", "invalid.permission", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPermission(tt.perm)
			if result != tt.expected {
				t.Errorf("IsValidPermission(%q) = %v, want %v", tt.perm, result, tt.expected)
			}
		})
	}
}

func TestRouter_IsSuperAdmin(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db, logger)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create admin user
	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create regular user
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	tests := []struct {
		name     string
		userID   int64
		expected bool
	}{
		{"super admin", adminUser.ID, true},
		{"regular user", regularUser.ID, false},
		{"non-existent user", 999999, false},
		{"zero user ID", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), userIDKey, tt.userID)
			result := router.IsSuperAdmin(ctx)
			if result != tt.expected {
				t.Errorf("IsSuperAdmin(userID=%d) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}

	// IsSuperAdmin with nil userRepo (router created without SetDependencies)
	routerNoDeps := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	ctx := context.WithValue(context.Background(), userIDKey, adminUser.ID)
	if routerNoDeps.IsSuperAdmin(ctx) {
		t.Error("IsSuperAdmin with nil userRepo should return false")
	}

	// IsSuperAdmin when userRepo is not *repository.UserRepository (type assert fails)
	router.userRepo = 123
	if router.IsSuperAdmin(ctx) {
		t.Error("IsSuperAdmin with non-*UserRepository userRepo should return false")
	}
}

func TestRouter_HasPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create admin user
	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create regular user
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Create category with permissions
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	if err != nil {
		t.Fatalf("Failed to set category permissions: %v", err)
	}

	// Assign category to regular user (words.read_all only)
	err = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to assign category to user: %v", err)
	}

	fullAccessCategoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "Full Access"})
	if err != nil {
		t.Fatalf("Failed to create full access category: %v", err)
	}
	err = accessCategoryRepo.SetCategoryPermissions(fullAccessCategoryID, []string{string(PermissionFullAccess)})
	if err != nil {
		t.Fatalf("Failed to set full access permissions: %v", err)
	}

	// User that has both read-only and full_access (for "full_access has any permission" case)
	fullAccessUser, err := userRepo.GetOrCreateUser(88888)
	if err != nil {
		t.Fatalf("Failed to create full_access user: %v", err)
	}
	err = accessCategoryRepo.SetUserCategories(fullAccessUser.ID, []int64{categoryID, fullAccessCategoryID})
	if err != nil {
		t.Fatalf("Failed to assign categories to full_access user: %v", err)
	}

	// User with no categories in DB
	noCatUser, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("Failed to create no-categories user: %v", err)
	}

	tests := []struct {
		name        string
		userID      int64
		categories  []int64
		cachedPerms []string // if non-nil, pre-fill context with these permissions (covers getUserPermissionsFromDB cache path)
		permission  Permission
		expected    bool
	}{
		{"super admin has all permissions", adminUser.ID, []int64{}, nil, PermissionWordsReadAll, true},
		{"user with permission", regularUser.ID, []int64{categoryID}, nil, PermissionWordsReadAll, true},
		{"user without permission", regularUser.ID, []int64{categoryID}, nil, PermissionWordsEditAll, false},
		{"user with no categories", noCatUser.ID, []int64{}, nil, PermissionWordsReadAll, false},
		{"user with full_access has any permission", fullAccessUser.ID, []int64{fullAccessCategoryID}, nil, PermissionWordsEditAll, true},
		{"user with cached permissions", regularUser.ID, []int64{}, []string{"words.read_all"}, PermissionWordsReadAll, true},
		{"user with cached full_access", regularUser.ID, []int64{}, []string{string(PermissionFullAccess)}, PermissionWordsEditAll, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), userIDKey, tt.userID)
			ctx = context.WithValue(ctx, userCategoriesKey, tt.categories)
			if tt.cachedPerms != nil {
				ctx = context.WithValue(ctx, userPermissionsKey, tt.cachedPerms)
			}
			result := router.HasPermission(ctx, tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission(userID=%d, permission=%s) = %v, want %v", tt.userID, tt.permission, result, tt.expected)
			}
		})
	}
}

// TestRouter_HasPermission_GetUserPermissionsError covers HasPermission when getUserPermissionsFromDB returns error.
func TestRouter_HasPermission_GetUserPermissionsError(t *testing.T) {
	testutil.SetupTestDB(t) // ensure postgres_compat driver is registered
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDatabase(t)
	conn := goodDB.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(conn, logger)

	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 12345},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret"},
	}

	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "PermCat"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	_ = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	_ = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})

	closedConn, err := sql.Open("postgres_compat", testutil.GetTestDSN(t))
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	if err := closedConn.Ping(); err != nil {
		_ = closedConn.Close()
		t.Skip("ping failed:", err)
	}
	closedConn.Close()

	router := NewRouter(logger, cfg, closedConn, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	ctx := context.WithValue(context.Background(), userIDKey, regularUser.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{categoryID})
	if router.HasPermission(ctx, PermissionWordsReadAll) {
		t.Error("HasPermission with failing GetUserPermissions should return false")
	}
}

func TestRouter_RequirePermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "ReadCategory"})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	if err != nil {
		t.Fatalf("Failed to set category permissions: %v", err)
	}
	err = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to assign category to user: %v", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	tests := []struct {
		name         string
		userID       int64
		categories   []int64
		cachedPerms  []string // if set, pre-fill context (covers loadUserPermissionsIntoContext cache path)
		expected     int
	}{
		{"super admin allowed", adminUser.ID, []int64{}, nil, http.StatusOK},
		{"user without permission", regularUser.ID, []int64{}, nil, http.StatusForbidden},
		{"user with permission from DB", regularUser.ID, []int64{categoryID}, nil, http.StatusOK},
		{"user with cached permissions allowed", regularUser.ID, []int64{}, []string{"words.read_all"}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			ctx = context.WithValue(ctx, userCategoriesKey, tt.categories)
			if tt.cachedPerms != nil {
				ctx = context.WithValue(ctx, userPermissionsKey, tt.cachedPerms)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			protectedHandler := router.RequirePermission(PermissionWordsReadAll)(handler)
			protectedHandler(rr, req)

			if rr.Code != tt.expected {
				t.Errorf("RequirePermission() status = %d, want %d", rr.Code, tt.expected)
			}
		})
	}
}

func TestRouter_checkPermissionInHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{
		Admin:  config.AdminConfig{TelegramID: 12345},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret"},
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	adminUser, _ := userRepo.GetOrCreateUser(12345)
	regularUser, _ := userRepo.GetOrCreateUser(99999)

	categoryID, _ := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "PermCat"})
	_ = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	_ = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})

	tests := []struct {
		name       string
		ctx        context.Context
		permission Permission
		want       bool
	}{
		{"super admin has permission", context.WithValue(context.Background(), userIDKey, adminUser.ID), PermissionWordsReadAll, true},
		{"user with permission", context.WithValue(context.WithValue(context.Background(), userIDKey, regularUser.ID), userCategoriesKey, []int64{categoryID}), PermissionWordsReadAll, true},
		{"user without permission", context.WithValue(context.WithValue(context.Background(), userIDKey, regularUser.ID), userCategoriesKey, []int64{categoryID}), PermissionWordsEditAll, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := router.checkPermissionInHandler(tt.ctx, tt.permission)
			if got != tt.want {
				t.Errorf("checkPermissionInHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_RequireAnyPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "Any Perms"})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	if err != nil {
		t.Fatalf("Failed to set category permissions: %v", err)
	}
	err = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to assign category to user: %v", err)
	}

	editOnlyCategoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "Edit Only"})
	if err != nil {
		t.Fatalf("Failed to create edit-only category: %v", err)
	}
	err = accessCategoryRepo.SetCategoryPermissions(editOnlyCategoryID, []string{"words.edit_all"})
	if err != nil {
		t.Fatalf("Failed to set edit-only permissions: %v", err)
	}
	editOnlyUser, err := userRepo.GetOrCreateUser(66666)
	if err != nil {
		t.Fatalf("Failed to create edit-only user: %v", err)
	}
	err = accessCategoryRepo.SetUserCategories(editOnlyUser.ID, []int64{editOnlyCategoryID})
	if err != nil {
		t.Fatalf("Failed to assign edit-only category to user: %v", err)
	}

	fullAccessCategoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{Name: "Full Access"})
	if err != nil {
		t.Fatalf("Failed to create full access category: %v", err)
	}
	err = accessCategoryRepo.SetCategoryPermissions(fullAccessCategoryID, []string{string(PermissionFullAccess)})
	if err != nil {
		t.Fatalf("Failed to set full access permissions: %v", err)
	}
	fullAccessOnlyUser, err := userRepo.GetOrCreateUser(55555)
	if err != nil {
		t.Fatalf("Failed to create full-access-only user: %v", err)
	}
	err = accessCategoryRepo.SetUserCategories(fullAccessOnlyUser.ID, []int64{fullAccessCategoryID})
	if err != nil {
		t.Fatalf("Failed to assign full_access category to user: %v", err)
	}

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	tests := []struct {
		name       string
		userID     int64
		categories []int64
		wantStatus int
	}{
		{"super admin allowed", adminUser.ID, []int64{}, http.StatusOK},
		{"user with any permission allowed", regularUser.ID, []int64{categoryID}, http.StatusOK},
		{"user with second required permission allowed", editOnlyUser.ID, []int64{editOnlyCategoryID}, http.StatusOK},
		{"user with full access allowed", fullAccessOnlyUser.ID, []int64{fullAccessCategoryID}, http.StatusOK},
		{"user without permissions forbidden", regularUser.ID, []int64{}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			ctx = context.WithValue(ctx, userCategoriesKey, tt.categories)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			router.RequireAnyPermission(PermissionWordsReadAll, PermissionWordsEditAll)(handler)(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
