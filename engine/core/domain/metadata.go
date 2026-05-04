package domain

import (
	"fmt"
)

// BuildAnalysisContext performs a first pass over the nodes to collect metadata.
func BuildAnalysisContext(nodes []USN, cfg ProjectConfig) *AnalysisContext {
	ctx := &AnalysisContext{
		Nodes:          nodes,
		Config:         cfg,
		LabelsByFor:    make(map[string]string),
		LandmarkLabels: make(map[SemanticRole]map[string]Source),
		LinksByLabel:   make(map[string]map[string][]int),
	}

	for i, node := range nodes {
		// Structural link analysis
		if node.Role == RoleLink && node.Label != "" {
			href, _ := node.Traits["href"].(string)
			if _, ok := ctx.LinksByLabel[node.Label]; !ok {
				ctx.LinksByLabel[node.Label] = make(map[string][]int)
			}
			ctx.LinksByLabel[node.Label][href] = append(ctx.LinksByLabel[node.Label][href], i)
		}

		// Identify Web document-level traits
		if node.UID == "html" || node.UID == "html-tag" {
			if lang, ok := node.Traits["lang"].(string); ok && lang != "" {
				ctx.HasLang = true
			}
		}

		// Collect <label for="..."> — support both 'for' (HTML) and 'htmlFor' (JSX/React)
		forAttr := ""
		if f, ok := node.Traits["for"].(string); ok && f != "" {
			forAttr = f
		} else if f, ok := node.Traits["htmlFor"].(string); ok && f != "" {
			forAttr = f
		}
		if forAttr != "" {
			if node.Label != "" {
				ctx.LabelsByFor[forAttr] = node.Label
			}
		}

		// Initialize landmark tracking maps
		if _, exists := ctx.LandmarkLabels[node.Role]; !exists && isLandmark(node.Role) {
			ctx.LandmarkLabels[node.Role] = make(map[string]Source)
		}

		if isLandmark(node.Role) {
			if node.Role == RoleMain {
				ctx.MainCount++
			}
			if node.Label != "" {
				ctx.LandmarkLabels[node.Role][node.Label] = node.Source
			}
		}

		if node.Role == RoleHeading {
			level := 0
			_, _ = fmt.Sscanf(string(node.UID), "h%d", &level)
			if level == 1 {
				ctx.HasH1 = true
			}
		}
	}

	return ctx
}
