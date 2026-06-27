package service

import (
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestIsSpanishVerbTrainingScope(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"es.presente.indicativo", true},
		{"es.pretérito.indicativo", true},
		{"es.grammar.past_preterito.foo", false},
		{"es.grammar.orientation_alphabet", false},
		{"en.presente.indicativo", false},
		{"es.presente", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isSpanishVerbTrainingScope(tt.in); got != tt.want {
			t.Errorf("isSpanishVerbTrainingScope(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestResolveVerbScopes_grammarBundleStageFallsBackToDefault(t *testing.T) {
	g := "es.grammar.past_preterito_perfecto.haber_plus_participio"
	settings := &models.UserSettings{GrammarStage: &g}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	want := models.DefaultSpanishVerbScopes()
	if len(got) != len(want) {
		t.Fatalf("ResolveVerbScopes: got %v, want %v", got, want)
	}
	if len(want) > 0 && got[0] != want[0] {
		t.Fatalf("ResolveVerbScopes: got %v, want %v", got, want)
	}
}

func TestResolveVerbScopes_validGrammarStageUsed(t *testing.T) {
	g := "es.presente.indicativo"
	settings := &models.UserSettings{GrammarStage: &g}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	if len(got) != 1 || got[0] != "es.presente.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestEnsureVerbFormUserCards_DisabledIsNoOp(t *testing.T) {
	repo := &repository.VerbFormsRepository{}
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "en"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, zap.NewNop())
	if err := svc.EnsureVerbFormUserCards(1, []string{"es.presente.indicativo"}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveVerbScopes_enabledVerbScopesWin(t *testing.T) {
	g := "es.grammar.foo"
	settings := &models.UserSettings{
		GrammarStage:      &g,
		EnabledVerbScopes: []string{"es.futuro.indicativo"},
	}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	if len(got) != 1 || got[0] != "es.futuro.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveVerbScopes_enabledVerbScopesFiltersJunkKeepsValid(t *testing.T) {
	settings := &models.UserSettings{
		EnabledVerbScopes: []string{"es.grammar.bad", "es.presente.indicativo"},
	}
	got := ResolveVerbScopes(settings, config.LearningConfig{TargetLang: "es"})
	if len(got) != 1 || got[0] != "es.presente.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveVerbScopes_onlyInvalidEnabledScopesFallsBack(t *testing.T) {
	g := "es.grammar.past_preterito.foo"
	settings := &models.UserSettings{
		GrammarStage:      &g,
		EnabledVerbScopes: []string{"es.grammar.orientation", "es.grammar.past_preterito.foo"},
	}
	got := ResolveVerbScopes(settings, config.LearningConfig{TargetLang: "es"})
	want := models.DefaultSpanishVerbScopes()
	if len(got) != len(want) || (len(want) > 0 && got[0] != want[0]) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestEnsureVerbFormUserCards_MaterializesRuntimeClozeCards(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	userID := int64(9901)
	wordCardID := int64(10)
	formID := int64(55)
	trainingCardID := int64(77)
	verbRepo := repository.NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT DISTINCT wc\.id`).
		WithArgs(userID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word"}).AddRow(wordCardID, "hablar"))
	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma`).
		WithArgs("hablar", "es").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))
	mock.ExpectExec(`INSERT INTO word_verb_lemmas`).
		WithArgs(wordCardID, int64(99), 1.0, "auto_user_vocab").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT w\.id, w\.word, d\.id, d\.mood`).
		WithArgs("es.presente.indicativo", userID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"word_card_id", "lemma", "verb_form_dict_id", "mood", "tense", "person", "number", "surface_form"}).
			AddRow(wordCardID, "hablar", formID, "indicativo", "presente", "1", "singular", "hablo"))
	mock.ExpectQuery(`SELECT lemma, COALESCE\(metadata_json,''\) FROM verb_lemmas`).
		WithArgs("hablar").
		WillReturnRows(sqlmock.NewRows([]string{"lemma", "metadata_json"}).AddRow("hablar", `{"ru":{"gloss":"говорить"}}`))
	mock.ExpectExec(`INSERT INTO verb_training_cards`).
		WithArgs(wordCardID, formID, models.VerbCardTypeCloze, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM verb_training_cards WHERE word_card_id=\? AND verb_form_dict_id=\? AND card_type=\?`).
		WithArgs(wordCardID, formID, models.VerbCardTypeCloze).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(trainingCardID))
	mock.ExpectQuery(`SELECT DISTINCT c\.id`).
		WithArgs(models.VerbCardTypeCloze, "es.presente.indicativo", userID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(trainingCardID))
	mock.ExpectQuery(`SELECT id FROM user_verb_cards WHERE user_id=\? AND verb_training_card_id=\?`).
		WithArgs(userID, trainingCardID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(88)))

	svc := NewVerbTrainingService(verbRepo, config.LearningConfig{TargetLang: "es"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, zap.NewNop())
	if err := svc.EnsureVerbFormUserCards(userID, []string{"es.presente.indicativo"}); err != nil {
		t.Fatalf("EnsureVerbFormUserCards: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
