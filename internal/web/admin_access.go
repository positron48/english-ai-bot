package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// handleAdminAccessCategories handles CRUD for user access categories
func (r *Router) handleAdminAccessCategories(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		categories, err := r.accessCategoryRepo.GetAllCategories()
		if err != nil {
			r.logger.Error("failed to get access categories", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Get permissions for each category
		categoriesWithPerms := make([]*models.UserAccessCategoryWithPermissions, 0, len(categories))
		for _, cat := range categories {
			perms, err := r.accessCategoryRepo.GetCategoryPermissions(cat.ID)
			if err != nil {
				r.logger.Warn("failed to get category permissions", zap.Error(err), zap.Int64("category_id", cat.ID))
				perms = []string{}
			}
			categoriesWithPerms = append(categoriesWithPerms, &models.UserAccessCategoryWithPermissions{
				UserAccessCategory: *cat,
				Permissions:        perms,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": categoriesWithPerms,
		})

	case http.MethodPost:
		var category models.UserAccessCategory
		if err := json.NewDecoder(req.Body).Decode(&category); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if category.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		id, err := r.accessCategoryRepo.CreateCategory(&category)
		if err != nil {
			r.logger.Error("failed to create access category", zap.Error(err))
			errLower := strings.ToLower(err.Error())
			if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(errLower, "unique constraint") || strings.Contains(errLower, "duplicate key") {
				http.Error(w, "Category with this name already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}


// handleAdminAccessCategoryRoutes routes category sub-paths
func (r *Router) handleAdminAccessCategoryRoutes(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/access/categories/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Check for sub-paths: permissions or users
	if len(parts) >= 2 {
		switch parts[1] {
		case "permissions":
			r.handleAdminAccessCategoryPermissions(w, req, categoryID)
			return
		case "users":
			r.handleAdminAccessCategoryUsers(w, req, categoryID)
			return
		}
	}

	// Default: handle category CRUD (create a new request with category ID in path)
	// We need to call handleAdminAccessCategory, but it expects the ID in the path
	// So we'll handle it inline here
	r.handleAdminAccessCategoryByID(w, req, categoryID)
}

// handleAdminAccessCategoryByID handles individual category operations by ID
func (r *Router) handleAdminAccessCategoryByID(w http.ResponseWriter, req *http.Request, categoryID int64) {
	switch req.Method {
	case http.MethodGet:
		category, err := r.accessCategoryRepo.GetCategory(categoryID)
		if err != nil {
			r.logger.Error("failed to get category", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if category == nil {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}

		perms, err := r.accessCategoryRepo.GetCategoryPermissions(categoryID)
		if err != nil {
			r.logger.Warn("failed to get category permissions", zap.Error(err))
			perms = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"category": &models.UserAccessCategoryWithPermissions{
				UserAccessCategory: *category,
				Permissions:        perms,
			},
		})

	case http.MethodPut:
		var category models.UserAccessCategory
		if err := json.NewDecoder(req.Body).Decode(&category); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		category.ID = categoryID
		if err := r.accessCategoryRepo.UpdateCategory(&category); err != nil {
			r.logger.Error("failed to update category", zap.Error(err))
			if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Category with this name already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	case http.MethodDelete:
		if err := r.accessCategoryRepo.DeleteCategory(categoryID); err != nil {
			r.logger.Error("failed to delete category", zap.Error(err))
			if strings.Contains(err.Error(), "cannot delete") {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminAccessCategoryPermissions handles category permissions
func (r *Router) handleAdminAccessCategoryPermissions(w http.ResponseWriter, req *http.Request, categoryID int64) {

	switch req.Method {
	case http.MethodGet:
		perms, err := r.accessCategoryRepo.GetCategoryPermissions(categoryID)
		if err != nil {
			r.logger.Error("failed to get category permissions", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"permissions": perms,
		})

	case http.MethodPut:
		var requestData struct {
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate permissions
		for _, perm := range requestData.Permissions {
			if !IsValidPermission(perm) {
				http.Error(w, "Invalid permission: "+perm, http.StatusBadRequest)
				return
			}
		}

		if err := r.accessCategoryRepo.SetCategoryPermissions(categoryID, requestData.Permissions); err != nil {
			r.logger.Error("failed to set category permissions", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminAccessAvailablePermissions returns list of all available permissions
func (r *Router) handleAdminAccessAvailablePermissions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"permissions": AllPermissionStrings(),
	})
}

// handleAdminAccessUsers handles user category assignments
func (r *Router) handleAdminAccessUsers(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/access/users/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		categories, err := r.accessCategoryRepo.GetUserCategories(userID)
		if err != nil {
			r.logger.Error("failed to get user categories", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":    userID,
			"categories": categories,
		})

	case http.MethodPut:
		var requestData struct {
			CategoryIDs []int64 `json:"category_ids"`
		}
		if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := r.accessCategoryRepo.SetUserCategories(userID, requestData.CategoryIDs); err != nil {
			r.logger.Error("failed to set user categories", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminAccessCategoryUsers handles users in a category
func (r *Router) handleAdminAccessCategoryUsers(w http.ResponseWriter, req *http.Request, categoryID int64) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDs, err := r.accessCategoryRepo.GetUsersByCategory(categoryID)
	if err != nil {
		r.logger.Error("failed to get users by category", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"category_id": categoryID,
		"user_ids":    userIDs,
	})
}
