package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/cache"
	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
)

const maxDownloadSize = 10 * 1024 * 1024

type StaticResolver struct {
	IndexURL   string
	HTTPClient *http.Client
	Cache      *cache.Cache
	Verbose    bool
}

func NewStaticResolver(c *cache.Cache) *StaticResolver {
	return &StaticResolver{
		Cache: c,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *StaticResolver) Index(ctx context.Context) (*index.Index, error) {
	cacheKey := "index:" + r.IndexURL

	if r.Cache != nil {
		if data, err := r.Cache.Get(cacheKey); err == nil {
			r.debugf("Using cached index for %s", r.IndexURL)
			return index.Parse(data)
		}
	}

	r.debugf("Fetching index from %s", r.IndexURL)
	data, err := r.httpGet(ctx, r.IndexURL)
	if err == nil {
		if r.Cache != nil {
			if cacheErr := r.Cache.Set(cacheKey, data); cacheErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cache index: %v\n", cacheErr)
			}
		}
		return index.Parse(data)
	}

	r.debugf("Remote fetch failed: %v", err)

	if r.Cache != nil {
		if cachedData, cacheErr := r.Cache.Get(cacheKey); cacheErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: using cached index (network error: %v)\n", err)
			return index.Parse(cachedData)
		}
	}

	return nil, fmt.Errorf("failed to fetch index: %w (no cache available)", err)
}

func (r *StaticResolver) Search(ctx context.Context, query string, opts resolver.SearchOpts) ([]resolver.ProvisionerSummary, error) {
	idx, err := r.Index(ctx)
	if err != nil {
		return nil, err
	}

	results := idx.Search(query, opts.Category, opts.Platform)

	summaries := make([]resolver.ProvisionerSummary, len(results))
	for i, p := range results {
		summaries[i] = resolver.ProvisionerSummary{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Category:    p.Category,
			Platforms:   p.PlatformNames(),
			Variants:    p.VariantCount(),
			Version:     p.LatestVersion(),
			Description: p.Description,
		}
	}
	return summaries, nil
}

func (r *StaticResolver) Resolve(ctx context.Context, name string) (*index.Provisioner, error) {
	idx, err := r.Index(ctx)
	if err != nil {
		return nil, err
	}

	p := idx.FindProvisioner(name)
	if p == nil {
		return nil, resolver.ErrNotFound
	}
	return p, nil
}

func (r *StaticResolver) ResolveVersion(ctx context.Context, name, version string) (*index.Provisioner, error) {
	p, err := r.Resolve(ctx, name)
	if err != nil {
		return nil, err
	}

	for _, v := range p.Versions {
		if v.Version == version {
			return p, nil
		}
	}
	return nil, fmt.Errorf("version %q not found for provisioner %q", version, name)
}

func (r *StaticResolver) FetchFile(ctx context.Context, p *index.Provisioner, variant, platform string) ([]byte, string, error) {
	v := p.FindVariant(variant)
	if v == nil {
		return nil, "", fmt.Errorf("%w: variant %q not found for provisioner %q",
			resolver.ErrVariantNotFound, variant, p.Name)
	}

	plat, ok := v.Platforms[platform]
	if !ok {
		return nil, "", fmt.Errorf("%w: platform %q not supported for variant %q of %q",
			resolver.ErrPlatformNotFound, platform, variant, p.Name)
	}

	if plat.DownloadURL == "" {
		return nil, "", fmt.Errorf("provisioner %s/%s/%s: downloadURL is required for static HTTP registries", p.Name, variant, platform)
	}

	r.debugf("Fetching file from %s", plat.DownloadURL)
	data, err := r.httpGet(ctx, plat.DownloadURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	hash := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(hash[:])
	return data, checksum, nil
}

func (r *StaticResolver) httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "score-hub/0.1.0")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func (r *StaticResolver) debugf(format string, args ...interface{}) {
	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}
