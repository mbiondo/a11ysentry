package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8"))
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#00ADD8")).Padding(0, 1)
)

// Choice represents an option in a menu.
type Choice struct {
	Label       string
	Description string
	Selected    bool
}

// MultiSelectModel for interactive multi-choice menus.
type MultiSelectModel struct {
	Title    string
	Choices  []Choice
	Cursor   int
	Finished bool
}

func (m MultiSelectModel) Init() tea.Cmd { return nil }

func (m MultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case " ":
			m.Choices[m.Cursor].Selected = !m.Choices[m.Cursor].Selected
		case "enter":
			m.Finished = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MultiSelectModel) View() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(m.Title) + "\n\n")

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		checked := "[ ]"
		if choice.Selected {
			checked = "[x]"
		}

		line := fmt.Sprintf("%s %s %s", cursor, checked, choice.Label)
		if m.Cursor == i {
			s.WriteString(focusedStyle.Render(line))
		} else if choice.Selected {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(line)
		}
		
		if choice.Description != "" {
			s.WriteString(fmt.Sprintf(" - %s", choice.Description))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n(space to select, enter to confirm, q to quit)\n")
	return s.String()
}

// PromptYesNo for simple boolean questions.
type PromptYesNo struct {
	Question string
	Result   bool
	Finished bool
}

func (m PromptYesNo) Init() tea.Cmd { return nil }

func (m PromptYesNo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "y", "Y":
			m.Result = true
			m.Finished = true
			return m, tea.Quit
		case "n", "N":
			m.Result = false
			m.Finished = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PromptYesNo) View() string {
	return fmt.Sprintf("%s %s\n\n(y/n to select, q to quit)\n", titleStyle.Render("?"), m.Question)
}
