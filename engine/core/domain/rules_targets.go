package domain

import (
	"context"
	"fmt"
)

type ruleTouchTargets struct{}

func (r *ruleTouchTargets) Name() string             { return "Touch Target Size" }
func (r *ruleTouchTargets) ErrorCode() string        { return "WCAG_2_5_5" }
func (r *ruleTouchTargets) ACTID() string            { return "q8950f" } // Example ACT ID
func (r *ruleTouchTargets) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G202" }

func (r *ruleTouchTargets) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	for _, node := range analysisCtx.Nodes {
		if node.Role == RoleButton || node.Role == RoleLink {
			p := node.Source.Platform
			isMobile := p == PlatformAndroidCompose || p == PlatformAndroidView || p == PlatformIOSSwiftUI || p == PlatformFlutterDart || p == PlatformReactNative
			isGaming := p == PlatformUnity || p == PlatformGodot

			w, _ := node.Traits["width"].(float64)
			h, _ := node.Traits["height"].(float64)

			targetSize := 44.0 // Mobile default
			errorCode := "WCAG_2_5_5"

			if !isMobile && !isGaming {
				targetSize = 24.0 // Web WCAG 2.2 SC 2.5.8
				errorCode = "WCAG_2_5_8"
			}

			if (w > 0 && w < targetSize) || (h > 0 && h < targetSize) {
				violations = append(violations, Violation{
					ErrorCode:        errorCode,
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Touch target too small (%.0fx%.0f). Target is %.0fx%.0f px/DP.", w, h, targetSize, targetSize),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Increase the size to at least %.0fx%.0f.", targetSize, targetSize),
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}
