package nextjs

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProbeFixture(t *testing.T, dir string, dirs []string, files map[string]string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// Next.js supports a root app/ directory and a src/app/ directory. Only the root one was checked,
// so a project laid out under src/ fell through to the generic scanner: the layout chain was lost
// and every orphan component was analyzed as if it were a page.
func TestProbeDetectsSrcAppDirectory(t *testing.T) {
	dir := t.TempDir()
	writeProbeFixture(t, dir, []string{"src/app"}, map[string]string{"next.config.mjs": "export default {};"})

	if !New().Probe(dir) {
		t.Error("Probe() = false for a project that uses src/app")
	}
}

func TestProbeDetectsRootAppDirectory(t *testing.T) {
	dir := t.TempDir()
	writeProbeFixture(t, dir, []string{"app"}, map[string]string{"next.config.js": "module.exports = {};"})

	if !New().Probe(dir) {
		t.Error("Probe() = false for a project that uses app/")
	}
}

func TestProbeRejectsProjectWithoutNextConfig(t *testing.T) {
	dir := t.TempDir()
	writeProbeFixture(t, dir, []string{"src/app"}, nil)

	if New().Probe(dir) {
		t.Error("Probe() = true for a directory with no Next.js config")
	}
}

func TestProbeRejectsProjectWithoutAppDirectory(t *testing.T) {
	dir := t.TempDir()
	writeProbeFixture(t, dir, []string{"src/components"}, map[string]string{"next.config.ts": "export default {};"})

	if New().Probe(dir) {
		t.Error("Probe() = true for a project with no app directory")
	}
}
