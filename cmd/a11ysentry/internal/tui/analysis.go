package tui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
	"fmt"
)

var (
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672")).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	
	codeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E1E")).
			Foreground(lipgloss.Color("#D4D4D4")).
			Padding(0, 1).
			Italic(true)
	
	violationStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginBottom(1)
	
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8")).
			Bold(true)
)

func (m MainModel) resultsView() string {
	var b strings.Builder

	if len(m.results.Violations) == 0 {
		return successStyle.Render("  ✅ No accessibility violations found! Excellent work.")
	}

	isNarrow := m.terminalW < 80
	contentWidth := m.terminalW - 12
	if isNarrow {
		contentWidth = m.terminalW - 6
	}

	errors := 0
	warnings := 0
	for _, v := range m.results.Violations {
		if v.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}

	summary := fmt.Sprintf("Found %d issues (%d errors, %d warnings)\n", len(m.results.Violations), errors, warnings)
	b.WriteString(lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render(summary))
	b.WriteString("\n")

	for _, v := range m.results.Violations {
		var severityTag string
		if v.Severity == "error" {
			severityTag = errorStyle.Render(" ERROR ")
		} else {
			severityTag = warningStyle.Render(" WARNING ")
		}

		var content strings.Builder
		fmt.Fprintf(&content, "%s %s\n\n", severityTag, labelStyle.Render(v.ErrorCode))
		
		msgStyle := lipgloss.NewStyle().Width(contentWidth - 4)
		content.WriteString(msgStyle.Render(v.Message) + "\n\n")
		
		// Hide source code in narrow mode to save space
		if !isNarrow && v.SourceRef.RawHTML != "" {
			content.WriteString(labelStyle.Render("SOURCE CODE:") + "\n")
			content.WriteString(codeStyle.Width(contentWidth-6).Render(v.SourceRef.RawHTML) + "\n\n")
		}

		if v.FixSnippet != "" {
			content.WriteString(labelStyle.Render("SUGGESTED FIX:") + "\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Width(contentWidth-6).Render(v.FixSnippet) + "\n\n")
		}

		if v.DocumentationURL != "" {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(contentWidth-6).Render("Learn more: "+v.DocumentationURL))
		}

		style := violationStyle.Width(contentWidth)
		if isNarrow {
			style = style.Padding(0, 1).MarginBottom(1)
		}

		b.WriteString(style.Render(content.String()))
		b.WriteString("\n")
	}

	return b.String()
}
