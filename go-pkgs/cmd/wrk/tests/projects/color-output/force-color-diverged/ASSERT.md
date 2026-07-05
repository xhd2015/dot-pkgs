## Expected

- Exit code 0.
- `Remote:` value `Needs Merge(2 commits diverged)` is wrapped in red ANSI.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"fmt"

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
	assertColorProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := colorCompareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	if remote != "Remote:       Needs Merge(2 commits diverged)" {
		t.Fatalf("Remote: want Needs Merge(2 commits diverged), got %q", remote)
	}

	block := fmt.Sprintf(`---
version: 2
---
%s
%s
%s
Status:       clean
Remote:       <ansi-color red>Needs Merge(2 commits diverged)</ansi-color>
Worktrees:    0 total, 0 dirty
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo))
	assert.Output(t, resp.Stdout, block)
}
```