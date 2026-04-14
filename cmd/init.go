package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/philip-730/groundwork/internal/config"
	"github.com/philip-730/groundwork/internal/registry"
	"github.com/spf13/cobra"
)

var initName string

var initCmd = &cobra.Command{
	Use:   "init <url>",
	Short: "Add a registry and sync it",
	Long: `Adds a template registry to ~/.groundwork/config.toml and clones it locally.

If the registry is already configured with the same URL, it is re-synced.
Running init again with a different URL for the same name is an error.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		name := initName
		if name == "" {
			name = registryNameFromURL(url)
		}
		if name == "" {
			return fmt.Errorf("could not derive registry name from URL %q — use --name to set one", url)
		}

		// Check for conflicts with the already-loaded cfg.
		for _, r := range cfg.Registries {
			if r.Name == name {
				if r.URL != url {
					return fmt.Errorf("registry %q already configured with a different URL (%s)", name, r.URL)
				}
				// Same name and URL — just re-sync.
				fmt.Fprintf(os.Stdout, "registry %q already configured, re-syncing\n", name)
				return syncOne(config.Registry{Name: name, URL: url})
			}
		}

		// Append and write config.
		cfg.Registries = append(cfg.Registries, config.Registry{Name: name, URL: url})
		if err := writeConfig(cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		fmt.Fprintf(os.Stdout, "added registry %q\n", name)

		return syncOne(config.Registry{Name: name, URL: url})
	},
}

func syncOne(reg config.Registry) error {
	cacheRoot, err := registry.CacheRoot()
	if err != nil {
		return err
	}
	_, err = registry.Sync(reg, cacheRoot, os.Stdout)
	return err
}

func writeConfig(cfg *config.Config) error {
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(cfg.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// registryNameFromURL derives a registry name from the last path component of
// a URL. For example "https://github.com/my-org/groundwork-registry" → "groundwork-registry".
func registryNameFromURL(url string) string {
	url = strings.TrimRight(url, "/")
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		name := url[idx+1:]
		name = strings.TrimSuffix(name, ".git")
		return name
	}
	return ""
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "registry name (default: derived from URL)")
}
