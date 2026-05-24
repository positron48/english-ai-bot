package models

import "testing"

func TestUserTier_RankAndFeatures(t *testing.T) {
	if TierFree.Rank() >= TierPro.Rank() {
		t.Fatal("free should rank below pro")
	}
	if !TierAllowsFeature(TierPro, "speaking") {
		t.Fatal("pro should allow speaking")
	}
	if TierAllowsFeature(TierFree, "speaking") {
		t.Fatal("free should not allow speaking")
	}
	if !TierAllowsFeature(TierProPlus, "speaking_roleplay") {
		t.Fatal("pro_plus should allow roleplay")
	}
	if TierAllowsFeature(TierPro, "speaking_roleplay") {
		t.Fatal("pro should not allow roleplay")
	}
}

func TestParseUserTier(t *testing.T) {
	if ParseUserTier(" PRO ") != TierPro {
		t.Fatalf("got %q", ParseUserTier(" PRO "))
	}
	if ParseUserTier("unknown") != TierFree {
		t.Fatalf("unknown tier should be free")
	}
}

func TestIsValidUserTier(t *testing.T) {
	if !IsValidUserTier("pro_plus") {
		t.Fatal("pro_plus should be valid")
	}
	if IsValidUserTier("enterprise") {
		t.Fatal("enterprise is not a known tier yet")
	}
}
