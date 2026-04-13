package registry

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/philip-730/groundwork/internal/config"
)

// SyncResult describes the outcome of a single registry sync.
type SyncResult struct {
	Name   string
	Action string // "cloned" or "updated"
	Dir    string
}

// Sync clones or pulls a single registry into cacheRoot/<reg.Name>.
// Progress output from git is written to w. Returns a SyncResult on success.
func Sync(reg config.Registry, cacheRoot string, w io.Writer) (SyncResult, error) {
	dest := CacheDir(cacheRoot, reg.Name)

	if IsCached(cacheRoot, reg.Name) {
		return pull(reg, dest, w)
	}
	return clone(reg, dest, w)
}

// SyncAll syncs every registry in the slice, returning all results.
// It continues on error, collecting all failures before returning.
func SyncAll(registries []config.Registry, cacheRoot string, w io.Writer) ([]SyncResult, error) {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create cache root %s: %w", cacheRoot, err)
	}

	var results []SyncResult
	var errs []error

	for _, reg := range registries {
		fmt.Fprintf(w, "syncing %s (%s)\n", reg.Name, reg.URL)
		res, err := Sync(reg, cacheRoot, w)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", reg.Name, err))
			fmt.Fprintf(w, "  error: %v\n", err)
			continue
		}
		fmt.Fprintf(w, "  %s → %s\n", res.Action, res.Dir)
		results = append(results, res)
	}

	if len(errs) > 0 {
		return results, joinErrors(errs)
	}
	return results, nil
}

func clone(reg config.Registry, dest string, w io.Writer) (SyncResult, error) {
	cmd := exec.Command("git", "clone", "--", reg.URL, dest)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return SyncResult{}, fmt.Errorf("git clone: %w", err)
	}
	return SyncResult{Name: reg.Name, Action: "cloned", Dir: dest}, nil
}

func pull(reg config.Registry, dest string, w io.Writer) (SyncResult, error) {
	// Fetch then reset to origin/HEAD so the local clone is always clean.
	fetch := exec.Command("git", "-C", dest, "fetch", "--prune", "origin")
	fetch.Stdout = w
	fetch.Stderr = w
	if err := fetch.Run(); err != nil {
		return SyncResult{}, fmt.Errorf("git fetch: %w", err)
	}

	reset := exec.Command("git", "-C", dest, "reset", "--hard", "origin/HEAD")
	reset.Stdout = w
	reset.Stderr = w
	if err := reset.Run(); err != nil {
		return SyncResult{}, fmt.Errorf("git reset: %w", err)
	}

	return SyncResult{Name: reg.Name, Action: "updated", Dir: dest}, nil
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	msg := errs[0].Error()
	for _, e := range errs[1:] {
		msg += "; " + e.Error()
	}
	return fmt.Errorf("%s", msg)
}
