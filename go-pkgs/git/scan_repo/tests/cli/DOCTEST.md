# scan_repo CLI — `RunCLI` Flag Parsing and Output

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo.RunCLI`. The CLI
parses `less-flags` arguments, calls `Scan`, and prints discovered repos as
tab-separated lines (default) or JSON (`--json`). Errors go to stderr; successful
empty scans exit 0 with empty stdout.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies CLI argv (`[]string`) with flags only (no config file).
- **RunCLI** — parses flags via `less-flags`, validates `--root` is present,
  maps flags to `Options` (full-path `--ignore-dir`, basename `--ignore-dir-basename`,
  `--verbose` for permission-skip warnings, cache control flags below), invokes `Scan`.
- **Scan** — filesystem walk, optional git enrichment, optional mirror cache
  (same as library tests).
- **Stdout formatter** — lines mode: `{path}\t{RepoType}` plus optional
  `\torigin:{owner}/{repo}@{host}` when `--list-remotes` and origin exists;
  JSON mode: marshaled `[]Repo` with `RepoType` as `"main"`/`"worktree"`.
- **Stderr** — usage/help text, validation errors, scan failures.
- **Cache control** — CLI maps `--no-cache`, `--refresh`, `--cache-dir PATH` onto
  `Options.NoCache`, `Options.Refresh`, `Options.CacheRoot`. When cache is on and
  `--cache-dir` is omitted, product default is `$HOME/.cache/git-repo-scan`.

### Behaviors

**Invocation modes**

- `--help` / `-h` — print usage to stdout, exit 0.
- Missing `--root` — error on stderr mentioning roots required, non-zero exit.
- Unknown flag — parse error on stderr, non-zero exit.
- Valid scan — map flags to `Options`, run `Scan`, format stdout.

**Cache flags (P5)**

- `--no-cache` — `NoCache=true`: full live walk; no mirror read/write under
  `--cache-dir` (or default).
- `--refresh` — `Refresh=true`: force cold full walk + rewrite even when warm-
  eligible (finds brand-new repos warm would miss).
- `--cache-dir PATH` — `CacheRoot=PATH`: cold/warm mirror under that root.
- Default cache root when cache enabled and `--cache-dir` empty:
  `$HOME/.cache/git-repo-scan`.

**Output**

- Default lines: one repo per line, path-sorted; empty scan → empty stdout, exit 0.
- `--json`: JSON array; empty scan → `[]`.
- `--list-remotes`: append origin column on lines output when origin remote exists.
- `--list-worktrees`: enrichment only; lines output still `{path}\t{RepoType}`.

## Decision Tree

```
cli
├── help/                         [argv contains -h or --help]
│   └── show/                     # documents discovery + cache flags
├── errors/                       [invalid argv — no scan]
│   ├── no-roots/               [no --root flag]
│   └── unknown-flag/           [unrecognized flag]
├── flags/                        [valid scan, default lines; no cache flags]
│   ├── single-root/
│   ├── multiple-roots/
│   ├── max-depth/
│   ├── ignore-dir/                    [full normalized path]
│   ├── ignore-dir-no-basename-match/  [relative path does not basename-skip]
│   ├── ignore-dir-basename/
│   ├── verbose-permission-skip/       [-v + unreadable dir]
│   ├── verbose-quiet-default/         [no -v, silent skip]
│   ├── verbose-remote-skip/           [-v + CloudStorage skip warning]
│   └── verbose-quiet-remote/          [no -v, silent CloudStorage skip]
├── cache/                        [P5 — --no-cache / --refresh / --cache-dir]
│   ├── no-cache/               # --no-cache + --cache-dir → discover, no mirror write
│   ├── cache-dir/              # --cache-dir cold scan writes under given dir
│   └── refresh/                # warm seed then --refresh finds brand-new
├── output/                       [format selection]
│   ├── lines-default/          [--json absent, multi-repo fixture]
│   ├── json/                   [--json, multi-repo fixture]
│   └── json-empty/             [--json, empty workspace]
└── enrich/                       [--list-remotes or --list-worktrees]
    ├── list-remotes/
    └── list-worktrees/
```

## Test Index

| Leaf | Branch | Description |
|------|--------|-------------|
| `help/show` | Help | `--help` prints usage with all flags (incl. cache) |
| `errors/no-roots` | Errors | No `--root` → stderr mentions roots |
| `errors/unknown-flag` | Errors | Unknown flag → parse error |
| `flags/single-root` | Flags | One `--root`, one repo discovered |
| `flags/multiple-roots` | Flags | Two `--root` values union results |
| `flags/max-depth` | Flags | `--max-depth` excludes deep repo |
| `flags/ignore-dir` | Flags | `--ignore-dir` skips by full path |
| `flags/ignore-dir-no-basename-match` | Flags | Relative `--ignore-dir` does not basename-skip |
| `flags/ignore-dir-basename` | Flags | `--ignore-dir-basename` skips by basename |
| `flags/verbose-permission-skip` | Flags | `-v` warns on permission-denied skip |
| `flags/verbose-quiet-default` | Flags | Default: no stderr on permission skip |
| `flags/verbose-remote-skip` | Flags | `-v` warns on CloudStorage skip |
| `flags/verbose-quiet-remote` | Flags | Default: no stderr on CloudStorage skip |
| `cache/no-cache` | Cache | `--no-cache` discovers repos; no `entry.json` under `--cache-dir` |
| `cache/cache-dir` | Cache | `--cache-dir` cold scan writes mirror under given path |
| `cache/refresh` | Cache | after warm seed, `--refresh` lists brand-new repo |
| `output/lines-default` | Output | Tab-separated lines, path-sorted |
| `output/json` | Output | JSON array with string RepoType |
| `output/json-empty` | Output | Empty workspace → `[]` |
| `enrich/list-remotes` | Enrich | Lines with `origin:owner/repo@host` |
| `enrich/list-worktrees` | Enrich | Main + worktree rows via git |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cli/
doctest test -v ./go-pkgs/git/scan_repo/tests/cli/
```

From monorepo root:

```sh
doctest test -v ./external/dot-pkgs-cli/go-pkgs/git/scan_repo/tests/cli/
```

```go
import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Args []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW

	runErr := scan_repo.RunCLI(req.Args)

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, stdoutR); err != nil {
		return nil, err
	}
	if _, err := io.Copy(&stderrBuf, stderrR); err != nil {
		return nil, err
	}

	exitCode := 0
	if runErr != nil {
		exitCode = 1
	}

	return &Response{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}, nil
}
```