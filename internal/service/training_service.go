package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	learncfg "tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// userWordMasteringRepoForSession is used by TrainingService for FinishSession and generateQueue (allows mocks in tests).
type userWordMasteringRepoForSession interface {
	GetWordCardIDsBySessionID(sessionID int64) ([]repository.UserWordPair, error)
	GetWordMasteringStatsBatch(pairs []repository.UserWordPair) (map[repository.UserWordPair]repository.WordMasteringStatsRow, error)
	GetKnownForPairs(pairs []repository.UserWordPair) (map[repository.UserWordPair]bool, error)
	UpsertBatch(entries []struct {
		UserID, WordCardID int64
		Score              int
	}) error
	GetScore(userID, wordCardID int64) (int, error)
}

// userCardRepoForTraining is used by TrainingService for generateQueue, GetDueCount, RestoreQueue (allows mocks in tests).
type userCardRepoForTraining interface {
	GetDueCards(userID int64, now time.Time, limit int) ([]*models.UserCard, error)
	GetNewCards(userID int64, limit int) ([]*models.UserCard, error)
	GetWordMasteringStats(userID, wordCardID int64) (*repository.WordMasteringStats, error)
	GetUserCard(userCardID int64) (*models.UserCard, error)
	GetDueCount(userID int64, now time.Time) (int, error)
}

// trainingCardRepoForQueue is used by TrainingService for generateQueue and RestoreQueue (allows mocks in tests).
type trainingCardRepoForQueue interface {
	GetTrainingCard(id int64) (*models.TrainingCard, error)
}

// TrainingService handles training session management
type TrainingService struct {
	userCardRepo          userCardRepoForTraining
	trainingCardRepo      trainingCardRepoForQueue
	sessionRepo           *repository.SessionRepository
	userWordMasteringRepo userWordMasteringRepoForSession // nil ok
	learning              learncfg.LearningConfig
	logger                *zap.Logger
}

// NewTrainingService creates a new training service (userWordMasteringRepo can be nil).
func NewTrainingService(
	userCardRepo userCardRepoForTraining,
	trainingCardRepo trainingCardRepoForQueue,
	sessionRepo *repository.SessionRepository,
	userWordMasteringRepo userWordMasteringRepoForSession,
	learning learncfg.LearningConfig,
	logger *zap.Logger,
) *TrainingService {
	return &TrainingService{
		userCardRepo:          userCardRepo,
		trainingCardRepo:      trainingCardRepo,
		sessionRepo:           sessionRepo,
		userWordMasteringRepo: userWordMasteringRepo,
		learning:              learning,
		logger:                logger,
	}
}

// SessionConfig holds session configuration
type SessionConfig struct {
	MaxCardsPerSession      int
	MaxNewPerSession        int
	AlgoVersion             string
	SpellEnabled            bool // inject spell (compose word) challenges
	SpellMasteringThreshold int  // min mastering_score 0-100 for words eligible for spell
	TypeEnabled             bool // inject type-the-word (no letters) challenges
	TypeMasteringThreshold  int  // min mastering_score 0-100 for words eligible for type
}

// StartSession starts a new training session. If config is nil, defaults are used (spell enabled, threshold 50).
func (s *TrainingService) StartSession(userID int64, source models.SessionSource, config *SessionConfig) (*models.TrainingSession, []*models.TrainingQueueItem, error) {
	// Check if there's already an active session
	activeSession, err := s.sessionRepo.GetActiveSession(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if activeSession != nil {
		// Finish the old session
		s.logger.Info("finishing old active session", zap.Int64("session_id", activeSession.ID))
		if err := s.sessionRepo.FinishSession(activeSession.ID, activeSession.DoneCount); err != nil {
			s.logger.Warn("failed to finish old session", zap.Error(err))
		}
	}

	if config == nil {
		config = &SessionConfig{
			MaxCardsPerSession:      models.DefaultMaxCardsPerSession,
			MaxNewPerSession:        models.DefaultMaxNewPerSession,
			AlgoVersion:             "srs_v2_delayed_mcq_sm2_autoquality",
			SpellEnabled:            true,
			SpellMasteringThreshold: 50,
			TypeEnabled:             true,
			TypeMasteringThreshold:  70,
		}
	}

	// Generate card queue (may include spell challenges if enabled)
	queue, err := s.generateQueue(userID, *config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate queue: %w", err)
	}

	if len(queue) == 0 {
		return nil, nil, fmt.Errorf("no cards available for training")
	}

	// Create session
	configJSON, _ := json.Marshal(config)
	session := &models.TrainingSession{
		UserID:       userID,
		Source:       source,
		PlannedCount: len(queue),
		DoneCount:    0,
		SessionJSON:  string(configJSON),
	}

	sessionID, err := s.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	session.ID = sessionID

	s.logger.Info("started training session",
		zap.Int64("session_id", sessionID),
		zap.Int64("user_id", userID),
		zap.String("source", string(source)),
		zap.String("learning_pair", s.learning.Pair),
		zap.Int("planned_count", len(queue)),
	)

	return session, queue, nil
}

// FinishSession finishes a training session and updates mastering scores for words touched in this session.
func (s *TrainingService) FinishSession(sessionID int64, doneCount int) error {
	if err := s.sessionRepo.FinishSession(sessionID, doneCount); err != nil {
		return err
	}
	if s.userWordMasteringRepo != nil {
		if err := s.updateMasteringScoresForSession(sessionID); err != nil {
			s.logger.Warn("failed to update mastering scores for session", zap.Int64("session_id", sessionID), zap.Error(err))
		}
	}
	return nil
}

// updateMasteringScoresForSession recalculates and upserts mastering_score for all (user_id, word_card_id) that had review_events in this session.
func (s *TrainingService) updateMasteringScoresForSession(sessionID int64) error {
	pairs, err := s.userWordMasteringRepo.GetWordCardIDsBySessionID(sessionID)
	if err != nil {
		return err
	}
	if len(pairs) == 0 {
		return nil
	}
	statsMap, err := s.userWordMasteringRepo.GetWordMasteringStatsBatch(pairs)
	if err != nil {
		return err
	}
	knownMap, err := s.userWordMasteringRepo.GetKnownForPairs(pairs)
	if err != nil {
		return err
	}
	entries := make([]struct {
		UserID     int64
		WordCardID int64
		Score      int
	}, 0, len(pairs))
	for _, p := range pairs {
		stats, hasStats := statsMap[p]
		known := knownMap[p]
		var total, correct, recentTotal, recentCorrect int64
		if hasStats {
			total = stats.Total
			correct = stats.Correct
			recentTotal = stats.RecentTotal
			recentCorrect = stats.RecentCorrect
		}
		score := ComputeMasteringScore(total, correct, recentTotal, recentCorrect, known)
		entries = append(entries, struct {
			UserID     int64
			WordCardID int64
			Score      int
		}{p.UserID, p.WordCardID, score})
	}
	return s.userWordMasteringRepo.UpsertBatch(entries)
}

// generateQueue generates a queue of cards for training (and optionally one spell challenge).
// From all cards available for training (due + new), we pick random MaxCardsPerSession; then
// within the session we shuffle and avoid adjacent duplicates (same word).
func (s *TrainingService) generateQueue(userID int64, config SessionConfig) ([]*models.TrainingQueueItem, error) {
	now := time.Now()

	// Build pool: all due (up to limit) + new (up to MaxNewPerSession)
	dueCards, err := s.userCardRepo.GetDueCards(userID, now, models.MaxDuePoolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get due cards: %w", err)
	}
	newCards, err := s.userCardRepo.GetNewCards(userID, config.MaxNewPerSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get new cards: %w", err)
	}

	s.logger.Info("built training pool for user",
		zap.Int64("user_id", userID),
		zap.Int("due_count", len(dueCards)),
		zap.Int("new_count", len(newCards)),
	)

	// Combine into pool and dedupe by UserCard.ID
	seenCardIDs := make(map[int64]bool)
	pool := make([]*models.UserCard, 0, len(dueCards)+len(newCards))
	for _, card := range dueCards {
		if !seenCardIDs[card.ID] {
			pool = append(pool, card)
			seenCardIDs[card.ID] = true
		}
	}
	for _, card := range newCards {
		if !seenCardIDs[card.ID] {
			pool = append(pool, card)
			seenCardIDs[card.ID] = true
		}
	}

	if len(pool) == 0 {
		return nil, nil
	}

	// Random sample of size MaxCardsPerSession from pool (no priority order — mix of familiar and new)
	allCards := pool
	if len(pool) > config.MaxCardsPerSession {
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		allCards = make([]*models.UserCard, config.MaxCardsPerSession)
		copy(allCards, pool[:config.MaxCardsPerSession])
	}

	// Fetch training card data
	cardQueue := make([]*models.UserCardWithTraining, 0, len(allCards))
	skippedCount := 0
	for _, userCard := range allCards {
		trainingCard, err := s.trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil {
			s.logger.Warn("failed to get training card",
				zap.Int64("user_card_id", userCard.ID),
				zap.Int64("training_card_id", userCard.TrainingCardID),
				zap.String("state", string(userCard.State)),
				zap.Error(err),
			)
			skippedCount++
			continue
		}
		if trainingCard == nil {
			s.logger.Warn("training card not found - this should not happen after INNER JOIN",
				zap.Int64("user_card_id", userCard.ID),
				zap.Int64("training_card_id", userCard.TrainingCardID),
				zap.String("state", string(userCard.State)),
			)
			skippedCount++
			continue
		}

		cardQueue = append(cardQueue, &models.UserCardWithTraining{
			UserCard:     *userCard,
			TrainingCard: *trainingCard,
		})
	}

	if skippedCount > 0 {
		s.logger.Warn("skipped cards during queue generation",
			zap.Int64("user_id", userID),
			zap.Int("total_cards", len(allCards)),
			zap.Int("skipped", skippedCount),
			zap.Int("queue_size", len(cardQueue)),
		)
	}

	// First, shuffle all cards randomly
	rand.Shuffle(len(cardQueue), func(i, j int) {
		cardQueue[i], cardQueue[j] = cardQueue[j], cardQueue[i]
	})

	// Then apply algorithm to prevent same words appearing close together
	cardQueue = s.shufflePreventDuplicates(cardQueue)

	if len(cardQueue) == 0 && len(allCards) > 0 {
		s.logger.Error("queue is empty but cards were fetched - all cards were skipped",
			zap.Int64("user_id", userID),
			zap.Int("total_cards_fetched", len(allCards)),
			zap.Int("skipped", skippedCount),
		)
		return nil, nil
	}

	// Build queue of TrainingQueueItem (all cards first)
	queue := make([]*models.TrainingQueueItem, 0, len(cardQueue))
	for _, c := range cardQueue {
		queue = append(queue, &models.TrainingQueueItem{Type: "card", Card: c})
	}

	s.applySpellTypeChallenges(queue, userID, config)
	return queue, nil
}

// applySpellTypeChallenges replaces some card items with spell or type challenge based on mastering score (mutates queue in place).
func (s *TrainingService) applySpellTypeChallenges(queue []*models.TrainingQueueItem, userID int64, config SessionConfig) {
	spellThresh := config.SpellMasteringThreshold
	if spellThresh < 0 {
		spellThresh = 0
	}
	if spellThresh > 100 {
		spellThresh = 100
	}
	typeThresh := config.TypeMasteringThreshold
	if typeThresh < 0 {
		typeThresh = 0
	}
	if typeThresh > 100 {
		typeThresh = 100
	}
	for i := range queue {
		if queue[i].Type != "card" || queue[i].Card == nil {
			continue
		}
		// Spell and type (compose/type English word) only for RU→EN; EN→RU stays as normal options
		if queue[i].Card.UserCard.Direction != models.DirectionRUtoEN {
			continue
		}
		tc := &queue[i].Card.TrainingCard
		wordCardID := tc.WordCardID
		displayWord := tc.WordEN
		if tc.DisplayWord != nil && *tc.DisplayWord != "" {
			displayWord = *tc.DisplayWord
		}
		if len(displayWord) < 2 {
			continue
		}
		var score int
		if s.userWordMasteringRepo != nil {
			score, _ = s.userWordMasteringRepo.GetScore(userID, wordCardID)
		} else {
			stats, err := s.userCardRepo.GetWordMasteringStats(userID, wordCardID)
			if err != nil || stats == nil {
				continue
			}
			score = computeMasteringScore(stats)
		}
		// Type threshold: 33% card, 33% spell, 33% type
		if config.TypeEnabled && score >= typeThresh {
			replacedID := queue[i].Card.UserCard.ID
			switch rand.Intn(3) {
			case 0:
				// keep card
			case 1:
				prefix, letters := spellPrefixAndLetters(displayWord)
				if len(letters) > 0 {
					queue[i] = &models.TrainingQueueItem{Type: "spell", Spell: &models.SpellChallenge{
						WordCardID: wordCardID, DisplayWord: displayWord, WordTarget: displayWord, WordRU: tc.WordRU, WordNative: tc.WordRU,
						Prefix: prefix, ShuffledLetters: letters,
						ReplacedUserCardID: replacedID,
					}}
				}
			case 2:
				queue[i] = &models.TrainingQueueItem{Type: "type", TypeChallenge: &models.TypeChallenge{
					WordCardID: wordCardID, DisplayWord: displayWord, WordTarget: displayWord, WordRU: tc.WordRU, WordNative: tc.WordRU,
					ReplacedUserCardID: replacedID,
				}}
			}
			continue
		}
		// Spell threshold: 50% card, 50% spell
		if config.SpellEnabled && score >= spellThresh {
			if rand.Intn(2) == 1 {
				prefix, letters := spellPrefixAndLetters(displayWord)
				if len(letters) > 0 {
					queue[i] = &models.TrainingQueueItem{Type: "spell", Spell: &models.SpellChallenge{
						WordCardID: wordCardID, DisplayWord: displayWord, WordTarget: displayWord, WordRU: tc.WordRU, WordNative: tc.WordRU,
						Prefix: prefix, ShuffledLetters: letters,
						ReplacedUserCardID: queue[i].Card.UserCard.ID,
					}}
				}
			}
		}
	}
}

// computeMasteringScore returns 0-100 using the same formula as vocab (mastering_score_calc)
func computeMasteringScore(stats *repository.WordMasteringStats) int {
	if stats.IsKnown {
		return 100
	}
	if stats.ReviewStateCount == stats.TotalCards && stats.TotalReps > 0 {
		cap20 := stats.TotalReps / 2
		if cap20 > 20 {
			cap20 = 20
		}
		return 75 + cap20
	}
	if stats.ReviewStateCount > 0 || stats.LearningStateCount > 0 {
		if stats.TotalCards == 0 {
			return 25
		}
		cap25 := (stats.ReviewStateCount + stats.LearningStateCount) * 25 / stats.TotalCards
		if cap25 > 25 {
			cap25 = 25
		}
		return 25 + cap25
	}
	return 0
}

// spellPrefixAndLetters returns prefix (e.g. "to " for verbs) and shuffled letters for the rest. For "to spy" -> ("to ", ["s","p","y"]).
func spellPrefixAndLetters(displayWord string) (prefix string, letters []string) {
	if strings.HasPrefix(displayWord, "to ") && len(displayWord) > 3 {
		prefix = "to "
		letters = shuffleLetters(displayWord[3:])
		return prefix, letters
	}
	return "", shuffleLetters(displayWord)
}

// shuffleLetters returns the word's runes as separate strings, shuffled (keeps spaces for "to spy" etc.)
func shuffleLetters(word string) []string {
	var runes []rune
	for _, r := range word {
		if unicode.IsLetter(r) || r == ' ' {
			runes = append(runes, r)
		}
	}
	if len(runes) < 2 {
		return nil
	}
	rand.Shuffle(len(runes), func(i, j int) { runes[i], runes[j] = runes[j], runes[i] })
	out := make([]string, len(runes))
	for i, r := range runes {
		out[i] = string(r)
	}
	return out
}

// shufflePreventDuplicates shuffles cards while preventing same word appearing close together
// Also prevents cards with the same word and direction from appearing close together
// If initial shuffle has duplicates, it will try to fix them with manual permutations or re-shuffling
func (s *TrainingService) shufflePreventDuplicates(queue []*models.UserCardWithTraining) []*models.UserCardWithTraining {
	if len(queue) <= 1 {
		return queue
	}

	// Group by word_card_id (same word, different senses/directions)
	wordGroups := make(map[int64][]*models.UserCardWithTraining)
	for _, card := range queue {
		wordGroups[card.TrainingCard.WordCardID] = append(wordGroups[card.TrainingCard.WordCardID], card)
	}

	// If no duplicates, already shuffled, return as is
	if len(wordGroups) == len(queue) {
		return queue
	}

	// Try shuffling with a reasonable limit on attempts
	maxAttempts := 10
	bestResult := queue
	bestScore := s.calculateShuffleScore(queue) // Worst possible score

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result := s.shufflePreventDuplicatesAttempt(queue, wordGroups)
		score := s.calculateShuffleScore(result)

		// If we found a perfect shuffle (no adjacent duplicates), use it immediately
		if score == 0 {
			return result
		}

		// Keep track of the best result so far
		if score < bestScore {
			bestResult = result
			bestScore = score
		}
	}

	// If we still have duplicates, try manual fixes
	if bestScore > 0 {
		bestResult = s.fixAdjacentDuplicates(bestResult)
	}

	return bestResult
}

// shufflePreventDuplicatesAttempt performs one attempt at shuffling
func (s *TrainingService) shufflePreventDuplicatesAttempt(
	queue []*models.UserCardWithTraining,
	wordGroups map[int64][]*models.UserCardWithTraining,
) []*models.UserCardWithTraining {
	// Build new queue spreading duplicates apart with larger minimum distance
	result := make([]*models.UserCardWithTraining, 0, len(queue))
	seenUserCardIDs := make(map[int64]bool) // Track UserCard.ID to prevent duplicates

	// Calculate minimum distance based on queue size
	// For larger queues, use larger distance to better spread words
	minDistance := 5
	if len(queue) < 10 {
		minDistance = 3
	} else if len(queue) < 20 {
		minDistance = 4
	}

	// Shuffle groups to randomize which word we try first
	groupKeys := make([]int64, 0, len(wordGroups))
	for k := range wordGroups {
		groupKeys = append(groupKeys, k)
	}
	rand.Shuffle(len(groupKeys), func(i, j int) {
		groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i]
	})

	// Create a copy of wordGroups for this attempt
	attemptWordGroups := make(map[int64][]*models.UserCardWithTraining)
	for k, v := range wordGroups {
		cards := make([]*models.UserCardWithTraining, len(v))
		copy(cards, v)
		attemptWordGroups[k] = cards
	}

	// Track total cards available to prevent infinite loops
	totalCardsAvailable := 0
	for _, cards := range attemptWordGroups {
		totalCardsAvailable += len(cards)
	}

	for len(result) < totalCardsAvailable {
		added := false

		// Check if we have any cards left
		hasCardsLeft := false
		for _, cards := range attemptWordGroups {
			if len(cards) > 0 {
				hasCardsLeft = true
				break
			}
		}
		if !hasCardsLeft {
			// No more cards available, stop
			break
		}

		// Try each word group in random order
		for _, wordCardID := range groupKeys {
			cards := attemptWordGroups[wordCardID]
			if len(cards) == 0 {
				continue
			}

			// Check if we can add this word (not used in last N positions)
			canAdd := true
			checkDistance := minDistance
			if len(result) < checkDistance {
				checkDistance = len(result)
			}

			for i := len(result) - checkDistance; i < len(result); i++ {
				// Prevent same word from appearing close together
				if result[i].TrainingCard.WordCardID == wordCardID {
					canAdd = false
					break
				}
			}

			if canAdd || len(result) == 0 {
				// Shuffle cards within this word group to randomize which card we take
				rand.Shuffle(len(cards), func(i, j int) {
					cards[i], cards[j] = cards[j], cards[i]
				})

				// Prefer cards with different direction than recent cards
				// Check last few cards to see their directions
				recentDirections := make(map[models.CardDirection]bool)
				checkDirDistance := 3
				if len(result) < checkDirDistance {
					checkDirDistance = len(result)
				}
				for i := len(result) - checkDirDistance; i < len(result); i++ {
					recentDirections[result[i].UserCard.Direction] = true
				}

				// Find first card that hasn't been added yet, preferring different direction
				cardAdded := false
				preferredCardIndex := -1
				for i, card := range cards {
					if !seenUserCardIDs[card.UserCard.ID] {
						// Prefer card with different direction than recent cards
						if !recentDirections[card.UserCard.Direction] {
							preferredCardIndex = i
							break
						}
						// If no preferred card found yet, remember this one as fallback
						if preferredCardIndex == -1 {
							preferredCardIndex = i
						}
					}
				}

				if preferredCardIndex >= 0 {
					card := cards[preferredCardIndex]
					result = append(result, card)
					seenUserCardIDs[card.UserCard.ID] = true
					// Remove this card from the group
					attemptWordGroups[wordCardID] = append(cards[:preferredCardIndex], cards[preferredCardIndex+1:]...)
					cardAdded = true
					added = true
				}

				if !cardAdded {
					// All cards in this group are already added, remove the group
					attemptWordGroups[wordCardID] = []*models.UserCardWithTraining{}
				}
			}
		}

		// If we couldn't add any card (all words are too recent), add the next available
		// This should rarely happen with proper distance
		if !added {
			for _, wordCardID := range groupKeys {
				cards := attemptWordGroups[wordCardID]
				if len(cards) > 0 {
					rand.Shuffle(len(cards), func(i, j int) {
						cards[i], cards[j] = cards[j], cards[i]
					})
					// Find first card that hasn't been added yet
					cardAdded := false
					for i, card := range cards {
						if !seenUserCardIDs[card.UserCard.ID] {
							result = append(result, card)
							seenUserCardIDs[card.UserCard.ID] = true
							// Remove this card from the group
							attemptWordGroups[wordCardID] = append(cards[:i], cards[i+1:]...)
							cardAdded = true
							break
						}
					}
					if cardAdded {
						break
					} else {
						// All cards in this group are already added, remove the group
						attemptWordGroups[wordCardID] = []*models.UserCardWithTraining{}
					}
				}
			}
		}

		// Clean up empty groups
		for i := len(groupKeys) - 1; i >= 0; i-- {
			if len(attemptWordGroups[groupKeys[i]]) == 0 {
				groupKeys = append(groupKeys[:i], groupKeys[i+1:]...)
			}
		}
	}

	return result
}

// calculateShuffleScore calculates a score for shuffle quality (lower is better)
// Returns the number of adjacent duplicates, with extra penalty for duplicates at the end
func (s *TrainingService) calculateShuffleScore(result []*models.UserCardWithTraining) int {
	if len(result) <= 1 {
		return 0
	}

	score := 0
	// Check all adjacent pairs
	for i := 1; i < len(result); i++ {
		if result[i].TrainingCard.WordCardID == result[i-1].TrainingCard.WordCardID {
			// Extra penalty for duplicates at the end (last 2 positions)
			if i >= len(result)-2 {
				score += 3 // Higher penalty for end duplicates
			} else {
				score += 1
			}
		}
	}

	return score
}

// fixAdjacentDuplicates tries to fix adjacent duplicates by swapping cards
func (s *TrainingService) fixAdjacentDuplicates(result []*models.UserCardWithTraining) []*models.UserCardWithTraining {
	if len(result) <= 1 {
		return result
	}

	// Create a copy to work with
	fixed := make([]*models.UserCardWithTraining, len(result))
	copy(fixed, result)

	// Try to fix duplicates, starting from the end (where they're most common)
	maxSwaps := len(fixed) * 2 // Reasonable limit
	swaps := 0

	for attempt := 0; attempt < maxSwaps && swaps < maxSwaps; attempt++ {
		improved := false

		// Check from the end backwards (prioritize fixing end duplicates)
		for i := len(fixed) - 1; i > 0; i-- {
			if fixed[i].TrainingCard.WordCardID == fixed[i-1].TrainingCard.WordCardID {
				// Try to find a card to swap with that fixes the duplicate without creating new ones
				// The duplicate is at positions i-1 and i (same WordCardID)
				// We want to swap fixed[i] with fixed[j] to fix it
				for j := 0; j < len(fixed); j++ {
					if j == i || j == i-1 {
						continue
					}

					// After swap: fixed[i] becomes fixed[j], fixed[j] becomes fixed[i]
					// Conditions for a valid swap:
					// 1. fixed[j].WordCardID != fixed[i-1].WordCardID (fixes the duplicate at i-1, i)
					// 2. fixed[j].WordCardID != fixed[j-1].WordCardID (if j > 0, no new duplicate when moved to position i)
					// 3. fixed[j].WordCardID != fixed[j+1].WordCardID (if j < len-1, no new duplicate when moved to position i)
					// 4. fixed[i].WordCardID != fixed[j-1].WordCardID (if j > 0, no new duplicate when moved to position j)
					// 5. fixed[i].WordCardID != fixed[j+1].WordCardID (if j < len-1, no new duplicate when moved to position j)

					// Must fix the original duplicate
					if fixed[j].TrainingCard.WordCardID == fixed[i-1].TrainingCard.WordCardID {
						continue
					}

					// Check if fixed[j] (which will be at position i after swap) would create duplicates
					if j > 0 && fixed[j-1].TrainingCard.WordCardID == fixed[j].TrainingCard.WordCardID {
						continue
					}
					if j < len(fixed)-1 && fixed[j+1].TrainingCard.WordCardID == fixed[j].TrainingCard.WordCardID {
						continue
					}

					// Check if fixed[i] (which will be at position j after swap) would create duplicates
					if j > 0 && fixed[j-1].TrainingCard.WordCardID == fixed[i].TrainingCard.WordCardID {
						continue
					}
					if j < len(fixed)-1 && fixed[j+1].TrainingCard.WordCardID == fixed[i].TrainingCard.WordCardID {
						continue
					}

					// All checks passed, swap cards
					fixed[i], fixed[j] = fixed[j], fixed[i]
					improved = true
					swaps++
					break
				}

				if improved {
					break
				}
			}
		}

		// If no improvement was made, break
		if !improved {
			break
		}
	}

	return fixed
}

// GetDueCount gets the count of due cards for a user
func (s *TrainingService) GetDueCount(userID int64) (int, error) {
	return s.userCardRepo.GetDueCount(userID, time.Now())
}

// GetSession gets a session by ID
func (s *TrainingService) GetSession(sessionID int64) (*models.TrainingSession, error) {
	return s.sessionRepo.GetSession(sessionID)
}

// GetActiveSession gets the active session for a user
func (s *TrainingService) GetActiveSession(userID int64) (*models.TrainingSession, error) {
	return s.sessionRepo.GetActiveSession(userID)
}

// UpdateSessionState updates session state in database
func (s *TrainingService) UpdateSessionState(sessionID int64, sessionJSON string) error {
	session, err := s.sessionRepo.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	session.SessionJSON = sessionJSON
	return s.sessionRepo.UpdateSession(session)
}

// RestoreQueue restores queue from user card IDs
func (s *TrainingService) RestoreQueue(userID int64, userCardIDs []int64) ([]*models.UserCardWithTraining, error) {
	if len(userCardIDs) == 0 {
		return nil, nil
	}

	queue := make([]*models.UserCardWithTraining, 0, len(userCardIDs))

	for _, userCardID := range userCardIDs {
		// Get user card
		userCard, err := s.userCardRepo.GetUserCard(userCardID)
		if err != nil {
			s.logger.Warn("failed to get user card during restore",
				zap.Int64("user_card_id", userCardID),
				zap.Error(err),
			)
			continue
		}
		if userCard == nil {
			s.logger.Warn("user card not found during restore",
				zap.Int64("user_card_id", userCardID),
			)
			continue
		}

		// Verify user owns this card
		if userCard.UserID != userID {
			s.logger.Warn("user card belongs to different user",
				zap.Int64("user_card_id", userCardID),
				zap.Int64("expected_user_id", userID),
				zap.Int64("actual_user_id", userCard.UserID),
			)
			continue
		}

		// Get training card
		trainingCard, err := s.trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil {
			s.logger.Warn("failed to get training card during restore",
				zap.Int64("training_card_id", userCard.TrainingCardID),
				zap.Error(err),
			)
			continue
		}
		if trainingCard == nil {
			s.logger.Warn("training card not found during restore",
				zap.Int64("training_card_id", userCard.TrainingCardID),
			)
			continue
		}

		queue = append(queue, &models.UserCardWithTraining{
			UserCard:     *userCard,
			TrainingCard: *trainingCard,
		})
	}

	return queue, nil
}
