package domain

import (
	"context"
	"fmt"
)

type ruleWCAG131ARIA struct{}

func (r *ruleWCAG131ARIA) Name() string             { return "ARIA Required Parent/Child" }
func (r *ruleWCAG131ARIA) ErrorCode() string        { return "ARIA_REQ_PARENT" }
func (r *ruleWCAG131ARIA) ACTID() string            { return "bc659a" }
func (r *ruleWCAG131ARIA) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA11" }

func (r *ruleWCAG131ARIA) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	// Parent requirements map
	requiredParents := map[SemanticRole][]SemanticRole{
		RoleListItem: {RoleList},
		RoleTab:      {RoleTabList},
	}

	for _, node := range analysisCtx.Nodes {
		// Only check if it's a web platform for now
		if !isWebPlatform(node.Source.Platform) {
			continue
		}

		if parents, required := requiredParents[node.Role]; required {
			// In USN, the hierarchy is stored in the node.
			// However, in a flat list (analysisCtx.Nodes), we need to check the ParentID.
			
			foundParent := false
			if node.Hierarchy.ParentID != "" {
				for _, p := range analysisCtx.Nodes {
					// In some adapters, ParentID might be UID or a internal counter.
					// We check if the parent node has the required role.
					if p.UID == node.Hierarchy.ParentID || p.Traits["_internal_id"] == node.Hierarchy.ParentID {
						for _, reqRole := range parents {
							if p.Role == reqRole {
								foundParent = true
								break
							}
						}
					}
					if foundParent {
						break
					}
				}
			}

			if !foundParent {
				violations = append(violations, Violation{
					ErrorCode:        r.ErrorCode(),
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Role '%s' must be contained within a '%s'.", node.Role, parents[0]),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Wrap this element in a container with role=\"%s\".", parents[0]),
					DocumentationURL: r.DocumentationURL(),
				})
			}
		}
	}

	return violations, nil
}
