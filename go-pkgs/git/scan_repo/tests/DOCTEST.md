# scan_repo — Git Repository Discovery and Enrichment

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo`. `Scan` walks
filesystem roots to discover git checkouts (`.git` directory or gitlink file).
`ParseRemoteOwnerRepo` parses remote URLs into host, owner, and repo name.
Optional enrichment lists remotes and worktrees via git subprocesses.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies one or more filesystem roots and scan options.
- **Scan** — validates roots, walks each tree, discovers repos, optionally enriches.
- **Walk** — `filepath.WalkDir` from each root; applies ignore config (full paths
  and basenames); on permission errors skips the directory (`SkipDir`) instead of
  aborting; stops descending into a discovered repo (`SkipDir` boundary).
- **Ignore config** — `IgnoreDirs` are normalized full paths (exact match);
  `IgnoreDirBasenames` union default basenames (`.git`, `node_modules`, …) for
  directory name matches anywhere in the tree.
- **Repo detector** — classifies `.git` as directory (`RepoTypeMain`) or gitlink
  file (`RepoTypeWorktree`); resolves `GitDir` to absolute storage path.
- **Enricher** — when `ListRemotes` or `ListWorktrees` is set, runs `git -C`
  subprocesses per discovered entry; worktrees attach only to main rows.
- **ParseRemoteOwnerRepo** — pure URL parser; no filesystem or git access.

### Behaviors

**Scan (discovery)**

- Require at least one root; each root must exist and be a directory.
- Expand `~`, absolutize and clean paths; sort results by `Path` ascending.
- Apply default ignore basenames unioned with `IgnoreDirBasenames`.
- Skip directories whose normalized full path is listed in `IgnoreDirs`.
- When `Verbose` is true, log permission-denied skips to stderr as warnings.
- Respect `MaxDepth` relative to each root (0 = unlimited).
- Option A: every checkout with `.git` is its own row; no dedup by `GitDir`.

**Enrichment**

- `ListRemotes=false` — no git calls; `Remotes` empty.
- `ListRemotes=true` — `git remote` + config URL per remote on every row.
- `ListWorktrees=false` — no git calls; `Worktrees` empty.
- `ListWorktrees=true` — `git worktree list --porcelain` only on `RepoTypeMain`.

**ParseRemoteOwnerRepo**

- Parse GitHub HTTPS, SSH, and SCP-style URLs into owner and repo.
- Return `ok=false` for unparseable input.

## Decision Tree

```
scan-repo
├── parse-remote/              [req.ParseURL set — pure parser, no FS/git]
│   ├── parse-github-ssh/
│   ├── parse-github-https/
│   ├── parse-scp-style/
│   └── parse-invalid/
├── scan/                      [ListRemotes=false, ListWorktrees=false]
│   ├── single-repo/
│   ├── sibling-repos/
│   ├── no-repos/
│   ├── repo-boundary/
│   ├── max-depth/
│   ├── ignore-dirs/              [default basename: node_modules]
│   ├── ignore-dir-basename/      [custom IgnoreDirBasenames]
│   ├── ignore-dir-full-path/     [IgnoreDirs full path]
│   ├── permission-denied-skip/   [WalkDir EACCES → SkipDir]
│   ├── gitlink-worktree/
│   ├── main-and-linked/
│   ├── empty-roots-error/
│   ├── missing-root-error/
│   └── not-a-directory-error/
├── enrich-remotes/            [ListRemotes=true, ListWorktrees=false]
│   ├── no-remotes/
│   ├── single-origin/
│   ├── multiple-remotes/
│   └── flags-false-skips-git/
├── enrich-worktrees/          [ListWorktrees=true]
│   ├── main-only/
│   ├── main-plus-linked/
│   └── flags-false-skips-git/
└── find-github/               [FindLocalMainByGitHub]
    ├── basename-mismatch/     clone dir name != github repo name
    └── skips-worktree/        returns main, not linked worktree
```

## Test Index

| Leaf | Mode | Description |
|------|------|-------------|
| `parse-remote/parse-github-ssh` | Parse | SSH URL → owner/repo |
| `parse-remote/parse-github-https` | Parse | HTTPS URL → owner/repo |
| `parse-remote/parse-scp-style` | Parse | Enterprise SCP URL → owner/repo |
| `parse-remote/parse-invalid` | Parse | Unparseable URL → ok=false |
| `scan/single-repo` | Scan | One main repo discovered |
| `scan/sibling-repos` | Scan | Two sibling repos, path-sorted |
| `scan/no-repos` | Scan | Empty tree → empty slice |
| `scan/repo-boundary` | Scan | Nested `.git` inside found repo skipped |
| `scan/max-depth` | Scan | Deep repo excluded by MaxDepth |
| `scan/ignore-dirs` | Scan | `node_modules` default basename ignore |
| `scan/ignore-dir-basename` | Scan | Custom `IgnoreDirBasenames` skips tree |
| `scan/ignore-dir-full-path` | Scan | `IgnoreDirs` exact full path skips tree |
| `scan/permission-denied-skip` | Scan | Unreadable child dir; scan still succeeds |
| `scan/gitlink-worktree` | Scan | Gitlink → RepoTypeWorktree |
| `scan/main-and-linked` | Scan | Main + linked worktree as two rows |
| `scan/empty-roots-error` | Scan | No roots → error |
| `scan/missing-root-error` | Scan | Missing root path → error |
| `scan/not-a-directory-error` | Scan | File root → error |
| `enrich-remotes/no-remotes` | Enrich | Git init, empty Remotes |
| `enrich-remotes/single-origin` | Enrich | Single origin remote parsed |
| `enrich-remotes/multiple-remotes` | Enrich | origin + upstream remotes |
| `enrich-remotes/flags-false-skips-git` | Enrich | ListRemotes=false, fake repo OK |
| `enrich-worktrees/main-only` | Enrich | Worktrees on main row only |
| `enrich-worktrees/main-plus-linked` | Enrich | Two rows; Worktrees only on main |
| `enrich-worktrees/flags-false-skips-git` | Enrich | ListWorktrees=false, fake repo OK |
| `find-github/basename-mismatch` | Find | `myproject-clone` + origin `xhd2015/myproject` |
| `find-github/skips-worktree` | Find | linked worktree skipped; main path returned |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/
doctest test -v ./go-pkgs/git/scan_repo/tests/
```

From monorepo root:

```sh
doctest test -v ./external/dot-pkgs-cli/go-pkgs/git/scan_repo/tests/
```

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Roots                []string
	MaxDepth             int
	IgnoreDirs           []string
	IgnoreDirBasenames   []string
	Verbose              bool
	ListRemotes          bool
	ListWorktrees        bool
	ParseURL             string // non-empty → ParseRemoteOwnerRepo only
	FindGitHubOwner      string
	FindGitHubRepo       string
}

type Response struct {
	Repos   []scan_repo.Repo
	Found   *scan_repo.Repo
	Owner   string
	Repo    string
	ParseOK bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.ParseURL != "" {
		owner, repo, ok := scan_repo.ParseRemoteOwnerRepo(req.ParseURL)
		return &Response{Owner: owner, Repo: repo, ParseOK: ok}, nil
	}
	if req.FindGitHubOwner != "" || req.FindGitHubRepo != "" {
		found, err := scan_repo.FindLocalMainByGitHub(context.Background(), scan_repo.Options{
			Roots: req.Roots,
		}, req.FindGitHubOwner, req.FindGitHubRepo)
		if err != nil {
			return nil, err
		}
		return &Response{Found: found}, nil
	}
	repos, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:                req.Roots,
		MaxDepth:             req.MaxDepth,
		IgnoreDirs:           req.IgnoreDirs,
		IgnoreDirBasenames:   req.IgnoreDirBasenames,
		Verbose:              req.Verbose,
		ListRemotes:          req.ListRemotes,
		ListWorktrees:        req.ListWorktrees,
	})
	if err != nil {
		return nil, err
	}
	return &Response{Repos: repos}, nil
}
```