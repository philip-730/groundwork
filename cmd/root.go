package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "groundwork",
	Short: "A code generator for GCP service scaffolding",
	Long: `Groundwork generates files for GCP service archetypes from registry templates.
It is topology-aware, resolving real project IDs and resource references
rather than leaving placeholders for you to fill in manually.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(topologyCmd)
}
