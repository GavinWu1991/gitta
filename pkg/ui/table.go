package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gavin/gitta/internal/core"
)

// DisplayStory is a view model for rendering a story row.
type DisplayStory struct {
	Source   string
	Story    *core.Story
	Priority core.Priority
	Status   core.Status
}

// RenderStorySections renders one or more sections (e.g., Sprint vs Backlog) with
// lipgloss styling. Each section gets its own bordered table.
func RenderStorySections(sections map[string][]DisplayStory) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true)

	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1)

	var blocks []string
	keys := make([]string, 0, len(sections))
	for k := range sections {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, title := range keys {
		stories := sections[title]
		if len(stories) == 0 {
			continue
		}
		table := renderTable(stories)
		blocks = append(blocks, sectionStyle.Render(title))
		blocks = append(blocks, tableStyle.Render(table))
	}

	// If no sections had content, return empty string so caller can handle empty state.
	if len(blocks) == 0 {
		return ""
	}

	return strings.Join(blocks, "\n\n")
}

const (
	idWidth       = 10
	titleWidth    = 40
	statusWidth   = 12
	assigneeWidth = 20
	priorityWidth = 10
)

func renderTable(stories []DisplayStory) string {
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		idWidth, "ID",
		titleWidth, "Title",
		statusWidth, "Status",
		assigneeWidth, "Assignee",
		priorityWidth, "Priority",
	)

	var rows []string
	for _, item := range stories {
		row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
			idWidth, truncate(item.Story.ID, idWidth),
			titleWidth, truncate(item.Story.Title, titleWidth),
			statusWidth, statusStyle(item.Status),
			assigneeWidth, truncate(valueOrEmpty(item.Story.Assignee), assigneeWidth),
			priorityWidth, truncate(string(item.Priority), priorityWidth),
		)
		rows = append(rows, row)
	}

	return strings.Join(append([]string{header}, rows...), "\n")
}

func truncate(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	return value[:width-1] + "…"
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func statusStyle(status core.Status) string {
	base := lipgloss.NewStyle().Bold(true)
	switch status {
	case core.StatusDoing:
		return base.Foreground(lipgloss.Color("178")).Render("Doing")
	case core.StatusReview:
		return base.Foreground(lipgloss.Color("75")).Render("Review")
	case core.StatusDone:
		return base.Foreground(lipgloss.Color("77")).Render("Done")
	default:
		return base.Foreground(lipgloss.Color("246")).Render("Todo")
	}
}
