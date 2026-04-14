package cmd

import (
	"fmt"
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
	// Config is loaded before every subcommand. Commands that don't need a
	// config file (e.g. help, completion) are not affected because cobra only
	// calls PersistentPreRunE for commands that have a RunE/Run of their own.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if configFlag != "" {
			cfg, err = config.Load(configFlag)
		} else {
			wd, wdErr := os.Getwd()
			if wdErr != nil {
				return fmt.Errorf("get working directory: %w", wdErr)
			}
			cfg, err = config.FindAndLoad(wd)
		}
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "path to groundwork.toml (default: search up from current directory)")

	rootCmd.AddCommand(scaffoldCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(topologyCmd)
}
