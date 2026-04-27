package repository

import (
	"testing"

	"tgbot-skeleton/internal/grammarbundle"

	"go.uber.org/zap"
)

func TestBuildTheoryBlockIndex_EmbeddedBundle(t *testing.T) {
	logger := zap.NewNop()
	repo := NewGrammarContentRepositoryWithFS(grammarbundle.FS, logger)

	idx, err := BuildTheoryBlockIndex(repo)
	if err != nil {
		t.Fatalf("BuildTheoryBlockIndex error: %v", err)
	}
	if idx == nil {
		t.Fatalf("index is nil")
	}
	if len(idx.ByBlockID) == 0 {
		t.Fatalf("expected non-empty theory blocks index")
	}

	foundConcept := false
	for _, info := range idx.ByBlockID {
		if info.ConceptID != "" {
			foundConcept = true
			break
		}
	}
	if !foundConcept {
		t.Fatalf("expected at least one theory block with concept_id")
	}
}

