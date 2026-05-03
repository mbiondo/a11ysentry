package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"a11ysentry/scanner/angular"
	"a11ysentry/scanner/vue"
)

func TestAngularScanner(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "angular.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	fw := angular.New()
	if !fw.Probe(tmp) {
		t.Fatal("Angular scanner should probe positive for project with angular.json")
	}
}

func TestVueScanner(t *testing.T) {
	tmp := t.TempDir()
	
	// Create a package.json with vue dependency
	packageJSON := `{"dependencies": {"vue": "^3.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	fw := vue.New()
	if !fw.Probe(tmp) {
		t.Fatal("Vue scanner should probe positive for project with vue in package.json")
	}
}

func TestVueScanner_SkipNuxt(t *testing.T) {
	tmp := t.TempDir()
	
	// Create a package.json with nuxt dependency
	packageJSON := `{"dependencies": {"vue": "^3.0.0", "nuxt": "^3.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	fw := vue.New()
	if fw.Probe(tmp) {
		t.Fatal("Vue scanner should probe negative for Nuxt projects")
	}
}
