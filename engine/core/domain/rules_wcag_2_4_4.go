package domain

import (
	"context"
	"fmt"
)

type ruleWCAG244 struct{}

func (r *ruleWCAG244) Name() string             { return "Link Purpose (In Context)" }
func (r *ruleWCAG244) ErrorCode() string        { return "WCAG_2_4_4" }
func (r *ruleWCAG244) ACTID() string            { return "c487ae" }
func (r *ruleWCAG244) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G91" }

func (r *ruleWCAG244) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for label, hrefs := range analysisCtx.LinksByLabel {
		if len(hrefs) > 1 {
			for _, indices := range hrefs {
				for _, idx := range indices {
					node := analysisCtx.Nodes[idx]
					if _, hasALabel := node.Traits["aria-label"].(string); hasALabel {
						continue
					}
					if _, hasALabelled := node.Traits["aria-labelledby"].(string); hasALabelled {
						continue
					}

					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Multiple links have the same label '%s' but point to different destinations. This is ambiguous for screen reader users.", label),
						SourceRef:        node.Source,
						FixSnippet:       fmt.Sprintf("Add aria-label=\"%s - [Context]\" to distinguish this link.", label),
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}
	}

	return violations, nil
}
