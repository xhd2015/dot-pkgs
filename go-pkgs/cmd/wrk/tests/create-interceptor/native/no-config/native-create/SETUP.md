# Scenario

**Feature**: native create when config.json is missing

```
# no config.json under WRK_HOME
myrepo (main) -> wrk -> {WRK_HOME}/worktrees/myrepo-main-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main`.
2. Run bare `wrk` with cwd = `myrepo` (create mode).
3. Do not write `config.json`; do not install a fake interceptor.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForInterceptor(t, req)
	return nil
}
```
