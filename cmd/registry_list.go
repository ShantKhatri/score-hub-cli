package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ShantKhatri/score-hub-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRegistryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured registries",
		Long: `Show all configured registries with their URL, status, and default indicator.

Examples:
  score-hub registry list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()

			if len(cfg.Registries) == 0 {
				fmt.Println("No registries configured.")
				return nil
			}

			aliases := make([]string, 0, len(cfg.Registries))
			for alias := range cfg.Registries {
				aliases = append(aliases, alias)
			}
			sortRegistryAliases(aliases)

			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)
			red := color.New(color.FgRed)
			dim := color.New(color.Faint)

			bold.Printf("%-15s %-60s %-15s\n", "NAME", "URL", "STATUS")
			fmt.Println(strings.Repeat("─", 92))

			for _, alias := range aliases {
				entry := cfg.Registries[alias]

				urlDisplay := entry.URL
				if len(urlDisplay) > 58 {
					urlDisplay = urlDisplay[:55] + "..."
				}

				status := checkRegistryStatus(entry.URL)

				defaultMarker := ""
				if alias == cfg.DefaultRegistry {
					defaultMarker = " ★"
				}

				name := alias + defaultMarker
				if status == "ok" {
					fmt.Printf("%-15s %-60s %s\n", name, urlDisplay, green.Sprint("✓ ok"))
				} else {
					fmt.Printf("%-15s %-60s %s\n", name, urlDisplay, red.Sprint("✗ unreachable"))
				}
			}

			fmt.Println()
			dim.Printf("Default: %s\n", cfg.DefaultRegistry)

			return nil
		},
	}

	return cmd
}

func checkRegistryStatus(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return "unreachable"
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "unreachable"
	}
	resp.Body.Close()

	// Some servers don't support HEAD, try GET
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "unreachable"
		}
		resp, err = client.Do(req)
		if err != nil {
			return "unreachable"
		}
		resp.Body.Close()
	}

	if resp.StatusCode == http.StatusOK {
		return "ok"
	}
	return "unreachable"
}

func sortRegistryAliases(aliases []string) {
	for i, a := range aliases {
		if a == "public" && i != 0 {
			aliases[0], aliases[i] = aliases[i], aliases[0]
			break
		}
	}
	start := 0
	if len(aliases) > 0 && aliases[0] == "public" {
		start = 1
	}
	remaining := aliases[start:]
	for i := 0; i < len(remaining); i++ {
		for j := i + 1; j < len(remaining); j++ {
			if remaining[j] < remaining[i] {
				remaining[i], remaining[j] = remaining[j], remaining[i]
			}
		}
	}
}
