package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/cache"
	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
)

const (
	DefaultIndexURL        = "https://raw.githubusercontent.com/ShantKhatri/score-hub-cli/main/cmd/index.yaml"
	DefaultUpstreamBaseURL = "https://raw.githubusercontent.com/score-spec/community-provisioners/main/"

	maxDownloadSize = 10 * 1024 * 1024
)

type GitHubResolver struct {
	IndexURL        string
	UpstreamBaseURL string
	HTTPClient      *http.Client
	Cache           *cache.Cache

	// Used as last-resort fallback when remote and cache both fail.
	EmbeddedIndex []byte

	// Verbose enables debug-level logging to stderr.
	Verbose bool
}

func NewGitHubResolver(c *cache.Cache) *GitHubResolver {
	return &GitHubResolver{
		IndexURL:        DefaultIndexURL,
		UpstreamBaseURL: DefaultUpstreamBaseURL,
		Cache:           c,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *GitHubResolver) Index(ctx context.Context) (*index.Index, error) {
	cacheKey := "index:" + r.IndexURL

	// Try cache first
	if r.Cache != nil {
		if data, err := r.Cache.Get(cacheKey); err == nil {
			r.debugf("Using cached index for %s", r.IndexURL)
			return index.Parse(data)
		}
	}

	r.debugf("Fetching index from %s", r.IndexURL)
	data, err := r.httpGet(ctx, r.IndexURL)
	if err == nil {
		// Cache the fresh index
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

	if r.EmbeddedIndex != nil {
		r.debugf("Using embedded index as fallback")
		fmt.Fprintf(os.Stderr, "Warning: using embedded index (network error: %v)\n", err)
		return index.Parse(r.EmbeddedIndex)
	}

	return nil, fmt.Errorf("failed to fetch index: %w (no cache or embedded index available)", err)
}

func (r *GitHubResolver) Search(ctx context.Context, query string, opts resolver.SearchOpts) ([]resolver.ProvisionerSummary, error) {
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

// Resolve looks up a provisioner by exact name.
func (r *GitHubResolver) Resolve(ctx context.Context, name string) (*index.Provisioner, error) {
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

func (r *GitHubResolver) ResolveVersion(ctx context.Context, name, version string) (*index.Provisioner, error) {
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

func (r *GitHubResolver) FetchFile(ctx context.Context, p *index.Provisioner, variant, platform string) ([]byte, string, error) {
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

	// Determine the URL to fetch from
	var fileURL string
	if plat.DownloadURL != "" {
		fileURL = plat.DownloadURL
	} else if plat.Path != "" {
		fileURL = r.UpstreamBaseURL + plat.Path
	} else {
		return nil, "", fmt.Errorf("provisioner %s/%s/%s: no downloadURL or path configured",
			p.Name, variant, platform)
	}

	r.debugf("Fetching file from %s", fileURL)
	data, err := r.httpGet(ctx, fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	checksum := ComputeChecksum(data)
	return data, checksum, nil
}

// VerifyChecksum verifies SHA256 checksum in "sha256:hex" format.
func VerifyChecksum(data []byte, expected string) error {
	if len(expected) < 8 || expected[:7] != "sha256:" {
		return fmt.Errorf("invalid checksum format: %s (expected sha256:<hex>)", expected)
	}

	expectedHex := expected[7:]
	hash := sha256.Sum256(data)
	actualHex := hex.EncodeToString(hash[:])

	if actualHex != expectedHex {
		return fmt.Errorf("%w:\n  expected: %s\n  actual:   sha256:%s",
			resolver.ErrChecksumMismatch, expected, actualHex)
	}

	return nil
}

func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func (r *GitHubResolver) httpGet(ctx context.Context, url string) ([]byte, error) {
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

	// Limit read size to prevent memory exhaustion
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func (r *GitHubResolver) debugf(format string, args ...interface{}) {
	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}

// If the value starts with http:// or https://, it's treated as a direct URL.
// Otherwise, it's treated as an alias, but GitHubResolver doesn't handle aliases,
// so this returns the value as-is for the caller to resolve.
func ResolveRegistryFlag(value string) (url string, isAlias bool) {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value, false
	}
	return value, true
}
