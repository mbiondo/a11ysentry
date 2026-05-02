package domain

import (
	"encoding/json"
	"os"
)

// RuleConfig defines the behavior of a specific accessibility rule.
type RuleConfig struct {
	Enabled  bool     `json:"enabled"`
	Severity Severity `json:"severity"`
}

// ProjectConfig holds the full configuration for an a11ysentry project.
type ProjectConfig struct {
	Version       string                `json:"version"`
	Paths         []string              `json:"paths"`
	Exclude       []string              `json:"exclude"`
	Format        string                `json:"format"`
	ExitOnWarning bool                  `json:"exitOnWarning"`
	Platform      string                `json:"platform"`
	Rules         map[string]RuleConfig `json:"rules"`
}

// DefaultConfig returns a configuration with all rules enabled by default.
func DefaultConfig() ProjectConfig {
	return ProjectConfig{
		Version: "1.0",
		Paths:   []string{"src", "app", "pages", "components"},
		Exclude: []string{"node_modules", ".git", "examples", "landing"},
		Format:  "text",
		Rules: map[string]RuleConfig{
			"WCAG_1_1_1":        {Enabled: true, Severity: SeverityError},
			"WCAG_1_3_1":        {Enabled: true, Severity: SeverityError},
			"WCAG_1_3_1_LEGEND": {Enabled: true, Severity: SeverityError},
			"WCAG_1_3_5":        {Enabled: true, Severity: SeverityWarning},
			"WCAG_1_4_1":        {Enabled: true, Severity: SeverityWarning},
			"WCAG_1_4_3":        {Enabled: true, Severity: SeverityError},
			"WCAG_1_4_3_DARK":   {Enabled: true, Severity: SeverityError},
			"WCAG_1_4_11":       {Enabled: true, Severity: SeverityError},
			"WCAG_2_1_1":        {Enabled: true, Severity: SeverityError},
			"WCAG_2_4_1":        {Enabled: true, Severity: SeverityError},
			"WCAG_2_4_3":        {Enabled: true, Severity: SeverityWarning},
			"WCAG_2_4_3_HIDDEN": {Enabled: true, Severity: SeverityError},
			"WCAG_2_4_4":        {Enabled: true, Severity: SeverityError},
			"WCAG_2_4_6":        {Enabled: true, Severity: SeverityWarning},
			"WCAG_2_4_7":        {Enabled: true, Severity: SeverityError},
			"WCAG_2_5_5":        {Enabled: true, Severity: SeverityError},
			"WCAG_2_5_8":        {Enabled: true, Severity: SeverityError},
			"WCAG_3_1_1":        {Enabled: true, Severity: SeverityWarning},
			"WCAG_3_3_2":        {Enabled: true, Severity: SeverityError},
			"WCAG_4_1_1":        {Enabled: true, Severity: SeverityError},
			"WCAG_4_1_2":        {Enabled: true, Severity: SeverityError},
			"WCAG_4_1_3":        {Enabled: true, Severity: SeverityWarning},
			"ARIA_1_1":          {Enabled: true, Severity: SeverityWarning},
			"G141":              {Enabled: true, Severity: SeverityWarning},
			"REDUNDANT_TITLE":   {Enabled: true, Severity: SeverityWarning},
			"ADV_FOCUS_TRAP":    {Enabled: true, Severity: SeverityError},
		},
	}
}

// LoadConfig reads a JSON configuration file and merges it with defaults.
func LoadConfig(path string) (ProjectConfig, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}
