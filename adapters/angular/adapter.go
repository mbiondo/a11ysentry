package angular

import (
	"context"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type angularAdapter struct {
	webAdapter ports.Adapter
}

func NewAngularAdapter() ports.Adapter {
	return &angularAdapter{
		webAdapter: web.NewHTMLAdapter(),
	}
}

func (a *angularAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	// The webAdapter already has specific support for [alt] and [attr.alt] bindings built-in 
	// (see adapters/web/adapter.go lines 220-224) 
	// So we can just delegate directly and override the platform name.
	nodes, err := a.webAdapter.Ingest(ctx, root)
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		nodes[i].Source.Platform = domain.Platform("angular")
		
		// Additional fixup for [attr.aria-label] which isn't natively handled in webAdapter yet
		if val, ok := nodes[i].Traits["[attr.aria-label]"]; ok {
			strVal, _ := val.(string)
			if strVal != "" {
				nodes[i].Label = "{{" + strVal + "}}"
			}
		}
	}

	return nodes, nil
}
