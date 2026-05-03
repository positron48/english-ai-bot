package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"

	"go.uber.org/zap"
)

type VerbFormsRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewVerbFormsRepository(db *sql.DB, logger *zap.Logger) *VerbFormsRepository {
	return &VerbFormsRepository{db: db, logger: logger}
}

// verbTrainingEligibleByWordCardSQL keeps only verb_training rows whose word_card is linked to a Spanish
// verb_lemmas row via word_verb_lemmas. POS-only heuristics miss mis-tagged nouns ("palabra") that still
// have stray verb_training_cards from old syncs.
func verbTrainingEligibleByWordCardSQL(verbTrainingTableAlias string) string {
	return `
AND EXISTS (
  SELECT 1 FROM word_verb_lemmas wvl
  INNER JOIN verb_lemmas vl ON vl.id = wvl.verb_lemma_id AND vl.language = 'es'
  WHERE wvl.word_card_id = ` + verbTrainingTableAlias + `.word_card_id
)`
}

// verbTrainingPromptHasExampleTranslationSQL keeps only cards with a non-empty example_translation in prompt_json
// (skips legacy/partial rows before full lemma pack import).
func verbTrainingPromptHasExampleTranslationSQL(vtcAlias string) string {
	return ` AND NULLIF(TRIM(COALESCE(` + vtcAlias + `.prompt_json::jsonb->>'example_translation','')), '') IS NOT NULL`
}

func (r *VerbFormsRepository) UpsertVerbLemma(lemma, language, source, sourceVersion, checksum, metadataJSON string) (int64, error) {
	q := `INSERT INTO verb_lemmas (lemma, language, source, source_version, checksum, metadata_json, updated_at)
	      VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	      ON CONFLICT(lemma, language) DO UPDATE SET
	        source=excluded.source,
	        source_version=excluded.source_version,
	        checksum=excluded.checksum,
	        metadata_json=CASE
	          WHEN TRIM(COALESCE(excluded.metadata_json, '')) IN ('', '{}') THEN COALESCE(verb_lemmas.metadata_json, '{}')
	          ELSE excluded.metadata_json
	        END,
	        updated_at=CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, strings.ToLower(strings.TrimSpace(lemma)), strings.ToLower(strings.TrimSpace(language)), source, sourceVersion, checksum, metadataJSON); err != nil {
		return 0, fmt.Errorf("upsert verb lemma: %w", err)
	}
	var id int64
	if err := r.db.QueryRow(`SELECT id FROM verb_lemmas WHERE lemma = ? AND language = ?`, strings.ToLower(strings.TrimSpace(lemma)), strings.ToLower(strings.TrimSpace(language))).Scan(&id); err != nil {
		return 0, fmt.Errorf("select verb lemma id: %w", err)
	}
	return id, nil
}

func irregularInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *VerbFormsRepository) UpsertVerbForm(form *models.VerbFormDict) (int64, error) {
	q := `INSERT INTO verb_forms_dict (verb_lemma_id, mood, tense, person, number, surface_form, is_irregular, tags_json, updated_at)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	      ON CONFLICT(verb_lemma_id, mood, tense, person, number) DO UPDATE SET
	        surface_form=excluded.surface_form,
	        is_irregular=excluded.is_irregular,
	        tags_json=excluded.tags_json,
	        updated_at=CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, form.VerbLemmaID, strings.ToLower(form.Mood), strings.ToLower(form.Tense), strings.ToLower(form.Person), strings.ToLower(form.Number), strings.ToLower(form.SurfaceForm), irregularInt(form.IsIrregular), form.TagsJSON); err != nil {
		return 0, fmt.Errorf("upsert verb form: %w", err)
	}
	var id int64
	if err := r.db.QueryRow(`SELECT id FROM verb_forms_dict WHERE verb_lemma_id=? AND mood=? AND tense=? AND person=? AND number=?`,
		form.VerbLemmaID, strings.ToLower(form.Mood), strings.ToLower(form.Tense), strings.ToLower(form.Person), strings.ToLower(form.Number)).Scan(&id); err != nil {
		return 0, fmt.Errorf("select verb form id: %w", err)
	}
	return id, nil
}

func (r *VerbFormsRepository) LinkWordCardToLemma(wordCardID, verbLemmaID int64, confidence float64, source string) error {
	q := `INSERT INTO word_verb_lemmas (word_card_id, verb_lemma_id, confidence, source, updated_at)
	      VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	      ON CONFLICT(word_card_id) DO UPDATE SET
	        verb_lemma_id=excluded.verb_lemma_id,
	        confidence=excluded.confidence,
	        source=excluded.source,
	        updated_at=CURRENT_TIMESTAMP`
	_, err := r.db.Exec(q, wordCardID, verbLemmaID, confidence, source)
	if err != nil {
		return fmt.Errorf("link word card to lemma: %w", err)
	}
	return nil
}

func (r *VerbFormsRepository) LinkWordCardByLemma(wordCardID int64, lemma, language, source string) (bool, error) {
	var verbLemmaID int64
	err := r.db.QueryRow(`SELECT id FROM verb_lemmas WHERE lemma = ? AND language = ?`, strings.ToLower(strings.TrimSpace(lemma)), strings.ToLower(strings.TrimSpace(language))).Scan(&verbLemmaID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lookup verb lemma: %w", err)
	}
	if err := r.LinkWordCardToLemma(wordCardID, verbLemmaID, 1.0, source); err != nil {
		return false, err
	}
	return true, nil
}

// LinkMissingSpanishVerbLemmasForUser links each vocabulary word_card to verb_lemmas when
// word_cards.word matches a Spanish dictionary lemma. UI can show verb_forms_json from the LLM
// while training uses word_verb_lemmas + verb_forms_dict; this reconciles missing links after
// the conjugation dictionary is populated or if the first link attempt happened too early.
func (r *VerbFormsRepository) LinkMissingSpanishVerbLemmasForUser(userID int64) error {
	q := `SELECT DISTINCT wc.id, LOWER(TRIM(wc.word))
	      FROM word_cards wc
	      WHERE wc.id IN (
	        SELECT tc.word_card_id FROM user_cards uc
	          INNER JOIN training_cards tc ON tc.id = uc.training_card_id
	          WHERE uc.user_id = ?
	        UNION
	        SELECT uwk.word_card_id FROM user_word_knowledge uwk
	          WHERE uwk.user_id = ? AND uwk.status = 'known'
	      )`
	rows, err := r.db.Query(q, userID, userID)
	if err != nil {
		return fmt.Errorf("list vocab word cards for verb lemma link: %w", err)
	}
	defer rows.Close()
	type wcRow struct {
		id   int64
		word string
	}
	var list []wcRow
	for rows.Next() {
		var row wcRow
		if err := rows.Scan(&row.id, &row.word); err != nil {
			return fmt.Errorf("scan vocab word card: %w", err)
		}
		list = append(list, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	const lang = "es"
	for _, row := range list {
		if row.word == "" {
			continue
		}
		if _, err := r.LinkWordCardByLemma(row.id, row.word, lang, "auto_user_vocab"); err != nil {
			return fmt.Errorf("link verb lemma word_card_id=%d: %w", row.id, err)
		}
	}
	return nil
}

func (r *VerbFormsRepository) UpsertVerbFormExample(example *models.VerbFormExample) (int64, error) {
	q := `INSERT INTO verb_form_examples (verb_form_dict_id, example_target, gloss_native, source, quality_score, updated_at)
	      VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	      ON CONFLICT(verb_form_dict_id, example_target) DO UPDATE SET
	        gloss_native=excluded.gloss_native,
	        source=excluded.source,
	        quality_score=excluded.quality_score,
	        updated_at=CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, example.VerbFormDictID, example.ExampleTarget, example.GlossNative, example.Source, example.QualityScore); err != nil {
		return 0, fmt.Errorf("upsert verb form example: %w", err)
	}
	var id int64
	if err := r.db.QueryRow(`SELECT id FROM verb_form_examples WHERE verb_form_dict_id=? AND example_target=?`, example.VerbFormDictID, example.ExampleTarget).Scan(&id); err != nil {
		return 0, fmt.Errorf("select verb form example id: %w", err)
	}
	return id, nil
}

type VerbFormViewRow struct {
	WordCardID  int64  `json:"word_card_id"`
	Lemma       string `json:"lemma"`
	Mood        string `json:"mood"`
	Tense       string `json:"tense"`
	Person      string `json:"person"`
	Number      string `json:"number"`
	SurfaceForm string `json:"surface_form"`
	IsIrregular bool   `json:"is_irregular"`
}

type PendingVerbLemmaRow struct {
	WordCardID int64  `json:"word_card_id"`
	Lemma      string `json:"lemma"`
}

type LinkedVerbFormRow struct {
	WordCardID     int64
	Lemma          string
	VerbFormDictID int64
	Mood           string
	Tense          string
	Person         string
	Number         string
	SurfaceForm    string
}

// VerbFormPreviewRow is one finite/imperative surface from verb_forms_dict (tooling / preview).
type VerbFormPreviewRow struct {
	VerbFormDictID int64
	Mood           string
	Tense          string
	Person         string
	Number         string
	SurfaceForm    string
}

// ListVerbFormsForLemma returns all paradigm rows for a lemma (language default "es").
func (r *VerbFormsRepository) ListVerbFormsForLemma(lemma, language string) ([]VerbFormPreviewRow, error) {
	lemma = strings.ToLower(strings.TrimSpace(lemma))
	if lemma == "" {
		return nil, fmt.Errorf("empty lemma")
	}
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "" {
		lang = "es"
	}
	q := `SELECT d.id, d.mood, d.tense, d.person, d.number, d.surface_form
	      FROM verb_forms_dict d
	      JOIN verb_lemmas l ON l.id = d.verb_lemma_id
	      WHERE l.lemma = ? AND l.language = ?`
	rows, err := r.db.Query(q, lemma, lang)
	if err != nil {
		return nil, fmt.Errorf("list verb forms for lemma: %w", err)
	}
	defer rows.Close()
	out := make([]VerbFormPreviewRow, 0, 80)
	for rows.Next() {
		var row VerbFormPreviewRow
		if err := rows.Scan(&row.VerbFormDictID, &row.Mood, &row.Tense, &row.Person, &row.Number, &row.SurfaceForm); err != nil {
			return nil, fmt.Errorf("scan verb form preview: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	SortVerbFormPreviewRowsSpanish(out)
	return out, nil
}

// ListVerbFormViewRowsForLemma returns all stored paradigm forms for a Spanish lemma (dictionary only, no user filter).
func (r *VerbFormsRepository) ListVerbFormViewRowsForLemma(lemma, language string) ([]VerbFormViewRow, error) {
	lemma = strings.ToLower(strings.TrimSpace(lemma))
	if lemma == "" {
		return nil, fmt.Errorf("empty lemma")
	}
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "" {
		lang = "es"
	}
	rows, err := r.db.Query(`SELECT 0, l.lemma, d.mood, d.tense, d.person, d.number, d.surface_form, d.is_irregular
		FROM verb_lemmas l
		JOIN verb_forms_dict d ON d.verb_lemma_id = l.id
		WHERE l.lemma = ? AND l.language = ?`, lemma, lang)
	if err != nil {
		return nil, fmt.Errorf("list verb form view rows for lemma: %w", err)
	}
	defer rows.Close()
	out := make([]VerbFormViewRow, 0, 80)
	for rows.Next() {
		var row VerbFormViewRow
		var ir int64
		if err := rows.Scan(&row.WordCardID, &row.Lemma, &row.Mood, &row.Tense, &row.Person, &row.Number, &row.SurfaceForm, &ir); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		row.IsIrregular = ir != 0
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	SortVerbFormViewRowsSpanish(out)
	return out, nil
}

func (r *VerbFormsRepository) GetUserVerbForms(userID, wordCardID int64) ([]VerbFormViewRow, error) {
	q := `SELECT w.id, w.word, d.mood, d.tense, d.person, d.number, d.surface_form, d.is_irregular
	      FROM word_cards w
	      JOIN word_verb_lemmas l ON l.word_card_id = w.id
	      JOIN verb_forms_dict d ON d.verb_lemma_id = l.verb_lemma_id
	      WHERE w.id = ?
	        AND EXISTS (
	          SELECT 1
	          FROM (
	            SELECT tc.word_card_id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id = ?
	            UNION
	            SELECT uwk.word_card_id FROM user_word_knowledge uwk WHERE uwk.user_id = ? AND uwk.status='known'
	          ) x
	          WHERE x.word_card_id = w.id
	        )
	      ORDER BY d.mood, d.tense, d.person, d.number`
	rows, err := r.db.Query(q, wordCardID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("get user verb forms: %w", err)
	}
	defer rows.Close()
	out := make([]VerbFormViewRow, 0)
	for rows.Next() {
		var row VerbFormViewRow
		var ir int64
		if err := rows.Scan(&row.WordCardID, &row.Lemma, &row.Mood, &row.Tense, &row.Person, &row.Number, &row.SurfaceForm, &ir); err != nil {
			return nil, fmt.Errorf("scan user verb forms: %w", err)
		}
		row.IsIrregular = ir != 0
		out = append(out, row)
	}
	SortVerbFormViewRowsSpanish(out)
	return out, nil
}

func (r *VerbFormsRepository) GetLinkedVerbFormsForUser(userID int64, scopes []string) ([]LinkedVerbFormRow, error) {
	if len(scopes) == 0 {
		return nil, nil
	}
	scopeList := strings.Repeat("?,", len(scopes))
	scopeList = strings.TrimSuffix(scopeList, ",")
	args := make([]interface{}, 0, len(scopes)+2)
	for _, s := range scopes {
		args = append(args, strings.ToLower(strings.TrimSpace(s)))
	}
	args = append(args, userID, userID)
	q := `SELECT w.id, w.word, d.id, d.mood, d.tense, d.person, d.number, d.surface_form
	      FROM word_cards w
	      JOIN word_verb_lemmas l ON l.word_card_id = w.id
	      JOIN verb_forms_dict d ON d.verb_lemma_id = l.verb_lemma_id
	      WHERE ('es.' || d.tense || '.' || d.mood) IN (` + scopeList + `)
	        AND w.id IN (
	          SELECT tc.word_card_id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=?
	          UNION
	          SELECT word_card_id FROM user_word_knowledge WHERE user_id=? AND status='known'
	        )
	      ORDER BY w.id, d.mood, d.tense, d.person, d.number`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("get linked verb forms for user: %w", err)
	}
	defer rows.Close()
	out := make([]LinkedVerbFormRow, 0)
	for rows.Next() {
		var row LinkedVerbFormRow
		if err := rows.Scan(&row.WordCardID, &row.Lemma, &row.VerbFormDictID, &row.Mood, &row.Tense, &row.Person, &row.Number, &row.SurfaceForm); err != nil {
			return nil, fmt.Errorf("scan linked verb form: %w", err)
		}
		out = append(out, row)
	}
	return out, nil
}

func (r *VerbFormsRepository) GetVerbFormExamples(verbFormDictID int64, limit int) ([]models.VerbFormExample, error) {
	rows, err := r.db.Query(`SELECT id, verb_form_dict_id, example_target, COALESCE(gloss_native,''), COALESCE(source,''), quality_score
		FROM verb_form_examples WHERE verb_form_dict_id=? ORDER BY quality_score DESC, id ASC LIMIT ?`, verbFormDictID, limit)
	if err != nil {
		return nil, fmt.Errorf("get verb form examples: %w", err)
	}
	defer rows.Close()
	out := make([]models.VerbFormExample, 0)
	for rows.Next() {
		var ex models.VerbFormExample
		if err := rows.Scan(&ex.ID, &ex.VerbFormDictID, &ex.ExampleTarget, &ex.GlossNative, &ex.Source, &ex.QualityScore); err != nil {
			return nil, fmt.Errorf("scan verb form example: %w", err)
		}
		out = append(out, ex)
	}
	return out, nil
}

type VerbQueueCard struct {
	UserVerbCardID  int64
	WordCardID      int64
	CardType        string
	PromptJSON      string
	AnswerJSON      string
	DistractorsJSON string
}

func (r *VerbFormsRepository) GetOrCreateUserVerbCard(userID, verbTrainingCardID int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(`SELECT id FROM user_verb_cards WHERE user_id=? AND verb_training_card_id=?`, userID, verbTrainingCardID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup user verb card: %w", err)
	}
	q := `INSERT INTO user_verb_cards (
		user_id, verb_training_card_id, state, ef, reps, interval_days, learning_step, lapse_count, next_due_at, created_at, updated_at
	) VALUES (?, ?, 'new', 2.5, 0, 0, 0, 0, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	newID, err := database.InsertAndReturnID(r.db, q, userID, verbTrainingCardID)
	if err != nil {
		return 0, fmt.Errorf("create user verb card: %w", err)
	}
	return newID, nil
}

func (r *VerbFormsRepository) UpsertVerbTrainingCard(card *models.VerbTrainingCard) (int64, error) {
	q := `INSERT INTO verb_training_cards (
		word_card_id, verb_form_dict_id, card_type, prompt_json, answer_json, distractors_json, example_id, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(word_card_id, verb_form_dict_id, card_type) DO UPDATE SET
		prompt_json=excluded.prompt_json,
		answer_json=excluded.answer_json,
		distractors_json=excluded.distractors_json,
		example_id=excluded.example_id,
		updated_at=CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, card.WordCardID, card.VerbFormDictID, card.CardType, card.PromptJSON, card.AnswerJSON, card.DistractorsJSON, card.ExampleID); err != nil {
		return 0, fmt.Errorf("upsert verb training card: %w", err)
	}
	var id int64
	if err := r.db.QueryRow(`SELECT id FROM verb_training_cards WHERE word_card_id=? AND verb_form_dict_id=? AND card_type=?`,
		card.WordCardID, card.VerbFormDictID, card.CardType).Scan(&id); err != nil {
		return 0, fmt.Errorf("select verb training card id: %w", err)
	}
	return id, nil
}

func (r *VerbFormsRepository) EnsureUserCardsForUserWords(userID int64, scopes []string) error {
	if len(scopes) == 0 {
		return nil
	}
	scopeList := strings.Repeat("?,", len(scopes))
	scopeList = strings.TrimSuffix(scopeList, ",")
	// Placeholders order: card_type, then each scope for IN (...), then user_id twice in subquery.
	args := make([]interface{}, 0, len(scopes)+3)
	args = append(args, models.VerbCardTypeCloze)
	for _, s := range scopes {
		args = append(args, strings.ToLower(strings.TrimSpace(s)))
	}
	args = append(args, userID, userID)
	q := `SELECT DISTINCT c.id
	      FROM verb_training_cards c
	      JOIN verb_forms_dict d ON d.id = c.verb_form_dict_id
	      WHERE c.card_type = ?
	        AND (('es.' || d.tense || '.' || d.mood) IN (` + scopeList + `))
	        AND c.word_card_id IN (
	          SELECT tc.word_card_id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=?
	          UNION
	          SELECT word_card_id FROM user_word_knowledge WHERE user_id=? AND status='known'
	        )` + verbTrainingEligibleByWordCardSQL("c") + verbTrainingPromptHasExampleTranslationSQL("c")
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return fmt.Errorf("load candidate verb training cards: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cardID int64
		if err := rows.Scan(&cardID); err != nil {
			return err
		}
		if _, err := r.GetOrCreateUserVerbCard(userID, cardID); err != nil {
			return err
		}
	}
	return nil
}

// verbNewQueueRow is a new (state=new) verb card plus word_card_id for fair mixing across lemmas.
type verbNewQueueRow struct {
	card       VerbQueueCard
	wordCardID int64
}

// roundRobinVerbNewCards picks up to maxPick cards rotating by word_card_id so one verb
// cannot occupy the entire "new" budget when many lemmas have fresh cards (common right after import).
func roundRobinVerbNewCards(pooled []verbNewQueueRow, maxPick int) []VerbQueueCard {
	if maxPick <= 0 || len(pooled) == 0 {
		return nil
	}
	byWord := make(map[int64][]VerbQueueCard)
	order := make([]int64, 0)
	for _, row := range pooled {
		if _, ok := byWord[row.wordCardID]; !ok {
			order = append(order, row.wordCardID)
		}
		byWord[row.wordCardID] = append(byWord[row.wordCardID], row.card)
	}
	out := make([]VerbQueueCard, 0, maxPick)
	for len(out) < maxPick {
		progress := false
		for _, wid := range order {
			q := byWord[wid]
			if len(q) == 0 {
				continue
			}
			out = append(out, q[0])
			byWord[wid] = q[1:]
			progress = true
			if len(out) >= maxPick {
				return out
			}
		}
		if !progress {
			break
		}
	}
	return out
}

func (r *VerbFormsRepository) GetVerbQueue(userID int64, now time.Time, maxCards, maxNew int) ([]VerbQueueCard, error) {
	if maxCards < 0 {
		maxCards = 0
	}
	if maxNew < 0 {
		maxNew = 0
	}
	dueLimit := models.MaxDuePoolSize
	if dueLimit < maxCards {
		dueLimit = maxCards
	}
	q := `SELECT uvc.id, vtc.word_card_id, vtc.card_type, vtc.prompt_json, vtc.answer_json, COALESCE(vtc.distractors_json,'')
	      FROM user_verb_cards uvc
	      JOIN verb_training_cards vtc ON vtc.id = uvc.verb_training_card_id
	      INNER JOIN verb_forms_dict d ON d.id = vtc.verb_form_dict_id
	      WHERE uvc.user_id = ? AND vtc.card_type = ?
	        AND (uvc.next_due_at IS NULL OR uvc.next_due_at <= ?)` + verbTrainingEligibleByWordCardSQL("vtc") + verbTrainingPromptHasExampleTranslationSQL("vtc") + `
	      ORDER BY CASE WHEN uvc.state='learning' THEN 0 ELSE 1 END, uvc.next_due_at NULLS FIRST
	      LIMIT ?`
	rows, err := r.db.Query(q, userID, models.VerbCardTypeCloze, now, dueLimit)
	if err != nil {
		return nil, fmt.Errorf("get due verb queue: %w", err)
	}
	defer rows.Close()
	out := make([]VerbQueueCard, 0, maxCards+maxNew)
	seenIDs := make(map[int64]struct{}, maxCards+maxNew)
	for rows.Next() {
		var c VerbQueueCard
		if err := rows.Scan(&c.UserVerbCardID, &c.WordCardID, &c.CardType, &c.PromptJSON, &c.AnswerJSON, &c.DistractorsJSON); err != nil {
			return nil, err
		}
		out = append(out, c)
		seenIDs[c.UserVerbCardID] = struct{}{}
	}
	if maxNew > 0 {
		poolCap := maxNew * 30
		if poolCap > 600 {
			poolCap = 600
		}
		if poolCap < maxNew {
			poolCap = maxNew
		}
		nq := `SELECT uvc.id, vtc.word_card_id, vtc.card_type, vtc.prompt_json, vtc.answer_json, COALESCE(vtc.distractors_json,'')
		      FROM user_verb_cards uvc
		      JOIN verb_training_cards vtc ON vtc.id = uvc.verb_training_card_id
		      INNER JOIN verb_forms_dict d ON d.id = vtc.verb_form_dict_id
		      WHERE uvc.user_id = ? AND uvc.state='new' AND vtc.card_type = ?` + verbTrainingEligibleByWordCardSQL("vtc") + verbTrainingPromptHasExampleTranslationSQL("vtc") + `
		      ORDER BY random()
		      LIMIT ?`
		rowsNew, err := r.db.Query(nq, userID, models.VerbCardTypeCloze, poolCap)
		if err != nil {
			return nil, fmt.Errorf("get new verb queue: %w", err)
		}
		pool := make([]verbNewQueueRow, 0, poolCap)
		for rowsNew.Next() {
			var c VerbQueueCard
			if err := rowsNew.Scan(&c.UserVerbCardID, &c.WordCardID, &c.CardType, &c.PromptJSON, &c.AnswerJSON, &c.DistractorsJSON); err != nil {
				_ = rowsNew.Close()
				return nil, err
			}
			if _, exists := seenIDs[c.UserVerbCardID]; exists {
				continue
			}
			pool = append(pool, verbNewQueueRow{card: c, wordCardID: c.WordCardID})
		}
		if err := rowsNew.Close(); err != nil {
			return nil, err
		}
		out = append(out, roundRobinVerbNewCards(pool, maxNew)...)
	}
	if len(out) > maxCards {
		out = out[:maxCards]
	}
	return out, nil
}

// CountUserVerbClozeCards returns how many verb-form cloze cards exist for the user (full pool, not session queue).
func (r *VerbFormsRepository) CountUserVerbClozeCards(userID int64) (int64, error) {
	q := `SELECT COUNT(*) FROM user_verb_cards uvc
		INNER JOIN verb_training_cards vtc ON vtc.id = uvc.verb_training_card_id
		INNER JOIN verb_forms_dict d ON d.id = vtc.verb_form_dict_id
		WHERE uvc.user_id = ? AND vtc.card_type = ?` + verbTrainingEligibleByWordCardSQL("vtc") + verbTrainingPromptHasExampleTranslationSQL("vtc")
	var n int64
	if err := r.db.QueryRow(q, userID, models.VerbCardTypeCloze).Scan(&n); err != nil {
		return 0, fmt.Errorf("count user verb cloze cards: %w", err)
	}
	return n, nil
}

type VerbUserCardSRS struct {
	ID           int64
	State        string
	EF           float64
	Reps         int
	IntervalDays int
	LearningStep int
	LapseCount   int
}

func (r *VerbFormsRepository) GetVerbUserCardSRS(userVerbCardID int64) (*VerbUserCardSRS, error) {
	var c VerbUserCardSRS
	err := r.db.QueryRow(`SELECT id, state, ef, reps, interval_days, learning_step, lapse_count FROM user_verb_cards WHERE id=?`, userVerbCardID).
		Scan(&c.ID, &c.State, &c.EF, &c.Reps, &c.IntervalDays, &c.LearningStep, &c.LapseCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get verb user card srs: %w", err)
	}
	return &c, nil
}

func (r *VerbFormsRepository) UpdateVerbUserCardSRS(c *VerbUserCardSRS, nextDueAt time.Time, lastQuality int) error {
	_, err := r.db.Exec(`UPDATE user_verb_cards
		SET state=?, ef=?, reps=?, interval_days=?, learning_step=?, lapse_count=?, next_due_at=?, last_review_at=CURRENT_TIMESTAMP, last_quality=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		c.State, c.EF, c.Reps, c.IntervalDays, c.LearningStep, c.LapseCount, nextDueAt, lastQuality, c.ID)
	if err != nil {
		return fmt.Errorf("update verb user card srs: %w", err)
	}
	return nil
}

func (r *VerbFormsRepository) CreateVerbReviewEvent(sessionID, userID, userVerbCardID int64, isCorrect bool, quality int) error {
	correctInt := 0
	if isCorrect {
		correctInt = 1
	}
	_, err := r.db.Exec(`INSERT INTO verb_review_events
		(session_id, user_id, user_verb_card_id, shown_at, answered_at, is_correct, quality)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?)`,
		sessionID, userID, userVerbCardID, correctInt, quality)
	if err != nil {
		return fmt.Errorf("create verb review event: %w", err)
	}
	return nil
}

func (r *VerbFormsRepository) StartVerbSession(userID int64, planned int, sessionJSON string) (int64, error) {
	id, err := database.InsertAndReturnID(r.db, `INSERT INTO verb_training_sessions (user_id, planned_count, done_count, session_json, started_at) VALUES (?, ?, 0, ?, CURRENT_TIMESTAMP)`,
		userID, planned, sessionJSON)
	if err != nil {
		return 0, fmt.Errorf("start verb session: %w", err)
	}
	return id, nil
}

func (r *VerbFormsRepository) FinishVerbSession(sessionID int64, done int) error {
	_, err := r.db.Exec(`UPDATE verb_training_sessions SET done_count=?, ended_at=CURRENT_TIMESTAMP WHERE id=?`, done, sessionID)
	if err != nil {
		return fmt.Errorf("finish verb session: %w", err)
	}
	return nil
}

func (r *VerbFormsRepository) GetVerbSessionStats(sessionID int64) (totalCards int, correctCards int, err error) {
	query := `SELECT
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct
	FROM verb_review_events
	WHERE session_id = ? AND answered_at IS NOT NULL`
	err = r.db.QueryRow(query, sessionID).Scan(&totalCards, &correctCards)
	if err != nil {
		return 0, 0, fmt.Errorf("get verb session stats: %w", err)
	}
	return totalCards, correctCards, nil
}

// GetVerbLemmaMetadataJSONBatch returns metadata_json keyed by lowercased lemma (language=es).
func (r *VerbFormsRepository) GetVerbLemmaMetadataJSONBatch(lemmas []string) (map[string]string, error) {
	out := map[string]string{}
	if len(lemmas) == 0 {
		return out, nil
	}
	seen := map[string]struct{}{}
	list := make([]string, 0, len(lemmas))
	for _, l := range lemmas {
		k := strings.ToLower(strings.TrimSpace(l))
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		list = append(list, k)
	}
	if len(list) == 0 {
		return out, nil
	}
	ph := strings.Repeat("?,", len(list))
	ph = strings.TrimSuffix(ph, ",")
	args := make([]interface{}, 0, len(list))
	for _, l := range list {
		args = append(args, l)
	}
	q := `SELECT lemma, COALESCE(metadata_json,'') FROM verb_lemmas WHERE language='es' AND lemma IN (` + ph + `)`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("get verb lemma metadata batch: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var lemma, meta string
		if err := rows.Scan(&lemma, &meta); err != nil {
			return nil, fmt.Errorf("scan lemma metadata: %w", err)
		}
		out[strings.ToLower(strings.TrimSpace(lemma))] = meta
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateVerbLemmaMetadataJSON sets metadata_json for a Spanish lemma row.
func (r *VerbFormsRepository) UpdateVerbLemmaMetadataJSON(lemma, metadataJSON string) error {
	lemma = strings.ToLower(strings.TrimSpace(lemma))
	if lemma == "" {
		return fmt.Errorf("empty lemma")
	}
	_, err := r.db.Exec(`UPDATE verb_lemmas SET metadata_json=?, updated_at=CURRENT_TIMESTAMP WHERE lemma=? AND language='es'`,
		metadataJSON, lemma)
	if err != nil {
		return fmt.Errorf("update verb lemma metadata: %w", err)
	}
	return nil
}

// ListSpanishVerbLemmas returns all Spanish lemmas ordered lexicographically (for batch jobs).
func (r *VerbFormsRepository) ListSpanishVerbLemmas() ([]string, error) {
	rows, err := r.db.Query(`SELECT lemma FROM verb_lemmas WHERE language='es' ORDER BY lemma`)
	if err != nil {
		return nil, fmt.Errorf("list spanish verb lemmas: %w", err)
	}
	defer rows.Close()
	out := make([]string, 0, 700)
	for rows.Next() {
		var lemma string
		if err := rows.Scan(&lemma); err != nil {
			return nil, err
		}
		out = append(out, strings.TrimSpace(lemma))
	}
	return out, rows.Err()
}

// spanishLemmaLooksLikeInfinitiveSQL restricts lemmas to plausible Spanish infinitives:
// -ir/-ar/-er (length ≥ 4) plus short irregulars ir, dar, ser, ver.
// Excludes nouns/participles mistaken as lemmas (pasado, embargo, hecho, cuenta, …).
func spanishLemmaLooksLikeInfinitiveSQL() string {
	// Leading space so concatenation after ")" or ") < ?" never yields ")AND" / "?AND" (PostgreSQL 42601).
	return ` AND (
  LOWER(TRIM(w.word)) IN ('ir','dar','ser','ver')
  OR (
    LENGTH(LOWER(TRIM(w.word))) >= 4
    AND (
      LOWER(TRIM(w.word)) LIKE '%ar'
      OR LOWER(TRIM(w.word)) LIKE '%er'
      OR LOWER(TRIM(w.word)) LIKE '%ir'
    )
  )
)`
}

// ListPendingVerbTrainingLemmas lists Spanish infinitive headwords suitable for LLM verb-pack authoring.
// formsGapOnly=true (default for API): lemmas where verb_training_cards (cloze_form) count is below full V1 coverage
// (same count as one generated lemma file). So vocabulary verbs without any synced pack, or with a partial pack, appear here.
// Training UI only shows verbs that already have materialized cards (user_verb_cards); this endpoint lists the rest for generation.
// formsGapOnly=false: all infinitive-like headwords with verb_lemmas + a vocabulary training_card with verb POS (ignore pack completeness).
// Cursor is word_card_id based.
func (r *VerbFormsRepository) ListPendingVerbTrainingLemmas(limit int, cursorWordCardID int64, formsGapOnly bool) ([]PendingVerbLemmaRow, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}
	base := `SELECT w.id, LOWER(TRIM(w.word)) AS lemma
	      FROM word_cards w
	      INNER JOIN verb_lemmas vl ON vl.language = 'es' AND vl.lemma = LOWER(TRIM(w.word))
	      WHERE LOWER(TRIM(w.word)) <> ''
	        AND w.id > ?
	        AND EXISTS (
	          SELECT 1 FROM training_cards tc
	          WHERE tc.word_card_id = w.id
	            AND LOWER(TRIM(COALESCE(tc.pos, ''))) LIKE 'verb%'
	        )`
	tail := spanishLemmaLooksLikeInfinitiveSQL() + `
	      ORDER BY w.id ASC
	      LIMIT ?`
	var rows *sql.Rows
	var err error
	if formsGapOnly {
		full := verbtraining.FullCoverageClozeCardCountV1()
		gap := `
	        AND (
	          SELECT COUNT(*) FROM verb_training_cards vtc
	          WHERE vtc.word_card_id = w.id AND vtc.card_type = ?
	        ) < ?`
		rows, err = r.db.Query(base+gap+tail, cursorWordCardID, models.VerbCardTypeCloze, full, limit)
	} else {
		rows, err = r.db.Query(base+tail, cursorWordCardID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list pending verb training lemmas: %w", err)
	}
	defer rows.Close()
	out := make([]PendingVerbLemmaRow, 0, limit)
	for rows.Next() {
		var row PendingVerbLemmaRow
		if err := rows.Scan(&row.WordCardID, &row.Lemma); err != nil {
			return nil, fmt.Errorf("scan pending lemma: %w", err)
		}
		if row.WordCardID <= 0 || strings.TrimSpace(row.Lemma) == "" {
			continue
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// AdminVerbLemmaSummary is one Spanish lemma linked from vocabulary (word_verb_lemmas) for admin browsing.
type AdminVerbLemmaSummary struct {
	WordCardID int64  `json:"word_card_id"`
	Lemma      string `json:"lemma"`
	ClozeCount int64  `json:"cloze_count"`
	RuGloss    string `json:"ru_gloss,omitempty"`
}

// ListAdminVerbTrainingLemmas returns lemmas with verb-training cloze card counts (cursor on word_card.id).
func (r *VerbFormsRepository) ListAdminVerbTrainingLemmas(search string, limit int, afterWordCardID int64) ([]AdminVerbLemmaSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	search = strings.TrimSpace(search)
	args := []interface{}{afterWordCardID}
	q := `
SELECT w.id, LOWER(TRIM(w.word)),
  (SELECT COUNT(*) FROM verb_training_cards vtc
   WHERE vtc.word_card_id = w.id AND vtc.card_type = '` + models.VerbCardTypeCloze + `'),
  COALESCE(vl.metadata_json,'')
FROM word_cards w
INNER JOIN word_verb_lemmas wvl ON wvl.word_card_id = w.id
INNER JOIN verb_lemmas vl ON vl.id = wvl.verb_lemma_id AND vl.language = 'es'
WHERE w.id > ?`
	if search != "" {
		q += ` AND LOWER(TRIM(w.word)) LIKE ?`
		args = append(args, "%"+strings.ToLower(search)+"%")
	}
	q += `
ORDER BY w.id ASC
LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin verb training lemmas: %w", err)
	}
	defer rows.Close()
	out := make([]AdminVerbLemmaSummary, 0, limit)
	for rows.Next() {
		var row AdminVerbLemmaSummary
		var meta string
		if err := rows.Scan(&row.WordCardID, &row.Lemma, &row.ClozeCount, &meta); err != nil {
			return nil, fmt.Errorf("scan admin verb lemma: %w", err)
		}
		row.RuGloss = spanishverbs.RuGlossFromLemmaMetadataJSON(meta)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// AdminVerbTrainingCardDetail is one materialized verb_training_cards row with conjugation slice for admin UI.
type AdminVerbTrainingCardDetail struct {
	ID          int64           `json:"id"`
	CardType    string          `json:"card_type"`
	Mood        string          `json:"mood"`
	Tense       string          `json:"tense"`
	Person      string          `json:"person"`
	Number      string          `json:"number"`
	SurfaceForm string          `json:"surface_form"`
	Prompt      json.RawMessage `json:"prompt"`
	Answer      json.RawMessage `json:"answer"`
	Distractors json.RawMessage `json:"distractors,omitempty"`
}

// ListAdminVerbTrainingCardsByWordCard returns all verb training cards for a word_card_id with joined dictionary slice.
func (r *VerbFormsRepository) ListAdminVerbTrainingCardsByWordCard(wordCardID int64) ([]AdminVerbTrainingCardDetail, error) {
	if wordCardID <= 0 {
		return nil, fmt.Errorf("invalid word_card_id")
	}
	q := `SELECT vtc.id, vtc.card_type,
	  vtc.prompt_json, vtc.answer_json, COALESCE(vtc.distractors_json,''),
	  d.mood, d.tense, d.person, d.number, d.surface_form
	FROM verb_training_cards vtc
	INNER JOIN verb_forms_dict d ON d.id = vtc.verb_form_dict_id
	WHERE vtc.word_card_id = ?
	ORDER BY d.mood, d.tense, d.person, d.number, vtc.id`
	rows, err := r.db.Query(q, wordCardID)
	if err != nil {
		return nil, fmt.Errorf("list admin verb training cards: %w", err)
	}
	defer rows.Close()
	out := make([]AdminVerbTrainingCardDetail, 0, 64)
	for rows.Next() {
		var row AdminVerbTrainingCardDetail
		var promptStr, answerStr, distStr string
		if err := rows.Scan(&row.ID, &row.CardType, &promptStr, &answerStr, &distStr, &row.Mood, &row.Tense, &row.Person, &row.Number, &row.SurfaceForm); err != nil {
			return nil, fmt.Errorf("scan admin verb training card: %w", err)
		}
		row.Prompt = json.RawMessage(strings.TrimSpace(promptStr))
		if len(row.Prompt) == 0 {
			row.Prompt = json.RawMessage("{}")
		}
		row.Answer = json.RawMessage(strings.TrimSpace(answerStr))
		if len(row.Answer) == 0 {
			row.Answer = json.RawMessage("{}")
		}
		d := strings.TrimSpace(distStr)
		if d != "" {
			row.Distractors = json.RawMessage(d)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	SortAdminVerbTrainingCardDetailsSpanish(out)
	return out, nil
}

// AdminVerbLemmaLookup resolves lemma for a word_card linked to Spanish verb_lemmas (admin).
func (r *VerbFormsRepository) AdminVerbLemmaLookup(wordCardID int64) (lemma string, ok bool, err error) {
	if wordCardID <= 0 {
		return "", false, nil
	}
	err = r.db.QueryRow(`
SELECT LOWER(TRIM(w.word))
FROM word_cards w
INNER JOIN word_verb_lemmas wvl ON wvl.word_card_id = w.id
INNER JOIN verb_lemmas vl ON vl.id = wvl.verb_lemma_id AND vl.language = 'es'
WHERE w.id = ?`, wordCardID).Scan(&lemma)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("admin verb lemma lookup: %w", err)
	}
	return strings.TrimSpace(lemma), true, nil
}

const verbExampleCatalogTTL = 5 * time.Minute

var (
	verbExampleCatalogMu   sync.Mutex
	verbExampleCatalogTpl  []spanishverbs.CatalogTemplate
	verbExampleCatalogAt   time.Time
	verbExampleCatalogHave bool // false until first successful load (including empty result)
)

// ListVerbExampleCatalogTemplatesCached loads verb_example_templates with a short in-process TTL.
func (r *VerbFormsRepository) ListVerbExampleCatalogTemplatesCached() ([]spanishverbs.CatalogTemplate, error) {
	verbExampleCatalogMu.Lock()
	defer verbExampleCatalogMu.Unlock()
	if verbExampleCatalogHave && time.Since(verbExampleCatalogAt) < verbExampleCatalogTTL {
		out := make([]spanishverbs.CatalogTemplate, len(verbExampleCatalogTpl))
		copy(out, verbExampleCatalogTpl)
		return out, nil
	}
	tpl, err := r.listVerbExampleCatalogTemplatesUncached()
	if err != nil {
		return nil, err
	}
	verbExampleCatalogTpl = tpl
	verbExampleCatalogAt = time.Now()
	verbExampleCatalogHave = true
	out := make([]spanishverbs.CatalogTemplate, len(tpl))
	copy(out, tpl)
	return out, nil
}

func (r *VerbFormsRepository) listVerbExampleCatalogTemplatesUncached() ([]spanishverbs.CatalogTemplate, error) {
	rows, err := r.db.Query(`SELECT code, lemma_match, COALESCE(verb_class,''), COALESCE(mood,''), COALESCE(tense,''), es_suffix, ru_pattern
		FROM verb_example_templates WHERE active = true ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list verb example templates: %w", err)
	}
	defer rows.Close()
	var out []spanishverbs.CatalogTemplate
	for rows.Next() {
		var code, lemma, vclass, mood, tense, esSuf, ruPat string
		if err := rows.Scan(&code, &lemma, &vclass, &mood, &tense, &esSuf, &ruPat); err != nil {
			return nil, fmt.Errorf("scan verb example template: %w", err)
		}
		out = append(out, spanishverbs.CatalogTemplate{
			ID:         strings.TrimSpace(code),
			VerbClass:  strings.TrimSpace(vclass),
			Mood:       strings.TrimSpace(mood),
			Tense:      strings.TrimSpace(tense),
			EsSuffix:   esSuf,
			RuPattern:  ruPat,
			LemmaMatch: strings.ToLower(strings.TrimSpace(lemma)),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ResetVerbExampleCatalogCacheForTests clears cached DB templates (tests only).
func ResetVerbExampleCatalogCacheForTests() {
	verbExampleCatalogMu.Lock()
	defer verbExampleCatalogMu.Unlock()
	verbExampleCatalogHave = false
	verbExampleCatalogTpl = nil
	verbExampleCatalogAt = time.Time{}
}
