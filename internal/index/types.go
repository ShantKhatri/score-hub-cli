package index

import "time"

// Index is the top-level index.yaml structure
type Index struct {
	APIVersion   string        `yaml:"apiVersion"`
	Kind         string        `yaml:"kind"`
	Metadata     IndexMetadata `yaml:"metadata"`
	Provisioners []Provisioner `yaml:"provisioners"`
}

// IndexMetadata contains index-level metadata
type IndexMetadata struct {
	Generated     string `yaml:"generated"`
	SchemaVersion string `yaml:"schemaVersion"`
}

// Provisioner is a named entry in the index
type Provisioner struct {
	Name        string    `yaml:"name"`
	DisplayName string    `yaml:"displayName"`
	Description string    `yaml:"description"`
	Category    string    `yaml:"category"`
	Tags        []string  `yaml:"tags"`
	Upstream    string    `yaml:"upstream"`
	Variants    []Variant `yaml:"variants"`
	Versions    []Version `yaml:"versions"`
}

// Variant is an implementation variant of a provisioner
type Variant struct {
	ID            string              `yaml:"id"`
	DisplayName   string              `yaml:"displayName,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	Platforms     map[string]Platform `yaml:"platforms"`
	Prerequisites []string            `yaml:"prerequisites,omitempty"`
}

type Platform struct {
	Path     string `yaml:"path,omitempty"` // Relative path in community-provisioners
	Filename string `yaml:"filename"`       // Original filename to preserve on install
	Checksum string `yaml:"checksum"`       // sha256:hex checksum of file content

	// Present in org-registry provisioners. Absent in community provisioners.
	// FetchFile checks this field first; falls back to Path if empty.
	DownloadURL string `yaml:"downloadURL,omitempty"`
}

// Version represents a version entry for a provisioner
type Version struct {
	Version        string `yaml:"version"`
	Date           string `yaml:"date"`
	UpstreamCommit string `yaml:"upstreamCommit"`
	Changelog      string `yaml:"changelog"`
}

// LatestVersion returns the latest version or "0.0.0"
func (p *Provisioner) LatestVersion() string {
	if len(p.Versions) == 0 {
		return "0.0.0"
	}
	return p.Versions[0].Version
}

// LatestDate returns the latest version's date
func (p *Provisioner) LatestDate() string {
	if len(p.Versions) == 0 {
		return "unknown"
	}
	return p.Versions[0].Date
}

// LatestCommit returns the upstream commit hash
func (p *Provisioner) LatestCommit() string {
	if len(p.Versions) == 0 {
		return ""
	}
	return p.Versions[0].UpstreamCommit
}

// PlatformNames returns unique platform names across all variants
func (p *Provisioner) PlatformNames() []string {
	seen := make(map[string]bool)
	var platforms []string
	for _, v := range p.Variants {
		for platform := range v.Platforms {
			if !seen[platform] {
				seen[platform] = true
				platforms = append(platforms, platform)
			}
		}
	}
	return platforms
}

// VariantCount returns the number of variants
func (p *Provisioner) VariantCount() int {
	return len(p.Variants)
}

// FindVariant returns the variant with given ID
func (p *Provisioner) FindVariant(id string) *Variant {
	for i := range p.Variants {
		if p.Variants[i].ID == id {
			return &p.Variants[i]
		}
	}
	return nil
}

// HasPlatform reports if variant supports the platform
func (v *Variant) HasPlatform(platform string) bool {
	_, ok := v.Platforms[platform]
	return ok
}

// InstalledAt helper for lock file
func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
