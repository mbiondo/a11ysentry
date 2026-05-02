package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8")).
			Bold(true).
			Padding(1, 2)

	mainStyle = lipgloss.NewStyle().
			Padding(0, 2)

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

func (m MainModel) projectsView() string {
	header := headerStyle.Render("🛡️  A11ySentry Dashboard")

	m.projectsList.SetSize(m.terminalW-4, m.terminalH-8)
	mainView := mainStyle.Width(m.terminalW - 4).Render(m.projectsList.View())

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingTop(1).Render("q: quit • enter: view reports")

	return docStyle.Render(header + "\n\n" + mainView + "\n" + footer)
}

func (m MainModel) reportsView() string {
	header := headerStyle.Render(fmt.Sprintf("📂 Project: %s", m.selectedProject))

	m.reportsList.SetSize(m.terminalW-4, m.terminalH-8)
	mainView := mainStyle.Width(m.terminalW - 4).Render(m.reportsList.View())

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingTop(1).Render("esc: back • enter: details")

	return docStyle.Render(header + "\n\n" + mainView + "\n" + footer)
}

func (m MainModel) progressView() string {
	header := headerStyle.Render("⏳ Analyzing Files...")
	
	bar := m.progress.ViewAs(float64(m.analyzedCount) / float64(m.totalToAnalyze))
	
	stats := fmt.Sprintf("\n\nProgress: %d / %d files\n", m.analyzedCount, m.totalToAnalyze)
	
	return docStyle.Render(header + "\n\n" + bar + stats)
}
