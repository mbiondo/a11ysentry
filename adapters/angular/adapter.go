package angular

import (
	"context"
	"strings"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type angularAdapter struct {
	webAdapter ports.Adapter
}

func angularMapper(key, val string) (string, string) {
	// Map event bindings: (click) -> onclick
	if strings.HasPrefix(key, "(") && strings.HasSuffix(key, ")") {
		eventName := key[1 : len(key)-1]
		return "on" + eventName, val
	}
	
	// Map attribute bindings: [attr.aria-label] -> aria-label
	if strings.HasPrefix(key, "[attr.") && strings.HasSuffix(key, "]") {
		attrName := key[6 : len(key)-1]
		return attrName, "{{" + val + "}}"
	}
	
	// Map property bindings: [alt] -> alt
	if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
		attrName := key[1 : len(key)-1]
		return attrName, "{{" + val + "}}"
	}
	
	return key, val
}

func NewAngularAdapter() ports.Adapter {
	return &angularAdapter{
		webAdapter: web.NewHTMLAdapterWithMapper(domain.Platform("angular"), angularMapper),
	}
}

func (a *angularAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	return a.webAdapter.Ingest(ctx, root)
}
