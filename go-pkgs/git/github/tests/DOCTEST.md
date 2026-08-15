# GitHub Repos — `ListOwned`, `ListRepos`, and URL helpers via `gh`

## Version
0.0.2

Doc-style tests for the library package at
`github.com/xhd2015/dot-pkgs/go-pkgs/git/github`. The package lists
repositories via `gh` (owned repos, description search, code search), merges
multi-owner results with `matched_by` provenance, and normalizes repo URLs.
Tests use mock `gh` shell scripts — no real GitHub network or auth.

## DSN (Domain Specific Notion)

### Participants

- **`ListOwned`** — public entry point: validates `Options`, iterates
  `opts.Owners`, invokes `gh` per owner, merges JSON results, dedupes by
  `FullName`, sorts ascending, returns `[]Repo`.
- **`Options`** — caller configuration: required non-empty owner list, per-owner
  limit, archived/fork inclusion flags, and `GhBin` path override.
- **`gh` CLI** — external process: `gh repo list <owner> --json … --limit N`
  with optional `--no-archived` and `--source` flags; stdout is JSON array of
  repo objects with nested `owner.login`.
- **`GH_BIN`** — environment override read by `DefaultOptions` / exec layer;
  doctest harness sets it to a mock script path.
- **`Repo`** — normalized domain record: `Owner`, `Name`, `FullName`,
  `URL` (https canonical), `Description`, `IsFork`, `IsArchived`.
- **`NormalizeRepoURL`** — pure helper: converts SSH/git/https raw URLs from
  `gh` into `https://github.com/<owner>/<name>`.
- **`EnsureAuthenticated`** — auth gate: runs `gh api user`, returns login on
  success; on failure error contains `gh auth login`.
- **`ListRepos`** — unified entry point: auth, resolve owners (explicit or
  inferred login), optional description/code search, union by `FullName`, sort
  ascending, return `[]RepoResult`.
- **`ListReposOptions`** — caller configuration: optional owners (empty → infer
  from auth login), `SearchDescription`, `SearchCode`, limit (0 → default 30),
  and `GhBin` override.
- **`RepoResult`** — `Repo` plus `MatchedBy []MatchReason` (`owned`,
  `description`, `code`).
- **`MatchReason`** — string enum tagging why a repo appeared in results.

### Behaviors

**ListOwned**

- **Validate** — reject empty `Owners` slice; reject any empty owner string
  before spawning `gh`.
- **Invoke** — `exec.CommandContext` per owner with JSON field list, limit
  (0 → default 100), `--no-archived` when `IncludeArchived` is false,
  `--source` when `IncludeForks` is false.
- **Decode** — map `owner.login` → `Repo.Owner`; build `FullName` as
  `owner/name`; normalize `URL` via `NormalizeRepoURL`.
- **Merge** — concatenate per-owner results; dedupe by `FullName` keeping
  first occurrence (owner order, then JSON order); sort `FullName` ascending.
- **Empty** — all owners return `[]` → `[]Repo{}`, not an error.

**gh errors**

- Missing binary → error contains `gh not found`.
- Exit 4 + auth stderr → error hints `gh auth login`.
- Other non-zero exit → error mentions owner and stderr snippet.
- Invalid JSON stdout → decode/json error.

**NormalizeRepoURL**

- Accept SSH (`git@github.com:o/r.git`), git (`git://…`), and https inputs;
  always return `https://github.com/<owner>/<name>`.

**EnsureAuthenticated**

- Invoke `gh api user`; parse JSON `login` field (or `--jq .login` equivalent).
- Not logged in → error containing `gh auth login`; no further gh calls.

**ListRepos**

- **Auth** — `EnsureAuthenticated` before any repo query; fail fast.
- **Owners** — non-empty `Owners` used as-is (validate non-blank strings); empty
  → single owner from auth login.
- **Limit** — 0 → default **30** on all gh subcommands.
- **Plain** — both search keywords empty → `ListOwned` per owner; each result
  `matched_by: ["owned"]`.
- **Description search** — `gh search repos "<kw>" --owner <owner> --json …
  --limit N`; `matched_by: ["description"]`.
- **Code search** — `gh search code "<kw>" --owner <owner> --json repository
  --limit N`; dedupe per owner; `matched_by: ["code"]`.
- **Union** — both keywords set → union by `FullName`; merge `matched_by` (e.g.
  `["description","code"]` when both match).
- **Sort** — `FullName` ascending; empty → `[]RepoResult{}`, not error.

## Decision Tree

```
git/github/tests
├── list-owned                     [ListOwned: gh repo list per owner]
│   ├── success                    [happy path: merge, map, sort]
│   │   ├── single-owner           1 owner → 2 repos, fields + sort
│   │   ├── multi-owner            2 owners → disjoint merged sorted list
│   │   ├── dedupe-same-repo       same FullName from 2 owners → 1 entry
│   │   └── empty-result           mock returns [] → empty slice, no error
│   ├── options                    [gh flag forwarding]
│   │   ├── limit                  Limit 42 → --limit 42
│   │   ├── skip-archived          IncludeArchived false → --no-archived
│   │   ├── include-archived       IncludeArchived true → no --no-archived
│   │   └── exclude-forks          IncludeForks false → --source
│   └── errors
│       ├── validation             [pre-exec validation]
│       │   ├── empty-owners       Owners [] → error, gh never called
│       │   └── empty-owner-string blank owner → invalid owner error
│       └── gh                     [gh process failures]
│           ├── gh-missing         GhBin nonexistent → gh not found
│           ├── gh-auth-fail       exit 4 → gh auth login hint
│           ├── gh-exit-other        exit 1 → owner + stderr in error
│           └── gh-invalid-json      non-JSON stdout → decode error
├── list-repos/DOCTEST.md          [nested root: ListRepos auth + search union]
│   ├── auth                       [EnsureAuthenticated gate]
│   │   ├── not-authenticated      gh api user fails → gh auth login hint
│   │   └── infer-owner            empty Owners → login drives repo list owner
│   ├── plain                      [no search keywords]
│   │   └── owned-only             ListOwned path → matched_by ["owned"]
│   ├── search                     [gh search repos/code]
│   │   ├── description-only       search repos → matched_by ["description"]
│   │   ├── code-only              search code deduped → matched_by ["code"]
│   │   └── union-both             OR merge → ["description"], ["code"], both
│   └── options                    [ListRepos-specific options]
│       ├── limit-default-30       Limit 0 → gh sees --limit 30
│       └── multi-owner            two owners merged sorted with ["owned"]
└── parse                          [pure URL / name helpers]
    ├── normalize-url              SSH, git, https → canonical https URL
    └── full-name                  owner + name → owner/name
```

## Test Index

| Leaf | Branch | Description |
|------|--------|-------------|
| `list-owned/success/single-owner` | success | One owner, two repos mapped and sorted |
| `list-owned/success/multi-owner` | success | Two owners, disjoint repos merged sorted |
| `list-owned/success/dedupe-same-repo` | success | Duplicate FullName deduped, first wins |
| `list-owned/success/empty-result` | success | Empty JSON array → `[]Repo{}`, nil error |
| `list-owned/options/limit` | options | `--limit 42` forwarded to gh |
| `list-owned/options/skip-archived` | options | Default skips archived via `--no-archived` |
| `list-owned/options/include-archived` | options | IncludeArchived omits `--no-archived` |
| `list-owned/options/exclude-forks` | options | IncludeForks false adds `--source` |
| `list-owned/errors/validation/empty-owners` | validation | No owners → error before gh |
| `list-owned/errors/validation/empty-owner-string` | validation | Blank owner string rejected |
| `list-owned/errors/gh/gh-missing` | gh | Missing gh binary path |
| `list-owned/errors/gh/gh-auth-fail` | gh | Auth failure exit code 4 |
| `list-owned/errors/gh/gh-exit-other` | gh | Generic non-zero exit |
| `list-owned/errors/gh/gh-invalid-json` | gh | Malformed JSON on stdout |
| `parse/normalize-url` | parse | SSH/git/https URL normalization |
| `parse/full-name` | parse | FullName construction `o/r` |
| `list-repos/auth/not-authenticated` | auth | `gh api user` fails → auth hint |
| `list-repos/auth/infer-owner` | auth | Empty owners → inferred login `alice` |
| `list-repos/plain/owned-only` | plain | No search → 2 repos `["owned"]` |
| `list-repos/search/description-only` | search | Description search results |
| `list-repos/search/code-only` | search | Code search deduped results |
| `list-repos/search/union-both` | search | Union merges `matched_by` |
| `list-repos/options/limit-default-30` | options | Default limit 30 forwarded to gh |
| `list-repos/options/multi-owner` | options | Two owners merged sorted |

## How to Run

```sh
doctest vet ./go-pkgs/git/github/tests/
doctest test ./go-pkgs/git/github/tests/...

# ListRepos nested root (separate DOCTEST.md boundary):
doctest vet ./go-pkgs/git/github/tests/list-repos/
doctest test ./go-pkgs/git/github/tests/list-repos/...
```

```go
import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	ghrepos "github.com/xhd2015/dot-pkgs/go-pkgs/git/github"
)

type Request struct {
	Owners          []string
	Limit           int
	IncludeArchived bool
	IncludeForks    bool
	GhBin           string
	// parse leaves:
	NormalizeOwner string
	NormalizeName  string
	NormalizeInput string
	FullNameOwner  string
	FullNameName   string
	ParseRefInput  string
}

type Response struct {
	Repos         []ghrepos.Repo
	Normalized    string
	FullName      string
	ParseRefOwner string
	ParseRefName  string
	GhArgv        string // captured "$*" from mock gh, for options/* leaves
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.ParseRefInput != "" {
		owner, name, err := ghrepos.ParseRef(req.ParseRefInput)
		if err != nil {
			return nil, err
		}
		return &Response{ParseRefOwner: owner, ParseRefName: name}, nil
	}
	if req.NormalizeInput != "" {
		normalized := ghrepos.NormalizeRepoURL(req.NormalizeOwner, req.NormalizeName, req.NormalizeInput)
		return &Response{Normalized: normalized}, nil
	}
	if req.FullNameOwner != "" {
		fullName := req.FullNameOwner + "/" + req.FullNameName
		return &Response{FullName: fullName}, nil
	}

	if req.GhBin != "" {
		t.Setenv("GH_BIN", req.GhBin)
	}
	opts := ghrepos.Options{
		Owners:          req.Owners,
		Limit:           req.Limit,
		IncludeArchived: req.IncludeArchived,
		IncludeForks:    req.IncludeForks,
		GhBin:           req.GhBin,
	}
	repos, err := ghrepos.ListOwned(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	resp := &Response{Repos: repos}
	if req.GhBin != "" {
		argvPath := filepath.Join(filepath.Dir(req.GhBin), "gh.argv")
		if data, readErr := os.ReadFile(argvPath); readErr == nil {
			resp.GhArgv = strings.TrimSpace(string(data))
		}
	}
	return resp, nil
}
```