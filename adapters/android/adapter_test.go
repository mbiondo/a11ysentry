package android

import (
	"context"
	"os"
	"testing"
	"a11ysentry/engine/core/domain"
)

func TestAndroidAdapter_Ingest(t *testing.T) {
	adapter := NewAndroidAdapter()
	
	content := `
@Composable
fun Test() {
    Image(contentDescription = "Test Image")
    Button() { Text("Test Button") }
}
`
	tmpfile, err := os.CreateTemp("", "test*.kt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	os.WriteFile(tmpfile.Name(), []byte(content), 0644)

	nodes, err := adapter.Ingest(context.Background(), []string{tmpfile.Name()})
	if err != nil {
		t.Fatal(err)
	}

	hasImage := false
	hasButton := false
	for _, n := range nodes {
		if n.Role == domain.RoleImage && n.Label == "Test Image" {
			hasImage = true
		}
		if n.Role == domain.RoleButton && n.Label == "Test Button" {
			hasButton = true
		}
	}

	if !hasImage || !hasButton {
		t.Errorf("failed to ingest mobile components. Image: %v, Button: %v", hasImage, hasButton)
	}
}
