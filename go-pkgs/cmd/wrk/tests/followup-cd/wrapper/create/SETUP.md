# Scenario

**Feature**: wrapper create auto-cd controls

```
source bash.sh; wrk [--no-cd]  # optional WRK_AUTO_CD=0
  -> maybe cd into new worktree
```

## Steps

1. Descendants seed main repo as StartDir/RepoDir.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "wrapper")
	return nil
}
```
