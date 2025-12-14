package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDays int
		wantErr  bool
	}{
		{
			name:     "empty duration defaults to 2 weeks",
			input:    "",
			wantDays: 14,
			wantErr:  false,
		},
		{
			name:     "1 week",
			input:    "1w",
			wantDays: 7,
			wantErr:  false,
		},
		{
			name:     "2 weeks",
			input:    "2w",
			wantDays: 14,
			wantErr:  false,
		},
		{
			name:     "4 weeks",
			input:    "4w",
			wantDays: 28,
			wantErr:  false,
		},
		{
			name:     "uppercase weeks",
			input:    "2W",
			wantDays: 14,
			wantErr:  false,
		},
		{
			name:     "14 days",
			input:    "14d",
			wantDays: 14,
			wantErr:  false,
		},
		{
			name:     "30 days",
			input:    "30d",
			wantDays: 30,
			wantErr:  false,
		},
		{
			name:     "uppercase days",
			input:    "14D",
			wantDays: 14,
			wantErr:  false,
		},
		{
			name:     "invalid format - missing unit",
			input:    "14",
			wantDays: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - missing number",
			input:    "w",
			wantDays: 0,
			wantErr:  true,
		},
		{
			name:     "invalid unit",
			input:    "14m",
			wantDays: 0,
			wantErr:  true,
		},
		{
			name:     "zero days",
			input:    "0d",
			wantDays: 0,
			wantErr:  true,
		},
		{
			name:     "negative days",
			input:    "-1d",
			wantDays: 0,
			wantErr:  true,
		},
		{
			name:     "zero weeks",
			input:    "0w",
			wantDays: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDays, err := core.ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotDays != tt.wantDays {
				t.Errorf("ParseDuration(%q) = %d days, want %d days", tt.input, gotDays, tt.wantDays)
			}
		})
	}
}
