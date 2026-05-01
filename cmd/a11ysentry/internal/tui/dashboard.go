package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#008080")).
			Bold(true).
			Padding(1, 2)
	
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

func (m MainModel) dashboardView() string {
	s := headerStyle.Render("🛡️  A11ySentry Dashboard")
	s += "\n\n"

	if len(m.history.Items()) == 0 {
		s += "No analysis history found. Run your first audit!\n\n"
	} else {
		s += m.history.View()
		s += "\n"
	}

	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("press 'q' to quit • 'enter' to view details")

	return docStyle.Render(s)
}
