# gotool — replace, resolve, update, pin

## Version

0.0.2

Library doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/gotool` (migrated from
`kool-wrk/tools/go`). Each leaf calls gotool APIs directly via root `Run()` — no CLI.

**Classic TDD (Pin):** `update.Pin` is scaffolded but not implemented. Leaves under
`update/pin/` expect **RED** until the implementer lands real Pin semantics.
Existing `update/basic` still exercises `Update` and should stay GREEN until the
implementer rewires `Update` as a thin wrapper around `Pin`.

## DSN (Domain Specific Notion)

- **`replace.Replace(dir)`** — equivalent to `go mod edit -replace "$mod=$absDir"` from the
  consumer module cwd; `dir` must be an existing Go module directory.
- **`resolve.ResolveLocalModules(currentDir, []string{localDir})`** — reads consumer
  `go.mod`, resolves local module path, returns `LocalModuleInfo.IsDependency`.
- **`update.Update(dir)`** — drops replace for the target module and sets require to the
  latest matching git version tag (consumer = process cwd; legacy).
- **`update.UpdateIn(consumerDir, dir)`** — same as Update with an explicit consumer module.
- **`update.Pin(opts)`** — library pin with explicit `ConsumerDir`, `DepDir`, optional
  `Version`, and `DryRun`. No process-global Chdir; edits use
  `commands.GoModEditOptions{Dir: ConsumerDir}`.
- **`ConsumerDir`** — module directory whose `go.mod` is edited (replace/update/pin) or read
  (resolve).
- **`TargetDir` / `DepDir`** — local Go module directory passed to `Replace` / `Update` /
  `Pin` as the dependency source (module path + tags).
- **`LocalModDir`** — dependency directory checked by resolve leaves.
- **`Version`** — optional exact go require version for Pin (e.g. `v0.0.5`); empty =
  resolve latest tag.
- **`DryRun`** — Pin plans result without writing consumer `go.mod`.
- **`DiskVersion`** — require version observed on disk after the call (may differ from
  planned `ModuleVersion` on dry-run).

## Decision Tree

```
gotool tests
├── replace/
│   └── basic/                    # Replace adds correct -replace directive
├── resolve/
│   ├── is-dependency/            # module in require → IsDependency true
│   └── not-dependency/           # unknown module → IsDependency false
└── update/
    ├── basic/                    # Update drops replace, sets require from git tag (GREEN)
    └── pin/                      # Pin API — classic TDD RED until implementer
        ├── basic/                # Pin: drop replace, require latest tag
        ├── explicit-version/     # Pin with Version=v0.0.5 (forced; no tag lookup required)
        ├── dry-run/              # would pin; go.mod unchanged; PinResult filled
        └── missing-tag/          # untagged DepDir + empty Version → error
```

## Test Index

| # | Leaf | Description | Expected |
|---|------|-------------|----------|
| 1 | `replace/basic` | `Replace` adds `replace old => absDir` in consumer go.mod | GREEN |
| 2 | `resolve/is-dependency` | Module listed in require → `IsDependency` true | GREEN |
| 3 | `resolve/not-dependency` | Module not in consumer go.mod → `IsDependency` false | GREEN |
| 4 | `update/basic` | `Update` drops replace and sets require to latest tag | GREEN |
| 5 | `update/pin/basic` | `Pin` drops replace, require latest tag; PinResult filled | RED |
| 6 | `update/pin/explicit-version` | `Pin` with `Version=v0.0.5` pins that version | RED |
| 7 | `update/pin/dry-run` | DryRun plans pin; consumer go.mod unchanged on disk | RED |
| 8 | `update/pin/missing-tag` | Untagged DepDir + empty Version → error mentioning tag | RED |

## How to Run

```sh
doctest vet ./go-pkgs/gotool/tests
doctest test ./go-pkgs/gotool/tests
doctest test ./go-pkgs/gotool/tests/update
doctest test ./go-pkgs/gotool/tests/update/pin
```

```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
)

type Request struct {
	Operation   string // "replace", "resolve", "update", "pin"
	ConsumerDir string
	TargetDir   string // dep module dir for replace/update/pin
	LocalModDir string
	Version     string // Pin: optional exact require version
	DryRun      bool   // Pin: plan only, no go.mod writes
}

type Response struct {
	Err           error
	IsDependency  bool
	ModulePath    string
	AbsDir        string
	ModuleVersion string // planned/applied require version (PinResult.Version or update disk)
	Tag           string // PinResult.Tag
	HasReplace    bool   // replace still present on disk after call
	DiskVersion   string // require version on disk after call
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Operation {
	case "replace":
		resp.AbsDir, resp.ModulePath, resp.Err = replace.ReplaceIn(req.ConsumerDir, req.TargetDir)
		if resp.Err == nil {
			modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
			if err != nil {
				return nil, err
			}
			resp.HasReplace = hasReplaceFor(modInfo, resp.ModulePath, resp.AbsDir)
		}

	case "resolve":
		_, resolved, resolveErr := resolve.ResolveLocalModules(req.ConsumerDir, []string{req.LocalModDir})
		resp.Err = resolveErr
		if resp.Err == nil {
			if len(resolved) != 1 {
				return nil, fmt.Errorf("expected 1 resolved module, got %d", len(resolved))
			}
			resp.IsDependency = resolved[0].IsDependency
			resp.ModulePath = resolved[0].ModuleInfo.Module.Path
		}

	case "update":
		resp.Err = update.UpdateIn(req.ConsumerDir, req.TargetDir)
		if resp.Err == nil {
			modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
			if err != nil {
				return nil, err
			}
			modulePath, err := modulePathFromTarget(req.TargetDir)
			if err != nil {
				return nil, err
			}
			resp.ModulePath = modulePath
			resp.ModuleVersion = requireVersion(modInfo, resp.ModulePath)
			resp.DiskVersion = resp.ModuleVersion
			resp.HasReplace = hasReplaceForPath(modInfo, resp.ModulePath)
		}

	case "pin":
		result, pinErr := update.Pin(update.PinOptions{
			ConsumerDir: req.ConsumerDir,
			DepDir:      req.TargetDir,
			Version:     req.Version,
			DryRun:      req.DryRun,
		})
		resp.Err = pinErr
		resp.ModulePath = result.ModulePath
		resp.ModuleVersion = result.Version
		resp.Tag = result.Tag

		// Always inspect consumer go.mod when present (success, dry-run, and error paths).
		if req.ConsumerDir != "" {
			modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
			if err != nil {
				// Surface read errors only when Pin itself succeeded; otherwise keep pinErr.
				if pinErr == nil {
					return nil, err
				}
			} else {
				path := resp.ModulePath
				if path == "" {
					path, _ = modulePathFromTarget(req.TargetDir)
					if resp.ModulePath == "" {
						resp.ModulePath = path
					}
				}
				if path != "" {
					resp.DiskVersion = requireVersion(modInfo, path)
					resp.HasReplace = hasReplaceForPath(modInfo, path)
				}
			}
		}

	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	return resp, nil
}

func hasReplaceFor(modInfo *resolve.ModuleInfo, modulePath, absDir string) bool {
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == modulePath && repl.New.Path == absDir {
			return true
		}
	}
	return false
}

func hasReplaceForPath(modInfo *resolve.ModuleInfo, modulePath string) bool {
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == modulePath {
			return true
		}
	}
	return false
}

func requireVersion(modInfo *resolve.ModuleInfo, modulePath string) string {
	for _, req := range modInfo.Require {
		if req.Path == modulePath {
			return req.Version
		}
	}
	return ""
}

func modulePathFromTarget(targetDir string) (string, error) {
	opts := &commands.GoModEditOptions{Dir: targetDir}
	mod, err := commands.GoModEditJSON(opts)
	if err != nil {
		return "", err
	}
	return mod.Module.Path, nil
}
```
