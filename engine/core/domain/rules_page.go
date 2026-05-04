package domain

import (
	"context"
)

type rulePageLevel struct{}

func (r *rulePageLevel) Name() string             { return "Page Level Structure" }
func (r *rulePageLevel) ErrorCode() string        { return "WCAG_3_1_1" }
func (r *rulePageLevel) ACTID() string            { return "bf051a" }
func (r *rulePageLevel) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/html/H57" }

func (r *rulePageLevel) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	isWebProject := false
	isComponent := false
	for _, n := range analysisCtx.Nodes {
		p := n.Source.Platform
		if p == PlatformWebReact || p == PlatformWebVue || p == PlatformWebSvelte ||
			p == PlatformWebAngular || p == PlatformWebAstro ||
			p == PlatformBlazor || p == PlatformElectron || p == PlatformTauri {
			isWebProject = true
		}
		if n.Source.IsComponent {
			isComponent = true
		}
	}

	if isWebProject {
		severity := SeverityError
		contextNote := ""
		if isComponent {
			severity = SeverityWarning
			contextNote = " (component file - verify if this is a root document or a nested component)"
		}

		var firstSource Source
		if len(analysisCtx.Nodes) > 0 {
			for _, n := range analysisCtx.Nodes {
				if n.Source.Line > 0 {
					firstSource = n.Source
					break
				}
			}
			if firstSource.Line == 0 {
				firstSource = analysisCtx.Nodes[0].Source
			}
		}
		if firstSource.Line == 0 {
			firstSource.Line = 1
			firstSource.Column = 1
		}

		if !analysisCtx.HasLang {
			violations = append(violations, Violation{
				ErrorCode:        "WCAG_3_1_1",
				Severity:         severity,
				Message:          "Document missing language attribute." + contextNote,
				SourceRef:        firstSource,
				DocumentationURL: r.DocumentationURL(),
			})
		}
		if !analysisCtx.HasH1 {
			violations = append(violations, Violation{
				ErrorCode:        "G141",
				Severity:         severity,
				Message:          "Page missing an H1 heading." + contextNote,
				SourceRef:        firstSource,
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G141",
			})
		}
	}

	return violations, nil
}
