package service

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"

	"go.uber.org/zap"
)

type VerbTrainingSession struct {
	SessionID int64
	Queue     []repository.VerbQueueCard
	Index     int
}

type VerbTrainingService struct {
	repo     *repository.VerbFormsRepository
	learning config.LearningConfig
	cfg      config.TrainingConfig
	logger   *zap.Logger
}

func NewVerbTrainingService(repo *repository.VerbFormsRepository, learning config.LearningConfig, cfg config.TrainingConfig, logger *zap.Logger) *VerbTrainingService {
	return &VerbTrainingService{repo: repo, learning: learning, cfg: cfg, logger: logger}
}

func (s *VerbTrainingService) Enabled() bool {
	return s.cfg.SpanishVerbFormsEnabled && strings.EqualFold(s.learning.TargetLang, "es")
}

// isSpanishVerbTrainingScope is true when s matches keys built as 'es.'+tense+'.'+mood in verb_forms_dict.
// grammar_stage often stores a grammar-bundle chapter id (es.grammar.*); those must not be used as verb scopes
// or GetLinkedVerbFormsForUser returns no rows despite words being in the vocabulary.
func isSpanishVerbTrainingScope(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if !strings.HasPrefix(s, "es.") {
		return false
	}
	rest := strings.TrimPrefix(s, "es.")
	if rest == "" || !strings.Contains(rest, ".") {
		return false
	}
	// Spanish grammar course chapter slugs in user settings
	if strings.HasPrefix(rest, "grammar.") {
		return false
	}
	return true
}

func ResolveVerbScopes(settings *models.UserSettings, learning config.LearningConfig) []string {
	if !strings.EqualFold(learning.TargetLang, "es") {
		return nil
	}
	if settings != nil {
		if len(settings.EnabledVerbScopes) > 0 {
			out := make([]string, 0, len(settings.EnabledVerbScopes))
			for _, scope := range settings.EnabledVerbScopes {
				scope = strings.ToLower(strings.TrimSpace(scope))
				if scope != "" && isSpanishVerbTrainingScope(scope) {
					out = append(out, scope)
				}
			}
			// Ignore garbage (e.g. es.grammar.* chapter ids saved by mistake): must match
			// ('es.'||tense||'.'||mood) in verb_forms_dict or training stays empty while vocab UI still shows forms.
			if len(out) > 0 {
				return out
			}
		}
		if settings.GrammarStage != nil {
			g := strings.ToLower(strings.TrimSpace(*settings.GrammarStage))
			if g != "" && isSpanishVerbTrainingScope(g) {
				return []string{g}
			}
		}
	}
	return models.DefaultSpanishVerbScopes()
}

// EnsureVerbFormUserCards links Spanish lemmas, upserts verb_training_cards, and creates user_verb_cards
// for the user's vocabulary. Call after user_cards are created (same moment as word training cards).
func (s *VerbTrainingService) EnsureVerbFormUserCards(userID int64, scopes []string) error {
	if !s.Enabled() {
		return nil
	}
	if err := s.repo.LinkMissingSpanishVerbLemmasForUser(userID); err != nil {
		return err
	}
	if err := s.ensureTrainingCardsForUser(userID, scopes); err != nil {
		return err
	}
	return s.repo.EnsureUserCardsForUserWords(userID, scopes)
}

func (s *VerbTrainingService) StartSession(userID int64, scopes []string) (*VerbTrainingSession, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("verb forms training is disabled")
	}
	if err := s.EnsureVerbFormUserCards(userID, scopes); err != nil {
		return nil, err
	}
	queue, err := s.repo.GetVerbQueue(userID, time.Now(), s.cfg.VerbFormsMaxCards, s.cfg.VerbFormsMaxNew)
	if err != nil {
		return nil, err
	}
	shuffleSeed := time.Now().UnixNano() ^ userID ^ int64(len(queue)<<20)
	ShuffleVerbQueue(queue, shuffleSeed)
	queue = SpreadAdjacentDuplicateVerbPromptKeys(queue)
	if len(queue) == 0 {
		return nil, fmt.Errorf("no cards available for training")
	}
	sessionJSON, _ := json.Marshal(map[string]interface{}{
		"scopes": scopes,
		"max":    s.cfg.VerbFormsMaxCards,
	})
	sessionID, err := s.repo.StartVerbSession(userID, len(queue), string(sessionJSON))
	if err != nil {
		return nil, err
	}
	return &VerbTrainingSession{SessionID: sessionID, Queue: queue, Index: 0}, nil
}

func (s *VerbTrainingService) ensureTrainingCardsForUser(userID int64, scopes []string) error {
	rows, err := s.repo.GetLinkedVerbFormsForUser(userID, scopes)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	lemmas := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		lemma := strings.ToLower(strings.TrimSpace(row.Lemma))
		if lemma == "" {
			continue
		}
		if _, ok := seen[lemma]; ok {
			continue
		}
		seen[lemma] = struct{}{}
		lemmas = append(lemmas, lemma)
	}
	metadata, err := s.repo.GetVerbLemmaMetadataJSONBatch(lemmas)
	if err != nil {
		return err
	}
	for _, row := range rows {
		card, err := buildRuntimeVerbTrainingCard(row, metadata[strings.ToLower(strings.TrimSpace(row.Lemma))])
		if err != nil {
			return err
		}
		if _, err := s.repo.UpsertVerbTrainingCard(card); err != nil {
			return err
		}
	}
	return nil
}

func stableVerbTrainingSeed(row repository.LinkedVerbFormRow) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(row.Lemma))))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(row.Mood))))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(row.Tense))))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(row.Person))))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(row.Number))))
	return int64(h.Sum64() ^ uint64(row.VerbFormDictID))
}

func buildRuntimeVerbTrainingCard(row repository.LinkedVerbFormRow, metadataJSON string) (*models.VerbTrainingCard, error) {
	lemma := strings.ToLower(strings.TrimSpace(row.Lemma))
	mood := strings.ToLower(strings.TrimSpace(row.Mood))
	tense := strings.ToLower(strings.TrimSpace(row.Tense))
	person := strings.ToLower(strings.TrimSpace(row.Person))
	number := strings.ToLower(strings.TrimSpace(row.Number))
	surface := strings.ToLower(strings.TrimSpace(row.SurfaceForm))
	if lemma == "" || mood == "" || tense == "" || person == "" || number == "" || surface == "" {
		return nil, fmt.Errorf("incomplete linked verb form row: %+v", row)
	}

	ruGloss := spanishverbs.RuGlossFromLemmaMetadataJSON(metadataJSON)
	if ruGloss == "" {
		ruGloss = spanishverbs.DefaultRuGloss(lemma)
	}
	verbClass := spanishverbs.VerbClassFromLemmaMetadataJSON(metadataJSON)
	allowed := spanishverbs.AllowedTemplateIDsFromLemmaMetadataJSON(metadataJSON)
	seed := stableVerbTrainingSeed(row)
	question, translation := spanishverbs.GenerateVerbExamplePair(seed, lemma, mood, tense, person, number, surface, ruGloss, verbClass, allowed)
	if strings.TrimSpace(question) == "" {
		question = spanishverbs.BuildVerbTrainingClozeQuestion(person, number, lemma, mood, tense)
	}
	if strings.TrimSpace(translation) == "" {
		translation = spanishverbs.PlainRussianVerbTrainingHintLine(lemma, person, number, ruGloss, mood, tense)
	}
	if strings.TrimSpace(translation) == "" {
		translation = "форма глагола"
	}

	prompt := map[string]interface{}{
		"type":                models.VerbCardTypeCloze,
		"example_mode":        "runtime_generated",
		"example_source":      "verb_forms_dict_runtime",
		"lemma":               lemma,
		"mood":                mood,
		"tense":               tense,
		"person":              person,
		"number":              number,
		"expected_form":       surface,
		"question":            question,
		"example_translation": translation,
	}
	if ruGloss != "" {
		prompt["ru_gloss"] = ruGloss
	}
	if verbClass != "" {
		prompt["verb_class"] = verbClass
	}
	if len(allowed) > 0 {
		prompt["allowed_template_ids"] = allowed
	}
	promptJSON, err := json.Marshal(prompt)
	if err != nil {
		return nil, err
	}
	answerJSON, err := json.Marshal(map[string]string{"surface_form": surface})
	if err != nil {
		return nil, err
	}
	optionsJSON, err := json.Marshal(BuildVerbFormMultipleChoiceOptions(surface, lemma, seed))
	if err != nil {
		return nil, err
	}
	return &models.VerbTrainingCard{
		WordCardID:      row.WordCardID,
		VerbFormDictID:  row.VerbFormDictID,
		CardType:        models.VerbCardTypeCloze,
		PromptJSON:      string(promptJSON),
		AnswerJSON:      string(answerJSON),
		DistractorsJSON: string(optionsJSON),
	}, nil
}

func (s *VerbTrainingService) Grade(userID, sessionID, userVerbCardID int64, isCorrect bool) error {
	if !s.Enabled() {
		return fmt.Errorf("verb forms training is disabled")
	}
	card, err := s.repo.GetVerbUserCardSRS(userVerbCardID)
	if err != nil {
		return err
	}
	if card == nil {
		return fmt.Errorf("user verb card not found")
	}
	quality := 2
	if isCorrect {
		quality = 5
	}
	if quality >= 4 {
		card.Reps++
		switch card.Reps {
		case 1:
			card.IntervalDays = 1
		case 2:
			card.IntervalDays = 3
		default:
			card.IntervalDays = int(math.Round(float64(card.IntervalDays) * card.EF))
			if card.IntervalDays < 1 {
				card.IntervalDays = 1
			}
		}
		card.EF = card.EF + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
		if card.EF < 1.3 {
			card.EF = 1.3
		}
		card.State = "review"
	} else {
		card.Reps = 0
		card.IntervalDays = 0
		card.LearningStep = 0
		card.LapseCount++
		card.State = "learning"
		if card.EF > 2.5 {
			card.EF = 2.5
		}
	}
	nextDueAt := time.Now().Add(time.Duration(card.IntervalDays) * 24 * time.Hour)
	if card.State == "learning" {
		nextDueAt = time.Now().Add(10 * time.Minute)
	}
	if err := s.repo.UpdateVerbUserCardSRS(card, nextDueAt, quality); err != nil {
		return err
	}
	return s.repo.CreateVerbReviewEvent(sessionID, userID, userVerbCardID, isCorrect, quality)
}
