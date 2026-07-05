# Scenario

**Feature**: npm package manager detection, resolve, and install argv helpers

```
# detection from project root
projectDir fixtures -> DetectFromDir -> Trace.Manager + Signals

# detection from node_modules path
node_modules path -> DetectFromNodeModules -> Trace + HasPackageJSON

# resolve with explicit or auto pref
projectDir + pref -> Resolve -> Manager on PATH or error

# install argv per manager
Manager + FrozenLockfile -> InstallArgs -> CLI argv slice
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/npm` is importable.
- Filesystem fixtures are created in temp dirs per leaf `Setup`.
- `resolve` leaves that require a specific CLI skip when that binary is absent
  from PATH.

## Context

- Detection priority when multiple indicators: `pnpm > bun > npm > yarn`.
- `writeProject` creates a temp project tree from a `map[relativePath]content`.
- `mkdirProjectDir` creates nested directories (e.g. `node_modules/.pnpm/...`).

```go
import (
	"os"
	"path/filepath"
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

const (
	pkgJSONDemo           = `{"name":"demo"}`
	pkgJSONNpmPM          = `{"name":"demo","packageManager":"npm@10.0.0"}`
	pkgJSONPnpmPM         = `{"name":"demo","packageManager":"pnpm@11.10.0"}`
	pnpmLockYAML          = "lockfileVersion: '9.0'\n"
	bunLockJSON           = "{}"
	bunLockbStub          = "\x00bun.lockb"
	packageLockJSON       = `{"lockfileVersion":3}`
	yarnLockStub          = "# yarn lockfile v1\n\n"
)

func writeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mkdirProjectDir(t *testing.T, projectDir, rel string) string {
	t.Helper()
	abs := filepath.Join(projectDir, rel)
	if err := os.MkdirAll(abs, 0755); err != nil {
		t.Fatal(err)
	}
	return abs
}

func nodeModulesPath(t *testing.T, projectDir string) string {
	t.Helper()
	return mkdirProjectDir(t, projectDir, "node_modules")
}

func requireManagerOnPath(t *testing.T, manager npm.Manager) {
	t.Helper()
	if !npm.Available(manager) {
		t.Skipf("%s not available on PATH", manager)
	}
}
```