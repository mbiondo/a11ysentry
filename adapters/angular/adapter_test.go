package angular_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	
	"a11ysentry/adapters/angular"
	"a11ysentry/engine/core/domain"
)

func TestAngularAdapter_Bindings(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create an Angular template with dynamic bindings
	htmlContent := `
	<img [attr.aria-label]="dynamicLabel" src="/logo.png" />
	<img [alt]="dynamicAlt" src="/hero.png" />
	<button (click)="submit()">Submit</button>
	<div *ngIf="showMenu">
		<nav></nav>
	</div>
	`

	os.WriteFile(filepath.Join(tmpDir, "app.component.html"), []byte(htmlContent), 0644)

	adapter := angular.NewAngularAdapter()
	rootNode := &domain.FileNode{FilePath: filepath.Join(tmpDir, "app.component.html")}

	nodes, err := adapter.Ingest(context.Background(), rootNode)
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	foundAriaLabel := false
	foundAlt := false
	foundClick := false
	
	for _, n := range nodes {
		if n.Role == domain.RoleImage {
			if n.Label == "{{dynamicLabel}}" {
				foundAriaLabel = true
			}
			if n.Label == "{{dynamicAlt}}" {
				foundAlt = true
			}
		}
		if n.Role == domain.RoleButton {
			if val, ok := n.Traits["(click)"]; ok && val == "submit()" {
				foundClick = true
			}
		}
	}
	
	if !foundAriaLabel {
		t.Errorf("Expected [attr.aria-label] to be parsed as dynamicLabel")
	}
	if !foundAlt {
		t.Errorf("Expected [alt] to be parsed as dynamicAlt")
	}
	if !foundClick {
		t.Errorf("Expected (click) to be preserved in traits")
	}
}
