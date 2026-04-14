// Package config loads and validates groundwork.toml.
//
// The on-disk format uses [topologies.<name>] table keys so that topology
// names are unambiguous. For example:
//
//	[topologies.primary]
//	default = true
//
//	[topologies.primary.environments]
//	dev  = "my-org-dev"
//	prod = "my-org-prod"
//
//	[topologies.primary.shared]
//	project = "my-org-shared"
//	region  = "us-central1"
//
//	  [topologies.primary.shared.artifact_registry]
//	  location   = "us-central1"
//	  repository = "my-org-images"
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const filename = "groundwork.toml"

// Config is the top-level structure of groundwork.toml.
type Config struct {
	Workspace  Workspace           `toml:"workspace"`
	Registries []Registry          `toml:"registries"`
	Topologies map[string]Topology `toml:"topologies"`

	// Path is the absolute path of the file that was loaded.
	// It is set by Load and not present in the TOML file itself.
	Path string `toml:"-"`
}

// Workspace holds workspace-wide metadata.
type Workspace struct {
	Name string `toml:"name"`
}

// Registry is a named git remote that contains templates.
type Registry struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// Topology maps a named GCP org structure used during scaffolding.
type Topology struct {
	Default      bool              `toml:"default"`
	Environments map[string]string `toml:"environments"`
	Shared       SharedConfig      `toml:"shared"`
}

// SharedConfig holds references to shared GCP infrastructure.
type SharedConfig struct {
	Project          string                 `toml:"project"`
	Region           string                 `toml:"region"`
	ArtifactRegistry ArtifactRegistryConfig `toml:"artifact_registry"`
}

// ArtifactRegistryConfig identifies the shared container image repository.
type ArtifactRegistryConfig struct {
	Location   string `toml:"location"`
	Repository string `toml:"repository"`
}

// DefaultTopology returns the name and value of the topology that should be
// used when none is specified explicitly. It returns an error if no default
// can be determined (zero or more than one topology marked default=true).
func (c *Config) DefaultTopology() (string, Topology, error) {
	if len(c.Topologies) == 1 {
		for name, t := range c.Topologies {
			return name, t, nil
		}
	}
	for name, t := range c.Topologies {
		if t.Default {
			return name, t, nil
		}
	}
	return "", Topology{}, fmt.Errorf("no default topology: mark one topology with default = true in %s", c.Path)
}

// Topology looks up a topology by name. It returns an error if not found.
func (c *Config) Topology(name string) (Topology, error) {
	t, ok := c.Topologies[name]
	if !ok {
		return Topology{}, fmt.Errorf("topology %q not found in %s", name, c.Path)
	}
	return t, nil
}

// Load reads and parses groundwork.toml at path, then validates it.
func Load(path string) (*Config, error) {
	var cfg Config
	md, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(md.Undecoded()) > 0 {
		return nil, fmt.Errorf("parse %s: unknown keys: %v", path, md.Undecoded())
	}
	cfg.Path = path
	if err := Validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Find walks up the directory tree from dir, looking for groundwork.toml.
// It returns the absolute path of the first file found, or an error if none
// exists before the filesystem root.
func Find(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(abs, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("%s not found (searched from %s to filesystem root)", filename, dir)
}

// FindAndLoad locates groundwork.toml starting from dir and loads it.
func FindAndLoad(dir string) (*Config, error) {
	path, err := Find(dir)
	if err != nil {
		return nil, err
	}
	return Load(path)
}
