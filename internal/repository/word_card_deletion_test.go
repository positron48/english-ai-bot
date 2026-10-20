package repository

import (
	"testing"
	"tgbot-skeleton/internal/models"
)

func TestWordCardReportDetachQueriesExplain(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	if err := repo.SaveWordCard("delete-explain", "definition", ""); err != nil {
		t.Fatal(err)
	}
	word, err := repo.GetWordCard("delete-explain")
	if err != nil {
		t.Fatal(err)
	}
	var userID int64
	if err := db.QueryRow(`INSERT INTO users(telegram_id) VALUES(987123) RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	tc, err := NewTrainingCardRepository(db, repo.logger).CreateTrainingCard(&models.TrainingCard{WordCardID: word.ID, WordEN: word.Word, WordRU: "пример", MeaningEN: "example"})
	if err != nil {
		t.Fatal(err)
	}
	uc, err := NewUserCardRepository(db, repo.logger).CreateUserCard(&models.UserCard{UserID: userID, TrainingCardID: tc, Direction: models.DirectionRUtoEN, State: models.StateReview})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewContentReportRepository(db, repo.logger).Create(CreateContentReportInput{UserID: userID, SourceType: "word_training", WordCardID: &word.ID, TrainingCardID: &tc, UserCardID: &uc, ReportCategory: "other"}); err != nil {
		t.Fatal(err)
	}
	for _, query := range wordCardReportDetachQueries {
		rows, err := db.Query("EXPLAIN "+query, word.ID)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var plan string
			if err := rows.Scan(&plan); err != nil {
				t.Fatal(err)
			}
			t.Log(plan)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
}
