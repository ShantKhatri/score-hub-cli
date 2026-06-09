package registry

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/cache"
	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
	gh "github.com/ShantKhatri/score-hub-cli/internal/resolver/github"
	httppkg "github.com/ShantKhatri/score-hub-cli/internal/resolver/http"
)

func NewManager(cfg *config.Config) (*Manager, error) {
	c := cache.New(config.CacheDir(), 1*time.Hour)

	m := &Manager{
		resolvers: make(map[string]resolver.Resolver),
		urls:      make(map[string]string),
		default_:  cfg.DefaultRegistry,
	}

	for alias, entry := range cfg.Registries {
		r, err := newResolver(entry, alias, c)
		if err != nil {
			return nil, fmt.Errorf("registry %q: %w", alias, err)
		}
		m.resolvers[alias] = r
		m.urls[alias] = entry.URL
		m.order = append(m.order, alias)
	}

	// Stable order: "public" first, then alphabetical.
	sortAliases(m.order)

	if _, ok := m.resolvers[m.default_]; !ok {
		return nil, fmt.Errorf("defaultRegistry %q is not in the registries map", m.default_)
	}

	return m, nil
}

func newResolver(entry config.RegistryEntry, alias string, c *cache.Cache) (resolver.Resolver, error) {
	url := entry.URL

	switch {
	case url == config.DefaultPublicIndexURL || strings.Contains(url, "score-spec/community-provisioners"):
		res := gh.NewGitHubResolver(c)
		res.IndexURL = url
		return res, nil

	case strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://"):
		res := httppkg.NewStaticResolver(c)
		res.IndexURL = url
		return res, nil

	case strings.HasPrefix(url, "file://") || strings.HasPrefix(url, "/") || strings.HasPrefix(url, "./"):
		return nil, fmt.Errorf("local file URLs are not yet supported (coming in Phase 2). URL: %s", url)

	default:
		return nil, fmt.Errorf("unsupported URL scheme: %q. Supported: https://, http://, file://, /path, ./path", url)
	}
}

func sortAliases(aliases []string) {
	sort.Slice(aliases, func(i, j int) bool {
		if aliases[i] == "public" {
			return true
		}
		if aliases[j] == "public" {
			return false
		}
		return aliases[i] < aliases[j]
	})
}
