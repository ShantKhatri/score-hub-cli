package cmd

import (
	"context"

	"github.com/ShantKhatri/score-hub-cli/internal/output"
	"github.com/ShantKhatri/score-hub-cli/internal/registry"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	searchCategory string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for provisioners by name, category, or keyword",
	Long: `Search the score-hub index for provisioners matching a query.

Examples:
  score-hub search dapr
  score-hub search messaging
  score-hub search --category networking
  score-hub search --platform k8s
  score-hub search              # list all provisioners grouped by category`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		m, err := getManager()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var results []registry.TaggedResult

		alias := resolveRegistryFlag(flagRegistry)
		if alias != "" {
			res, err := m.ResolverFor(alias)
			if err != nil {
				return err
			}
			summaries, err := res.Search(ctx, query, resolver.SearchOpts{
				Category: searchCategory,
				Platform: flagPlatform,
			})
			if err != nil {
				return err
			}
			url, _ := m.URLForAlias(alias)
			for _, s := range summaries {
				results = append(results, registry.TaggedResult{
					ProvisionerSummary: s,
					RegistryAlias:      alias,
					RegistryURL:        url,
				})
			}
		} else {
			results, err = m.SearchAll(ctx, query, resolver.SearchOpts{
				Category: searchCategory,
				Platform: flagPlatform,
			})
			if err != nil {
				return err
			}
		}

		output.PrintSearchResultsTagged(results, query, flagJSON)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchCategory, "category", "",
		"Filter by category (e.g., messaging, networking, ai, compute)")
}
