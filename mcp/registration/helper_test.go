package registration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readFile is a small helper so each test can fail with a precise message.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func TestPiMCPConfigPath(t *testing.T) {
	got := piMCPConfigPath(`C:\Users\dev`)

	want := filepath.Join(`C:\Users\dev`, ".pi", "agent", "mcp.json")
	if got != want {
		t.Errorf("piMCPConfigPath() = %q, want %q", got, want)
	}
}

func TestRegisterPiWritesMCPConfig(t *testing.T) {
	home := t.TempDir()

	if err := registerPiAt(home, `C:\bin\a11ysentry.exe`); err != nil {
		t.Fatalf("registerPiAt() error = %v", err)
	}

	path := piMCPConfigPath(home)
	raw := readFile(t, path)

	var cfg MCPConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("written config is not valid JSON: %v\n%s", err, raw)
	}

	server, ok := cfg.MCPServers["a11ysentry"]
	if !ok {
		t.Fatalf("a11ysentry missing from mcpServers: %s", raw)
	}
	if server.Command != `C:\bin\a11ysentry.exe` {
		t.Errorf("command = %q, want the registered binary path", server.Command)
	}
	if len(server.Args) != 1 || server.Args[0] != "mcp" {
		t.Errorf("args = %v, want [mcp]", server.Args)
	}
}

func TestRegisterPiPreservesExistingServers(t *testing.T) {
	home := t.TempDir()
	path := piMCPConfigPath(home)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{"mcpServers":{"codegraph":{"command":"codegraph","args":["serve","--mcp"]}}}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := registerPiAt(home, "a11ysentry"); err != nil {
		t.Fatalf("registerPiAt() error = %v", err)
	}

	var cfg MCPConfig
	if err := json.Unmarshal([]byte(readFile(t, path)), &cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := cfg.MCPServers["codegraph"]; !ok {
		t.Error("registering a11ysentry dropped the pre-existing codegraph server")
	}
	if _, ok := cfg.MCPServers["a11ysentry"]; !ok {
		t.Error("a11ysentry was not added alongside the existing server")
	}
}

func TestRegisterPiIsIdempotent(t *testing.T) {
	home := t.TempDir()

	if err := registerPiAt(home, "a11ysentry"); err != nil {
		t.Fatalf("first registerPiAt() error = %v", err)
	}
	first := readFile(t, piMCPConfigPath(home))

	if err := registerPiAt(home, "a11ysentry"); err != nil {
		t.Fatalf("second registerPiAt() error = %v", err)
	}
	second := readFile(t, piMCPConfigPath(home))

	if first != second {
		t.Errorf("registration is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestGlobalSkillTargetsCoverAgentRuntimes(t *testing.T) {
	targets := GlobalSkillTargets(`C:\Users\dev`)

	var sawPi bool
	for _, target := range targets {
		if strings.Contains(target.Label, "pi") && strings.Contains(target.Path, ".pi") {
			sawPi = true
		}
	}

	if !sawPi {
		t.Errorf("no pi skill target in %v", targets)
	}
}

func TestProjectSkillTargetsFollowDiscoveryConvention(t *testing.T) {
	targets := ProjectSkillTargets(`C:\work\app`)

	want := []string{
		filepath.Join(`C:\work\app`, ".pi", "skills"),
		filepath.Join(`C:\work\app`, ".agent", "skills"),
		filepath.Join(`C:\work\app`, ".agents", "skills"),
	}

	if len(targets) != len(want) {
		t.Fatalf("got %d targets, want %d: %v", len(targets), len(want), targets)
	}
	for i, w := range want {
		if targets[i].Path != w {
			t.Errorf("target[%d].Path = %q, want %q", i, targets[i].Path, w)
		}
	}
}

func TestSkillIsEmbedded(t *testing.T) {
	data, err := skillFS.ReadFile(skillRoot + "/SKILL.md")
	if err != nil {
		t.Fatalf("embedded skill is unreadable: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("embedded SKILL.md is missing its frontmatter delimiter:\n%s", content[:80])
	}

	// The skill must document every tool the MCP server actually exposes.
	for _, tool := range []string{"analyze_accessibility", "get_component_context", "get_audit_history"} {
		if !strings.Contains(content, "### `"+tool) {
			t.Errorf("embedded SKILL.md does not document the %s tool", tool)
		}
	}
}

func TestWriteSkillInstallsIntoDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")

	if err := writeSkill(dir); err != nil {
		t.Fatalf("writeSkill() error = %v", err)
	}

	installed := filepath.Join(dir, "a11ysentry-mcp", "SKILL.md")
	content := readFile(t, installed)

	if content != string(mustSkillBytes(t)) {
		t.Error("installed SKILL.md does not match the embedded source")
	}
}

func mustSkillBytes(t *testing.T) []byte {
	t.Helper()
	data, err := skillFS.ReadFile(skillRoot + "/SKILL.md")
	if err != nil {
		t.Fatalf("reading embedded skill: %v", err)
	}
	return data
}
