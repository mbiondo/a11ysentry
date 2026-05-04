package domain

import (
	"context"
	"strings"
)

type ruleBestPractices struct{}

func (r *ruleBestPractices) Name() string             { return "A11y Best Practices" }
func (r *ruleBestPractices) ErrorCode() string        { return "BEST_PRACTICES" }
func (r *ruleBestPractices) ACTID() string            { return "N/A" }
func (r *ruleBestPractices) DocumentationURL() string { return "https://www.w3.org/WAI/standards-guidelines/wcag/" }

func (r *ruleBestPractices) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		// Redundant Title
		if title, ok := node.Traits["title"].(string); ok && title != "" {
			if strings.TrimSpace(title) == strings.TrimSpace(node.Label) {
				violations = append(violations, Violation{
					ErrorCode:        "REDUNDANT_TITLE",
					Severity:         SeverityWarning,
					Message:          "The 'title' attribute is identical to the element's text/label. This creates redundant announcements for screen reader users.",
					SourceRef:        node.Source,
					FixSnippet:       "Remove the title attribute as it provides no additional information.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H65",
				})
			}
		}

		// Modal/Dialog aria-modal
		if node.Role == RoleModal {
			hasAriaModal := node.Traits["aria-modal"] == "true"
			if !hasAriaModal {
				violations = append(violations, Violation{
					ErrorCode:        "ADV_FOCUS_TRAP",
					Severity:         SeverityError,
					Message:          "Modal/Dialog missing 'aria-modal=\"true\"'. Without it, screen readers may allow users to navigate outside the modal while it is active.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-modal=\"true\" to the dialog element.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}
		}
	}

	return violations, nil
}
