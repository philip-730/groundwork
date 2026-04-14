// Package template handles parsing and validation of template.toml files
// found inside registry clones.
//
// A registry is expected to have the layout:
//
//	<registry-cache>/
//	  templates/
//	    <template-name>/
//	      template.toml
//	      **/*.tmpl
package template

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

const Filename = "template.toml"

// Template is the parsed, validated content of a template.toml together with
// the discovery metadata (which registry it came from, where it lives on disk).
type Template struct {
	Meta     Meta                `toml:"template"`
	Inputs   map[string]Input    `toml:"inputs"`
	Topology TopologyRequirement `toml:"topology"`

	// Set by discovery, not present in template.toml.
	RegistryName string `toml:"-"`
	Dir          string `toml:"-"`
}

// Meta holds the human-readable identity of the template.
type Meta struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
}

// Input describes a single scaffold input variable.
type Input struct {
	Type     string `toml:"type"`
	Required bool   `toml:"required"`
	Default  string `toml:"default"`
}

// TopologyRequirement lists the topology keys a template needs resolved
// before rendering (e.g. "environments.dev", "shared.artifact_registry").
type TopologyRequirement struct {
	Requires []string `toml:"requires"`
}

// Load reads and parses a template.toml at path, then validates it.
func Load(path string) (*Template, error) {
	var t Template
	md, err := toml.DecodeFile(path, &t)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(md.Undecoded()) > 0 {
		return nil, fmt.Errorf("parse %s: unknown keys: %v", path, md.Undecoded())
	}
	if err := validate(&t, path); err != nil {
		return nil, err
	}
	return &t, nil
}

func validate(t *Template, path string) error {
	var errs []string
	if strings.TrimSpace(t.Meta.Name) == "" {
		errs = append(errs, "template.name is required")
	}
	for name, inp := range t.Inputs {
		if strings.TrimSpace(inp.Type) == "" {
			errs = append(errs, fmt.Sprintf("inputs.%s: type is required", name))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: %s", path, strings.Join(errs, "; "))
	}
	return nil
}
