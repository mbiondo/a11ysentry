package domain

import (
	"context"
	"fmt"
)

type ruleWCAG143 struct{}

func (r *ruleWCAG143) Name() string             { return "Contrast (Minimum)" }
func (r *ruleWCAG143) ErrorCode() string        { return "WCAG_1_4_3" }
func (r *ruleWCAG143) ACTID() string            { return "afw4f7" }
func (r *ruleWCAG143) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G18" }

func (r *ruleWCAG143) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleHeading || node.Role == RoleButton || node.Role == "generic" {
			fg, hasFg := node.Traits["color"].(string)
			bg, hasBg := node.Traits["background-color"].(string)
			if hasFg && hasBg {
				ratio := CalculateContrast(fg, bg)
				if ratio < 4.5 {
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Low contrast ratio (%.2f:1). Target is 4.5:1.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Adjust colors to meet WCAG AA standards.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			} else if hasFg || hasBg {
				isTextBearer := node.Role == RoleHeading ||
					node.Role == RoleButton ||
					node.Role == RoleLink ||
					(node.Role == "generic" && node.Label != "")
				if isTextBearer {
					missing := "color and background-color"
					if hasFg {
						missing = "background-color"
					} else if hasBg {
						missing = "color"
					}
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_UNRESOLVED",
						Severity:         SeverityWarning,
						Message:          fmt.Sprintf("Color contrast could not be validated: %s is not statically resolvable (CSS variable, custom token, or runtime value).", missing),
						SourceRef:        node.Source,
						FixSnippet:       "Provide the resolved color values via --css or ensure colors are defined in static CSS/Tailwind classes.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}

			// Dark mode contrast
			darkFg, hasDarkFg := node.Traits["dark:color"].(string)
			darkBg, hasDarkBg := node.Traits["dark:background-color"].(string)
			if hasDarkFg && hasDarkBg {
				ratio := CalculateContrast(darkFg, darkBg)
				if ratio < 4.5 {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_DARK",
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Dark mode contrast ratio (%.2f:1) is below 4.5:1 minimum.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Adjust the dark: color classes to ensure sufficient contrast in dark mode.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			}
		}
	}

	return violations, nil
}
