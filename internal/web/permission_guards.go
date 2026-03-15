package web

import (
	"context"
	"encoding/json"
	"net/http"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// IsSuperAdmin checks if the current user is the super admin from .env
func (r *Router) IsSuperAdmin(ctx context.Context) bool {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return false
	}

	// Get user to check telegram ID
	if r.userRepo == nil {
		return false
	}

	userRepo, ok := r.userRepo.(*repository.UserRepository)
	if !ok {
		return false
	}

	user, err := userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return false
	}

	// Check if telegram ID matches admin config
	return int64(r.config.Admin.TelegramID) == user.TelegramID
}

// getUserPermissionsFromDB calculates user permissions from their categories
// Caches result in context to avoid multiple DB queries per request
func (r *Router) getUserPermissionsFromDB(ctx context.Context) ([]string, error) {
	// Check cache in context first
	if cached := getUserPermissionsFromContext(ctx); len(cached) > 0 {
		return cached, nil
	}

	// Super admin has all permissions
	if r.IsSuperAdmin(ctx) {
		allPerms := AllPermissionStrings()
		return allPerms, nil
	}

	// Get categories from context
	categories := getUserCategoriesFromContext(ctx)
	if len(categories) == 0 {
		// No categories, no permissions
		return []string{}, nil
	}

	// Get permissions from database
	permissions, err := r.accessCategoryRepo.GetUserPermissions(getUserIDFromContext(ctx))
	if err != nil {
		r.logger.Error("failed to get user permissions", zap.Error(err))
		return []string{}, err
	}

	return permissions, nil
}

// loadUserPermissionsIntoContext loads permissions into context (call this once per request)
func (r *Router) loadUserPermissionsIntoContext(ctx context.Context) context.Context {
	// Don't reload if already cached
	if cached := getUserPermissionsFromContext(ctx); len(cached) > 0 {
		return ctx
	}

	perms, _ := r.getUserPermissionsFromDB(ctx)
	return context.WithValue(ctx, userPermissionsKey, perms)
}

// HasPermission checks if user has a specific permission
func (r *Router) HasPermission(ctx context.Context, permission Permission) bool {
	// Super admin always has all permissions
	if r.IsSuperAdmin(ctx) {
		return true
	}

	permissions, err := r.getUserPermissionsFromDB(ctx)
	if err != nil {
		return false
	}

	permStr := string(permission)
	for _, p := range permissions {
		if p == permStr {
			return true
		}
		// full_access grants everything
		if p == string(PermissionFullAccess) {
			return true
		}
	}

	return false
}

// RequirePermission wraps a handler to require a specific permission
func (r *Router) RequirePermission(permission Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			// Load permissions into context for this request
			ctx = r.loadUserPermissionsIntoContext(ctx)
			req = req.WithContext(ctx)

			if !r.HasPermission(ctx, permission) {
				r.logger.Warn("permission denied",
					zap.String("path", req.URL.Path),
					zap.String("permission", string(permission)),
					zap.Int64("user_id", getUserIDFromContext(ctx)))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Forbidden",
					"message": "You don't have permission to access this resource.",
				})
				return
			}

			next(w, req)
		}
	}
}

// checkPermissionInHandler checks permission in handler context (for method-specific checks)
func (r *Router) checkPermissionInHandler(ctx context.Context, permission Permission) bool {
	return r.HasPermission(ctx, permission)
}

// RequireAnyPermission wraps a handler to require at least one of the specified permissions
func (r *Router) RequireAnyPermission(permissions ...Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			// Load permissions into context for this request
			ctx = r.loadUserPermissionsIntoContext(ctx)
			// Update request with new context
			req = req.WithContext(ctx)

			// Super admin always has access
			if r.IsSuperAdmin(ctx) {
				next(w, req)
				return
			}

			// Get permissions from context (already loaded by loadUserPermissionsIntoContext)
			userPerms := getUserPermissionsFromContext(ctx)

			// Check if user has full_access
			for _, p := range userPerms {
				if p == string(PermissionFullAccess) {
					next(w, req)
					return
				}
			}

			// Check if user has any of the required permissions
			for _, requiredPerm := range permissions {
				requiredPermStr := string(requiredPerm)
				for _, userPerm := range userPerms {
					if userPerm == requiredPermStr {
						next(w, req)
						return
					}
				}
			}

			r.logger.Warn("permission denied - none of required permissions",
				zap.String("path", req.URL.Path),
				zap.Int64("user_id", getUserIDFromContext(ctx)))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to access this resource.",
			})
		}
	}
}
