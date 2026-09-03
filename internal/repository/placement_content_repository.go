package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"tgbot-skeleton/internal/placement"
	"tgbot-skeleton/internal/placementbundle"
)

type PlacementContentRepository struct {
	db     *sql.DB
	source string
}

func NewPlacementContentRepository(db *sql.DB, source string) *PlacementContentRepository {
	return &PlacementContentRepository{db, source}
}
func (r *PlacementContentRepository) Load(ctx context.Context, course string) (*placement.Bank, error) {
	if r.source != "db" {
		return placementbundle.Load(course)
	}
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `SELECT data_json FROM placement_banks WHERE course_code = ? AND active`, course).Scan(&raw); err != nil {
		return nil, fmt.Errorf("active placement bank: %w", err)
	}
	var b placement.Bank
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	if b.CourseCode != course {
		return nil, fmt.Errorf("placement bank course mismatch")
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return &b, nil
}

// ImportPlacementBank retains old versions; sessions independently keep snapshots.
func ImportPlacementBank(ctx context.Context, db *sql.DB, b *placement.Bank) error {
	if err := b.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(b)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize imports for this course without deactivating another import midway.
	var course string
	if err = tx.QueryRowContext(ctx, `SELECT code FROM courses WHERE code = ? FOR UPDATE`, b.CourseCode).Scan(&course); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE placement_banks SET active=false WHERE course_code=? AND active`, b.CourseCode); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO placement_banks(course_code,version,data_json,active) VALUES(?,?,?::jsonb,true)
		ON CONFLICT(course_code,version) DO UPDATE SET active=true`, b.CourseCode, b.Version, string(raw)); err != nil {
		return err
	}
	return tx.Commit()
}
