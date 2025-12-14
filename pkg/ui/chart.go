package ui

import (
	"fmt"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

const (
	chartWidth  = 60
	chartHeight = 15
)

// RenderBurndownChart renders an ASCII burndown chart from burndown data points.
// The chart displays remaining work over time, with optional filtering for points or tasks.
// Parameters:
//   - dataPoints: Array of burndown data points (one per day)
//   - showPoints: If true, displays story points trend
//   - showTasks: If true, displays task count trend
//
// Returns a multi-line string containing the ASCII chart with axes, data points, and legend.
func RenderBurndownChart(dataPoints []core.BurndownDataPoint, showPoints bool, showTasks bool) string {
	if len(dataPoints) == 0 {
		return "No data available for burndown chart"
	}

	var lines []string

	// Determine what to show
	if !showPoints && !showTasks {
		showTasks = true // Default to tasks
	}

	// Calculate chart dimensions
	maxValue := 0
	for _, dp := range dataPoints {
		if showPoints && dp.RemainingPoints > maxValue {
			maxValue = dp.RemainingPoints
		}
		if showTasks && dp.RemainingTasks > maxValue {
			maxValue = dp.RemainingTasks
		}
	}

	if maxValue == 0 {
		maxValue = 1 // Avoid division by zero
	}

	// Normalize to chart height
	scale := float64(chartHeight) / float64(maxValue)

	// Create chart grid
	grid := make([][]rune, chartHeight+1)
	for i := range grid {
		grid[i] = make([]rune, chartWidth+1)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Draw axes
	for i := 0; i <= chartHeight; i++ {
		grid[i][0] = '|'
	}
	for j := 0; j <= chartWidth; j++ {
		grid[chartHeight][j] = '-'
	}

	// Plot data points
	if len(dataPoints) > 0 {
		step := float64(chartWidth) / float64(len(dataPoints)-1)

		for idx, dp := range dataPoints {
			x := int(float64(idx) * step)
			if x > chartWidth {
				x = chartWidth
			}

			if showPoints {
				y := chartHeight - int(float64(dp.RemainingPoints)*scale)
				if y < 0 {
					y = 0
				}
				if y <= chartHeight {
					grid[y][x] = '*'
				}
			}

			if showTasks {
				y := chartHeight - int(float64(dp.RemainingTasks)*scale)
				if y < 0 {
					y = 0
				}
				if y <= chartHeight {
					if grid[y][x] == '*' {
						grid[y][x] = '+' // Both overlap
					} else {
						grid[y][x] = '#'
					}
				}
			}
		}

		// Draw lines between points
		for idx := 0; idx < len(dataPoints)-1; idx++ {
			x1 := int(float64(idx) * step)
			x2 := int(float64(idx+1) * step)

			var y1, y2 int
			if showPoints {
				y1 = chartHeight - int(float64(dataPoints[idx].RemainingPoints)*scale)
				y2 = chartHeight - int(float64(dataPoints[idx+1].RemainingPoints)*scale)
			} else {
				y1 = chartHeight - int(float64(dataPoints[idx].RemainingTasks)*scale)
				y2 = chartHeight - int(float64(dataPoints[idx+1].RemainingTasks)*scale)
			}

			drawLine(grid, x1, y1, x2, y2)
		}
	}

	// Convert grid to strings
	for i := 0; i <= chartHeight; i++ {
		line := string(grid[i])
		lines = append(lines, line)
	}

	// Add labels
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Max: %d", maxValue))

	// Add legend
	if showPoints && showTasks {
		lines = append(lines, "Legend: * = Points, # = Tasks, + = Both")
	} else if showPoints {
		lines = append(lines, "Legend: * = Story Points")
	} else {
		lines = append(lines, "Legend: # = Tasks")
	}

	// Add date range
	if len(dataPoints) > 0 {
		startDate := dataPoints[0].Date.Format("2006-01-02")
		endDate := dataPoints[len(dataPoints)-1].Date.Format("2006-01-02")
		lines = append(lines, fmt.Sprintf("Date range: %s to %s", startDate, endDate))
	}

	return strings.Join(lines, "\n")
}

// drawLine draws a line between two points using simple characters.
func drawLine(grid [][]rune, x1, y1, x2, y2 int) {
	if x1 < 0 {
		x1 = 0
	}
	if x2 < 0 {
		x2 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if y2 < 0 {
		y2 = 0
	}
	if x1 > len(grid[0])-1 {
		x1 = len(grid[0]) - 1
	}
	if x2 > len(grid[0])-1 {
		x2 = len(grid[0]) - 1
	}
	if y1 > len(grid)-1 {
		y1 = len(grid) - 1
	}
	if y2 > len(grid)-1 {
		y2 = len(grid) - 1
	}

	dx := x2 - x1
	dy := y2 - y1

	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}

	if steps == 0 {
		return
	}

	xStep := float64(dx) / float64(steps)
	yStep := float64(dy) / float64(steps)

	for i := 0; i <= steps; i++ {
		x := x1 + int(float64(i)*xStep)
		y := y1 + int(float64(i)*yStep)

		if x >= 0 && x < len(grid[0]) && y >= 0 && y < len(grid) {
			if grid[y][x] == ' ' {
				grid[y][x] = '.'
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// FormatBurndownCSV formats burndown data points as CSV output.
// The CSV includes columns: Date, RemainingPoints, RemainingTasks, TotalPoints, TotalTasks.
// Returns a multi-line string with header row followed by data rows.
func FormatBurndownCSV(dataPoints []core.BurndownDataPoint) string {
	var lines []string
	lines = append(lines, "Date,RemainingPoints,RemainingTasks,TotalPoints,TotalTasks")

	for _, dp := range dataPoints {
		totalPoints := ""
		if dp.TotalPoints != nil {
			totalPoints = fmt.Sprintf("%d", *dp.TotalPoints)
		}
		totalTasks := ""
		if dp.TotalTasks != nil {
			totalTasks = fmt.Sprintf("%d", *dp.TotalTasks)
		}

		line := fmt.Sprintf("%s,%d,%d,%s,%s",
			dp.Date.Format("2006-01-02"),
			dp.RemainingPoints,
			dp.RemainingTasks,
			totalPoints,
			totalTasks,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
