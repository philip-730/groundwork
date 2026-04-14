// Package topology loads and validates topologies.toml from a registry cache.
package topology

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const filename = "topologies.toml"

// File is the parsed content of a topologies.toml at the root of a registry.
type File struct {
	Topologies map[string]Topology `toml:"topologies"`

	// Path is the absolute path of the file that was loaded.
	// It is set by Load and not present in the TOML file itself.
	Path string `toml:"-"`
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

// Load reads and parses topologies.toml from registryCacheDir.
func Load(registryCacheDir string) (*File, error) {
	path := filepath.Join(registryCacheDir, filename)
	var f File
	md, err := toml.DecodeFile(path, &f)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no topologies.toml found in registry at %s — add one to the registry", registryCacheDir)
		}
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(md.Undecoded()) > 0 {
		return nil, fmt.Errorf("parse %s: unknown keys: %v", path, md.Undecoded())
	}
	f.Path = path
	if err := validate(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

// DefaultTopology returns the name and value of the topology to use when none
// is specified explicitly. Returns an error if no default can be determined.
func (f *File) DefaultTopology() (string, Topology, error) {
	if len(f.Topologies) == 1 {
		for name, t := range f.Topologies {
			return name, t, nil
		}
	}
	for name, t := range f.Topologies {
		if t.Default {
			return name, t, nil
		}
	}
	return "", Topology{}, fmt.Errorf("no default topology: mark one with default = true in %s", f.Path)
}

// Topology looks up a topology by name. Returns an error if not found.
func (f *File) Topology(name string) (Topology, error) {
	t, ok := f.Topologies[name]
	if !ok {
		return Topology{}, fmt.Errorf("topology %q not found in %s", name, f.Path)
	}
	return t, nil
}

func validate(f *File) error {
	var errs []string

	for name := range f.Topologies {
		if strings.TrimSpace(name) == "" {
			errs = append(errs, "topology name must not be empty")
		}
	}

	defaults := 0
	for _, t := range f.Topologies {
		if t.Default {
			defaults++
		}
	}
	if len(f.Topologies) > 1 && defaults == 0 {
		errs = append(errs, "multiple topologies defined but none has default = true")
	}
	if defaults > 1 {
		errs = append(errs, fmt.Sprintf("%d topologies have default = true; at most one is allowed", defaults))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
