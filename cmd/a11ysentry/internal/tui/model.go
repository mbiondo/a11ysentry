package tui

import (
	"context"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	stateDashboard sessionState = iota
	stateAnalyzing
	stateResults
)

type MainModel struct {
	state      sessionState
	repo       ports.Repository
	history    list.Model
	results    domain.ViolationReport
	terminalW  int
	terminalH  int
}

func NewMainModel(repo ports.Repository) MainModel {
	return MainModel{
		state:   stateDashboard,
		repo:    repo,
		history: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

type reportItem struct {
	report domain.ViolationReport
}

func (i reportItem) Title() string       { return i.report.FilePath }
func (i reportItem) Description() string { return string(i.report.Platform) }
func (i reportItem) FilterValue() string { return i.report.FilePath }

func (m MainModel) fetchHistory() tea.Cmd {
	return func() tea.Msg {
		history, err := m.repo.GetHistory(context.Background(), 20)
		if err != nil {
			return err
		}
		return history
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.fetchHistory()
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []domain.ViolationReport:
		var items []list.Item
		for _, r := range msg {
			items = append(items, reportItem{report: r})
		}
		m.history.SetItems(items)
		m.history.Title = "Recent Audits"

	case tea.WindowSizeMsg:
		m.terminalW = msg.Width
		m.terminalH = msg.Height
		m.history.SetSize(msg.Width-4, msg.Height-10)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state != stateDashboard {
				m.state = stateDashboard
				return m, nil
			}
		case "enter":
			if m.state == stateDashboard {
				if i, ok := m.history.SelectedItem().(reportItem); ok {
					m.results = i.report
					m.state = stateResults
				}
			}
		}
	}

	if m.state == stateDashboard {
		m.history, cmd = m.history.Update(msg)
	}

	return m, cmd
}

func (m MainModel) View() string {
	switch m.state {
	case stateDashboard:
		return m.dashboardView()
	case stateResults:
		return m.resultsView()
	default:
		return "Initializing..."
	}
}
