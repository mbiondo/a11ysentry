package domain

import (
	"context"
)

// AnalysisContext provides shared metadata and project configuration to rules.
type AnalysisContext struct {
	Nodes            []USN
	Config           ProjectConfig
	LabelsByFor      map[string]string
	LandmarkLabels   map[SemanticRole]map[string]Source
	MainCount        int
	LinksByLabel     map[string]map[string][]int
	HasLang          bool
	HasH1            bool
	LastHeadingLevel int
}

// Rule defines the interface for an accessibility audit rule.
type Rule interface {
	Name() string
	ErrorCode() string
	ACTID() string
	DocumentationURL() string
	Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error)
}
