# Scenario

**Feature**: `RunCLI` help paths print usage to stdout without calling `ListRepos`

```
# top-level help (empty args, --help, -h, help)
RunCLI [] OR --help -> top-level usage (mentions repo) -> stdout, exit 0

# repo-level help (repo alone or repo --help)
RunCLI repo OR repo --help -> repo usage (list; points to repo list --help) -> stdout, exit 0

# list leaf help
RunCLI repo list --help -> list usage with flags -> stdout, exit 0
```

## Preconditions

- Help leaves do not set `req.GhBin`; no `gh` invocation expected.

## Steps

1. Descendant `Setup` sets `req.Args` to the help argv for that level.

## Context

- Empty top-level args behave like `--help` (exit 0, top-level usage).
- Top-level help lists `repo` among available commands.
- Repo-level help mentions `list` and that `repo list --help` shows list options;
  it is not identical to list leaf help (no requirement to list all list flags).
- `repo list --help` documents list flags (`--owner`, `--json`, search flags).
- Help stdout ends with trailing `\n`.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GhBin = ""
	return nil
}

func assertHelpStdout(t *testing.T, resp *Response, mustContain ...string) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatal("expected non-empty stdout usage")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with trailing newline, got %q", resp.Stdout)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stdout)
	for _, want := range mustContain {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("expected %q in usage, got stdout=%q", want, resp.Stdout)
		}
	}
}

func assertNotContainsFold(t *testing.T, stdout, needle string) {
	t.Helper()
	if strings.Contains(strings.ToLower(stdout), strings.ToLower(needle)) {
		t.Fatalf("did not expect %q in repo-level help, got stdout=%q", needle, stdout)
	}
}
```
