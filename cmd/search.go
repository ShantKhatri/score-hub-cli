package cmd

import (
	"context"
	"fmt"

	"github.com/score-hub/cli/internal/output"
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
  score-hub search              # list all provisioners`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		res, err := getResolver()
		if err != nil {
			return err
		}

		ctx := context.Background()
		idx, err := res.GetIndex(ctx, flagNoCache)
		if err != nil {
			return fmt.Errorf("failed to load index: %w", err)
		}

		results := idx.Search(query, searchCategory, flagPlatform)
		output.PrintSearchResults(results, flagJSON)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchCategory, "category", "",
		"Filter by category (e.g., messaging, networking)")
}
