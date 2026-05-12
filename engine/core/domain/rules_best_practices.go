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

		// Modal/Dialog aria-modal and Focus Trap
		if node.Role == RoleModal {
			ariaModal, _ := node.Traits["aria-modal"].(string)
			hasAriaModal := ariaModal == "true"
			
			if !hasAriaModal {
				violations = append(violations, Violation{
					ErrorCode:        "ADV_FOCUS_TRAP",
					Severity:         SeverityError,
					Message:          "Modal/Dialog missing 'aria-modal=\"true\"'. This is required to tell assistive technologies that the rest of the page is inactive.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-modal=\"true\" to the element.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}

			// Best practice: modals should have an accessible label
			if node.Label == "" {
				violations = append(violations, Violation{
					ErrorCode:        "MODAL_MISSING_LABEL",
					Severity:         SeverityWarning,
					Message:          "Modal/Dialog missing an accessible label (aria-label or aria-labelledby). Users won't know the purpose of the modal when it opens.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-label=\"Modal Title\" or aria-labelledby=\"id-of-title\".",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}
		}
	}

	return violations, nil
}
