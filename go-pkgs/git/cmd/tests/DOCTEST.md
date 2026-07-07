# git/cmd — Git Subprocess Helpers

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd`. `Run` executes
`git -C <dir>` with stdout capture and one-line error gist. `RunOptional` returns
`(output, ok, err)` when git is missing or the command fails benignly.
`RunCombined` captures combined stdout+stderr.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies working directory, git args, and operation mode.
- **Runner** — spawns `git` with `GIT_OPTIONAL_LOCKS=0` when appropriate.
- **Git** — external binary on PATH; tests skip when unavailable.

### Behaviors

- **Run** — success returns trimmed stdout; failure returns normalized one-line error.
- **RunOptional** — distinguishes missing git / command failure via `ok` flag.
- **RunCombined** — like Run but merges stderr into output on success path.

## Decision Tree

```
cmd
└── run/
    ├── success/        # rev-parse in temp git repo
    └── missing-repo/   # non-repo directory → error
```

## Test Index

| Leaf | Description |
|------|-------------|
| `run/success` | `git rev-parse --is-inside-work-tree` in initialized repo |
| `run/missing-repo` | Run in plain temp dir without `.git` → error |

## How to Run

```sh
doctest vet ./go-pkgs/git/cmd/tests/
doctest test -v ./go-pkgs/git/cmd/tests/
```

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
)

type Request struct {
	Dir  string
	Args []string
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	out, err := cmd.Run(context.Background(), req.Dir, req.Args...)
	if err != nil {
		return nil, err
	}
	return &Response{Output: out}, nil
}
```