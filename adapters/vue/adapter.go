package vue

import (
	"context"
	"strings"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type vueAdapter struct {
	webAdapter ports.Adapter
}

func vueMapper(key, val string) (string, string) {
	// Map shorthand events: @click -> onclick
	if strings.HasPrefix(key, "@") {
		return "on" + key[1:], val
	}
	
	// Map explicit events: v-on:click -> onclick
	if strings.HasPrefix(key, "v-on:") {
		return "on" + key[5:], val
	}

	// Map shorthand properties: :alt -> alt
	if strings.HasPrefix(key, ":") {
		return key[1:], "{{" + val + "}}"
	}

	// Map explicit properties: v-bind:alt -> alt
	if strings.HasPrefix(key, "v-bind:") {
		return key[7:], "{{" + val + "}}"
	}
	
	return key, val
}

func NewVueAdapter() ports.Adapter {
	return &vueAdapter{
		webAdapter: web.NewHTMLAdapterWithMapper(domain.Platform("vue"), vueMapper),
	}
}

func (a *vueAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	nodes, err := a.webAdapter.Ingest(ctx, root)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
