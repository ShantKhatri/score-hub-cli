package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/score-hub/cli/internal/index"
	"github.com/score-hub/cli/internal/lockfile"
)

func PrintSearchResults(results []index.Provisioner, jsonOut bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(results)
		return
	}
	if len(results) == 0 {
		fmt.Println("No provisioners found.")
		return
	}
	fmt.Printf("%-22s %-14s %-16s %-10s %s\n",
		"NAME", "CATEGORY", "PLATFORMS", "VARIANTS", "VERSION")
	fmt.Println(strings.Repeat("─", 75))
	for _, p := range results {
		platforms := strings.Join(p.PlatformNames(), ", ")
		fmt.Printf("%-22s %-14s %-16s %-10d %s\n",
			p.Name, p.Category, platforms, p.VariantCount(), p.LatestVersion())
	}
	fmt.Printf("\n%d result(s). Run 'score-hub info <name>' for details.\n", len(results))
}

func PrintInfo(p *index.Provisioner) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)

	bold.Printf("\n%s", p.Name)
	dim.Printf(" v%s\n", p.LatestVersion())
	cyan.Println(p.DisplayName)
	fmt.Println()
	fmt.Printf("Category:  %s\n", p.Category)
	fmt.Printf("Platforms: %s\n", strings.Join(p.PlatformNames(), ", "))
	fmt.Printf("Upstream:  %s\n", p.Upstream)
	fmt.Printf("Updated:   %s\n", p.LatestDate())
	fmt.Println()
	bold.Println("DESCRIPTION")
	fmt.Printf("  %s\n", p.Description)
	fmt.Println()
	bold.Println("VARIANTS")
	for _, v := range p.Variants {
		desc := v.Description
		if desc == "" {
			desc = v.DisplayName
		}
		var plats []string
		for pl := range v.Platforms {
			plats = append(plats, pl)
		}
		fmt.Printf("  %-15s %s (%s)\n", v.ID, desc, strings.Join(plats, " + "))
	}
	hasPrereqs := false
	for _, v := range p.Variants {
		if len(v.Prerequisites) > 0 {
			hasPrereqs = true
			break
		}
	}
	if hasPrereqs {
		fmt.Println()
		bold.Println("PREREQUISITES")
		seen := make(map[string]bool)
		for _, v := range p.Variants {
			for _, pr := range v.Prerequisites {
				if !seen[pr] {
					seen[pr] = true
					fmt.Printf("  • %s\n", pr)
				}
			}
		}
	}
	fmt.Println()
	bold.Println("INSTALL")
	for _, v := range p.Variants {
		fmt.Printf("  score-hub install %s --variant %s\n", p.Name, v.ID)
	}
	if len(p.Versions) > 0 {
		fmt.Println()
		bold.Println("VERSION HISTORY")
		for _, ver := range p.Versions {
			fmt.Printf("  %-8s  %-12s  %s\n", ver.Version, ver.Date, ver.Changelog)
		}
	}
	fmt.Println()
}

func PrintInstallSuccess(name, version, variant, platform, installPath, resType string) {
	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	green.Printf("✓ Installed %s@%s (%s, %s)\n", name, version, variant, platform)
	dim.Printf("→ %s\n", installPath)
	if resType != "" {
		fmt.Println()
		fmt.Println("Add to your score.yaml:")
		cyan.Printf("  resources:\n    my-resource:\n      type: %s\n", resType)
	}
	fmt.Println()
	tool := "score-k8s"
	if platform == "compose" {
		tool = "score-compose"
	}
	fmt.Printf("Then run: %s generate score.yaml\n", tool)
}

func PrintList(entries []lockfile.LockEntry, idx *index.Index, jsonOut bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(entries)
		return
	}
	if len(entries) == 0 {
		yellow := color.New(color.FgYellow)
		yellow.Println("No provisioners installed in this project.")
		fmt.Println("Run 'score-hub install <name>' to install provisioners.")
		return
	}
	cwd, _ := os.Getwd()
	fmt.Printf("Installed provisioners in %s\n\n", cwd)
	fmt.Printf("%-22s %-12s %-10s %-10s %s\n",
		"NAME", "VARIANT", "PLATFORM", "VERSION", "STATUS")
	fmt.Println(strings.Repeat("─", 70))
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	for _, e := range entries {
		status := green.Sprint("✓ installed")
		if idx != nil {
			p := idx.FindProvisioner(e.Name)
			if p != nil && p.LatestVersion() != e.Version {
				status = yellow.Sprintf("↑ update available (%s)", p.LatestVersion())
			}
		}
		fmt.Printf("%-22s %-12s %-10s %-10s %s\n",
			e.Name, e.Variant, e.Platform, e.Version, status)
	}
	fmt.Printf("\n%d provisioner(s) installed.\n", len(entries))
}
