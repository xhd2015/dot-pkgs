## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <root short>  main ahead commit
Status:       clean

Dir:          wt-linked
Branch:       wt-side
Commit:       <wt short>  status main root
Status:       clean
Compare with Master: main is newer(wt-side +1 commit -> main)
                     to fast forward, on wt-side: 
                        git merge --ff-only  main
```

## Expected

- Exit code 0.
- Stdout contains two status blocks: root `.` and linked worktree `wt-linked`.
- Root block has **no** `Compare with Master` line.
- Linked worktree block includes `Compare with Master:` showing main branch is newer than the worktree branch (kool multi-line format).
- Stderr is empty.

## Side Effects

- No repository files are changed.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	rootBlock := statusBlockTemplate(t, req.MainRepo, ".", "clean")
	assert.Output(t, resp.Stdout, rootBlock)

	compare := compareWithMasterField(t, req.MainRepo, "main", req.WtBranch)
	linkedBlock := statusBlockWithCompare(t, req.WtDir, "wt-linked", "clean", compare)
	assert.Output(t, resp.Stdout, linkedBlock)
}
```