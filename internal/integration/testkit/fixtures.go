package testkit

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// UserFixture creates or returns a user by telegram ID
func UserFixture(t *testing.T, conn *sql.DB, telegramID int64) *models.User {
	t.Helper()
	logger := zap.NewNop()
	repo := repository.NewUserRepository(conn, logger)
	user, err := repo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("UserFixture: %v", err)
	}
	return user
}

// WordFixture creates a word_card and training_cards; returns word_card_id and training_card IDs
func WordFixture(t *testing.T, conn *sql.DB, word, definition, wordRu string) (wordCardID int64, trainingCardIDs []int64) {
	t.Helper()
	logger := zap.NewNop()
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingRepo := repository.NewTrainingCardRepository(conn, logger)

	wid, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:       word,
		Definition: definition,
	})
	if err != nil {
		t.Fatalf("WordFixture word card: %v", err)
	}

	tcID, err := trainingRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wid,
		WordEN:     word,
		SenseIndex: 0,
		WordRU:     wordRu,
		MeaningEN:  definition,
	})
	if err != nil {
		t.Fatalf("WordFixture training card: %v", err)
	}
	return wid, []int64{tcID}
}

// TrainingDeckFixture creates N words with training cards and user_cards for user
func TrainingDeckFixture(t *testing.T, conn *sql.DB, userID int64, count int) (wordCardIDs, trainingCardIDs []int64) {
	t.Helper()
	logger := zap.NewNop()
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)

	for i := 0; i < count; i++ {
		word := fmt.Sprintf("fixture-word-%d", i+1)
		wid, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: word, Definition: "definition"})
		if err != nil {
			t.Fatalf("TrainingDeckFixture word: %v", err)
		}
		wordCardIDs = append(wordCardIDs, wid)

		tcID, err := trainingRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wid,
			WordEN:     word,
			SenseIndex: 0,
			WordRU:     "перевод",
			MeaningEN:  "definition",
		})
		if err != nil {
			t.Fatalf("TrainingDeckFixture training card: %v", err)
		}
		trainingCardIDs = append(trainingCardIDs, tcID)

		for _, dir := range []models.CardDirection{models.DirectionENtoRU, models.DirectionRUtoEN} {
			_, err = userCardRepo.CreateUserCard(&models.UserCard{
				UserID:         userID,
				TrainingCardID: tcID,
				Direction:      dir,
				State:          models.StateNew,
				EF:             2.5,
			})
			if err != nil {
				t.Fatalf("TrainingDeckFixture user card: %v", err)
			}
		}
	}
	return wordCardIDs, trainingCardIDs
}

// GrammarPublishFixture publishes a grammar section and its chapters
func GrammarPublishFixture(t *testing.T, conn *sql.DB, sectionID string, chapterIDs []string) {
	t.Helper()
	for _, cid := range chapterIDs {
		_, err := conn.Exec(
			`INSERT INTO grammar_published_items (item_type, item_id, is_published, name, updated_at)
			 VALUES ($1, $2, 1, $2, NOW())
			 ON CONFLICT (item_type, item_id) DO UPDATE SET is_published=1, name=EXCLUDED.name, updated_at=NOW()`,
			"chapter", cid,
		)
		if err != nil {
			t.Fatalf("GrammarPublishFixture chapter %s: %v", cid, err)
		}
	}
	_, err := conn.Exec(
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, name, updated_at)
		 VALUES ($1, $2, 1, $2, NOW())
		 ON CONFLICT (item_type, item_id) DO UPDATE SET is_published=1, name=EXCLUDED.name, updated_at=NOW()`,
		"section", sectionID,
	)
	if err != nil {
		t.Fatalf("GrammarPublishFixture section %s: %v", sectionID, err)
	}
}

// GrammarProgressFixture creates grammar_progress for user+chapter
func GrammarProgressFixture(t *testing.T, conn *sql.DB, userID int64, chapterID string, bestScore int, passed bool) {
	t.Helper()
	var passedAt *time.Time
	if passed {
		now := time.Now().UTC()
		passedAt = &now
	}
	_, err := conn.Exec(
		`INSERT INTO grammar_progress (user_id, chapter_id, best_score, passed_at, last_attempt_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (user_id, chapter_id) DO UPDATE SET best_score=GREATEST(grammar_progress.best_score, EXCLUDED.best_score), passed_at=COALESCE(EXCLUDED.passed_at, grammar_progress.passed_at), last_attempt_at=EXCLUDED.last_attempt_at`,
		userID, chapterID, bestScore, passedAt,
	)
	if err != nil {
		t.Fatalf("GrammarProgressFixture: %v", err)
	}
}
