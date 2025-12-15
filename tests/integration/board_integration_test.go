package integration

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gavin/gitta/ui/tui"
)

// TestBoard_ThreeColumnLayout tests three-column layout display (visual verification)
func TestBoard_ThreeColumnLayout(t *testing.T) {
	model := tui.BoardModel{
		Columns: [3]tui.Column{
			{Title: "To Do", Status: "todo", Tasks: []tui.Task{}},
			{Title: "In Progress", Status: "in-progress", Tasks: []tui.Task{}},
			{Title: "Done", Status: "done", Tasks: []tui.Task{}},
		},
		Width:  120,
		Height: 24,
	}

	view := model.View()

	// Verify all three column titles appear
	expectedTitles := []string{"To Do", "In Progress", "Done"}
	for _, title := range expectedTitles {
		if !strings.Contains(view, title) {
			t.Errorf("View() missing column title %q", title)
		}
	}

	// Verify no error message (width is sufficient)
	if strings.Contains(view, "Terminal too narrow") {
		t.Errorf("View() should not show error for width 120")
	}
}

// TestBoard_HardcodedDataDisplay tests hardcoded data display in all three columns
func TestBoard_HardcodedDataDisplay(t *testing.T) {
	model := tui.BoardModel{
		Columns: [3]tui.Column{
			{
				Title:  "To Do",
				Status: "todo",
				Tasks: []tui.Task{
					{ID: "TASK-001", Title: "Research Bubble Tea layouts", Status: "todo"},
					{ID: "TASK-002", Title: "Design three-column structure", Status: "todo"},
				},
			},
			{
				Title:  "In Progress",
				Status: "in-progress",
				Tasks: []tui.Task{
					{ID: "TASK-003", Title: "Implement cursor navigation", Status: "in-progress"},
				},
			},
			{
				Title:  "Done",
				Status: "done",
				Tasks: []tui.Task{
					{ID: "TASK-004", Title: "Set up project structure", Status: "done"},
				},
			},
		},
		Width:  120,
		Height: 24,
	}

	view := model.View()

	// Verify all task IDs appear in view
	expectedTaskIDs := []string{"TASK-001", "TASK-002", "TASK-003", "TASK-004"}
	for _, taskID := range expectedTaskIDs {
		if !strings.Contains(view, taskID) {
			t.Errorf("View() missing task ID %q", taskID)
		}
	}

	// Verify tasks appear in correct columns (by checking task titles or partial matches)
	// Titles may be truncated, so check for partial matches
	expectedTitleParts := []string{
		"Research",  // Part of "Research Bubble Tea layouts"
		"Design",    // Part of "Design three-column structure"
		"Implement", // Part of "Implement cursor navigation"
		"Set up",    // Part of "Set up project structure"
	}
	for _, titlePart := range expectedTitleParts {
		if !strings.Contains(view, titlePart) {
			t.Errorf("View() missing task title part %q", titlePart)
		}
	}
}

// TestBoard_CursorNavigation tests full cursor navigation flow (all four directions)
func TestBoard_CursorNavigation(t *testing.T) {
	model := tui.BoardModel{
		Columns: [3]tui.Column{
			{
				Title:  "To Do",
				Status: "todo",
				Tasks: []tui.Task{
					{ID: "T1", Title: "Task 1", Status: "todo"},
					{ID: "T2", Title: "Task 2", Status: "todo"},
				},
			},
			{
				Title:  "In Progress",
				Status: "in-progress",
				Tasks: []tui.Task{
					{ID: "T3", Title: "Task 3", Status: "in-progress"},
				},
			},
			{
				Title:  "Done",
				Status: "done",
				Tasks: []tui.Task{
					{ID: "T4", Title: "Task 4", Status: "done"},
				},
			},
		},
		Cursor: tui.CursorPosition{
			ColumnIndex: 0,
			TaskIndex:   [3]int{0, 0, 0},
		},
		Width:  120,
		Height: 24,
	}

	// Test navigation sequence: right -> down -> right -> up -> left
	steps := []struct {
		name     string
		key      tea.KeyType
		wantCol  int
		wantTask [3]int
	}{
		{
			name:     "move right to column 1",
			key:      tea.KeyRight,
			wantCol:  1,
			wantTask: [3]int{0, 0, 0},
		},
		{
			name:     "move down in column 1 (should stay at 0, only 1 task)",
			key:      tea.KeyDown,
			wantCol:  1,
			wantTask: [3]int{0, 0, 0},
		},
		{
			name:     "move right to column 2",
			key:      tea.KeyRight,
			wantCol:  2,
			wantTask: [3]int{0, 0, 0},
		},
		{
			name:     "move up in column 2 (should stay at 0, only 1 task)",
			key:      tea.KeyUp,
			wantCol:  2,
			wantTask: [3]int{0, 0, 0},
		},
		{
			name:     "move left back to column 1",
			key:      tea.KeyLeft,
			wantCol:  1,
			wantTask: [3]int{0, 0, 0},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: step.key}
			updated, _ := model.Update(keyMsg)
			model = updated.(tui.BoardModel)

			if model.Cursor.ColumnIndex != step.wantCol {
				t.Errorf("After %s: ColumnIndex = %d, want %d", step.name, model.Cursor.ColumnIndex, step.wantCol)
			}

			for i := range step.wantTask {
				if model.Cursor.TaskIndex[i] != step.wantTask[i] {
					t.Errorf("After %s: TaskIndex[%d] = %d, want %d", step.name, i, model.Cursor.TaskIndex[i], step.wantTask[i])
				}
			}
		})
	}

	// Test navigation in column 0 with multiple tasks
	model.Cursor.ColumnIndex = 0
	model.Cursor.TaskIndex = [3]int{0, 0, 0}

	// Move down in column 0
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(keyMsg)
	model = updated.(tui.BoardModel)

	if model.Cursor.TaskIndex[0] != 1 {
		t.Errorf("After moving down in column 0: TaskIndex[0] = %d, want 1", model.Cursor.TaskIndex[0])
	}

	// Move up back to first task
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = model.Update(keyMsg)
	model = updated.(tui.BoardModel)

	if model.Cursor.TaskIndex[0] != 0 {
		t.Errorf("After moving up in column 0: TaskIndex[0] = %d, want 0", model.Cursor.TaskIndex[0])
	}
}

// TestBoard_ShowBoard_ContextCancellation tests ShowBoard with context cancellation
func TestBoard_ShowBoard_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// ShowBoard should handle cancelled context gracefully
	err := tui.ShowBoard(ctx)
	if err == nil {
		t.Error("ShowBoard() should return error for cancelled context")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("ShowBoard() error = %v, want to contain 'context cancelled'", err)
	}
}
