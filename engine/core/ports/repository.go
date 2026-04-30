package ports

import (
	"context"
	"a11ysentry/engine/core/domain"
)

// Repository defines the port for persisting analysis history.
type Repository interface {
	// SaveReport stores a new violation report in the database.
	SaveReport(ctx context.Context, report domain.ViolationReport) error
	// GetHistory retrieves the last N violation reports.
	GetHistory(ctx context.Context, limit int) ([]domain.ViolationReport, error)
}
