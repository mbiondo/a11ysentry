package domain

import (
	"context"
	"fmt"
)

type ruleWCAG413 struct{}

func (r *ruleWCAG413) Name() string             { return "Status Messages" }
func (r *ruleWCAG413) ErrorCode() string        { return "WCAG_4_1_3" }
func (r *ruleWCAG413) ACTID() string            { return "be4d0c" } // Example ACT ID
func (r *ruleWCAG413) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA22" }

func (r *ruleWCAG413) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if roleAttr, ok := node.Traits["role"].(string); ok {
			switch roleAttr {
			case "status", "alert", "log":
				if _, hasLive := node.Traits["aria-live"].(string); !hasLive {
					implicit := map[string]string{
						"status": "polite",
						"alert":  "assertive",
						"log":    "polite",
					}
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityWarning,
						Message:          fmt.Sprintf("Element with role=\"%s\" is missing an explicit aria-live attribute. Some AT may not announce it automatically.", roleAttr),
						SourceRef:        node.Source,
						FixSnippet:       fmt.Sprintf("Add aria-live=\"%s\" to ensure screen readers announce status changes without requiring focus.", implicit[roleAttr]),
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}

		if node.Role == RoleLiveRegion {
			_, hasLive := node.Traits["aria-live"].(string)
			if !hasLive {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityWarning,
					Message:          "Live region missing explicit 'aria-live' attribute. Dynamic content updates may not be announced to screen reader users.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-live=\"polite\" (for status updates) or aria-live=\"assertive\" (for critical alerts).",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA19",
				})
			}
		}
	}

	return violations, nil
}
