package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/ShantKhatri/score-hub-cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a provisioner",
	Long: `Display detailed information about a provisioner including
variants, platforms, prerequisites, install commands, and version history.

Examples:
  score-hub info dapr-pubsub
  score-hub info route --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryAlias, name, _ := parseNameVersion(args[0])

		if registryAlias == "" {
			registryAlias = resolveRegistryFlag(flagRegistry)
		} else if flagRegistry != "" {
			return fmt.Errorf("cannot use both a scoped name (%s) and the --registry flag", args[0])
		}

		m, err := getManager()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var prov *index.Provisioner
		var targetRegistry string

		if registryAlias != "" {
			res, err := m.ResolverFor(registryAlias)
			if err != nil {
				return err
			}
			p, err := res.Resolve(ctx, name)
			if err != nil {
				return fmt.Errorf("provisioner %q not found in registry %q", name, registryAlias)
			}
			prov = p
			targetRegistry = registryAlias
		} else {
			found, err := m.FindAcrossRegistries(ctx, name)
			if err != nil {
				return err
			}
			if len(found) == 0 {
				return fmt.Errorf("provisioner %q not found in any registry.\n"+
					"Run 'score-hub search %s' to find similar provisioners", name, name)
			}
			if len(found) > 1 {
				if p, inDefault := found[m.DefaultAlias()]; inDefault {
					prov = p
					targetRegistry = m.DefaultAlias()
					fmt.Fprintf(os.Stderr, "Warning: %q found in multiple registries. Showing info from %q.\n", name, targetRegistry)
				} else {
					var aliases []string
					for a := range found {
						aliases = append(aliases, a)
					}
					return fmt.Errorf("conflict: %q found in multiple registries (%s). Use scoped name (e.g. %s/%s) to specify", name, strings.Join(aliases, ", "), aliases[0], name)
				}
			} else {
				for a, p := range found {
					prov = p
					targetRegistry = a
				}
			}
		}

		registryURL, _ := m.URLForAlias(targetRegistry)

		output.PrintInfo(prov, targetRegistry, registryURL, flagJSON)
		return nil
	},
}
