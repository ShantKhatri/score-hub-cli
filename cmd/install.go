package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/score-hub/cli/internal/config"
	"github.com/score-hub/cli/internal/installer"
	"github.com/score-hub/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	installVariant string
	installYes     bool
	installDir     string
)

var installCmd = &cobra.Command{
	Use:   "install <name>[@version]",
	Short: "Install a provisioner into your Score project",
	Long: `Install a provisioner from the score-hub index into your
current Score project directory.

Examples:
  score-hub install dapr-pubsub --variant redis
  score-hub install dapr-pubsub --variant redis --platform k8s
  score-hub install dapr-pubsub@1.0.0 --variant redis
  score-hub install route --variant ingress --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, version := parseNameVersion(args[0])

		res, err := getResolver()
		if err != nil {
			return err
		}

		ctx := context.Background()
		idx, err := res.GetIndex(ctx, flagNoCache)
		if err != nil {
			return fmt.Errorf("failed to load index: %w", err)
		}

		cfg := config.Load()

		ver := version
		if ver == "" {
			ver = "latest"
		}
		fmt.Printf("Resolving %s@%s...\n", name, ver)

		result, err := installer.Install(
			ctx, res, idx,
			name, installVariant,
			flagPlatform, cfg.Platform,
			version, installYes, installDir,
		)
		if err != nil {
			return err
		}

		output.PrintInstallSuccess(
			result.ProvisionerName,
			result.Version,
			result.Variant,
			result.Platform,
			result.InstallPath,
			name,
		)
		return nil
	},
}

func parseNameVersion(arg string) (name, version string) {
	if idx := strings.LastIndex(arg, "@"); idx > 0 {
		return arg[:idx], arg[idx+1:]
	}
	return arg, ""
}

func init() {
	installCmd.Flags().StringVar(&installVariant, "variant", "",
		"Provisioner variant (e.g., redis, rabbitmq)")
	installCmd.Flags().BoolVar(&installYes, "yes", false,
		"Skip confirmation prompts")
	installCmd.Flags().StringVar(&installDir, "dir", "",
		"Override install directory (default: auto-detect)")
}
