# gotool — replace, resolve, update

## Version

0.0.1

Library doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/gotool` (migrated from
`kool-wrk/tools/go`). Each leaf calls gotool APIs directly via root `Run()` — no CLI.

## DSN (Domain Specific Notion)

- **`replace.Replace(dir)`** — equivalent to `go mod edit -replace "$mod=$absDir"` from the
  consumer module cwd; `dir` must be an existing Go module directory.
- **`resolve.ResolveLocalModules(currentDir, []string{localDir})`** — reads consumer
  `go.mod`, resolves local module path, returns `LocalModuleInfo.IsDependency`.
- **`update.Update(dir)`** — drops replace for the target module and sets require to the
  latest matching git version tag.
- **`ConsumerDir`** — module directory whose `go.mod` is edited (replace/update) or read
  (resolve).
- **`TargetDir`** — local Go module directory passed to `Replace` / `Update`.
- **`LocalModDir`** — dependency directory checked by resolve leaves.

## Decision Tree

```
gotool tests
├── replace/
│   └── basic/                    # Replace adds correct -replace directive
├── resolve/
│   ├── is-dependency/            # module in require → IsDependency true
│   └── not-dependency/           # unknown module → IsDependency false
└── update/
    └── basic/                    # Update drops replace, sets require from git tag
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `replace/basic` | `Replace` adds `replace old => absDir` in consumer go.mod |
| 2 | `resolve/is-dependency` | Module listed in require → `IsDependency` true |
| 3 | `resolve/not-dependency` | Module not in consumer go.mod → `IsDependency` false |
| 4 | `update/basic` | `Update` drops replace and sets require to latest tag |

## How to Run

```sh
doctest vet ./go-pkgs/gotool/tests
doctest test ./go-pkgs/gotool/tests
```

```go
import (
	"fmt"
	"os"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
)

type Request struct {
	Operation   string // "replace", "resolve", "update"
	ConsumerDir string
	TargetDir   string
	LocalModDir string
}

type Response struct {
	Err           error
	IsDependency  bool
	ModulePath    string
	AbsDir        string
	ModuleVersion string
	HasReplace    bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	oldWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(req.ConsumerDir); err != nil {
		return nil, err
	}
	defer os.Chdir(oldWd)

	switch req.Operation {
	case "replace":
		resp.AbsDir, resp.ModulePath, resp.Err = replace.Replace(req.TargetDir)
		if resp.Err == nil {
			modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
			if err != nil {
				return nil, err
			}
			resp.HasReplace = hasReplaceFor(modInfo, resp.ModulePath, resp.AbsDir)
		}

	case "resolve":
		_, resolved, resp.Err := resolve.ResolveLocalModules(req.ConsumerDir, []string{req.LocalModDir})
		if resp.Err == nil {
			if len(resolved) != 1 {
				return nil, fmt.Errorf("expected 1 resolved module, got %d", len(resolved))
			}
			resp.IsDependency = resolved[0].IsDependency
			resp.ModulePath = resolved[0].ModuleInfo.Module.Path
		}

	case "update":
		resp.Err = update.Update(req.TargetDir)
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
			resp.HasReplace = hasReplaceForPath(modInfo, resp.ModulePath)
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