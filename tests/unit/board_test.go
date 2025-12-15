package unit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gavin/gitta/ui/tui"
)

// TestBoardModel_View_ThreeColumns tests that View() renders three columns with empty data
func TestBoardModel_View_ThreeColumns(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		columns  [3]tui.Column
		wantCols int // Expected number of column headers
	}{
		{
			name:   "three empty columns",
			width:  120,
			height: 24,
			columns: [3]tui.Column{
				{Title: "To Do", Status: "todo", Tasks: []tui.Task{}},
				{Title: "In Progress", Status: "in-progress", Tasks: []tui.Task{}},
				{Title: "Done", Status: "done", Tasks: []tui.Task{}},
			},
			wantCols: 3,
		},
		{
			name:   "columns with different titles",
			width:  100,
			height: 30,
			columns: [3]tui.Column{
				{Title: "Column 1", Status: "todo", Tasks: []tui.Task{}},
				{Title: "Column 2", Status: "in-progress", Tasks: []tui.Task{}},
				{Title: "Column 3", Status: "done", Tasks: []tui.Task{}},
			},
			wantCols: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: tt.columns,
				Cursor: tui.CursorPosition{
					ColumnIndex: 0,
					TaskIndex:   [3]int{0, 0, 0},
				},
				Width:  tt.width,
				Height: tt.height,
			}

			view := model.View()

			// Check that all three column titles appear
			for _, col := range tt.columns {
				if !strings.Contains(view, col.Title) {
					t.Errorf("View() missing column title %q", col.Title)
				}
			}

			// Check minimum width error message doesn't appear
			if strings.Contains(view, "Terminal too narrow") {
				t.Errorf("View() should not show error for width %d", tt.width)
			}
		})
	}
}

// TestBoardModel_View_MinimumWidth tests View() handling terminal width < 80 characters
func TestBoardModel_View_MinimumWidth(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		wantError   bool
		wantMessage string
	}{
		{
			name:        "width too narrow",
			width:       60,
			height:      24,
			wantError:   true,
			wantMessage: "Terminal too narrow",
		},
		{
			name:        "exact minimum width",
			width:       80,
			height:      24,
			wantError:   false,
			wantMessage: "",
		},
		{
			name:        "width above minimum",
			width:       100,
			height:      24,
			wantError:   false,
			wantMessage: "",
		},
		{
			name:        "very narrow terminal",
			width:       40,
			height:      24,
			wantError:   true,
			wantMessage: "Terminal too narrow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Status: "todo", Tasks: []tui.Task{}},
					{Title: "In Progress", Status: "in-progress", Tasks: []tui.Task{}},
					{Title: "Done", Status: "done", Tasks: []tui.Task{}},
				},
				Width:  tt.width,
				Height: tt.height,
			}

			view := model.View()

			if tt.wantError {
				if !strings.Contains(view, tt.wantMessage) {
					t.Errorf("View() = %q, want to contain %q", view, tt.wantMessage)
				}
				if !strings.Contains(view, "80") {
					t.Errorf("View() error message should mention minimum width 80")
				}
			} else {
				if strings.Contains(view, "Terminal too narrow") {
					t.Errorf("View() should not show error for width %d", tt.width)
				}
			}
		})
	}
}

// TestBoardModel_View_WithTasks tests View() displaying tasks in correct columns
func TestBoardModel_View_WithTasks(t *testing.T) {
	tests := []struct {
		name      string
		columns   [3]tui.Column
		width     int
		wantTasks map[int][]string // Column index -> Task IDs that should appear
	}{
		{
			name: "columns with tasks",
			columns: [3]tui.Column{
				{
					Title:  "To Do",
					Status: "todo",
					Tasks: []tui.Task{
						{ID: "TASK-001", Title: "Task 1", Status: "todo"},
					},
				},
				{
					Title:  "In Progress",
					Status: "in-progress",
					Tasks: []tui.Task{
						{ID: "TASK-002", Title: "Task 2", Status: "in-progress"},
						{ID: "TASK-003", Title: "Task 3", Status: "in-progress"},
					},
				},
				{
					Title:  "Done",
					Status: "done",
					Tasks:  []tui.Task{},
				},
			},
			width: 120,
			wantTasks: map[int][]string{
				0: {"TASK-001"},
				1: {"TASK-002", "TASK-003"},
				2: {},
			},
		},
		{
			name: "empty columns",
			columns: [3]tui.Column{
				{Title: "To Do", Status: "todo", Tasks: []tui.Task{}},
				{Title: "In Progress", Status: "in-progress", Tasks: []tui.Task{}},
				{Title: "Done", Status: "done", Tasks: []tui.Task{}},
			},
			width: 120,
			wantTasks: map[int][]string{
				0: {},
				1: {},
				2: {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: tt.columns,
				Width:   tt.width,
				Height:  24,
			}

			view := model.View()

			// Check all column titles appear
			for i, col := range tt.columns {
				if !strings.Contains(view, col.Title) {
					t.Errorf("View() missing column title %q for column %d", col.Title, i)
				}
			}

			// Check tasks appear in correct columns (by checking task IDs in view)
			for colIdx, taskIDs := range tt.wantTasks {
				for _, taskID := range taskIDs {
					if !strings.Contains(view, taskID) {
						t.Errorf("View() missing task ID %q in column %d", taskID, colIdx)
					}
				}
			}

			// Check empty message for empty columns
			for i, col := range tt.columns {
				if len(col.Tasks) == 0 {
					// Empty message should appear somewhere in the view
					// (we can't test exact column, but can verify it's present)
					if !strings.Contains(view, "empty") {
						t.Errorf("View() should show empty message for column %d with no tasks", i)
					}
				}
			}
		})
	}
}

// TestBoardModel_View_TaskFormat tests task card rendering with ID and Title format in View()
func TestBoardModel_View_TaskFormat(t *testing.T) {
	tests := []struct {
		name       string
		tasks      [3][]tui.Task
		cursorCol  int
		cursorTask int
		width      int
		wantID     string
		wantTitle  string
		wantCursor bool
	}{
		{
			name: "selected task with cursor indicator",
			tasks: [3][]tui.Task{
				{{ID: "TASK-001", Title: "Test Task", Status: "todo"}},
				{},
				{},
			},
			cursorCol:  0,
			cursorTask: 0,
			width:      120,
			wantID:     "TASK-001",
			wantTitle:  "Test Task",
			wantCursor: true,
		},
		{
			name: "unselected task",
			tasks: [3][]tui.Task{
				{{ID: "TASK-002", Title: "Another Task", Status: "in-progress"}},
				{},
				{},
			},
			cursorCol:  1, // Cursor in different column
			cursorTask: 0,
			width:      120,
			wantID:     "TASK-002",
			wantTitle:  "Another Task",
			wantCursor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskIndex := [3]int{0, 0, 0}
			if tt.cursorCol < 3 && len(tt.tasks[tt.cursorCol]) > 0 {
				taskIndex[tt.cursorCol] = tt.cursorTask
			}

			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: tt.tasks[0]},
					{Title: "In Progress", Tasks: tt.tasks[1]},
					{Title: "Done", Tasks: tt.tasks[2]},
				},
				Cursor: tui.CursorPosition{
					ColumnIndex: tt.cursorCol,
					TaskIndex:   taskIndex,
				},
				Width:  tt.width,
				Height: 24,
			}

			view := model.View()

			// Check ID appears in [ID] format
			if !strings.Contains(view, "["+tt.wantID+"]") {
				t.Errorf("View() missing ID in brackets %q, got view", tt.wantID)
			}

			// Check title appears
			if tt.wantTitle != "" && !strings.Contains(view, tt.wantTitle) {
				t.Errorf("View() missing title %q", tt.wantTitle)
			}

			// Check cursor indicator for selected task
			if tt.wantCursor {
				// Selected task should have ">" indicator
				// We can't test exact position, but can verify it appears near the task
				if !strings.Contains(view, ">") {
					t.Errorf("View() missing cursor indicator for selected task")
				}
			}
		})
	}
}

// TestBoardModel_RenderTask_Truncation tests task title truncation when exceeding column width
func TestBoardModel_RenderTask_Truncation(t *testing.T) {
	tests := []struct {
		name         string
		task         tui.Task
		colWidth     int
		wantTruncate bool
		wantEllipsis bool
	}{
		{
			name: "title fits within column",
			task: tui.Task{
				ID:     "TASK-001",
				Title:  "Short",
				Status: "todo",
			},
			colWidth:     30,
			wantTruncate: false,
			wantEllipsis: false,
		},
		{
			name: "title exceeds column width",
			task: tui.Task{
				ID:     "TASK-002",
				Title:  "This is a very long task title that definitely exceeds the column width",
				Status: "todo",
			},
			colWidth:     30, // Narrow enough to force truncation
			wantTruncate: true,
			wantEllipsis: true,
		},
		{
			name: "title fits within column",
			task: tui.Task{
				ID:     "TASK-003",
				Title:  "Short title",
				Status: "todo",
			},
			colWidth:     40,
			wantTruncate: false,
			wantEllipsis: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: []tui.Task{tt.task}},
					{Title: "In Progress", Tasks: []tui.Task{}},
					{Title: "Done", Tasks: []tui.Task{}},
				},
				Width:  tt.colWidth * 3,
				Height: 24,
			}

			view := model.View()

			// Check ID always appears
			if !strings.Contains(view, "["+tt.task.ID+"]") {
				t.Errorf("View() missing ID %q", tt.task.ID)
			}

			// Check truncation
			if tt.wantTruncate {
				// Title should be truncated, so full title should not appear
				if strings.Contains(view, tt.task.Title) {
					t.Errorf("View() should truncate long title, but full title appears")
				}
				if tt.wantEllipsis && !strings.Contains(view, "...") {
					t.Errorf("View() should add ellipsis for truncated title")
				}
			} else {
				// Title should appear fully
				if !strings.Contains(view, tt.task.Title) {
					t.Errorf("View() missing title %q", tt.task.Title)
				}
			}
		})
	}
}

// TestBoardModel_Update_CursorMovement_LeftRight tests cursor movement left/right between columns
func TestBoardModel_Update_CursorMovement_LeftRight(t *testing.T) {
	tests := []struct {
		name        string
		initialCol  int
		key         string
		wantCol     int
		wantChanged bool
	}{
		{
			name:        "move right from first column",
			initialCol:  0,
			key:         "right",
			wantCol:     1,
			wantChanged: true,
		},
		{
			name:        "move right from middle column",
			initialCol:  1,
			key:         "right",
			wantCol:     2,
			wantChanged: true,
		},
		{
			name:        "move right from last column (boundary)",
			initialCol:  2,
			key:         "right",
			wantCol:     2, // Should not move
			wantChanged: false,
		},
		{
			name:        "move left from last column",
			initialCol:  2,
			key:         "left",
			wantCol:     1,
			wantChanged: true,
		},
		{
			name:        "move left from middle column",
			initialCol:  1,
			key:         "left",
			wantCol:     0,
			wantChanged: true,
		},
		{
			name:        "move left from first column (boundary)",
			initialCol:  0,
			key:         "left",
			wantCol:     0, // Should not move
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: []tui.Task{{ID: "T1", Title: "Task 1"}}},
					{Title: "In Progress", Tasks: []tui.Task{{ID: "T2", Title: "Task 2"}}},
					{Title: "Done", Tasks: []tui.Task{{ID: "T3", Title: "Task 3"}}},
				},
				Cursor: tui.CursorPosition{
					ColumnIndex: tt.initialCol,
					TaskIndex:   [3]int{0, 0, 0},
				},
				Width:  120,
				Height: 24,
			}

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "left" {
				keyMsg = tea.KeyMsg{Type: tea.KeyLeft}
			} else if tt.key == "right" {
				keyMsg = tea.KeyMsg{Type: tea.KeyRight}
			}

			updated, _ := model.Update(keyMsg)
			updatedModel := updated.(tui.BoardModel)

			if updatedModel.Cursor.ColumnIndex != tt.wantCol {
				t.Errorf("Update() ColumnIndex = %d, want %d", updatedModel.Cursor.ColumnIndex, tt.wantCol)
			}
		})
	}
}

// TestBoardModel_Update_CursorMovement_UpDown tests cursor movement up/down within a column
func TestBoardModel_Update_CursorMovement_UpDown(t *testing.T) {
	tests := []struct {
		name        string
		colIdx      int
		initialTask int
		tasks       []tui.Task
		key         string
		wantTask    int
		wantChanged bool
	}{
		{
			name:        "move down from first task",
			colIdx:      0,
			initialTask: 0,
			tasks: []tui.Task{
				{ID: "T1", Title: "Task 1"},
				{ID: "T2", Title: "Task 2"},
			},
			key:         "down",
			wantTask:    1,
			wantChanged: true,
		},
		{
			name:        "move down from last task (boundary)",
			colIdx:      0,
			initialTask: 1,
			tasks: []tui.Task{
				{ID: "T1", Title: "Task 1"},
				{ID: "T2", Title: "Task 2"},
			},
			key:         "down",
			wantTask:    1, // Should not move
			wantChanged: false,
		},
		{
			name:        "move up from last task",
			colIdx:      0,
			initialTask: 1,
			tasks: []tui.Task{
				{ID: "T1", Title: "Task 1"},
				{ID: "T2", Title: "Task 2"},
			},
			key:         "up",
			wantTask:    0,
			wantChanged: true,
		},
		{
			name:        "move up from first task (boundary)",
			colIdx:      0,
			initialTask: 0,
			tasks: []tui.Task{
				{ID: "T1", Title: "Task 1"},
			},
			key:         "up",
			wantTask:    0, // Should not move
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskIndex := [3]int{0, 0, 0}
			taskIndex[tt.colIdx] = tt.initialTask

			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: tt.tasks},
					{Title: "In Progress", Tasks: []tui.Task{}},
					{Title: "Done", Tasks: []tui.Task{}},
				},
				Cursor: tui.CursorPosition{
					ColumnIndex: tt.colIdx,
					TaskIndex:   taskIndex,
				},
				Width:  120,
				Height: 24,
			}

			var keyMsg tea.KeyMsg
			if tt.key == "up" {
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == "down" {
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			}

			updated, _ := model.Update(keyMsg)
			updatedModel := updated.(tui.BoardModel)

			if updatedModel.Cursor.TaskIndex[tt.colIdx] != tt.wantTask {
				t.Errorf("Update() TaskIndex[%d] = %d, want %d", tt.colIdx, updatedModel.Cursor.TaskIndex[tt.colIdx], tt.wantTask)
			}
		})
	}
}

// TestBoardModel_Update_CursorBoundaries tests cursor boundary conditions
func TestBoardModel_Update_CursorBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		initialCol  int
		initialTask [3]int
		key         string
		wantCol     int
		wantTask    [3]int
	}{
		{
			name:        "cannot move left from first column",
			initialCol:  0,
			initialTask: [3]int{0, 0, 0},
			key:         "left",
			wantCol:     0,
			wantTask:    [3]int{0, 0, 0},
		},
		{
			name:        "cannot move right from last column",
			initialCol:  2,
			initialTask: [3]int{0, 0, 0},
			key:         "right",
			wantCol:     2,
			wantTask:    [3]int{0, 0, 0},
		},
		{
			name:        "cannot move up from first task",
			initialCol:  0,
			initialTask: [3]int{0, 0, 0},
			key:         "up",
			wantCol:     0,
			wantTask:    [3]int{0, 0, 0},
		},
		{
			name:        "cannot move down from last task",
			initialCol:  0,
			initialTask: [3]int{1, 0, 0}, // Task 1 in column 0 (assuming 2 tasks)
			key:         "down",
			wantCol:     0,
			wantTask:    [3]int{1, 0, 0}, // Should stay at last task
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model with tasks in first column
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: []tui.Task{
						{ID: "T1", Title: "Task 1"},
						{ID: "T2", Title: "Task 2"},
					}},
					{Title: "In Progress", Tasks: []tui.Task{}},
					{Title: "Done", Tasks: []tui.Task{}},
				},
				Cursor: tui.CursorPosition{
					ColumnIndex: tt.initialCol,
					TaskIndex:   tt.initialTask,
				},
				Width:  120,
				Height: 24,
			}

			var keyMsg tea.KeyMsg
			switch tt.key {
			case "left":
				keyMsg = tea.KeyMsg{Type: tea.KeyLeft}
			case "right":
				keyMsg = tea.KeyMsg{Type: tea.KeyRight}
			case "up":
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			case "down":
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			}

			updated, _ := model.Update(keyMsg)
			updatedModel := updated.(tui.BoardModel)

			if updatedModel.Cursor.ColumnIndex != tt.wantCol {
				t.Errorf("Update() ColumnIndex = %d, want %d", updatedModel.Cursor.ColumnIndex, tt.wantCol)
			}

			for i := range tt.wantTask {
				if updatedModel.Cursor.TaskIndex[i] != tt.wantTask[i] {
					t.Errorf("Update() TaskIndex[%d] = %d, want %d", i, updatedModel.Cursor.TaskIndex[i], tt.wantTask[i])
				}
			}
		})
	}
}

// TestBoardModel_Update_CursorEmptyColumn tests cursor behavior with empty columns
func TestBoardModel_Update_CursorEmptyColumn(t *testing.T) {
	tests := []struct {
		name        string
		colIdx      int
		tasks       []tui.Task
		initialTask int
		key         string
		wantTask    int
	}{
		{
			name:        "cannot move down in empty column",
			colIdx:      1,
			tasks:       []tui.Task{},
			initialTask: 0,
			key:         "down",
			wantTask:    0, // Should stay at 0
		},
		{
			name:        "cannot move up in empty column",
			colIdx:      1,
			tasks:       []tui.Task{},
			initialTask: 0,
			key:         "up",
			wantTask:    0, // Should stay at 0
		},
		{
			name:        "can move to empty column from other column",
			colIdx:      0,
			tasks:       []tui.Task{{ID: "T1", Title: "Task 1"}},
			initialTask: 0,
			key:         "right",
			wantTask:    0, // Task index preserved when moving to empty column
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskIndex := [3]int{0, 0, 0}
			taskIndex[tt.colIdx] = tt.initialTask

			columns := [3]tui.Column{
				{Title: "To Do", Tasks: []tui.Task{{ID: "T1", Title: "Task 1"}}},
				{Title: "In Progress", Tasks: tt.tasks},
				{Title: "Done", Tasks: []tui.Task{}},
			}

			model := tui.BoardModel{
				Columns: columns,
				Cursor: tui.CursorPosition{
					ColumnIndex: tt.colIdx,
					TaskIndex:   taskIndex,
				},
				Width:  120,
				Height: 24,
			}

			var keyMsg tea.KeyMsg
			switch tt.key {
			case "up":
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			case "down":
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			case "right":
				keyMsg = tea.KeyMsg{Type: tea.KeyRight}
			}

			updated, _ := model.Update(keyMsg)
			updatedModel := updated.(tui.BoardModel)

			if tt.key == "right" {
				// Check column changed
				if updatedModel.Cursor.ColumnIndex == tt.colIdx {
					t.Errorf("Update() should move to next column")
				}
			} else {
				// Check task index in current column
				if updatedModel.Cursor.TaskIndex[tt.colIdx] != tt.wantTask {
					t.Errorf("Update() TaskIndex[%d] = %d, want %d", tt.colIdx, updatedModel.Cursor.TaskIndex[tt.colIdx], tt.wantTask)
				}
			}
		})
	}
}

// TestBoardModel_Update_WindowSize tests terminal resize handling
func TestBoardModel_Update_WindowSize(t *testing.T) {
	tests := []struct {
		name     string
		initialW int
		initialH int
		newW     int
		newH     int
		wantW    int
		wantH    int
	}{
		{
			name:     "resize to larger size",
			initialW: 80,
			initialH: 24,
			newW:     120,
			newH:     40,
			wantW:    120,
			wantH:    40,
		},
		{
			name:     "resize to smaller size",
			initialW: 120,
			initialH: 40,
			newW:     90,
			newH:     30,
			wantW:    90,
			wantH:    30,
		},
		{
			name:     "resize preserves cursor position",
			initialW: 100,
			initialH: 30,
			newW:     150,
			newH:     50,
			wantW:    150,
			wantH:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: []tui.Task{{ID: "T1", Title: "Task 1"}}},
					{Title: "In Progress", Tasks: []tui.Task{}},
					{Title: "Done", Tasks: []tui.Task{}},
				},
				Cursor: tui.CursorPosition{
					ColumnIndex: 1,
					TaskIndex:   [3]int{0, 0, 0},
				},
				Width:  tt.initialW,
				Height: tt.initialH,
			}

			windowSizeMsg := tea.WindowSizeMsg{
				Width:  tt.newW,
				Height: tt.newH,
			}

			updated, _ := model.Update(windowSizeMsg)
			updatedModel := updated.(tui.BoardModel)

			if updatedModel.Width != tt.wantW {
				t.Errorf("Update() Width = %d, want %d", updatedModel.Width, tt.wantW)
			}
			if updatedModel.Height != tt.wantH {
				t.Errorf("Update() Height = %d, want %d", updatedModel.Height, tt.wantH)
			}

			// Verify cursor position preserved
			if updatedModel.Cursor.ColumnIndex != 1 {
				t.Errorf("Update() should preserve cursor ColumnIndex, got %d", updatedModel.Cursor.ColumnIndex)
			}
		})
	}
}

// TestBoardModel_Update_Quit tests quit handling
func TestBoardModel_Update_Quit(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantQuit bool
		wantCmd  bool // Should return tea.Quit command
	}{
		{
			name:     "quit with 'q'",
			key:      "q",
			wantQuit: true,
			wantCmd:  true,
		},
		{
			name:     "quit with 'esc'",
			key:      "esc",
			wantQuit: true,
			wantCmd:  true,
		},
		{
			name:     "quit with ctrl+c",
			key:      "ctrl+c",
			wantQuit: true,
			wantCmd:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tui.BoardModel{
				Columns: [3]tui.Column{
					{Title: "To Do", Tasks: []tui.Task{}},
					{Title: "In Progress", Tasks: []tui.Task{}},
					{Title: "Done", Tasks: []tui.Task{}},
				},
				Width:  120,
				Height: 24,
			}

			var keyMsg tea.KeyMsg
			switch tt.key {
			case "q":
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
			case "esc":
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			case "ctrl+c":
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			updated, cmd := model.Update(keyMsg)
			updatedModel := updated.(tui.BoardModel)

			if updatedModel.Quitting != tt.wantQuit {
				t.Errorf("Update() Quitting = %v, want %v", updatedModel.Quitting, tt.wantQuit)
			}

			if tt.wantCmd && cmd == nil {
				t.Errorf("Update() should return tea.Quit command for quit keys")
			}
		})
	}
}
