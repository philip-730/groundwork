package config

import (
	"errors"
	"fmt"
	"strings"
)

// Validate checks a parsed Config for semantic correctness.
func Validate(c *Config) error {
	var errs []string

	for i, r := range c.Registries {
		if strings.TrimSpace(r.Name) == "" {
			errs = append(errs, fmt.Sprintf("registries[%d]: name is required", i))
		}
		if strings.TrimSpace(r.URL) == "" {
			errs = append(errs, fmt.Sprintf("registries[%d]: url is required", i))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
