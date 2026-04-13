package cmd

import (
	"fmt"
	"os"

	"github.com/philip-730/groundwork/internal/registry"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage template registries",
	Long:  `Commands for syncing and inspecting template registries defined in groundwork.toml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var registrySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Clone or pull all configured registries",
	Long: `Fetches the latest commits for every registry listed in groundwork.toml.

Each registry is cloned on first run and pulled on subsequent runs.
Clones are stored in ~/.groundwork/registries/<name>.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(cfg.Registries) == 0 {
			fmt.Fprintln(os.Stderr, "no registries configured in groundwork.toml")
			return nil
		}

		cacheRoot, err := registry.CacheRoot()
		if err != nil {
			return err
		}

		_, err = registry.SyncAll(cfg.Registries, cacheRoot, os.Stdout)
		return err
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates across all registries",
	Long:  `Prints all templates discovered in the local registry cache.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("registry list: (not yet implemented — coming in step 4: template discovery)")
		return nil
	},
}

var registryInspectCmd = &cobra.Command{
	Use:   "inspect <template>",
	Short: "Show inputs and topology requirements for a template",
	Long:  `Displays the template.toml contents for the named template in a human-readable form.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("registry inspect: template=%s (not yet implemented — coming in step 4: template discovery)\n", args[0])
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registrySyncCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryInspectCmd)
}
