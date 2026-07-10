# Scenario

**Feature**: create mode follow-up behavior (home-gated)

```
# shell cwd == user home (FakeHome) + WRK_FOLLOWUP_FILE
wrk <mainRepo> -> follow-up: cd <new-worktree>

# shell cwd == main repo (not home) + WRK_FOLLOWUP_FILE
wrk -> worktree created; follow-up empty

# --no-cd or unset env always suppress write
```

## Steps

1. Descendants seed main repo and choose shell cwd (FakeHome vs mainRepo).

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "binary")
	return nil
}
```
