package pyqt_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	
	"a11ysentry/adapters/pyqt"
	"a11ysentry/engine/core/domain"
)

func TestPyQtAdapter_XMLParsing(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a Qt Designer .ui file
	uiContent := `
<?xml version="1.0" encoding="UTF-8"?>
<ui version="4.0">
 <class>MainWindow</class>
 <widget class="QMainWindow" name="MainWindow">
  <widget class="QWidget" name="centralwidget">
   <widget class="QPushButton" name="pushButton">
    <property name="text">
     <string>Submit</string>
    </property>
    <property name="accessibleName">
     <string>Submit Form</string>
    </property>
   </widget>
   <widget class="QLabel" name="label">
    <property name="text">
     <string>Instructions</string>
    </property>
   </widget>
  </widget>
 </widget>
</ui>
	`

	os.WriteFile(filepath.Join(tmpDir, "main.ui"), []byte(uiContent), 0644)

	adapter := pyqt.NewPyQtAdapter()
	rootNode := &domain.FileNode{FilePath: filepath.Join(tmpDir, "main.ui")}

	nodes, err := adapter.Ingest(context.Background(), rootNode)
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	foundButton := false
	foundLabel := false
	
	for _, n := range nodes {
		t.Logf("Node: Role=%s, UID=%s, Label=%s, Traits=%v", n.Role, n.UID, n.Label, n.Traits)
		if n.UID == "pushButton" && n.Role == domain.RoleButton {
			foundButton = true
			if n.Label != "Submit Form" {
				t.Errorf("Expected button label to be 'Submit Form', got '%s'", n.Label)
			}
		}
		if n.UID == "label" {
			foundLabel = true
		}
	}
	
	if !foundButton {
		t.Errorf("Expected to find pushButton")
	}
	if !foundLabel {
		t.Errorf("Expected to find QLabel parsed")
	}
}
