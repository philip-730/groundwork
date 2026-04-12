package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var topologyCmd = &cobra.Command{
	Use:   "topology",
	Short: "Inspect and validate topology configuration",
	Long:  `Commands for listing and validating topologies defined in groundwork.toml.`,
	// No RunE — subcommands only. Print usage when called bare.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var topologyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured topologies",
	Long:  `Prints all topologies defined in groundwork.toml, marking the default.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("topology list: (not yet implemented)")
		return nil
	},
}

var topologyValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate topology configuration against GCP",
	Long: `Checks that all project IDs and shared resource references in groundwork.toml
exist and are accessible using Application Default Credentials (ADC).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("topology validate: (not yet implemented)")
		return nil
	},
}

func init() {
	topologyCmd.AddCommand(topologyListCmd)
	topologyCmd.AddCommand(topologyValidateCmd)
}
