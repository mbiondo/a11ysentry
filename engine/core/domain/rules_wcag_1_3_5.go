package domain

import (
	"context"
	"fmt"
)

type ruleWCAG135 struct{}

func (r *ruleWCAG135) Name() string             { return "Identify Input Purpose" }
func (r *ruleWCAG135) ErrorCode() string        { return "WCAG_1_3_5" }
func (r *ruleWCAG135) ACTID() string            { return "6a71e1" } // Example ACT ID
func (r *ruleWCAG135) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/html/H98" }

func (r *ruleWCAG135) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleInput {
			inputType, _ := node.Traits["type"].(string)
			autocomplete, _ := node.Traits["autocomplete"].(string)

			tokens := map[string]string{
				"email": "email",
				"tel":   "tel",
				"url":   "url",
			}

			if expected, ok := tokens[inputType]; ok && autocomplete == "" {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityWarning,
					Message:          fmt.Sprintf("Input of type '%s' is missing an autocomplete attribute. Providing it helps users with cognitive disabilities.", inputType),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("autocomplete=\"%s\"", expected),
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}
