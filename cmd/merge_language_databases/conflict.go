package main

import "sort"

// classifyTelegramEntries splits telegram presence across source DBs into expected
// multi-course overlaps (same user in English + Spanish) and blocking conflicts.
func classifyTelegramEntries(entries map[int64]*identityConflict) (multiCourse []telegramMultiCourseUser, conflicts []identityConflict) {
	for _, entry := range entries {
		sourceSet := uniqueStrings(entry.SourceLabels)
		if len(sourceSet) == 2 && containsAll(sourceSet, "english", "spanish") {
			multiCourse = append(multiCourse, telegramMultiCourseUser{
				TelegramID:    entry.TelegramID,
				SourceLabels:  sortedCopy(sourceSet),
				SourceUserIDs: sortedCopy(entry.SourceUserIDs),
			})
			continue
		}
		if len(sourceSet) > 1 || len(entry.TargetUserIDs) > 1 {
			conflicts = append(conflicts, *entry)
			continue
		}
		if len(entry.TargetUserIDs) == 1 && len(sourceSet) == 1 {
			// Source user maps to an existing target user; informational only.
			continue
		}
	}
	sort.Slice(multiCourse, func(i, j int) bool { return multiCourse[i].TelegramID < multiCourse[j].TelegramID })
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].TelegramID < conflicts[j].TelegramID })
	return multiCourse, conflicts
}

// classifyStableIdentityConflicts keeps only identities that cannot be auto-merged.
// Cross-app telegram_username overlap is resolved during user merge via telegram_id.
func classifyStableIdentityConflicts(entries map[string]*stableIdentityConflict) []stableIdentityConflict {
	var conflicts []stableIdentityConflict
	for _, entry := range entries {
		sourceSet := uniqueStrings(entry.SourceLabels)
		if entry.IdentityType == "telegram_username" && len(sourceSet) == 2 && containsAll(sourceSet, "english", "spanish") {
			continue
		}
		if entry.IdentityType == "telegram_id" && len(sourceSet) == 2 && containsAll(sourceSet, "english", "spanish") {
			continue
		}
		if len(sourceSet) > 1 || len(entry.TargetUserIDs) > 0 {
			conflicts = append(conflicts, *entry)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].IdentityType != conflicts[j].IdentityType {
			return conflicts[i].IdentityType < conflicts[j].IdentityType
		}
		return conflicts[i].IdentityValue < conflicts[j].IdentityValue
	})
	return conflicts
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func containsAll(values []string, required ...string) bool {
	seen := map[string]bool{}
	for _, v := range values {
		seen[v] = true
	}
	for _, req := range required {
		if !seen[req] {
			return false
		}
	}
	return true
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
