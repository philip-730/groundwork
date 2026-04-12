package config

import (
	"errors"
	"fmt"
	"strings"
)

// Validate checks a parsed Config for semantic correctness.
func Validate(c *Config) error {
	var errs []string

	if strings.TrimSpace(c.Workspace.Name) == "" {
		errs = append(errs, "workspace.name is required")
	}

	for i, r := range c.Registries {
		if strings.TrimSpace(r.Name) == "" {
			errs = append(errs, fmt.Sprintf("registries[%d]: name is required", i))
		}
		if strings.TrimSpace(r.URL) == "" {
			errs = append(errs, fmt.Sprintf("registries[%d]: url is required", i))
		}
	}

	defaults := 0
	for name, t := range c.Topologies {
		if strings.TrimSpace(name) == "" {
			errs = append(errs, "topology name must not be empty")
		}
		if t.Default {
			defaults++
		}
	}
	if len(c.Topologies) > 1 && defaults == 0 {
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
