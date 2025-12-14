package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gavin/gitta/internal/core"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// TaskSelectorModel represents the state of the task selection TUI.
type TaskSelectorModel struct {
	tasks     []*core.Story
	cursor    int
	selected  map[int]struct{}
	quitting  bool
	cancelled bool
	ctx       context.Context
}

// SelectTasks displays an interactive TUI for selecting tasks to rollover.
// Returns the selected task IDs and an error if cancelled or context cancelled.
func SelectTasks(ctx context.Context, tasks []*core.Story) ([]string, error) {
	if len(tasks) == 0 {
		return []string{}, nil
	}

	model := TaskSelectorModel{
		tasks:    tasks,
		selected: make(map[int]struct{}),
		ctx:      ctx,
	}

	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Create program with context support
	p := tea.NewProgram(model, tea.WithContext(ctx))

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m := finalModel.(TaskSelectorModel)

	if m.cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}

	// Extract selected task IDs
	var selectedIDs []string
	for idx := range m.selected {
		if idx >= 0 && idx < len(m.tasks) {
			selectedIDs = append(selectedIDs, m.tasks[idx].ID)
		}
	}

	return selectedIDs, nil
}

// Init initializes the TUI model.
func (m TaskSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m TaskSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check context cancellation
	if m.ctx != nil {
		select {
		case <-m.ctx.Done():
			m.cancelled = true
			return m, tea.Quit
		default:
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case "enter":
			// Confirm selection and quit
			m.quitting = true
			return m, tea.Quit

		case " ":
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

// View renders the TUI.
func (m TaskSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	if m.cancelled {
		return "\n  Selection cancelled.\n\n"
	}

	var s string
	s += titleStyle.Render("Select tasks to rollover (space to select, Enter to confirm, Esc to cancel)") + "\n\n"

	if len(m.tasks) == 0 {
		s += "  No tasks available.\n"
		return s
	}

	for i, task := range m.tasks {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = selectedStyle.Render("x")
		}

		// Format task display: ID - Title (Status)
		taskDisplay := fmt.Sprintf("%s - %s (%s)", task.ID, task.Title, task.Status)
		if len(taskDisplay) > 60 {
			taskDisplay = taskDisplay[:57] + "..."
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, taskDisplay)
	}

	s += "\n"
	s += helpStyle.Render("↑/↓: navigate  Space: select  Enter: confirm  Esc/q: cancel")

	return s
}
