package scaffold

import "github.com/philip-730/groundwork/internal/config"

// RenderContext is the data available to every template file during rendering.
//
// In a .tmpl file, reference values like:
//
//	{{ .Inputs.service_name }}
//	{{ .Topology.Environments.dev }}
//	{{ .Topology.Shared.Region }}
//	{{ .Topology.Shared.ArtifactRegistry.Repository }}
type RenderContext struct {
	Inputs   map[string]string
	Topology TopologyContext
}

// TopologyContext mirrors config.Topology in a form convenient for template
// authors. Because text/template supports map key access via dot notation,
// dynamic environment names (dev, prod, staging…) are reachable directly:
// {{ .Topology.Environments.dev }}
type TopologyContext struct {
	Name         string
	Environments map[string]string
	Shared       SharedContext
}

// SharedContext holds references to shared GCP infrastructure.
type SharedContext struct {
	Project          string
	Region           string
	ArtifactRegistry ArtifactRegistryContext
}

// ArtifactRegistryContext identifies the shared container image repository.
type ArtifactRegistryContext struct {
	Location   string
	Repository string
}

// BuildContext converts a topology name + config.Topology and the collected
// input values into a RenderContext ready for template execution.
func BuildContext(inputs map[string]string, topologyName string, t config.Topology) RenderContext {
	return RenderContext{
		Inputs: inputs,
		Topology: TopologyContext{
			Name:         topologyName,
			Environments: t.Environments,
			Shared: SharedContext{
				Project: t.Shared.Project,
				Region:  t.Shared.Region,
				ArtifactRegistry: ArtifactRegistryContext{
					Location:   t.Shared.ArtifactRegistry.Location,
					Repository: t.Shared.ArtifactRegistry.Repository,
				},
			},
		},
	}
}
