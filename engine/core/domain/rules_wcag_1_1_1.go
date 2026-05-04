package domain

import (
	"context"
	"fmt"
)

type ruleWCAG111 struct{}

func (r *ruleWCAG111) Name() string             { return "Non-text Content" }
func (r *ruleWCAG111) ErrorCode() string        { return "WCAG_1_1_1" }
func (r *ruleWCAG111) ACTID() string            { return "23a2a8" }
func (r *ruleWCAG111) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G94" }

func (r *ruleWCAG111) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		p := node.Source.Platform
		isMobile := p == PlatformAndroidCompose || p == PlatformAndroidView || p == PlatformIOSSwiftUI || p == PlatformFlutterDart || p == PlatformReactNative
		isGaming := p == PlatformUnity || p == PlatformGodot

		if node.Role == RoleImage && node.Label == "" {
			msg := "Image missing alternative text."
			if isMobile {
				msg = "Mobile image missing content description."
			} else if isGaming {
				msg = "Gaming texture/sprite missing accessibility label."
			}

			violations = append(violations, Violation{
				ErrorCode:        r.ErrorCode(),
				Severity:         SeverityError,
				Message:          fmt.Sprintf("%s Every image must have an 'alt', 'aria-label', or platform-specific description attribute.", msg),
				SourceRef:        node.Source,
				FixSnippet:       "Add a descriptive label for users with screen readers.",
				DocumentationURL: r.DocumentationURL(),
			})
		}
	}

	return violations, nil
}
