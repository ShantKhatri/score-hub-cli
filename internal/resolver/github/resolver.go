// Package github implements the Resolver interface backed by GitHub raw content.
package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/score-hub/cli/internal/index"
)

const (
	DefaultIndexURL        = "https://raw.githubusercontent.com/score-hub/index/main/index.yaml"
	DefaultUpstreamBaseURL = "https://raw.githubusercontent.com/score-spec/community-provisioners/refs/heads/main/"
	DefaultCacheTTL        = 1 * time.Hour

	maxDownloadSize = 10 * 1024 * 1024
)

// GitHubResolver implements Resolver using GitHub raw content URLs.
type GitHubResolver struct {
	IndexURL        string
	UpstreamBaseURL string
	CacheDir        string
	CacheTTL        time.Duration
	HTTPClient      *http.Client

	// EmbeddedIndex is the index compiled into the binary via go:embed.
	// Used as last-resort fallback when remote and cache both fail.
	EmbeddedIndex []byte

	// Verbose enables debug-level logging to stderr.
	Verbose bool
}

// NewGitHubResolver returns a resolver with default settings.
func NewGitHubResolver(cacheDir string) *GitHubResolver {
	return &GitHubResolver{
		IndexURL:        DefaultIndexURL,
		UpstreamBaseURL: DefaultUpstreamBaseURL,
		CacheDir:        cacheDir,
		CacheTTL:        DefaultCacheTTL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIndex fetches and parses the provisioner index.
func (r *GitHubResolver) GetIndex(ctx context.Context, noCache bool) (*index.Index, error) {
	cachePath := filepath.Join(r.CacheDir, "index.yaml")

	if !noCache {
		if data, err := r.readCache(cachePath); err == nil {
			r.debugf("Using cached index from %s", cachePath)
			return index.Parse(data)
		}
	}

	r.debugf("Fetching index from %s", r.IndexURL)
	data, err := r.httpGet(ctx, r.IndexURL)
	if err == nil {
		// Cache the fresh index
		if cacheErr := r.writeCache(cachePath, data); cacheErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cache index: %v\n", cacheErr)
		}
		return index.Parse(data)
	}

	r.debugf("Remote fetch failed: %v", err)

	if cachedData, cacheErr := os.ReadFile(cachePath); cacheErr == nil {
		fmt.Fprintf(os.Stderr, "Warning: using cached index (network error: %v)\n", err)
		return index.Parse(cachedData)
	}

	if r.EmbeddedIndex != nil {
		r.debugf("Using embedded index as fallback")
		fmt.Fprintf(os.Stderr, "Warning: using embedded index (network error: %v)\n", err)
		return index.Parse(r.EmbeddedIndex)
	}

	return nil, fmt.Errorf("failed to fetch index: %w (no cache or embedded index available)", err)
}

// FetchFile downloads a provisioner file from the upstream source.
func (r *GitHubResolver) FetchFile(ctx context.Context, path string) ([]byte, error) {
	url := r.UpstreamBaseURL + path
	r.debugf("Fetching file from %s", url)
	return r.httpGet(ctx, url)
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
		return fmt.Errorf("checksum mismatch:\n  expected: %s\n  actual:   sha256:%s", expected, actualHex)
	}

	return nil
}

// ComputeChecksum computes SHA256 checksum in "sha256:hex" format.
func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// httpGet performs a bounded HTTP GET request.
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

// readCache retrieves cached data if not expired.
func (r *GitHubResolver) readCache(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > r.CacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	return os.ReadFile(path)
}

// writeCache saves data to cache directory.
func (r *GitHubResolver) writeCache(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// debugf prints debug output when verbose mode is enabled.
func (r *GitHubResolver) debugf(format string, args ...interface{}) {
	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}
