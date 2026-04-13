package registry_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philip-730/groundwork/internal/config"
	"github.com/philip-730/groundwork/internal/registry"
)

// makeBarRepo creates a local bare git repo with one commit and returns its path.
// This acts as a stand-in for a remote, with no network involvement.
func makeBarRepo(t *testing.T) string {
	t.Helper()

	// Working repo: init, add a file, commit.
	src := t.TempDir()
	mustGit(t, src, "init")
	mustGit(t, src, "config", "user.email", "test@example.com")
	mustGit(t, src, "config", "user.name", "Test")

	readmePath := filepath.Join(src, "README.md")
	if err := os.WriteFile(readmePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, src, "add", ".")
	mustGit(t, src, "commit", "-m", "initial")

	// Clone it as a bare repo — this is the "remote".
	bare := t.TempDir()
	mustGitCmd(t, "git", "clone", "--bare", src, bare)
	return bare
}

// addCommit adds a new file and commits it to the repo at dir.
func addCommit(t *testing.T, dir, filename string) {
	t.Helper()
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, filename), []byte("data\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "add "+filename)
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	// -c commit.gpgsign=false prevents interference from environment-level
	// signing hooks that are not relevant to unit tests.
	full := append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func mustGitCmd(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestSync_Clone(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := t.TempDir()

	reg := config.Registry{Name: "test", URL: bare}
	var buf bytes.Buffer

	res, err := registry.Sync(reg, cacheRoot, &buf)
	if err != nil {
		t.Fatalf("Sync: %v\noutput: %s", err, buf.String())
	}
	if res.Action != "cloned" {
		t.Errorf("action = %q, want %q", res.Action, "cloned")
	}
	if res.Name != "test" {
		t.Errorf("name = %q, want %q", res.Name, "test")
	}

	// Verify the clone actually landed.
	if !registry.IsCached(cacheRoot, "test") {
		t.Error("IsCached returned false after successful clone")
	}
	if _, err := os.Stat(filepath.Join(res.Dir, "README.md")); err != nil {
		t.Errorf("README.md not found in clone: %v", err)
	}
}

func TestSync_Pull(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := t.TempDir()

	reg := config.Registry{Name: "test", URL: bare}
	var buf bytes.Buffer

	// First sync: clone.
	if _, err := registry.Sync(reg, cacheRoot, &buf); err != nil {
		t.Fatalf("initial Sync: %v", err)
	}

	// Push a new commit to the bare repo by cloning it, committing, pushing.
	work := t.TempDir()
	mustGitCmd(t, "git", "clone", bare, work)
	addCommit(t, work, "new-file.txt")
	mustGit(t, work, "push", "origin", "HEAD")

	// Second sync: pull.
	buf.Reset()
	res, err := registry.Sync(reg, cacheRoot, &buf)
	if err != nil {
		t.Fatalf("second Sync: %v\noutput: %s", err, buf.String())
	}
	if res.Action != "updated" {
		t.Errorf("action = %q, want %q", res.Action, "updated")
	}

	// Verify the new file appeared.
	cloneDir := registry.CacheDir(cacheRoot, "test")
	if _, err := os.Stat(filepath.Join(cloneDir, "new-file.txt")); err != nil {
		t.Errorf("new-file.txt not found after pull: %v", err)
	}
}

func TestSyncAll_CreatesCache(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := filepath.Join(t.TempDir(), "deep", "cache") // doesn't exist yet

	regs := []config.Registry{{Name: "r1", URL: bare}}
	var buf bytes.Buffer
	results, err := registry.SyncAll(regs, cacheRoot, &buf)
	if err != nil {
		t.Fatalf("SyncAll: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
}

func TestSyncAll_ContinuesOnError(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := t.TempDir()

	regs := []config.Registry{
		{Name: "bad", URL: "/nonexistent/path/that/does/not/exist"},
		{Name: "good", URL: bare},
	}
	var buf bytes.Buffer
	results, err := registry.SyncAll(regs, cacheRoot, &buf)

	// Should have an error for "bad" but still return "good".
	if err == nil {
		t.Fatal("expected error for bad registry, got nil")
	}
	if len(results) != 1 || results[0].Name != "good" {
		t.Errorf("results = %v, want [{good ...}]", results)
	}
}

func TestIsCached(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := t.TempDir()

	reg := config.Registry{Name: "myrepo", URL: bare}

	if registry.IsCached(cacheRoot, "myrepo") {
		t.Error("IsCached = true before any sync")
	}

	var buf bytes.Buffer
	if _, err := registry.Sync(reg, cacheRoot, &buf); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if !registry.IsCached(cacheRoot, "myrepo") {
		t.Error("IsCached = false after sync")
	}
}

func TestCacheDir(t *testing.T) {
	got := registry.CacheDir("/home/user/.groundwork/registries", "internal")
	want := "/home/user/.groundwork/registries/internal"
	if got != want {
		t.Errorf("CacheDir = %q, want %q", got, want)
	}
}

func TestSyncAll_Output(t *testing.T) {
	bare := makeBarRepo(t)
	cacheRoot := t.TempDir()

	regs := []config.Registry{{Name: "myrepo", URL: bare}}
	var buf bytes.Buffer
	registry.SyncAll(regs, cacheRoot, &buf) //nolint:errcheck

	out := buf.String()
	if !strings.Contains(out, "myrepo") {
		t.Errorf("output missing registry name: %q", out)
	}
	if !strings.Contains(out, "cloned") {
		t.Errorf("output missing action: %q", out)
	}
}
