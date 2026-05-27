// Package main provides a utility to sync checksums from upstream community-provisioners.
//
// Usage:
//
//	go run scripts/sync_checksums.go [--update]
//
// If --update is provided, the cmd/index.yaml file will be updated in place.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/score-hub/cli/internal/index"
	gh "github.com/score-hub/cli/internal/resolver/github"
)

func main() {
	updateFlag := flag.Bool("update", false, "Update cmd/index.yaml in place")
	flag.Parse()

	data, err := os.ReadFile("cmd/index.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading index: %v\n", err)
		os.Exit(1)
	}

	idx, err := index.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing index: %v\n", err)
		os.Exit(1)
	}

	res := gh.NewGitHubResolver("")
	ctx := context.Background()

	computedChecksums := make(map[string]string)
	existingChecksums := make(map[string]string)
	changes := 0

	fmt.Println("Fetching upstream provisioners to verify checksums...")
	for _, p := range idx.Provisioners {
		for _, v := range p.Variants {
			for _, plat := range v.Platforms {
				existingChecksums[plat.Path] = plat.Checksum

				body, err := res.FetchFile(ctx, plat.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "FAIL: %s: %v\n", plat.Path, err)
					continue
				}

				computedChecksums[plat.Path] = gh.ComputeChecksum(body)
			}
		}
	}

	fmt.Println("\nDiff Report:")
	for path, computed := range computedChecksums {
		existing := existingChecksums[path]
		if existing != computed {
			fmt.Printf("  %s\n    old: %s\n    new: %s\n", path, existing, computed)
			changes++
		}
	}

	if changes == 0 {
		fmt.Println("  All checksums are up to date.")
		os.Exit(0)
	}

	if !*updateFlag {
		fmt.Printf("\nFound %d mismatches. Run with --update to fix them.\n", changes)
		os.Exit(1)
	}

	// Update the file in-place line by line to preserve YAML comments and structure
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "path: ") {
			path := strings.TrimPrefix(trimmed, "path: ")
			if cs, ok := computedChecksums[path]; ok && existingChecksums[path] != cs {
				// Search ahead up to 3 lines for the corresponding checksum field
				for j := i + 1; j < len(lines) && j <= i+3; j++ {
					if strings.Contains(lines[j], "checksum:") {
						indent := lines[j][:len(lines[j])-len(strings.TrimLeft(lines[j], " "))]
						lines[j] = indent + "checksum: " + fmt.Sprintf("%q", cs)
						break
					}
				}
			}
		}
	}

	if err := os.WriteFile("cmd/index.yaml", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing updated index: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n✓ Successfully updated %d checksums in cmd/index.yaml\n", changes)
}
