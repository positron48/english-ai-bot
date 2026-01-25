package web

// Permission represents a permission string
type Permission string

// Available permissions (hardcoded list)
const (
	// FullAccess grants all permissions
	PermissionFullAccess Permission = "full_access"

	// Words permissions
	PermissionWordsReadAll Permission = "words.read_all"
	PermissionWordsEditAll Permission = "words.edit_all"

	// Word sets permissions
	PermissionWordSetsRead Permission = "word_sets.read"
	PermissionWordSetsEdit Permission = "word_sets.edit"

	// Users permissions
	PermissionUsersReadAll Permission = "users.read_all"

	// Stats permissions
	PermissionStatsRead Permission = "stats.read"
)

// AllPermissions returns all available permissions
func AllPermissions() []Permission {
	return []Permission{
		PermissionFullAccess,
		PermissionWordsReadAll,
		PermissionWordsEditAll,
		PermissionWordSetsRead,
		PermissionWordSetsEdit,
		PermissionUsersReadAll,
		PermissionStatsRead,
	}
}

// AllPermissionStrings returns all available permissions as strings
func AllPermissionStrings() []string {
	perms := AllPermissions()
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = string(p)
	}
	return result
}

// IsValidPermission checks if a permission string is valid
func IsValidPermission(perm string) bool {
	for _, p := range AllPermissions() {
		if string(p) == perm {
			return true
		}
	}
	return false
}
