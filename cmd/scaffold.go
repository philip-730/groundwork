package cmd

import (
	"fmt"
	"os"

	"github.com/philip-730/groundwork/internal/registry"
	"github.com/philip-730/groundwork/internal/scaffold"
	tmpl "github.com/philip-730/groundwork/internal/template"
	"github.com/spf13/cobra"
)

var (
	scaffoldTopology string
	scaffoldOut      string
	scaffoldDryRun   bool
	scaffoldVars     []string
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold <template>",
	Short: "Scaffold a new service from a template",
	Long: `Scaffold generates files for a service from the named template.

Steps:
  1. Pull the template from the local registry cache
  2. Prompt for required inputs not supplied via --var
  3. Resolve topology values (project IDs, shared resource references)
  4. Preview files to be written
  5. Ask for confirmation
  6. Write files to the output directory

Run 'groundwork registry sync' first if you haven't already.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		cacheRoot, err := registry.CacheRoot()
		if err != nil {
			return err
		}

		templates, errs := tmpl.DiscoverAll(cfg.Registries, cacheRoot)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "warning: %v\n", e)
		}

		t, err := tmpl.Find(templates, templateName)
		if err != nil {
			return err
		}

		return scaffold.Scaffold(cfg, t, scaffold.Options{
			TopologyName: scaffoldTopology,
			OutDir:       scaffoldOut,
			DryRun:       scaffoldDryRun,
			Vars:         scaffoldVars,
			Stdin:        os.Stdin,
			Stdout:       os.Stdout,
		})
	},
}

func init() {
	scaffoldCmd.Flags().StringVar(&scaffoldTopology, "topology", "", "topology to use (defaults to the topology marked default)")
	scaffoldCmd.Flags().StringVar(&scaffoldOut, "out", "", "output directory (defaults to ./<service_name>)")
	scaffoldCmd.Flags().BoolVar(&scaffoldDryRun, "dry-run", false, "print rendered files without writing")
	scaffoldCmd.Flags().StringArrayVar(&scaffoldVars, "var", nil, "pass input values as key=value (may be repeated; skips interactive prompt)")
}
