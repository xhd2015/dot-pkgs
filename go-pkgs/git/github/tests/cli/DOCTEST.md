# RunCLI — `kool github` command-line interface

## Version
0.0.2

Nested doc-style test root for `RunCLI` in
`github.com/xhd2015/dot-pkgs/go-pkgs/git/github`. Exercises top-level
and `repo list` subcommand routing, flag parsing, stdout/stderr output modes,
and auth error surfacing. Mock `gh` backs integration leaves.

## DSN (Domain Specific Notion)

### Participants

- **`RunCLI`** — parses `args` (tokens after `kool github`), routes subcommands,
  prints usage or formatted results, returns error on failure.
- **`repo` subcommand** — dispatches `list` and dedicated repo-level help;
  errors on unknown verbs.
- **`repo list`** — parses less-flags (`--owner`, `--search-description`,
  `--search-code`, `--limit`, `--json`), calls `ListRepos`, formats output.
- **`ListRepos`** — library entry used by `repo list`; auth gate and search
  union (tested separately under `list-repos/`).
- **`gh` CLI** — invoked indirectly via `ListRepos`; mocked in integration
  leaves.

### Behaviors

- Empty top-level args → same top-level help as `--help`/`-h`/`help` on stdout,
  exit 0 (trailing `\n`).
- Unknown top-level command → unrecognized-command error on stderr, non-zero exit.
- `repo` alone or `repo --help`/`-h`/`help` → **repo-level** help (mentions
  `list` and that `repo list --help` shows list options); not list leaf help;
  exit 0.
- Unknown `repo` subcommand → error.
- `repo list --help` → list usage on stdout (flags: `--owner`, `--json`, …), exit 0.
- Default output: one line per repo `{full_name}\t{matched_by}` (comma-joined).
- `--json`: indented JSON array of `RepoResult` objects.
- Errors from `ListRepos` → message on stderr, non-zero exit.
- `GH_BIN` env overrides gh executable path (same as library tests).

## Decision Tree

```
cli/
├── help
│   ├── top-level                 RunCLI(["--help"])
│   ├── empty-args                RunCLI([])
│   ├── repo                      RunCLI(["repo"])
│   ├── repo-help                 RunCLI(["repo","--help"])
│   └── repo-list                 RunCLI(["repo","list","--help"])
├── errors
│   ├── unknown-command           RunCLI(["nope"])
│   ├── unknown-repo-sub          RunCLI(["repo","nope"])
│   ├── unknown-flag              RunCLI(["repo","list","--nope"])
│   └── unexpected-args           RunCLI(["repo","list","extra"])
├── output
│   ├── lines-default             tab-separated owned repos
│   ├── json                      --json multi-repo array
│   └── json-empty                --json empty result
├── flags
│   ├── owner                     --owner alice
│   └── search-description        --search-description widget
└── auth
    └── not-authenticated         stderr gh auth login hint
```

## Test Index

| Leaf | Description |
|------|-------------|
| `help/top-level` | Top-level `--help` prints usage, exit 0 |
| `help/empty-args` | Empty args print top-level usage, exit 0 |
| `help/repo` | Bare `repo` prints repo-level help, exit 0 |
| `help/repo-help` | `repo --help` prints repo-level help, exit 0 |
| `help/repo-list` | `repo list --help` prints list flags, exit 0 |
| `errors/unknown-command` | Unrecognized top-level command |
| `errors/unknown-repo-sub` | Unrecognized `repo` subcommand |
| `errors/unknown-flag` | Unknown `repo list` flag |
| `errors/unexpected-args` | Trailing positional args rejected |
| `output/lines-default` | Default tab-separated line output |
| `output/json` | `--json` indented multi-repo output |
| `output/json-empty` | `--json` with empty result `[]` |
| `flags/owner` | `--owner` passes through to gh repo list |
| `flags/search-description` | Description search line output |
| `auth/not-authenticated` | Auth failure stderr hint |

## How to Run

```sh
doctest vet ./go-pkgs/git/github/tests/cli/
doctest test -v ./go-pkgs/git/github/tests/cli/...
```

```go
import (
	"io"
	"os"
	"testing"

	ghrepos "github.com/xhd2015/dot-pkgs/go-pkgs/git/github"
)

type Request struct {
	Args  []string
	GhBin string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.GhBin != "" {
		t.Setenv("GH_BIN", req.GhBin)
	} else {
		t.Setenv("GH_BIN", "")
	}

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

	runErr := ghrepos.RunCLI(req.Args)

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutBytes, readErr := io.ReadAll(stdoutR)
	if readErr != nil {
		return nil, readErr
	}
	stderrBytes, readErr := io.ReadAll(stderrR)
	if readErr != nil {
		return nil, readErr
	}

	exitCode := 0
	if runErr != nil {
		exitCode = 1
	}

	return &Response{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
	}, nil
}
```