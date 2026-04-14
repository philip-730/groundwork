package scaffold_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philip-730/groundwork/internal/config"
	"github.com/philip-730/groundwork/internal/scaffold"
	tmpl "github.com/philip-730/groundwork/internal/template"
)

// ── ParseVars ────────────────────────────────────────────────────────────────

func TestParseVars_Valid(t *testing.T) {
	got, err := scaffold.ParseVars([]string{"name=my-api", "region=us-east1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["name"] != "my-api" {
		t.Errorf("name = %q", got["name"])
	}
	if got["region"] != "us-east1" {
		t.Errorf("region = %q", got["region"])
	}
}

func TestParseVars_ValueContainsEquals(t *testing.T) {
	got, err := scaffold.ParseVars([]string{"conn=host=localhost;port=5432"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["conn"] != "host=localhost;port=5432" {
		t.Errorf("conn = %q", got["conn"])
	}
}

func TestParseVars_MissingEquals(t *testing.T) {
	_, err := scaffold.ParseVars([]string{"noequalssign"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseVars_EmptyKey(t *testing.T) {
	_, err := scaffold.ParseVars([]string{"=value"})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestParseVars_Empty(t *testing.T) {
	got, err := scaffold.ParseVars(nil)
	if err != nil || len(got) != 0 {
		t.Errorf("ParseVars(nil) = %v, %v", got, err)
	}
}

// ── CollectInputs ────────────────────────────────────────────────────────────

func makeTemplate(inputs map[string]tmpl.Input) *tmpl.Template {
	return &tmpl.Template{
		Meta:   tmpl.Meta{Name: "test-tmpl"},
		Inputs: inputs,
	}
}

func TestCollectInputs_AllProvided(t *testing.T) {
	tt := makeTemplate(map[string]tmpl.Input{
		"service_name": {Type: "string", Required: true},
	})
	provided := map[string]string{"service_name": "my-api"}
	got, err := scaffold.CollectInputs(tt, provided, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["service_name"] != "my-api" {
		t.Errorf("service_name = %q", got["service_name"])
	}
}

func TestCollectInputs_PromptRequired(t *testing.T) {
	tt := makeTemplate(map[string]tmpl.Input{
		"service_name": {Type: "string", Required: true},
	})
	var out bytes.Buffer
	got, err := scaffold.CollectInputs(tt, map[string]string{}, strings.NewReader("my-svc\n"), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["service_name"] != "my-svc" {
		t.Errorf("service_name = %q", got["service_name"])
	}
}

func TestCollectInputs_UsesDefault(t *testing.T) {
	tt := makeTemplate(map[string]tmpl.Input{
		"region": {Type: "string", Default: "us-central1"},
	})
	// Empty line → accept default.
	got, err := scaffold.CollectInputs(tt, map[string]string{}, strings.NewReader("\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["region"] != "us-central1" {
		t.Errorf("region = %q, want us-central1", got["region"])
	}
}

func TestCollectInputs_RequiredStdinClosed(t *testing.T) {
	tt := makeTemplate(map[string]tmpl.Input{
		"service_name": {Type: "string", Required: true},
	})
	_, err := scaffold.CollectInputs(tt, map[string]string{}, strings.NewReader(""), &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error when stdin closed with required input unsatisfied")
	}
}

// ── BuildContext ─────────────────────────────────────────────────────────────

func TestBuildContext(t *testing.T) {
	topo := config.Topology{
		Environments: map[string]string{"dev": "my-dev", "prod": "my-prod"},
		Shared: config.SharedConfig{
			Project: "my-shared",
			Region:  "us-central1",
			ArtifactRegistry: config.ArtifactRegistryConfig{
				Location:   "us-central1",
				Repository: "my-images",
			},
		},
	}
	ctx := scaffold.BuildContext(map[string]string{"service_name": "api"}, "primary", topo)

	if ctx.Topology.Name != "primary" {
		t.Errorf("topology name = %q", ctx.Topology.Name)
	}
	if ctx.Topology.Environments["dev"] != "my-dev" {
		t.Errorf("env dev = %q", ctx.Topology.Environments["dev"])
	}
	if ctx.Topology.Shared.Region != "us-central1" {
		t.Errorf("region = %q", ctx.Topology.Shared.Region)
	}
	if ctx.Topology.Shared.ArtifactRegistry.Repository != "my-images" {
		t.Errorf("repo = %q", ctx.Topology.Shared.ArtifactRegistry.Repository)
	}
	if ctx.Inputs["service_name"] != "api" {
		t.Errorf("input service_name = %q", ctx.Inputs["service_name"])
	}
}

// ── ValidateTopologyKeys ─────────────────────────────────────────────────────

func fullTopo() config.Topology {
	return config.Topology{
		Environments: map[string]string{"dev": "my-dev", "prod": "my-prod"},
		Shared: config.SharedConfig{
			Project: "my-shared",
			Region:  "us-central1",
			ArtifactRegistry: config.ArtifactRegistryConfig{
				Location:   "us-central1",
				Repository: "my-images",
			},
		},
	}
}

func TestValidateTopologyKeys_AllPresent(t *testing.T) {
	keys := []string{
		"environments.dev",
		"environments.prod",
		"shared.project",
		"shared.region",
		"shared.artifact_registry",
		"shared.artifact_registry.location",
		"shared.artifact_registry.repository",
	}
	if err := scaffold.ValidateTopologyKeys(fullTopo(), keys); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTopologyKeys_MissingEnv(t *testing.T) {
	topo := fullTopo()
	topo.Environments = map[string]string{"dev": "my-dev"} // no prod
	err := scaffold.ValidateTopologyKeys(topo, []string{"environments.prod"})
	if err == nil {
		t.Fatal("expected error for missing environments.prod")
	}
}

func TestValidateTopologyKeys_MissingSharedProject(t *testing.T) {
	topo := fullTopo()
	topo.Shared.Project = ""
	err := scaffold.ValidateTopologyKeys(topo, []string{"shared.project"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateTopologyKeys_Empty(t *testing.T) {
	if err := scaffold.ValidateTopologyKeys(fullTopo(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── RenderAll ────────────────────────────────────────────────────────────────

// makeTemplateDir creates a fake template directory with given file contents.
func makeTemplateDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRenderAll_Basic(t *testing.T) {
	dir := makeTemplateDir(t, map[string]string{
		"template.toml":          `[template]\nname="x"`,
		"terraform/main.tf.tmpl": `project = "{{ .Topology.Environments.dev }}"`,
		"static.txt":             "verbatim content",
	})

	ctx := scaffold.BuildContext(
		map[string]string{},
		"primary",
		config.Topology{Environments: map[string]string{"dev": "my-dev"}},
	)
	files, err := scaffold.RenderAll(dir, ctx)
	if err != nil {
		t.Fatalf("RenderAll: %v", err)
	}

	byPath := make(map[string]string)
	for _, f := range files {
		byPath[f.RelPath] = string(f.Content)
	}

	// template.toml must be excluded
	if _, ok := byPath["template.toml"]; ok {
		t.Error("template.toml should be excluded from output")
	}
	// .tmpl extension stripped, content rendered
	if v := byPath["terraform/main.tf"]; v != `project = "my-dev"` {
		t.Errorf("terraform/main.tf = %q", v)
	}
	// verbatim file copied
	if v := byPath["static.txt"]; v != "verbatim content" {
		t.Errorf("static.txt = %q", v)
	}
}

func TestRenderAll_MissingKey(t *testing.T) {
	dir := makeTemplateDir(t, map[string]string{
		"bad.tf.tmpl": `{{ .Inputs.does_not_exist }}`,
	})
	// missingkey=error means this should fail
	ctx := scaffold.BuildContext(map[string]string{}, "p", config.Topology{})
	_, err := scaffold.RenderAll(dir, ctx)
	if err == nil {
		t.Fatal("expected error for missing template key")
	}
}

// ── Scaffold (end-to-end dry-run) ────────────────────────────────────────────

func makeFullTemplate(t *testing.T) *tmpl.Template {
	t.Helper()
	dir := makeTemplateDir(t, map[string]string{
		"template.toml":              "[template]\nname=\"cloud-run-service\"\n",
		"terraform/main.tf.tmpl":     `project = "{{ .Topology.Environments.dev }}"` + "\nname = \"{{ .Inputs.service_name }}\"",
		"cloudbuild/build.yaml.tmpl": `steps:\n- name: "{{ .Topology.Shared.ArtifactRegistry.Location }}-docker.pkg.dev/{{ .Topology.Shared.Project }}/{{ .Topology.Shared.ArtifactRegistry.Repository }}/{{ .Inputs.service_name }}"`,
	})
	return &tmpl.Template{
		Meta:   tmpl.Meta{Name: "cloud-run-service", Description: "Cloud Run service"},
		Inputs: map[string]tmpl.Input{"service_name": {Type: "string", Required: true}},
		Topology: tmpl.TopologyRequirement{
			Requires: []string{"environments.dev", "shared.artifact_registry"},
		},
		RegistryName: "internal",
		Dir:          dir,
	}
}

func makeCfg(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		Workspace: config.Workspace{Name: "my-org"},
		Topologies: map[string]config.Topology{
			"primary": {
				Default:      true,
				Environments: map[string]string{"dev": "my-org-dev", "prod": "my-org-prod"},
				Shared: config.SharedConfig{
					Project: "my-org-shared",
					Region:  "us-central1",
					ArtifactRegistry: config.ArtifactRegistryConfig{
						Location:   "us-central1",
						Repository: "my-org-images",
					},
				},
			},
		},
		Path: "/fake/groundwork.toml",
	}
}

func TestScaffold_DryRun(t *testing.T) {
	tt := makeFullTemplate(t)
	cfg := makeCfg(t)

	var out bytes.Buffer
	err := scaffold.Scaffold(cfg, tt, scaffold.Options{
		DryRun: true,
		Vars:   []string{"service_name=my-api"},
		Stdin:  strings.NewReader(""),
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	outStr := out.String()
	if !strings.Contains(outStr, "terraform/main.tf") {
		t.Errorf("output missing terraform/main.tf:\n%s", outStr)
	}
	if !strings.Contains(outStr, "my-org-dev") {
		t.Errorf("topology value not rendered:\n%s", outStr)
	}
	if !strings.Contains(outStr, "my-api") {
		t.Errorf("input value not rendered:\n%s", outStr)
	}
}

func TestScaffold_WritesFiles(t *testing.T) {
	tt := makeFullTemplate(t)
	cfg := makeCfg(t)
	outDir := t.TempDir()

	var out bytes.Buffer
	err := scaffold.Scaffold(cfg, tt, scaffold.Options{
		OutDir: outDir,
		Vars:   []string{"service_name=my-api"},
		Stdin:  strings.NewReader("y\n"),
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	// Verify rendered file exists and has correct content.
	mainTF, err := os.ReadFile(filepath.Join(outDir, "terraform", "main.tf"))
	if err != nil {
		t.Fatalf("read main.tf: %v", err)
	}
	if !strings.Contains(string(mainTF), "my-org-dev") {
		t.Errorf("main.tf missing topology value:\n%s", mainTF)
	}
	if !strings.Contains(string(mainTF), "my-api") {
		t.Errorf("main.tf missing input value:\n%s", mainTF)
	}
}

func TestScaffold_Abort(t *testing.T) {
	tt := makeFullTemplate(t)
	cfg := makeCfg(t)
	outDir := t.TempDir()

	var out bytes.Buffer
	err := scaffold.Scaffold(cfg, tt, scaffold.Options{
		OutDir: outDir,
		Vars:   []string{"service_name=my-api"},
		Stdin:  strings.NewReader("n\n"),
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// outDir should be empty — nothing written.
	entries, _ := os.ReadDir(outDir)
	if len(entries) != 0 {
		t.Errorf("expected no files after abort, got %d", len(entries))
	}
}

func TestScaffold_TopologyValidationFails(t *testing.T) {
	dir := makeTemplateDir(t, map[string]string{
		"template.toml": "[template]\nname=\"x\"\n",
		"main.tf.tmpl":  `{{ .Topology.Environments.staging }}`,
	})
	tt := &tmpl.Template{
		Meta:     tmpl.Meta{Name: "x"},
		Inputs:   map[string]tmpl.Input{},
		Topology: tmpl.TopologyRequirement{Requires: []string{"environments.staging"}},
		Dir:      dir,
	}
	cfg := makeCfg(t) // topology has dev+prod, not staging

	var out bytes.Buffer
	err := scaffold.Scaffold(cfg, tt, scaffold.Options{
		DryRun: true,
		Stdin:  strings.NewReader(""),
		Stdout: &out,
	})
	if err == nil {
		t.Fatal("expected topology validation error")
	}
	if !strings.Contains(err.Error(), "environments.staging") {
		t.Errorf("error = %q, want mention of environments.staging", err.Error())
	}
}
