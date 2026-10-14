package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"

	"github.com/mark3labs/mcp-go/mcp"
)

type s8Repo struct{}

func (s8Repo) SaveReport(context.Context, domain.ViolationReport) error { return nil }
func (s8Repo) GetHistory(context.Context, int) ([]domain.ViolationReport, error) {
	return nil, nil
}
func (s8Repo) ClearHistory(context.Context) error { return nil }

var _ ports.Repository = s8Repo{}

func writeS8Fixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// S8: the MCP directory path must already be cross-tree unique, so a component shared by three
// pages is reported once. This pins the claim the S8 commit message makes about the MCP path.
func TestS8MCPReportsSharedComponentOnce(t *testing.T) {
	dir := t.TempDir()
	shared := `export default function Shared(){return (<img src="/logo.png" />);}`

	writeS8Fixture(t, filepath.Join(dir, "package.json"), `{"name":"fx"}`)
	writeS8Fixture(t, filepath.Join(dir, "next.config.ts"), `export default {}`)
	writeS8Fixture(t, filepath.Join(dir, "components", "Shared.tsx"), shared)
	writeS8Fixture(t, filepath.Join(dir, "app", "page.tsx"),
		"import Shared from \"../components/Shared\";\nexport default function Home(){return (<div><Shared /></div>);}")
	writeS8Fixture(t, filepath.Join(dir, "app", "about", "page.tsx"),
		"import Shared from \"../../components/Shared\";\nexport default function About(){return (<div><Shared /></div>);}")
	writeS8Fixture(t, filepath.Join(dir, "app", "contact", "page.tsx"),
		"import Shared from \"../../components/Shared\";\nexport default function Contact(){return (<div><Shared /></div>);}")

	srv := &MCPServer{repo: s8Repo{}}
	res, err := srv.analyzeDirectory(context.Background(), dir)
	if err != nil {
		t.Fatalf("analyzeDirectory: %v", err)
	}

	var text strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	out := text.String()
	if n := strings.Count(out, filepath.Join("components", "Shared.tsx")); n != 1 {
		t.Errorf("MCP reported the shared component %d times, want 1:\n%s", n, out)
	}
}
