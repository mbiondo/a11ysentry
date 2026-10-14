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

// excludeTestRepo is a no-op repository so the analyzer can run without a database.
type excludeTestRepo struct{}

func (excludeTestRepo) SaveReport(context.Context, domain.ViolationReport) error { return nil }
func (excludeTestRepo) GetHistory(context.Context, int) ([]domain.ViolationReport, error) {
	return nil, nil
}
func (excludeTestRepo) ClearHistory(context.Context) error { return nil }

var _ ports.Repository = excludeTestRepo{}

func writeExcludeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// S7: the MCP directory path must honor the exclude list from a11ysentry.json.
//
// Project discovery stops as soon as it finds a project root, so excludes never reached files
// inside an already-discovered root. Collection now filters them explicitly.
func TestMCPHonorsProjectExclude(t *testing.T) {
	dir := t.TempDir()
	img := `export default function X(){return (<img src="x.png" />);}`

	writeExcludeFixture(t, filepath.Join(dir, "a11ysentry.json"), `{"version":"1.0","exclude":["excluded"]}`)
	writeExcludeFixture(t, filepath.Join(dir, "included", "package.json"), `{"name":"inc"}`)
	writeExcludeFixture(t, filepath.Join(dir, "included", "src", "A.tsx"), img)
	writeExcludeFixture(t, filepath.Join(dir, "excluded", "package.json"), `{"name":"exc"}`)
	writeExcludeFixture(t, filepath.Join(dir, "excluded", "src", "B.tsx"), img)

	srv := &MCPServer{repo: excludeTestRepo{}}
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

	if strings.Contains(text.String(), "B.tsx") {
		t.Errorf("excluded directory was analyzed through MCP; a11ysentry.json exclude was ignored:\n%s", text.String())
	}
	if !strings.Contains(text.String(), "A.tsx") {
		t.Errorf("included directory was not analyzed:\n%s", text.String())
	}
}
