package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type ruleWCAG143 struct{}

func (r *ruleWCAG143) Name() string             { return "Contrast (Minimum)" }
func (r *ruleWCAG143) ErrorCode() string        { return "WCAG_1_4_3" }
func (r *ruleWCAG143) ACTID() string            { return "afw4f7" }
func (r *ruleWCAG143) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G18" }

func (r *ruleWCAG143) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		// Skip hidden elements
		if _, hidden := node.Traits["aria-hidden-inherited"]; hidden {
			continue
		}
		if h, ok := node.Traits["aria-hidden"].(string); ok && h == "true" {
			continue
		}
		// Skip visually-hidden elements (sr-only, visually-hidden, etc.) — contrast is irrelevant
		if cls, ok := node.Traits["class"].(string); ok {
			if strings.Contains(cls, "sr-only") || strings.Contains(cls, "visually-hidden") {
				continue
			}
		}

		isTextBearer := node.Role == RoleHeading ||
			node.Role == RoleButton ||
			node.Role == RoleLink ||
			node.Role == RoleInput ||
			(node.Role == "generic" && node.Label != "")

		if isTextBearer {
			fg, hasFg := node.Traits["color"].(string)
			bg, hasBg := node.Traits["background-color"].(string)

			// Determine target ratio based on font size and weight
			targetRatio := 4.5
			isLargeText := false

			fontSize, _ := node.Traits["font-size"].(string)
			fontWeight, _ := node.Traits["font-weight"].(string)

			// Robust size detection
			sizeVal := 0.0
			if fontSize != "" {
				cleanSize := strings.TrimSuffix(fontSize, "px")
				cleanSize = strings.TrimSuffix(cleanSize, "pt")
				if val, err := strconv.ParseFloat(cleanSize, 64); err == nil {
					sizeVal = val
				}
			}
			
			// Large text: 18pt (24px) or 14pt (18.66px) bold
			isBold := fontWeight == "bold" || fontWeight == "700" || fontWeight == "800" || fontWeight == "900"
			
			if sizeVal >= 24 || (sizeVal >= 18.66 && isBold) {
				targetRatio = 3.0
				isLargeText = true
			}

			if hasFg && hasBg {
				ratio := CalculateContrast(fg, bg)
				if ratio < 0 {
					// Colors are present but cannot be parsed (empty or malformed) — skip silently
				} else if ratio < targetRatio {
					msg := fmt.Sprintf("Low contrast ratio (%.2f:1). Target is %.1f:1 for normal text.", ratio, targetRatio)
					if isLargeText {
						msg = fmt.Sprintf("Low contrast ratio (%.2f:1). Target is %.1f:1 for large/bold text.", ratio, targetRatio)
					}
					violations = append(violations, Violation{
						ErrorCode:        r.ErrorCode(),
						Severity:         SeverityError,
						Message:          msg,
						SourceRef:        node.Source,
						FixSnippet:       "Adjust colors to meet WCAG AA standards.",
						DocumentationURL: r.DocumentationURL(),
					})
				}
			} else if hasFg || hasBg {
				// Only emit UNRESOLVED if there was an explicit color class that failed to resolve.
				// If the color is simply absent (inherited from global CSS), silencing is correct —
				// the user cannot fix what they haven't declared.
				hasUnresolvedColor := node.Traits["has-unresolved-color"] != nil
				hasUnresolvedBg := node.Traits["has-unresolved-bg"] != nil
				if !hasFg && !hasUnresolvedColor {
					// FG missing but no unresolved color class — it's a pure inheritance case, skip
				} else if !hasBg && !hasUnresolvedBg {
					// BG missing but no unresolved bg class — it's a pure inheritance case, skip
				} else {
					missing := "color and background-color"
					if hasFg {
						missing = "background-color"
					} else if hasBg {
						missing = "color"
					}
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_UNRESOLVED",
						Severity:         SeverityWarning,
						Message:          fmt.Sprintf("Color contrast could not be validated: %s is not statically resolvable (CSS variable; custom token; or runtime value).", missing),
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
				if ratio < targetRatio {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_DARK",
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Dark mode contrast ratio (%.2f:1) is below %.1f:1 minimum.", ratio, targetRatio),
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
