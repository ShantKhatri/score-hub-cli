package cmd

import (
	"context"
	"fmt"

	"github.com/score-hub/cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a provisioner",
	Long: `Display detailed information about a provisioner including
variants, platforms, prerequisites, install commands, and version history.

Examples:
  score-hub info dapr-pubsub
  score-hub info route`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		res, err := getResolver()
		if err != nil {
			return err
		}

		ctx := context.Background()
		idx, err := res.GetIndex(ctx, flagNoCache)
		if err != nil {
			return fmt.Errorf("failed to load index: %w", err)
		}

		prov := idx.FindProvisioner(name)
		if prov == nil {
			return fmt.Errorf("provisioner '%s' not found.\n"+
				"Run 'score-hub search %s' to find similar provisioners", name, name)
		}

		output.PrintInfo(prov)
		return nil
	},
}
