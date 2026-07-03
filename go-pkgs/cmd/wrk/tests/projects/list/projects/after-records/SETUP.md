# Scenario

**Feature**: wrk --projects after multiple auto-records

```
wrk --list (repoA) + wrk --list (repoB) -> wrk --projects -> sorted paths
```

## Steps

1. Create two git repos `aaa` and `zzz` under `{WorkRoot}`.
2. Run `wrk --list` from each to auto-record.
3. Run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	repoA := initProjectsRepo(t, req.WorkRoot, "aaa")
	repoB := initProjectsRepo(t, req.WorkRoot, "zzz")
	runWrkWithArgs(t, req, repoA, "--list")
	runWrkWithArgs(t, req, repoB, "--list")
	req.SecondRepo = repoB
	req.MainRepo = repoA
	return nil
}
```