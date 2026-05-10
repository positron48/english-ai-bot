package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type ReadingTextProgress struct {
	UserID    int64
	ChapterID string
	ReadAt    time.Time
}

type ReadingTextProgressRepository struct {
	db *sql.DB
}

func NewReadingTextProgressRepository(db *sql.DB) *ReadingTextProgressRepository {
	return &ReadingTextProgressRepository{db: db}
}

func (r *ReadingTextProgressRepository) MarkRead(userID int64, chapterID string) error {
	const q = `
INSERT INTO reading_text_progress (user_id, chapter_id, read_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, chapter_id) DO UPDATE SET
	read_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, userID, chapterID); err != nil {
		return fmt.Errorf("mark reading text read: %w", err)
	}
	return nil
}

func (r *ReadingTextProgressRepository) Get(userID int64, chapterID string) (*ReadingTextProgress, error) {
	const q = `
SELECT user_id, chapter_id, read_at
FROM reading_text_progress
WHERE user_id = ? AND chapter_id = ?`
	var p ReadingTextProgress
	if err := r.db.QueryRow(q, userID, chapterID).Scan(&p.UserID, &p.ChapterID, &p.ReadAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get reading text progress: %w", err)
	}
	return &p, nil
}
