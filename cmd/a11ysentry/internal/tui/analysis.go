package tui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
	"fmt"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	codeStyle  = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a")).
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1)
	
	violationStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderLeftForeground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			Margin(1, 0)
)

func (m MainModel) resultsView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("🛡️  Analysis Results: %s", m.results.FilePath)))
	b.WriteString("\n\n")

	if len(m.results.Violations) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("✅ No accessibility violations found!"))
	} else {
		fmt.Fprintf(&b, "Found %d violations:\n", len(m.results.Violations))
		for _, v := range m.results.Violations {
			b.WriteString(violationStyle.Render(
				fmt.Sprintf("%s [%s]\n%s\n\n%s",
					errorStyle.Render(v.ErrorCode),
					v.SourceRef.Platform,
					v.Message,
					codeStyle.Render(v.SourceRef.RawHTML),
				),
			))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("press 'esc' to return to dashboard • 'q' to quit"))

	return docStyle.Render(b.String())
}
