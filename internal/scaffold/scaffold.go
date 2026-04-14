package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/philip-730/groundwork/internal/topology"
	tmpl "github.com/philip-730/groundwork/internal/template"
)

// Options controls scaffold behaviour.
type Options struct {
	// TopologyName selects which topology to use. Empty means use the default.
	TopologyName string
	// OutDir is where files are written. Empty defaults to ./<service_name>
	// (or ./<template-name> if service_name is not an input).
	OutDir string
	// DryRun prints rendered file contents without writing anything.
	DryRun bool
	// Vars are pre-supplied input values in "key=value" form (from --var flags).
	Vars []string
	// Stdin is used for interactive prompts. Defaults to os.Stdin in the cmd layer.
	Stdin io.Reader
	// Stdout is used for all output. Defaults to os.Stdout in the cmd layer.
	Stdout io.Writer
}

// Scaffold runs the full scaffold flow for the given template.
func Scaffold(topos *topology.File, t *tmpl.Template, opts Options) error {
	// 1. Parse --var flags.
	provided, err := ParseVars(opts.Vars)
	if err != nil {
		return err
	}

	// 2. Resolve topology.
	topoName, topo, err := resolveTopology(topos, opts.TopologyName)
	if err != nil {
		return err
	}

	// 3. Validate that the topology has everything the template declared it needs.
	if err := ValidateTopologyKeys(topo, t.Topology.Requires); err != nil {
		return fmt.Errorf("topology %q does not satisfy template requirements: %w", topoName, err)
	}

	// 4. Collect inputs interactively for anything not supplied via --var.
	fmt.Fprintln(opts.Stdout, "Inputs:")
	inputs, err := CollectInputs(t, provided, opts.Stdin, opts.Stdout)
	if err != nil {
		return err
	}

	// 5. Build render context.
	ctx := BuildContext(inputs, topoName, topo)

	// 6. Render template files.
	rendered, err := RenderAll(t.Dir, ctx)
	if err != nil {
		return err
	}

	// 7. Determine output directory.
	outDir := opts.OutDir
	if outDir == "" {
		if name, ok := inputs["service_name"]; ok && name != "" {
			outDir = "./" + name
		} else {
			outDir = "./" + t.Meta.Name
		}
	}

	// 8. Dry-run: print rendered contents and exit.
	if opts.DryRun {
		return printDryRun(opts.Stdout, rendered)
	}

	// 9. Show file list and ask for confirmation.
	fmt.Fprintf(opts.Stdout, "\nFiles to be written to %s:\n\n", outDir)
	for _, f := range rendered {
		fmt.Fprintf(opts.Stdout, "  %s\n", f.RelPath)
	}

	fmt.Fprintf(opts.Stdout, "\nProceed? [y/N] ")
	scanner := bufio.NewScanner(opts.Stdin)
	if !scanner.Scan() || !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
		fmt.Fprintln(opts.Stdout, "aborted.")
		return nil
	}

	// 10. Write files.
	if err := WriteAll(outDir, rendered); err != nil {
		return err
	}
	fmt.Fprintf(opts.Stdout, "wrote %d file(s) to %s\n", len(rendered), outDir)
	return nil
}

// ValidateTopologyKeys checks that every key path declared in required is
// non-empty in the given topology. Key paths use dot notation:
//
//	environments.dev
//	environments.prod
//	shared.project
//	shared.region
//	shared.artifact_registry
//	shared.artifact_registry.location
//	shared.artifact_registry.repository
func ValidateTopologyKeys(topo topology.Topology, required []string) error {
	var missing []string
	for _, key := range required {
		if !topologyKeyPresent(topo, key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required topology keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

func topologyKeyPresent(topo topology.Topology, key string) bool {
	parts := strings.SplitN(key, ".", 2)
	switch parts[0] {
	case "environments":
		if len(parts) < 2 {
			return len(topo.Environments) > 0
		}
		v, ok := topo.Environments[parts[1]]
		return ok && v != ""
	case "shared":
		if len(parts) < 2 {
			return topo.Shared.Project != "" || topo.Shared.Region != ""
		}
		return sharedKeyPresent(topo.Shared, parts[1])
	}
	return false
}

func sharedKeyPresent(shared topology.SharedConfig, key string) bool {
	parts := strings.SplitN(key, ".", 2)
	switch parts[0] {
	case "project":
		return shared.Project != ""
	case "region":
		return shared.Region != ""
	case "artifact_registry":
		if len(parts) < 2 {
			return shared.ArtifactRegistry.Location != "" && shared.ArtifactRegistry.Repository != ""
		}
		switch parts[1] {
		case "location":
			return shared.ArtifactRegistry.Location != ""
		case "repository":
			return shared.ArtifactRegistry.Repository != ""
		}
	}
	return false
}

func resolveTopology(topos *topology.File, name string) (string, topology.Topology, error) {
	if name != "" {
		t, err := topos.Topology(name)
		return name, t, err
	}
	return topos.DefaultTopology()
}

func printDryRun(w io.Writer, files []RenderedFile) error {
	for i, f := range files {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "--- %s ---\n", f.RelPath)
		w.Write(f.Content) //nolint:errcheck
		if len(f.Content) > 0 && f.Content[len(f.Content)-1] != '\n' {
			fmt.Fprintln(w)
		}
	}
	return nil
}
