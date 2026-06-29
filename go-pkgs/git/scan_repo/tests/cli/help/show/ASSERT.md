## Expected

- Exit code 0.
- Stdout contains usage text documenting `--root`, `--max-depth`, `--ignore-dir`,
  `--ignore-dir-basename`, `-v` / `--verbose`, `--list-remotes`, `--list-worktrees`,
  and `--json`.
- Stderr is empty.

## Side Effects

- No filesystem scan occurs.

## Exit Code

- `0`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	for _, flag := range []string{
		"--root", "--max-depth", "--ignore-dir", "--ignore-dir-basename",
		"-v", "--verbose", "--list-remotes", "--list-worktrees", "--json",
	} {
		if !strings.Contains(resp.Stdout, flag) {
			t.Fatalf("help should mention %q, got:\n%s", flag, resp.Stdout)
		}
	}
}
```