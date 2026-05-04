package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
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
	stateProjectTree
	stateAnalyzing
	stateResults
)

type MainModel struct {
	state           sessionState
	repo            ports.Repository
	projectsList    list.Model
	reportsList     list.Model
	treeView        list.Model // Reusing list model for the tree view
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
		treeView:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		progress:     progress.New(progress.WithDefaultGradient()),
		viewport:     viewport.New(0, 0),
	}
}

type treeItem struct {
	label   string
	level   int
	isRoot  bool
	isCycle bool
	report  domain.ViolationReport
}

func (i treeItem) Title() string {
	if i.label == "" {
		return "────────────────────────────────────────────────"
	}
	prefix := ""
	if i.level > 0 {
		prefix = strings.Repeat("  ", i.level-1) + "└─ "
	}
	name := filepath.Base(i.label)
	if i.isRoot {
		name = "🌳 " + name
	}
	if i.isCycle {
		name = "↺ " + name + " (circular)"
	}

	// Apply styles based on violations
	style := lipgloss.NewStyle()
	if i.report.ID != 0 {
		hasError, hasWarning := false, false
		for _, v := range i.report.Violations {
			if v.SourceRef.FilePath == i.label {
				if v.Severity == "error" {
					hasError = true
				} else {
					hasWarning = true
				}
			}
		}
		if hasError {
			style = style.Foreground(lipgloss.Color("#FF4672")).Bold(true)
		} else if hasWarning {
			style = style.Foreground(lipgloss.Color("#FFA500")).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("#A3BE8C"))
		}
	}

	return prefix + style.Render(name)
}

func (i treeItem) Description() string {
	if i.label == "" || i.isCycle {
		return ""
	}
	stats := ""
	if i.report.ID != 0 {
		errors, warnings := 0, 0
		for _, v := range i.report.Violations {
			// Only count violations for THIS file in the tree
			if v.SourceRef.FilePath != i.label {
				continue
			}
			if v.Severity == "error" {
				errors++
			} else {
				warnings++
			}
		}
		if errors > 0 {
			stats += fmt.Sprintf(" • %d 🔴", errors)
		}
		if warnings > 0 {
			stats += fmt.Sprintf(" • %d 🟠", warnings)
		}
		if errors == 0 && warnings == 0 {
			stats += " • ✅"
		}
	}
	return i.label + stats
}
func (i treeItem) FilterValue() string { return i.label }

type projectItem struct {
	name     string
	count    int
	errors   int
	warnings int
}

func (i projectItem) Title() string       { return i.name }
func (i projectItem) Description() string {
	res := fmt.Sprintf("%d files", i.count)
	if i.errors > 0 {
		res += fmt.Sprintf(" • 🔴 %d", i.errors)
	}
	if i.warnings > 0 {
		res += fmt.Sprintf(" • 🟠 %d", i.warnings)
	}
	if i.errors == 0 && i.warnings == 0 {
		res += " • ✅ Clean"
	}
	return res
}
func (i projectItem) FilterValue() string { return i.name }

type reportItem struct {
	report domain.ViolationReport
}

func (i reportItem) Title() string {
	return filepath.Base(i.report.FilePath)
}
func (i reportItem) Description() string {
	ts := time.Unix(i.report.Timestamp, 0).Format("2006-01-02 15:04")
	badge := getPlatformBadge(string(i.report.Platform))
	
	errors, warnings := 0, 0
	for _, v := range i.report.Violations {
		if v.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}
	
	stats := ""
	if errors > 0 {
		stats += fmt.Sprintf(" • %d 🔴", errors)
	}
	if warnings > 0 {
		stats += fmt.Sprintf(" • %d 🟠", warnings)
	}
	if errors == 0 && warnings == 0 {
		stats += " • ✅"
	}

	return fmt.Sprintf("%s %s%s", badge, ts, stats)
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
		m.treeView.SetSize(msg.Width-4, msg.Height-6)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		if m.state == stateResults {
			m.viewport.SetContent(m.resultsView())
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			switch m.state {
			case stateProjectReports:
				m.updateTreeList()
				m.state = stateProjectTree
				return m, nil
			case stateProjectTree:
				m.state = stateProjectReports
				return m, nil
			}
		case "esc":
			switch m.state {
			case stateProjectReports, stateProjectTree:
				m.state = stateProjects
				return m, nil
			case stateResults:
				if len(m.treeView.Items()) > 0 {
					m.state = stateProjectTree
				} else {
					m.state = stateProjectReports
				}
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
			case stateProjectTree:
				if i, ok := m.treeView.SelectedItem().(treeItem); ok {
					if i.label == "" {
						return m, nil // Skip separators
					}
					if i.report.ID != 0 {
						// Filter violations to only show those for the selected file
						filtered := i.report
						var vList []domain.Violation
						for _, v := range i.report.Violations {
							if v.SourceRef.FilePath == i.label {
								vList = append(vList, v)
							}
						}
						filtered.Violations = vList
						filtered.FilePath = i.label // Update path to show the selected file

						m.results = filtered
						m.state = stateResults
						m.viewport.SetContent(m.resultsView())
						m.viewport.YOffset = 0
						return m, nil
					}
				}
			}
		}
	}

	switch m.state {
	case stateProjects:
		m.projectsList, cmd = m.projectsList.Update(msg)
	case stateProjectReports:
		m.reportsList, cmd = m.reportsList.Update(msg)
	case stateProjectTree:
		m.treeView, cmd = m.treeView.Update(msg)
	case stateResults:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m *MainModel) updateTreeList() {
	var items []list.Item

	for _, r := range m.allReports {
		if r.ProjectName == m.selectedProject {
			if r.Hierarchy != nil {
				// Each report represents a PageTree entry point.
				// We show each one as an independent tree.
				items = append(items, m.flattenTree(r.Hierarchy, r, 0, true)...)
				
				// Add an empty item for visual separation between different page trees
				items = append(items, treeItem{label: "", level: 0})
			}
		}
	}
	m.treeView.SetItems(items)
	m.treeView.Title = fmt.Sprintf("Component Explorer: %s (press 't' for list)", m.selectedProject)
}

func (m *MainModel) flattenTree(node *domain.FileNode, report domain.ViolationReport, level int, isRoot bool) []list.Item {
	var items []list.Item
	
	// Create item for current node with the full report context and cycle marker
	item := treeItem{
		label:   node.FilePath,
		level:   level,
		isRoot:  isRoot,
		isCycle: node.IsCycle,
		report:  report,
	}
	
	items = append(items, item)
	
	// Stop recursion if this is a cycle node
	if node.IsCycle {
		return items
	}
	
	for _, child := range node.Children {
		items = append(items, m.flattenTree(child, report, level+1, false)...)
	}
	
	return items
}

func (m *MainModel) updateProjectsList() {
	type runStats struct {
		runID     string
		timestamp int64
		count     int
		errors    int
		warnings  int
		reports   []domain.ViolationReport
	}
	
	// Map project name -> latest run
	latestRuns := make(map[string]*runStats)

	for _, r := range m.allReports {
		name := r.ProjectName
		if name == "" {
			name = "Unknown Project"
		}
		
		run, ok := latestRuns[name]
		if !ok || r.Timestamp > run.timestamp {
			latestRuns[name] = &runStats{
				runID:     r.RunID,
				timestamp: r.Timestamp,
				reports:   []domain.ViolationReport{r},
			}
			run = latestRuns[name]
		} else if r.RunID == run.runID {
			run.reports = append(run.reports, r)
		} else {
			// Older run, ignore for the main project list
			continue
		}

		run.count++
		for _, v := range r.Violations {
			switch v.Severity {
			case "error":
				run.errors++
			case "warning":
				run.warnings++
			}
		}
	}

	var items []list.Item
	// Sort by timestamp or name if needed, here we just iterate
	for name, run := range latestRuns {
		items = append(items, projectItem{
			name:     name,
			count:    run.count,
			errors:   run.errors,
			warnings: run.warnings,
		})
	}
	m.projectsList.SetItems(items)
	m.projectsList.Title = "Projects (Latest Snapshots)"
}

func (m *MainModel) updateReportsList() {
	var items []list.Item
	
	// Find the latest RunID for the selected project
	latestRunID := ""
	var latestTS int64
	for _, r := range m.allReports {
		if r.ProjectName == m.selectedProject || (m.selectedProject == "Unknown Project" && r.ProjectName == "") {
			if r.Timestamp > latestTS {
				latestTS = r.Timestamp
				latestRunID = r.RunID
			}
		}
	}

	// Show only reports from that RunID
	for _, r := range m.allReports {
		if r.RunID == latestRunID {
			items = append(items, reportItem{report: r})
		}
	}
	m.reportsList.SetItems(items)
	m.reportsList.Title = fmt.Sprintf("Last Run: %s", m.selectedProject)
}

func (m MainModel) View() string {
	switch m.state {
	case stateProjects:
		return m.projectsView()
	case stateProjectReports:
		return m.reportsView()
	case stateProjectTree:
		return m.treeListView()
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

func (m MainModel) treeListView() string {
	return docStyle.Render(m.treeView.View())
}
