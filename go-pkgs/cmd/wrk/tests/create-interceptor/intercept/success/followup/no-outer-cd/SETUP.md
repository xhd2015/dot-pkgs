# Scenario

**Feature**: outer wrk writes no follow-up `cd` when intercepting create

```
# home gate would apply if native: cwd=FakeHome + WRK_FOLLOWUP_FILE
FakeHome (cwd) + WRK_FOLLOWUP_FILE
wrk <mainRepo> + intercept enabled
  -> fake runs; follow-up file has no "cd …" line from outer wrk
```

## Steps

1. Grouping already has `myrepo`, enabled config, and fake `kool`.
2. Set `FakeHome`, export `WRK_FOLLOWUP_FILE`, process cwd = FakeHome.
3. Create via positional `wrk <mainRepo>` so shell cwd remains FakeHome (home gate would open on native path).

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	// MainRepo already set by intercept Setup; re-home shell cwd for home gate.
	req.FakeHome = filepath.Join(req.WorkRoot, "home")
	mkdirAll(t, req.FakeHome)
	req.FollowupFile = filepath.Join(req.WorkRoot, "followup.txt")
	req.UseFollowupEnv = true
	req.RepoDir = req.FakeHome
	req.TargetDir = req.MainRepo
	return nil
}
```
