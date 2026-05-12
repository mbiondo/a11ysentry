package domain

import (
	"context"
	"fmt"
)

type ruleWCAG211 struct{}

func (r *ruleWCAG211) Name() string             { return "Keyboard" }
func (r *ruleWCAG211) ErrorCode() string        { return "WCAG_2_1_1" }
func (r *ruleWCAG211) ACTID() string            { return "674b10" }
func (r *ruleWCAG211) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G90" }

func (r *ruleWCAG211) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		// Only applicable to Web for now as mobile platforms handle keyboard/focus differently
		if !isWebPlatform(node.Source.Platform) {
			continue
		}

		// Skip nodes where the source position was not resolved (line 1, col 1 is the fallback
		// used when the adapter cannot determine the exact location of an element within the file).
		// Reporting these would produce misleading violations pointing at the top of the file.
		if node.Source.Line == 1 && node.Source.Column == 1 {
			continue
		}

		// Detect elements with click handlers but no keyboard support
		hasClick := hasAnyTrait(node, "onclick", "@click", "v-on:click", "(click)")
		hasKeyboard := hasAnyTrait(node, "onkeydown", "onkeyup", "onkeypress", "@keydown", "v-on:keydown", "(keydown)")
		tabIndex, hasTabIndex := node.Traits["tabindex"]

		// 1. Elements with click handlers MUST be focusable and have keyboard handlers
		if hasClick {
			isFocusable := hasTabIndex && isPositiveOrZero(tabIndex)

			// If it's a generic element (div, span, etc.) and has a click handler,
			// it MUST have a tabindex and a keyboard handler.
			if node.Role == "generic" {
				if !isFocusable {
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          "Element has a click handler but is not focusable (missing tabindex=\"0\").",
						SourceRef:        node.Source,
						FixSnippet:       "Add tabindex=\"0\" to make the element focusable via keyboard.",
						DocumentationURL: r.DocumentationURL(),
					})
				} else if !hasKeyboard {
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          "Element has a click handler and is focusable, but missing a keyboard handler (onkeydown).",
						SourceRef:        node.Source,
						FixSnippet:       "Add an onkeydown or onkeyup handler to support keyboard interaction.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}

		// 2. Elements with focusable tabindex but no role (for generic elements)
		if node.Role == "generic" && hasTabIndex && isPositiveOrZero(tabIndex) {
			_, hasRole := node.Traits["role"]
			if !hasRole {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_1_1_ROLE",
					Severity:         SeverityWarning,
					Message:          "Focusable element is missing a semantic role. Screen readers won't know what this element does.",
					SourceRef:        node.Source,
					FixSnippet:       "Add a role attribute (e.g., role=\"button\") to provide semantic meaning.",
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}

func isWebPlatform(p Platform) bool {
	switch p {
	case PlatformWebReact, PlatformWebVue, PlatformWebSvelte, PlatformWebAngular, PlatformWebAstro:
		return true
	default:
		return false
	}
}

func hasAnyTrait(node USN, traits ...string) bool {
	for _, t := range traits {
		if _, ok := node.Traits[t]; ok {
			return true
		}
	}
	return false
}

func isPositiveOrZero(val any) bool {
	switch v := val.(type) {
	case int:
		return v >= 0
	case float64:
		return v >= 0
	case string:
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		return err == nil && i >= 0
	default:
		return false
	}
}
