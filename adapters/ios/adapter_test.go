package ios

import (
	"context"
	"os"
	"testing"
	"a11ysentry/engine/core/domain"
)

func TestIOSAdapter_Ingest(t *testing.T) {
	adapter := NewIOSAdapter()
	
	content := `
struct Test: View {
    var body: some View {
        Image("logo").accessibilityLabel("Test Logo")
        Button("Test Button") { }
    }
}
`
	tmpfile, err := os.CreateTemp("", "test*.swift")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) //nolint:errcheck
	
	if err := os.WriteFile(tmpfile.Name(), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: tmpfile.Name()})
	if err != nil {
		t.Fatal(err)
	}

	hasImage := false
	hasButton := false
	for _, n := range nodes {
		if n.Role == domain.RoleImage && n.Label == "Test Logo" {
			hasImage = true
		}
		if n.Role == domain.RoleButton && n.Label == "Test Button" {
			hasButton = true
		}
	}

	if !hasImage || !hasButton {
		t.Errorf("failed to ingest ios components. Image: %v, Button: %v", hasImage, hasButton)
	}
}
