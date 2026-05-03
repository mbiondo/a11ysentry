package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"a11ysentry/scanner/dotnet"
	"a11ysentry/scanner/pyqt"
	"a11ysentry/scanner/electron"
)

func TestDotNetScanner(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "app.csproj"), []byte("<Project></Project>"), 0644); err != nil {
		t.Fatal(err)
	}

	fw := dotnet.New()
	if !fw.Probe(tmp) {
		t.Fatal("DotNet scanner should probe positive for project with .csproj")
	}
}

func TestPyQtScanner(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "main.ui"), []byte("<ui></ui>"), 0644); err != nil {
		t.Fatal(err)
	}

	fw := pyqt.New()
	if !fw.Probe(tmp) {
		t.Fatal("PyQt scanner should probe positive for project with .ui files")
	}
}

func TestElectronScanner(t *testing.T) {
	tmp := t.TempDir()
	packageJSON := `{"devDependencies": {"electron": "^30.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	fw := electron.New()
	if !fw.Probe(tmp) {
		t.Fatal("Electron scanner should probe positive for project with electron dependency")
	}
}
