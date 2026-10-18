package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

type DailyJourneyWordsAdapter struct {
	db                      *sql.DB
	logger                  *zap.Logger
	learning                config.LearningConfig
	MirrorSRS               bool
	OptionsDelayMS          int
	WrongAnswerDelaySeconds int
}

func NewDailyJourneyWordsAdapter(db *sql.DB, logger *zap.Logger, learning config.LearningConfig) *DailyJourneyWordsAdapter {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DailyJourneyWordsAdapter{db: db, logger: logger, learning: learning}
}

type journeyWordsState struct {
	WrongAnswerDelaySeconds int                         `json:"wrong_answer_delay_seconds"`
	DelayMS                 int                         `json:"delay_ms"`
	Phase                   string                      `json:"phase"`
	DueWords                int                         `json:"due_words"`
	AddTarget               int                         `json:"add_target"`
	Added                   []int64                     `json:"added"`
	Queue                   []*models.TrainingQueueItem `json:"queue"`
	Index                   int                         `json:"index"`
	Correct                 int                         `json:"correct"`
	Card                    map[string]any              `json:"card"`
	Options                 []string                    `json:"options"`
	CorrectAnswer           string                      `json:"correct_answer"`
	ShownAt                 time.Time                   `json:"shown_at"`
	OptionsShownAt          *time.Time                  `json:"options_shown_at"`
	Feedback                map[string]any              `json:"feedback"`
}
type JourneyWordsExercise struct {
	Phase      string            `json:"phase"`
	AddTarget  int               `json:"add_target"`
	AddedCount int               `json:"added_count"`
	Addable    map[int64][]int64 `json:"addable"`
	Card       map[string]any    `json:"card,omitempty"`
	Feedback   map[string]any    `json:"feedback,omitempty"`
	Options    []string          `json:"options,omitempty"`
}

func (a *DailyJourneyWordsAdapter) openSets(ctx context.Context, q database.ContextDBTX, user int64, course string) (map[int64]bool, error) {
	levels, err := dailyJourneyOpenLevels(ctx, a.db, q, user, course)
	if err != nil {
		return nil, err
	}
	type category struct {
		parent    sql.NullInt64
		level     sql.NullString
		published bool
	}
	cats := map[int64]category{}
	rows, err := q.QueryContext(ctx, `SELECT id,parent_id,level_code,is_published=1 FROM word_set_categories WHERE course_code=?`, course)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var c category
		if err := rows.Scan(&id, &c.parent, &c.level, &c.published); err != nil {
			rows.Close()
			return nil, err
		}
		cats[id] = c
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = q.QueryContext(ctx, `SELECT id,category_id,level_code FROM word_sets WHERE course_code=? AND is_published=1`, course)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sets := map[int64]bool{}
	for rows.Next() {
		var id int64
		var parent sql.NullInt64
		var level sql.NullString
		if err := rows.Scan(&id, &parent, &level); err != nil {
			return nil, err
		}
		allowed := !level.Valid || levels[strings.ToUpper(level.String)]
		visited := map[int64]bool{}
		for parent.Valid {
			c, exists := cats[parent.Int64]
			if !exists || visited[parent.Int64] || !c.published || (c.level.Valid && !levels[strings.ToUpper(c.level.String)]) {
				allowed = false
				break
			}
			visited[parent.Int64] = true
			parent = c.parent
		}
		if allowed {
			sets[id] = true
		}
	}
	return sets, rows.Err()
}
func (a *DailyJourneyWordsAdapter) addable(ctx context.Context, q database.ContextDBTX, user int64, course string) (map[int64][]int64, error) {
	return a.addableWith(ctx, q, user, course)
}
func (a *DailyJourneyWordsAdapter) addableWith(ctx context.Context, q database.ContextDBTX, user int64, course string) (map[int64][]int64, error) {
	sets, err := a.openSets(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, `SELECT wi.word_set_id,wi.word_card_id FROM word_set_items wi JOIN word_cards wc ON wc.id=wi.word_card_id WHERE wc.course_code=? AND EXISTS(SELECT 1 FROM training_cards tc WHERE tc.word_card_id=wc.id AND tc.word_ru<>'' AND tc.word_en<>'') AND NOT EXISTS(SELECT 1 FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=? AND tc.word_card_id=wc.id) AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=? AND k.word_card_id=wc.id AND k.status='known') ORDER BY wi.word_set_id,wi.sort_order,wi.word_card_id`, course, user, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[int64][]int64{}
	for rows.Next() {
		var set, word int64
		if err := rows.Scan(&set, &word); err != nil {
			return nil, err
		}
		if sets[set] {
			result[set] = append(result[set], word)
		}
	}
	return result, rows.Err()
}
func uniqueAddable(pool map[int64][]int64) int {
	seen := map[int64]bool{}
	for _, words := range pool {
		for _, id := range words {
			seen[id] = true
		}
	}
	return len(seen)
}
func journeyDueWordIDs(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]int64, error) {
	rows, err := q.QueryContext(ctx, `SELECT DISTINCT ON(tc.word_card_id) uc.id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id JOIN word_cards wc ON wc.id=tc.word_card_id WHERE uc.user_id=? AND wc.course_code=? AND (uc.next_due_at IS NULL OR uc.next_due_at<=CURRENT_TIMESTAMP) AND tc.word_ru<>'' AND tc.word_en<>'' AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=uc.user_id AND k.word_card_id=wc.id AND k.status='known') ORDER BY tc.word_card_id,uc.reps,uc.id`, user, course)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
func (a *DailyJourneyWordsAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	ids, err := journeyDueWordIDs(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	pool, err := a.addable(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 && uniqueAddable(pool) == 0 {
		return nil, nil
	}
	ref, _ := json.Marshal(map[string]string{"course_code": course})
	return []dailyjourney.Step{{Mode: dailyjourney.Words, Title: "Слова", Goal: "Добавь новые слова и повтори знакомые", TargetCount: 20, ContentRef: ref, ContentVersion: "word-training-v1", Launch: &dailyjourney.Launch{RouteName: "Training"}}}, nil
}
func (a *DailyJourneyWordsAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	var ref struct {
		Course string `json:"course_code"`
	}
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return err
	}
	if ref.Course != course {
		return dailyjourney.ErrAccess
	}
	var allowed bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=? AND uc.status='active' AND c.status='active')`, user, course).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return dailyjourney.ErrAccess
	}
	var state journeyWordsState
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	if state.Phase == "train" && state.Index < len(state.Queue) {
		item := state.Queue[state.Index]
		var cardID int64
		switch item.Type {
		case "card":
			cardID = item.Card.UserCard.ID
		case "spell":
			cardID = item.Spell.ReplacedUserCardID
		case "type":
			cardID = item.TypeChallenge.ReplacedUserCardID
		default:
			return dailyjourney.ErrInvalidAction
		}
		var exists bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id JOIN word_cards wc ON wc.id=tc.word_card_id WHERE uc.id=? AND uc.user_id=? AND wc.course_code=?)`, cardID, user, course).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrDailyJourneyContentUnavailable
		}
	}
	return nil
}
func (a *DailyJourneyWordsAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	var state journeyWordsState
	if len(step.State) > 0 && string(step.State) != "{}" {
		if err := json.Unmarshal(step.State, &state); err != nil {
			return err
		}
	}
	switch action.Type {
	case "start":
		var payload struct {
			Reveal bool `json:"reveal"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if state.Phase == "" {
			ids, err := journeyDueWordIDs(ctx, tx, user, j.CourseCode)
			if err != nil {
				return err
			}
			pool, err := a.addable(ctx, tx, user, j.CourseCode)
			if err != nil {
				return err
			}
			state.DueWords = len(ids)
			state.AddTarget = min(dailyjourney.WordAdditionTarget(len(ids)), uniqueAddable(pool))
			state.Added = []int64{}
			if state.AddTarget > 0 {
				state.Phase = "add"
				step.Launch = &dailyjourney.Launch{RouteName: "WordSets"}
				step.TargetCount = state.AddTarget
			} else {
				if err := a.prepare(ctx, tx, user, j, step, &state); err != nil {
					return err
				}
			}
		}
		if payload.Reveal {
			if state.Phase != "train" || state.Feedback != nil {
				return dailyjourney.ErrInvalidAction
			}
			if state.OptionsShownAt == nil {
				now := time.Now()
				state.OptionsShownAt = &now
			}
		}
	case "add_word":
		var payload struct {
			SetID  int64 `json:"set_id"`
			WordID int64 `json:"word_card_id"`
			Known  bool  `json:"known"`
		}
		if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
			return err
		}
		if state.Phase != "add" {
			return dailyjourney.ErrInvalidAction
		}
		for _, id := range state.Added {
			if id == payload.WordID {
				step.State, _ = json.Marshal(state)
				return nil
			}
		}
		sets, err := a.openSets(ctx, tx, user, j.CourseCode)
		if err != nil {
			return err
		}
		if !sets[payload.SetID] {
			return dailyjourney.ErrAccess
		}
		var wordID int64
		err = tx.QueryRowContext(ctx, `SELECT wc.id FROM word_set_items wi JOIN word_cards wc ON wc.id=wi.word_card_id WHERE wi.word_set_id=? AND wc.id=? AND wc.course_code=? FOR UPDATE OF wc`, payload.SetID, payload.WordID, j.CourseCode).Scan(&wordID)
		if err == sql.ErrNoRows {
			return dailyjourney.ErrAccess
		}
		if err != nil {
			return err
		}
		var enrolled, known bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=? AND tc.word_card_id=?),EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=? AND k.word_card_id=? AND k.status='known')`, user, wordID, user, wordID).Scan(&enrolled, &known); err != nil {
			return err
		}
		if !enrolled && !known {
			if payload.Known {
				if _, err := tx.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,word_card_id,status) VALUES(?,?,'known') ON CONFLICT(user_id,word_card_id) DO NOTHING`, user, wordID); err != nil {
					return err
				}
			} else {
				cards, err := repository.NewTrainingCardRepository(tx, a.logger).GetTrainingCardsByWordCardID(wordID)
				if err != nil {
					return err
				}
				if len(cards) == 0 {
					return fmt.Errorf("word training cards are not ready")
				}
				repo := repository.NewUserCardRepository(tx, a.logger)
				for _, tc := range cards {
					if tc.WordEN == "" || tc.WordRU == "" {
						continue
					}
					for _, direction := range []models.CardDirection{models.DirectionRUtoEN, models.DirectionENtoRU} {
						if _, err := repo.CreateUserCard(&models.UserCard{UserID: user, TrainingCardID: tc.ID, Direction: direction, State: models.StateNew, EF: models.InitialEF}); err != nil {
							return err
						}
					}
				}
				if err := repository.NewUserWordMasteringRepository(tx, a.logger).Upsert(user, wordID, 0); err != nil {
					return err
				}
				if a.MirrorSRS {
					lc := a.learning
					parts := strings.SplitN(j.CourseCode, "_", 2)
					lc.TargetLang = parts[0]
					if len(parts) > 1 {
						lc.NativeLang = parts[1]
					}
					if err := repository.NewLinglowWordSRSBackfillRepository(tx).SyncWordLearningItemForUser(ctx, lc, user, wordID); err != nil {
						return err
					}
				}
				state.Added = append(state.Added, wordID)
				step.CompletedCount = len(state.Added)
			}
		}
		remaining, err := a.addableWith(ctx, tx, user, j.CourseCode)
		if err != nil {
			return err
		}
		state.AddTarget = min(state.AddTarget, len(state.Added)+uniqueAddable(remaining))
		if state.AddTarget == 0 && state.DueWords == 0 {
			step.Status = "unavailable"
			reason := "Нет готовых слов для тренировки"
			step.Result = &reason
		} else if len(state.Added) >= state.AddTarget {
			if err := a.prepare(ctx, tx, user, j, step, &state); err != nil {
				return err
			}
		}
	case "answer":
		if state.Phase != "train" || state.Feedback != nil || state.Index >= len(state.Queue) {
			return dailyjourney.ErrInvalidAction
		}
		if err := a.answer(ctx, tx, user, j, step, &state, action); err != nil {
			return err
		}
	case "ack_feedback":
		var empty struct{}
		if err := decodeJourneyPayload(action.Payload, &empty); err != nil {
			return err
		}
		if state.Feedback == nil {
			return dailyjourney.ErrInvalidAction
		}
		state.Feedback = nil
		step.Feedback = nil
		state.Index++
		if state.Index >= len(state.Queue) {
			step.Status = "completed"
			result := fmt.Sprintf("Повторено %d слов, добавлено %d", len(state.Queue), len(state.Added))
			step.Result = &result
			step.ResultDetails, _ = json.Marshal(map[string]any{"added_words": len(state.Added), "reviewed_words": len(state.Queue), "correct": state.Correct})
		} else {
			if err := a.card(ctx, tx, user, j, step, &state); err != nil {
				return err
			}
		}
	default:
		return dailyjourney.ErrInvalidAction
	}
	var err error
	step.State, err = json.Marshal(state)
	return err
}
func (a *DailyJourneyWordsAdapter) prepare(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, state *journeyWordsState) error {
	ids, err := journeyDueWordIDs(ctx, tx, user, j.CourseCode)
	if err != nil {
		return err
	}
	repo := repository.NewUserCardRepository(tx, a.logger)
	cards := []*models.UserCard{}
	// Newly added words are fetched explicitly: an out-of-band review cannot
	// push them out of this queue by changing their due date.
	for _, word := range state.Added {
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT uc.id FROM user_cards uc JOIN training_cards tc ON tc.id=uc.training_card_id WHERE uc.user_id=? AND tc.word_card_id=? AND tc.word_ru<>'' AND tc.word_en<>'' ORDER BY uc.reps,uc.id LIMIT 1`, user, word).Scan(&id); err != nil {
			return err
		}
		card, err := repo.GetUserCard(id)
		if err != nil {
			return err
		}
		if card != nil {
			cards = append(cards, card)
		}
	}
	rand.Shuffle(len(ids), func(i, k int) { ids[i], ids[k] = ids[k], ids[i] })
	for _, id := range ids {
		card, err := repo.GetUserCard(id)
		if err != nil {
			return err
		}
		if card != nil {
			cards = append(cards, card)
		}
		if len(cards) >= 20+len(state.Added) {
			break
		}
	}
	cfg := DefaultSessionConfig()
	cfg.CourseCode = j.CourseCode
	cfg.IsEnglishTarget = strings.HasPrefix(j.CourseCode, "en_")
	// Respect the same per-user spell/type settings as normal training.
	var raw string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(settings_json,'') FROM users WHERE id=?`, user).Scan(&raw); err != nil {
		return err
	}
	var settings models.UserSettings
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			return err
		}
	}
	if settings.SpellModeEnabled != nil {
		cfg.SpellEnabled = *settings.SpellModeEnabled
	}
	if settings.TypeModeEnabled != nil {
		cfg.TypeEnabled = *settings.TypeModeEnabled
	}
	if settings.SpellMasteringThreshold != nil {
		cfg.SpellMasteringThreshold = *settings.SpellMasteringThreshold
	}
	if settings.TypeMasteringThreshold != nil {
		cfg.TypeMasteringThreshold = *settings.TypeMasteringThreshold
	}
	svc := NewTrainingService(repo, repository.NewTrainingCardRepository(tx, a.logger), repository.NewSessionRepository(tx, a.logger), repository.NewUserWordMasteringRepository(tx, a.logger), a.learning, a.logger)
	queue, err := svc.GenerateJourneyQueue(user, cards, *cfg)
	if err != nil {
		return err
	}
	if len(queue) == 0 {
		return fmt.Errorf("no ready words for daily journey")
	}
	state.Phase = "train"
	state.Queue = queue
	state.Index = 0
	step.CompletedCount = 0
	step.TargetCount = len(queue)
	step.Launch = &dailyjourney.Launch{RouteName: "Training"}
	return a.card(ctx, tx, user, j, step, state)
}
func journeySessionID(id string) int64 {
	sum := sha256.Sum256([]byte(id))
	return -int64(binary.LittleEndian.Uint64(sum[:8])%(1<<52) + 1)
}
func (a *DailyJourneyWordsAdapter) card(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, state *journeyWordsState) error {
	item := state.Queue[state.Index]
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	state.Options = nil
	state.CorrectAnswer = ""
	state.DelayMS = a.OptionsDelayMS
	state.WrongAnswerDelaySeconds = a.WrongAnswerDelaySeconds
	var raw string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(settings_json,'') FROM users WHERE id=?`, user).Scan(&raw); err != nil {
		return err
	}
	var settings models.UserSettings
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			return err
		}
	}
	if settings.OptionsDelaySeconds != nil {
		state.DelayMS = *settings.OptionsDelaySeconds * 1000
	}
	if settings.WrongAnswerDelaySeconds != nil {
		state.WrongAnswerDelaySeconds = *settings.WrongAnswerDelaySeconds
	}
	state.WrongAnswerDelaySeconds = max(0, state.WrongAnswerDelaySeconds)
	state.DelayMS = max(0, state.DelayMS)
	response := map[string]any{"type": item.Type, "session_id": journeySessionID(step.ID), "card_index": state.Index + 1, "total_cards": len(state.Queue), "delay_ms": state.DelayMS}
	switch item.Type {
	case "card":
		c := item.Card
		if c == nil {
			return dailyjourney.ErrInvalidAction
		}
		cfg := a.learning
		cfg.TargetLang = strings.SplitN(j.CourseCode, "_", 2)[0]
		opts, correct, err := NewOptionsService(repository.NewTrainingCardRepository(tx, a.logger), a.logger, cfg.TargetLang).GenerateOptions(c, models.DefaultOptionCount, nil, nil, nil)
		if err != nil {
			return err
		}
		state.Options = opts
		state.CorrectAnswer = correct
		response["user_card_id"] = c.UserCard.ID
		response["training_card_id"] = c.TrainingCard.ID
		response["word_card_id"] = c.TrainingCard.WordCardID
		response["direction"] = c.UserCard.Direction
		response["example_target"] = c.TrainingCard.ExampleEN
		response["example_en"] = c.TrainingCard.ExampleEN
		response["word_category"] = c.TrainingCard.POS
		response["transcription"] = c.TrainingCard.Transcription
		if c.UserCard.Direction == models.DirectionRUtoEN {
			response["question"] = "Переведи: " + c.TrainingCard.WordRU
			response["word_native"] = c.TrainingCard.WordRU
			response["word_ru"] = c.TrainingCard.WordRU
		} else {
			response["question"] = "Что означает: " + c.TrainingCard.WordEN
			response["word_target"] = c.TrainingCard.WordEN
			response["word_en"] = c.TrainingCard.WordEN
			response["display_word"] = TrainingDisplayWord(&c.TrainingCard)
		}
	case "spell":
		c := item.Spell
		if c == nil {
			return dailyjourney.ErrInvalidAction
		}
		state.CorrectAnswer = c.DisplayWord
		if !strings.HasPrefix(j.CourseCode, "en_") {
			state.CorrectAnswer = strings.TrimPrefix(state.CorrectAnswer, "to ")
		}
		response["question"] = "Составь слово: " + c.WordRU
		response["word_native"] = c.WordRU
		response["word_ru"] = c.WordRU
		response["user_card_id"] = c.ReplacedUserCardID
		response["word_card_id"] = c.WordCardID
		response["direction"] = "spell"
		response["letters"] = c.ShuffledLetters
		response["prefix"] = c.Prefix
	case "type":
		c := item.TypeChallenge
		if c == nil {
			return dailyjourney.ErrInvalidAction
		}
		state.CorrectAnswer = c.DisplayWord
		if !strings.HasPrefix(j.CourseCode, "en_") {
			state.CorrectAnswer = strings.TrimPrefix(state.CorrectAnswer, "to ")
		}
		word := state.CorrectAnswer
		prefix := ""
		if strings.HasPrefix(j.CourseCode, "en_") && strings.HasPrefix(word, "to ") {
			prefix = "to "
			word = strings.TrimPrefix(word, prefix)
		}
		first := ""
		if word != "" {
			r, _ := utf8.DecodeRuneInString(word)
			first = string(r)
		}
		response["question"] = "Введи слово: " + c.WordRU
		response["word_native"] = c.WordRU
		response["word_ru"] = c.WordRU
		response["user_card_id"] = c.ReplacedUserCardID
		response["word_card_id"] = c.WordCardID
		response["direction"] = "type"
		response["prefix"] = prefix
		response["hint_length"] = utf8.RuneCountInString(word)
		response["hint_first_letter"] = first
	default:
		return dailyjourney.ErrInvalidAction
	}
	state.Card = response
	return nil
}
func (a *DailyJourneyWordsAdapter) answer(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, state *journeyWordsState, action dailyjourney.Action) error {
	var payload struct {
		CardIndex   int     `json:"card_index"`
		UserCardID  int64   `json:"user_card_id"`
		OptionIndex *int    `json:"option_index"`
		AnswerText  *string `json:"answer_text"`
	}
	if err := decodeJourneyPayload(action.Payload, &payload); err != nil {
		return err
	}
	if payload.CardIndex != state.Index+1 {
		return dailyjourney.ErrConflict
	}
	item := state.Queue[state.Index]
	cardID := int64(0)
	chosen := ""
	options := state.Options
	if item.Type == "card" {
		cardID = item.Card.UserCard.ID
		if payload.OptionIndex == nil || payload.AnswerText != nil || state.OptionsShownAt == nil || *payload.OptionIndex < 0 || *payload.OptionIndex >= len(options) {
			return dailyjourney.ErrInvalidAction
		}
		chosen = options[*payload.OptionIndex]
	} else {
		if payload.AnswerText == nil || payload.OptionIndex != nil {
			return dailyjourney.ErrInvalidAction
		}
		chosen = strings.TrimSpace(*payload.AnswerText)
		if item.Type == "spell" {
			cardID = item.Spell.ReplacedUserCardID
		} else {
			cardID = item.TypeChallenge.ReplacedUserCardID
		}
		options = []string{}
	}
	if payload.UserCardID != cardID {
		return dailyjourney.ErrConflict
	}
	var locked int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM user_cards WHERE id=? AND user_id=? FOR UPDATE`, cardID, user).Scan(&locked); err != nil {
		if err == sql.ErrNoRows {
			return ErrDailyJourneyContentUnavailable
		}
		return err
	}
	repo := repository.NewUserCardRepository(tx, a.logger)
	card, err := repo.GetUserCard(cardID)
	if err != nil {
		return err
	}
	if card == nil {
		return ErrDailyJourneyContentUnavailable
	}
	correct := strings.EqualFold(strings.TrimSpace(chosen), strings.TrimSpace(state.CorrectAnswer))
	now := time.Now()
	answerMS := max(0, int(now.Sub(state.ShownAt).Milliseconds()))
	delayMS := 0
	if state.OptionsShownAt != nil {
		answerMS = max(0, int(now.Sub(*state.OptionsShownAt).Milliseconds()))
		delayMS = max(0, int(state.OptionsShownAt.Sub(state.ShownAt).Milliseconds()))
	}
	attempt := models.AttemptData{Correct: correct, ChosenOption: chosen, OptionCount: len(options), AnswerTimeMS: answerMS, TDelayMS: delayMS, GradedAt: &now, EarlyReveal: item.Type == "card" && delayMS < state.DelayMS}
	if item.Type != "card" {
		attempt.OptionCount = 1
		attempt.TimeMultiplier = models.TimeMultiplierForMode(item.Type, utf8.RuneCountInString(state.CorrectAnswer))
	}
	before := srsStateForJourney(card)
	srs := NewSRSService(repo, a.learning, a.logger)
	if !correct {
		if err := srs.RecordWrongAnswer(card, chosen); err != nil {
			return err
		}
	}
	if err := srs.GradeCard(card, attempt); err != nil {
		return err
	}
	tc, err := repository.NewTrainingCardRepository(tx, a.logger).GetTrainingCard(card.TrainingCardID)
	if err != nil {
		return err
	}
	if tc == nil {
		return ErrDailyJourneyContentUnavailable
	}
	if a.MirrorSRS {
		lc := a.learning
		parts := strings.SplitN(j.CourseCode, "_", 2)
		lc.TargetLang = parts[0]
		if len(parts) > 1 {
			lc.NativeLang = parts[1]
		}
		if err := repository.NewLinglowWordSRSBackfillRepository(tx).SyncWordLearningItemForUser(ctx, lc, user, tc.WordCardID); err != nil {
			return err
		}
	}
	after := srsStateForJourney(card)
	opts, _ := json.Marshal(options)
	metrics, _ := json.Marshal(map[string]any{"mode": item.Type, "answer_time_ms": answerMS, "daily_journey_id": j.ID, "daily_journey_step_id": step.ID})
	event := &models.ReviewEvent{UserID: user, UserCardID: cardID, CourseCode: j.CourseCode, ClientAttemptID: "daily-journey:" + j.ID + ":" + action.ClientActionID, Direction: card.Direction, ShownAt: state.ShownAt, OptionsShownAt: state.OptionsShownAt, AnsweredAt: &now, TDelayMS: delayMS, EarlyReveal: attempt.EarlyReveal, OptionCount: attempt.OptionCount, OptionsJSON: string(opts), ChosenOption: chosen, IsCorrect: correct, Quality: int(models.CalculateQuality(attempt)), MetricsJSON: string(metrics), SRSBeforeJSON: before, SRSAfterJSON: after}
	reviewID, err := repository.NewSessionRepository(tx, a.logger).CreateReviewEvent(event)
	if err != nil {
		return err
	}
	if _, err := repository.NewLinglowEventRepository(a.db).RecordWordReviewEventTx(ctx, tx, a.learning, repository.WordReviewEventInput{UserID: user, ReviewEventID: reviewID, CourseCode: j.CourseCode, UserCardID: cardID, Direction: string(card.Direction), IsCorrect: correct, Quality: event.Quality, OptionsJSON: event.OptionsJSON, ChosenOption: chosen, MetricsJSON: event.MetricsJSON, SRSBeforeJSON: before, SRSAfterJSON: after, ClientAttemptID: "daily-journey:" + j.ID + ":" + action.ClientActionID, AnsweredAt: now}); err != nil {
		return err
	}
	master := repository.NewUserWordMasteringRepository(tx, a.logger)
	pair := repository.UserWordPair{UserID: user, WordCardID: tc.WordCardID}
	stats, err := master.GetWordMasteringStatsBatch([]repository.UserWordPair{pair})
	if err != nil {
		return err
	}
	known, err := master.GetKnownForPairs([]repository.UserWordPair{pair})
	if err != nil {
		return err
	}
	score := ComputeMasteringScore(stats[pair].Total, stats[pair].Correct, stats[pair].RecentTotal, stats[pair].RecentCorrect, known[pair])
	if err := master.Upsert(user, tc.WordCardID, score); err != nil {
		return err
	}
	if correct {
		state.Correct++
	}
	step.CompletedCount++
	state.Feedback = map[string]any{"is_correct": correct, "chosen_option": chosen, "correct_answer": state.CorrectAnswer}
	if !correct {
		state.Feedback["delay_seconds"] = state.WrongAnswerDelaySeconds
	}
	if item.Type == "card" {
		state.Feedback["hint"] = item.Card.TrainingCard.Hint
		state.Feedback["example"] = item.Card.TrainingCard.ExampleEN
		state.Feedback["example_target"] = item.Card.TrainingCard.ExampleEN
	}
	text := "Ответ принят"
	step.Feedback = &text
	return nil
}
func srsStateForJourney(card *models.UserCard) string {
	data, _ := json.Marshal(models.SRSState{State: card.State, EF: card.EF, Reps: card.Reps, IntervalDays: card.IntervalDays, LearningStep: card.LearningStep, LapseCount: card.LapseCount})
	return string(data)
}
func (a *DailyJourneyWordsAdapter) Exercise(ctx context.Context, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (any, error) {
	var state journeyWordsState
	if err := json.Unmarshal(step.State, &state); err != nil {
		return nil, err
	}
	view := JourneyWordsExercise{Phase: state.Phase, AddTarget: state.AddTarget, AddedCount: len(state.Added), Card: state.Card, Feedback: state.Feedback}
	if state.Phase == "add" {
		pool, err := a.addable(ctx, a.db, user, j.CourseCode)
		if err != nil {
			return nil, err
		}
		view.Addable = pool
	}
	if state.OptionsShownAt != nil {
		view.Options = state.Options
	}
	return view, nil
}

// Keep database.DBTX explicit here to document the shared transactional engine contract.
var _ database.DBTX = (*sql.Tx)(nil)
