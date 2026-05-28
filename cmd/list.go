package cmd

import (
	"context"
	"fmt"

	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/lockfile"
	"github.com/ShantKhatri/score-hub-cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed provisioners in the current project",
	Long: `Show all provisioners installed in the current project directory
by reading the .score-hub.lock file.

Examples:
  score-hub list
  score-hub list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lock, err := lockfile.Load(".")
		if err != nil {
			fmt.Println("No .score-hub.lock found. Run 'score-hub install <name>' to start.")
			return nil
		}

		// Try to load index for update status comparison
		var idx *index.Index
		res, resErr := getResolver()
		if resErr == nil {
			ctx := context.Background()
			if loadedIdx, loadErr := res.GetIndex(ctx, false); loadErr == nil {
				idx = loadedIdx
			}
		}

		output.PrintList(lock.Entries, idx, flagJSON)
		return nil
	},
}
