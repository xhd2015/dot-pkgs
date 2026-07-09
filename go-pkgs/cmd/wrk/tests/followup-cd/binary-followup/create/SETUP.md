# Scenario

**Feature**: create mode follow-up behavior

```
wrk (create) + optional WRK_FOLLOWUP_FILE / --no-cd
  -> follow-up file empty or cd <new-worktree>
```

## Steps

1. Descendants seed main repo and create options.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "binary")
	return nil
}
```
