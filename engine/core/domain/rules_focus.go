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
		tabindexAttr, hasTabIndex := node.Traits["tabindex"].(string)

		// Focus Visibility
		if isNativeInteractive {
			if style, ok := node.Traits["style"].(string); ok {
				styleLower := strings.ReplaceAll(style, " ", "")
				if strings.Contains(styleLower, "outline:none") ||
					strings.Contains(styleLower, "outline:0") ||
					strings.Contains(styleLower, "outline:0px") {
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          "Focus indicator hidden via 'outline: none'. Keyboard users won't know where the focus is.",
						SourceRef:        node.Source,
						FixSnippet:       "Remove 'outline: none' or provide a high-contrast custom :focus style.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}

		// Tabindex negative on interactive
		if hasTabIndex && tabindexAttr == "-1" && isNativeInteractive {
			violations = append(violations, Violation{
				ErrorCode:        "WCAG_2_4_3",
				Severity:         SeverityWarning,
				Message:          "Interactive element has tabindex=\"-1\", removing it from the natural tab order. Ensure it is reachable via a custom focus management strategy.",
				SourceRef:        node.Source,
				FixSnippet:       "Remove tabindex=\"-1\" unless you are managing focus programmatically (e.g. modal, roving tabindex).",
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
			})
		}

		// Tabindex positive
		if hasTabIndex {
			var tabVal int
			if n, err := fmt.Sscanf(tabindexAttr, "%d", &tabVal); n == 1 && err == nil && tabVal > 0 {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_4_3",
					Severity:         SeverityWarning,
					Message:          fmt.Sprintf("tabindex=\"%d\" (positive value) forces a custom focus order that may confuse keyboard and AT users.", tabVal),
					SourceRef:        node.Source,
					FixSnippet:       "Use tabindex=\"0\" to include elements in the natural DOM order instead of forcing a custom sequence.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
				})
			}
		}

		// Keyboard navigation for clickables (WCAG 2.1.1)
		hasMouseClick := node.Traits["onclick"] != nil || node.Traits["@click"] != nil || node.Traits["v-on:click"] != nil || node.Traits["(click)"] != nil || node.Traits["on:click"] != nil
		hasKeyboard := node.Traits["onkeydown"] != nil || node.Traits["@keydown"] != nil || node.Traits["v-on:keydown"] != nil || node.Traits["(keydown)"] != nil || node.Traits["on:keydown"] != nil ||
			node.Traits["onkeyup"] != nil || node.Traits["@keyup"] != nil || node.Traits["v-on:keyup"] != nil || node.Traits["(keyup)"] != nil || node.Traits["on:keyup"] != nil ||
			node.Traits["onkeypress"] != nil || node.Traits["@keypress"] != nil || node.Traits["v-on:keypress"] != nil || node.Traits["(keypress)"] != nil || node.Traits["on:keypress"] != nil

		if hasMouseClick && !isNativeInteractive {
			if !hasKeyboard || !hasTabIndex {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_1_1",
					Severity:         SeverityError,
					Message:          "Non-interactive element with a click handler is missing keyboard support (keydown/keyup) or is not focusable (tabindex).",
					SourceRef:        node.Source,
					FixSnippet:       "Add a keydown handler and tabindex=\"0\", or change the element to a native <button>.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G90",
				})
			}
		}

		// Interactive elements in aria-hidden areas
		if isNativeInteractive || hasTabIndex {
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
