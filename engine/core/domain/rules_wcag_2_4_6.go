package domain

import (
	"context"
)

type ruleWCAG246 struct{}

func (r *ruleWCAG246) Name() string             { return "Headings and Labels" }
func (r *ruleWCAG246) ErrorCode() string        { return "WCAG_2_4_6" }
func (r *ruleWCAG246) ACTID() string            { return "9eb3f6" } // Example ACT ID
func (r *ruleWCAG246) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G130" }

func (r *ruleWCAG246) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleHeading && node.Label == "" {
			violations = append(violations, Violation{
				ErrorCode:        r.ErrorCode(),
				Severity:         SeverityWarning,
				Message:          "Heading element has no visible text content. Headings must be descriptive to aid navigation.",
				SourceRef:        node.Source,
				FixSnippet:       "Add meaningful text to the heading or remove it if it is decorative.",
				DocumentationURL: r.DocumentationURL(),
			})
		}
	}

	return violations, nil
}
