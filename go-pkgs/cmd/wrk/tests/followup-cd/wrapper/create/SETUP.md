# Scenario

**Feature**: wrapper create auto-cd controls (home-gated)

```
# StartDir = FakeHome (user home) + auto-cd on
source bash.sh; wrk <mainRepo> -> stderr "cd <wt>"; FinalPWD = worktree

# StartDir = main repo (not home)
source bash.sh; wrk -> no stderr cd; FinalPWD stays main

# WRK_AUTO_CD=0 / --no-cd suppress regardless of home
```

## Steps

1. Descendants seed main repo and choose StartDir (FakeHome vs mainRepo).

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "wrapper")
	return nil
}
```
