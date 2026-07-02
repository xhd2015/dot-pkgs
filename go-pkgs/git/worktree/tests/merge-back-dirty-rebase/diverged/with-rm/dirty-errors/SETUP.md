# Scenario

**Feature**: diverged + dirty + --rm → still errors "worktree not clean"

```
# Remove=true with dirty worktree → IsClean guard still applies
dirty feat -> MergeBack(Remove=true) -> IsClean fails -> error
```

## Steps

1. Make feature worktree dirty.
2. Call MergeBack with `Remove=true`.

```go
func Setup(t *testing.T, req *Request) error {
	makeDirty(t, req.SourcePath)
	return nil
}
```
