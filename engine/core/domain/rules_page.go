package domain

import (
	"context"
)

type rulePageLevel struct{}

func (r *rulePageLevel) Name() string             { return "Page Level Structure" }
func (r *rulePageLevel) ErrorCode() string        { return "G141" }
func (r *rulePageLevel) ACTID() string            { return "" }
func (r *rulePageLevel) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/general/G141" }

func (r *rulePageLevel) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	isWebProject := false
	isComponent := false
	for _, n := range analysisCtx.Nodes {
		p := n.Source.Platform
		if isWebPlatform(p) {
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

		if !analysisCtx.HasH1 {
			firstSource := getFirstAvailableSource(analysisCtx.Nodes)
			violations = append(violations, Violation{
				ErrorCode:        "G141",
				Severity:         severity,
				Message:          "Page missing an H1 heading." + contextNote,
				SourceRef:        firstSource,
				DocumentationURL: r.DocumentationURL(),
			})
		}
	}

	return violations, nil
}
