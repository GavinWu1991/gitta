package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    core.SprintStatus
		to      core.SprintStatus
		wantErr bool
		errMsg  string
	}{
		// Valid transitions
		{"Planning to Active", core.StatusPlanning, core.StatusActive, false, ""},
		{"Ready to Active", core.StatusReady, core.StatusActive, false, ""},
		{"Active to Archived", core.StatusActive, core.StatusArchived, false, ""},
		{"Same status (no-op)", core.StatusActive, core.StatusActive, false, ""},
		{"Planning to Archived", core.StatusPlanning, core.StatusArchived, false, ""},
		{"Ready to Archived", core.StatusReady, core.StatusArchived, false, ""},

		// Invalid transitions
		{"Archived to Active", core.StatusArchived, core.StatusActive, true, "cannot transition from archived"},
		{"Archived to Ready", core.StatusArchived, core.StatusReady, true, "cannot transition from archived"},
		{"Archived to Planning", core.StatusArchived, core.StatusPlanning, true, "cannot transition from archived"},
		{"Active to Ready", core.StatusActive, core.StatusReady, true, "cannot transition from active to ready"},
		{"Active to Planning", core.StatusActive, core.StatusPlanning, true, "cannot transition from active to planning"},
		{"Ready to Planning", core.StatusReady, core.StatusPlanning, true, "invalid transition"},
		{"Planning to Ready", core.StatusPlanning, core.StatusReady, true, "invalid transition"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := core.ValidateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransition(%v, %v) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("ValidateTransition(%v, %v) expected error message containing %q, got %v", tt.from, tt.to, tt.errMsg, err)
				} else if err != nil && !errorContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTransition(%v, %v) error message = %q, want containing %q", tt.from, tt.to, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func errorContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
