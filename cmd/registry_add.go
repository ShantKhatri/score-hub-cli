package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/ShantKhatri/score-hub-cli/internal/index"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newRegistryAddCmd() *cobra.Command {
	var (
		setDefault bool
		noValidate bool
		overwrite  bool
	)

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a provisioner registry",
		Long: `Add a named registry pointing to an index.yaml file.

The URL must point to a valid score-hub index.yaml file (unless --no-validate is used).

Examples:
  score-hub registry add myorg https://raw.githubusercontent.com/myorg/score-provisioners/main/index.yaml
  score-hub registry add myorg https://example.com/index.yaml --set-default
  score-hub registry add internal https://internal.example.com/index.yaml --no-validate`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			url := args[1]

			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("registry URL must start with http:// or https://\n  Got: %s", url)
			}

			cfg := config.Load()

			if _, exists := cfg.Registries[alias]; exists && !overwrite {
				return fmt.Errorf("registry %q already exists. Use --overwrite to replace it", alias)
			}
			var provCount int
			var provNames []string
			if !noValidate {
				count, names, err := validateRegistryURL(url)
				if err != nil {
					return fmt.Errorf("failed to validate registry at %s:\n  %w\n\nUse --no-validate to skip validation", url, err)
				}
				provCount = count
				provNames = names
			}

			entry := config.RegistryEntry{URL: url}
			if err := cfg.AddRegistry(alias, entry); err != nil {
				return err
			}

			if setDefault {
				if err := cfg.SetDefaultRegistry(alias); err != nil {
					return err
				}
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			green := color.New(color.FgGreen)
			dim := color.New(color.Faint)

			green.Printf("✓ Registry %q added\n", alias)
			dim.Printf("  URL: %s\n", url)
			if !noValidate && provCount > 0 {
				display := provNames
				suffix := ""
				if len(display) > 5 {
					display = display[:5]
					suffix = ", ..."
				}
				fmt.Printf("  Found %d provisioners: %s%s\n", provCount, strings.Join(display, ", "), suffix)
			}
			if setDefault {
				fmt.Printf("  Set as default registry\n")
			}
			fmt.Printf("\n  Run \"score-hub registry show %s\" to browse available provisioners.\n", alias)
			return nil
		},
	}

	cmd.Flags().BoolVar(&setDefault, "set-default", false, "Set as default registry for unqualified installs")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "Skip fetching and validating the index.yaml before saving")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing registry with the same alias")

	return cmd
}

// Returns the provisioner count and list of names.
func validateRegistryURL(url string) (int, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid URL: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("URL unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("URL returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to parse as index.yaml
	var idx index.Index
	if err := yaml.Unmarshal(body, &idx); err != nil {
		return 0, nil, fmt.Errorf("invalid index.yaml: YAML parse error: %w", err)
	}

	if idx.APIVersion == "" {
		return 0, nil, fmt.Errorf("invalid index.yaml: missing 'apiVersion' field")
	}

	var names []string
	for _, p := range idx.Provisioners {
		names = append(names, p.Name)
	}

	return len(idx.Provisioners), names, nil
}
