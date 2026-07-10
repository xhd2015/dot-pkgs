# Scenario

**Feature**: create interceptor + `--exec` errors without escape hatch

```
config create.interceptor enabled
myrepo -> wrk --exec true
  -> non-zero; stderr mentions interceptor and/or --no-interceptor
  -> no successful native+exec path (fail before or instead of mixed intercept+exec)
```

## Steps

1. Write enabled interceptor config under `WRK_HOME`.
2. Initialize `myrepo`.
3. Run `wrk --exec true` **without** `--no-interceptor` / `WRK_NO_INTERCEPTOR`.

```go
func Setup(t *testing.T, req *Request) error {
	writeEnabledExecInterceptor(t, req.WrkHome)
	initGitRepoOnMain(t, req.RepoDir)
	req.Args = []string{"--exec", "true"}
	return nil
}
```
