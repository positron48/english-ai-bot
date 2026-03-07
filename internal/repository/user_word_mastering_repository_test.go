package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupUserWordMasteringRepo(t *testing.T) (*UserWordMasteringRepository, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	return NewUserWordMasteringRepository(db, zap.NewNop()), db
}

// seedUserWordMasteringFixtures creates user, word_cards, training_cards, user_cards, session, review_events for stats tests.
func seedUserWordMasteringFixtures(t *testing.T, db *sql.DB) (userID, wordCardID, trainingCardID, userCardID, sessionID int64) {
	t.Helper()
	var uid int64
	err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (90001) RETURNING id`).Scan(&uid)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var wcid int64
	err = db.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('testword', 'def') RETURNING id`).Scan(&wcid)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var tcid int64
	err = db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES ($1, 'testword', 0, 'тест', 'meaning') RETURNING id`, wcid).Scan(&tcid)
	if err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	var ucid int64
	err = db.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'en_ru', 'review') RETURNING id`, uid, tcid).Scan(&ucid)
	if err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	var sid int64
	err = db.QueryRow(`INSERT INTO training_sessions (user_id, source) VALUES ($1, 'test') RETURNING id`, uid).Scan(&sid)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	_, err = db.Exec(`INSERT INTO review_events (session_id, user_id, user_card_id, direction, answered_at, is_correct) VALUES ($1, $2, $3, 'en_ru', CURRENT_TIMESTAMP, 1)`, sid, uid, ucid)
	if err != nil {
		t.Fatalf("insert review_event: %v", err)
	}
	return uid, wcid, tcid, ucid, sid
}

func TestNewUserWordMasteringRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	if repo == nil {
		t.Fatal("NewUserWordMasteringRepository() should not return nil")
	}
}

func TestUserWordMasteringRepository_GetWordMasteringStatsBatch(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)

	t.Run("empty pairs", func(t *testing.T) {
		out, err := repo.GetWordMasteringStatsBatch(nil)
		if err != nil {
			t.Fatalf("GetWordMasteringStatsBatch(nil) error = %v", err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %d entries", len(out))
		}
	})

	t.Run("pair with review_events", func(t *testing.T) {
		pairs := []UserWordPair{{UserID: userID, WordCardID: wordCardID}}
		out, err := repo.GetWordMasteringStatsBatch(pairs)
		if err != nil {
			t.Fatalf("GetWordMasteringStatsBatch() error = %v", err)
		}
		row, ok := out[UserWordPair{UserID: userID, WordCardID: wordCardID}]
		if !ok {
			t.Fatal("expected one row for seeded pair")
		}
		if row.Total < 1 || row.Correct < 1 {
			t.Errorf("expected at least 1 total/correct, got total=%d correct=%d", row.Total, row.Correct)
		}
	})

	t.Run("pair with no events returns empty", func(t *testing.T) {
		pairs := []UserWordPair{{UserID: 99999, WordCardID: 99999}}
		out, err := repo.GetWordMasteringStatsBatch(pairs)
		if err != nil {
			t.Fatalf("GetWordMasteringStatsBatch() error = %v", err)
		}
		if len(out) != 0 {
			t.Errorf("expected no rows for non-existent pair, got %d", len(out))
		}
	})
}

func TestUserWordMasteringRepository_GetWordCardIDsBySessionID(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	_, wordCardID, _, _, sessionID := seedUserWordMasteringFixtures(t, db)
	userID := int64(0)
	_ = db.QueryRow(`SELECT user_id FROM training_sessions WHERE id = $1`, sessionID).Scan(&userID)

	pairs, err := repo.GetWordCardIDsBySessionID(sessionID)
	if err != nil {
		t.Fatalf("GetWordCardIDsBySessionID() error = %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].UserID != userID || pairs[0].WordCardID != wordCardID {
		t.Errorf("expected pair (user=%d, word=%d), got (%d, %d)", userID, wordCardID, pairs[0].UserID, pairs[0].WordCardID)
	}

	pairsEmpty, _ := repo.GetWordCardIDsBySessionID(99999)
	if len(pairsEmpty) != 0 {
		t.Errorf("expected 0 pairs for unknown session, got %d", len(pairsEmpty))
	}
}

func TestUserWordMasteringRepository_GetKnownWordCardIDsForUser(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)
	_, _ = db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known') ON CONFLICT (user_id, word_card_id) DO UPDATE SET status = 'known'`, userID, wordCardID)

	t.Run("empty list", func(t *testing.T) {
		out, err := repo.GetKnownWordCardIDsForUser(userID, nil)
		if err != nil {
			t.Fatalf("GetKnownWordCardIDsForUser(nil) error = %v", err)
		}
		if len(out) != 0 {
			t.Error("expected empty map")
		}
	})

	t.Run("known and unknown", func(t *testing.T) {
		ids := []int64{wordCardID, 99998}
		out, err := repo.GetKnownWordCardIDsForUser(userID, ids)
		if err != nil {
			t.Fatalf("GetKnownWordCardIDsForUser() error = %v", err)
		}
		if !out[wordCardID] {
			t.Error("wordCardID should be known")
		}
		if out[99998] {
			t.Error("99998 should not be known")
		}
	})
}

func TestUserWordMasteringRepository_GetKnownForPairs(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)
	_, _ = db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known') ON CONFLICT (user_id, word_card_id) DO UPDATE SET status = 'known'`, userID, wordCardID)

	t.Run("empty pairs", func(t *testing.T) {
		out, err := repo.GetKnownForPairs(nil)
		if err != nil {
			t.Fatalf("GetKnownForPairs(nil) error = %v", err)
		}
		if len(out) != 0 {
			t.Error("expected empty map")
		}
	})

	t.Run("one known one unknown", func(t *testing.T) {
		pairs := []UserWordPair{{userID, wordCardID}, {userID, 99997}}
		out, err := repo.GetKnownForPairs(pairs)
		if err != nil {
			t.Fatalf("GetKnownForPairs() error = %v", err)
		}
		if !out[UserWordPair{userID, wordCardID}] {
			t.Error("(userID, wordCardID) should be known")
		}
		if out[UserWordPair{userID, 99997}] {
			t.Error("(userID, 99997) should not be known")
		}
	})
}

func TestUserWordMasteringRepository_Upsert(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)

	if err := repo.Upsert(userID, wordCardID, 75); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	score, err := repo.GetScore(userID, wordCardID)
	if err != nil {
		t.Fatalf("GetScore() error = %v", err)
	}
	if score != 75 {
		t.Errorf("expected score 75, got %d", score)
	}

	if err := repo.Upsert(userID, wordCardID, 90); err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}
	score, _ = repo.GetScore(userID, wordCardID)
	if score != 90 {
		t.Errorf("expected score 90 after update, got %d", score)
	}
}

func TestUserWordMasteringRepository_UpsertBatch(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)

	t.Run("empty entries no-op", func(t *testing.T) {
		if err := repo.UpsertBatch(nil); err != nil {
			t.Fatalf("UpsertBatch(nil) error = %v", err)
		}
	})

	entries := []struct {
		UserID     int64
		WordCardID int64
		Score      int
	}{
		{userID, wordCardID, 50},
		{userID, wordCardID + 1, 60},
	}
	// Ensure word_card exists for second entry
	_, _ = db.Exec(`INSERT INTO word_cards (word, definition) VALUES ('word2', 'def2') ON CONFLICT DO NOTHING`)
	var wc2 int64
	_ = db.QueryRow(`SELECT id FROM word_cards WHERE word = 'word2'`).Scan(&wc2)
	if wc2 != 0 {
		entries[1].WordCardID = wc2
	} else {
		entries = entries[:1]
	}

	if err := repo.UpsertBatch(entries); err != nil {
		t.Fatalf("UpsertBatch() error = %v", err)
	}
	score, err := repo.GetScore(userID, wordCardID)
	if err != nil || score != 50 {
		t.Errorf("expected score 50, got %d err=%v", score, err)
	}
}

func TestUserWordMasteringRepository_GetScore(t *testing.T) {
	repo, db := setupUserWordMasteringRepo(t)
	userID, wordCardID, _, _, _ := seedUserWordMasteringFixtures(t, db)

	score, err := repo.GetScore(userID, wordCardID)
	if err != nil {
		t.Fatalf("GetScore() error = %v", err)
	}
	if score != 0 {
		t.Errorf("expected 0 for missing row, got %d", score)
	}

	_, _ = db.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES ($1, $2, 42) ON CONFLICT (user_id, word_card_id) DO UPDATE SET mastering_score = 42`, userID, wordCardID)
	score, err = repo.GetScore(userID, wordCardID)
	if err != nil {
		t.Fatalf("GetScore() error = %v", err)
	}
	if score != 42 {
		t.Errorf("expected 42, got %d", score)
	}
}
