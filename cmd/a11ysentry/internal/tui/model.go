package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateProjects sessionState = iota
	stateProjectReports
	stateAnalyzing
	stateResults
)

type MainModel struct {
	state           sessionState
	repo            ports.Repository
	projectsList    list.Model
	reportsList     list.Model
	progress        progress.Model
	viewport        viewport.Model
	allReports      []domain.ViolationReport
	selectedProject string
	results         domain.ViolationReport
	terminalW       int
	terminalH       int
	totalToAnalyze  int
	analyzedCount   int
}

func NewMainModel(repo ports.Repository) MainModel {
	return MainModel{
		state:        stateProjects,
		repo:         repo,
		projectsList: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		reportsList:  list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		progress:     progress.New(progress.WithDefaultGradient()),
		viewport:     viewport.New(0, 0),
	}
}

type projectItem struct {
	name  string
	count int
}

func (i projectItem) Title() string       { return i.name }
func (i projectItem) Description() string { return fmt.Sprintf("%d analysis reports", i.count) }
func (i projectItem) FilterValue() string { return i.name }

type reportItem struct {
	report domain.ViolationReport
}

func (i reportItem) Title() string {
	return filepath.Base(i.report.FilePath)
}
func (i reportItem) Description() string {
	return fmt.Sprintf("%s - %s", i.report.Platform, time.Unix(i.report.Timestamp, 0).Format("2006-01-02 15:04"))
}
func (i reportItem) FilterValue() string { return i.report.FilePath }

func (m MainModel) fetchHistory() tea.Cmd {
	return func() tea.Msg {
		history, err := m.repo.GetHistory(context.Background(), 100)
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
		m.allReports = msg
		m.updateProjectsList()

	case tea.WindowSizeMsg:
		m.terminalW = msg.Width
		m.terminalH = msg.Height
		m.projectsList.SetSize(msg.Width-4, msg.Height-6)
		m.reportsList.SetSize(msg.Width-4, msg.Height-6)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		if m.state == stateResults {
			m.viewport.SetContent(m.resultsView())
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			switch m.state {
			case stateProjectReports:
				m.state = stateProjects
				return m, nil
			case stateResults:
				m.state = stateProjectReports
				return m, nil
			}
		case "enter":
			switch m.state {
			case stateProjects:
				if i, ok := m.projectsList.SelectedItem().(projectItem); ok {
					m.selectedProject = i.name
					m.updateReportsList()
					m.state = stateProjectReports
					return m, nil
				}
			case stateProjectReports:
				if i, ok := m.reportsList.SelectedItem().(reportItem); ok {
					m.results = i.report
					m.state = stateResults
					m.viewport.SetContent(m.resultsView())
					m.viewport.YOffset = 0
					return m, nil
				}
			}
		}
	}

	switch m.state {
	case stateProjects:
		m.projectsList, cmd = m.projectsList.Update(msg)
	case stateProjectReports:
		m.reportsList, cmd = m.reportsList.Update(msg)
	case stateResults:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m *MainModel) updateProjectsList() {
	projectCounts := make(map[string]int)
	var projectNames []string

	for _, r := range m.allReports {
		name := r.ProjectName
		if name == "" {
			name = "Unknown Project"
		}
		if _, exists := projectCounts[name]; !exists {
			projectNames = append(projectNames, name)
		}
		projectCounts[name]++
	}

	var items []list.Item
	for _, name := range projectNames {
		items = append(items, projectItem{name: name, count: projectCounts[name]})
	}
	m.projectsList.SetItems(items)
	m.projectsList.Title = "Projects"
}

func (m *MainModel) updateReportsList() {
	var items []list.Item
	for _, r := range m.allReports {
		if r.ProjectName == m.selectedProject || (m.selectedProject == "Unknown Project" && r.ProjectName == "") {
			items = append(items, reportItem{report: r})
		}
	}
	m.reportsList.SetItems(items)
	m.reportsList.Title = fmt.Sprintf("Reports: %s", m.selectedProject)
}

func (m MainModel) View() string {
	switch m.state {
	case stateProjects:
		return m.projectsView()
	case stateProjectReports:
		return m.reportsView()
	case stateAnalyzing:
		return m.progressView()
	case stateResults:
		header := headerStyle.Render(fmt.Sprintf("🛡️  Analysis Results: %s", m.results.FilePath))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 2).Render("press 'esc' to return • 'q' to quit • use arrows/pgup/pgdn to scroll")
		return docStyle.Render(header + "\n\n" + m.viewport.View() + "\n" + footer)
	default:
		return "Initializing..."
	}
}
