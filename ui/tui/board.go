package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Task represents a single task item displayed in a kanban column.
type Task struct {
	ID          string
	Title       string
	Status      string
	Description string // Optional
}

// Column represents one of three vertical sections in the kanban board.
type Column struct {
	Title  string
	Tasks  []Task
	Status string
}

// CursorPosition represents the current interactive focus point in the board.
type CursorPosition struct {
	ColumnIndex int
	TaskIndex   [3]int // Task index per column (length 3, one per column)
}

// BoardModel represents the entire board state for the Bubble Tea TUI.
type BoardModel struct {
	Columns  [3]Column
	Cursor   CursorPosition
	Width    int
	Height   int
	Quitting bool
	Ctx      context.Context
}

// newSampleBoard creates a hardcoded sample board with 4 tasks across 3 columns.
func newSampleBoard() BoardModel {
	return BoardModel{
		Columns: [3]Column{
			{
				Title:  "To Do",
				Status: "todo",
				Tasks: []Task{
					{ID: "TASK-001", Title: "Research Bubble Tea layouts", Status: "todo"},
					{ID: "TASK-002", Title: "Design three-column structure", Status: "todo"},
				},
			},
			{
				Title:  "In Progress",
				Status: "in-progress",
				Tasks: []Task{
					{ID: "TASK-003", Title: "Implement cursor navigation", Status: "in-progress"},
				},
			},
			{
				Title:  "Done",
				Status: "done",
				Tasks: []Task{
					{ID: "TASK-004", Title: "Set up project structure", Status: "done"},
				},
			},
		},
		Cursor: CursorPosition{
			ColumnIndex: 0,
			TaskIndex:   [3]int{0, 0, 0},
		},
		Width:  80,
		Height: 24,
	}
}

var (
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)
)

// Init initializes the TUI model with sample data and cursor position.
// It returns a command to get the initial window size.
func (m BoardModel) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update handles messages and updates the model state.
// It processes window resize events, keyboard input, and context cancellation.
func (m BoardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check context cancellation
	if m.Ctx != nil {
		select {
		case <-m.Ctx.Done():
			m.Quitting = true
			return m, tea.Quit
		default:
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Preserve cursor position when terminal is resized
		m.Width = msg.Width
		m.Height = msg.Height
		// Ensure cursor position is still valid after resize
		if m.Cursor.ColumnIndex >= len(m.Columns) {
			m.Cursor.ColumnIndex = 0
		}
		colIdx := m.Cursor.ColumnIndex
		if m.Cursor.TaskIndex[colIdx] >= len(m.Columns[colIdx].Tasks) && len(m.Columns[colIdx].Tasks) > 0 {
			m.Cursor.TaskIndex[colIdx] = len(m.Columns[colIdx].Tasks) - 1
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Quitting = true
			return m, tea.Quit

		case "left":
			// Move to previous column
			if m.Cursor.ColumnIndex > 0 {
				m.Cursor.ColumnIndex--
			}

		case "right":
			// Move to next column
			if m.Cursor.ColumnIndex < 2 {
				m.Cursor.ColumnIndex++
			}

		case "up":
			// Move to previous task in current column
			if m.Cursor.TaskIndex[m.Cursor.ColumnIndex] > 0 {
				m.Cursor.TaskIndex[m.Cursor.ColumnIndex]--
			}

		case "down":
			// Move to next task in current column
			colIdx := m.Cursor.ColumnIndex
			maxTasks := len(m.Columns[colIdx].Tasks)
			if m.Cursor.TaskIndex[colIdx] < maxTasks-1 {
				m.Cursor.TaskIndex[colIdx]++
			}
		}
	}

	return m, nil
}

// renderTask renders a single task card with ID and Title format.
func (m BoardModel) renderTask(task Task, colWidth int, isSelected bool) string {
	// Format: [ID] Title
	taskText := fmt.Sprintf("[%s] %s", task.ID, task.Title)

	// Truncate if exceeds column width - 6 (for ID + spacing)
	maxWidth := colWidth - 6
	if lipgloss.Width(taskText) > maxWidth {
		// Truncate with ellipsis
		taskText = taskText[:maxWidth-3] + "..."
	}

	// Apply cursor indicator if selected
	if isSelected {
		return "> " + taskText
	}
	return "  " + taskText
}

// renderColumn renders a single column with borders, header, and tasks.
func (m BoardModel) renderColumn(colIdx int) string {
	if colIdx < 0 || colIdx >= len(m.Columns) {
		return ""
	}

	col := m.Columns[colIdx]
	colWidth := m.Width / 3

	// Build column content
	var lines []string

	// Header
	header := headerStyle.Width(colWidth - 2).Render(col.Title)
	lines = append(lines, header)
	lines = append(lines, strings.Repeat("─", colWidth-2))

	// Tasks or empty message
	if len(col.Tasks) == 0 {
		lines = append(lines, "  (empty)")
	} else {
		// Render each task
		for i, task := range col.Tasks {
			isSelected := colIdx == m.Cursor.ColumnIndex && i == m.Cursor.TaskIndex[colIdx]
			taskLine := m.renderTask(task, colWidth, isSelected)
			lines = append(lines, taskLine)
		}
	}

	// Join lines and apply border style
	content := strings.Join(lines, "\n")
	return columnStyle.Width(colWidth).Render(content)
}

// View renders the TUI.
func (m BoardModel) View() string {
	// Check minimum width
	if m.Width < 80 {
		return fmt.Sprintf("Terminal too narrow (width: %d, minimum: 80). Please resize your terminal.", m.Width)
	}

	// Render three columns
	col1 := m.renderColumn(0)
	col2 := m.renderColumn(1)
	col3 := m.renderColumn(2)

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)
}

// ShowBoard displays an interactive kanban board TUI.
// It creates a Bubble Tea program with the board model and runs it until the user quits.
// Returns an error if the TUI fails to initialize or run.
func ShowBoard(ctx context.Context) error {
	model := newSampleBoard()
	model.Ctx = ctx

	// Check context cancellation before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Create program with context support
	p := tea.NewProgram(model, tea.WithContext(ctx))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if user cancelled
	m := finalModel.(BoardModel)
	if m.Quitting {
		return nil
	}

	return nil
}
