package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"

	// Global flags
	flagPlatform string
	flagJSON     bool
	flagNoCache  bool
	flagVerbose  bool
	flagRegistry string
)

var rootCmd = &cobra.Command{
	Use:          "score-hub",
	Short:        "The CLI install layer for the Score community provisioners ecosystem",
	Long:         "score-hub makes Score community provisioners discoverable and installable.",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagPlatform, "platform", "",
		"Score platform: k8s or compose (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false,
		"Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagNoCache, "no-cache", false,
		"Skip local index cache, fetch fresh")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false,
		"Verbose output for debugging")
	rootCmd.PersistentFlags().StringVar(&flagRegistry, "registry", "",
		"Registry URL (overrides config)")

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print score-hub version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("score-hub v%s\n", Version)
	},
}
