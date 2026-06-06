package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIVersion string      `yaml:"apiVersion"`
	Platform   string      `yaml:"platform"`
	Cache      CacheConfig `yaml:"cache"`

	// Named registry map. Key is alias (e.g., "public", "myorg").
	Registries      map[string]RegistryEntry `yaml:"registries,omitempty"`
	DefaultRegistry string                   `yaml:"defaultRegistry,omitempty"`
}

type RegistryEntry struct {
	URL string `yaml:"url"`
}

type CacheConfig struct {
	TTL string `yaml:"ttl"`
	Dir string `yaml:"dir"`
}

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

const DefaultPublicIndexURL = "https://raw.githubusercontent.com/ShantKhatri/score-hub-cli/main/cmd/index.yaml"

func DefaultConfig() *Config {
	homeDir := ScoreHubDir()
	return &Config{
		APIVersion: "score-hub/v1alpha1",
		Platform:   "auto",
		Cache: CacheConfig{
			TTL: "1h",
			Dir: filepath.Join(homeDir, "cache"),
		},
	}
}

// (~/.score-hub)
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
		cfg.InitDefaults()
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		cfg.InitDefaults()
		return cfg
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

	if fileCfg.Registries != nil {
		cfg.Registries = fileCfg.Registries
	}
	if fileCfg.DefaultRegistry != "" {
		cfg.DefaultRegistry = fileCfg.DefaultRegistry
	}
	if platform := os.Getenv("SCORE_HUB_PLATFORM"); platform != "" {
		cfg.Platform = platform
	}

	cfg.InitDefaults()

	if envVal := os.Getenv("SCORE_HUB_REGISTRY_URLS"); envVal != "" {
		pairs := strings.Split(envVal, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				alias := strings.TrimSpace(parts[0])
				url := strings.TrimSpace(parts[1])
				if alias != "" && url != "" {
					cfg.Registries[alias] = RegistryEntry{URL: url}
				}
			}
		}
	}

	return cfg
}

func (c *Config) InitDefaults() {
	if c.Registries == nil {
		c.Registries = make(map[string]RegistryEntry)
	}

	if _, ok := c.Registries["public"]; !ok {
		c.Registries["public"] = RegistryEntry{
			URL: DefaultPublicIndexURL,
		}
	}

	if c.DefaultRegistry == "" {
		c.DefaultRegistry = "public"
	}
}

func (c *Config) AddRegistry(alias string, entry RegistryEntry) error {
	if !isValidAlias(alias) {
		return fmt.Errorf("invalid registry alias %q: must be alphanumeric, hyphens, or underscores", alias)
	}
	if c.Registries == nil {
		c.Registries = make(map[string]RegistryEntry)
	}
	c.Registries[alias] = entry
	return nil
}

func (c *Config) RemoveRegistry(alias string, force bool) error {
	if alias == "public" && !force {
		return fmt.Errorf("cannot remove the default public registry without --force")
	}
	if alias == c.DefaultRegistry && !force {
		return fmt.Errorf("cannot remove the default registry %q without --force. Change defaultRegistry first", alias)
	}
	delete(c.Registries, alias)
	return nil
}

func (c *Config) SetDefaultRegistry(alias string) error {
	if _, ok := c.Registries[alias]; !ok {
		return fmt.Errorf("registry %q not found in config", alias)
	}
	c.DefaultRegistry = alias
	return nil
}

func (c *Config) RegistryURL(alias string) (string, error) {
	entry, ok := c.Registries[alias]
	if !ok {
		return "", fmt.Errorf("registry %q not configured. Run 'score-hub registry list' to see configured registries", alias)
	}
	return entry.URL, nil
}

// to ~/.score-hub/config.yaml.
func (c *Config) Save() error {
	configDir := ScoreHubDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return os.WriteFile(configPath, data, 0644)
}

func isValidAlias(alias string) bool {
	return alias != "" && aliasRegex.MatchString(alias)
}
