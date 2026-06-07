package cmd

import (
	"github.com/spf13/cobra"
)

func newRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage provisioner registries",
		Long: `Add, remove, and list provisioner registries.

A registry is a named URL pointing to an index.yaml file that describes
available provisioners. The public score-hub registry is always configured.

Examples:
  score-hub registry add myorg https://raw.githubusercontent.com/myorg/score-provisioners/main/index.yaml
  score-hub registry list
  score-hub registry show myorg
  score-hub registry remove myorg`,
	}
	cmd.AddCommand(
		newRegistryAddCmd(),
		newRegistryListCmd(),
		newRegistryRemoveCmd(),
		newRegistryShowCmd(),
	)
	return cmd
}
