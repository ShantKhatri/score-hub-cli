package cmd

import (
	"fmt"

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

		m, _ := getManager()

		output.PrintList(lock.Entries, m, flagJSON)
		return nil
	},
}
