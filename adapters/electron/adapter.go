package electron

import (
	"context"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type electronAdapter struct {
	webAdapter ports.Adapter
}

func NewElectronAdapter() ports.Adapter {
	return &electronAdapter{
		webAdapter: web.NewHTMLAdapter(),
	}
}

func (a *electronAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	nodes, err := a.webAdapter.Ingest(ctx, root)
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		nodes[i].Source.Platform = domain.Platform("electron")
	}

	return nodes, nil
}
