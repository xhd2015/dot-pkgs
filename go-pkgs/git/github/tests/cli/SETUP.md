# Scenario

**Feature**: `RunCLI` harness with mock `gh` for integration leaves

```
# CLI captures stdout/stderr while calling RunCLI
RunCLI args -> route subcommand -> ListRepos -> mock gh -> format output

# help and parse-error leaves skip gh
RunCLI args -> usage or error -> stdout/stderr
```

## Preconditions

- `RunCLI` is the single entry for `kool github` argv (no `github` token).
- Integration leaves install a shell mock at `req.GhBin`; `Run` sets `GH_BIN`.
- Help and CLI-parse error leaves do not require mock `gh`.

## Steps

1. Root `Setup` clears `GH_BIN` so help/error leaves do not invoke a real gh.
2. Descendant `Setup` sets `req.Args`, configures mock `gh`, or both.
3. `Run` (in `cli/DOCTEST.md`) captures I/O and calls `RunCLI`.

## Context

- Line output format: `{full_name}\t{matched_by}` per repo, one line each, sorted.
- JSON output uses `json.MarshalIndent` with two-space indent.
- Auth errors must surface `gh auth login` on stderr.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
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
	script := mockGhHeader + body
	if err := os.WriteFile(ghPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return ghPath
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

func writeEmptyOwnedGh(t *testing.T, owner string) string {
	t.Helper()
	body := authAliceBody + fmt.Sprintf(`if [ "$1" = "repo" ] && [ "$2" = "list" ] && [ "$3" = %q ]; then
  echo '[]'
  exit 0
fi
`, owner) + unexpectedGhBody
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
```