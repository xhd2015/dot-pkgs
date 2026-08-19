# Scenario

**Feature**: mock `gh` executable for `ListOwned` integration tests

```
# per-owner gh invocation
ListOwned -> exec gh repo list <owner> --json ... -> JSON stdout

# doctest replaces gh with shell script
ListOwned -> mock gh script -> canned JSON or error exit
```

## Preconditions

- Each leaf writes an executable shell script to `t.TempDir()/gh`.
- Mock scripts record `"$*"` to `gh.argv` and touch `gh.called` on invocation.

## Steps

1. Leaf `Setup` calls `writeFakeGh` with a script body tailored to the scenario.
2. Set `req.GhBin` to the mock script path.
3. `Run` invokes `ListOwned`, which execs the mock.

## Context

- Fixture JSON for success cases lives in leaf `testdata/` directories.
- Options leaves assert captured argv via `resp.GhArgv` or `readGhArgv`.
- Validation leaves use a mock that fails if ever invoked.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Limit == 0 {
		req.Limit = 100
	}
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

func writeFakeGhFromFixture(t *testing.T, d *session.Doctest, fixtureRelPath string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureFile(d, fixtureRelPath))
	if err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  cat <<'EOF'
%s
EOF
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`, string(data))
	return writeFakeGh(t, body)
}

func writeTrapGh(t *testing.T) string {
	t.Helper()
	return writeFakeGh(t, `echo "gh should not have been called: $*" >&2
exit 99
`)
}

func writeOwnerFixtureGh(t *testing.T, d *session.Doctest, ownerToFixture map[string]string) string {
	t.Helper()
	body := ""
	for owner, fixture := range ownerToFixture {
		data, err := os.ReadFile(fixtureFile(d, fixture))
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
	body += `echo "unexpected args: $*" >&2
exit 1
`
	return writeFakeGh(t, body)
}```