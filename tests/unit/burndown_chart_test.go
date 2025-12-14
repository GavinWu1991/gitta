package unit

import (
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

func TestBurndownASCIIChart(t *testing.T) {
	t.Skip("TODO: implement ASCII chart rendering tests (T059/T060)")

	_ = []core.BurndownDataPoint{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), RemainingTasks: 5},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), RemainingTasks: 3},
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), RemainingTasks: 1},
	}
}
