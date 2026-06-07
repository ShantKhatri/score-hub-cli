package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/lockfile"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRegistryRemoveCmd() *cobra.Command {
	var (
		forceFlag bool
		yesFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a provisioner registry",
		Long: `Remove a named registry from the configuration.

Installed provisioners from the removed registry remain on disk and in the lockfile.

Examples:
  score-hub registry remove myorg
  score-hub registry remove myorg --force
  score-hub registry remove myorg --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			cfg := config.Load()

			if _, exists := cfg.Registries[alias]; !exists {
				fmt.Printf("Registry %q not found.\n", alias)
				return nil
			}

			lock, _ := lockfile.Load(".")
			if lock != nil && !yesFlag {
				var fromRegistry int
				for _, e := range lock.Entries {
					effReg := e.EffectiveRegistry()
					if effReg == alias {
						fromRegistry++
					}
				}
				if fromRegistry > 0 {
					yellow := color.New(color.FgYellow)
					yellow.Printf("Warning: %d provisioner(s) in .score-hub.lock were installed from %q.\n", fromRegistry, alias)
					fmt.Println("They remain installed on disk. Remove from config? [y/N]")

					reader := bufio.NewReader(os.Stdin)
					response, _ := reader.ReadString('\n')
					response = strings.TrimSpace(strings.ToLower(response))
					if response != "y" && response != "yes" {
						fmt.Println("Aborted.")
						return nil
					}
				}
			}

			if err := cfg.RemoveRegistry(alias, forceFlag); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			green := color.New(color.FgGreen)
			green.Printf("✓ Registry %q removed\n", alias)
			return nil
		},
	}

	cmd.Flags().BoolVar(&forceFlag, "force", false, "Remove even if it is the default or public registry")
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Skip confirmation prompts")

	return cmd
}
