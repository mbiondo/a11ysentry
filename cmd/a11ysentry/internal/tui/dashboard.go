package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8")).
			Bold(true).
			Padding(1, 2)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginRight(1)

	mainContentStyle = lipgloss.NewStyle().Padding(0, 1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

func (m MainModel) sidebarView() string {
	// Simple sidebar showing the current project if any, and some stats
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("PROJECTS"))
	s.WriteString("\n\n")

	// Highlight current selection in sidebar context if we are in reports view
	for _, item := range m.projectsList.Items() {
		p := item.(projectItem)
		if m.state != stateProjects && p.name == m.selectedProject {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Render("> " + p.name))
		} else {
			s.WriteString("  " + p.name)
		}
		s.WriteString("\n")
	}

	return sidebarStyle.Render(s.String())
}

func (m MainModel) projectsView() string {
	header := headerStyle.Render("🛡️  A11ySentry Dashboard")
	
	// Create a two-column layout: Sidebar with project names, and Main with the interactive list
	sidebar := m.sidebarView()
	content := m.projectsList.View()

	combined := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingTop(1).Render("press 'q' to quit • 'enter' to view project reports")

	return docStyle.Render(header + "\n\n" + combined + "\n" + footer)
}

func (m MainModel) reportsView() string {
	header := headerStyle.Render(fmt.Sprintf("📂 Project: %s", m.selectedProject))
	
	sidebar := m.sidebarView()
	content := m.reportsList.View()

	combined := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingTop(1).Render("press 'esc' to back • 'enter' to view details")

	return docStyle.Render(header + "\n\n" + combined + "\n" + footer)
}

func (m MainModel) progressView() string {
	header := headerStyle.Render("⏳ Analyzing Files...")
	
	bar := m.progress.ViewAs(float64(m.analyzedCount) / float64(m.totalToAnalyze))
	
	stats := fmt.Sprintf("\n\nProgress: %d / %d files\n", m.analyzedCount, m.totalToAnalyze)
	
	return docStyle.Render(header + "\n\n" + bar + stats)
}
