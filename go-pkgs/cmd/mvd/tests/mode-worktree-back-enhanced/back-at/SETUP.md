# Scenario

**Feature**: cmdWorktreeBackAt splices non-last worktree under CASE B/C auto-yes

```
# chain [root, mid, wt, later]; --back wt auto-yes merge/rebase → [root, mid, later]
mvd --back wt → auto-yes plan confirm → remove wt + splice later entries
```

## Steps
- All tests in this branch exercise `cmdWorktreeBackAt`: removing a worktree entry that is NOT the last in the chain.
- After removal, subsequent entries are preserved (spliced).
- The chain pattern is: [root, mid, wt, later] → `--back wt` → [root, mid, later].
- The enhanced CASE B / CASE C behaviors apply; default auto-yes (no `--confirm` required for success paths).

```go
func Setup(t *testing.T, req *Request) error {
	t.Logf("variant: cmdWorktreeBackAt (non-last entry)")
	return nil
}
```
