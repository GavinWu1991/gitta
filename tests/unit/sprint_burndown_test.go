package unit

import (
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

func TestBurndownCalculation(t *testing.T) {
	t.Skip("TODO: implement burndown calculation tests (T058)")

	_ = []core.BurndownDataPoint{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), RemainingTasks: 5},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), RemainingTasks: 3},
	}
}
