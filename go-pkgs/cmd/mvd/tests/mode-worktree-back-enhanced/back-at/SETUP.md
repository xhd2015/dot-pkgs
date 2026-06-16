## Steps
- All tests in this branch exercise `cmdWorktreeBackAt`: removing a worktree entry that is NOT the last in the chain.
- After removal, subsequent entries are preserved (spliced).
- The chain pattern is: [root, mid, wt, later] → `--back wt` → [root, mid, later].
- The enhanced CASE B / CASE C behaviors apply to this variant as well.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("variant: cmdWorktreeBackAt (non-last entry)")
    return nil
}
```
