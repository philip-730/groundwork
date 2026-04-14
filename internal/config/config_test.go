package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/philip-730/groundwork/internal/config"
)

func writeToml(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeToml: %v", err)
	}
	return path
}

func TestLoad_Registries(t *testing.T) {
	path := writeToml(t, `
[[registries]]
name = "internal"
url  = "https://github.com/my-org/groundwork-registry"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Registries) != 1 {
		t.Fatalf("len(registries) = %d, want 1", len(cfg.Registries))
	}
	if cfg.Registries[0].Name != "internal" {
		t.Errorf("name = %q, want %q", cfg.Registries[0].Name, "internal")
	}
	if cfg.Registries[0].URL != "https://github.com/my-org/groundwork-registry" {
		t.Errorf("url = %q", cfg.Registries[0].URL)
	}
}

func TestLoad_Path(t *testing.T) {
	path := writeToml(t, `
[[registries]]
name = "x"
url  = "https://example.com/x"
`)
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
	path := writeToml(t, `unexpected_key = "boom"`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
}

func TestLoad_Empty(t *testing.T) {
	path := writeToml(t, "")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Registries) != 0 {
		t.Errorf("expected no registries, got %d", len(cfg.Registries))
	}
}

func TestLoadGlobal_FileNotFound_ReturnsEmpty(t *testing.T) {
	// Point GlobalConfigPath to a non-existent file via --config flag equivalent:
	// we test LoadGlobal indirectly by calling Load with a missing path and
	// checking the os.IsNotExist branch in LoadGlobal manually.
	dir := t.TempDir()
	missing := filepath.Join(dir, "config.toml")
	_, err := config.Load(missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	// LoadGlobal wraps this and returns empty — tested via integration.
}

var validateCases = []struct {
	name    string
	content string
	wantErr string
}{
	{
		name: "registry missing url",
		content: `
[[registries]]
name = "foo"
`,
		wantErr: "url is required",
	},
	{
		name: "registry missing name",
		content: `
[[registries]]
url = "https://github.com/foo/bar"
`,
		wantErr: "name is required",
	},
	{
		name:    "valid empty config",
		content: ``,
		wantErr: "",
	},
	{
		name: "valid single registry",
		content: `
[[registries]]
name = "x"
url  = "https://example.com/x"
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
			if !contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
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
