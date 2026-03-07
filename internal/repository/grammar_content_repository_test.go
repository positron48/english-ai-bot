package repository

import (
	"testing"
	"testing/fstest"

	"go.uber.org/zap"
)

func setupGrammarContentRepo(t *testing.T) *GrammarContentRepository {
	t.Helper()
	return NewGrammarContentRepository(zap.NewNop())
}

func TestNewGrammarContentRepository(t *testing.T) {
	repo := NewGrammarContentRepository(zap.NewNop())
	if repo == nil {
		t.Fatal("NewGrammarContentRepository() should not return nil")
	}
}

func TestGrammarContentRepository_GetSections(t *testing.T) {
	repo := setupGrammarContentRepo(t)
	sections, err := repo.GetSections()
	if err != nil {
		t.Fatalf("GetSections() error = %v", err)
	}
	if sections == nil {
		t.Fatal("GetSections() should not return nil")
	}
	if sections.Version == "" {
		t.Error("sections version should be set")
	}
	if len(sections.Sections) == 0 {
		t.Error("sections list should not be empty")
	}
	// First section from real bundle
	for _, s := range sections.Sections {
		if s.SectionID != "" && len(s.ChapterIDs) > 0 {
			break
		}
	}
}

func TestGrammarContentRepository_GetSections_Errors(t *testing.T) {
	logger := zap.NewNop()

	t.Run("missing sections.json returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{fs: fstest.MapFS{}, logger: logger}
		_, err := repo.GetSections()
		if err == nil {
			t.Fatal("GetSections() expected error for missing file")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs:     fstest.MapFS{"sections.json": &fstest.MapFile{Data: []byte("not json")}},
			logger: logger,
		}
		_, err := repo.GetSections()
		if err == nil {
			t.Fatal("GetSections() expected error for invalid JSON")
		}
	})
}

func TestGrammarContentRepository_GetIndex(t *testing.T) {
	repo := setupGrammarContentRepo(t)
	index, err := repo.GetIndex()
	if err != nil {
		t.Fatalf("GetIndex() error = %v", err)
	}
	if index == nil {
		t.Fatal("GetIndex() should not return nil")
	}
	if index.Version == "" {
		t.Error("index version should be set")
	}
	if len(index.Chapters) == 0 {
		t.Error("index chapters should not be empty")
	}
}

func TestGrammarContentRepository_GetIndex_Errors(t *testing.T) {
	logger := zap.NewNop()

	t.Run("missing index.json returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{fs: fstest.MapFS{}, logger: logger}
		_, err := repo.GetIndex()
		if err == nil {
			t.Fatal("GetIndex() expected error for missing file")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs:     fstest.MapFS{"index.json": &fstest.MapFile{Data: []byte("{]")}},
			logger: logger,
		}
		_, err := repo.GetIndex()
		if err == nil {
			t.Fatal("GetIndex() expected error for invalid JSON")
		}
	})
}

func TestGrammarContentRepository_GetChapter(t *testing.T) {
	repo := setupGrammarContentRepo(t)
	index, _ := repo.GetIndex()
	var firstID string
	for id := range index.Chapters {
		firstID = id
		break
	}
	if firstID == "" {
		t.Fatal("need at least one chapter in index")
	}

	t.Run("existing chapter", func(t *testing.T) {
		ch, err := repo.GetChapter(firstID)
		if err != nil {
			t.Fatalf("GetChapter() error = %v", err)
		}
		if ch == nil {
			t.Fatal("GetChapter() should not return nil")
		}
		if ch.ID != firstID {
			t.Errorf("expected chapter id %q, got %q", firstID, ch.ID)
		}
	})

	t.Run("nonexistent chapter", func(t *testing.T) {
		ch, err := repo.GetChapter("nonexistent.chapter.id")
		if err == nil {
			t.Fatal("GetChapter() expected error for nonexistent chapter")
		}
		if ch != nil {
			t.Error("GetChapter() should return nil for nonexistent chapter")
		}
	})
}

func TestGrammarContentRepository_GetChapter_Errors(t *testing.T) {
	logger := zap.NewNop()

	t.Run("chapter not in index returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs: fstest.MapFS{
				"index.json": &fstest.MapFile{Data: []byte(`{"version":"1","generated_at":"","chapters":{}}`)},
			},
			logger: logger,
		}
		_, err := repo.GetChapter("missing-chapter")
		if err == nil {
			t.Fatal("GetChapter() expected error when chapter not in index")
		}
	})

	t.Run("chapter file missing returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs: fstest.MapFS{
				"index.json": &fstest.MapFile{Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"ch1.json"}}`)},
				// chapters/ch1.json missing
			},
			logger: logger,
		}
		_, err := repo.GetChapter("ch1")
		if err == nil {
			t.Fatal("GetChapter() expected error when chapter file missing")
		}
	})

	t.Run("invalid chapter JSON returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs: fstest.MapFS{
				"index.json":       &fstest.MapFile{Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch2":"ch2.json"}}`)},
				"chapters/ch2.json": &fstest.MapFile{Data: []byte("invalid")},
			},
			logger: logger,
		}
		_, err := repo.GetChapter("ch2")
		if err == nil {
			t.Fatal("GetChapter() expected error for invalid chapter JSON")
		}
	})

	t.Run("GetIndex fails returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{fs: fstest.MapFS{}, logger: logger}
		_, err := repo.GetChapter("any-id")
		if err == nil {
			t.Fatal("GetChapter() expected error when GetIndex fails")
		}
	})
}

func TestGrammarContentRepository_ChapterExists(t *testing.T) {
	repo := setupGrammarContentRepo(t)
	index, _ := repo.GetIndex()
	var firstID string
	for id := range index.Chapters {
		firstID = id
		break
	}

	if !repo.ChapterExists(firstID) {
		t.Errorf("ChapterExists(%q) should be true", firstID)
	}
	if repo.ChapterExists("nonexistent.chapter.id") {
		t.Error("ChapterExists(nonexistent) should be false")
	}

	// When GetIndex fails, ChapterExists returns false
	badRepo := &GrammarContentRepository{fs: fstest.MapFS{}, logger: zap.NewNop()}
	if badRepo.ChapterExists("any") {
		t.Error("ChapterExists() should return false when GetIndex fails")
	}
}

func TestGrammarContentRepository_GetAllChapterIDs(t *testing.T) {
	repo := setupGrammarContentRepo(t)
	ids, err := repo.GetAllChapterIDs()
	if err != nil {
		t.Fatalf("GetAllChapterIDs() error = %v", err)
	}
	index, _ := repo.GetIndex()
	if len(ids) != len(index.Chapters) {
		t.Errorf("GetAllChapterIDs() length = %d, index chapters = %d", len(ids), len(index.Chapters))
	}
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate chapter id %q", id)
		}
		seen[id] = true
		if _, ok := index.Chapters[id]; !ok {
			t.Errorf("id %q not in index", id)
		}
	}
}

func TestGrammarContentRepository_GetAllChapterIDs_Errors(t *testing.T) {
	logger := zap.NewNop()

	t.Run("GetIndex fails returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{fs: fstest.MapFS{}, logger: logger}
		ids, err := repo.GetAllChapterIDs()
		if err == nil {
			t.Fatal("GetAllChapterIDs() expected error when GetIndex fails")
		}
		if ids != nil {
			t.Error("GetAllChapterIDs() should return nil ids on error")
		}
	})

	t.Run("invalid index JSON returns error", func(t *testing.T) {
		repo := &GrammarContentRepository{
			fs:     fstest.MapFS{"index.json": &fstest.MapFile{Data: []byte("not json")}},
			logger: logger,
		}
		_, err := repo.GetAllChapterIDs()
		if err == nil {
			t.Fatal("GetAllChapterIDs() expected error for invalid index JSON")
		}
	})
}
