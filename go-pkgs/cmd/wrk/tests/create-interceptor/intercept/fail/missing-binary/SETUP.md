# Scenario

**Feature**: interceptor binary missing from PATH fails hard

```
argv[0]=wrk-interceptor-tool-not-installed (not on PATH)
  -> non-zero; no worktree; no silent native create
```

## Steps

1. Override config to use a unique argv[0] that is not installed.
2. Clear fake-tool PATH install from grouping (empty prepend dir).
3. Run bare create.

```go
import (
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	// Undo intercept grouping's fake install.
	req.PathPrepend = filepath.Join(req.WorkRoot, "empty-bin")
	mkdirAll(t, req.PathPrepend)
	writeInterceptorConfig(t, req.WrkHome, true, []string{
		"wrk-interceptor-tool-not-installed", "--mark",
	}, nil)
	req.InterceptorLog = filepath.Join(req.WorkRoot, "interceptor-argv.log")
	writeFile(t, req.InterceptorLog, "")
	// Strip ExtraEnv fake log/exit from installFakeInterceptor.
	var cleaned []string
	for _, e := range req.ExtraEnv {
		if strings.HasPrefix(e, envFakeInterceptorLog+"=") || strings.HasPrefix(e, envFakeInterceptorExit+"=") {
			continue
		}
		cleaned = append(cleaned, e)
	}
	req.ExtraEnv = cleaned
	return nil
}
```
