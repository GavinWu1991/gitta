package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration.
type Config struct {
	// Add configuration fields as needed
	LogLevel string `mapstructure:"log_level"`
	DataDir  string `mapstructure:"data_dir"`
}

// Load reads configuration from files, environment variables, and defaults.
// Configuration is loaded in this order (later sources override earlier):
// 1. Defaults
// 2. Config file (`.gitta/config.yaml`)
// 3. Environment variables (GITTA_*)
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("data_dir", ".gitta")

	// Set config file name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default: .gitta/config.yaml
		v.AddConfigPath(".gitta")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvPrefix("GITTA")
	v.AutomaticEnv()

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK - use defaults/env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Resolve relative paths
	if !filepath.IsAbs(cfg.DataDir) {
		cfg.DataDir = filepath.Join(".gitta", cfg.DataDir)
	}

	return &cfg, nil
}
