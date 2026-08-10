# Scenario

**Feature**: gotool library helpers for go.mod replace, resolve, update, and pin

```
# isolated temp workspace per leaf; go + git on PATH
consumer go.mod + local module dir -> gotool API -> go.mod side effects / PinResult
```

## Preconditions

- The `gotool` package tree is importable under `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/...`
- `go` and `git` are available on PATH

## Steps

1. Verify `go` and `git` are available.
2. Leaf `Setup` builds consumer + target module fixtures.
3. Root `Run` invokes the requested gotool API (`pin` without Chdir; `replace`/`update` chdir to consumer).

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	consumerModulePath = "example.com/consumer"
	depModulePath      = "example.com/dep"
	fixtureModulePath  = "github.com/example/fixture-mod"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	return nil
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func mustGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initDepModuleRepo(t *testing.T, workspace, modulePath string) string {
	t.Helper()
	repo := filepath.Join(workspace, "dep-repo")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repo, "dep.go"), "package dep\n")
	mustGit(t, repo, "init", "-b", "main")
	mustGit(t, repo, "config", "user.email", "test@example.com")
	mustGit(t, repo, "config", "user.name", "Test User")
	mustGit(t, repo, "add", ".")
	mustGit(t, repo, "commit", "-m", "init dep")
	return repo
}

func initConsumerModule(t *testing.T, workspace string, withRequire bool) string {
	t.Helper()
	consumer := filepath.Join(workspace, "consumer")
	if err := os.MkdirAll(consumer, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(consumer, "go.mod"), "module "+consumerModulePath+"\n\ngo 1.22\n")
	if withRequire {
		mustGo(t, consumer, "mod", "edit", "-require="+depModulePath+"@v0.0.0")
	}
	return consumer
}

func initTaggedFixtureRepo(t *testing.T, workspace string) string {
	t.Helper()
	repo := filepath.Join(workspace, "fixture-mod")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "go.mod"), "module "+fixtureModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repo, "mod.go"), "package fixturemod\n")
	mustGit(t, repo, "init", "-b", "main")
	mustGit(t, repo, "config", "user.email", "test@example.com")
	mustGit(t, repo, "config", "user.name", "Test User")
	mustGit(t, repo, "add", ".")
	mustGit(t, repo, "commit", "-m", "initial")
	mustGit(t, repo, "tag", "v1.0.0")
	writeFile(t, filepath.Join(repo, "README.md"), "post-tag change\n")
	mustGit(t, repo, "add", "README.md")
	mustGit(t, repo, "commit", "-m", "post-tag commit")
	return repo
}

// initUntaggedFixtureRepo is like initTaggedFixtureRepo but never creates a version tag.
func initUntaggedFixtureRepo(t *testing.T, workspace string) string {
	t.Helper()
	repo := filepath.Join(workspace, "fixture-mod")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "go.mod"), "module "+fixtureModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repo, "mod.go"), "package fixturemod\n")
	mustGit(t, repo, "init", "-b", "main")
	mustGit(t, repo, "config", "user.email", "test@example.com")
	mustGit(t, repo, "config", "user.name", "Test User")
	mustGit(t, repo, "add", ".")
	mustGit(t, repo, "commit", "-m", "initial untagged")
	return repo
}

func initConsumerWithReplace(t *testing.T, workspace, targetDir string) string {
	t.Helper()
	consumer := filepath.Join(workspace, "consumer")
	if err := os.MkdirAll(consumer, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(consumer, "go.mod"), "module "+consumerModulePath+"\n\ngo 1.22\n")
	mustGo(t, consumer, "mod", "edit", "-require="+fixtureModulePath+"@v0.0.1")
	mustGo(t, consumer, "mod", "edit", "-replace="+fixtureModulePath+"="+targetDir)
	return consumer
}

func newWorkspace(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gotool-doctest-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}
```
