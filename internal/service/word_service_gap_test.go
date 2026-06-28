//go:build test

package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const wordServiceGapUserID int64 = 900001

func wordServiceGapSpanishLearning() config.LearningConfig {
	return config.LearningConfig{
		Pair:            "ru-es",
		NativeLang:      "ru",
		TargetLang:      "es",
		AppCode:         "spanish",
		GrammarBundleID: "es",
		ContentSource:   "bundle",
	}
}

func wordServiceGapSetField(t *testing.T, svc *WordService, field string, value any) {
	t.Helper()
	v := reflect.ValueOf(svc).Elem().FieldByName(field)
	if !v.IsValid() {
		t.Fatalf("field %q not found on WordService", field)
	}
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func wordServiceGapSetLexicon(t *testing.T, svc *WordService, lex map[string]models.SpanishGenderLexiconEntry) {
	t.Helper()
	wordServiceGapSetField(t, svc, "spanishGenderLexicon", lex)
}

func wordServiceGapSpanishService(t *testing.T, wordRepo *repository.WordRepository, aiService *ai.Service) *WordService {
	t.Helper()
	svc := NewWordServiceWithMastering(wordRepo, nil, nil, nil, aiService, wordServiceGapSpanishLearning(), zap.NewNop())
	if len(svc.spanishGenderLexicon) == 0 {
		wordServiceGapSetLexicon(t, svc, map[string]models.SpanishGenderLexiconEntry{
			"gato": {Lemma: "gato", Gender: "m", OppositeGenderWord: "gata"},
			"actor": {Lemma: "actor", Gender: "m", OppositeGenderWord: "actriz"},
		})
	}
	return svc
}

func newAIServiceWithPromptResponses(t *testing.T, logger *zap.Logger, byNeedle map[string]string, fallback string) *ai.Service {
	t.Helper()
	aiService := ai.NewService("http://example.com", "model", "key", "prompt", logger)
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			bodyBytes, _ := io.ReadAll(r.Body)
			var payload struct {
				Messages []struct {
					Content string `json:"content"`
				} `json:"messages"`
			}
			_ = json.Unmarshal(bodyBytes, &payload)
			userMsg := ""
			for _, m := range payload.Messages {
				if strings.TrimSpace(m.Content) != "" {
					userMsg = m.Content
				}
			}
			content := fallback
			for needle, resp := range byNeedle {
				if strings.Contains(userMsg, needle) {
					content = resp
					break
				}
			}
			out := ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Role: "assistant", Content: content}}},
			}
			body, _ := json.Marshal(out)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}
	aiValue := reflect.ValueOf(aiService).Elem()
	clientField := aiValue.FieldByName("client")
	reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(mockClient))
	return aiService
}

func wordServiceGapCardColumns() []string {
	return []string{
		"id", "word", "definition", "pos", "noun_gender", "opposite_gender_word", "transcription", "definition_ru",
		"examples_json", "verb_forms_json", "display_en", "processed_at", "processing_error", "course_code", "created_at", "updated_at",
	}
}

func wordServiceGapCardRow(id int64, word, defRU, course string, nounGender, opposite *string) *sqlmock.Rows {
	return sqlmock.NewRows(wordServiceGapCardColumns()).AddRow(
		id, word, "", nil, nounGender, opposite, nil, defRU, nil, nil, nil, "", "", course, "2024-01-01", "2024-01-01",
	)
}

func wordServiceGapExpectLemmaNotFound(mock sqlmock.Sqlmock, lemma, course string) {
	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs(lemma, course).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`FROM word_forms wf\s+JOIN word_cards wc ON wc\.id = wf\.word_card_id`).
		WithArgs(lemma, course).
		WillReturnRows(sqlmock.NewRows([]string{"form", "word_card_id"}))
}

func wordServiceGapExpectAISave(mock sqlmock.Sqlmock, lemma, course string, cardID int64) {
	mock.ExpectExec(`INSERT INTO word_cards`).
		WillReturnResult(sqlmock.NewResult(cardID, 1))
	mock.ExpectExec(`INSERT INTO word_forms`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO word_request_history`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func wordServiceGapWriteLexicon(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "lexicon.tsv")
	content := "lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes\nperro\tm\tel perro\tla perra\ttest\ttest\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write lexicon: %v", err)
	}
	return path
}

func TestWordServiceGap_buildForcedNativeLanguagePrompt(t *testing.T) {
	got := buildForcedNativeLanguagePrompt("hola", "ru", "es", "ru-es")
	for _, want := range []string{"ru-es", "ru", "es", "hola", "definition_native MUST be in ru"} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q:\n%s", want, got)
		}
	}
}

func TestWordServiceGap_hasCyrillicText(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"", false},
		{"  ", false},
		{"hello", false},
		{"привет", true},
		{" mix кириллица ", true},
	}
	for _, tt := range tests {
		if got := hasCyrillicText(tt.text); got != tt.want {
			t.Fatalf("hasCyrillicText(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestWordServiceGap_normalizeNounGender(t *testing.T) {
	tests := []struct {
		pos  string
		raw  string
		want *string
	}{
		{"verb", "m", nil},
		{"noun", "f", strPtr("f")},
		{"sustantivo femenino", "", strPtr("f")},
		{"noun", "invalid", nil},
	}
	for _, tt := range tests {
		got := normalizeNounGender(tt.pos, tt.raw)
		if tt.want == nil {
			if got != nil {
				t.Fatalf("normalizeNounGender(%q,%q) = %v, want nil", tt.pos, tt.raw, *got)
			}
			continue
		}
		if got == nil || *got != *tt.want {
			t.Fatalf("normalizeNounGender(%q,%q) = %v, want %v", tt.pos, tt.raw, got, *tt.want)
		}
	}
}

func TestWordServiceGap_canonicalPOSValue(t *testing.T) {
	if got := canonicalPOSValue("Noun"); got != "noun" {
		t.Fatalf("canonical known POS = %q", got)
	}
	if got := canonicalPOSValue("particle"); got != "particle" {
		t.Fatalf("canonical unknown POS = %q", got)
	}
}

func TestWordServiceGap_nativeFieldsLookValidForConfig(t *testing.T) {
	tests := []struct {
		name     string
		def      string
		examples []models.WordInfoExample
		require  bool
		want     bool
	}{
		{"no cyrillic required", "hello", nil, false, true},
		{"missing cyrillic definition", "hello", nil, true, false},
		{"valid cyrillic", "кот", nil, true, true},
		{"latin gloss fails", "кот", []models.WordInfoExample{{GlossNative: "cat"}}, true, false},
		{"cyrillic gloss native", "кот", []models.WordInfoExample{{GlossNative: "кошка"}}, true, true},
		{"cyrillic gloss ru fallback", "кот", []models.WordInfoExample{{GlossRU: "кошка"}}, true, true},
		{"empty gloss ok", "кот", []models.WordInfoExample{{GlossNative: ""}}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nativeFieldsLookValidForConfig(tt.def, tt.examples, tt.require)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestWordServiceGap_SetVerbFormCardsSync_and_tryVerbFormCardsSync(t *testing.T) {
	logger := zap.NewNop()
	svc := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)

	var calledWith int64
	svc.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: true}, func(userID int64) error {
		calledWith = userID
		return nil
	})
	svc.tryVerbFormCardsSync(wordServiceGapUserID)
	if calledWith != wordServiceGapUserID {
		t.Fatalf("hook called with %d, want %d", calledWith, wordServiceGapUserID)
	}

	svc.tryVerbFormCardsSync(wordServiceGapUserID + 1) // nil hook path via fresh service
	nilHook := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)
	nilHook.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: true}, nil)
	nilHook.tryVerbFormCardsSync(wordServiceGapUserID)

	disabled := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)
	disabled.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: false}, func(int64) error {
		t.Fatal("hook should not run when disabled")
		return nil
	})
	disabled.tryVerbFormCardsSync(wordServiceGapUserID)

	english := NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), logger)
	english.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: true}, func(int64) error {
		t.Fatal("hook should not run for English target")
		return nil
	})
	english.tryVerbFormCardsSync(wordServiceGapUserID)

	var warnCalled bool
	warnSvc := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)
	warnSvc.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: true}, func(int64) error {
		return fmt.Errorf("sync failed")
	})
	warnSvc.tryVerbFormCardsSync(wordServiceGapUserID)
	_ = warnCalled
}

func TestWordServiceGap_NewWordServiceWithMastering(t *testing.T) {
	t.Setenv("SPANISH_GENDER_LEXICON_PATH", wordServiceGapWriteLexicon(t))
	logger := zap.NewNop()
	svc := NewWordServiceWithMastering(nil, nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)
	if svc == nil {
		t.Fatal("expected service")
	}
	if len(svc.spanishGenderLexicon) == 0 {
		t.Fatal("expected lexicon loaded from custom path")
	}

	t.Setenv("SPANISH_GENDER_LEXICON_PATH", filepath.Join(t.TempDir(), "missing.tsv"))
	svcMissing := NewWordServiceWithMastering(nil, nil, nil, nil, nil, wordServiceGapSpanishLearning(), logger)
	if len(svcMissing.spanishGenderLexicon) != 0 {
		t.Fatal("expected empty lexicon when custom path missing")
	}
	_ = NewWordServiceWithMastering(nil, nil, nil, nil, nil, wordServiceGapSpanishLearning(), nil)
}

func TestWordServiceGap_applySpanishGenderLexicon(t *testing.T) {
	svc := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), zap.NewNop())
	wordServiceGapSetLexicon(t, svc, map[string]models.SpanishGenderLexiconEntry{
		"gato": {Gender: "m", OppositeGenderWord: "gata"},
	})

	svc.applySpanishGenderLexicon("gato", nil)

	empty := &models.WordInfoResponse{}
	svc.applySpanishGenderLexicon("", empty)
	if empty.POS != "" {
		t.Fatal("empty lemma should not mutate")
	}

	noLex := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), zap.NewNop())
	wordServiceGapSetLexicon(t, noLex, nil)
	wi := &models.WordInfoResponse{}
	noLex.applySpanishGenderLexicon("gato", wi)
	if wi.POS != "" {
		t.Fatal("empty lexicon should not apply")
	}

	unknown := &models.WordInfoResponse{}
	svc.applySpanishGenderLexicon("unknown", unknown)
	if unknown.POS != "" {
		t.Fatal("unknown lemma should not apply")
	}

	noun := &models.WordInfoResponse{POS: "noun"}
	svc.applySpanishGenderLexicon("gato", noun)
	if noun.POS != "noun" || noun.NounGender != "m" || noun.OppositeGenderWord != "gata" {
		t.Fatalf("noun apply = %+v", noun)
	}

	blankPOS := &models.WordInfoResponse{}
	svc.applySpanishGenderLexicon("gato", blankPOS)
	if blankPOS.POS != "noun" || blankPOS.NounGender != "m" {
		t.Fatalf("blank POS apply = %+v", blankPOS)
	}

	verb := &models.WordInfoResponse{POS: "verb"}
	svc.applySpanishGenderLexicon("gato", verb)
	if verb.POS != "verb" || verb.NounGender != "" {
		t.Fatalf("verb POS should stay verb: %+v", verb)
	}
}

func TestWordServiceGap_normalizeOppositeGenderWord(t *testing.T) {
	svc := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), zap.NewNop())
	wordServiceGapSetLexicon(t, svc, map[string]models.SpanishGenderLexiconEntry{
		"actor": {OppositeGenderWord: "actriz"},
	})

	if got := svc.normalizeOppositeGenderWord("noun", "gata", "gato", "en"); got != nil {
		t.Fatal("non-es target should return nil")
	}
	if got := svc.normalizeOppositeGenderWord("verb", "gata", "gato", "es"); got != nil {
		t.Fatal("non-noun should return nil")
	}
	if got := svc.normalizeOppositeGenderWord("noun", "", "gato", "es"); got != nil {
		t.Fatal("empty raw should return nil")
	}
	if got := svc.normalizeOppositeGenderWord("noun", "gato", "gato", "es"); got != nil {
		t.Fatal("same lemma/opposite should return nil")
	}

	simple := svc.normalizeOppositeGenderWord("noun", "gata", "gato", "es")
	if simple == nil || *simple != "gata" {
		t.Fatalf("simple pair = %v", simple)
	}

	lexicon := svc.normalizeOppositeGenderWord("noun", "actriz", "actor", "es")
	if lexicon == nil || *lexicon != "actriz" {
		t.Fatalf("lexicon pair = %v", lexicon)
	}

	if got := svc.normalizeOppositeGenderWord("noun", "actriz", "gato", "es"); got != nil {
		t.Fatal("invalid opposite should return nil")
	}
}

func TestWordServiceGap_wordCardNativeFieldsLookValid(t *testing.T) {
	es := NewWordService(nil, nil, nil, nil, wordServiceGapSpanishLearning(), zap.NewNop())
	if es.wordCardNativeFieldsLookValid(nil) {
		t.Fatal("nil card should be invalid")
	}

	defRU := "кот"
	if !es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &defRU}) {
		t.Fatal("cyrillic DefinitionRU should be valid")
	}

	defNative := "собака"
	if !es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionNative: &defNative}) {
		t.Fatal("cyrillic DefinitionNative should be valid")
	}

	latin := "cat"
	if es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &latin}) {
		t.Fatal("latin-only definition should be invalid for es_ru")
	}

	examplesJSON := `[{"gloss_native":"кошка"}]`
	if !es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &defRU, ExamplesJSON: &examplesJSON}) {
		t.Fatal("valid examples should pass")
	}
	badExamplesJSON := `[{"gloss_native":"cat"}]`
	if es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &defRU, ExamplesJSON: &badExamplesJSON}) {
		t.Fatal("latin gloss in examples should fail")
	}
	_ = es.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &defRU, ExamplesJSON: strPtr("not-json")})

	en := NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), zap.NewNop())
	if !en.wordCardNativeFieldsLookValid(&models.WordCard{DefinitionRU: &latin}) {
		t.Fatal("english deployment should not require cyrillic")
	}
}

func TestWordServiceGap_GetWordDefinitionForCourse_EmptyCourseCode(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs("gaphello", "en_ru").
		WillReturnRows(wordServiceGapCardRow(11, "gaphello", "приветствие", "en_ru", nil, nil))
	mock.ExpectExec(`INSERT INTO word_request_history`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), zap.NewNop())
	resp, err := svc.GetWordDefinitionForCourse(context.Background(), wordServiceGapUserID, "gaphello", "  ")
	if err != nil {
		t.Fatalf("GetWordDefinitionForCourse: %v", err)
	}
	if !strings.Contains(resp, "gaphello") {
		t.Fatalf("expected default-course lookup, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_GetWordDefinitionForCourse_TargetLangFallback(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordServiceGapExpectLemmaNotFound(mock, "gapword", "_only")
	wordServiceGapExpectAISave(mock, "gapword", "_only", 42)

	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	aiSvc := newAIServiceWithResponse(t, zap.NewNop(), `{"lemma":"gapword","pos":"noun","definition_ru":"тест"}`)
	svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())

	resp, err := svc.GetWordDefinitionForCourse(context.Background(), wordServiceGapUserID, "gapword", "_only")
	if err != nil {
		t.Fatalf("GetWordDefinitionForCourse: %v", err)
	}
	if !strings.Contains(resp, "тест") {
		t.Fatalf("expected saved definition, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_GetWordDefinitionForCourse_ScopedLookup(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs("gapalgo", "es_ru").
		WillReturnRows(wordServiceGapCardRow(21, "gapalgo", "что-то", "es_ru", nil, nil))
	mock.ExpectExec(`INSERT INTO word_request_history`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), zap.NewNop())
	resp, err := svc.GetWordDefinitionForCourse(context.Background(), wordServiceGapUserID+1, "gapalgo", "es_ru")
	if err != nil {
		t.Fatalf("GetWordDefinitionForCourse: %v", err)
	}
	if !strings.Contains(resp, "что-то") {
		t.Fatalf("expected es definition, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_SpanishForcedNativeLanguageRetry(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordServiceGapExpectLemmaNotFound(mock, "perro", "es_ru")
	wordServiceGapExpectAISave(mock, "perro", "es_ru", 31)

	first := `{"lemma":"perro","pos":"noun","definition_native":"dog","examples":[{"example_target":"El perro ladra.","gloss_native":"The dog barks."}]}`
	second := `{"lemma":"perro","pos":"noun","definition_native":"собака","examples":[{"example_target":"El perro ladra.","gloss_native":"Собака лает."}]}`
	aiSvc := newAIServiceWithPromptResponses(t, zap.NewNop(), map[string]string{
		"Critical language constraint": second,
	}, first)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, aiSvc)

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+2, "perro")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "собака") {
		t.Fatalf("expected retried Russian definition, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_SpanishNativeFieldsRejected(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordServiceGapExpectLemmaNotFound(mock, "perroreject", "es_ru")
	englishOnly := `{"lemma":"perroreject","pos":"noun","definition_native":"dog","examples":[{"example_target":"x","gloss_native":"y"}]}`
	aiSvc := newAIServiceWithResponse(t, zap.NewNop(), englishOnly)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, aiSvc)

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+3, "perroreject")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "русском") {
		t.Fatalf("expected rejection message, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_DBInvalidNativeFieldsForcesAIRefresh(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs("gapperro", "es_ru").
		WillReturnRows(wordServiceGapCardRow(41, "gapperro", "dog", "es_ru", nil, nil))
	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma = \? AND language = \?`).
		WithArgs("gapperro", "es").
		WillReturnError(sql.ErrNoRows)
	wordServiceGapExpectAISave(mock, "gapperro", "es_ru", 41)

	aiResp := `{"lemma":"gapperro","pos":"noun","definition_native":"собака"}`
	aiSvc := newAIServiceWithResponse(t, zap.NewNop(), aiResp)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, aiSvc)

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+4, "gapperro")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "собака") {
		t.Fatalf("expected refreshed Russian card, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_SpanishGenderLexiconOnSave(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordServiceGapExpectLemmaNotFound(mock, "gato", "es_ru")
	wordServiceGapExpectAISave(mock, "gato", "es_ru", 51)

	aiResp := `{"lemma":"gato","pos":"","definition_native":"кот","opposite_gender_word":"gata"}`
	aiSvc := newAIServiceWithResponse(t, zap.NewNop(), aiResp)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, aiSvc)

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+5, "gato")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "кот") {
		t.Fatalf("expected definition in response, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_tryVerbFormCardsSyncOnDBHit(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs("gapsync", "es_ru").
		WillReturnRows(wordServiceGapCardRow(61, "gapsync", "кот", "es_ru", nil, nil))
	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma = \? AND language = \?`).
		WithArgs("gapsync", "es").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO word_request_history`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var synced int64
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, nil)
	svc.SetVerbFormCardsSync(config.TrainingConfig{SpanishVerbFormsEnabled: true}, func(uid int64) error {
		synced = uid
		return nil
	})

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+6, "gapsync")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if resp == "" {
		t.Fatal("expected markdown response")
	}
	if synced != wordServiceGapUserID+6 {
		t.Fatalf("verb form sync user = %d, want %d", synced, wordServiceGapUserID+6)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_tryLinkVerbLemmaForLang(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := wordServiceGapSpanishService(t, wordRepo, nil)

	svc.tryLinkVerbLemmaForLang(nil, "es")
	svc.tryLinkVerbLemmaForLang(&models.WordCard{Word: "  "}, "es")
	svc.tryLinkVerbLemmaForLang(&models.WordCard{Word: "hablar", ID: 1}, "en")

	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma = \? AND language = \?`).
		WithArgs("hablar", "es").
		WillReturnError(sql.ErrNoRows)
	svc.tryLinkVerbLemmaForLang(&models.WordCard{Word: "hablar", ID: 99}, "es")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_courseTargetLang(t *testing.T) {
	tests := []struct {
		course string
		want   string
	}{
		{"es_ru", "es"},
		{"en_ru", "en"},
		{"_only", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := courseTargetLang(tt.course); got != tt.want {
			t.Fatalf("courseTargetLang(%q) = %q, want %q", tt.course, got, tt.want)
		}
	}
}

func TestWordServiceGap_renderWordCardMarkdown_EmptyTargetLang(t *testing.T) {
	def := "тест"
	svc := NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), zap.NewNop())
	md := svc.renderWordCardMarkdown(&models.WordCard{Word: "gaprender", DefinitionRU: &def}, "")
	if md == "" {
		t.Fatal("expected markdown with default target lang fallback")
	}
}

func TestWordServiceGap_getWordDefinitionForCourse_WordFormGetWordCardByIDError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM word_cards\s+WHERE LOWER\(word\) = LOWER\(\?\) AND course_code IS NOT DISTINCT FROM \?`).
		WithArgs("gapform", "en_ru").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`FROM word_forms wf\s+JOIN word_cards wc ON wc\.id = wf\.word_card_id`).
		WithArgs("gapform", "en_ru").
		WillReturnRows(sqlmock.NewRows([]string{"form", "word_card_id"}).AddRow("gapform", 77))
	mock.ExpectQuery(`FROM word_cards\s+WHERE id = \?`).
		WithArgs(int64(77)).
		WillReturnError(fmt.Errorf("card gone"))
	wordServiceGapExpectAISave(mock, "gapform", "en_ru", 88)

	aiSvc := newAIServiceWithResponse(t, zap.NewNop(), `{"lemma":"gapform","pos":"noun","definition_ru":"форма"}`)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())
	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+7, "gapform")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "форма") {
		t.Fatalf("expected AI fallback, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_getWordDefinitionForCourse_ForcedJSONRetrySuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wordServiceGapExpectLemmaNotFound(mock, "gapjson", "en_ru")
	wordServiceGapExpectAISave(mock, "gapjson", "en_ru", 91)

	aiSvc := newAIServiceWithPromptResponses(t, zap.NewNop(), map[string]string{
		"SINGLE-WORD LOOKUP ONLY": `{"lemma":"gapjson","pos":"noun","definition_ru":"json ok"}`,
	}, "plain text answer")
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())

	resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+8, "gapjson")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "json ok") {
		t.Fatalf("expected forced JSON retry result, got %q", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWordServiceGap_getWordDefinitionForCourse_ForcedJSONRetryFailures(t *testing.T) {
	t.Run("retry AI error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		wordServiceGapExpectLemmaNotFound(mock, "gapjsonfail", "en_ru")
		mock.ExpectExec(`INSERT INTO word_cards`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		aiSvc := ai.NewService("http://example.com", "model", "key", "prompt", zap.NewNop())
		call := 0
		mockClient := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			call++
			if call == 1 {
				body, _ := json.Marshal(ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: "not json"}}}})
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(body)), Header: http.Header{"Content-Type": []string{"application/json"}}}, nil
			}
			return nil, fmt.Errorf("forced retry unavailable")
		})}
		aiValue := reflect.ValueOf(aiSvc).Elem()
		clientField := aiValue.FieldByName("client")
		reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(mockClient))

		wordRepo := repository.NewWordRepository(db, zap.NewNop())
		svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())
		resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+9, "gapjsonfail")
		if err != nil {
			t.Fatalf("GetWordDefinition: %v", err)
		}
		if resp != "not json" {
			t.Fatalf("expected legacy save response, got %q", resp)
		}
	})

	t.Run("retry still not JSON", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		wordServiceGapExpectLemmaNotFound(mock, "gapjsonbad", "en_ru")
		mock.ExpectExec(`INSERT INTO word_cards`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		aiSvc := newAIServiceWithPromptResponses(t, zap.NewNop(), map[string]string{
			"SINGLE-WORD LOOKUP ONLY": "still not json",
		}, "still not json")
		wordRepo := repository.NewWordRepository(db, zap.NewNop())
		svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())
		resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+10, "gapjsonbad")
		if err != nil {
			t.Fatalf("GetWordDefinition: %v", err)
		}
		if resp != "still not json" {
			t.Fatalf("expected legacy response, got %q", resp)
		}
	})
}

func TestWordServiceGap_getWordDefinitionForCourse_LemmaCommaAndSpanishVerb(t *testing.T) {
	t.Run("lemma comma fallback", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		wordServiceGapExpectLemmaNotFound(mock, "gapcomma", "en_ru")
		wordServiceGapExpectAISave(mock, "gapcomma", "en_ru", 101)

		aiSvc := newAIServiceWithResponse(t, zap.NewNop(), `{"lemma":"foo,bar","pos":"noun","definition_ru":"запятая"}`)
		wordRepo := repository.NewWordRepository(db, zap.NewNop())
		svc := NewWordService(wordRepo, nil, nil, aiSvc, config.DefaultLearningConfig(), zap.NewNop())
		resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+11, "gapcomma")
		if err != nil {
			t.Fatalf("GetWordDefinition: %v", err)
		}
		if !strings.Contains(resp, "запятая") {
			t.Fatalf("expected saved card, got %q", resp)
		}
	})

	t.Run("spanish verb display without to prefix", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		wordServiceGapExpectLemmaNotFound(mock, "hablar", "es_ru")
		wordServiceGapExpectAISave(mock, "hablar", "es_ru", 102)

		aiResp := `{"lemma":"hablar","pos":"verb","definition_native":"говорить","verb_forms":{"v1":"hablar","v2":"habló","v3":"hablado"}}`
		wordRepo := repository.NewWordRepository(db, zap.NewNop())
		svc := wordServiceGapSpanishService(t, wordRepo, newAIServiceWithResponse(t, zap.NewNop(), aiResp))
		resp, err := svc.GetWordDefinition(context.Background(), wordServiceGapUserID+12, "hablar")
		if err != nil {
			t.Fatalf("GetWordDefinition: %v", err)
		}
		if strings.Contains(resp, "to hablar") {
			t.Fatalf("Spanish verb should not use English to-prefix, got %q", resp)
		}
		if !strings.Contains(resp, "говорить") {
			t.Fatalf("expected Russian definition, got %q", resp)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func strPtr(s string) *string { return &s }
