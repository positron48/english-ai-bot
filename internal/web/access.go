package web

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// handleAccessMe returns current user's categories and effective permissions
// @Summary      Получить мои права доступа
// @Description  Возвращает список категорий пользователя и эффективные права доступа (вычисленные из категорий)
// @Tags         Access
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Категории и права пользователя"
// @Failure      401  {string}  string  "Неавторизован"
// @Router       /api/access/me [get]
func (r *Router) handleAccessMe(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := req.Context()
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get categories from context (from JWT)
	categories := getUserCategoriesFromContext(ctx)

	// Calculate effective permissions
	permissions, err := r.getUserPermissionsFromDB(ctx)
	if err != nil {
		r.logger.Error("failed to get user permissions", zap.Error(err))
		permissions = []string{}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories":  categories,
		"permissions": permissions,
	})
}
