package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gavin/gitta/internal/version"
)

func TestVersionFormatter(t *testing.T) {
	tests := []struct {
		name     string
		info     version.Info
		jsonMode bool
		want     string
		wantJSON map[string]interface{}
	}{
		{
			name: "human readable format",
			info: version.Info{
				Version:   "0.1.0",
				Commit:    "abc123",
				BuildDate: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				GoVersion: "go1.21.5",
			},
			jsonMode: false,
			want:     "gitta v0.1.0\ncommit: abc123\nbuild date: 2024-01-15T10:30:00Z\ngo: go1.21.5\n",
		},
		{
			name: "JSON format",
			info: version.Info{
				Version:   "0.1.0",
				Commit:    "abc123",
				BuildDate: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				GoVersion: "go1.21.5",
			},
			jsonMode: true,
			wantJSON: map[string]interface{}{
				"version":   "0.1.0",
				"commit":    "abc123",
				"buildDate": "2024-01-15T10:30:00Z",
				"goVersion": "go1.21.5",
			},
		},
		{
			name: "missing commit",
			info: version.Info{
				Version:   "0.1.0",
				Commit:    "",
				BuildDate: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				GoVersion: "go1.21.5",
			},
			jsonMode: false,
			want:     "gitta v0.1.0\ncommit: <unknown>\nbuild date: 2024-01-15T10:30:00Z\ngo: go1.21.5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.jsonMode {
				jsonStr, err := tt.info.FormatJSON()
				if err != nil {
					t.Fatalf("failed to format JSON: %v", err)
				}
				result = jsonStr

				// Verify JSON structure
				var gotJSON map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &gotJSON); err != nil {
					t.Fatalf("failed to unmarshal JSON: %v", err)
				}

				// Compare key fields
				for key, wantVal := range tt.wantJSON {
					gotVal, ok := gotJSON[key]
					if !ok {
						t.Errorf("missing JSON key: %s", key)
						continue
					}
					if gotVal != wantVal {
						t.Errorf("JSON[%s] = %v, want %v", key, gotVal, wantVal)
					}
				}
			} else {
				result = tt.info.Format()
				if result != tt.want {
					t.Errorf("Format() = %q, want %q", result, tt.want)
				}
			}
		})
	}
}
