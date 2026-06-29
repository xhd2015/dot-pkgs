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
  `--verbose` for permission-skip warnings), invokes `Scan`.
- **Scan** — filesystem walk and optional git enrichment (same as library tests).
- **Stdout formatter** — lines mode: `{path}\t{RepoType}` plus optional
  `\torigin:{owner}/{repo}@{host}` when `--list-remotes` and origin exists;
  JSON mode: marshaled `[]Repo` with `RepoType` as `"main"`/`"worktree"`.
- **Stderr** — usage/help text, validation errors, scan failures.

### Behaviors

**Invocation modes**

- `--help` / `-h` — print usage to stdout, exit 0.
- Missing `--root` — error on stderr mentioning roots required, non-zero exit.
- Unknown flag — parse error on stderr, non-zero exit.
- Valid scan — map flags to `Options`, run `Scan`, format stdout.

**Output**

- Default lines: one repo per line, path-sorted; empty scan → empty stdout, exit 0.
- `--json`: JSON array; empty scan → `[]`.
- `--list-remotes`: append origin column on lines output when origin remote exists.
- `--list-worktrees`: enrichment only; lines output still `{path}\t{RepoType}`.

## Decision Tree

```
cli
├── help/                         [argv contains -h or --help]
│   └── show/
├── errors/                       [invalid argv — no scan]
│   ├── no-roots/               [no --root flag]
│   └── unknown-flag/           [unrecognized flag]
├── flags/                        [valid scan, default lines output]
│   ├── single-root/
│   ├── multiple-roots/
│   ├── max-depth/
│   ├── ignore-dir/                    [full normalized path]
│   ├── ignore-dir-no-basename-match/  [relative path does not basename-skip]
│   ├── ignore-dir-basename/
│   ├── verbose-permission-skip/       [-v + unreadable dir]
│   └── verbose-quiet-default/         [no -v, silent skip]
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
| `help/show` | Help | `--help` prints usage with all flags |
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