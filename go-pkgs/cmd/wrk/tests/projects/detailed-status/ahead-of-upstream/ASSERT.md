## Expected

- Exit code 0.
- `Compare with Remote: main is newer(origin/main +1 commit -> main)`.
- `Worktrees: 0 Clean, 0 Dirty`.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	if !strings.Contains(remote, "main is newer(origin/main +1 commit -> main)") {
		t.Fatalf("Compare with Remote: want ahead message, got %q", remote)
	}
	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", remote, "0 Clean, 0 Dirty")
	assert.Output(t, resp.Stdout, block)
}
```