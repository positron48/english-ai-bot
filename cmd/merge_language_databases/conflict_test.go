package main

import "testing"

func TestClassifyTelegramEntries_MultiCourseExpected(t *testing.T) {
	entries := map[int64]*identityConflict{
		42: {
			TelegramID:    42,
			SourceLabels:  []string{"english", "spanish"},
			SourceUserIDs: []string{"english:1", "spanish:9"},
		},
	}
	multi, conflicts := classifyTelegramEntries(entries)
	if len(multi) != 1 {
		t.Fatalf("expected 1 multi-course user, got %d", len(multi))
	}
	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
	if multi[0].TelegramID != 42 {
		t.Fatalf("unexpected telegram id %d", multi[0].TelegramID)
	}
}

func TestClassifyTelegramEntries_TargetConflictBlocks(t *testing.T) {
	entries := map[int64]*identityConflict{
		7: {
			TelegramID:    7,
			SourceLabels:  []string{"english"},
			SourceUserIDs: []string{"english:3"},
			TargetUserIDs: []string{"target:10", "target:11"},
		},
	}
	_, conflicts := classifyTelegramEntries(entries)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
}

func TestClassifyStableIdentityConflicts_UsernameOverlapAllowed(t *testing.T) {
	entries := map[string]*stableIdentityConflict{
		"telegram_username\x00alice": {
			IdentityType:  "telegram_username",
			IdentityValue: "alice",
			SourceLabels:  []string{"english", "spanish"},
			SourceUserIDs: []string{"english:1", "spanish:1"},
		},
	}
	conflicts := classifyStableIdentityConflicts(entries)
	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts for same telegram user ids, got %d", len(conflicts))
	}
}

func TestClassifyStableIdentityConflicts_TargetUsernameBlocks(t *testing.T) {
	entries := map[string]*stableIdentityConflict{
		"telegram_username\x00bob": {
			IdentityType:  "telegram_username",
			IdentityValue: "bob",
			SourceLabels:  []string{"english"},
			SourceUserIDs: []string{"english:1"},
			TargetUserIDs: []string{"target:9"},
		},
	}
	conflicts := classifyStableIdentityConflicts(entries)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
}
