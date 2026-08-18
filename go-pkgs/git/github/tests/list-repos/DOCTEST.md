# ListRepos — auth, search union, `matched_by`

## Version
0.0.2

Nested doc-style test root for `ListRepos` and `EnsureAuthenticated` in
`github.com/xhd2015/dot-pkgs/go-pkgs/git/github`. Exercises auth
gate, owner inference, plain owned mode, description/code search, and union
merge. Mock `gh` handles `api user`, `repo list`, `search repos`, and
`search code`.

## DSN (Domain Specific Notion)

### Participants

- **`ListRepos`** — unified entry: auth, resolve owners, search union, sort,
  return `[]RepoResult` with `matched_by` provenance.
- **`EnsureAuthenticated`** — `gh api user` gate; returns login or auth error.
- **`ListReposOptions`** — owners (empty → infer), search keywords, limit
  (0 → 30), `GhBin`.
- **`RepoResult`** — `Repo` plus `MatchedBy []MatchReason`.
- **`gh` CLI** — `api user`, `repo list`, `search repos`, `search code`.

### Behaviors

- Auth before any repo query; failure hints `gh auth login`.
- Plain mode (no search) delegates to owned listing with `matched_by: ["owned"]`.
- Search modes tag `description` or `code`; both keywords union by `FullName`.
- Results sorted ascending; empty → `[]RepoResult{}`.

## Decision Tree

```
list-repos/
├── auth
│   ├── not-authenticated      gh api user fails
│   └── infer-owner            empty Owners → login drives repo list
├── plain
│   └── owned-only             no search → matched_by ["owned"]
├── search
│   ├── description-only       search repos
│   ├── code-only              search code deduped
│   │   └── gh-export-repository-shape  gh --json repository (nameWithOwner, no owner.login)
│   └── union-both             OR merge matched_by
└── options
    ├── limit-default-30       Limit 0 → --limit 30
    └── multi-owner            two owners merged sorted
```

## Test Index

| Leaf | Description |
|------|-------------|
| `auth/not-authenticated` | Auth failure hints `gh auth login` |
| `auth/infer-owner` | Empty owners inferred from login |
| `plain/owned-only` | Plain mode `["owned"]` tags |
| `search/description-only` | Description search results |
| `search/code-only` | Code search deduped |
| `search/code-only/gh-export-repository-shape` | Real gh repository JSON (`nameWithOwner`) |
| `search/union-both` | Union merges `matched_by` |
| `options/limit-default-30` | Default limit 30 |
| `options/multi-owner` | Multi-owner merge |

## How to Run

```sh
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
	Owners            []string
	Limit             int
	SearchDescription string
	SearchCode        string
	GhBin             string
}

type Response struct {
	Results []ghrepos.RepoResult
	Login   string
	GhArgv  string
}

func captureGhArgv(ghBin string) string {
	if ghBin == "" {
		return ""
	}
	argvPath := filepath.Join(filepath.Dir(ghBin), "gh.argv")
	data, err := os.ReadFile(argvPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	opts := ghrepos.ListReposOptions{
		Owners:            req.Owners,
		SearchDescription: req.SearchDescription,
		SearchCode:        req.SearchCode,
		Limit:             req.Limit,
		GhBin:             req.GhBin,
	}
	results, err := ghrepos.ListRepos(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	resp := &Response{
		Results: results,
		GhArgv:  captureGhArgv(req.GhBin),
	}
	login, authErr := ghrepos.EnsureAuthenticated(context.Background(), req.GhBin)
	if authErr == nil {
		resp.Login = login
	}
	return resp, nil
}
```