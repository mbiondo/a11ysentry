package vue

import (
	"context"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type vueAdapter struct {
	webAdapter ports.Adapter
}

func NewVueAdapter() ports.Adapter {
	return &vueAdapter{
		webAdapter: web.NewHTMLAdapter(),
	}
}

func (a *vueAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	// The webAdapter already has specific support for Vue bindings built-in 
	// (e.g. :alt, v-bind:alt, @click, etc.)
	// We can delegate directly to the web adapter and override the platform name.
	nodes, err := a.webAdapter.Ingest(ctx, root)
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		nodes[i].Source.Platform = domain.Platform("vue")
	}

	return nodes, nil
}
