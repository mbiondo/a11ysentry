package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"a11ysentry/scanner/django"
	"a11ysentry/scanner/flask"
)

func TestDjangoScanner(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "manage.py"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	fw := django.New()
	if !fw.Probe(tmp) {
		t.Fatal("Django scanner should probe positive for project with manage.py")
	}
}

func TestFlaskScanner(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "app.py"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	fw := flask.New()
	if !fw.Probe(tmp) {
		t.Fatal("Flask scanner should probe positive for project with app.py")
	}
}
