package index

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse reads and parses an index.yaml file from bytes
func Parse(data []byte) (*Index, error) {
	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	if idx.APIVersion == "" {
		return nil, fmt.Errorf("index missing apiVersion field")
	}

	if idx.Kind != "Index" {
		return nil, fmt.Errorf("expected kind 'Index', got '%s'", idx.Kind)
	}

	return &idx, nil
}

// ParseFile reads and parses an index.yaml file from disk
func ParseFile(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}
	return Parse(data)
}

// Search filters provisioners by a query string, matching against
// name, displayName, description, category, and tags
func (idx *Index) Search(query string, category string, platform string) []Provisioner {
	query = strings.ToLower(strings.TrimSpace(query))
	category = strings.ToLower(strings.TrimSpace(category))
	platform = strings.ToLower(strings.TrimSpace(platform))

	var results []Provisioner

	for _, p := range idx.Provisioners {
		// If query is specified, check name/description/tags match
		if query != "" {
			match := false
			if strings.Contains(strings.ToLower(p.Name), query) {
				match = true
			}
			if strings.Contains(strings.ToLower(p.DisplayName), query) {
				match = true
			}
			if strings.Contains(strings.ToLower(p.Description), query) {
				match = true
			}
			if strings.Contains(strings.ToLower(p.Category), query) {
				match = true
			}
			for _, tag := range p.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by category
		if category != "" && strings.ToLower(p.Category) != category {
			continue
		}

		// Filter by platform
		if platform != "" {
			hasPlatform := false
			for _, v := range p.Variants {
				if v.HasPlatform(platform) {
					hasPlatform = true
					break
				}
			}
			if !hasPlatform {
				continue
			}
		}

		results = append(results, p)
	}

	return results
}

// FindProvisioner looks up a provisioner by exact name
func (idx *Index) FindProvisioner(name string) *Provisioner {
	name = strings.ToLower(strings.TrimSpace(name))
	for i := range idx.Provisioners {
		if strings.ToLower(idx.Provisioners[i].Name) == name {
			return &idx.Provisioners[i]
		}
	}
	return nil
}
