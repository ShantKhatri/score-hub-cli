package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/registry"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRegistryShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show details and provisioners for a registry",
		Long: `Fetch the index.yaml for a named registry and display its provisioners.

Examples:
  score-hub registry show public
  score-hub registry show myorg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			cfg := config.Load()

			entry, exists := cfg.Registries[alias]
			if !exists {
				return fmt.Errorf("registry %q not configured. Run 'score-hub registry add' to add it", alias)
			}
			mgr, err := registry.NewManager(cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize registries: %w", err)
			}

			res, err := mgr.ResolverFor(alias)
			if err != nil {
				return err
			}

			ctx := context.Background()
			idx, err := res.Index(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch index from registry %q (%s): %w", alias, entry.URL, err)
			}

			bold := color.New(color.Bold)
			dim := color.New(color.Faint)

			bold.Printf("Registry: %s\n", alias)
			dim.Printf("URL: %s\n", entry.URL)
			fmt.Println()

			if len(idx.Provisioners) == 0 {
				fmt.Println("No provisioners found in this registry.")
				return nil
			}

			bold.Printf("%-22s %-10s %-20s %s\n", "NAME", "VERSION", "VARIANTS", "PLATFORMS")
			fmt.Println(strings.Repeat("─", 75))

			for _, p := range idx.Provisioners {
				var variantNames []string
				for _, v := range p.Variants {
					variantNames = append(variantNames, v.ID)
				}

				platformSet := make(map[string]bool)
				for _, v := range p.Variants {
					for plat := range v.Platforms {
						platformSet[plat] = true
					}
				}
				var platforms []string
				for plat := range platformSet {
					platforms = append(platforms, plat)
				}
				sort.Strings(platforms)

				fmt.Printf("%-22s %-10s %-20s %s\n",
					p.Name,
					p.LatestVersion(),
					strings.Join(variantNames, ", "),
					strings.Join(platforms, ", "))
			}

			fmt.Printf("\n%d provisioners.\n", len(idx.Provisioners))
			return nil
		},
	}

	return cmd
}
