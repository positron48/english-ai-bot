package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const wordReposGapUserTelegramID int64 = 900001

const wordReposGapInvalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

func wordReposGapDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
}

func wordReposGapInvalidDB(t *testing.T) *sql.DB {
	t.Helper()
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", wordReposGapInvalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered:", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func wordReposGapUser(t *testing.T, db *sql.DB, telegramID int64) int64 {
	t.Helper()
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user %d: %v", telegramID, err)
	}
	return user.ID
}

func TestWordReposGap_TagWordCardCourse(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	if err := repo.TagWordCardCourse(1, "  "); err != nil {
		t.Fatalf("empty course: %v", err)
	}

	if err := repo.SaveWordCard("gap-tag-untagged", "def", ""); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, err := repo.GetWordCard("gap-tag-untagged")
	if err != nil || card == nil {
		t.Fatalf("GetWordCard: err=%v card=%v", err, card)
	}
	if err := repo.TagWordCardCourse(card.ID, " EN_RU "); err != nil {
		t.Fatalf("TagWordCardCourse: %v", err)
	}
	got, _ := repo.GetWordCardByID(card.ID)
	if got == nil || got.CourseCode != "en_ru" {
		t.Fatalf("course_code = %q, want en_ru", got.CourseCode)
	}

	if err := repo.SaveWordCard("gap-tag-tagged", "def", "es_ru"); err != nil {
		t.Fatalf("SaveWordCard es: %v", err)
	}
	esCard, _ := repo.GetWordCardByLemmaForCourse("gap-tag-tagged", "es_ru")
	if err := repo.TagWordCardCourse(esCard.ID, "en_ru"); err != nil {
		t.Fatalf("TagWordCardCourse overwrite attempt: %v", err)
	}
	still, _ := repo.GetWordCardByID(esCard.ID)
	if still.CourseCode != "es_ru" {
		t.Fatalf("course_code overwritten to %q", still.CourseCode)
	}
}

func TestWordReposGap_GetWordFormMappingForCourse(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	if err := repo.SaveWordCard("gap-lemma-en", "def", "en_ru"); err != nil {
		t.Fatalf("SaveWordCard en: %v", err)
	}
	if err := repo.SaveWordCard("gap-lemma-es", "def", "es_ru"); err != nil {
		t.Fatalf("SaveWordCard es: %v", err)
	}
	enCard, _ := repo.GetWordCardByLemmaForCourse("gap-lemma-en", "en_ru")
	esCard, _ := repo.GetWordCardByLemmaForCourse("gap-lemma-es", "es_ru")
	if err := repo.UpsertWordFormMapping("gapform", enCard.ID); err != nil {
		t.Fatalf("UpsertWordFormMapping en: %v", err)
	}
	if err := repo.UpsertWordFormMapping("gapformes", esCard.ID); err != nil {
		t.Fatalf("UpsertWordFormMapping es: %v", err)
	}

	fallback, err := repo.GetWordFormMappingForCourse("gapform", "")
	if err != nil || fallback == nil || fallback.WordCardID != enCard.ID {
		t.Fatalf("empty course fallback = %+v err=%v", fallback, err)
	}

	enMap, err := repo.GetWordFormMappingForCourse("gapform", "en_ru")
	if err != nil || enMap == nil || enMap.WordCardID != enCard.ID {
		t.Fatalf("en_ru map = %+v err=%v", enMap, err)
	}

	wrong, err := repo.GetWordFormMappingForCourse("gapform", "es_ru")
	if err != nil {
		t.Fatalf("es lookup on en form: %v", err)
	}
	if wrong != nil {
		t.Fatalf("expected nil for cross-course form, got %+v", wrong)
	}

	esMap, err := repo.GetWordFormMappingForCourse("gapformes", "es_ru")
	if err != nil || esMap == nil || esMap.WordCardID != esCard.ID {
		t.Fatalf("es_ru map = %+v err=%v", esMap, err)
	}

	missing, err := repo.GetWordFormMappingForCourse("no-such-gap-form", "en_ru")
	if err != nil || missing != nil {
		t.Fatalf("missing form = %+v err=%v", missing, err)
	}
}

func TestWordReposGap_ListRecentWords(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	words := []struct {
		word, course string
	}{
		{"gap-recent-alpha", "en_ru"},
		{"gap-recent-beta", "en_ru"},
		{"gap-recent-gamma", "es_ru"},
		{"gap-recent-legacy", ""},
	}
	for _, w := range words {
		if err := repo.SaveWordCard(w.word, "def", w.course); err != nil {
			t.Fatalf("SaveWordCard %q: %v", w.word, err)
		}
	}
	if _, err := db.Exec(`UPDATE word_cards SET word = '' WHERE word = 'gap-recent-legacy'`); err != nil {
		t.Fatalf("blank word: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO word_cards (word, definition, course_code) VALUES ('gap-recent-dup', 'd', NULL)`); err != nil {
		t.Fatalf("dup legacy insert: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO word_cards (word, definition, course_code) VALUES ('gap-recent-dup', 'd2', 'en_ru')`); err != nil {
		t.Fatalf("dup course insert: %v", err)
	}

	recent, err := repo.ListRecentWords("en_ru", 2)
	if err != nil {
		t.Fatalf("ListRecentWords: %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("ListRecentWords len = %d, want 2", len(recent))
	}
	seen := make(map[string]struct{}, len(recent))
	for _, w := range recent {
		seen[w] = struct{}{}
	}
	if _, ok := seen["gap-recent-dup"]; !ok {
		t.Fatalf("expected deduped gap-recent-dup in %v", recent)
	}

	defaultLimit, err := repo.ListRecentWordsPage("en_ru", 0, -1)
	if err != nil {
		t.Fatalf("ListRecentWordsPage default limit: %v", err)
	}
	if len(defaultLimit) == 0 {
		t.Fatal("expected default limit results")
	}

	page, err := repo.ListRecentWordsPage("en_ru", 1, 1)
	if err != nil {
		t.Fatalf("ListRecentWordsPage offset: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("page len = %d, want 1", len(page))
	}

	esOnly, err := repo.ListRecentWords("es_ru", 10)
	if err != nil {
		t.Fatalf("ListRecentWords es: %v", err)
	}
	for _, w := range esOnly {
		if w == "gap-recent-alpha" || w == "gap-recent-beta" {
			t.Fatalf("en-only word %q in es list %v", w, esOnly)
		}
	}

	for i := 0; i < 6; i++ {
		word := fmt.Sprintf("gap-recent-bulk-%02d", i)
		if err := repo.SaveWordCard(word, "def", "en_ru"); err != nil {
			t.Fatalf("bulk save %s: %v", word, err)
		}
	}
	bulk, err := repo.ListRecentWordsPage("en_ru", 3, 0)
	if err != nil || len(bulk) != 3 {
		t.Fatalf("bulk limit = %v err=%v", bulk, err)
	}

	if _, err := db.Exec(`INSERT INTO word_cards (word, definition, course_code) VALUES ('   ', 'blank', 'en_ru')`); err != nil {
		t.Fatalf("spaces word: %v", err)
	}
	spaced, err := repo.ListRecentWordsPage("en_ru", 50, 0)
	if err != nil {
		t.Fatalf("ListRecentWordsPage spaced: %v", err)
	}
	for _, w := range spaced {
		if w == "" || w == "   " {
			t.Fatalf("blank word leaked: %q in %v", w, spaced)
		}
	}
}

func TestWordReposGap_TTSHelpers(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	if path, err := repo.GetTTSAudioRelPath("en_ru", "missing-gap-tts"); err != nil || path != "" {
		t.Fatalf("missing path = %q err=%v", path, err)
	}

	if _, err := db.Exec(`
		INSERT INTO tts_generation_status (course_code, word, state, audio_rel_path)
		VALUES ('en_ru', 'gap-tts-word', 'ready', 'ab/cd/gap-tts-word.mp3')
		ON CONFLICT (course_code, word) DO UPDATE SET audio_rel_path = EXCLUDED.audio_rel_path
	`); err != nil {
		t.Fatalf("insert tts ready: %v", err)
	}
	path, err := repo.GetTTSAudioRelPath("en_ru", "GAP-TTS-WORD")
	if err != nil || path != "ab/cd/gap-tts-word.mp3" {
		t.Fatalf("ready path = %q err=%v", path, err)
	}

	if _, err := db.Exec(`
		INSERT INTO tts_generation_status (course_code, word, state, audio_rel_path)
		VALUES ('en_ru', 'gap-tts-null', 'pending', NULL)
		ON CONFLICT (course_code, word) DO UPDATE SET audio_rel_path = NULL, state = 'pending'
	`); err != nil {
		t.Fatalf("insert tts null path: %v", err)
	}
	nullPath, err := repo.GetTTSAudioRelPath("en_ru", "gap-tts-null")
	if err != nil || nullPath != "" {
		t.Fatalf("null path = %q err=%v", nullPath, err)
	}

	if err := repo.DeleteTTSStatus("en_ru", "gap-tts-word"); err != nil {
		t.Fatalf("DeleteTTSStatus: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tts_generation_status WHERE course_code = 'en_ru' AND word = 'gap-tts-word'`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("row not deleted, count=%d", count)
	}
}

func TestWordReposGap_MergeWordFormInto(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	ctx := context.Background()

	var formID, canonID int64
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, course_code)
		VALUES ('gap-running', 'to run', 'en_ru') RETURNING id
	`).Scan(&formID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		INSERT INTO word_cards (word, definition, course_code)
		VALUES ('gap-run', 'to run', 'en_ru') RETURNING id
	`).Scan(&canonID); err != nil {
		t.Fatal(err)
	}

	if err := repo.MergeWordFormInto(ctx, formID, formID); err == nil {
		t.Fatal("expected error for same IDs")
	}

	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID)
	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, userID, formID); err != nil {
		t.Fatalf("user_word_knowledge: %v", err)
	}

	if err := repo.MergeWordFormInto(ctx, formID, canonID); err != nil {
		t.Fatalf("MergeWordFormInto: %v", err)
	}

	var formCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM word_cards WHERE id = ?`, formID).Scan(&formCount); err != nil {
		t.Fatal(err)
	}
	if formCount != 0 {
		t.Fatal("form card should be deleted")
	}
	var knowCardID int64
	if err := db.QueryRow(`SELECT word_card_id FROM user_word_knowledge WHERE user_id = ?`, userID).Scan(&knowCardID); err != nil {
		t.Fatal(err)
	}
	if knowCardID != canonID {
		t.Fatalf("knowledge card = %d, want %d", knowCardID, canonID)
	}
}

func TestWordReposGap_WordSetCategoryForCourse(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())

	desc := "gap category"
	level := "A1"
	cat := &models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap EN Category",
		Description: &desc,
		IsPublished: true,
		SortOrder:   1,
		LevelCode:   &level,
	}
	id, err := repo.CreateCategory(cat)
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	got, err := repo.GetCategoryForCourse(id, "en_ru")
	if err != nil || got == nil || got.Name != "Gap EN Category" || got.CourseCode != "en_ru" {
		t.Fatalf("GetCategoryForCourse = %+v err=%v", got, err)
	}

	wrong, err := repo.GetCategoryForCourse(id, "es_ru")
	if err != nil || wrong != nil {
		t.Fatalf("wrong course = %+v err=%v", wrong, err)
	}

	all, err := repo.GetAllCategoriesForCourse("en_ru")
	if err != nil {
		t.Fatalf("GetAllCategoriesForCourse en: %v", err)
	}
	found := false
	for _, c := range all {
		if c.ID == id {
			found = true
			if c.CourseCode != "en_ru" {
				t.Fatalf("course_code = %q", c.CourseCode)
			}
		}
	}
	if !found {
		t.Fatal("category not in en_ru list")
	}

	allAny, err := repo.GetAllCategoriesForCourse("")
	if err != nil || len(allAny) == 0 {
		t.Fatalf("GetAllCategoriesForCourse all = %d err=%v", len(allAny), err)
	}

	updatedDesc := "updated gap"
	got.Description = &updatedDesc
	got.Name = "Gap EN Updated"
	got.SortOrder = 2
	if err := repo.UpdateCategoryForCourse(got); err != nil {
		t.Fatalf("UpdateCategoryForCourse: %v", err)
	}
	after, _ := repo.GetCategoryForCourse(id, "en_ru")
	if after.Name != "Gap EN Updated" {
		t.Fatalf("name = %q", after.Name)
	}

	got.CourseCode = "es_ru"
	if err := repo.UpdateCategoryForCourse(got); err == nil {
		t.Fatal("expected not found for wrong course update")
	}
}

func TestWordReposGap_WordSetForCourse(t *testing.T) {
	db := wordReposGapDB(t)
	catRepo := NewWordSetCategoryRepository(db, zap.NewNop())
	setRepo := NewWordSetRepository(db, zap.NewNop())

	catID, err := catRepo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap Set Category",
		IsPublished: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	desc := "gap set"
	level := "A2"
	pos := "noun"
	pub := &models.WordSet{
		CourseCode:   "en_ru",
		CategoryID:   &catID,
		Title:        "Gap Published Set",
		Description:  &desc,
		IsPublished:  true,
		SortOrder:    1,
		PreferredPOS: &pos,
		LevelCode:    &level,
	}
	pubID, err := setRepo.CreateWordSet(pub)
	if err != nil {
		t.Fatalf("CreateWordSet pub: %v", err)
	}
	unpubID, err := setRepo.CreateWordSet(&models.WordSet{
		CourseCode:  "en_ru",
		CategoryID:  &catID,
		Title:       "Gap Draft Set",
		IsPublished: false,
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("CreateWordSet unpub: %v", err)
	}

	got, err := setRepo.GetWordSetForCourse(pubID, "en_ru")
	if err != nil || got == nil || got.CourseCode != "en_ru" || got.Title != "Gap Published Set" {
		t.Fatalf("GetWordSetForCourse = %+v err=%v", got, err)
	}
	if got.PreferredPOS == nil || *got.PreferredPOS != "noun" {
		t.Fatalf("PreferredPOS = %v", got.PreferredPOS)
	}
	if got.LevelCode == nil || *got.LevelCode != "A2" {
		t.Fatalf("LevelCode = %v", got.LevelCode)
	}

	noCourse, err := setRepo.GetWordSetForCourse(pubID, "")
	if err != nil || noCourse == nil {
		t.Fatalf("GetWordSetForCourse no filter = %+v err=%v", noCourse, err)
	}

	wrong, err := setRepo.GetWordSetForCourse(pubID, "es_ru")
	if err != nil || wrong != nil {
		t.Fatalf("wrong course set = %+v err=%v", wrong, err)
	}

	published, err := setRepo.ListWordSetsForCourse("en_ru", &catID, 10, 0)
	if err != nil {
		t.Fatalf("ListWordSetsForCourse published: %v", err)
	}
	if len(published) != 1 || published[0].ID != pubID {
		t.Fatalf("published sets = %+v", published)
	}

	all, err := setRepo.ListWordSetsForCourse("en_ru", &catID, 10, 0, true)
	if err != nil || len(all) < 2 {
		t.Fatalf("all sets = %d err=%v", len(all), err)
	}
	_ = unpubID

	updated := *got
	updated.Title = "Gap Updated Title"
	updated.CourseCode = "en_ru"
	if err := setRepo.UpdateWordSet(&updated); err != nil {
		t.Fatalf("UpdateWordSet scoped: %v", err)
	}

	updated.CourseCode = "es_ru"
	if err := setRepo.UpdateWordSet(&updated); err == nil {
		t.Fatal("expected not found for wrong course UpdateWordSet")
	}

	if err := setRepo.DeleteWordSetForCourse(pubID, "es_ru"); err == nil {
		t.Fatal("expected not found for wrong course delete")
	}
	if err := setRepo.DeleteWordSetForCourse(unpubID, "en_ru"); err != nil {
		t.Fatalf("DeleteWordSetForCourse: %v", err)
	}
}

func TestWordReposGap_EffectiveLevelCodeAndAggregateProgress(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+1)

	catLevel := "B1"
	setLevel := "A2"
	setRepoLevel := EffectiveLevelCode(&models.WordSet{LevelCode: &setLevel}, &catLevel)
	if setRepoLevel != "A2" {
		t.Fatalf("set level = %q", setRepoLevel)
	}
	catOnly := EffectiveLevelCode(&models.WordSet{}, &catLevel)
	if catOnly != "B1" {
		t.Fatalf("cat level = %q", catOnly)
	}
	if empty := EffectiveLevelCode(&models.WordSet{}, nil); empty != "" {
		t.Fatalf("empty level = %q", empty)
	}

	total, known, inVocab, err := setRepo.GetCategoriesAggregateProgress(nil, userID)
	if err != nil || total != 0 || known != 0 || inVocab != 0 {
		t.Fatalf("empty categories = %d/%d/%d err=%v", total, known, inVocab, err)
	}

	catRepo := NewWordSetCategoryRepository(db, zap.NewNop())
	catID, _ := catRepo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap Progress Cat",
		IsPublished: true,
		SortOrder:   1,
	})
	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode:  "en_ru",
		CategoryID:  &catID,
		Title:       "Gap Progress Set",
		IsPublished: true,
		SortOrder:   1,
	})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-progress-word", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatalf("SetWordSetItems: %v", err)
	}

	total, known, inVocab, err = setRepo.GetCategoriesAggregateProgress([]int64{catID}, userID)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if total != 1 || known != 0 || inVocab != 0 {
		t.Fatalf("before training = %d/%d/%d", total, known, inVocab)
	}

	progress, err := setRepo.GetWordSetProgress(setID, userID)
	if err != nil || progress.TotalWords != 1 {
		t.Fatalf("progress = %+v err=%v", progress, err)
	}
	if progress.UnknownWords != 1 {
		t.Fatalf("unknown = %d", progress.UnknownWords)
	}
}

func TestWordReposGap_WordCardNullableFields(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	processedAt := time.Now().UTC().Truncate(time.Second)
	procErr := "gap proc err"
	pos := "noun"
	gender := "m"
	opp := "gap-opposite"
	trans := "[ɡæp]"
	defRU := "гэп"
	examples := `["ex"]`
	verbs := `{"past":"gapped"}`
	display := "GAP"
	var id int64
	err := db.QueryRow(`
		INSERT INTO word_cards (
			word, definition, pos, noun_gender, opposite_gender_word, transcription,
			definition_ru, examples_json, verb_forms_json, display_en,
			processed_at, processing_error, course_code
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`, "gap-full-card", "def", pos, gender, opp, trans, defRU, examples, verbs, display,
		processedAt.Format("2006-01-02 15:04:05"), procErr, "en_ru").Scan(&id)
	if err != nil {
		t.Fatalf("insert full card: %v", err)
	}

	byID, err := repo.GetWordCardByID(id)
	if err != nil || byID == nil {
		t.Fatalf("GetWordCardByID: err=%v card=%v", err, byID)
	}
	if byID.POS == nil || byID.ProcessedAt == nil || byID.ProcessingError == nil {
		t.Fatalf("nullable fields missing on byID: %+v", byID)
	}

	byLemma, err := repo.GetWordCardByLemma("GAP-FULL-CARD")
	if err != nil || byLemma == nil || byLemma.Transcription == nil {
		t.Fatalf("GetWordCardByLemma: err=%v card=%v", err, byLemma)
	}

	byCourse, err := repo.GetWordCardByLemmaForCourse("gap-full-card", "en_ru")
	if err != nil || byCourse == nil || byCourse.DisplayEN == nil {
		t.Fatalf("GetWordCardByLemmaForCourse: err=%v card=%v", err, byCourse)
	}

	legacy, err := repo.GetWordCardByLemmaForCourse("gap-full-card", "")
	if err != nil || legacy != nil {
		t.Fatalf("legacy course empty should not match tagged row: %+v err=%v", legacy, err)
	}

	missing, err := repo.GetWordCardByID(999999001)
	if err != nil || missing != nil {
		t.Fatalf("missing by id = %+v err=%v", missing, err)
	}
}

func TestWordReposGap_GetUserIDsByWordCardID(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+2)

	if err := repo.SaveWordCard("gap-history-word", "def", "en_ru"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemmaForCourse("gap-history-word", "en_ru")
	input := "gap-history-input"
	if err := repo.AddWordRequestHistoryWithCard(userID, input, &card.ID, &card.Word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard with card: %v", err)
	}
	if err := repo.AddWordRequestHistoryWithCard(userID, input, nil, nil); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard legacy: %v", err)
	}

	byCard, err := repo.GetUserIDsByWordCardID(card.ID, card.Word)
	if err != nil {
		t.Fatalf("GetUserIDsByWordCardID: %v", err)
	}
	if len(byCard) != 1 || byCard[0] != userID {
		t.Fatalf("by card = %v, want [%d]", byCard, userID)
	}

	byWord, err := repo.GetUserIDsByWord(card.Word)
	if err != nil || len(byWord) != 1 {
		t.Fatalf("GetUserIDsByWord = %v err=%v", byWord, err)
	}

	onlyInput := "gap-only-input-word"
	if err := repo.AddWordRequestHistoryWithCard(userID, onlyInput, nil, nil); err != nil {
		t.Fatalf("legacy only: %v", err)
	}
	legacyIDs, err := repo.GetUserIDsByWord(onlyInput)
	if err != nil || len(legacyIDs) != 1 {
		t.Fatalf("legacy GetUserIDsByWord = %v err=%v", legacyIDs, err)
	}
}

func TestWordReposGap_ListWordCardsAdminCourseCodes(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	if err := repo.SaveWordCard("gap-admin-course", "def", "en_ru"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemmaForCourse("gap-admin-course", "en_ru")
	var courseID int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE code = 'en_ru'`).Scan(&courseID); err != nil {
		t.Fatalf("course id: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO learning_items (course_id, item_type, source_kind, source_id, title, status)
		VALUES (?, 'word', 'word_card', ?, 'gap admin item', 'published')
		ON CONFLICT (course_id, source_kind, source_id) DO NOTHING
	`, courseID, fmt.Sprintf("%d", card.ID)); err != nil {
		t.Fatalf("learning item: %v", err)
	}

	items, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "gap-admin", "", 10, 0, "word", "asc")
	if err != nil {
		t.Fatalf("ListWordCardsAdminForCourse: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected admin items")
	}
	found := false
	for _, it := range items {
		if it.Word == "gap-admin-course" {
			found = true
			if len(it.CourseCodes) == 0 {
				t.Fatalf("course codes empty for %+v", it)
			}
		}
	}
	if !found {
		t.Fatalf("gap-admin-course not in %+v", items)
	}

	hasAudio := true
	withAudio, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, &hasAudio, "", "", 10, 0, "has_cards", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdminForCourse hasAudio: %v", err)
	}
	_ = withAudio
}

func TestWordReposGap_WordSetProgressKnownAndVocab(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+3)

	catID, _ := NewWordSetCategoryRepository(db, zap.NewNop()).CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap Progress Full",
		IsPublished: true,
		SortOrder:   1,
	})
	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode:  "en_ru",
		CategoryID:  &catID,
		Title:       "Gap Full Set",
		IsPublished: true,
		SortOrder:   1,
	})

	knownID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-known-word", Definition: "def"})
	vocabID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-vocab-word", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{knownID, vocabID}); err != nil {
		t.Fatalf("SetWordSetItems: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, userID, knownID); err != nil {
		t.Fatalf("known: %v", err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES (?, 'gap-vocab-word', 0, 'слово', 'gap vocab', 'noun') RETURNING id
	`, vocabID).Scan(&trainingCardID); err != nil {
		t.Fatalf("training card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'en_ru', 'new', 2.5)`, userID, trainingCardID); err != nil {
		t.Fatalf("user card: %v", err)
	}

	progress, err := setRepo.GetWordSetProgress(setID, userID)
	if err != nil {
		t.Fatalf("GetWordSetProgress: %v", err)
	}
	if progress.TotalWords != 2 || progress.KnownWords != 1 || progress.WordsInVocab != 1 || progress.UnknownWords != 0 {
		t.Fatalf("progress = %+v", progress)
	}
	if progress.ProgressPercent != 100 {
		t.Fatalf("percent = %f", progress.ProgressPercent)
	}

	total, known, inVocab, err := setRepo.GetCategoriesAggregateProgress([]int64{catID}, userID)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if total != 2 || known != 1 || inVocab != 1 {
		t.Fatalf("aggregate = %d/%d/%d", total, known, inVocab)
	}
}

// A word card shared by two sets in the same category must not be double-counted in the
// category denominator: each set reads 100% and so must the category aggregate. Guards the
// COUNT(DISTINCT word_card_id) total fix (previously COUNT(*) over rows inflated the total).
func TestWordReposGap_AggregateProgressSharedWordAcrossSets(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+50)

	catID, _ := NewWordSetCategoryRepository(db, zap.NewNop()).CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap Shared Word",
		IsPublished: true,
		SortOrder:   1,
	})
	set1, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode: "en_ru", CategoryID: &catID, Title: "Shared Set 1", IsPublished: true, SortOrder: 1,
	})
	set2, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode: "en_ru", CategoryID: &catID, Title: "Shared Set 2", IsPublished: true, SortOrder: 2,
	})

	sharedID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-shared-word", Definition: "def"})
	uniq1, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-uniq-word-1", Definition: "def"})
	uniq2, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-uniq-word-2", Definition: "def"})
	if err := setRepo.SetWordSetItems(set1, []int64{sharedID, uniq1}); err != nil {
		t.Fatalf("SetWordSetItems set1: %v", err)
	}
	if err := setRepo.SetWordSetItems(set2, []int64{sharedID, uniq2}); err != nil {
		t.Fatalf("SetWordSetItems set2: %v", err)
	}

	for _, id := range []int64{sharedID, uniq1, uniq2} {
		if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, userID, id); err != nil {
			t.Fatalf("known %d: %v", id, err)
		}
	}

	// Each set is fully learned.
	for _, setID := range []int64{set1, set2} {
		p, err := setRepo.GetWordSetProgress(setID, userID)
		if err != nil {
			t.Fatalf("GetWordSetProgress %d: %v", setID, err)
		}
		if p.ProgressPercent != 100 {
			t.Fatalf("set %d percent = %f, want 100", setID, p.ProgressPercent)
		}
	}

	// Category aggregate dedups the shared card: 3 distinct cards, all known → 100%.
	total, known, inVocab, err := setRepo.GetCategoriesAggregateProgress([]int64{catID}, userID)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if total != 3 || known != 3 || inVocab != 0 {
		t.Fatalf("aggregate = %d/%d/%d, want 3/3/0", total, known, inVocab)
	}
}

func TestWordReposGap_GetWordSetWordsPreferredPOS(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+4)

	pos := "verb"
	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode:   "en_ru",
		Title:        "Gap POS Set",
		IsPublished:  true,
		SortOrder:    1,
		PreferredPOS: &pos,
	})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-pos-word", Definition: "def", DisplayEN: strPtr("GAP POS")})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatalf("SetWordSetItems: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word, transcription, example_en, example_ru)
		VALUES (?, 'gap-pos-word', 0, 'поз', 'gap pos meaning', 'verb', 'gap-pos-display', '[pos]', 'example en', 'example ru')
	`, cardID); err != nil {
		t.Fatalf("training card: %v", err)
	}

	words, err := setRepo.GetWordSetWords(setID, userID)
	if err != nil {
		t.Fatalf("GetWordSetWords: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("words = %+v", words)
	}
	w := words[0]
	if w.DisplayWord != "gap-pos-display" || w.Transcription == nil || w.WordRU == nil || w.MeaningEN == nil || w.ExampleEN == nil || w.ExampleRU == nil {
		t.Fatalf("word info = %+v", w)
	}
}

func TestWordReposGap_DeleteCategoryForCourseNotFound(t *testing.T) {
	repo := NewWordSetCategoryRepository(wordReposGapDB(t), zap.NewNop())
	if err := repo.DeleteCategoryForCourse(999999002, "en_ru"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestWordReposGap_GetCategoryRFC3339Time(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	var id int64
	if err := db.QueryRow(`
		INSERT INTO word_set_categories (course_code, name, is_published, sort_order, created_at, updated_at)
		VALUES ('en_ru', 'Gap RFC3339', 1, 1, '2024-06-01T12:00:00Z', '2024-06-01T12:00:00Z')
		RETURNING id
	`).Scan(&id); err != nil {
		t.Fatalf("insert: %v", err)
	}
	cat, err := repo.GetCategory(id)
	if err != nil || cat == nil || cat.CreatedAt.IsZero() {
		t.Fatalf("GetCategory = %+v err=%v", cat, err)
	}
}

func TestWordReposGap_DBErrors(t *testing.T) {
	db := wordReposGapInvalidDB(t)
	ctx := context.Background()

	if err := NewWordRepository(db, zap.NewNop()).TagWordCardCourse(1, "en_ru"); err == nil {
		t.Fatal("TagWordCardCourse")
	}
	if _, err := NewWordRepository(db, zap.NewNop()).GetWordFormMappingForCourse("form", "en_ru"); err == nil {
		t.Fatal("GetWordFormMappingForCourse")
	}
	if _, err := NewWordRepository(db, zap.NewNop()).ListRecentWordsPage("en_ru", 10, 0); err == nil {
		t.Fatal("ListRecentWordsPage")
	}
	if err := NewWordRepository(db, zap.NewNop()).MergeWordFormInto(ctx, 1, 2); err == nil {
		t.Fatal("MergeWordFormInto")
	}
	if _, err := NewWordRepository(db, zap.NewNop()).GetTTSAudioRelPath("en_ru", "word"); err == nil {
		t.Fatal("GetTTSAudioRelPath")
	}
	if err := NewWordRepository(db, zap.NewNop()).DeleteTTSStatus("en_ru", "word"); err == nil {
		t.Fatal("DeleteTTSStatus")
	}
	if _, err := NewWordSetCategoryRepository(db, zap.NewNop()).GetCategoryForCourse(1, "en_ru"); err == nil {
		t.Fatal("GetCategoryForCourse")
	}
	if _, err := NewWordSetCategoryRepository(db, zap.NewNop()).GetAllCategoriesForCourse("en_ru"); err == nil {
		t.Fatal("GetAllCategoriesForCourse")
	}
	if _, err := NewWordSetRepository(db, zap.NewNop()).ListWordSetsForCourse("en_ru", nil, 10, 0); err == nil {
		t.Fatal("ListWordSetsForCourse")
	}
}

func TestWordReposGap_ScanErrors(t *testing.T) {
	t.Run("GetUserIDsByWord scan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT .+ FROM word_cards").WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT DISTINCT user_id").
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("bad"))
		repo := NewWordRepository(db, zap.NewNop())
		if _, err := repo.GetUserIDsByWord("word"); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("GetUserIDsByWord with lemma scan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT .+ FROM word_cards").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "word", "definition", "pos", "noun_gender", "opposite_gender_word", "transcription",
				"definition_ru", "examples_json", "verb_forms_json", "display_en", "processed_at", "processing_error",
				"course_code", "created_at", "updated_at",
			}).AddRow(1, "lemma", "d", nil, nil, nil, nil, nil, nil, nil, nil, "", "", "en_ru", "2024-01-01", "2024-01-01"))
		mock.ExpectQuery("SELECT DISTINCT user_id").
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("bad"))
		repo := NewWordRepository(db, zap.NewNop())
		if _, err := repo.GetUserIDsByWord("lemma"); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("GetWordSetWords scan warn", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, category_id").WillReturnRows(sqlmock.NewRows([]string{
			"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "level_code", "created_at", "updated_at",
		}).AddRow(1, nil, "set", nil, 1, 1, nil, nil, "2024-01-01", "2024-01-01"))
		rows := sqlmock.NewRows([]string{
			"id", "word", "status", "display_word_pref", "transcription_pref", "word_ru_pref", "meaning_en_pref", "example_en_pref", "example_ru_pref",
		}).AddRow("bad", "w", "unknown", nil, nil, nil, nil, nil, nil).
			AddRow(2, "ok", "unknown", "ok", nil, nil, nil, nil, nil)
		mock.ExpectQuery("FROM word_set_items").WillReturnRows(rows)
		repo := NewWordSetRepository(db, zap.NewNop())
		words, err := repo.GetWordSetWords(1, 900001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(words) != 1 || words[0].Word != "ok" {
			t.Fatalf("words = %+v", words)
		}
	})

	t.Run("GetUserIDsByWordCardID scan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT DISTINCT user_id").
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("bad"))
		repo := NewWordRepository(db, zap.NewNop())
		if _, err := repo.GetUserIDsByWordCardID(1, "word"); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("ListRecentWordsPage scan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT wc.word").
			WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow(nil))
		repo := NewWordRepository(db, zap.NewNop())
		if _, err := repo.ListRecentWordsPage("en_ru", 10, 0); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("ListWordCardsAdminForCourse scan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{
			"id", "word", "definition", "pos", "noun_gender", "opposite_gender_word", "transcription",
			"definition_ru", "examples_json", "verb_forms_json", "display_en", "processed_at", "processing_error",
			"tts_state", "tts_error", "tts_audio_rel_path", "course_codes", "created_at", "updated_at", "has_training_cards",
		}).AddRow("bad", "w", "d", "", "", "", "", "", "", "", "", "", "", "", "", "", "en_ru", "2024-01-01", "2024-01-01", 0)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		repo := NewWordRepository(db, zap.NewNop())
		if _, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "", "desc"); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("GetAllCategoriesForCourse scan warn", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{
			"id", "course_code", "parent_id", "name", "description", "is_published", "sort_order", "level_code", "created_at", "updated_at",
		}).AddRow("bad", "en_ru", nil, "n", nil, 1, 1, nil, "2024-01-01", "2024-01-01").
			AddRow(1, "en_ru", nil, "ok", nil, 1, 1, nil, "2024-01-01", "2024-01-01")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		repo := NewWordSetCategoryRepository(db, zap.NewNop())
		cats, err := repo.GetAllCategoriesForCourse("en_ru")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cats) != 1 || cats[0].Name != "ok" {
			t.Fatalf("cats = %+v", cats)
		}
	})

	t.Run("ListWordSetsForCourse scan warn", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows := sqlmock.NewRows([]string{
			"id", "course_code", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "level_code", "created_at", "updated_at",
		}).AddRow("bad", "en_ru", nil, "t", nil, 1, 1, nil, nil, "2024-01-01", "2024-01-01").
			AddRow(1, "en_ru", nil, "ok", nil, 1, 1, nil, nil, "2024-01-01", "2024-01-01")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		repo := NewWordSetRepository(db, zap.NewNop())
		sets, err := repo.ListWordSetsForCourse("en_ru", nil, 10, 0, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sets) != 1 || sets[0].Title != "ok" {
			t.Fatalf("sets = %+v", sets)
		}
	})
}

func strPtr(s string) *string { return &s }

func TestWordReposGap_WordSetProgressErrors(t *testing.T) {
	db := wordReposGapInvalidDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())

	if _, err := setRepo.GetWordSetProgress(1, wordReposGapUserTelegramID); err == nil {
		t.Fatal("GetWordSetProgress invalid db")
	}
	if _, _, _, err := setRepo.GetCategoriesAggregateProgress([]int64{1}, wordReposGapUserTelegramID); err == nil {
		t.Fatal("GetCategoriesAggregateProgress invalid db")
	}

	realDB := wordReposGapDB(t)
	setRepo = NewWordSetRepository(realDB, zap.NewNop())
	if _, err := setRepo.GetWordSetProgress(999999003, 1); err == nil {
		t.Fatal("GetWordSetProgress missing set")
	}
}

func TestWordReposGap_GetUserIDsByWordErrors(t *testing.T) {
	db := wordReposGapInvalidDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	if _, err := repo.GetUserIDsByWord("word"); err == nil {
		t.Fatal("GetUserIDsByWord invalid db")
	}
	if _, err := repo.GetUserIDsByWordCardID(1, "word"); err == nil {
		t.Fatal("GetUserIDsByWordCardID invalid db")
	}
}

func TestWordReposGap_ListWordCardsAdminFilters(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+5)

	if err := repo.SaveWordCard("gap-admin-filter", "def", "en_ru"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemmaForCourse("gap-admin-filter", "en_ru")
	if err := repo.MarkWordCardProcessedError(card.ID, "gap error"); err != nil {
		t.Fatalf("MarkWordCardProcessedError: %v", err)
	}
	if err := repo.AddWordRequestHistoryWithCard(userID, "gap-admin-filter", &card.ID, &card.Word); err != nil {
		t.Fatalf("history: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO tts_generation_status (course_code, word, state, last_error_message)
		VALUES ('en_ru', 'gap-admin-filter', 'failed_retryable', 'tts err')
		ON CONFLICT (course_code, word) DO UPDATE SET state = 'failed_retryable', audio_rel_path = NULL, last_error_message = EXCLUDED.last_error_message
	`); err != nil {
		t.Fatalf("tts: %v", err)
	}

	items, err := repo.ListWordCardsAdminForCourse("en_ru", &userID, true, nil, "gap-admin", "noun", 10, 0, "id", "asc")
	if err != nil {
		t.Fatalf("ListWordCardsAdminForCourse filters: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected filtered admin items")
	}
	for _, it := range items {
		if it.Word == "gap-admin-filter" {
			if it.ProcessingError == nil || it.TTSError == nil {
				t.Fatalf("item = %+v", it)
			}
		}
	}

	hasAudio := true
	if _, err := db.Exec(`
		UPDATE tts_generation_status
		SET state = 'ready', audio_rel_path = 'ab/cd/gap-admin-filter.mp3'
		WHERE course_code = 'en_ru' AND word = 'gap-admin-filter'
	`); err != nil {
		t.Fatalf("tts ready: %v", err)
	}
	withAudio, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, &hasAudio, "gap-admin", "", 10, 0, "word", "desc")
	if err != nil {
		t.Fatalf("hasAudio filter: %v", err)
	}
	for _, it := range withAudio {
		if it.Word == "gap-admin-filter" && it.TTSAudioURL == nil {
			t.Fatalf("expected audio url on %+v", it)
		}
	}

	if _, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "pos", "asc"); err != nil {
		t.Fatalf("sort by pos: %v", err)
	}
	if _, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "invalid_sort", "desc"); err != nil {
		t.Fatalf("invalid sort: %v", err)
	}
}

func TestWordReposGap_MergeWordFormIntoCommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	execs := []string{
		"DELETE FROM word_set_items",
		"UPDATE word_set_items",
		"DELETE FROM user_word_knowledge",
		"UPDATE user_word_knowledge",
		"DELETE FROM user_word_mastering",
		"UPDATE user_word_mastering",
		"DELETE FROM learning_items",
		"UPDATE learning_items",
		"DELETE FROM word_cards",
	}
	for _, q := range execs {
		mock.ExpectExec(q).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	repo := NewWordRepository(db, zap.NewNop())
	if err := repo.MergeWordFormInto(context.Background(), 10, 20); err == nil {
		t.Fatal("expected commit error")
	}
}

func TestWordReposGap_GetCategoryForCourseNotFound(t *testing.T) {
	repo := NewWordSetCategoryRepository(wordReposGapDB(t), zap.NewNop())
	got, err := repo.GetCategoryForCourse(999999006, "en_ru")
	if err != nil || got != nil {
		t.Fatalf("GetCategoryForCourse missing = %+v err=%v", got, err)
	}
}

func TestWordReposGap_ListRecentWordsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT wc.word").WillReturnError(fmt.Errorf("query failed"))
	repo := NewWordRepository(db, zap.NewNop())
	if _, err := repo.ListRecentWords("en_ru", 5); err == nil {
		t.Fatal("expected query error")
	}
}

func TestWordReposGap_GetCategoryDBError(t *testing.T) {
	db := wordReposGapInvalidDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	if _, err := repo.GetCategory(1); err == nil {
		t.Fatal("GetCategory invalid db")
	}
}

func TestWordReposGap_GetWordSetProgressCountErrors(t *testing.T) {
	newSetRow := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{
			"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "level_code", "created_at", "updated_at",
		}).AddRow(1, nil, "set", nil, 1, 1, nil, nil, "2024-01-01", "2024-01-01")
	}

	t.Run("total count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, category_id").WillReturnRows(newSetRow())
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT word_card_id\\) FROM word_set_items").
			WillReturnError(fmt.Errorf("count total failed"))
		repo := NewWordSetRepository(db, zap.NewNop())
		if _, err := repo.GetWordSetProgress(1, 900001); err == nil {
			t.Fatal("expected total count error")
		}
	})

	t.Run("known count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, category_id").WillReturnRows(newSetRow())
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT word_card_id\\) FROM word_set_items").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnError(fmt.Errorf("known count failed"))
		repo := NewWordSetRepository(db, zap.NewNop())
		if _, err := repo.GetWordSetProgress(1, 900001); err == nil {
			t.Fatal("expected known count error")
		}
	})

	t.Run("vocab count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, category_id").WillReturnRows(newSetRow())
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT word_card_id\\) FROM word_set_items").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnError(fmt.Errorf("vocab count failed"))
		repo := NewWordSetRepository(db, zap.NewNop())
		if _, err := repo.GetWordSetProgress(1, 900001); err == nil {
			t.Fatal("expected vocab count error")
		}
	})

	t.Run("clamp unknown words", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, category_id").WillReturnRows(newSetRow())
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT word_card_id\\) FROM word_set_items").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		repo := NewWordSetRepository(db, zap.NewNop())
		progress, err := repo.GetWordSetProgress(1, 900001)
		if err != nil {
			t.Fatal(err)
		}
		if progress.UnknownWords != 0 {
			t.Fatalf("unknown = %d", progress.UnknownWords)
		}
	})
}

func TestWordReposGap_GetCategoriesAggregateProgressErrors(t *testing.T) {
	t.Run("known count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnError(fmt.Errorf("known count failed"))
		repo := NewWordSetRepository(db, zap.NewNop())
		if _, _, _, err := repo.GetCategoriesAggregateProgress([]int64{1}, 900001); err == nil {
			t.Fatal("expected known count error")
		}
	})

	t.Run("vocab count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT wsi.word_card_id\\)").
			WillReturnError(fmt.Errorf("vocab count failed"))
		repo := NewWordSetRepository(db, zap.NewNop())
		if _, _, _, err := repo.GetCategoriesAggregateProgress([]int64{1}, 900001); err == nil {
			t.Fatal("expected vocab count error")
		}
	})
}

func TestWordReposGap_GetWordSetProgressGetWordSetError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, category_id").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "level_code", "created_at", "updated_at",
		}).AddRow("bad", nil, "set", nil, 1, 1, nil, nil, "2024-01-01", "2024-01-01"))
	repo := NewWordSetRepository(db, zap.NewNop())
	if _, err := repo.GetWordSetProgress(1, 900001); err == nil {
		t.Fatal("expected get word set error")
	}
}

func TestWordReposGap_GetWordSetProgressEmptySet(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	catID, _ := NewWordSetCategoryRepository(db, zap.NewNop()).CreateCategory(&models.WordSetCategory{
		CourseCode: "en_ru", Name: "Gap Empty", IsPublished: true, SortOrder: 1,
	})
	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode: "en_ru", CategoryID: &catID, Title: "Empty", IsPublished: true, SortOrder: 1,
	})
	progress, err := setRepo.GetWordSetProgress(setID, wordReposGapUser(t, db, wordReposGapUserTelegramID+7))
	if err != nil {
		t.Fatal(err)
	}
	if progress.TotalWords != 0 || progress.ProgressPercent != 0 {
		t.Fatalf("empty progress = %+v", progress)
	}
}

func TestWordReposGap_GetWordSetWordsDisplayFallback(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+8)

	pos := "noun"
	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode: "en_ru", Title: "Display Fallback", IsPublished: true, SortOrder: 1, PreferredPOS: &pos,
	})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-fallback-word", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word)
		VALUES (?, 'gap-fallback-word', 0, 'слово', 'meaning', 'noun', '')
	`, cardID); err != nil {
		t.Fatal(err)
	}
	words, err := setRepo.GetWordSetWords(setID, userID)
	if err != nil || len(words) != 1 || words[0].DisplayWord != "gap-fallback-word" {
		t.Fatalf("words = %+v err=%v", words, err)
	}
}

func TestWordReposGap_GetWordSetWordsKnownStatus(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+9)

	setID, _ := setRepo.CreateWordSet(&models.WordSet{CourseCode: "en_ru", Title: "Known Set", IsPublished: true, SortOrder: 1})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-known-status", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, userID, cardID); err != nil {
		t.Fatal(err)
	}
	words, err := setRepo.GetWordSetWords(setID, userID)
	if err != nil || len(words) != 1 || words[0].Status != "known" {
		t.Fatalf("words = %+v err=%v", words, err)
	}
}

func TestWordReposGap_GetWordSetWordsInVocabStatus(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+10)

	setID, _ := setRepo.CreateWordSet(&models.WordSet{CourseCode: "en_ru", Title: "Vocab Set", IsPublished: true, SortOrder: 1})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-vocab-status", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatal(err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES (?, 'gap-vocab-status', 0, 'слово', 'meaning', 'noun') RETURNING id
	`, cardID).Scan(&trainingCardID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'en_ru', 'new', 2.5)`, userID, trainingCardID); err != nil {
		t.Fatal(err)
	}
	words, err := setRepo.GetWordSetWords(setID, userID)
	if err != nil || len(words) != 1 || words[0].Status != "in_vocab" {
		t.Fatalf("words = %+v err=%v", words, err)
	}
}

func TestWordReposGap_GetCategoryISOFormatTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, parent_id").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "description", "is_published", "sort_order", "level_code", "created_at", "updated_at",
		}).AddRow(1, nil, "ISO Date", nil, 1, 1, nil, "2024-06-01T12:00:00", "2024-06-01T12:00:00"))
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	got, err := repo.GetCategory(1)
	if err != nil || got == nil || got.CreatedAt.IsZero() {
		t.Fatalf("category = %+v err=%v", got, err)
	}
}

func TestWordReposGap_GetCategoryUnparseableCreatedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, parent_id").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "description", "is_published", "sort_order", "level_code", "created_at", "updated_at",
		}).AddRow(1, nil, "Bad Date", nil, 1, 1, nil, "not-a-valid-date", "also-bad"))
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	got, err := repo.GetCategory(1)
	if err != nil || got == nil || !got.CreatedAt.IsZero() || !got.UpdatedAt.IsZero() {
		t.Fatalf("category = %+v err=%v", got, err)
	}
}

func TestWordReposGap_ListWordCardsAdminEmptyCourseCodes(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	if err := repo.SaveWordCard("gap-no-course-code", "def", ""); err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "gap-no-course", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.Word == "gap-no-course-code" && len(it.CourseCodes) != 0 {
			t.Fatalf("course codes = %v", it.CourseCodes)
		}
	}
}

func TestWordReposGap_GetCategoryForCourseFull(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	parentID, _ := repo.CreateCategory(&models.WordSetCategory{CourseCode: "en_ru", Name: "Gap Parent Full", SortOrder: 0})
	desc := "full desc"
	level := "B2"
	id, err := repo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		ParentID:    &parentID,
		Name:        "Gap Full Course Cat",
		Description: &desc,
		IsPublished: true,
		SortOrder:   1,
		LevelCode:   &level,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetCategoryForCourse(id, "en_ru")
	if err != nil || got == nil || got.ParentID == nil || got.Description == nil || got.LevelCode == nil {
		t.Fatalf("full category = %+v err=%v", got, err)
	}
}

func TestWordReposGap_GetCategoryUnpublishedWithParent(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	parentID, err := repo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		Name:        "Gap Parent",
		IsPublished: true,
		SortOrder:   0,
	})
	if err != nil {
		t.Fatal(err)
	}
	childID, err := repo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "en_ru",
		ParentID:    &parentID,
		Name:        "Gap Child",
		IsPublished: false,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetCategory(childID)
	if err != nil || got == nil || got.ParentID == nil || got.IsPublished {
		t.Fatalf("GetCategory child = %+v err=%v", got, err)
	}
}

func TestWordReposGap_GetWordSetWordsNoPreferredPOS(t *testing.T) {
	db := wordReposGapDB(t)
	setRepo := NewWordSetRepository(db, zap.NewNop())
	wordRepo := NewWordRepository(db, zap.NewNop())
	userID := wordReposGapUser(t, db, wordReposGapUserTelegramID+6)

	setID, _ := setRepo.CreateWordSet(&models.WordSet{
		CourseCode:  "en_ru",
		Title:       "Gap Plain Set",
		IsPublished: true,
		SortOrder:   1,
	})
	cardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "gap-plain-word", Definition: "def"})
	if err := setRepo.SetWordSetItems(setID, []int64{cardID}); err != nil {
		t.Fatal(err)
	}
	words, err := setRepo.GetWordSetWords(setID, userID)
	if err != nil || len(words) != 1 || words[0].DisplayWord != "gap-plain-word" {
		t.Fatalf("words = %+v err=%v", words, err)
	}
}

func TestWordReposGap_ListWordCardsAdminRequestingUsersWarn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{
		"id", "word", "definition", "pos", "noun_gender", "opposite_gender_word", "transcription",
		"definition_ru", "examples_json", "verb_forms_json", "display_en", "processed_at", "processing_error",
		"tts_state", "tts_error", "tts_audio_rel_path", "course_codes", "created_at", "updated_at", "has_training_cards",
	}).AddRow(1, "warnword", "d", "", "", "", "", "", "", "", "", "", "", "", "", "", "en_ru", "2024-01-01", "2024-01-01", 0)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectQuery("SELECT .+ FROM word_cards").WillReturnError(fmt.Errorf("lemma lookup failed"))
	repo := NewWordRepository(db, zap.NewNop())
	items, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || len(items[0].RequestingUsers) != 0 {
		t.Fatalf("items = %+v", items)
	}
}

func TestWordReposGap_GetUserIDsByWordQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT .+ FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "word", "definition", "pos", "noun_gender", "opposite_gender_word", "transcription",
			"definition_ru", "examples_json", "verb_forms_json", "display_en", "processed_at", "processing_error",
			"course_code", "created_at", "updated_at",
		}).AddRow(1, "lemma", "d", nil, nil, nil, nil, nil, nil, nil, nil, "", "", "en_ru", "2024-01-01", "2024-01-01"))
	mock.ExpectQuery("SELECT DISTINCT user_id").WillReturnError(fmt.Errorf("query failed"))
	repo := NewWordRepository(db, zap.NewNop())
	if _, err := repo.GetUserIDsByWord("lemma"); err == nil {
		t.Fatal("expected query error")
	}
}

func TestWordReposGap_GetCategoryForCourseMinimal(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	id, err := repo.CreateCategory(&models.WordSetCategory{
		CourseCode:  "es_ru",
		Name:        "Gap Minimal Cat",
		IsPublished: false,
		SortOrder:   3,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetCategoryForCourse(id, "es_ru")
	if err != nil || got == nil || got.ParentID != nil || got.Description != nil || got.LevelCode != nil {
		t.Fatalf("minimal category = %+v err=%v", got, err)
	}
}

func TestWordReposGap_ListWordCardsAdminDefaultSort(t *testing.T) {
	db := wordReposGapDB(t)
	repo := NewWordRepository(db, zap.NewNop())
	if err := repo.SaveWordCard("gap-default-sort", "def", "en_ru"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	if _, err := repo.ListWordCardsAdminForCourse("en_ru", nil, false, nil, "", "", 10, 0, "word", ""); err != nil {
		t.Fatalf("ListWordCardsAdminForCourse default sort: %v", err)
	}
}
