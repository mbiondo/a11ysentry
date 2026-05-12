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
		if node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput {
			p := node.Source.Platform
			isMobile := p == PlatformAndroidCompose || p == PlatformAndroidView || p == PlatformIOSSwiftUI || p == PlatformFlutterDart || p == PlatformReactNative
			isGaming := p == PlatformUnity || p == PlatformGodot

			// Get dimensions from traits or geometry
			w, _ := node.Traits["width"].(float64)
			if w == 0 {
				w = node.Geometry.W
			}
			h, _ := node.Traits["height"].(float64)
			if h == 0 {
				h = node.Geometry.H
			}

			// If both are 0, we can't validate (might be dynamic or unresolved)
			if w == 0 && h == 0 {
				continue
			}

			targetSize := 44.0 // Mobile default (WCAG 2.5.5 AAA)
			errorCode := "WCAG_2_5_5"
			severity := SeverityError

			if !isMobile && !isGaming {
				// Web WCAG 2.2 SC 2.5.8 (AA)
				targetSize = 24.0 
				errorCode = "WCAG_2_5_8"
				severity = SeverityWarning // AA is usually Warning/Error depending on strictness, but 2.5.8 is 24px
			}

			if w < targetSize || h < targetSize {
				violations = append(violations, Violation{
					ErrorCode:        errorCode,
					Severity:         severity,
					Message:          fmt.Sprintf("Interactive target is too small (%.0fx%.0f). Minimum recommended size for this platform is %.0fx%.0fpx.", w, h, targetSize, targetSize),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Increase the element size or padding to reach at least %.0fx%.0fpx.", targetSize, targetSize),
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}
