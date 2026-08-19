# Scenario

**Feature**: `ListRepos` entry point with auth, search union, and `matched_by`

```
# auth gate then owner resolution
ListRepos -> EnsureAuthenticated -> gh api user -> login

# plain mode: owned repos only
ListRepos no search -> ListOwned per owner -> matched_by ["owned"]

# search modes union by FullName
ListRepos search -> gh search repos/code per owner -> merge matched_by
```

## Preconditions

- `ListRepos` requires authenticated `gh` before any repo query.
- Mock `gh` scripts handle `api user`, `repo list`, `search repos`, and `search code`.
## Steps

1. Descendant `Setup` configures search keywords, owners, and mock `gh` behavior.
2. Nested root `Run` (in `list-repos/DOCTEST.md`) calls `ListRepos`.

## Context

- Default limit for `ListRepos` is 30 when `Limit` is 0 (distinct from `ListOwned` default 100).
- `matched_by` values are `owned`, `description`, or `code`; union merges preserve all reasons.
- This nested root is self-contained: helpers below include argv readers and mock-gh utilities.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	ghrepos "github.com/xhd2015/dot-pkgs/go-pkgs/git/github"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}

func fixtureFile(d *session.Doctest, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	base := d.DOCTEST_CASE
	if base == "" || !filepath.IsAbs(base) {
		base = filepath.Join(d.DOCTEST_ROOT, base)
	}
	return filepath.Join(base, rel)
}

func ghArgvPath(ghBin string) string {
	return filepath.Join(filepath.Dir(ghBin), "gh.argv")
}

func readGhArgv(t *testing.T, ghBin string) string {
	t.Helper()
	data, err := os.ReadFile(ghArgvPath(ghBin))
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatal(err)
	}
	return strings.TrimSpace(string(data))
}

func assertSortedFullNames(t *testing.T, names []string) {
	t.Helper()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Fatalf("names not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

const mockGhHeader = `#!/bin/sh
set -eu
dir=$(dirname "$0")
echo "$*" >> "$dir/gh.argv"
: > "$dir/gh.called"
`

func writeFakeGh(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	ghPath := filepath.Join(dir, "gh")
	tmpPath := ghPath + ".tmp"
	script := mockGhHeader + body
	// Write then rename so Linux overlayfs does not ETXTBSY on immediate exec.
	if err := os.WriteFile(tmpPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpPath, ghPath); err != nil {
		t.Fatal(err)
	}
	return ghPath
}

const authAliceBody = `if [ "$1" = "api" ] && [ "$2" = "user" ]; then
  echo '{"login":"alice"}'
  exit 0
fi
`

const unexpectedGhBody = `echo "unexpected args: $*" >&2
exit 1
`

func writeAuthFailGh(t *testing.T) string {
	t.Helper()
	body := `if [ "$1" = "api" ] && [ "$2" = "user" ]; then
  echo 'To authenticate, please run gh auth login' >&2
  exit 4
fi
` + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeInferOwnerGh(t *testing.T, fixtureRelPath string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureRelPath)
	if err != nil {
		t.Fatal(err)
	}
	body := authAliceBody + fmt.Sprintf(`if [ "$1" = "repo" ] && [ "$2" = "list" ] && [ "$3" = "alice" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, string(data)) + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeOwnedOnlyGh(t *testing.T, owner, fixtureRelPath string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureRelPath)
	if err != nil {
		t.Fatal(err)
	}
	body := authAliceBody + fmt.Sprintf(`if [ "$1" = "repo" ] && [ "$2" = "list" ] && [ "$3" = %q ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, owner, string(data)) + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeSearchReposGh(t *testing.T, fixtureRelPath string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureRelPath)
	if err != nil {
		t.Fatal(err)
	}
	body := authAliceBody + fmt.Sprintf(`if [ "$1" = "search" ] && [ "$2" = "repos" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, string(data)) + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeSearchCodeGh(t *testing.T, fixtureRelPath string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureRelPath)
	if err != nil {
		t.Fatal(err)
	}
	body := authAliceBody + fmt.Sprintf(`if [ "$1" = "search" ] && [ "$2" = "code" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, string(data)) + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeUnionSearchGh(t *testing.T, reposFixture, codeFixture string) string {
	t.Helper()
	reposData, err := os.ReadFile(reposFixture)
	if err != nil {
		t.Fatal(err)
	}
	codeData, err := os.ReadFile(codeFixture)
	if err != nil {
		t.Fatal(err)
	}
	body := authAliceBody +
		fmt.Sprintf(`if [ "$1" = "search" ] && [ "$2" = "repos" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, string(reposData)) +
		fmt.Sprintf(`if [ "$1" = "search" ] && [ "$2" = "code" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, string(codeData)) + unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeMultiOwnerGh(t *testing.T, ownerToFixture map[string]string) string {
	t.Helper()
	body := authAliceBody
	for owner, fixture := range ownerToFixture {
		data, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		body += fmt.Sprintf(`if [ "$1" = "repo" ] && [ "$2" = "list" ] && [ "$3" = %q ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
`, owner, string(data))
	}
	body += unexpectedGhBody
	return writeFakeGh(t, body)
}

func writeLimitDefaultGh(t *testing.T) string {
	t.Helper()
	body := authAliceBody + `if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  echo '[]'
  exit 0
fi
` + unexpectedGhBody
	return writeFakeGh(t, body)
}

func assertMatchedBy(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("matched_by len %d, want %d: got %v want %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("matched_by[%d]=%q, want %q (full: got %v want %v)", i, got[i], want[i], got, want)
		}
	}
}

func resultByFullName(t *testing.T, results []ghrepos.RepoResult, fullName string) ghrepos.RepoResult {
	t.Helper()
	for _, r := range results {
		if r.FullName == fullName {
			return r
		}
	}
	t.Fatalf("no result for %q in %+v", fullName, results)
	return ghrepos.RepoResult{}
}

func matchedByStrings(reasons []ghrepos.MatchReason) []string {
	out := make([]string, len(reasons))
	for i, r := range reasons {
		out[i] = string(r)
	}
	return out
}
```