package services

import (
	"github.com/spf13/viper"
)

// StatusEngineConfig holds configuration for the StatusEngine service.
type StatusEngineConfig struct {
	// BranchPrefix is the prefix pattern for matching branches to story IDs.
	// Default: "feat/"
	BranchPrefix string `mapstructure:"branch_prefix"`

	// CaseSensitive determines whether branch name matching is case-sensitive.
	// Default: true (Git branch names are case-sensitive)
	CaseSensitive bool `mapstructure:"case_sensitive"`

	// TargetBranches are the branch names to check for merge status.
	// Default: ["main", "master"]
	TargetBranches []string `mapstructure:"target_branches"`
}

// loadConfig loads StatusEngine configuration from Viper with defaults.
// Configuration keys:
//   - branch.prefix: Branch naming prefix pattern (default: "feat/")
//   - branch.case_sensitive: Case sensitivity for matching (default: true)
//   - branch.target_branches: Target branches for merge check (default: ["main", "master"])
func loadConfig() StatusEngineConfig {
	v := viper.New()

	// Set defaults
	v.SetDefault("branch.prefix", "feat/")
	v.SetDefault("branch.case_sensitive", true)
	v.SetDefault("branch.target_branches", []string{"main", "master"})

	// Environment variables (GITTA_BRANCH_*)
	v.SetEnvPrefix("GITTA")
	v.AutomaticEnv()

	var cfg StatusEngineConfig
	if err := v.UnmarshalKey("branch", &cfg); err != nil {
		// If unmarshal fails, use defaults
		return StatusEngineConfig{
			BranchPrefix:   "feat/",
			CaseSensitive:  true,
			TargetBranches: []string{"main", "master"},
		}
	}

	// Ensure defaults if unmarshaling didn't set them
	if cfg.BranchPrefix == "" {
		cfg.BranchPrefix = "feat/"
	}
	if len(cfg.TargetBranches) == 0 {
		cfg.TargetBranches = []string{"main", "master"}
	}

	// Validation: Ensure target branches are not empty
	if len(cfg.TargetBranches) > 0 {
		// Filter out empty strings
		validTargets := make([]string, 0, len(cfg.TargetBranches))
		for _, target := range cfg.TargetBranches {
			if target != "" {
				validTargets = append(validTargets, target)
			}
		}
		if len(validTargets) > 0 {
			cfg.TargetBranches = validTargets
		} else {
			cfg.TargetBranches = []string{"main", "master"}
		}
	}

	return cfg
}
