package cmd

import (
	"os"

	"github.com/philip-730/groundwork/internal/config"
	"github.com/spf13/cobra"
)

var (
	configFlag string         // --config override
	cfg        *config.Config // loaded once in PersistentPreRunE, available to all subcommands
)

var rootCmd = &cobra.Command{
	Use:   "groundwork",
	Short: "A code generator for GCP service scaffolding",
	Long: `Groundwork generates files for GCP service archetypes from registry templates.
It is topology-aware, resolving real project IDs and resource references
rather than leaving placeholders for you to fill in manually.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if configFlag != "" {
			cfg, err = config.Load(configFlag)
		} else {
			cfg, err = config.LoadGlobal()
		}
		return err
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "path to config file (default: ~/.groundwork/config.toml)")

	rootCmd.AddCommand(scaffoldCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(initCmd)
}
