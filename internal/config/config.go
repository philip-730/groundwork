// Package config loads the global groundwork config at ~/.groundwork/config.toml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the structure of ~/.groundwork/config.toml.
type Config struct {
	Registries []Registry `toml:"registries"`

	// Path is the absolute path of the file that was loaded.
	// It is set by Load and not present in the TOML file itself.
	Path string `toml:"-"`
}

// Registry is a named git remote that contains templates and a topologies.toml.
type Registry struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// GlobalConfigPath returns the path to the global config file.
func GlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".groundwork", "config.toml"), nil
}

// LoadGlobal loads the global config. If the file does not exist, an empty
// Config is returned so commands can show a helpful message rather than crash.
func LoadGlobal() (*Config, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Path: path}, nil
		}
		return nil, err
	}
	return cfg, nil
}

// Load reads and parses the config file at path, then validates it.
func Load(path string) (*Config, error) {
	var cfg Config
	md, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
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
