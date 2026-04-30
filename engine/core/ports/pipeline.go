package ports

import (
	"context"
	"a11ysentry/engine/core/domain"
)

// Adapter is responsible for ingesting source code and normalizing it to USN.
type Adapter interface {
	Ingest(ctx context.Context, files []string) ([]domain.USN, error)
}

// Analyzer applies accessibility rules to a USN tree.
type Analyzer interface {
	Analyze(ctx context.Context, nodes []domain.USN) ([]domain.Violation, error)
}

// Emitter outputs the results of the analysis.
type Emitter interface {
	Emit(ctx context.Context, violations []domain.Violation) error
}

