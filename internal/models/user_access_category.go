package models

import "time"

// UserAccessCategory represents a user access category
type UserAccessCategory struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserAccessCategoryWithPermissions represents a category with its permissions
type UserAccessCategoryWithPermissions struct {
	UserAccessCategory
	Permissions []string `json:"permissions"`
}
