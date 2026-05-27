package cmd

import (
	"github.com/score-hub/cli/internal/config"
	"github.com/score-hub/cli/internal/resolver"
	gh "github.com/score-hub/cli/internal/resolver/github"
)

func getResolver() (resolver.Resolver, error) {
	cfg := config.Load()
	res := gh.NewGitHubResolver(config.CacheDir())

	if flagRegistry != "" {
		res.IndexURL = flagRegistry
	} else if cfg.Registry != "" {
		res.IndexURL = cfg.Registry
	}

	res.EmbeddedIndex = getEmbeddedIndex()
	res.Verbose = flagVerbose
	return res, nil
}
