// Package sarif converts A11ySentry violation reports to SARIF 2.1.0 format,
// which is consumed by GitHub Code Scanning and other CI/CD tools.
package sarif

import (
	"a11ysentry/engine/core/domain"
	"fmt"
	"strings"
)

// ── SARIF 2.1.0 structs (subset required for GitHub Code Scanning) ──────────

type Log struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	InformationURI string `json:"informationUri"`
	Rules          []Rule `json:"rules"`
}

type Rule struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	ShortDescription Message        `json:"shortDescription"`
	HelpURI          string         `json:"helpUri,omitempty"`
	Properties       RuleProperties `json:"properties"`
}

type RuleProperties struct {
	Tags     []string `json:"tags"`
	Severity string   `json:"severity"`
}

type Message struct {
	Text string `json:"text"`
}

type Result struct {
	RuleID    string     `json:"ruleId"`
	Level     string     `json:"level"`
	Message   Message    `json:"message"`
	Locations []Location `json:"locations"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           *Region          `json:"region,omitempty"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
}

// FromReports converts a slice of ViolationReports to a SARIF Log.
func FromReports(reports []domain.ViolationReport) Log {
	// Collect unique rules across all reports.
	ruleMap := make(map[string]Rule)
	var results []Result

	for _, rep := range reports {
		for _, v := range rep.Violations {
			// Register rule if not already present.
			if _, exists := ruleMap[v.ErrorCode]; !exists {
				severity := "error"
				if v.Severity == domain.SeverityWarning {
					severity = "warning"
				}
				ruleMap[v.ErrorCode] = Rule{
					ID:   v.ErrorCode,
					Name: v.ErrorCode,
					ShortDescription: Message{
						Text: v.Message,
					},
					HelpURI: v.DocumentationURL,
					Properties: RuleProperties{
						Tags:     []string{"accessibility", "wcag"},
						Severity: severity,
					},
				}
			}

			level := "error"
			if v.Severity == domain.SeverityWarning {
				level = "warning"
			}

			msg := v.Message
			if v.SourceRef.RawHTML != "" {
				msg = fmt.Sprintf("%s | Code: %s", v.Message, v.SourceRef.RawHTML)
			}

			loc := Location{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{URI: fileURI(v.SourceRef.FilePath)},
				},
			}
			if v.SourceRef.Line > 0 {
				loc.PhysicalLocation.Region = &Region{
					StartLine:   v.SourceRef.Line,
					StartColumn: v.SourceRef.Column,
				}
			}

			results = append(results, Result{
				RuleID:    v.ErrorCode,
				Level:     level,
				Message:   Message{Text: msg},
				Locations: []Location{loc},
			})
		}
	}

	// Convert rule map to ordered slice.
	rules := make([]Rule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	return Log{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:           "A11ySentry",
						Version:        "1.0.0",
						InformationURI: "https://github.com/anomalyco/a11ysentry",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}
}

// fileURI converts an absolute path to a file:// URI (SARIF expects URIs).
func fileURI(path string) string {
	if path == "" {
		return ""
	}
	// Normalize Windows backslashes.
	path = replacer.Replace(path)
	if len(path) > 0 && path[0] != '/' {
		return "file:///" + path
	}
	return "file://" + path
}

var replacer = strings.NewReplacer(`\`, `/`)
