package domain

import (
	"context"
	"fmt"
)

type ruleWCAG131 struct{}

func (r *ruleWCAG131) Name() string             { return "Info and Relationships" }
func (r *ruleWCAG131) ErrorCode() string        { return "WCAG_1_3_1" }
func (r *ruleWCAG131) ACTID() string            { return "bf051a" } // Language/Structure
func (r *ruleWCAG131) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G141" }

func (r *ruleWCAG131) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation
	lastHeadingLevel := 0

	for i, node := range analysisCtx.Nodes {
		// Heading hierarchy
		if node.Role == RoleHeading {
			level := 0
			_, _ = fmt.Sscanf(string(node.UID), "h%d", &level)
			if level > lastHeadingLevel+1 && lastHeadingLevel != 0 {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Heading levels should only increase by one. Jumped from H%d to H%d.", lastHeadingLevel, level),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Adjust level to H%d to maintain hierarchy.", lastHeadingLevel+1),
					DocumentationURL: r.DocumentationURL(),
				})
			}
			lastHeadingLevel = level
		}

		// Fieldset Legend
		if node.Role == RoleFieldset {
			hasLegend := false
			for _, child := range analysisCtx.Nodes[i+1:] {
				if child.Role == RoleLegend {
					hasLegend = true
					break
				}
				if child.Role == RoleFieldset || child.Role == RoleMain || child.Role == RoleHeader {
					break
				}
			}
			if !hasLegend {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_1_3_1_LEGEND",
					Severity:         SeverityError,
					Message:          "<fieldset> is missing a <legend>. Grouped inputs need a descriptive legend for context.",
					SourceRef:        node.Source,
					FixSnippet:       "<legend>Group Description</legend>",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H71",
				})
			}
		}
	}

	return violations, nil
}
