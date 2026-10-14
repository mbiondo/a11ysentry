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
		// role="alert", role="status" and role="log" already imply aria-live in WAI-ARIA 1.2
		// (assertive, polite and polite respectively), so demanding an explicit aria-live on them
		// was redundant and produced false positives on every status message.
		//
		// A malformed explicit value is still a real defect: it overrides the implicit live
		// behaviour with something an assistive technology cannot honour.
		live, hasLive := node.Traits["aria-live"].(string)
		if !hasLive || live == "" {
			continue
		}

		switch live {
		case "off", "polite", "assertive":
			// Valid values.
		default:
			violations = append(violations, Violation{
				ErrorCode:        r.ErrorCode(),
				Severity:         SeverityWarning,
				Message:          fmt.Sprintf("Invalid aria-live value %q. Assistive technologies will not announce updates with an unrecognised value.", live),
				SourceRef:        node.Source,
				FixSnippet:       "Use aria-live=\"polite\" for status updates or aria-live=\"assertive\" for critical alerts.",
				DocumentationURL: r.DocumentationURL(),
			})
		}
	}

	return violations, nil
}
