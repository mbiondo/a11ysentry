package domain

import (
	"context"
	"fmt"
	"strings"
)

type ruleFocus struct{}

func (r *ruleFocus) Name() string             { return "Focus and Keyboard Navigation" }
func (r *ruleFocus) ErrorCode() string        { return "WCAG_2_4_7" }
func (r *ruleFocus) ACTID() string            { return "oj04fd" }
func (r *ruleFocus) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G149" }

func (r *ruleFocus) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		isNativeInteractive := node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput
		tabindexAttr, _ := node.Traits["tabindex"].(string)
		hasTabIndex := tabindexAttr != ""

		// 1. Focus Visibility (WCAG 2.4.7)
		if isNativeInteractive || (hasTabIndex && isPositiveOrZero(tabindexAttr)) {
			if style, ok := node.Traits["style"].(string); ok {
				styleLower := strings.ReplaceAll(strings.ToLower(style), " ", "")
				
				// Detect if outline is disabled
				isOutlineDisabled := strings.Contains(styleLower, "outline:none") ||
					strings.Contains(styleLower, "outline:0") ||
					strings.Contains(styleLower, "outline:0px")

				if isOutlineDisabled {
					// Check for alternative focus indicators like box-shadow or border-color
					hasAlternative := strings.Contains(styleLower, "box-shadow") || 
									 strings.Contains(styleLower, "border-color") ||
									 strings.Contains(styleLower, "background-color")

					if !hasAlternative {
						violations = append(violations, Violation{
							ErrorCode:        r.ErrorCode(),
							Severity:         SeverityError,
							Message:          "Focus indicator hidden via 'outline: none' without a visible alternative (like box-shadow or border). Keyboard users won't know where the focus is.",
							SourceRef:        node.Source,
							FixSnippet:       "Remove 'outline: none' or provide a high-contrast custom :focus style using box-shadow or border.",
							DocumentationURL: r.DocumentationURL(),
						})
					}
				}
			}
		}

		// 2. Tabindex management (WCAG 2.4.3)
		if hasTabIndex {
			if tabindexAttr == "-1" && isNativeInteractive {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_4_3",
					Severity:         SeverityWarning,
					Message:          "Interactive element has tabindex=\"-1\", removing it from the natural tab order. Ensure it is reachable via a custom focus management strategy.",
					SourceRef:        node.Source,
					FixSnippet:       "Remove tabindex=\"-1\" unless you are managing focus programmatically (e.g. modal, roving tabindex).",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
				})
			}

			var tabVal int
			if n, err := fmt.Sscanf(tabindexAttr, "%d", &tabVal); n == 1 && err == nil && tabVal > 0 {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_4_3_POSITIVE",
					Severity:         SeverityWarning,
					Message:          fmt.Sprintf("tabindex=\"%d\" (positive value) forces a custom focus order that may confuse keyboard and AT users.", tabVal),
					SourceRef:        node.Source,
					FixSnippet:       "Use tabindex=\"0\" to include elements in the natural DOM order instead of forcing a custom sequence.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
				})
			}
		}

		// 3. Interactive elements in aria-hidden areas (WCAG 2.4.3 / 4.1.2)
		if isNativeInteractive || (hasTabIndex && isPositiveOrZero(tabindexAttr)) {
			if _, inheritedHidden := node.Traits["aria-hidden-inherited"]; inheritedHidden {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_4_3_HIDDEN",
					Severity:         SeverityError,
					Message:          "Interactive element is inside an 'aria-hidden=\"true\"' container. Keyboard users can reach it, but screen reader users will ignore it.",
					SourceRef:        node.Source,
					FixSnippet:       "Remove 'aria-hidden=\"true\"' from the parent or remove the element from the tab order using tabindex=\"-1\".",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}
		}
	}

	return violations, nil
}
