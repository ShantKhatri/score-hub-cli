package registry

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
)

type Manager struct {
	resolvers map[string]resolver.Resolver // keyed by alias
	urls      map[string]string            // registry alias -> URL
	order     []string                     // stable order: "public" first, then alphabetical
	default_  string                       // default alias for unqualified lookups
}

func (m *Manager) ResolverFor(alias string) (resolver.Resolver, error) {
	if alias == "" {
		alias = m.default_
	}
	r, ok := m.resolvers[alias]
	if !ok {
		return nil, fmt.Errorf("registry %q not found. Run 'score-hub registry list' to see configured registries", alias)
	}
	return r, nil
}

func (m *Manager) DefaultAlias() string {
	return m.default_
}
func (m *Manager) Aliases() []string {
	return m.order
}

func (m *Manager) URLForAlias(alias string) (string, bool) {
	url, ok := m.urls[alias]
	return url, ok
}

type TaggedResult struct {
	resolver.ProvisionerSummary
	RegistryAlias string
	RegistryURL   string
}

func (m *Manager) SearchAll(ctx context.Context, query string, opts resolver.SearchOpts) ([]TaggedResult, error) {
	type result struct {
		alias   string
		results []resolver.ProvisionerSummary
		err     error
	}

	ch := make(chan result, len(m.order))
	var wg sync.WaitGroup

	for _, alias := range m.order {
		alias := alias
		r := m.resolvers[alias]
		wg.Add(1)
		go func() {
			defer wg.Done()
			summaries, err := r.Search(ctx, query, opts)
			ch <- result{alias: alias, results: summaries, err: err}
		}()
	}

	wg.Wait()
	close(ch)

	var merged []TaggedResult
	for res := range ch {
		if res.err != nil {
			fmt.Fprintf(os.Stderr, "Warning: registry %q unreachable: %v\n", res.alias, res.err)
			continue
		}
		url, _ := m.URLForAlias(res.alias)
		for _, s := range res.results {
			merged = append(merged, TaggedResult{
				ProvisionerSummary: s,
				RegistryAlias:      res.alias,
				RegistryURL:        url,
			})
		}
	}

	return merged, nil
}

func (m *Manager) FindAcrossRegistries(ctx context.Context, name string) (map[string]*index.Provisioner, error) {
	found := make(map[string]*index.Provisioner)
	for alias, r := range m.resolvers {
		p, err := r.Resolve(ctx, name)
		if err == resolver.ErrNotFound {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("registry %q: %w", alias, err)
		}
		found[alias] = p
	}
	return found, nil
}
