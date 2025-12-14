package unit

import (
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

func TestCalculateEndDate(t *testing.T) {
	baseDate := time.Date(2025, 1, 27, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		startDate time.Time
		duration  string
		wantDate  time.Time
		wantErr   bool
	}{
		{
			name:      "2 weeks from start",
			startDate: baseDate,
			duration:  "2w",
			wantDate:  time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "1 week from start",
			startDate: baseDate,
			duration:  "1w",
			wantDate:  time.Date(2025, 2, 3, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "14 days from start",
			startDate: baseDate,
			duration:  "14d",
			wantDate:  time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "30 days from start",
			startDate: baseDate,
			duration:  "30d",
			wantDate:  time.Date(2025, 2, 26, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "default duration (2 weeks)",
			startDate: baseDate,
			duration:  "",
			wantDate:  time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "month boundary crossing",
			startDate: time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC),
			duration:  "14d",
			wantDate:  time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "year boundary crossing",
			startDate: time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC),
			duration:  "14d",
			wantDate:  time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "invalid duration",
			startDate: baseDate,
			duration:  "invalid",
			wantDate:  time.Time{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, err := core.CalculateEndDate(tt.startDate, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateEndDate(%v, %q) error = %v, wantErr %v", tt.startDate, tt.duration, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !gotDate.Equal(tt.wantDate) {
				t.Errorf("CalculateEndDate(%v, %q) = %v, want %v", tt.startDate, tt.duration, gotDate, tt.wantDate)
			}
		})
	}
}
