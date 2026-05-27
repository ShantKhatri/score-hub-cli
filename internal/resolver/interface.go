// Package resolver defines the abstract interface for resolving provisioner data.
// This is the CRITICAL abstraction that allows score-hub to work with different
// backends (GitHub, local filesystem, HTTP API) without changing CLI commands.
//
// v0.1 ships with a GitHub-backed implementation. v0.2 adds local filesystem.
// v0.3 adds HTTP API for self-hosted registries.
package resolver

import (
	"context"

	"github.com/score-hub/cli/internal/index"
)

// Resolver is the abstraction for accessing provisioner data
type Resolver interface {
	GetIndex(ctx context.Context, noCache bool) (*index.Index, error)
	FetchFile(ctx context.Context, path string) ([]byte, error)
}
