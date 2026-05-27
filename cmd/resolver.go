package cmd

import (
	"github.com/score-hub/cli/internal/config"
	"github.com/score-hub/cli/internal/resolver"
	gh "github.com/score-hub/cli/internal/resolver/github"
)

// getResolver creates a resolver from config and flags
func getResolver() (resolver.Resolver, error) {
	cfg := config.Load()
	res := gh.NewGitHubResolver(config.CacheDir())

	if flagRegistry != "" {
		res.IndexURL = flagRegistry
	} else if cfg.Registry != "" {
		res.IndexURL = cfg.Registry
	}

	res.EmbeddedIndex = getEmbeddedIndex()
	return res, nil
}
