package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

type row struct {
	TrainingCardID  int64
	WordCardID      int64
	Lemma           string
	POS             sql.NullString
	NounGender      sql.NullString
	TranscriptionWC sql.NullString
	DefinitionRU    sql.NullString
	ExamplesJSON    sql.NullString
	VerbFormsJSON   sql.NullString
	DisplayEN       sql.NullString
	SenseIndex      int
	WordEN          string
	Transcription   sql.NullString
	WordRU          string
	MeaningEN       string
	ExampleEN       sql.NullString
	ExampleRU       sql.NullString
	DistractorsEN   sql.NullString
	DistractorsRU   sql.NullString
	Hint            sql.NullString
	DisplayWord     sql.NullString
}

type group struct {
	wordCard        models.WordCard
	resp            models.TrainingCardResponse
	trainingCardIDs []int64
}

type invalidGroup struct {
	g      group
	reason string
}

type invalidTrainingCard struct {
	TrainingCardID int64
	WordCardID     int64
	Lemma          string
	SenseIndex     int
	Reason         string
}

type definitionRegenerationResult struct {
	Lemma        string
	DefinitionRU string
}

type duplicateGroup struct {
	WordCardID int64
	Lemma      string
	DeleteIDs  []int64
	Reason     string
}

type invalidTTS struct {
	Word   string
	Reason string
}

const invalidReasonDefinitionRUNotCyrillic = "definition_ru_not_cyrillic"
const invalidReasonWordCardFieldsPrefix = "word_card_fields:"

func main() {
	os.Exit(runCLI())
}

func runCLI() int {
	commit := flag.Bool("commit", false, "apply DB changes (default: dry-run)")
	limit := flag.Int("limit", 0, "max number of word_cards to requeue (0 = no limit)")
	onlyWord := flag.String("only-word", "", "process only this lemma (optional)")
	checkTTS := flag.Bool("check-tts", true, "check and softly repair invalid TTS statuses")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		return 1
	}
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init db: %v\n", err)
		return 1
	}
	defer db.Close()

	groups, err := loadGroups(db.GetConnection(), *onlyWord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load cards: %v\n", err)
		return 1
	}

	invalidWordCards := make([]invalidGroup, 0)
	invalidTrainingCards := make([]invalidTrainingCard, 0)
	duplicates := make([]duplicateGroup, 0)
	for _, g := range groups {
		if issues := invalidWordCardFieldIssues(g, cfg.Learning); len(issues) > 0 {
			invalidWordCards = append(invalidWordCards, invalidGroup{
				g:      g,
				reason: invalidReasonWordCardFieldsPrefix + strings.Join(issues, ","),
			})
			continue
		}
		for i, sense := range g.resp.Senses {
			singleResp := models.TrainingCardResponse{
				WordEN:        g.resp.WordEN,
				WordTarget:    g.resp.WordTarget,
				Lemma:         g.resp.Lemma,
				Transcription: g.resp.Transcription,
				Senses:        []models.TrainingCardSense{sense},
			}
			if errMsg := service.ValidateTrainingCardResponse(cfg.Learning.TargetLang, &g.wordCard, &singleResp); errMsg != "" {
				if i < len(g.trainingCardIDs) {
					invalidTrainingCards = append(invalidTrainingCards, invalidTrainingCard{
						TrainingCardID: g.trainingCardIDs[i],
						WordCardID:     g.wordCard.ID,
						Lemma:          g.wordCard.Word,
						SenseIndex:     i,
						Reason:         errMsg,
					})
				}
			}
		}
		if deleteIDs, dupErr := duplicateTrainingCardIDs(g); len(deleteIDs) > 0 {
			duplicates = append(duplicates, duplicateGroup{
				WordCardID: g.wordCard.ID,
				Lemma:      g.wordCard.Word,
				DeleteIDs:  deleteIDs,
				Reason:     dupErr,
			})
		}
	}

	sort.Slice(invalidWordCards, func(i, j int) bool {
		return invalidWordCards[i].g.wordCard.ID < invalidWordCards[j].g.wordCard.ID
	})
	if *limit > 0 && len(invalidWordCards) > *limit {
		invalidWordCards = invalidWordCards[:*limit]
	}
	sort.Slice(invalidTrainingCards, func(i, j int) bool {
		if invalidTrainingCards[i].WordCardID == invalidTrainingCards[j].WordCardID {
			return invalidTrainingCards[i].TrainingCardID < invalidTrainingCards[j].TrainingCardID
		}
		return invalidTrainingCards[i].WordCardID < invalidTrainingCards[j].WordCardID
	})
	if *limit > 0 && len(invalidTrainingCards) > *limit {
		invalidTrainingCards = invalidTrainingCards[:*limit]
	}
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].WordCardID < duplicates[j].WordCardID
	})
	if *limit > 0 && len(duplicates) > *limit {
		duplicates = duplicates[:*limit]
	}

	invalidTTSList := make([]invalidTTS, 0)
	if *checkTTS {
		invalidTTSList, err = loadInvalidTTS(db.GetConnection(), cfg.TTS.AudioDir, *onlyWord)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load invalid tts statuses: %v\n", err)
			return 1
		}
		if *limit > 0 && len(invalidTTSList) > *limit {
			invalidTTSList = invalidTTSList[:*limit]
		}
	}

	mode := "DRY-RUN"
	if *commit {
		mode = "COMMIT"
	}
	fmt.Printf("[%s] total_word_cards=%d invalid_word_cards=%d invalid_training_cards=%d duplicates=%d invalid_tts=%d\n",
		mode, len(groups), len(invalidWordCards), len(invalidTrainingCards), len(duplicates), len(invalidTTSList))
	for _, g := range invalidWordCards {
		fmt.Printf("INVALID_WORD_CARD word_card_id=%d lemma=%q senses=%d reason=%q\n", g.g.wordCard.ID, g.g.wordCard.Word, len(g.g.resp.Senses), g.reason)
	}
	for _, c := range invalidTrainingCards {
		fmt.Printf("INVALID_TRAINING_CARD training_card_id=%d word_card_id=%d lemma=%q sense=%d reason=%q\n",
			c.TrainingCardID, c.WordCardID, c.Lemma, c.SenseIndex, c.Reason)
	}
	for _, d := range duplicates {
		fmt.Printf("DUPLICATE word_card_id=%d lemma=%q delete_training_cards=%v reason=%q\n", d.WordCardID, d.Lemma, d.DeleteIDs, d.Reason)
	}
	for _, t := range invalidTTSList {
		fmt.Printf("INVALID_TTS word=%q reason=%q\n", t.Word, t.Reason)
	}

	if !*commit {
		return 0
	}

	deletedDuplicates := 0
	for _, d := range duplicates {
		for _, trainingCardID := range d.DeleteIDs {
			if err := deleteTrainingCardByID(ctx, db.GetConnection(), trainingCardID); err != nil {
				fmt.Printf("ERROR delete duplicate training_card_id=%d lemma=%q: %v\n", trainingCardID, d.Lemma, err)
				continue
			}
			deletedDuplicates++
		}
	}

	wordRepo := repository.NewWordRepository(db.GetConnection(), log)
	aiService := ai.NewService(cfg.AI.URL, cfg.AI.Model, cfg.AI.APIKey, cfg.AI.Prompt, log)

	requeued := 0
	regeneratedDefinition := 0
	regeneratedDefinitionFailed := 0
	invalidWordCardSet := make(map[int64]struct{}, len(invalidWordCards))
	for _, g := range invalidWordCards {
		invalidWordCardSet[g.g.wordCard.ID] = struct{}{}
	}
	for _, g := range invalidWordCards {
		if isWordCardRegenerationReason(g.reason) {
			result, err := regenerateWordCardDefinition(ctx, db.GetConnection(), wordRepo, aiService, cfg.Learning, g.g.wordCard.ID)
			if err != nil {
				fmt.Printf("ERROR regenerate definition_ru word_card_id=%d lemma=%q: %v\n", g.g.wordCard.ID, g.g.wordCard.Word, err)
				regeneratedDefinitionFailed++
				continue
			}
			fmt.Printf("REGENERATED_DEFINITION word_card_id=%d lemma=%q definition_ru=%q\n", g.g.wordCard.ID, result.Lemma, result.DefinitionRU)
			regeneratedDefinition++
			continue
		}
		if err := requeueWordCard(ctx, db.GetConnection(), g.g.wordCard.ID, shouldNullDefinitionRU(g.g.wordCard, cfg.Learning)); err != nil {
			fmt.Printf("ERROR requeue word_card_id=%d lemma=%q: %v\n", g.g.wordCard.ID, g.g.wordCard.Word, err)
			continue
		}
		requeued++
	}
	deletedInvalidTrainingCards := 0
	for _, c := range invalidTrainingCards {
		if _, exists := invalidWordCardSet[c.WordCardID]; exists {
			continue
		}
		if err := deleteTrainingCardByID(ctx, db.GetConnection(), c.TrainingCardID); err != nil {
			fmt.Printf("ERROR delete invalid training_card_id=%d word_card_id=%d lemma=%q: %v\n", c.TrainingCardID, c.WordCardID, c.Lemma, err)
			continue
		}
		deletedInvalidTrainingCards++
	}

	resetTTS := 0
	if *checkTTS {
		ttsRepo := repository.NewTTSStatusRepository(db.GetConnection(), log, cfg.TTS.MaxRetries)
		for _, t := range invalidTTSList {
			if err := ttsRepo.ResetForForceRegenerate(t.Word); err != nil {
				fmt.Printf("ERROR reset invalid tts word=%q: %v\n", t.Word, err)
				continue
			}
			resetTTS++
		}
	}
	fmt.Printf("Done. requeued_word_cards=%d regenerated_definition=%d regen_definition_failed=%d deleted_invalid_training_cards=%d deleted_duplicate_training_cards=%d reset_tts=%d\n",
		requeued, regeneratedDefinition, regeneratedDefinitionFailed, deletedInvalidTrainingCards, deletedDuplicates, resetTTS)
	return 0
}

func loadGroups(conn *sql.DB, onlyWord string) ([]group, error) {
	query := `SELECT
		tc.id, wc.id, wc.word, wc.pos, wc.noun_gender, wc.transcription, wc.definition_ru, wc.examples_json, wc.verb_forms_json, wc.display_en,
		tc.sense_index, tc.word_en, tc.transcription, tc.word_ru, tc.meaning_en,
		tc.example_en, tc.example_ru, tc.distractors_en, tc.distractors_ru, tc.hint, tc.display_word
	FROM word_cards wc
	INNER JOIN training_cards tc ON tc.word_card_id = wc.id`
	args := []any{}
	if strings.TrimSpace(onlyWord) != "" {
		query += ` WHERE LOWER(wc.word) = LOWER(?)`
		args = append(args, strings.TrimSpace(onlyWord))
	}
	query += ` ORDER BY wc.id, tc.sense_index`

	rows, err := conn.Query(query, args...)
	if err != nil {
		if isMissingTableErr(err) {
			return []group{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	byID := map[int64]*group{}
	for rows.Next() {
		var r row
		if err := rows.Scan(
			&r.TrainingCardID, &r.WordCardID, &r.Lemma, &r.POS, &r.NounGender, &r.TranscriptionWC, &r.DefinitionRU, &r.ExamplesJSON, &r.VerbFormsJSON, &r.DisplayEN,
			&r.SenseIndex, &r.WordEN, &r.Transcription, &r.WordRU, &r.MeaningEN,
			&r.ExampleEN, &r.ExampleRU, &r.DistractorsEN, &r.DistractorsRU, &r.Hint, &r.DisplayWord,
		); err != nil {
			return nil, err
		}

		g, ok := byID[r.WordCardID]
		if !ok {
			wc := models.WordCard{
				ID:   r.WordCardID,
				Word: r.Lemma,
			}
			if r.POS.Valid && strings.TrimSpace(r.POS.String) != "" {
				pos := strings.TrimSpace(r.POS.String)
				wc.POS = &pos
			}
			if r.NounGender.Valid && strings.TrimSpace(r.NounGender.String) != "" {
				g := strings.TrimSpace(r.NounGender.String)
				wc.NounGender = &g
			}
			if r.TranscriptionWC.Valid && strings.TrimSpace(r.TranscriptionWC.String) != "" {
				tr := strings.TrimSpace(r.TranscriptionWC.String)
				wc.Transcription = &tr
			}
			if r.DefinitionRU.Valid && strings.TrimSpace(r.DefinitionRU.String) != "" {
				defRU := strings.TrimSpace(r.DefinitionRU.String)
				wc.DefinitionRU = &defRU
			}
			if r.ExamplesJSON.Valid && strings.TrimSpace(r.ExamplesJSON.String) != "" {
				ex := strings.TrimSpace(r.ExamplesJSON.String)
				wc.ExamplesJSON = &ex
			}
			if r.VerbFormsJSON.Valid && strings.TrimSpace(r.VerbFormsJSON.String) != "" {
				vf := strings.TrimSpace(r.VerbFormsJSON.String)
				wc.VerbFormsJSON = &vf
			}
			if r.DisplayEN.Valid && strings.TrimSpace(r.DisplayEN.String) != "" {
				d := strings.TrimSpace(r.DisplayEN.String)
				wc.DisplayEN = &d
			}
			gr := &group{
				wordCard: wc,
				resp: models.TrainingCardResponse{
					WordEN:        r.Lemma,
					WordTarget:    r.Lemma,
					Lemma:         r.Lemma,
					Transcription: strings.TrimSpace(r.Transcription.String),
					Senses:        []models.TrainingCardSense{},
				},
				trainingCardIDs: []int64{},
			}
			byID[r.WordCardID] = gr
			g = gr
		}

		var de, dr []string
		if strings.TrimSpace(r.DistractorsEN.String) != "" {
			_ = json.Unmarshal([]byte(r.DistractorsEN.String), &de)
		}
		if strings.TrimSpace(r.DistractorsRU.String) != "" {
			_ = json.Unmarshal([]byte(r.DistractorsRU.String), &dr)
		}
		sense := models.TrainingCardSense{
			POS:           strings.TrimSpace(r.POS.String),
			DisplayWord:   strings.TrimSpace(r.DisplayWord.String),
			WordRU:        strings.TrimSpace(r.WordRU),
			MeaningEN:     strings.TrimSpace(r.MeaningEN),
			ExampleEN:     strings.TrimSpace(r.ExampleEN.String),
			ExampleRU:     strings.TrimSpace(r.ExampleRU.String),
			DistractorsEN: de,
			DistractorsRU: dr,
			Hint:          strings.TrimSpace(r.Hint.String),
		}
		g.resp.Senses = append(g.resp.Senses, sense)
		g.trainingCardIDs = append(g.trainingCardIDs, r.TrainingCardID)
	}

	out := make([]group, 0, len(byID))
	for _, g := range byID {
		out = append(out, *g)
	}
	return out, rows.Err()
}

func isMissingTableErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") || strings.Contains(msg, "does not exist")
}

func isNativeDefinitionValid(card models.WordCard, learning config.LearningConfig) bool {
	if !strings.EqualFold(learning.NativeLang, "ru") || !strings.EqualFold(learning.TargetLang, "es") {
		return true
	}
	if card.DefinitionRU == nil {
		return false
	}
	return service.ContainsCyrillic(strings.TrimSpace(*card.DefinitionRU))
}

func shouldNullDefinitionRU(card models.WordCard, learning config.LearningConfig) bool {
	if !strings.EqualFold(learning.NativeLang, "ru") || !strings.EqualFold(learning.TargetLang, "es") {
		return false
	}
	if card.DefinitionRU == nil {
		return false
	}
	return !service.ContainsCyrillic(strings.TrimSpace(*card.DefinitionRU))
}

func requeueWordCard(ctx context.Context, conn *sql.DB, wordCardID int64, nullDefinitionRU bool) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM training_cards WHERE word_card_id = ?`, wordCardID); err != nil {
		return err
	}
	if nullDefinitionRU {
		if _, err := tx.ExecContext(ctx, `UPDATE word_cards
			SET definition_ru = NULL, processed_at = NULL, processing_error = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, wordCardID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE word_cards
			SET processed_at = NULL, processing_error = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, wordCardID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func deleteTrainingCardByID(ctx context.Context, conn *sql.DB, trainingCardID int64) error {
	_, err := conn.ExecContext(ctx, `DELETE FROM training_cards WHERE id = ?`, trainingCardID)
	return err
}

func regenerateWordCardDefinition(
	ctx context.Context,
	_ *sql.DB,
	wordRepo *repository.WordRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	wordCardID int64,
) (definitionRegenerationResult, error) {
	out := definitionRegenerationResult{}
	wordCard, err := wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		return out, fmt.Errorf("load word card: %w", err)
	}
	if wordCard == nil {
		return out, fmt.Errorf("word card not found")
	}

	response, err := aiService.GenerateResponse(ctx, wordCard.Word)
	if err != nil {
		return out, fmt.Errorf("llm generate response: %w", err)
	}

	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		return out, fmt.Errorf("parse llm json: %w", err)
	}
	models.SyncWordInfoResponseNeutralAliases(&wordInfo)
	if wordInfo.Error.IsTrue() {
		return out, fmt.Errorf("llm returned error: %s", wordInfo.Error.Message)
	}

	definitionRU := strings.TrimSpace(wordInfo.DefinitionRU)
	if !isDefinitionForLearningValid(definitionRU, learning) {
		return out, fmt.Errorf("llm returned invalid definition_native for pair %s", learning.Pair)
	}

	updated := &models.WordCard{
		ID:                 wordCard.ID,
		Word:               wordCard.Word,
		Definition:         wordCard.Definition,
		POS:                wordCard.POS,
		NounGender:         wordCard.NounGender,
		OppositeGenderWord: wordCard.OppositeGenderWord,
		Transcription:      wordCard.Transcription,
		DefinitionRU:       wordCard.DefinitionRU,
		ExamplesJSON:       wordCard.ExamplesJSON,
		VerbFormsJSON:      wordCard.VerbFormsJSON,
		DisplayEN:          wordCard.DisplayEN,
	}

	// Always refresh definition_ru in this recovery branch.
	updated.DefinitionRU = ptr(strings.TrimSpace(wordInfo.DefinitionRU))

	// Fill remaining fields only when empty/missing.
	if isNilOrBlank(updated.POS) && strings.TrimSpace(wordInfo.POS) != "" {
		canonicalPOS := models.CanonicalWordPOS(wordInfo.POS)
		if canonicalPOS != "" {
			updated.POS = ptr(canonicalPOS)
		}
	}
	if isNilOrBlank(updated.Transcription) && strings.TrimSpace(wordInfo.Transcription) != "" {
		updated.Transcription = ptr(strings.TrimSpace(wordInfo.Transcription))
	}
	if (updated.ExamplesJSON == nil || strings.TrimSpace(*updated.ExamplesJSON) == "") && len(wordInfo.Examples) > 0 {
		examplesBytes, err := json.Marshal(wordInfo.Examples)
		if err == nil {
			updated.ExamplesJSON = ptr(string(examplesBytes))
		}
	}
	if (updated.VerbFormsJSON == nil || strings.TrimSpace(*updated.VerbFormsJSON) == "") && wordInfo.VerbForms != nil {
		verbFormsBytes, err := json.Marshal(wordInfo.VerbForms)
		if err == nil {
			updated.VerbFormsJSON = ptr(string(verbFormsBytes))
		}
	}
	if isNilOrBlank(updated.NounGender) {
		if g := models.NormalizeNounGenderValue(wordInfo.NounGender); g != "" {
			updated.NounGender = ptr(g)
		} else if updated.POS != nil && models.IsNounPOS(*updated.POS) {
			if inferred := models.InferNounGenderFromPOSText(wordInfo.POS); inferred != "" {
				updated.NounGender = ptr(inferred)
			}
		}
	}
	if isNilOrBlank(updated.OppositeGenderWord) && strings.TrimSpace(wordInfo.OppositeGenderWord) != "" {
		updated.OppositeGenderWord = ptr(strings.TrimSpace(wordInfo.OppositeGenderWord))
	}
	if isNilOrBlank(updated.DisplayEN) {
		lemma := strings.ToLower(strings.TrimSpace(wordInfo.Lemma))
		if lemma == "" {
			lemma = strings.ToLower(strings.TrimSpace(wordCard.Word))
		}
		display := lemma
		if models.IsVerbPOS(wordInfo.POS) && wordInfo.VerbForms != nil && strings.TrimSpace(wordInfo.VerbForms.V1) != "" {
			if strings.EqualFold(learning.TargetLang, "en") {
				display = "to " + strings.TrimSpace(wordInfo.VerbForms.V1)
			} else {
				display = strings.TrimSpace(wordInfo.VerbForms.V1)
			}
		}
		if strings.TrimSpace(display) != "" {
			updated.DisplayEN = ptr(display)
		}
	}

	if err := wordRepo.UpdateWordCard(updated); err != nil {
		return out, fmt.Errorf("update word card: %w", err)
	}

	out.Lemma = wordCard.Word
	out.DefinitionRU = definitionRU
	return out, nil
}

func duplicateTrainingCardIDs(g group) ([]int64, string) {
	seen := make(map[string]int, len(g.resp.Senses))
	deleteIDs := make([]int64, 0)
	var firstReason string
	for i, s := range g.resp.Senses {
		key := strings.Join([]string{
			normalizeDupField(s.POS),
			normalizeDupField(s.DisplayWord),
			normalizeDupField(s.WordRU),
			normalizeDupField(s.MeaningEN),
		}, "|")
		if prevIdx, ok := seen[key]; ok {
			if i < len(g.trainingCardIDs) {
				deleteIDs = append(deleteIDs, g.trainingCardIDs[i])
			}
			if firstReason == "" {
				firstReason = fmt.Sprintf("duplicate_training_sense sense=%d duplicates sense=%d", i, prevIdx)
			}
			continue
		}
		seen[key] = i
	}
	return deleteIDs, firstReason
}

func normalizeDupField(v string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
}

func isDefinitionForLearningValid(definition string, learning config.LearningConfig) bool {
	definition = strings.TrimSpace(definition)
	if definition == "" {
		return false
	}
	if strings.EqualFold(learning.NativeLang, "ru") && strings.EqualFold(learning.TargetLang, "es") {
		return service.ContainsCyrillic(definition)
	}
	return true
}

func ptr(v string) *string { return &v }

func isNilOrBlank(v *string) bool {
	return v == nil || strings.TrimSpace(*v) == ""
}

func invalidWordCardFieldIssues(g group, learning config.LearningConfig) []string {
	card := g.wordCard
	issues := make([]string, 0, 8)
	if !isNativeDefinitionValid(card, learning) {
		issues = append(issues, invalidReasonDefinitionRUNotCyrillic)
	}
	posMissing := isNilOrBlank(card.POS)
	respPOS := sensesPOS(g.resp)
	if posMissing && strings.TrimSpace(respPOS) == "" {
		issues = append(issues, "missing_pos")
	}
	if isNilOrBlank(card.Transcription) && strings.TrimSpace(g.resp.Transcription) == "" {
		issues = append(issues, "missing_transcription")
	}
	if isNilOrBlank(card.DisplayEN) && !hasAnyDisplayWordInSenses(g.resp.Senses) {
		issues = append(issues, "missing_display_en")
	}
	if isNilOrBlank(card.ExamplesJSON) && !hasAnyExamplesInSenses(g.resp.Senses) {
		issues = append(issues, "missing_examples_json")
	}
	posValue := ""
	if card.POS != nil {
		posValue = *card.POS
	} else {
		posValue = respPOS
	}
	if models.IsVerbPOS(posValue) && isNilOrBlank(card.VerbFormsJSON) {
		issues = append(issues, "missing_verb_forms_json")
	}
	if models.IsNounPOS(posValue) && isNilOrBlank(card.NounGender) {
		issues = append(issues, "missing_noun_gender")
	}
	return issues
}

func isWordCardRegenerationReason(reason string) bool {
	return strings.HasPrefix(reason, invalidReasonWordCardFieldsPrefix)
}

func sensesPOS(r models.TrainingCardResponse) string {
	for _, s := range r.Senses {
		if strings.TrimSpace(s.POS) != "" {
			return strings.TrimSpace(s.POS)
		}
	}
	return ""
}

func hasAnyDisplayWordInSenses(senses []models.TrainingCardSense) bool {
	for _, s := range senses {
		if strings.TrimSpace(s.DisplayWord) != "" {
			return true
		}
	}
	return false
}

func hasAnyExamplesInSenses(senses []models.TrainingCardSense) bool {
	for _, s := range senses {
		if strings.TrimSpace(s.ExampleEN) != "" || strings.TrimSpace(s.ExampleRU) != "" {
			return true
		}
	}
	return false
}

func loadInvalidTTS(conn *sql.DB, audioDir, onlyWord string) ([]invalidTTS, error) {
	query := `SELECT word, state, COALESCE(audio_rel_path, '')
		FROM tts_generation_status`
	args := []any{}
	if strings.TrimSpace(onlyWord) != "" {
		query += ` WHERE LOWER(word) = LOWER(?)`
		args = append(args, strings.TrimSpace(onlyWord))
	}
	query += ` ORDER BY updated_at DESC, word`

	rows, err := conn.Query(query, args...)
	if err != nil {
		if isMissingTableErr(err) {
			return []invalidTTS{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]invalidTTS, 0)
	seen := make(map[string]struct{})
	for rows.Next() {
		var word, state, relPath string
		if err := rows.Scan(&word, &state, &relPath); err != nil {
			return nil, err
		}
		word = strings.TrimSpace(strings.ToLower(word))
		if word == "" {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}

		reason := ""
		switch state {
		case models.TTSStateFailedRetryable, models.TTSStateFailedTerminal:
			reason = "failed_state"
		case models.TTSStateReady:
			if !hasAnyAudioFile(audioDir, word, relPath) {
				reason = "ready_but_audio_missing"
			}
		}
		if reason == "" {
			continue
		}
		seen[word] = struct{}{}
		out = append(out, invalidTTS{
			Word:   word,
			Reason: reason,
		})
	}
	return out, rows.Err()
}

func hasAnyAudioFile(audioDir, word, relPath string) bool {
	if strings.TrimSpace(audioDir) == "" {
		return false
	}
	if audioRelPathExists(audioDir, relPath) {
		return true
	}
	legacyBase := legacyAudioRelBase(word)
	if legacyBase == "" {
		return false
	}
	return audioRelPathExists(audioDir, legacyBase+".mp3") || audioRelPathExists(audioDir, legacyBase+".wav")
}

func audioRelPathExists(audioDir, relPath string) bool {
	clean := filepath.Clean(strings.TrimSpace(relPath))
	if clean == "." || clean == "" || strings.HasPrefix(clean, "..") {
		return false
	}
	_, err := os.Stat(filepath.Join(audioDir, clean))
	return err == nil
}

func legacyAudioRelBase(word string) string {
	key := pronunciationWordKey(word)
	if len(key) < 4 {
		return ""
	}
	return filepath.Join(key[:2], key[2:4], pronunciationWordFileBase(word))
}

func pronunciationWordKey(word string) string {
	sum := sha256.Sum256([]byte(word))
	return hex.EncodeToString(sum[:])
}

func pronunciationWordFileBase(word string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(word)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		if r == '-' || r == '\'' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsSpace(r) {
			b.WriteByte('_')
		}
	}
	name := strings.Trim(b.String(), "_")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	if name == "" {
		return "word"
	}
	return name
}
