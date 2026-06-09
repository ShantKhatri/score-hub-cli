package output

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/lockfile"
	"github.com/ShantKhatri/score-hub-cli/internal/registry"
	"github.com/fatih/color"
)

func PrintSearchResultsTagged(results []registry.TaggedResult, query string, jsonOut bool) {
	if jsonOut {
		printSummaryJSON(results)
		return
	}
	if len(results) == 0 {
		fmt.Println("No provisioners found.")
		return
	}

	if query == "" {
		printSummaryGroupedByCategory(results)
		return
	}

	fmt.Printf("%-22s %-14s %-12s %-10s %-10s %s\n",
		"NAME", "CATEGORY", "REGISTRY", "PLATFORMS", "VARIANTS", "VERSION")
	fmt.Println(strings.Repeat("─", 85))
	for _, p := range results {
		platforms := strings.Join(p.Platforms, ", ")
		fmt.Printf("%-22s %-14s %-12s %-10s %-10d %s\n",
			p.Name, p.Category, p.RegistryAlias, platforms, p.Variants, p.Version)
	}
	fmt.Printf("\n%d result(s). Run 'score-hub info <name>' for details.\n", len(results))
}

func printSummaryGroupedByCategory(results []registry.TaggedResult) {
	categories := make(map[string][]registry.TaggedResult)
	var categoryOrder []string
	for _, p := range results {
		cat := p.Category
		if cat == "" {
			cat = "other"
		}
		if _, exists := categories[cat]; !exists {
			categoryOrder = append(categoryOrder, cat)
		}
		categories[cat] = append(categories[cat], p)
	}
	sort.Strings(categoryOrder)

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	for i, cat := range categoryOrder {
		if i > 0 {
			fmt.Println()
		}
		bold.Printf("[%s]\n", cat)
		for _, p := range categories[cat] {
			platforms := strings.Join(p.Platforms, ", ")
			fmt.Printf("  %-22s %s", p.Name, p.DisplayName)
			dim.Printf("  (%s | %s)\n", p.RegistryAlias, platforms)
		}
	}
	fmt.Printf("\n%d provisioner(s) available. Run 'score-hub info <name>' for details.\n", len(results))
}

func printSummaryJSON(results []registry.TaggedResult) {
	type searchResult struct {
		Name         string   `json:"name"`
		DisplayName  string   `json:"displayName"`
		Category     string   `json:"category"`
		Registry     string   `json:"registry"`
		Platforms    []string `json:"platforms"`
		VariantCount int      `json:"variantCount"`
		Version      string   `json:"latestVersion"`
	}
	var out []searchResult
	for _, p := range results {
		out = append(out, searchResult{
			Name:         p.Name,
			DisplayName:  p.DisplayName,
			Category:     p.Category,
			Registry:     p.RegistryAlias,
			Platforms:    p.Platforms,
			VariantCount: p.Variants,
			Version:      p.Version,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

func PrintSearchResults(results []index.Provisioner, query string, jsonOut bool) {
	if jsonOut {
		printSearchJSON(results)
		return
	}
	if len(results) == 0 {
		fmt.Println("No provisioners found.")
		return
	}

	// If no query was provided, group by category
	if query == "" {
		printGroupedByCategory(results)
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

func printGroupedByCategory(results []index.Provisioner) {
	categories := make(map[string][]index.Provisioner)
	var categoryOrder []string
	for _, p := range results {
		cat := p.Category
		if cat == "" {
			cat = "other"
		}
		if _, exists := categories[cat]; !exists {
			categoryOrder = append(categoryOrder, cat)
		}
		categories[cat] = append(categories[cat], p)
	}
	sort.Strings(categoryOrder)

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	for i, cat := range categoryOrder {
		if i > 0 {
			fmt.Println()
		}
		bold.Printf("[%s]\n", cat)
		for _, p := range categories[cat] {
			platforms := strings.Join(p.PlatformNames(), ", ")
			fmt.Printf("  %-22s %s", p.Name, p.DisplayName)
			dim.Printf("  (%s)\n", platforms)
		}
	}
	fmt.Printf("\n%d provisioner(s) available. Run 'score-hub info <name>' for details.\n", len(results))
}

func printSearchJSON(results []index.Provisioner) {
	type searchResult struct {
		Name         string   `json:"name"`
		DisplayName  string   `json:"displayName"`
		Category     string   `json:"category"`
		Platforms    []string `json:"platforms"`
		VariantCount int      `json:"variantCount"`
		Version      string   `json:"latestVersion"`
	}
	var out []searchResult
	for _, p := range results {
		out = append(out, searchResult{
			Name:         p.Name,
			DisplayName:  p.DisplayName,
			Category:     p.Category,
			Platforms:    p.PlatformNames(),
			VariantCount: p.VariantCount(),
			Version:      p.LatestVersion(),
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

func PrintInfo(p *index.Provisioner, registryAlias, registryURL string, jsonOut bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(p)
		return
	}

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)

	bold.Printf("\n%s", p.Name)
	dim.Printf(" — ")
	cyan.Println(p.DisplayName)
	fmt.Printf("Registry:  %s (%s)\n", registryAlias, registryURL)
	fmt.Printf("Category:  %s\n", p.Category)
	fmt.Printf("Platforms: %s\n", strings.Join(p.PlatformNames(), ", "))
	if p.Upstream != "" {
		fmt.Printf("Upstream:  %s\n", p.Upstream)
	}
	fmt.Printf("Version:   %s (%s)\n", p.LatestVersion(), p.LatestDate())
	fmt.Println()

	bold.Println("DESCRIPTION")
	lines := strings.Split(strings.TrimSpace(p.Description), "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", strings.TrimSpace(line))
	}
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
		sort.Strings(plats)
		platStr := strings.Join(plats, ", ")

		if len(plats) == 1 {
			fmt.Printf("  %-15s %-40s %s only\n", v.ID, desc, plats[0])
		} else {
			fmt.Printf("  %-15s %-40s %s\n", v.ID, desc, platStr)
		}
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
	bold.Println("USAGE IN score.yaml")
	cyan.Printf("  resources:\n    my-resource:\n      type: %s\n", p.Name)
	fmt.Println()
	bold.Println("INSTALL")
	for _, v := range p.Variants {
		var plats []string
		for pl := range v.Platforms {
			plats = append(plats, pl)
		}
		sort.Strings(plats)
		platInfo := ""
		if len(plats) == 1 {
			platInfo = fmt.Sprintf("  (%s only)", plats[0])
		} else {
			platInfo = fmt.Sprintf("  (%s)", strings.Join(plats, " + "))
		}
		fmt.Printf("  score-hub install %s --variant %s%s\n", p.Name, v.ID, platInfo)
	}

	if len(p.Versions) > 0 {
		fmt.Println()
		bold.Println("VERSION HISTORY")
		for _, ver := range p.Versions {
			fmt.Printf("  %-8s  %-12s  %s\n", ver.Version, ver.Date, ver.Changelog)
		}
	}

	fmt.Println()
	dim.Printf("Upstream: https://github.com/%s/tree/main/%s\n", p.Upstream, p.Name)
	fmt.Println()
}

func PrintInstallSuccess(name, version, variant, platform, installPath, resType string) {
	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	green.Printf("✓ Installed %s@%s (%s, %s)\n", name, version, variant, platform)
	dim.Printf("  → %s\n", installPath)
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

func PrintList(entries []lockfile.LockEntry, mgr *registry.Manager, jsonOut bool) {
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
	fmt.Printf("%-22s %-12s %-12s %-10s %-10s %s\n",
		"NAME", "REGISTRY", "VARIANT", "PLATFORM", "VERSION", "STATUS")
	fmt.Println(strings.Repeat("─", 85))
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	for _, e := range entries {
		status := green.Sprint("✓ installed")
		if mgr != nil {
			// Find update in the specific registry it was installed from
			regAlias := e.EffectiveRegistry()
			if res, err := mgr.ResolverFor(regAlias); err == nil {
				if p, err := res.Resolve(context.Background(), e.Name); err == nil && p.LatestVersion() != e.Version {
					status = yellow.Sprintf("↑ update available (%s)", p.LatestVersion())
				}
			}
		}
		fmt.Printf("%-22s %-12s %-12s %-10s %-10s %s\n",
			e.Name, e.EffectiveRegistry(), e.Variant, e.Platform, e.Version, status)
	}
	fmt.Printf("\n%d provisioner(s) installed.\n", len(entries))
}
