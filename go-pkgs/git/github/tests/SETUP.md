# Scenario

**Feature**: shared helpers for git/github doctest harness

```
# doctest harness prepares temp dirs and reads captured gh argv
test runner -> Setup chain -> Run(ListOwned | NormalizeRepoURL) -> Assert
```

## Preconditions

- Target package path is `github.com/xhd2015/dot-pkgs/go-pkgs/git/github`.
- Tests must not require a real `gh` binary or network access.

## Steps

1. Descendant `Setup` functions configure `Request` fields and mock `gh` scripts.
2. Root `Run` (in `DOCTEST.md`) executes `ListOwned` or parse helpers.

## Context

- `DOCTEST_ROOT` points at this `tests/` directory.
- `DOCTEST_SESSION_ID` is available for session-scoped temp state if needed.
- List-owned leaves inherit mock-gh helpers from `list-owned/SETUP.md`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ghDir(ghBin string) string {
	return filepath.Dir(ghBin)
}

func ghArgvPath(ghBin string) string {
	return filepath.Join(ghDir(ghBin), "gh.argv")
}

func ghCalledPath(ghBin string) string {
	return filepath.Join(ghDir(ghBin), "gh.called")
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

func ghWasCalled(ghBin string) bool {
	_, err := os.Stat(ghCalledPath(ghBin))
	return err == nil
}

func assertSortedFullNames(t *testing.T, names []string) {
	t.Helper()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Fatalf("names not sorted: %q before %q", names[i-1], names[i])
		}
	}
}
```