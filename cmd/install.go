package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/installer"
	"github.com/ShantKhatri/score-hub-cli/internal/output"
	"github.com/ShantKhatri/score-hub-cli/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	installVariant  string
	installYes      bool
	installDir      string
	installNoVerify bool
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
  score-hub install route --variant ingress --yes
  score-hub install environment --no-verify`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryAlias, name, version := parseNameVersion(args[0])

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
		cfg := config.Load()

		ver := version
		if ver == "" {
			ver = "latest"
		}

		var res resolver.Resolver
		var targetRegistry string

		if registryAlias != "" {
			r, err := m.ResolverFor(registryAlias)
			if err != nil {
				return err
			}
			res = r
			targetRegistry = registryAlias
		} else {
			found, err := m.FindAcrossRegistries(ctx, name)
			if err != nil {
				return err
			}

			if len(found) == 0 {
				return fmt.Errorf("provisioner %q not found in any registry", name)
			}
			if len(found) > 1 {
				// We have a conflict. Is it in the default registry?
				if _, inDefault := found[m.DefaultAlias()]; inDefault {
					targetRegistry = m.DefaultAlias()
					res, _ = m.ResolverFor(targetRegistry)
					fmt.Fprintf(os.Stderr, "Warning: %q found in multiple registries. Defaulting to %q.\n", name, targetRegistry)
				} else {
					var aliases []string
					for a := range found {
						aliases = append(aliases, a)
					}
					return fmt.Errorf("conflict: %q found in multiple registries (%s). Use scoped name (e.g. %s/%s) to specify.", name, strings.Join(aliases, ", "), aliases[0], name)
				}
			} else {
				// Only found in one registry
				for a := range found {
					targetRegistry = a
				}
				res, _ = m.ResolverFor(targetRegistry)
			}
		}

		fmt.Printf("Resolving %s@%s from %s...\n", name, ver, targetRegistry)

		registryURL, _ := m.URLForAlias(targetRegistry)

		result, err := installer.Install(
			ctx, res,
			name, installVariant,
			flagPlatform, cfg.Platform,
			version, installYes, installNoVerify, installDir,
			targetRegistry, registryURL,
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

func parseNameVersion(arg string) (registryAlias, name, version string) {
	nameAndVersion := arg
	if idx := strings.Index(arg, "/"); idx > 0 {
		registryAlias = arg[:idx]
		nameAndVersion = arg[idx+1:]
	}

	if idx := strings.LastIndex(nameAndVersion, "@"); idx > 0 {
		return registryAlias, nameAndVersion[:idx], nameAndVersion[idx+1:]
	}
	return registryAlias, nameAndVersion, ""
}

func init() {
	installCmd.Flags().StringVar(&installVariant, "variant", "",
		"Provisioner variant (e.g., redis, rabbitmq)")
	installCmd.Flags().BoolVar(&installYes, "yes", false,
		"Skip confirmation prompts")
	installCmd.Flags().StringVar(&installDir, "dir", "",
		"Override install directory (default: auto-detect)")
	installCmd.Flags().BoolVar(&installNoVerify, "no-verify", false,
		"Skip checksum verification (not recommended)")
}
