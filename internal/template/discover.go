package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/philip-730/groundwork/internal/config"
	"github.com/philip-730/groundwork/internal/registry"
)

// Discover walks <registryCacheDir>/templates/ and returns every template
// whose template.toml parses and validates successfully.
// Templates that fail to parse are collected as non-fatal errors.
func Discover(registryName, registryCacheDir string) ([]*Template, []error) {
	templatesDir := filepath.Join(registryCacheDir, "templates")

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("read %s: %w", templatesDir, err)}
	}

	var templates []*Template
	var errs []error

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		tmplPath := filepath.Join(templatesDir, e.Name(), Filename)
		if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
			continue // directory exists but has no template.toml — skip silently
		}

		t, err := Load(tmplPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		t.RegistryName = registryName
		t.Dir = filepath.Join(templatesDir, e.Name())
		templates = append(templates, t)
	}

	return templates, errs
}

// DiscoverAll discovers templates across all registries that have been synced.
// Registries that have not been synced yet are skipped with a warning in errs.
func DiscoverAll(registries []config.Registry, cacheRoot string) ([]*Template, []error) {
	var all []*Template
	var allErrs []error

	for _, reg := range registries {
		if !registry.IsCached(cacheRoot, reg.Name) {
			allErrs = append(allErrs, fmt.Errorf(
				"registry %q not synced — run: groundwork registry sync", reg.Name))
			continue
		}
		cacheDir := registry.CacheDir(cacheRoot, reg.Name)
		templates, errs := Discover(reg.Name, cacheDir)
		all = append(all, templates...)
		allErrs = append(allErrs, errs...)
	}

	return all, allErrs
}

// Find looks up a template by name across a set of already-discovered templates.
// If name contains a slash it is treated as "registry/template". Returns an
// error when zero or more than one match is found.
func Find(templates []*Template, name string) (*Template, error) {
	registryFilter := ""
	templateName := name
	if idx := indexByte(name, '/'); idx >= 0 {
		registryFilter = name[:idx]
		templateName = name[idx+1:]
	}

	var matches []*Template
	for _, t := range templates {
		if t.Meta.Name != templateName {
			continue
		}
		if registryFilter != "" && t.RegistryName != registryFilter {
			continue
		}
		matches = append(matches, t)
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("template %q not found", name)
	case 1:
		return matches[0], nil
	default:
		var registries []string
		for _, m := range matches {
			registries = append(registries, m.RegistryName)
		}
		return nil, fmt.Errorf(
			"template %q exists in multiple registries (%v); qualify with <registry>/<template>",
			name, registries)
	}
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
