package cmd

import (
	"fmt"

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

It prompts for required inputs, resolves topology values, previews the
files to be written, asks for confirmation, then writes output.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		template := args[0]
		fmt.Printf("scaffold: template=%s topology=%s out=%s dry-run=%v vars=%v\n",
			template, scaffoldTopology, scaffoldOut, scaffoldDryRun, scaffoldVars)
		fmt.Println("(not yet implemented)")
		return nil
	},
}

func init() {
	scaffoldCmd.Flags().StringVar(&scaffoldTopology, "topology", "", "topology to use (defaults to the topology marked default)")
	scaffoldCmd.Flags().StringVar(&scaffoldOut, "out", "", "output directory (defaults to ./<service_name>)")
	scaffoldCmd.Flags().BoolVar(&scaffoldDryRun, "dry-run", false, "print rendered files without writing")
	scaffoldCmd.Flags().StringArrayVar(&scaffoldVars, "var", nil, "pass input values as key=value (may be repeated; skips interactive prompt)")
}
