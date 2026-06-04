// Package config manages score-hub configuration from ~/.score-hub/config.yaml
// and environment variables.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the score-hub configuration
type Config struct {
	APIVersion string      `yaml:"apiVersion"`
	Registry   string      `yaml:"registry"`
	Platform   string      `yaml:"platform"`
	Cache      CacheConfig `yaml:"cache"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	TTL string `yaml:"ttl"`
	Dir string `yaml:"dir"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir := ScoreHubDir()
	return &Config{
		APIVersion: "score-hub/v1alpha1",
		Registry:   "https://raw.githubusercontent.com/ShantKhatri/score-hub-cli/main/cmd/index.yaml",
		Platform:   "auto",
		Cache: CacheConfig{
			TTL: "1h",
			Dir: filepath.Join(homeDir, "cache"),
		},
	}
}

// ScoreHubDir returns the score-hub config directory (~/.score-hub)
func ScoreHubDir() string {
	if dir := os.Getenv("SCORE_HUB_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".score-hub"
	}
	return filepath.Join(home, ".score-hub")
}

// CacheDir returns the cache directory
func CacheDir() string {
	return filepath.Join(ScoreHubDir(), "cache")
}

// Load reads configuration from ~/.score-hub/config.yaml
func Load() *Config {
	cfg := DefaultConfig()

	configPath := filepath.Join(ScoreHubDir(), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.Registry != "" {
		cfg.Registry = fileCfg.Registry
	}
	if fileCfg.Platform != "" {
		cfg.Platform = fileCfg.Platform
	}
	if fileCfg.Cache.TTL != "" {
		cfg.Cache.TTL = fileCfg.Cache.TTL
	}
	if fileCfg.Cache.Dir != "" {
		cfg.Cache.Dir = fileCfg.Cache.Dir
	}

	// Environment variable overrides
	if registry := os.Getenv("SCORE_HUB_REGISTRY"); registry != "" {
		cfg.Registry = registry
	}
	if platform := os.Getenv("SCORE_HUB_PLATFORM"); platform != "" {
		cfg.Platform = platform
	}

	return cfg
}
