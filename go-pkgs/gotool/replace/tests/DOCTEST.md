# CheckLocalReplaces — Doc-Style Test Tree

## Version
0.0.2

Tests for the shared `CheckLocalReplaces` library function that scans a Go module tree for local filesystem `replace` directives and classifies them as intra-repo or extra-repo.

# DSN (Domain Specific Notion)

- **Repo top** — the root directory of a git repository, passed to `CheckLocalReplaces` as `top`.
- **Scan** — the library uses `gotool/mod/scan` to walk all go.mod files under `top` and discover Go modules.
- **Module** — each discovered module has a `Dir` (absolute path) and a list of `replace` directives.
- **Local filesystem replace** — a replace directive whose `NewVersion` is empty and whose `NewPath` is a relative (`./` or `../`) or absolute path. These point at a local checkout rather than a published module version.
- **Intra-repo replace** — the replace target path resolves to a directory that exists and lives under the same git toplevel as `top`.
- **Extra-repo replace** — the replace target path resolves to a directory outside the `top` git toplevel, or the target directory does not exist.
- **ReplaceIssue** — each local replace becomes a `ReplaceIssue` with `GoModPath`, `OldPath`, `NewPath`, and `IsIntraRepo`.
- **Output** — the function returns `[]ReplaceIssue` and `error`. An empty slice with nil error means no local replaces found. A non-nil error means the scan itself failed.

## How to Run

```sh
doctest test -v ./
```

## Test Tree

```
no-go-mod/          — no go.mod files, no issues
has-go-mod/
├── no-replaces/    — go.mod with no replace directives, no issues
├── version-replace/
│   └── only/       — go.mod with only version-based replaces, no issues
├── local-replace/
│   ├── intra-repo/
│   │   ├── dot-slash/      — ./sub inside same repo → intra-repo issue
│   │   ├── dot-dot-slash/  — ../sibling inside same repo → intra-repo issue
│   │   ├── abs-path/       — absolute path inside same repo → intra-repo issue
│   │   └── git-dir-env/    — inherited GIT_DIR from hook env must not change intra-repo classification
│   ├── extra-repo/
│   │   ├── dot-slash/      — ./external/dep outside repo → extra-repo issue
│   │   └── abs-path/       — absolute path outside repo → extra-repo issue
│   └── multi-module/
│       ├── single-issue/   — root clean, sub has local replace → issue in sub
│       └── mixed/          — root intra-repo, sub extra-repo → both found
```

## Test Index

| Leaf | Description | Expected Issues |
|------|-------------|----------------|
| `no-go-mod` | No go.mod files in repo | 0 issues |
| `has-go-mod/no-replaces` | go.mod with no replace directives | 0 issues |
| `has-go-mod/version-replace/only` | go.mod with version-only replaces | 0 issues |
| `has-go-mod/local-replace/intra-repo/dot-slash` | `./sub` inside same repo | 1 issue, IsIntraRepo=true |
| `has-go-mod/local-replace/intra-repo/dot-dot-slash` | `../sibling` inside same repo | 1 issue, IsIntraRepo=true |
| `has-go-mod/local-replace/intra-repo/abs-path` | absolute path inside same repo | 1 issue, IsIntraRepo=true |
| `has-go-mod/local-replace/intra-repo/git-dir-env` | `../` replace from nested module while `GIT_DIR` is inherited from hook env | 1 issue, IsIntraRepo=true |
| `has-go-mod/local-replace/extra-repo/dot-slash` | `./external/dep` outside repo | 1 issue, IsIntraRepo=false |
| `has-go-mod/local-replace/extra-repo/abs-path` | absolute path outside repo | 1 issue, IsIntraRepo=false |
| `has-go-mod/local-replace/multi-module/single-issue` | root clean, sub has local replace | 1 issue in sub |
| `has-go-mod/local-replace/multi-module/mixed` | root intra-repo, sub extra-repo | 2 issues, mixed flags |

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
)

type Request struct {
	RootDir string
}

type Response struct {
	Issues []replace.ReplaceIssue
	Err    error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	issues, err := replace.CheckLocalReplaces(req.RootDir)
	return &Response{
		Issues: issues,
		Err:    err,
	}, nil
}
```
