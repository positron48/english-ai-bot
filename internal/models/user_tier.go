package models

import "strings"

// UserTier is the subscription tier for end-user features (separate from admin RBAC).
type UserTier string

const (
	TierFree    UserTier = "free"
	TierPro     UserTier = "pro"
	TierProPlus UserTier = "pro_plus"
)

// AllUserTiers returns known tiers in rank order (for admin UI).
func AllUserTiers() []UserTier {
	return []UserTier{TierFree, TierPro, TierProPlus}
}

// ParseUserTier normalizes and validates a tier string; unknown values become TierFree.
func ParseUserTier(s string) UserTier {
	t := UserTier(strings.TrimSpace(strings.ToLower(s)))
	switch t {
	case TierPro, TierProPlus:
		return t
	default:
		return TierFree
	}
}

// IsValidUserTier reports whether s is a known tier constant.
func IsValidUserTier(s string) bool {
	t := UserTier(strings.TrimSpace(strings.ToLower(s)))
	for _, known := range AllUserTiers() {
		if t == known {
			return true
		}
	}
	return false
}

// Rank returns ordering for tier comparison (higher = more access).
func (t UserTier) Rank() int {
	switch ParseUserTier(string(t)) {
	case TierProPlus:
		return 2
	case TierPro:
		return 1
	default:
		return 0
	}
}

// AtLeast returns true if t has rank >= other.
func (t UserTier) AtLeast(other UserTier) bool {
	return t.Rank() >= other.Rank()
}

// TierAllowsFeature checks feature access by tier.
func TierAllowsFeature(t UserTier, feature string) bool {
	switch strings.TrimSpace(strings.ToLower(feature)) {
	case "speaking":
		return ParseUserTier(string(t)).AtLeast(TierPro)
	case "conversation":
		return ParseUserTier(string(t)).AtLeast(TierPro)
	case "speaking_roleplay":
		return ParseUserTier(string(t)).AtLeast(TierProPlus)
	default:
		return false
	}
}

// UserFeaturesForTier returns feature flags for API responses.
func UserFeaturesForTier(t UserTier) map[string]bool {
	t = ParseUserTier(string(t))
	return map[string]bool{
		"speaking":          TierAllowsFeature(t, "speaking"),
		"conversation":      TierAllowsFeature(t, "conversation"),
		"speaking_roleplay": TierAllowsFeature(t, "speaking_roleplay"),
	}
}
