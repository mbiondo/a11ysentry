package domain

import (
	"context"
	"fmt"
)

type ruleLandmarks struct{}

func (r *ruleLandmarks) Name() string             { return "Landmarks" }
func (r *ruleLandmarks) ErrorCode() string        { return "WCAG_2_4_1" }
func (r *ruleLandmarks) ACTID() string            { return "bc659a" } // Example ACT ID
func (r *ruleLandmarks) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/html/H64" }

func (r *ruleLandmarks) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation
	mainFound := 0

	for _, node := range analysisCtx.Nodes {
		if !isLandmark(node.Role) {
			continue
		}

		if node.Role == RoleMain {
			mainFound++
			if mainFound > 1 {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityError,
					Message:          "Multiple <main> elements found. A document should only have one primary content landmark.",
					SourceRef:        node.Source,
					FixSnippet:       "Remove redundant <main> elements or convert them to <section>.",
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}

		// Check if multiple landmarks of same type exist without labels
		sameRoleCount := 0
		for _, n2 := range analysisCtx.Nodes {
			if n2.Role == node.Role {
				sameRoleCount++
			}
		}
		if sameRoleCount > 1 && node.Label == "" && node.Role != RoleMain {
			violations = append(violations, Violation{
				ErrorCode:        "ARIA_1_1",
				Severity:         SeverityWarning,
				Message:          fmt.Sprintf("Multiple %s landmarks found without distinguishing labels. Screen reader users won't know the difference between them.", node.Role),
				SourceRef:        node.Source,
				FixSnippet:       fmt.Sprintf("Add aria-label=\"...\" to distinguish this %s from others.", node.Role),
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA6",
			})
		}
	}

	return violations, nil
}
