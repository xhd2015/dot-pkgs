# Scenario

**Feature**: withgo pins a Go version, resolves a GOROOT, and execs under it

```
# pin then resolve dest under InstallDir; missing dest uses Install hook
caller goVersion -> PinPatch -> dest=$InstallDir/<pin> -> existing dir | Install hook

# ModuleGoLine reads go.mod; Exec/Run launch a child with GOROOT+PATH
modDir -> ModuleGoLine -> go1.19
goroot + args -> Exec -> child env GOROOT=$abs PATH=$abs/bin:$PATH
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo` is the SUT (not
  implemented yet — leaves stay compile-RED until it lands).
- Parallel-safe: no `os.Chdir` / `t.Chdir` / `os.Setenv` / `t.Setenv` / process
  stdio rewrite. InstallDir is always `t.TempDir()`. Resolve/run leaves inject a
  recording `Install` hook — never real `downloadgo` network.
- Process cwd is undetermined. Use absolute paths from `t.TempDir()` / `d`.
- `DefaultInstallDir` is read-only (`os.UserHomeDir()`); do not mutate HOME.

## Steps

1. Leaf/grouping `Setup` sets `req.Op` and scenario fixtures (temp install dir,
   go.mod, fake `$GOROOT/bin/go`).
2. Root `Run` dispatches to `withgo.PinPatch`, `ModuleGoLine`, `ResolveGoroot`,
   `Exec`, `Run`, or `DefaultInstallDir`.

## Context

- Pin table matches kool `ResolveGoroot` (go1.14…go1.25).
- Dest path is `$InstallDir/go1.19.13` for input `go1.19`.
- Fake `go` is a Unix shell script that prints `GOROOT`, first `PATH` entry, and
  optional `WITHGO_EXTRA`. This environment is macOS.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeGoMod(t *testing.T, dir, body string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "go.mod"), body)
}

func writeFakeGo(t *testing.T, goroot string) {
	t.Helper()
	script := `#!/bin/sh
printf 'GOROOT=%s\n' "$GOROOT"
IFS=:
set -- $PATH
printf 'PATH0=%s\n' "$1"
if [ -n "$WITHGO_EXTRA" ]; then
  printf 'WITHGO_EXTRA=%s\n' "$WITHGO_EXTRA"
fi
`
	writeFile(t, filepath.Join(goroot, "bin", "go"), script)
	if err := os.Chmod(filepath.Join(goroot, "bin", "go"), 0755); err != nil {
		t.Fatal(err)
	}
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func lastEnv(envOut, key string) (string, bool) {
	prefix := key + "="
	val := ""
	found := false
	for _, line := range strings.Split(envOut, "\n") {
		if strings.HasPrefix(line, prefix) {
			val = strings.TrimPrefix(line, prefix)
			found = true
		}
	}
	return val, found
}

func koolPinCases() [][2]string {
	return [][2]string{
		{"go1.25", "go1.25.0"},
		{"go1.24", "go1.24.1"},
		{"go1.23", "go1.23.6"},
		{"go1.22", "go1.22.12"},
		{"go1.21", "go1.21.13"},
		{"go1.20", "go1.20.14"},
		{"go1.19", "go1.19.13"},
		{"go1.18", "go1.18.10"},
		{"go1.17", "go1.17.13"},
		{"go1.16", "go1.16.15"},
		{"go1.15", "go1.15.15"},
		{"go1.14", "go1.14.15"},
		{"1.19", "go1.19.13"},
	}
}
```
