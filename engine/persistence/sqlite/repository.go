package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"

	_ "modernc.org/sqlite"
)

type sqliteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new repository and ensures schema is initialized.
func NewSQLiteRepository(dbPath string) (ports.Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	repo := &sqliteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *sqliteRepository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS analysis_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		platform TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		violations_json TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_reports_timestamp ON analysis_reports(timestamp);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *sqliteRepository) SaveReport(ctx context.Context, report domain.ViolationReport) error {
	violationsJSON, err := json.Marshal(report.Violations)
	if err != nil {
		return fmt.Errorf("failed to marshal violations: %w", err)
	}

	if report.Timestamp == 0 {
		report.Timestamp = time.Now().Unix()
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO analysis_reports (file_path, platform, timestamp, violations_json) VALUES (?, ?, ?, ?)",
		report.FilePath, report.Platform, report.Timestamp, string(violationsJSON),
	)
	return err
}

func (r *sqliteRepository) GetHistory(ctx context.Context, limit int) ([]domain.ViolationReport, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, file_path, platform, timestamp, violations_json FROM analysis_reports ORDER BY timestamp DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var reports []domain.ViolationReport
	for rows.Next() {
		var rep domain.ViolationReport
		var violationsJSON string
		if err := rows.Scan(&rep.ID, &rep.FilePath, &rep.Platform, &rep.Timestamp, &violationsJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(violationsJSON), &rep.Violations); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}

	return reports, nil
}
