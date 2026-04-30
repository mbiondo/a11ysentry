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

// RegisterAll registers A11ySentry in all supported AI agents.
func RegisterAll(binaryPath string) []error {
	var errors []error

	// List of registration functions
	regFuncs := []func(string) error{
		RegisterClaude,
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

	// Check if source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("skill source directory not found at %s", sourceDir)
	}

	_ = os.MkdirAll(targetDir, 0755)

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
