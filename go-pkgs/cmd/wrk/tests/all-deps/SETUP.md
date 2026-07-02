# Scenario

**Feature**: wrk --all-deps scans consumer go.mod + local git repos and links every matched dependency

```
# consumer requires dep1+dep2; scan-root has matching repos -> wrk --all-deps links each, replaces, tidies once
consumer (go.mod + git) + scan-root (mydep1, mydep2) -> wrk --all-deps --scan-root <root> -> stdout one line per dep + summary
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod`.
- Dep repos under the scan root must be `RepoTypeMain` on branch `main` with a committed `go.mod`.

## Steps

- Tests build an isolated consumer git repo plus named dep repos under a temp scan root.
- `req.RepoDir` is the consumer cwd for `wrk --all-deps`.
- `req.Args = []string{"--all-deps", "--scan-root", scanRoot}`.
- Dep repos use distinct module paths (`example.com/dep1`, `example.com/dep2`) so multiple can coexist under one scan root.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// allDepsGoModJSON mirrors the go.mod structure read via `go mod edit -json`.
type allDepsGoModJSON struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	return nil
}

// allDepsReadGoMod reads a directory's go.mod via `go mod edit -json`.
func allDepsReadGoMod(modDir string) (*allDepsGoModJSON, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var mod allDepsGoModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

// allDepsHasReplaceForModule reports whether go.mod has a replace for
// modulePath whose new path matches wantPath (empty wantPath → any path).
func allDepsHasReplaceForModule(mod *allDepsGoModJSON, modulePath, wantPath string) bool {
	for _, repl := range mod.Replace {
		if repl.Old.Path != modulePath {
			continue
		}
		if wantPath == "" || repl.New.Path == wantPath {
			return true
		}
	}
	return false
}

// allDepsReplacePathForModule returns the current replace new-path for
// modulePath, or "" if none.
func allDepsReplacePathForModule(mod *allDepsGoModJSON, modulePath string) string {
	for _, repl := range mod.Replace {
		if repl.Old.Path == modulePath {
			return repl.New.Path
		}
	}
	return ""
}

// allDepsGitignoreContainsExternal reports whether .gitignore has a `/external` line.
func allDepsGitignoreContainsExternal(top string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(top, ".gitignore"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			return true, nil
		}
	}
	return false, nil
}

// allDepsCountGitignoreExternalLines counts `/external` lines in .gitignore.
func allDepsCountGitignoreExternalLines(top string) (int, error) {
	data, err := os.ReadFile(filepath.Join(top, ".gitignore"))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			n++
		}
	}
	return n, nil
}

// allDepsExternalRelPath returns the relative `./external/<name>` form printed
// by wrk for a given dep basename (token "main", date wrkDate, no suffix).
func allDepsExternalRelPath(depBasename string) string {
	return fmt.Sprintf("./external/%s-main-%s", depBasename, wrkDate)
}

// allDepsExternalAbsPath returns the absolute external worktree path for a dep basename.
// consumerTop is resolved via EvalSymlinks because on macOS t.TempDir() returns an
// unresolved /var/folders/... path while wrk resolves to /private/var/folders/...;
// EvalSymlinks is a no-op on Linux where there is no symlink.
func allDepsExternalAbsPath(consumerTop, depBasename string) string {
	resolved, err := filepath.EvalSymlinks(consumerTop)
	if err != nil {
		resolved = consumerTop
	}
	return filepath.Join(resolved, "external", fmt.Sprintf("%s-main-%s", depBasename, wrkDate))
}

// allDepsDepMainRepo returns the resolved main-repo path of a dep repo created
// under {workRoot}/scan-root/<depBasename>. The external dep worktree is owned
// by the dep repo (registered under its .git/worktrees/), so ownership checks
// must run against this path, not the consumer. EvalSymlinks-resolved to match
// git's resolved paths on macOS (/var -> /private/var).
func allDepsDepMainRepo(workRoot, depBasename string) string {
	dep := filepath.Join(workRoot, "scan-root", depBasename)
	resolved, err := filepath.EvalSymlinks(dep)
	if err != nil {
		return dep
	}
	return resolved
}

// allDepsRunGo runs a go command in dir, failing the test on error.
func allDepsRunGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// allDepsRunGit runs a git command in dir with hooks disabled (-c core.hooksPath=)
// so local pre-commit hooks don't interfere with test repos.
func allDepsRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	fullArgs := append([]string{"-c", "core.hooksPath="}, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// initAllDepsRepo creates a git repo on branch main at path with a committed
// go.mod (module modulePath) and one .go file (package pkgName).
func initAllDepsRepo(t *testing.T, path, modulePath, pkgName string) {
	t.Helper()
	mkdirAll(t, path)
	allDepsRunGit(t, path, "init", "-b", "main")
	allDepsRunGit(t, path, "config", "user.email", "test@test.com")
	allDepsRunGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, pkgName+".go"), "package "+pkgName+"\n")
	allDepsRunGit(t, path, "add", "go.mod", pkgName+".go")
	allDepsRunGit(t, path, "commit", "-m", "init "+modulePath)
}

// initAllDepsConsumer creates a consumer git repo on main with a go.mod that
// requires the given module paths. extraGoMod (if non-empty) is appended to
// go.mod verbatim (used for pre-existing replace directives).
func initAllDepsConsumer(t *testing.T, workRoot string, requires []string, extraGoMod string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	mkdirAll(t, consumer)
	allDepsRunGit(t, consumer, "init", "-b", "main")
	allDepsRunGit(t, consumer, "config", "user.email", "test@test.com")
	allDepsRunGit(t, consumer, "config", "user.name", "Test")
	var sb strings.Builder
	sb.WriteString("module example.com/consumer\n\ngo 1.22\n")
	for _, m := range requires {
		sb.WriteString("\nrequire " + m + " v0.0.0\n")
	}
	if extraGoMod != "" {
		sb.WriteString("\n" + extraGoMod + "\n")
	}
	writeFile(t, filepath.Join(consumer, "go.mod"), sb.String())
	writeFile(t, filepath.Join(consumer, "main.go"), "package main\n")
	allDepsRunGit(t, consumer, "add", "go.mod", "main.go")
	allDepsRunGit(t, consumer, "commit", "-m", "init consumer")
	return consumer
}

// allDepsEnsureHelpersUsed keeps the prefixed helpers referenced even when a
// given leaf does not call every one (avoids unused-symbol compile errors in
// the inlined per-leaf test func).
func allDepsEnsureHelpersUsed() {
	_ = allDepsGoModJSON{}
	_ = allDepsReadGoMod
	_ = allDepsHasReplaceForModule
	_ = allDepsReplacePathForModule
	_ = allDepsGitignoreContainsExternal
	_ = allDepsCountGitignoreExternalLines
	_ = allDepsExternalRelPath
	_ = allDepsExternalAbsPath
	_ = allDepsDepMainRepo
	_ = allDepsRunGo
	_ = allDepsRunGit
	_ = initAllDepsRepo
	_ = initAllDepsConsumer
}
```
