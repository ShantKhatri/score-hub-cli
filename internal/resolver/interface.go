// Package resolver defines the abstract interface for resolving provisioner data.
package resolver

import (
	"context"
	"errors"

	"github.com/score-hub/cli/internal/index"
)

var (
	ErrNotFound         = errors.New("provisioner not found")
	ErrVariantNotFound  = errors.New("variant not found")
	ErrPlatformNotFound = errors.New("platform not supported for this variant")
	ErrChecksumMismatch = errors.New("checksum verification failed")
)

// Resolver is the abstraction for accessing provisioner data.
type Resolver interface {
	GetIndex(ctx context.Context, noCache bool) (*index.Index, error)
	FetchFile(ctx context.Context, path string) ([]byte, error)
}
