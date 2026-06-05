// Defines the abstract interface for resolving provisioner data.
package resolver

import (
	"context"
	"errors"

	"github.com/ShantKhatri/score-hub-cli/internal/index"
)

var (
	ErrNotFound         = errors.New("provisioner not found")
	ErrVariantNotFound  = errors.New("variant not found")
	ErrPlatformNotFound = errors.New("platform not supported for this variant")
	ErrChecksumMismatch = errors.New("checksum verification failed")
)

type SearchOpts struct {
	Category string
	Platform string
}

type ProvisionerSummary struct {
	Name        string
	DisplayName string
	Category    string
	Platforms   []string
	Variants    int
	Version     string
	Description string
}

type Resolver interface {
	// TODO: Implementation is remaining for these below methods
	Index(ctx context.Context) (*index.Index, error)
	Search(ctx context.Context, query string, opts SearchOpts) ([]ProvisionerSummary, error)
	Resolve(ctx context.Context, name string) (*index.Provisioner, error)
	ResolveVersion(ctx context.Context, name, version string) (*index.Provisioner, error)
	FetchFile(ctx context.Context, p *index.Provisioner, variant, platform string) ([]byte, string, error)
}
