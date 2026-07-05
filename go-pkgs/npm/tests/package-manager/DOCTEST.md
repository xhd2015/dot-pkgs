# Package Manager — detection, resolve, install helpers

## Version
0.0.2

Doc-style tests for JavaScript package manager detection and install command
helpers in `github.com/xhd2015/dot-pkgs/go-pkgs/npm`. Exercises filesystem
indicator scanning, PATH-aware resolution, and per-manager install argument
building.

## DSN (Domain Specific Notion)

### Participants

- **`DetectFromDir`** — scans a project root for lockfiles, `package.json`
  `packageManager` field, and defaults; returns a **`Trace`** with resolved
  **`Manager`** and collected **`Signal`** entries.
- **`DetectFromNodeModules`** — treats `node_modules` parent as project root;
  also inspects `node_modules/.pnpm` for pnpm store layout.
- **`HasPackageJSON`** — reports whether `package.json` exists beside
  `node_modules`.
- **`Resolve`** — picks a manager for a project: explicit name or `"auto"` /
  `""` with PATH availability fallback over detection candidates.
- **`InstallArgs` / `InstallCommand`** — build CLI argv for dependency install
  per manager and **`InstallOptions`** (`FrozenLockfile`).
- **`Manager`** — one of `pnpm`, `bun`, `npm`, `yarn`, `unknown`.
- **`Signal`** — `{Manager, Source}` indicator (lockfile name, pnpm store,
  packageManager field).
- **`Trace`** — `{ProjectRoot, NodeModulesAbsPath, Manager, HasPackageJSON,
  Signals, Steps}`.

### Behaviors

**Detection**

- Lockfile markers: `pnpm-lock.yaml` → pnpm; `bun.lockb` / `bun.lock` → bun;
  `package-lock.json` → npm; `yarn.lock` → yarn.
- `node_modules/.pnpm` directory → pnpm signal.
- `package.json` `packageManager` field (tool before `@`) → signal when known.
- **Priority** when multiple indicators: `pnpm > bun > npm > yarn`.
- No indicators + `package.json` only → default **npm**.
- No indicators + no `package.json` → **unknown**.

**Resolve**

- `pref` explicit known manager → that manager if CLI on PATH, else error.
- `pref` `"auto"` or `""` → detected candidates in priority order, first
  available on PATH; then global priority fallback; else error.
- Unknown `pref` → error listing expected values.

**Install**

- pnpm/bun/yarn frozen → `install --frozen-lockfile`; default → `install`.
- npm frozen → `ci`; default → `install --no-package-lock`.

## Decision Tree

```
package-manager
├── detect-from-dir              [DetectFromDir: project root indicators]
│   ├── pnpm-lock-only           package.json + pnpm-lock.yaml → pnpm
│   ├── bun-lock-only            package.json + bun.lock → bun
│   ├── bun-lockb-only           package.json + bun.lockb → bun
│   ├── npm-lock-only            package.json + package-lock.json → npm
│   ├── yarn-lock-only           package.json + yarn.lock → yarn
│   ├── mixed-all-lockfiles      all lockfiles + packageManager npm@10 → pnpm
│   ├── bun-over-npm             bun.lock + package-lock.json → bun
│   ├── package-manager-field    packageManager pnpm@11 only → pnpm
│   ├── package-json-only        package.json no field → npm
│   └── empty-project            no files → unknown
├── detect-from-node-modules     [DetectFromNodeModules + HasPackageJSON]
│   ├── pnpm-store               package.json + node_modules/.pnpm → pnpm
│   └── has-package-json         package.json + node_modules → HasPackageJSON
├── resolve                      [Resolve: explicit vs auto + PATH]
│   ├── explicit-pnpm            pref pnpm, empty dir → pnpm (PATH required)
│   ├── auto-pnpm-lock           pref auto, pnpm-lock.yaml → pnpm
│   ├── auto-mixed               pref auto, all lockfiles → pnpm
│   └── unknown-pref             pref yarnberry → error
└── install-args                 [InstallArgs / InstallCommand]
    ├── pnpm-default             pnpm, not frozen → install
    ├── pnpm-frozen              pnpm, frozen → install --frozen-lockfile
    ├── bun-default              bun, not frozen → install
    ├── npm-default              npm, not frozen → install --no-package-lock
    ├── npm-frozen               npm, frozen → ci
    └── yarn-frozen              yarn, frozen → install --frozen-lockfile
```

## Test Index

| Leaf | Op | Description |
|------|-----|-------------|
| `detect-from-dir/pnpm-lock-only` | detect-dir | pnpm-lock.yaml → pnpm |
| `detect-from-dir/bun-lock-only` | detect-dir | bun.lock → bun |
| `detect-from-dir/bun-lockb-only` | detect-dir | bun.lockb → bun |
| `detect-from-dir/npm-lock-only` | detect-dir | package-lock.json → npm |
| `detect-from-dir/yarn-lock-only` | detect-dir | yarn.lock → yarn |
| `detect-from-dir/mixed-all-lockfiles` | detect-dir | all lockfiles; pnpm wins |
| `detect-from-dir/bun-over-npm` | detect-dir | bun.lock beats package-lock.json |
| `detect-from-dir/package-manager-field` | detect-dir | packageManager field → pnpm |
| `detect-from-dir/package-json-only` | detect-dir | package.json only → npm default |
| `detect-from-dir/empty-project` | detect-dir | empty dir → unknown |
| `detect-from-node-modules/pnpm-store` | detect-node-modules | .pnpm store → pnpm |
| `detect-from-node-modules/has-package-json` | detect-node-modules | HasPackageJSON true |
| `resolve/explicit-pnpm` | resolve | explicit pnpm pref |
| `resolve/auto-pnpm-lock` | resolve | auto from pnpm-lock.yaml |
| `resolve/auto-mixed` | resolve | auto; all lockfiles → pnpm |
| `resolve/unknown-pref` | resolve | unknown pref → error |
| `install-args/pnpm-default` | install-args | pnpm default install |
| `install-args/pnpm-frozen` | install-args | pnpm frozen lockfile |
| `install-args/bun-default` | install-args | bun default install |
| `install-args/npm-default` | install-args | npm default install |
| `install-args/npm-frozen` | install-args | npm ci |
| `install-args/yarn-frozen` | install-args | yarn frozen lockfile |

## How to Run

```sh
doctest vet ./go-pkgs/npm/tests/package-manager/
doctest test ./go-pkgs/npm/tests/package-manager/
```

```go
import (
	"fmt"
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

type Request struct {
	Op              string
	ProjectDir      string
	NodeModulesPath string
	Pref            string
	Manager         npm.Manager
	FrozenLockfile  bool
}

type Response struct {
	Trace          npm.Trace
	Manager        npm.Manager
	HasPackageJSON bool
	Args           []string
	Command        string
	CommandArgs    []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Op {
	case "detect-dir":
		trace := npm.DetectFromDir(req.ProjectDir)
		return &Response{Trace: trace}, nil
	case "detect-node-modules":
		trace := npm.DetectFromNodeModules(req.NodeModulesPath)
		return &Response{Trace: trace}, nil
	case "has-package-json":
		has := npm.HasPackageJSON(req.NodeModulesPath)
		return &Response{HasPackageJSON: has}, nil
	case "resolve":
		manager, err := npm.Resolve(req.ProjectDir, req.Pref)
		if err != nil {
			return nil, err
		}
		return &Response{Manager: manager}, nil
	case "install-args":
		opts := npm.InstallOptions{FrozenLockfile: req.FrozenLockfile}
		args := npm.InstallArgs(req.Manager, opts)
		cmd, cmdArgs := npm.InstallCommand(req.Manager, opts)
		return &Response{Args: args, Command: cmd, CommandArgs: cmdArgs}, nil
	default:
		return nil, fmt.Errorf("unknown op %q", req.Op)
	}
}
```