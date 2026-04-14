package topology_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/philip-730/groundwork/internal/topology"
)

func writeTopologiesToml(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "topologies.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTopologiesToml: %v", err)
	}
	return dir // return the dir, not the file — Load takes the registry cache dir
}

const fullTopologies = `
[topologies.primary]
default = true

  [topologies.primary.environments]
  dev  = "my-org-dev"
  prod = "my-org-prod"

  [topologies.primary.shared]
  project = "my-org-shared"
  region  = "us-east1"

    [topologies.primary.shared.artifact_registry]
    location   = "us-east1"
    repository = "my-org-images"
`

func TestLoad_Full(t *testing.T) {
	dir := writeTopologiesToml(t, fullTopologies)
	f, err := topology.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	topo, ok := f.Topologies["primary"]
	if !ok {
		t.Fatal("topology 'primary' not found")
	}
	if !topo.Default {
		t.Error("primary: default should be true")
	}
	if topo.Environments["dev"] != "my-org-dev" {
		t.Errorf("environments.dev = %q", topo.Environments["dev"])
	}
	if topo.Shared.Project != "my-org-shared" {
		t.Errorf("shared.project = %q", topo.Shared.Project)
	}
	if topo.Shared.Region != "us-east1" {
		t.Errorf("shared.region = %q", topo.Shared.Region)
	}
	if topo.Shared.ArtifactRegistry.Location != "us-east1" {
		t.Errorf("artifact_registry.location = %q", topo.Shared.ArtifactRegistry.Location)
	}
	if topo.Shared.ArtifactRegistry.Repository != "my-org-images" {
		t.Errorf("artifact_registry.repository = %q", topo.Shared.ArtifactRegistry.Repository)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	dir := t.TempDir() // no topologies.toml inside
	_, err := topology.Load(dir)
	if err == nil {
		t.Fatal("expected error for missing topologies.toml")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	dir := writeTopologiesToml(t, "this is not [ valid toml }")
	_, err := topology.Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestLoad_UnknownKey(t *testing.T) {
	dir := writeTopologiesToml(t, `unexpected_key = "boom"`)
	_, err := topology.Load(dir)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestDefaultTopology_Single(t *testing.T) {
	dir := writeTopologiesToml(t, `[topologies.only]`)
	f, err := topology.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	name, _, err := f.DefaultTopology()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "only" {
		t.Errorf("name = %q, want %q", name, "only")
	}
}

func TestDefaultTopology_Explicit(t *testing.T) {
	dir := writeTopologiesToml(t, `
[topologies.a]
default = true

[topologies.b]
`)
	f, err := topology.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	name, _, err := f.DefaultTopology()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" {
		t.Errorf("name = %q, want %q", name, "a")
	}
}

func TestDefaultTopology_MultipleNoDefault(t *testing.T) {
	dir := writeTopologiesToml(t, `
[topologies.a]
[topologies.b]
`)
	_, err := topology.Load(dir)
	if err == nil {
		t.Fatal("expected validation error for multiple topologies with no default")
	}
}

func TestTopologyLookup_NotFound(t *testing.T) {
	dir := writeTopologiesToml(t, `[topologies.primary]`)
	f, err := topology.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	_, err = f.Topology("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing topology name")
	}
}
