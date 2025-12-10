package ui

import (
	"strings"
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestRenderStorySections_BuildsTable(t *testing.T) {
	title := "A very long title that should be truncated to fit within the table width"
	stories := []DisplayStory{
		{
			Source: "Sprint",
			Story: &core.Story{
				ID:       "US-001",
				Title:    title,
				Priority: core.PriorityHigh,
				Status:   core.StatusDoing,
			},
			Status: core.StatusDoing,
		},
	}

	output := RenderStorySections(map[string][]DisplayStory{"Sprint": stories})
	if output == "" {
		t.Fatalf("expected rendered output, got empty string")
	}
	if !containsAll(output, []string{"ID", "Title", "Status", "US-001", "╭"}) {
		t.Fatalf("output missing expected headers/values: %s", output)
	}
}

func TestRenderStorySections_EmptyReturnsEmpty(t *testing.T) {
	output := RenderStorySections(map[string][]DisplayStory{})
	if output != "" {
		t.Fatalf("expected empty output for empty sections, got %s", output)
	}
}

func containsAll(haystack string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}
