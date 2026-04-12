package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/philip-730/groundwork/internal/config"
)

// writeToml writes content to groundwork.toml in a temp dir and returns the
// full path to the file.
func writeToml(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "groundwork.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeToml: %v", err)
	}
	return path
}

const fullConfig = `
[workspace]
name = "my-org"

[[registries]]
name = "internal"
url  = "git@github.com:my-org/groundwork-registry"

[topologies.primary]
default = true

  [topologies.primary.environments]
  dev  = "my-org-dev"
  prod = "my-org-prod"

  [topologies.primary.shared]
  project = "my-org-shared"
  region  = "us-central1"

    [topologies.primary.shared.artifact_registry]
    location   = "us-central1"
    repository = "my-org-images"
`

func TestLoad_Full(t *testing.T) {
	path := writeToml(t, fullConfig)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workspace.Name != "my-org" {
		t.Errorf("workspace.name = %q, want %q", cfg.Workspace.Name, "my-org")
	}
	if len(cfg.Registries) != 1 {
		t.Fatalf("len(registries) = %d, want 1", len(cfg.Registries))
	}
	if cfg.Registries[0].Name != "internal" {
		t.Errorf("registries[0].name = %q, want %q", cfg.Registries[0].Name, "internal")
	}
	if cfg.Registries[0].URL != "git@github.com:my-org/groundwork-registry" {
		t.Errorf("registries[0].url = %q", cfg.Registries[0].URL)
	}

	topo, ok := cfg.Topologies["primary"]
	if !ok {
		t.Fatal("topology 'primary' not found")
	}
	if !topo.Default {
		t.Error("topology primary: default should be true")
	}
	if topo.Environments["dev"] != "my-org-dev" {
		t.Errorf("environments.dev = %q, want %q", topo.Environments["dev"], "my-org-dev")
	}
	if topo.Environments["prod"] != "my-org-prod" {
		t.Errorf("environments.prod = %q, want %q", topo.Environments["prod"], "my-org-prod")
	}
	if topo.Shared.Project != "my-org-shared" {
		t.Errorf("shared.project = %q, want %q", topo.Shared.Project, "my-org-shared")
	}
	if topo.Shared.Region != "us-central1" {
		t.Errorf("shared.region = %q, want %q", topo.Shared.Region, "us-central1")
	}
	if topo.Shared.ArtifactRegistry.Location != "us-central1" {
		t.Errorf("shared.artifact_registry.location = %q", topo.Shared.ArtifactRegistry.Location)
	}
	if topo.Shared.ArtifactRegistry.Repository != "my-org-images" {
		t.Errorf("shared.artifact_registry.repository = %q", topo.Shared.ArtifactRegistry.Repository)
	}
}

func TestLoad_Path(t *testing.T) {
	path := writeToml(t, fullConfig)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Path != path {
		t.Errorf("cfg.Path = %q, want %q", cfg.Path, path)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	path := writeToml(t, "this is not [ valid toml }")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestLoad_UnknownKey(t *testing.T) {
	path := writeToml(t, `
[workspace]
name = "x"
unexpected_key = "boom"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
}

var validateCases = []struct {
	name    string
	content string
	wantErr string
}{
	{
		name: "missing workspace name",
		content: `
[workspace]
name = ""
`,
		wantErr: "workspace.name is required",
	},
	{
		name: "registry missing url",
		content: `
[workspace]
name = "x"

[[registries]]
name = "foo"
`,
		wantErr: "url is required",
	},
	{
		name: "registry missing name",
		content: `
[workspace]
name = "x"

[[registries]]
url = "git@github.com:foo/bar"
`,
		wantErr: "name is required",
	},
	{
		name: "multiple topologies no default",
		content: `
[workspace]
name = "x"

[topologies.a]
[topologies.b]
`,
		wantErr: "none has default = true",
	},
	{
		name: "multiple topologies two defaults",
		content: `
[workspace]
name = "x"

[topologies.a]
default = true

[topologies.b]
default = true
`,
		wantErr: "at most one is allowed",
	},
	{
		name: "valid single topology no default field",
		content: `
[workspace]
name = "x"

[topologies.primary]
`,
		wantErr: "",
	},
	{
		name: "valid multiple topologies one default",
		content: `
[workspace]
name = "x"

[topologies.a]
default = true

[topologies.b]
`,
		wantErr: "",
	},
}

func TestValidate(t *testing.T) {
	for _, tc := range validateCases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeToml(t, tc.content)
			_, err := config.Load(path)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if msg := err.Error(); msg == "" {
				t.Errorf("error message is empty, want substring %q", tc.wantErr)
			} else if !contains(msg, tc.wantErr) {
				t.Errorf("error = %q, want substring %q", msg, tc.wantErr)
			}
		})
	}
}

func TestDefaultTopology_Single(t *testing.T) {
	path := writeToml(t, `
[workspace]
name = "x"

[topologies.only]
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	name, _, err := cfg.DefaultTopology()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "only" {
		t.Errorf("name = %q, want %q", name, "only")
	}
}

func TestDefaultTopology_Explicit(t *testing.T) {
	path := writeToml(t, `
[workspace]
name = "x"

[topologies.a]
default = true

[topologies.b]
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	name, _, err := cfg.DefaultTopology()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" {
		t.Errorf("name = %q, want %q", name, "a")
	}
}

func TestFind(t *testing.T) {
	// Write the file in a parent dir, search from a child dir.
	parent := t.TempDir()
	child := filepath.Join(parent, "nested", "dir")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(parent, "groundwork.toml")
	if err := os.WriteFile(configPath, []byte(`[workspace]
name = "x"`), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := config.Find(child)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got != configPath {
		t.Errorf("Find = %q, want %q", got, configPath)
	}
}

func TestFind_NotFound(t *testing.T) {
	// Use a temp dir that definitely has no groundwork.toml
	dir := t.TempDir()
	_, err := config.Find(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
