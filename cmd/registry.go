package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage template registries",
	Long:  `Commands for syncing and inspecting template registries defined in groundwork.toml.`,
	// No RunE — subcommands only. Print usage when called bare.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var registrySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Clone or pull all configured registries",
	Long:  `Fetches the latest commits for every registry listed in groundwork.toml into the local cache.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("registry sync: (not yet implemented)")
		return nil
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates across all registries",
	Long:  `Prints all templates discovered in the local registry cache.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("registry list: (not yet implemented)")
		return nil
	},
}

var registryInspectCmd = &cobra.Command{
	Use:   "inspect <template>",
	Short: "Show inputs and topology requirements for a template",
	Long:  `Displays the template.toml contents for the named template in a human-readable form.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("registry inspect: template=%s (not yet implemented)\n", args[0])
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registrySyncCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryInspectCmd)
}
