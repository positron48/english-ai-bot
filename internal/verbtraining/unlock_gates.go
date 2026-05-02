package verbtraining

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

const UnlockGatesPath = "verb_forms/unlock-gates.json"

func LoadUnlockGates(fsys fs.FS) (*UnlockGates, error) {
	if fsys == nil {
		return nil, fmt.Errorf("nil filesystem")
	}
	raw, err := fs.ReadFile(fsys, UnlockGatesPath)
	if err != nil {
		return nil, fmt.Errorf("read unlock gates: %w", err)
	}
	var g UnlockGates
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, fmt.Errorf("parse unlock gates: %w", err)
	}
	if g.Chapters == nil {
		g.Chapters = map[string][]string{}
	}
	// normalize in place
	for i := range g.AlwaysUnlocked {
		g.AlwaysUnlocked[i] = strings.ToLower(strings.TrimSpace(g.AlwaysUnlocked[i]))
	}
	normalized := make(map[string][]string, len(g.Chapters))
	for chapter, scopes := range g.Chapters {
		ch := strings.TrimSpace(chapter)
		if ch == "" {
			continue
		}
		dst := make([]string, 0, len(scopes))
		for _, s := range scopes {
			ss := strings.ToLower(strings.TrimSpace(s))
			if ss == "" {
				continue
			}
			dst = append(dst, ss)
		}
		normalized[ch] = dst
	}
	g.Chapters = normalized
	return &g, nil
}

func (g *UnlockGates) EnabledScopes(allowedChapters map[string]bool) []string {
	if g == nil {
		return nil
	}
	set := map[string]struct{}{}
	for _, s := range g.AlwaysUnlocked {
		if strings.TrimSpace(s) != "" {
			set[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
		}
	}
	for chapter, scopes := range g.Chapters {
		if !allowedChapters[chapter] {
			continue
		}
		for _, s := range scopes {
			ss := strings.ToLower(strings.TrimSpace(s))
			if ss != "" {
				set[ss] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}

