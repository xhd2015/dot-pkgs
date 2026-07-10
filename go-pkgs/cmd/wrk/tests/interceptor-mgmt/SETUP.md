# Scenario

**Feature**: wrk --interceptor management CLI for create.interceptor config

```
# management mode (does not exec create interceptor)
wrk --interceptor --status|--show|--path|--enable|--disable|--init|--check|--dry-run
  -> read/write $WRK_HOME/config.json create.interceptor
  -> stdout action-specific; mutually exclusive with other wrk modes

# lifecycle
--init (disabled stub) -> --enable / --disable -> file mutation, empty stdout preferred
--path / --status / --show / --check / --dry-run -> inspect or expand without create exec
```

## Preconditions

- Isolated `{WRK_HOME}` per leaf (root `Setup`).
- Management does not require a git checkout; leaves use neutral cwd under `{WorkRoot}`.
- Config path is always `{WRK_HOME}/config.json` (may be missing for read actions).

## Steps

- Descendants set `Request.Args` to `--interceptor` plus exactly one action (and optional `--force` / dry-run create args).
- Seed or omit `config.json` per scenario; assert exit code, stdout templates, and file side effects.

## Context

- Neutral init stub: `enabled: false`, `argv: ["echo", "wrk-interceptor-not-configured"]`.
- Status stdout is three stable lines: `state:`, `path:`, `argv0:`.
- Dry-run prints expanded argv one element per line; requires present + enabled interceptor.
- Reuses root `Request` / `Run` / `ExtraEnv` harness; no nested `DOCTEST.md`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	mgmtStubArgv0 = "echo"
	mgmtStubArgv1 = "wrk-interceptor-not-configured"
)

type mgmtConfigFile struct {
	Version int                    `json:"version"`
	Create  *mgmtCreateSection     `json:"create,omitempty"`
	Extra   map[string]interface{} `json:"-"` // not used; raw merge helpers below
}

type mgmtCreateSection struct {
	Interceptor *mgmtInterceptorSpec `json:"interceptor,omitempty"`
}

type mgmtInterceptorSpec struct {
	Enabled bool                   `json:"enabled"`
	Argv    []string               `json:"argv"`
	Vars    map[string]interface{} `json:"vars,omitempty"`
}

func Setup(t *testing.T, req *Request) error {
	// Neutral cwd: management must work outside a git repo.
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	ensureInterceptorMgmtHelpersUsed()
	return nil
}

func mgmtConfigPath(wrkHome string) string {
	return filepath.Join(wrkHome, "config.json")
}

func interceptorMgmtArgs(action string, extra ...string) []string {
	args := []string{"--interceptor", action}
	return append(args, extra...)
}

// writeMgmtInterceptorConfig writes a minimal versioned config with create.interceptor.
func writeMgmtInterceptorConfig(t *testing.T, wrkHome string, enabled bool, argv []string, vars map[string]interface{}) {
	t.Helper()
	cfg := mgmtConfigFile{Version: 1}
	cfg.Create = &mgmtCreateSection{
		Interceptor: &mgmtInterceptorSpec{
			Enabled: enabled,
			Argv:    argv,
			Vars:    vars,
		},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	data = append(data, '\n')
	writeFile(t, mgmtConfigPath(wrkHome), string(data))
}

// writeMgmtRawConfig writes arbitrary JSON text to config.json (for merge / extra keys).
func writeMgmtRawConfig(t *testing.T, wrkHome, content string) {
	t.Helper()
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	writeFile(t, mgmtConfigPath(wrkHome), content)
}

func readMgmtConfigBytes(t *testing.T, wrkHome string) []byte {
	t.Helper()
	data, err := os.ReadFile(mgmtConfigPath(wrkHome))
	if err != nil {
		t.Fatalf("read config.json: %v", err)
	}
	return data
}

func readMgmtInterceptor(t *testing.T, wrkHome string) *mgmtInterceptorSpec {
	t.Helper()
	data, err := os.ReadFile(mgmtConfigPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read config.json: %v", err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse config.json: %v", err)
	}
	createRaw, ok := root["create"]
	if !ok {
		return nil
	}
	var create map[string]json.RawMessage
	if err := json.Unmarshal(createRaw, &create); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	icRaw, ok := create["interceptor"]
	if !ok {
		return nil
	}
	var ic mgmtInterceptorSpec
	if err := json.Unmarshal(icRaw, &ic); err != nil {
		t.Fatalf("parse interceptor: %v", err)
	}
	return &ic
}

func assertMgmtConfigAbsent(t *testing.T, wrkHome string) {
	t.Helper()
	_, err := os.Stat(mgmtConfigPath(wrkHome))
	if err == nil {
		t.Fatalf("config.json should not exist at %s", mgmtConfigPath(wrkHome))
	}
	if !os.IsNotExist(err) {
		t.Fatalf("stat config.json: %v", err)
	}
}

func assertMgmtInterceptorEnabled(t *testing.T, wrkHome string, want bool) {
	t.Helper()
	ic := readMgmtInterceptor(t, wrkHome)
	if ic == nil {
		t.Fatal("expected create.interceptor block")
	}
	if ic.Enabled != want {
		t.Fatalf("enabled: want %v, got %v", want, ic.Enabled)
	}
}

func assertMgmtNeutralStub(t *testing.T, wrkHome string) {
	t.Helper()
	ic := readMgmtInterceptor(t, wrkHome)
	if ic == nil {
		t.Fatal("expected create.interceptor stub")
	}
	if ic.Enabled {
		t.Fatal("stub must have enabled: false")
	}
	if len(ic.Argv) < 1 {
		t.Fatalf("stub argv empty: %+v", ic.Argv)
	}
	if ic.Argv[0] != mgmtStubArgv0 {
		t.Fatalf("stub argv[0]: want %q, got %q", mgmtStubArgv0, ic.Argv[0])
	}
	if len(ic.Argv) >= 2 && ic.Argv[1] != mgmtStubArgv1 {
		t.Fatalf("stub argv[1]: want %q, got %q", mgmtStubArgv1, ic.Argv[1])
	}
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func statusStdoutTemplate(state, path, argv0 string) string {
	body := fmt.Sprintf("state: %s\npath: %s\nargv0: %s", state, path, argv0)
	return v2StdoutTemplate(body)
}

func ensureInterceptorMgmtHelpersUsed() {
	_ = mgmtConfigPath
	_ = interceptorMgmtArgs
	_ = writeMgmtInterceptorConfig
	_ = writeMgmtRawConfig
	_ = readMgmtConfigBytes
	_ = readMgmtInterceptor
	_ = assertMgmtConfigAbsent
	_ = assertMgmtInterceptorEnabled
	_ = assertMgmtNeutralStub
	_ = assertEmptyStdout
	_ = statusStdoutTemplate
}
```
