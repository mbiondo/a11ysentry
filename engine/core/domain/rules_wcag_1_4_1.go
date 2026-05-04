package domain

import (
	"context"
	"strings"
)

type ruleWCAG141 struct{}

func (r *ruleWCAG141) Name() string             { return "Use of Color" }
func (r *ruleWCAG141) ErrorCode() string        { return "WCAG_1_4_1" }
func (r *ruleWCAG141) ACTID() string            { return "1ec09b" } // Example ACT ID
func (r *ruleWCAG141) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G183" }

func (r *ruleWCAG141) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleLink {
			style, _ := node.Traits["style"].(string)
			noUnderline := strings.Contains(strings.ReplaceAll(style, " ", ""), "text-decoration:none") ||
				strings.Contains(strings.ReplaceAll(style, " ", ""), "text-decoration-line:none")
			
			if noUnderline {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityWarning,
					Message:          "Link has 'text-decoration: none' — if the only visual distinction from surrounding text is color, this violates WCAG 1.4.1.",
					SourceRef:        node.Source,
					FixSnippet:       "Ensure links are visually distinguishable from body text by at least one non-color cue (underline, bold, outline on hover, etc.).",
					DocumentationURL: r.DocumentationURL(),
				})
			}
			
			// Tailwind no-underline
			if noUnderlineClass, ok := node.Traits["no-underline"].(bool); ok && noUnderlineClass {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityWarning,
					Message:          "Link has 'no-underline' class — if the only visual distinction from surrounding text is color, this violates WCAG 1.4.1.",
					SourceRef:        node.Source,
					FixSnippet:       "Ensure links are visually distinguishable from body text by at least one non-color cue (underline, bold, outline on hover, etc.).",
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}
