package flask_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	
	"a11ysentry/adapters/flask"
	"a11ysentry/engine/core/domain"
)

func TestFlaskAdapter_InheritanceAndInclude(t *testing.T) {
	tmpDir := t.TempDir()
	
	baseHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head><title>Base</title></head>
	<body>
		<nav>{% include "nav.html" %}</nav>
		<main>
			{% block content %}{% endblock %}
		</main>
	</body>
	</html>
	`
	navHTML := `<button id="nav-btn" aria-label="Menu"></button>`
	childHTML := `
	{% extends "base.html" %}
	{% block content %}
		<h1 id="child-h1">Child Page</h1>
	{% endblock %}
	`

	os.WriteFile(filepath.Join(tmpDir, "base.html"), []byte(baseHTML), 0644)
	os.WriteFile(filepath.Join(tmpDir, "nav.html"), []byte(navHTML), 0644)
	os.WriteFile(filepath.Join(tmpDir, "child.html"), []byte(childHTML), 0644)

	adapter := flask.NewFlaskAdapter()
	rootNode := &domain.FileNode{FilePath: filepath.Join(tmpDir, "child.html")}
	
	rootNode.Children = []*domain.FileNode{
		{FilePath: filepath.Join(tmpDir, "base.html")},
		{FilePath: filepath.Join(tmpDir, "nav.html")},
	}

	nodes, err := adapter.Ingest(context.Background(), rootNode)
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	foundNavBtn := false
	foundChildH1 := false
	
	for _, n := range nodes {
		if n.UID == "nav-btn" {
			foundNavBtn = true
		}
		if n.UID == "child-h1" {
			foundChildH1 = true
		}
	}
	
	if !foundNavBtn {
		t.Errorf("Expected to find nav-btn from included nav.html")
	}
	if !foundChildH1 {
		t.Errorf("Expected to find child-h1 from child.html block content")
	}
}
