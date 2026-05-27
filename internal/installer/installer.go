// Package installer handles platform detection, downloads, checksums, and file placement.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/score-hub/cli/internal/index"
	"github.com/score-hub/cli/internal/lockfile"
	"github.com/score-hub/cli/internal/resolver"
	gh "github.com/score-hub/cli/internal/resolver/github"
)

// DetectedPlatform represents a detected Score platform
type DetectedPlatform struct {
	Name   string // "k8s" or "compose"
	Source string // how it was detected (flag, env, config, directory, binary)
}

// DetectPlatform detects platform from: flag, env var, config, directories, or binaries in PATH
func DetectPlatform(flagPlatform string, configPlatform string) (*DetectedPlatform, error) {
	if flagPlatform != "" {
		p := normalizePlatform(flagPlatform)
		if p == "" {
			return nil, fmt.Errorf("invalid platform: %s (expected 'k8s' or 'compose')", flagPlatform)
		}
		return &DetectedPlatform{Name: p, Source: "flag"}, nil
	}

	if envPlatform := os.Getenv("SCORE_HUB_PLATFORM"); envPlatform != "" {
		p := normalizePlatform(envPlatform)
		if p == "" {
			return nil, fmt.Errorf("invalid SCORE_HUB_PLATFORM: %s (expected 'k8s' or 'compose')", envPlatform)
		}
		return &DetectedPlatform{Name: p, Source: "env"}, nil
	}

	if configPlatform != "" && configPlatform != "auto" {
		p := normalizePlatform(configPlatform)
		if p != "" {
			return &DetectedPlatform{Name: p, Source: "config"}, nil
		}
	}

	hasK8s := dirExists(".score-k8s")
	hasCompose := dirExists(".score-compose")

	if hasK8s && hasCompose {
		return nil, fmt.Errorf("both .score-k8s/ and .score-compose/ directories found.\n" +
			"Use --platform k8s or --platform compose to specify which one to use")
	}
	if hasK8s {
		return &DetectedPlatform{Name: "k8s", Source: "directory (.score-k8s/)"}, nil
	}
	if hasCompose {
		return &DetectedPlatform{Name: "compose", Source: "directory (.score-compose/)"}, nil
	}

	hasK8sBin := binaryExists("score-k8s")
	hasComposeBin := binaryExists("score-compose")

	if hasK8sBin && hasComposeBin {
		return nil, fmt.Errorf("both score-k8s and score-compose found in PATH.\n" +
			"Use --platform k8s or --platform compose to specify which one to use")
	}
	if hasK8sBin {
		return &DetectedPlatform{Name: "k8s", Source: "binary (score-k8s in PATH)"}, nil
	}
	if hasComposeBin {
		return &DetectedPlatform{Name: "compose", Source: "binary (score-compose in PATH)"}, nil
	}

	return nil, fmt.Errorf("no Score implementation detected.\n" +
		"Install score-k8s or score-compose first, or use --platform k8s|compose")
}

// InstallResult holds the result of a successful installation
type InstallResult struct {
	ProvisionerName string
	Variant         string
	Platform        string
	Version         string
	Filename        string
	InstallPath     string
	Checksum        string
}

// Install resolves provisioner, detects platform, fetches file, verifies checksum, and updates lock file.
func Install(
	ctx context.Context,
	res resolver.Resolver,
	idx *index.Index,
	provName string,
	variantID string,
	flagPlatform string,
	configPlatform string,
	version string,
	skipConfirm bool,
	targetDir string,
) (*InstallResult, error) {
	prov := idx.FindProvisioner(provName)
	if prov == nil {
		return nil, fmt.Errorf("provisioner '%s' not found.\n"+
			"Run 'score-hub search %s' to find similar provisioners", provName, provName)
	}

	var variant *index.Variant
	if variantID != "" {
		variant = prov.FindVariant(variantID)
		if variant == nil {
			var available []string
			for _, v := range prov.Variants {
				available = append(available, v.ID)
			}
			return nil, fmt.Errorf("variant '%s' not found for provisioner '%s'.\n"+
				"Available variants: %s", variantID, provName, strings.Join(available, ", "))
		}
	} else if len(prov.Variants) == 1 {
		variant = &prov.Variants[0]
		fmt.Printf("Auto-selected variant: %s\n", variant.ID)
	} else {
		fmt.Printf("\nProvisioner '%s' has multiple variants:\n\n", prov.Name)
		for _, v := range prov.Variants {
			desc := v.Description
			if desc == "" {
				desc = v.DisplayName
			}
			var platforms []string
			for p := range v.Platforms {
				platforms = append(platforms, p)
			}
			fmt.Printf("  %-15s %s (%s)\n", v.ID, desc, strings.Join(platforms, ", "))
		}
		fmt.Printf("\nSpecify a variant with --variant <name>\n")
		fmt.Printf("Example: score-hub install %s --variant %s\n", prov.Name, prov.Variants[0].ID)
		return nil, fmt.Errorf("variant selection required")
	}

	platform, err := DetectPlatform(flagPlatform, configPlatform)
	if err != nil {
		return nil, err
	}

	platformData, ok := variant.Platforms[platform.Name]
	if !ok {
		var available []string
		for p := range variant.Platforms {
			available = append(available, p)
		}
		return nil, fmt.Errorf("variant '%s' does not support platform '%s' (available: %s)",
			variant.ID, platform.Name, strings.Join(available, ", "))
	}

	if len(variant.Prerequisites) > 0 && !skipConfirm {
		fmt.Printf("Prerequisites for %s (%s, %s):\n", prov.Name, variant.ID, platform.Name)
		for _, prereq := range variant.Prerequisites {
			fmt.Printf("  • %s\n", prereq)
		}
		fmt.Println()
	}

	ver := version
	if ver == "" {
		ver = prov.LatestVersion()
	}

	fmt.Printf("Downloading %s...\n", platformData.Filename)
	data, err := res.FetchFile(ctx, platformData.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if platformData.Checksum != "" {
		if err := gh.VerifyChecksum(data, platformData.Checksum); err != nil {
			return nil, fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("Checksum verified.")
	}

	var installDir string
	if targetDir != "" {
		installDir = targetDir
	} else if platform.Name == "k8s" {
		installDir = ".score-k8s"
	} else {
		installDir = ".score-compose"
	}

	installPath := filepath.Join(installDir, platformData.Filename)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(installPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	fmt.Printf("Installed to %s\n", installPath)
	checksum := gh.ComputeChecksum(data)
	lock, _ := lockfile.Load(".")
	if lock == nil {
		lock = lockfile.New()
	}
	lock.AddEntry(lockfile.LockEntry{
		Name:        prov.Name,
		Variant:     variant.ID,
		Platform:    platform.Name,
		Version:     ver,
		Commit:      prov.LatestCommit(),
		Filename:    platformData.Filename,
		Checksum:    checksum,
		InstalledAt: index.NowISO(),
	})
	if err := lock.Save("."); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update lock file: %v\n", err)
	}

	return &InstallResult{
		ProvisionerName: prov.Name,
		Variant:         variant.ID,
		Platform:        platform.Name,
		Version:         ver,
		Filename:        platformData.Filename,
		InstallPath:     installPath,
		Checksum:        checksum,
	}, nil
}

// normalizePlatform converts platform aliases to canonical names
func normalizePlatform(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "k8s", "kubernetes", "score-k8s":
		return "k8s"
	case "compose", "docker-compose", "score-compose":
		return "compose"
	default:
		return ""
	}
}

// dirExists reports if path is a directory
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// binaryExists reports if name is in PATH
func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
