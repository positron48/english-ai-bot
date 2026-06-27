package repository

import (
	"database/sql"
	"fmt"
	"strings"
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

// CountReadInSet returns how many of the given chapter IDs the user has read.
func (r *ReadingTextProgressRepository) CountReadInSet(userID int64, chapterIDs []string) (int, error) {
	if len(chapterIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(chapterIDs))
	args := make([]interface{}, 0, len(chapterIDs)+1)
	args = append(args, userID)
	for i, id := range chapterIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := fmt.Sprintf(`SELECT COUNT(*) FROM reading_text_progress WHERE user_id = ? AND chapter_id IN (%s)`,
		strings.Join(placeholders, ","))
	var count int
	if err := r.db.QueryRow(q, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count read in set: %w", err)
	}
	return count, nil
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
