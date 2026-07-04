# Scenario

**Feature**: replace points outside the scanning worktree (sibling checkout)

```
# worktree B go.mod has replace => <worktree A>/external/mydep
# target exists but is outside B's worktree directory tree — must block
# reproduces stale abs path left from another wrk checkout after wrk --done
worktree B + replace => sibling worktree A/external/mydep -> extra-repo issue
```

## Preconditions

- Two worktrees: `consumer-a` hosts the external linked worktree; `consumer-b` has the stale replace.
- The replace target path exists on disk but lies outside `consumer-b`'s worktree root.

## Steps

1. Create `consumer-a` with linked external worktree.
2. Create `consumer-b` with absolute replace pointing at `consumer-a`'s external path.
3. Call `replace.CheckLocalReplaces(consumer-b)`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	consumerA, externalPath := setupWrkExternalConsumer(t, req.RootDir)
	consumerB := filepath.Join(req.RootDir, "consumer-b")
	mkdirWrkExternal(t, consumerB)
	runWrkExternalGit(t, consumerB, "init")
	writeConsumerReplace(t, consumerB, externalPath)
	req.RootDir = consumerB
	_ = consumerA
	return nil
}

```