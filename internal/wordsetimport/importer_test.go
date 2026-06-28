package wordsetimport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestAssertImportProfile(t *testing.T) {
	t.Parallel()

	cfgEN := &config.Config{
		Learning: config.LearningConfig{
			TargetLang: "en",
			AppCode:    "english",
		},
	}
	if err := assertImportProfile(cfgEN, "en", "resources/wordsets/english_word_freq.csv"); err != nil {
		t.Fatalf("expected english profile to pass, got error: %v", err)
	}
	if err := assertImportProfile(cfgEN, "en", "resources/wordsets/spanish_word_freq.csv"); err == nil {
		t.Fatal("expected english profile to reject spanish path")
	}
	if err := assertImportProfile(cfgEN, "fr", "resources/wordsets/english_word_freq.csv"); err == nil {
		t.Fatal("expected unsupported lang error")
	}

	cfgES := &config.Config{
		Learning: config.LearningConfig{
			TargetLang: "es",
			AppCode:    "spanish",
		},
	}
	if err := assertImportProfile(cfgES, "es", "resources/wordsets/spanish_word_freq.csv"); err != nil {
		t.Fatalf("expected spanish profile to pass, got error: %v", err)
	}
	if err := assertImportProfile(cfgES, "es", "resources/wordsets/english_word_freq.csv"); err == nil {
		t.Fatal("expected spanish profile to reject english path")
	}
}

func TestParseRankRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		title string
		ok    bool
		s     int
		e     int
	}{
		{title: "Core Verbs — Top 50 (Ranks 1–50)", ok: true, s: 1, e: 50},
		{title: "Verbos esenciales — Top 50 (rangos 101-150)", ok: true, s: 101, e: 150},
		{title: "No ranks here", ok: false},
		{title: "Ranks 0-5", ok: false},
	}
	for _, tc := range tests {
		s, e, ok := parseRankRange(tc.title)
		if ok != tc.ok {
			t.Fatalf("title=%q: expected ok=%v, got %v", tc.title, tc.ok, ok)
		}
		if ok && (s != tc.s || e != tc.e) {
			t.Fatalf("title=%q: expected range=%d-%d, got %d-%d", tc.title, tc.s, tc.e, s, e)
		}
	}
}

func TestNormalizeLangAndDefaultCSVPath(t *testing.T) {
	t.Parallel()
	if normalizeLang("") != "es" {
		t.Fatal("empty lang")
	}
	if normalizeLang("EN") != "en" {
		t.Fatal("upper lang")
	}
	if !strings.Contains(defaultCSVPathForLang("en"), "english") {
		t.Fatal("english csv default")
	}
	if !strings.Contains(defaultCSVPathForLang("es"), "spanish") {
		t.Fatal("spanish csv default")
	}
}

func TestMapUDToTrainingPOSAndLemmaValidation(t *testing.T) {
	t.Parallel()
	pos, ok := mapUDToTrainingPOS("NOUN")
	if !ok || pos != "noun" {
		t.Fatalf("noun pos=%q ok=%v", pos, ok)
	}
	if _, ok := mapUDToTrainingPOS("X"); ok {
		t.Fatal("unexpected pos")
	}
	if !isValidLemma("en", "hello") {
		t.Fatal("hello valid")
	}
	if isValidLemma("en", "a") {
		t.Fatal("single char invalid")
	}
	if !isValidLemma("es", "adiós") {
		t.Fatal("spanish lemma")
	}
}

func TestExtractRankRangeFromSet(t *testing.T) {
	t.Parallel()
	title := "Core Nouns — Top 50 (Ranks 1–50)"
	desc := "ignored"
	ws := &models.WordSet{Title: title, Description: &desc}
	s, e, ok := extractRankRangeFromSet(ws)
	if !ok || s != 1 || e != 50 {
		t.Fatalf("range=%d-%d ok=%v", s, e, ok)
	}
	if _, _, ok := extractRankRangeFromSet(nil); ok {
		t.Fatal("nil set")
	}
}

func TestSelectRankedSets(t *testing.T) {
	t.Parallel()
	sets := []*models.WordSet{
		{ID: 1, Title: "Core Verbs — Top 50 (Ranks 1–50)", SortOrder: 1, PreferredPOS: strPtr("verb")},
		{ID: 2, Title: "No ranks", SortOrder: 0},
		{ID: 3, Title: "Core Nouns — Top 50 (Ranks 51–100)", SortOrder: 2, PreferredPOS: strPtr("noun")},
	}
	out := selectRankedSets(sets, map[string]struct{}{"Core Verbs — Top 50 (Ranks 1–50)": {}})
	if len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("filtered=%v", out)
	}
	all := selectRankedSets(sets, nil)
	if len(all) != 2 {
		t.Fatalf("all=%d", len(all))
	}
}

func TestLoadWordsByTrainingPOS(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	csv := "lemma,pos,popularity_count\nhello,NOUN,100\nworld,NOUN,90\nrun,VERB,80\nx,X,70\n"
	if err := os.WriteFile(path, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	byPOS, err := loadWordsByTrainingPOS(path, "en")
	if err != nil {
		t.Fatal(err)
	}
	if len(byPOS["noun"]) != 2 || byPOS["noun"][0].Lemma != "hello" {
		t.Fatalf("nouns=%v", byPOS["noun"])
	}
	if len(byPOS["verb"]) != 1 {
		t.Fatalf("verbs=%v", byPOS["verb"])
	}
}

func TestLoadWordsByTrainingPOSInvalidCSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(path, []byte("only,header\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadWordsByTrainingPOS(path, "en"); err == nil {
		t.Fatal("expected error for missing columns")
	}
}

func TestImportDryRunEnglish(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
	cfg := englishTestConfig()
	csvPath := testEnglishCSVPath(t)
	res, err := Import(context.Background(), cfg, conn, zap.NewNop(), ImportOptions{
		CSVPath:   csvPath,
		Lang:      "en",
		Commit:    false,
		LimitSets: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "DRY-RUN" || len(res.Sets) == 0 {
		t.Fatalf("result=%+v", res)
	}
	foundImported := false
	for _, s := range res.Sets {
		if !s.Skipped && s.Imported > 0 {
			foundImported = true
			break
		}
	}
	if !foundImported {
		t.Fatalf("sets=%+v", res.Sets)
	}
}

func TestImportCommitLimited(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
	cfg := englishTestConfig()
	csvPath := testEnglishCSVPath(t)
	res, err := Import(context.Background(), cfg, conn, zap.NewNop(), ImportOptions{
		CSVPath:   csvPath,
		Lang:      "en",
		Commit:    true,
		LimitSets: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Processed != 1 {
		t.Fatalf("processed=%d sets=%+v", res.Processed, res.Sets)
	}
}

func strPtr(s string) *string { return &s }
