package cmd

import (
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/registry"
)

func getManager() (*registry.Manager, error) {
	cfg := config.Load()

	// If the user provided a direct URL via --registry, we inject it into the
	// configuration as an ephemeral registry so the Manager can use it.
	if flagRegistry != "" && (strings.HasPrefix(flagRegistry, "http://") || strings.HasPrefix(flagRegistry, "https://")) {
		if cfg.Registries == nil {
			cfg.Registries = make(map[string]config.RegistryEntry)
		}
		cfg.Registries["_ephemeral"] = config.RegistryEntry{URL: flagRegistry}
	}

	m, err := registry.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// It returns the alias to use, or "_ephemeral" if a direct URL was provided.
func resolveRegistryFlag(flagRegistry string) string {
	if flagRegistry == "" {
		return ""
	}
	if strings.HasPrefix(flagRegistry, "http://") || strings.HasPrefix(flagRegistry, "https://") {
		return "_ephemeral"
	}
	return flagRegistry
}
