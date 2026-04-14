package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/philip-730/groundwork/internal/registry"
	tmpl "github.com/philip-730/groundwork/internal/template"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage template registries",
	Long:  `Commands for syncing and inspecting template registries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var registrySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Clone or pull all configured registries",
	Long: `Fetches the latest commits for every configured registry.

Each registry is cloned on first run and pulled on subsequent runs.
Clones are stored in ~/.groundwork/registries/<name>.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(cfg.Registries) == 0 {
			fmt.Fprintln(os.Stderr, "no registries configured — run: groundwork init <url>")
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
		cacheRoot, err := registry.CacheRoot()
		if err != nil {
			return err
		}

		templates, errs := tmpl.DiscoverAll(cfg.Registries, cacheRoot)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "warning: %v\n", e)
		}

		if len(templates) == 0 {
			fmt.Println("no templates found — run: groundwork registry sync")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "REGISTRY\tTEMPLATE\tDESCRIPTION")
		for _, t := range templates {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.RegistryName, t.Meta.Name, t.Meta.Description)
		}
		return w.Flush()
	},
}

var registryInspectCmd = &cobra.Command{
	Use:   "inspect <template>",
	Short: "Show inputs and topology requirements for a template",
	Long: `Displays the inputs and topology requirements for the named template.

<template> can be a plain name (e.g. cloud-run-service) or qualified with
the registry (e.g. internal/cloud-run-service) to resolve ambiguity when
the same name exists in multiple registries.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheRoot, err := registry.CacheRoot()
		if err != nil {
			return err
		}

		templates, errs := tmpl.DiscoverAll(cfg.Registries, cacheRoot)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "warning: %v\n", e)
		}

		t, err := tmpl.Find(templates, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Template:    %s\n", t.Meta.Name)
		fmt.Printf("Registry:    %s\n", t.RegistryName)
		fmt.Printf("Description: %s\n", t.Meta.Description)
		fmt.Printf("Directory:   %s\n", t.Dir)

		fmt.Println("\nInputs:")
		if len(t.Inputs) == 0 {
			fmt.Println("  (none)")
		} else {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			for name, inp := range t.Inputs {
				qualifier := ""
				switch {
				case inp.Required:
					qualifier = "required"
				case inp.Default != "":
					qualifier = "default: " + inp.Default
				default:
					qualifier = "optional"
				}
				fmt.Fprintf(w, "  %s\t%s\t%s\n", name, inp.Type, qualifier)
			}
			w.Flush()
		}

		fmt.Println("\nTopology requirements:")
		if len(t.Topology.Requires) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, req := range t.Topology.Requires {
				fmt.Printf("  %s\n", req)
			}
		}

		return nil
	},
}

func init() {
	registryCmd.AddCommand(registrySyncCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryInspectCmd)
}
