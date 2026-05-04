package domain

import (
	"context"
	"fmt"
)

type ruleParsing struct{}

func (r *ruleParsing) Name() string             { return "Parsing" }
func (r *ruleParsing) ErrorCode() string        { return "WCAG_4_1_1" }
func (r *ruleParsing) ACTID() string            { return "3ea0c8" }
func (r *ruleParsing) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/html/H93" }

func (r *ruleParsing) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation
	usedIDs := make(map[string]bool)

	for _, node := range analysisCtx.Nodes {
		if id, ok := node.Traits["id"].(string); ok && id != "" {
			if usedIDs[id] {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Duplicate ID found: '%s'. IDs must be unique for focus management.", id),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("id=\"%s-unique\"", id),
					DocumentationURL: r.DocumentationURL(),
				})
			}
			usedIDs[id] = true
		}
	}

	return violations, nil
}
