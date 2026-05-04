package domain

import (
	"context"
	"fmt"
)

type ruleAccessibleNames struct{}

func (r *ruleAccessibleNames) Name() string             { return "Name, Role, Value / Labels" }
func (r *ruleAccessibleNames) ErrorCode() string        { return "WCAG_4_1_2" } // Primary code
func (r *ruleAccessibleNames) ACTID() string            { return "674b10" }
func (r *ruleAccessibleNames) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14" }

func (r *ruleAccessibleNames) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput {
			effectiveLabel := node.Label
			if node.Role == RoleInput {
				if id, ok := node.Traits["id"].(string); ok {
					if l, found := analysisCtx.LabelsByFor[id]; found {
						effectiveLabel = l
					}
				}
			}

			if effectiveLabel == "" {
				code := "WCAG_4_1_2"
				if node.Role == RoleInput {
					code = "WCAG_3_3_2"
				}
				violations = append(violations, Violation{
					ErrorCode:        code,
					Severity:         SeverityError,
					Message:          fmt.Sprintf("%s missing accessible name or label.", node.Role),
					SourceRef:        node.Source,
					FixSnippet:       "Ensure the control has a visible label or an internal accessibility name.",
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}

		// Aria State for Interactive Elements
		if node.Role == RoleButton {
			role, _ := node.Traits["role"].(string)
			if role == "switch" || role == "checkbox" || role == "toggle" {
				_, hasPressed := node.Traits["aria-pressed"].(string)
				_, hasExpanded := node.Traits["aria-expanded"].(string)
				_, hasChecked := node.Traits["aria-checked"].(string)

				if !hasPressed && !hasExpanded && !hasChecked {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_4_1_2",
						Severity:         SeverityError,
						Message:          "Interactive button acting as toggle/switch missing state attribute (aria-pressed, aria-expanded, or aria-checked).",
						SourceRef:        node.Source,
						FixSnippet:       "Add the appropriate aria-state attribute to reflect the current toggle condition.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA5",
					})
				}
			}
		}
	}

	return violations, nil
}
