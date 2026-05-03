package registration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

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

const a11ysentrySkillContent = `---
name: a11ysentry-mcp
description: >
  Expert guidance for auditing and fixing accessibility violations using the A11ySentry MCP server.
  Triggers: "check accessibility", "audit UI", "fix a11y", "WCAG compliance", or UI changes in PRs.
allowed-tools: [analyze_accessibility, read_file, glob, grep_search]
version: "1.2.0"
---

# A11ySentry Expert Skill

You are an accessibility expert. Your goal is to ensure the project complies with WCAG 2.2 standards using the ` + "`" + `a11ysentry` + "`" + ` MCP tool.

## 🧠 Strategic Instructions

### 1. The Recursive Context Rule (CRITICAL)
Accessibility context is often fragmented. An ` + "`" + `<input>` + "`" + ` in one file might be labeled by a ` + "`" + `<label>` + "`" + ` in a parent or sibling.
- **NEVER** analyze a single component file in isolation if it belongs to a tree.
- **ACTION**: Use ` + "`" + `grep_search` + "`" + ` or ` + "`" + `read_file` + "`" + ` to find where the component is used.
- **ACTION**: Pass ALL related files to ` + "`" + `analyze_accessibility` + "`" + ` as a comma-separated list.

### 2. Multi-Platform Awareness
A11ySentry uses specific adapters. When you see these files, apply the corresponding mindset:
- ` + "`" + `.astro` + "`" + `, ` + "`" + `.vue` + "`" + `, ` + "`" + `.svelte` + "`" + `, ` + "`" + `.tsx` + "`" + `: Web/HTML standards.
- ` + "`" + `.kt` + "`" + `, ` + "`" + `.xml` + "`" + `: Android (Compose/Views).
- ` + "`" + `.swift` + "`" + `: iOS (SwiftUI).
- ` + "`" + `.dart` + "`" + `: Flutter.

### 3. Fixing Protocol
When violations are found:
1. **Analyze**: Read the ` + "`" + `Message` + "`" + ` and ` + "`" + `ErrorCode` + "`" + ` (e.g., WCAG 1.1.1).
2. **Consult**: Check the ` + "`" + `DocumentationURL` + "`" + ` if the fix isn't obvious.
3. **Execute**: Use the ` + "`" + `FixSnippet` + "`" + ` as a base, but ensure it matches the project's styling/component patterns.
4. **Verify**: ALWAYS re-run ` + "`" + `analyze_accessibility` + "`" + ` after making a change to confirm the fix.

## 🛠 Available Tools (via MCP)

- ` + "`" + `analyze_accessibility(path: string)` + "`" + `: Audits files or directories. ` + "`" + `path` + "`" + ` can be a single file path, a directory path, or a comma-separated list of files.
  - **Directory Mode**: If a directory is provided, A11ySentry automatically discovers project roots (e.g., Next.js, Astro, Android) and resolves full component trees for analysis.
  - **File Mode**: Audits specific files. Use comma-separated paths to provide multi-file context (Recursive Context Rule).
  **Note**: The output is in **TOON (Token-Oriented Object Notation)** for efficiency.
  **Format**: ` + "`" + `violations[count]{code,sev,file,line,snippet,msg,fix}:` + "`" + `
  - ` + "`" + `code` + "`" + `: WCAG Error Code.
  - ` + "`" + `sev` + "`" + `: Severity (E = Error, W = Warning).
  - ` + "`" + `snippet` + "`" + `: Truncated HTML/Code fragment.
  - ` + "`" + `msg` + "`" + `: Description of the issue.
  - ` + "`" + `fix` + "`" + `: Suggested fix.

## 📋 Examples

### Auditing a Full Project
If the user asks to check the entire project:
1. Identify the project root directory.
2. Run: ` + "`" + `analyze_accessibility(".")` + "`" + ` or ` + "`" + `analyze_accessibility("src/")` + "`" + `.

### Auditing a Page (Legacy/Specific)
If the user asks to check ` + "`" + `src/pages/index.astro` + "`" + ` and you want to be specific:
1. Identify components used in ` + "`" + `index.astro` + "`" + ` (e.g., ` + "`" + `Header.astro` + "`" + `, ` + "`" + `Footer.astro` + "`" + `).
2. Run: ` + "`" + `analyze_accessibility("src/pages/index.astro, src/components/Header.astro, src/components/Footer.astro")` + "`" + `.


### Fixing a contrast issue
1. Tool reports ` + "`" + `WCAG_1_4_3` + "`" + ` (Contrast) in ` + "`" + `Button.tsx` + "`" + `.
2. Look for CSS variables or Tailwind classes defining the color.
3. Apply fix and re-run analysis.
`

// RegisterAll registers A11ySentry in all supported AI agents.
func RegisterAll(binaryPath string) []error {
	var errors []error

	// List of registration functions
	regFuncs := []func(string) error{
		RegisterClaude,
		RegisterClaudeCode,
		RegisterCursor,
		RegisterGemini,
		RegisterVSCode,
		RegisterQwen,
		RegisterOpenCode,
	}

	for _, f := range regFuncs {
		if err := f(binaryPath); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
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

	var config MCPConfig
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &config)
	}

	if config.MCP == nil {
		config.MCP = make(map[string]OpenCodeServerConfig)
	}

	config.MCP["a11ysentry"] = OpenCodeServerConfig{
		Type:    "local",
		Command: []string{binaryPath, "mcp"},
		Enabled: true,
	}

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		_ = os.WriteFile(path+".bak", data, 0644)
	}

	return os.WriteFile(path, newData, 0644)
}

// RegisterSkill installs the A11ySentry skill in the global Gemini CLI skills directory.
func RegisterSkill(repoRoot string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	targetDir := filepath.Join(homeDir, ".gemini", "skills", "a11ysentry-mcp")
	sourceDir := filepath.Join(repoRoot, "skills", "a11ysentry-mcp")

	_ = os.MkdirAll(targetDir, 0755)

	// Check if source exists (local dev mode)
	if _, err := os.Stat(sourceDir); err == nil {
		// Simple recursive copy for the skill files (SKILL.md, etc.)
		return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return err
			}

			dest := filepath.Join(targetDir, rel)
			if info.IsDir() {
				return os.MkdirAll(dest, info.Mode())
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			return os.WriteFile(dest, data, info.Mode())
		})
	}

	// Fallback: Use bundled content (Homebrew / single-binary mode)
	skillPath := filepath.Join(targetDir, "SKILL.md")
	return os.WriteFile(skillPath, []byte(a11ysentrySkillContent), 0644)
}

func patchConfig(configPath, binaryPath string) error {
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)

	var config MCPConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		_ = json.Unmarshal(data, &config)
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServerConfig)
	}

	config.MCPServers["a11ysentry"] = MCPServerConfig{
		Command: binaryPath,
		Args:    []string{"mcp"},
	}

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
