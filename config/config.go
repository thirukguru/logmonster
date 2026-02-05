// Package config provides configuration loading and management.
package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	ScanPaths       []string      `mapstructure:"scan_paths"`
	ExcludePatterns []string      `mapstructure:"exclude_patterns"`
	Scan            ScanConfig    `mapstructure:"scan"`
	Thresholds      Thresholds    `mapstructure:"thresholds"`
	Display         DisplayConfig `mapstructure:"display"`
	Actions         ActionsConfig `mapstructure:"actions"`
}

// ScanConfig holds scan-related configuration.
type ScanConfig struct {
	Interval       int  `mapstructure:"interval"`
	MaxDepth       int  `mapstructure:"max_depth"`
	FollowSymlinks bool `mapstructure:"follow_symlinks"`
}

// Thresholds holds threshold configuration.
type Thresholds struct {
	GrowthMB     float64 `mapstructure:"growth_mb"`
	RateMBPerSec float64 `mapstructure:"rate_mb_per_sec"`
}

// DisplayConfig holds display-related configuration.
type DisplayConfig struct {
	TopN      int  `mapstructure:"top_n"`
	UseColors bool `mapstructure:"use_colors"`
}

// ActionsConfig holds action-related configuration.
type ActionsConfig struct {
	KillTimeout        int  `mapstructure:"kill_timeout"`
	ConfirmDestructive bool `mapstructure:"confirm_destructive"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		ScanPaths:       []string{"/var/log", "/tmp"},
		ExcludePatterns: []string{"*.gz", "*.zip", "*.bz2", "*.xz"},
		Scan: ScanConfig{
			Interval:       5,
			MaxDepth:       10,
			FollowSymlinks: false,
		},
		Thresholds: Thresholds{
			GrowthMB:     10,
			RateMBPerSec: 1.0,
		},
		Display: DisplayConfig{
			TopN:      10,
			UseColors: true,
		},
		Actions: ActionsConfig{
			KillTimeout:        5,
			ConfirmDestructive: true,
		},
	}
}

// Load loads configuration from file and environment.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Config file locations
	homeDir, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(homeDir, ".logmonster"))
	}
	viper.AddConfigPath("/etc/logmonster")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("scan_paths", cfg.ScanPaths)
	viper.SetDefault("exclude_patterns", cfg.ExcludePatterns)
	viper.SetDefault("scan.interval", cfg.Scan.Interval)
	viper.SetDefault("scan.max_depth", cfg.Scan.MaxDepth)
	viper.SetDefault("scan.follow_symlinks", cfg.Scan.FollowSymlinks)
	viper.SetDefault("thresholds.growth_mb", cfg.Thresholds.GrowthMB)
	viper.SetDefault("thresholds.rate_mb_per_sec", cfg.Thresholds.RateMBPerSec)
	viper.SetDefault("display.top_n", cfg.Display.TopN)
	viper.SetDefault("display.use_colors", cfg.Display.UseColors)
	viper.SetDefault("actions.kill_timeout", cfg.Actions.KillTimeout)
	viper.SetDefault("actions.confirm_destructive", cfg.Actions.ConfirmDestructive)

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal to struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetScanInterval returns the scan interval as a duration.
func (c *Config) GetScanInterval() time.Duration {
	return time.Duration(c.Scan.Interval) * time.Second
}

// GetThresholdBytes returns the threshold in bytes.
func (c *Config) GetThresholdBytes() int64 {
	return int64(c.Thresholds.GrowthMB * 1024 * 1024)
}

// GetKillTimeout returns the kill timeout as a duration.
func (c *Config) GetKillTimeout() time.Duration {
	return time.Duration(c.Actions.KillTimeout) * time.Second
}
