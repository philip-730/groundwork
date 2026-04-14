package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/philip-730/groundwork/internal/config"
	"github.com/philip-730/groundwork/internal/registry"
	tmpl "github.com/philip-730/groundwork/internal/template"
)

// makeRegistry builds a fake registry cache on disk with the given templates.
// Each entry in templates is a map of filename→content to write under
// templates/<name>/. A "template.toml" key is required for discovery to pick it up.
func makeRegistry(t *testing.T, templates map[string]map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, files := range templates {
		dir := filepath.Join(root, "templates", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for fname, content := range files {
			if err := os.WriteFile(filepath.Join(dir, fname), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

const validTOML = `
[template]
name        = "cloud-run-service"
description = "HTTP service on Cloud Run."

[inputs]
service_name = { type = "string", required = true }
region       = { type = "string", default = "us-central1" }

[topology]
requires = ["environments.dev", "environments.prod", "shared.artifact_registry"]
`

func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.toml")
	if err := os.WriteFile(path, []byte(validTOML), 0o644); err != nil {
		t.Fatal(err)
	}
	tmplate, err := tmpl.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmplate.Meta.Name != "cloud-run-service" {
		t.Errorf("name = %q", tmplate.Meta.Name)
	}
	if tmplate.Meta.Description != "HTTP service on Cloud Run." {
		t.Errorf("description = %q", tmplate.Meta.Description)
	}
	inp, ok := tmplate.Inputs["service_name"]
	if !ok {
		t.Fatal("inputs.service_name not found")
	}
	if !inp.Required {
		t.Error("service_name should be required")
	}
	if inp.Type != "string" {
		t.Errorf("service_name type = %q", inp.Type)
	}
	inp2 := tmplate.Inputs["region"]
	if inp2.Default != "us-central1" {
		t.Errorf("region default = %q", inp2.Default)
	}
	if tmplate.Topology.Requires[0] != "environments.dev" {
		t.Errorf("topology.requires[0] = %q", tmplate.Topology.Requires[0])
	}
}

func TestLoad_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.toml")
	if err := os.WriteFile(path, []byte(`[template]`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := tmpl.Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoad_UnknownKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.toml")
	content := `
[template]
name = "x"
surprise = "boom"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := tmpl.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestDiscover_Basic(t *testing.T) {
	root := makeRegistry(t, map[string]map[string]string{
		"cloud-run-service": {"template.toml": validTOML, "main.tf.tmpl": ""},
		"other-service":     {"template.toml": validTOML},
	})

	templates, errs := tmpl.Discover("internal", root)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(templates) != 2 {
		t.Errorf("len(templates) = %d, want 2", len(templates))
	}
	for _, tt := range templates {
		if tt.RegistryName != "internal" {
			t.Errorf("RegistryName = %q, want %q", tt.RegistryName, "internal")
		}
		if tt.Dir == "" {
			t.Error("Dir should be set")
		}
	}
}

func TestDiscover_SkipsNonTemplateDirs(t *testing.T) {
	root := makeRegistry(t, map[string]map[string]string{
		"has-toml":    {"template.toml": validTOML},
		"no-toml":     {"README.md": "hello"},
		"nested-only": {},
	})

	templates, errs := tmpl.Discover("r", root)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(templates) != 1 {
		t.Errorf("len(templates) = %d, want 1", len(templates))
	}
}

func TestDiscover_CollectsBadTOMLAsError(t *testing.T) {
	root := makeRegistry(t, map[string]map[string]string{
		"good": {"template.toml": validTOML},
		"bad":  {"template.toml": "this is not valid toml }"},
	})

	templates, errs := tmpl.Discover("r", root)
	if len(errs) != 1 {
		t.Errorf("len(errs) = %d, want 1 (bad template should be collected, not fatal)", len(errs))
	}
	if len(templates) != 1 {
		t.Errorf("len(templates) = %d, want 1 (good template should still be returned)", len(templates))
	}
}

func TestDiscover_NoTemplatesDir(t *testing.T) {
	root := t.TempDir() // no templates/ subdirectory
	templates, errs := tmpl.Discover("r", root)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(templates) != 0 {
		t.Errorf("expected no templates, got %d", len(templates))
	}
}

func TestDiscoverAll_SkipsUnsynced(t *testing.T) {
	cacheRoot := t.TempDir()
	// "internal" is not synced (no .git dir), "local" has been faked as synced.
	localDir := registry.CacheDir(cacheRoot, "local")
	if err := os.MkdirAll(filepath.Join(localDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a template into local.
	tmplDir := filepath.Join(localDir, "templates", "my-svc")
	if err := os.MkdirAll(tmplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmplDir, "template.toml"), []byte(validTOML), 0o644); err != nil {
		t.Fatal(err)
	}

	regs := []config.Registry{
		{Name: "internal", URL: "git@github.com:x/y"},
		{Name: "local", URL: "/tmp/local"},
	}

	templates, errs := tmpl.DiscoverAll(regs, cacheRoot)
	// Should have one error (internal not synced) and one template (local).
	if len(errs) != 1 {
		t.Errorf("len(errs) = %d, want 1", len(errs))
	}
	if len(templates) != 1 {
		t.Errorf("len(templates) = %d, want 1", len(templates))
	}
}

func TestFind_ByName(t *testing.T) {
	t1 := &tmpl.Template{Meta: tmpl.Meta{Name: "cloud-run-service"}, RegistryName: "internal"}
	t2 := &tmpl.Template{Meta: tmpl.Meta{Name: "bigquery-pipeline"}, RegistryName: "internal"}

	got, err := tmpl.Find([]*tmpl.Template{t1, t2}, "cloud-run-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != t1 {
		t.Error("Find returned wrong template")
	}
}

func TestFind_ByRegistrySlashName(t *testing.T) {
	t1 := &tmpl.Template{Meta: tmpl.Meta{Name: "svc"}, RegistryName: "alpha"}
	t2 := &tmpl.Template{Meta: tmpl.Meta{Name: "svc"}, RegistryName: "beta"}

	got, err := tmpl.Find([]*tmpl.Template{t1, t2}, "beta/svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != t2 {
		t.Error("Find returned wrong template")
	}
}

func TestFind_Ambiguous(t *testing.T) {
	t1 := &tmpl.Template{Meta: tmpl.Meta{Name: "svc"}, RegistryName: "alpha"}
	t2 := &tmpl.Template{Meta: tmpl.Meta{Name: "svc"}, RegistryName: "beta"}

	_, err := tmpl.Find([]*tmpl.Template{t1, t2}, "svc")
	if err == nil {
		t.Fatal("expected error for ambiguous name")
	}
}

func TestFind_NotFound(t *testing.T) {
	_, err := tmpl.Find([]*tmpl.Template{}, "missing")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}
