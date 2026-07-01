## Expected

- Exit code 0 (intra-repo replace is tolerated under the default lenient guard).
- Stderr names the offending go.mod file: `<wtDir>/go.mod`.
- Stderr names the offending directive: `replace example.com/foo => ./submod`.
- Stderr marks it as tolerated (contains `tolerated`).
- Merge-back proceeded: the consumer linked worktree is removed and no longer
  registered.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (intra-repo warn + proceed), got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, filepath.Join(req.WtDir, "go.mod"))
	assertContains(t, resp.Stderr, "replace example.com/foo => ./submod")
	assertContains(t, resp.Stderr, "tolerated")

	// Merge-back ran and removed the worktree.
	assertFileNotExists(t, req.WtDir)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```
