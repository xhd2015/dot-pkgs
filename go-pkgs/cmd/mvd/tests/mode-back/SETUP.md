## Steps

## Steps
- All tests in this mode exercise `mvd --back`, which moves a project back to its previous location.
- The resolution chain is: unique root basename → alias → absolute path.
- For worktree-backed entries, the back operation removes the git worktree and branch (after verifying clean/merged status).

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("mode: back")
    return nil
}
```
