package registration

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// skillFS is the single source of truth for the agent-facing skill. It is embedded so the
// compiled binary installs exactly the bytes that live in this repository, and so the two can
// never drift apart again.
//
//go:embed skills/a11ysentry-mcp
var skillFS embed.FS

// skillRoot is the directory of the embedded skill tree.
const skillRoot = "skills/a11ysentry-mcp"

// MCPConfig is the union of the agent configuration shapes A11ySentry writes.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers,omitempty"`
	// OpenCode specific root key
	MCP map[string]OpenCodeServerConfig `json:"mcp,omitempty"`
}

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type OpenCodeServerConfig struct {
	Type    string   `json:"type"`
	Command []string `json:"command"`
	Enabled bool     `json:"enabled"`
}

// SkillTarget is a directory an agent runtime scans for skills.
type SkillTarget struct {
	// Label names the runtime that reads the directory.
	Label string
	// Path is the skills directory to install into.
	Path string
}

// RegisterAll registers A11ySentry in every supported AI agent.
func RegisterAll(binaryPath string) []error {
	var errs []error

	// List of registration functions
	regFuncs := []func(string) error{
		RegisterClaude,
		RegisterClaudeCode,
		RegisterCursor,
		RegisterGemini,
		RegisterVSCode,
		RegisterQwen,
		RegisterOpenCode,
		RegisterPi,
	}

	for _, f := range regFuncs {
		if err := f(binaryPath); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func RegisterClaude(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Claude", "claude_desktop_config.json")
	}
	return patchConfig(path, binaryPath)
}

func RegisterClaudeCode(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("USERPROFILE"), ".claude.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".claude.json")
	}
	return patchConfig(path, binaryPath)
}

func RegisterCursor(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "globalStorage", "mcp-servers.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".config", "Cursor", "User", "globalStorage", "mcp-servers.json")
	}
	return patchConfig(path, binaryPath)
}

func RegisterGemini(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("APPDATA"), "gemini", "settings.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".gemini", "settings.json")
	}
	return patchConfig(path, binaryPath)
}

func RegisterVSCode(binaryPath string) error {
	var paths []string
	if runtime.GOOS == "windows" {
		userDir := filepath.Join(os.Getenv("APPDATA"), "Code", "User")
		paths = []string{
			filepath.Join(userDir, "mcp.json"),
			filepath.Join(userDir, "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"),
		}
	} else {
		userDir := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Code", "User")
		if runtime.GOOS == "linux" {
			userDir = filepath.Join(os.Getenv("HOME"), ".config", "Code", "User")
		}
		paths = []string{
			filepath.Join(userDir, "mcp.json"),
			filepath.Join(userDir, "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"),
		}
	}

	for _, p := range paths {
		_ = patchConfig(p, binaryPath)
	}
	return nil
}

func RegisterQwen(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("USERPROFILE"), ".qwen", "settings.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".qwen", "settings.json")
	}
	return patchConfig(path, binaryPath)
}

func RegisterOpenCode(binaryPath string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(os.Getenv("USERPROFILE"), ".config", "opencode", "opencode.json")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".config", "opencode", "opencode.json")
	}

	_ = os.MkdirAll(filepath.Dir(path), 0755)

	var config map[string]interface{}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	} else {
		config = make(map[string]interface{})
	}

	mcpRaw, ok := config["mcp"]
	var mcp map[string]interface{}
	if ok {
		if m, ok := mcpRaw.(map[string]interface{}); ok {
			mcp = m
		} else {
			mcp = make(map[string]interface{})
		}
	} else {
		mcp = make(map[string]interface{})
	}

	mcp["a11ysentry"] = map[string]interface{}{
		"type":    "local",
		"command": []string{binaryPath, "mcp"},
		"enabled": true,
	}

	config["mcp"] = mcp

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		_ = os.WriteFile(path+".bak", data, 0644)
	}

	return os.WriteFile(path, newData, 0644)
}

// piMCPConfigPath is the location the pi agent runtime reads MCP servers from. It is the same
// on every platform, but depends on the home directory supplied.
func piMCPConfigPath(home string) string {
	return filepath.Join(home, ".pi", "agent", "mcp.json")
}

// RegisterPi registers A11ySentry in the pi agent runtime. Pi reads the standard
// {"mcpServers": {...}} shape, so it shares the generic patcher with the other agents.
func RegisterPi(binaryPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return registerPiAt(home, binaryPath)
}

// registerPiAt is RegisterPi against an explicit home directory, so the written path is
// testable without touching the real user profile.
func registerPiAt(home, binaryPath string) error {
	return patchConfig(piMCPConfigPath(home), binaryPath)
}

// AgentRegistration pairs a user-facing label with the function that registers A11ySentry in that
// agent runtime.
type AgentRegistration struct {
	Label    string
	Register func(binaryPath string) error
}

// agentProbe is an AgentRegistration plus the paths that show the runtime is installed.
type agentProbe struct {
	label   string
	markers []string
	reg     func(string) error
}

// DetectedAgents returns the agent runtimes present on this machine, in menu order.
//
// Registration skips a runtime whose configuration directory does not exist, so installing
// A11ySentry never leaves configuration behind for an agent the user does not have.
func DetectedAgents() []AgentRegistration {
	home, _ := os.UserHomeDir()

	probes := agentProbes(home)

	var detected []AgentRegistration
	for _, p := range probes {
		for _, marker := range p.markers {
			if _, err := os.Stat(marker); err == nil {
				detected = append(detected, AgentRegistration{Label: p.label, Register: p.reg})
				break
			}
		}
	}
	return detected
}

// agentProbes lists every supported runtime with the markers that reveal it. The markers mirror
// the config locations the Register* functions write to.
func agentProbes(home string) []agentProbe {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		return []agentProbe{
			{"Claude Desktop", []string{filepath.Join(appData, "Claude")}, RegisterClaude},
			{"Claude Code (CLI)", []string{filepath.Join(home, ".claude.json")}, RegisterClaudeCode},
			{"Cursor IDE", []string{filepath.Join(appData, "Cursor")}, RegisterCursor},
			{"VSCode (Cline/Roo-Code)", []string{filepath.Join(appData, "Code")}, RegisterVSCode},
			{"Gemini CLI", []string{filepath.Join(appData, "gemini"), filepath.Join(home, ".gemini")}, RegisterGemini},
			{"Qwen", []string{filepath.Join(home, ".qwen")}, RegisterQwen},
			{"OpenCode", []string{filepath.Join(home, ".config", "opencode")}, RegisterOpenCode},
			{"Pi", []string{filepath.Join(home, ".pi", "agent")}, RegisterPi},
		}
	}

	return []agentProbe{
		{"Claude Desktop", []string{filepath.Join(home, "Library", "Application Support", "Claude")}, RegisterClaude},
		{"Claude Code (CLI)", []string{filepath.Join(home, ".claude.json")}, RegisterClaudeCode},
		{"Cursor IDE", []string{filepath.Join(home, "Library", "Application Support", "Cursor"), filepath.Join(home, ".config", "Cursor")}, RegisterCursor},
		{"VSCode (Cline/Roo-Code)", []string{filepath.Join(home, "Library", "Application Support", "Code"), filepath.Join(home, ".config", "Code")}, RegisterVSCode},
		{"Gemini CLI", []string{filepath.Join(home, ".gemini")}, RegisterGemini},
		{"Qwen", []string{filepath.Join(home, ".qwen")}, RegisterQwen},
		{"OpenCode", []string{filepath.Join(home, ".config", "opencode")}, RegisterOpenCode},
		{"Pi", []string{filepath.Join(home, ".pi", "agent")}, RegisterPi},
	}
}

// RegisterDetected registers A11ySentry in every agent runtime present on this machine. It
// returns the labels it registered in, and one error per runtime that failed, so a single
// failure does not stop the others.
func RegisterDetected(binaryPath string) (registered []string, errs []error) {
	for _, agent := range DetectedAgents() {
		if err := agent.Register(binaryPath); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", agent.Label, err))
			continue
		}
		registered = append(registered, agent.Label)
	}
	return registered, errs
}

// GlobalSkillTargets returns every global skills directory A11ySentry installs into, covering
// the pi runtime, the cross-runtime agents convention, Gemini CLI, and OpenCode.
func GlobalSkillTargets(home string) []SkillTarget {
	return []SkillTarget{
		{Label: "pi", Path: filepath.Join(home, ".pi", "agent", "skills")},
		{Label: "Shared agents", Path: filepath.Join(home, ".config", "agents", "skills")},
		{Label: "Shared agents", Path: filepath.Join(home, ".agents", "skills")},
		{Label: "Gemini CLI", Path: filepath.Join(home, ".gemini", "skills")},
		{Label: "OpenCode", Path: filepath.Join(home, ".config", "opencode", "skills")},
	}
}

// ProjectSkillTargets returns the project-level skills directories inside repoRoot, following
// the same discovery convention.
func ProjectSkillTargets(repoRoot string) []SkillTarget {
	return []SkillTarget{
		{Label: "pi", Path: filepath.Join(repoRoot, ".pi", "skills")},
		{Label: "Shared agents", Path: filepath.Join(repoRoot, ".agent", "skills")},
		{Label: "Shared agents", Path: filepath.Join(repoRoot, ".agents", "skills")},
	}
}

// SkillWriteError reports a failure installing the skill into one directory while letting the
// remaining directories still be attempted.
type SkillWriteError struct {
	Label string
	Path  string
	Err   error
}

func (e *SkillWriteError) Error() string {
	return fmt.Sprintf("%s (%s): %v", e.Label, e.Path, e.Err)
}

func (e *SkillWriteError) Unwrap() error { return e.Err }

// writeSkill copies the embedded skill tree into dir under the a11ysentry-mcp subdirectory.
func writeSkill(dir string) error {
	dest := filepath.Join(dir, "a11ysentry-mcp")
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	return fs.WalkDir(skillFS, skillRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(skillRoot, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		data, err := skillFS.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// RegisterSkills installs the agent skill into every global agent skills directory, and, when
// repoRoot is non-empty, into that project's skill directories as well.
//
// Pass an empty repoRoot to install only the global directories, which is what a standalone
// installer run wants.
func RegisterSkills(repoRoot string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	targets := GlobalSkillTargets(home)
	if repoRoot != "" {
		targets = append(targets, ProjectSkillTargets(repoRoot)...)
	}

	var errs []error
	for _, t := range targets {
		if err := writeSkill(t.Path); err != nil {
			errs = append(errs, &SkillWriteError{Label: t.Label, Path: t.Path, Err: err})
		}
	}
	return errors.Join(errs...)
}

func patchConfig(configPath, binaryPath string) error {
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)

	var config map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	} else {
		config = make(map[string]interface{})
	}

	mcpServersRaw, ok := config["mcpServers"]
	var mcpServers map[string]interface{}
	if ok {
		if m, ok := mcpServersRaw.(map[string]interface{}); ok {
			mcpServers = m
		} else {
			mcpServers = make(map[string]interface{})
		}
	} else {
		mcpServers = make(map[string]interface{})
	}

	mcpServers["a11ysentry"] = map[string]interface{}{
		"command": binaryPath,
		"args":    []string{"mcp"},
	}

	config["mcpServers"] = mcpServers

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Backup
	if _, err := os.Stat(configPath); err == nil {
		_ = os.WriteFile(configPath+".bak", data, 0644)
	}

	return os.WriteFile(configPath, newData, 0644)
}

func CheckRegistration(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}
	_, ok := config.MCPServers["a11ysentry"]
	if !ok {
		_, ok = config.MCP["a11ysentry"]
	}
	return ok
}
