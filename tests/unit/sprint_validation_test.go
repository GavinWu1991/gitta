package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestValidateSprintName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid sprint name",
			input:   "Sprint-01",
			wantErr: false,
		},
		{
			name:    "valid sprint name with larger number",
			input:   "Sprint-42",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "name with forward slash",
			input:   "Sprint/01",
			wantErr: true,
		},
		{
			name:    "name with backslash",
			input:   "Sprint\\01",
			wantErr: true,
		},
		{
			name:    "name with parent directory",
			input:   "Sprint-../01",
			wantErr: true,
		},
		{
			name:    "name with double dot",
			input:   "Sprint-..01",
			wantErr: true,
		},
		{
			name:    "simple name without numbers",
			input:   "MySprint",
			wantErr: false,
		},
		{
			name:    "name with spaces",
			input:   "Sprint 01",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := core.ValidateSprintName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSprintName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
