package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// CacheRoot returns the root directory where all registry clones are stored:
// ~/.groundwork/registries
func CacheRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".groundwork", "registries"), nil
}

// CacheDir returns the local clone path for a named registry.
func CacheDir(cacheRoot, name string) string {
	return filepath.Join(cacheRoot, name)
}

// IsCached reports whether a registry has been cloned locally.
func IsCached(cacheRoot, name string) bool {
	dir := CacheDir(cacheRoot, name)
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
