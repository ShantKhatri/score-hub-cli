package cmd

import (
	"strings"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/cache"
	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
	gh "github.com/ShantKhatri/score-hub-cli/internal/resolver/github"
)

func getResolver() (resolver.Resolver, error) {
	cfg := config.Load()

	c := cache.New(config.CacheDir(), 1*time.Hour)
	res := gh.NewGitHubResolver(c)

	if flagRegistry != "" {
		if strings.HasPrefix(flagRegistry, "http://") || strings.HasPrefix(flagRegistry, "https://") {
			res.IndexURL = flagRegistry
		} else {
			regURL, err := cfg.RegistryURL(flagRegistry)
			if err != nil {
				return nil, err
			}
			res.IndexURL = regURL
		}
	}

	res.EmbeddedIndex = getEmbeddedIndex()
	res.Verbose = flagVerbose
	return res, nil
}
