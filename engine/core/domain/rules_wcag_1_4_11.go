package domain

import (
	"context"
	"fmt"
)

type ruleWCAG1411 struct{}

func (r *ruleWCAG1411) Name() string             { return "Non-text Contrast" }
func (r *ruleWCAG1411) ErrorCode() string        { return "WCAG_1_4_11" }
func (r *ruleWCAG1411) ACTID() string            { return "0e3e20" } // Example ACT ID
func (r *ruleWCAG1411) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G195" }

func (r *ruleWCAG1411) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleInput {
			borderColor, hasBorder := node.Traits["border-color"].(string)
			bg, hasBg := node.Traits["background-color"].(string)
			if !hasBg {
				bg = "#ffffff"
			}
			if hasBorder {
				ratio := CalculateContrast(borderColor, bg)
				if ratio < 3.0 {
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Input border has insufficient contrast ratio (%.2f:1). Minimum for UI components is 3:1.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Increase the border color contrast so the input boundary is clearly visible against its background.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}
	}

	return violations, nil
}
