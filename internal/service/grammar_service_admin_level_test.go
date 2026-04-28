package service

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
	"testing/fstest"
)

func TestGrammarService_GetSectionBySectionID(t *testing.T) {
	svc, contentRepo, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if _, err := svc.GetSectionBySectionID(context.Background(), " "); err == nil {
		t.Fatal("expected error for empty section id")
	}

	sections, err := contentRepo.GetSections()
	if err != nil || len(sections.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	found, err := svc.GetSectionBySectionID(context.Background(), sections.Sections[0].SectionID)
	if err != nil {
		t.Fatalf("GetSectionBySectionID existing: %v", err)
	}
	if found == nil || found.SectionID != sections.Sections[0].SectionID {
		t.Fatal("expected found section")
	}

	if _, err := svc.GetSectionBySectionID(context.Background(), "__missing__"); err == nil {
		t.Fatal("expected error for missing section")
	}

	// GetSections error branch
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	badSvc := NewGrammarService(
		repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger),
		repository.NewGrammarPublishRepository(db.GetConnection(), logger),
		repository.NewGrammarAttemptRepository(db.GetConnection(), logger),
		config.DefaultLearningConfig(),
		logger,
	)
	if _, err := badSvc.GetSectionBySectionID(context.Background(), "s1"); err == nil {
		t.Fatal("expected GetSections error")
	}
}

func TestGrammarService_GrammarLevelHelpersAndFormatDisplay(t *testing.T) {
	svc, contentRepo, _, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if got := grammarLevelOrderMap(); got["A1"] != 1 || got["C2"] != 6 || got["mixed"] != -1 {
		t.Fatalf("unexpected grammar level map: %+v", got)
	}

	if got := svc.FormatPlacementLevelDisplay(nil, false); got != "" {
		t.Fatalf("expected empty label without placement row, got %q", got)
	}
	if got := svc.FormatPlacementLevelDisplay([]string{}, true); got != "Below A1" {
		t.Fatalf("expected Below A1 for empty opened, got %q", got)
	}

	sections, err := contentRepo.GetSections()
	if err != nil || len(sections.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	opened := []string{sections.Sections[0].SectionID, "__unknown__"}
	label := svc.FormatPlacementLevelDisplay(opened, true)
	if label == "" {
		t.Fatal("expected non-empty label")
	}
	if got := svc.FormatPlacementLevelDisplay([]string{"__unknown__"}, true); got != "—" {
		t.Fatalf("expected em dash for unknown opened sections, got %q", got)
	}

	// Content repo error branch => em dash
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	badContent := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	badSvc := NewGrammarService(
		badContent,
		repository.NewGrammarPublishRepository(db.GetConnection(), logger),
		repository.NewGrammarAttemptRepository(db.GetConnection(), logger),
		config.DefaultLearningConfig(),
		logger,
	)
	if got := badSvc.FormatPlacementLevelDisplay([]string{"s1"}, true); got != "—" {
		t.Fatalf("expected em dash on content error, got %q", got)
	}
}

func TestGrammarService_OpenPublishedSectionsThroughLevel(t *testing.T) {
	svc, contentRepo, publishRepo, _, _, cleanup := setupGrammarService(t)
	defer cleanup()

	if _, err := svc.OpenPublishedSectionsThroughLevel("unknown"); err == nil {
		t.Fatal("expected invalid level error")
	}

	sections, err := contentRepo.GetSections()
	if err != nil || len(sections.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	for _, sec := range sections.Sections {
		_ = publishRepo.SetPublished("section", sec.SectionID, true, nil)
	}
	opened, err := svc.OpenPublishedSectionsThroughLevel("B1")
	if err != nil {
		t.Fatalf("OpenPublishedSectionsThroughLevel: %v", err)
	}
	if len(opened) == 0 {
		t.Fatal("expected opened sections for B1")
	}

	// Branches: unpublished section and unknown/mixed level are skipped.
	customFS := fstest.MapFS{
		"sections.json": {Data: []byte(`{"version":"1","sections":[
			{"section_id":"s_pub","title":"Published","level":"A1","order":1,"chapter_ids":[]},
			{"section_id":"s_unpub","title":"Unpublished","level":"A1","order":2,"chapter_ids":[]},
			{"section_id":"s_mixed","title":"Mixed","level":"mixed","order":3,"chapter_ids":[]}
		]}`)},
		"index.json": {Data: []byte(`{"version":"1","generated_at":"","chapters":{}}`)},
	}
	logger := zap.NewNop()
	db2 := testutil.SetupTestDatabase(t)
	content2 := repository.NewGrammarContentRepositoryWithFS(customFS, logger)
	publish2 := repository.NewGrammarPublishRepository(db2.GetConnection(), logger)
	attempt2 := repository.NewGrammarAttemptRepository(db2.GetConnection(), logger)
	svc2 := NewGrammarService(content2, publish2, attempt2, config.DefaultLearningConfig(), logger)
	_ = publish2.SetPublished("section", "s_pub", true, nil)
	_ = publish2.SetPublished("section", "s_mixed", true, nil)
	opened2, err := svc2.OpenPublishedSectionsThroughLevel("A1")
	if err != nil {
		t.Fatalf("OpenPublishedSectionsThroughLevel custom: %v", err)
	}
	if len(opened2) != 1 || opened2[0] != "s_pub" {
		t.Fatalf("expected only published known-level section, got %+v", opened2)
	}

	// Content repo error branch
	logger2 := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	badSvc := NewGrammarService(
		repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger2),
		repository.NewGrammarPublishRepository(db.GetConnection(), logger2),
		repository.NewGrammarAttemptRepository(db.GetConnection(), logger2),
		config.DefaultLearningConfig(),
		logger2,
	)
	if _, err := badSvc.OpenPublishedSectionsThroughLevel("A1"); err == nil {
		t.Fatal("expected content repo error")
	}
}

func TestGrammarService_AdminSetGrammarPlacementLevel(t *testing.T) {
	svc, contentRepo, publishRepo, attemptRepo, _, cleanup := setupGrammarService(t)
	defer cleanup()

	sections, err := contentRepo.GetSections()
	if err != nil || len(sections.Sections) == 0 {
		t.Fatalf("GetSections: %v", err)
	}
	for _, sec := range sections.Sections {
		_ = publishRepo.SetPublished("section", sec.SectionID, true, nil)
	}

	// invalid level
	if err := svc.AdminSetGrammarPlacementLevel(context.Background(), 1, "nope"); err == nil {
		t.Fatal("expected invalid level error")
	}
	// below_a1 variants
	if err := svc.AdminSetGrammarPlacementLevel(context.Background(), 1, "below_a1"); err != nil {
		t.Fatalf("below_a1: %v", err)
	}
	if err := svc.AdminSetGrammarPlacementLevel(context.Background(), 1, "Below A1"); err != nil {
		t.Fatalf("Below A1: %v", err)
	}
	// normal level
	if err := svc.AdminSetGrammarPlacementLevel(context.Background(), 1, "a1"); err != nil {
		t.Fatalf("A1: %v", err)
	}
	res, err := attemptRepo.GetPlacementTestResult(1)
	if err != nil {
		t.Fatalf("GetPlacementTestResult: %v", err)
	}
	if res == nil {
		t.Fatal("expected placement row after admin set")
	}
	// empty => delete row
	if err := svc.AdminSetGrammarPlacementLevel(context.Background(), 1, ""); err != nil {
		t.Fatalf("clear placement: %v", err)
	}
	res, err = attemptRepo.GetPlacementTestResult(1)
	if err != nil {
		t.Fatalf("GetPlacementTestResult after delete: %v", err)
	}
	if res != nil {
		t.Fatal("expected placement row deleted")
	}

	// OpenPublishedSectionsThroughLevel error branch propagation.
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	badSvc := NewGrammarService(
		repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger),
		repository.NewGrammarPublishRepository(db.GetConnection(), logger),
		repository.NewGrammarAttemptRepository(db.GetConnection(), logger),
		config.DefaultLearningConfig(),
		logger,
	)
	if err := badSvc.AdminSetGrammarPlacementLevel(context.Background(), 1, "A1"); err == nil {
		t.Fatal("expected error propagated from OpenPublishedSectionsThroughLevel")
	}
}
