# Scenario

**Feature**: `TargetLabel` follows the branch checked out at `TargetPath`

```
# target may be main default branch or another linked worktree
ReadBranch(targetAbs) -> prompt question + fast-forward comment
```

## Context

Leaves under this node override `DefaultBranch`, `TargetPath`, and fixture layout.

```go
func Setup(t *testing.T, req *Request) error {
	req.CapturePrompt = true
	req.DryRun = false
	return nil
}
```