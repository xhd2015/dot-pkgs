# Scenario

**Feature**: `--exec` after native create runs the command in the new worktree

```
# create mode + --exec
myrepo (main) -> wrk --no-interceptor --exec <cmd>
  -> native worktree under WRK_HOME
  -> stdout: <wt-abs>\n + child stdout
  -> child cmd.Dir = wt-abs

# interceptor enabled without escape + --exec → error (no silent intercept)
```

## Preconditions

- No interceptor config unless a leaf writes one.
- Success leaves pass `--no-interceptor` so create is always native.

## Steps

- Default: init `myrepo` on main; `req.RepoDir` = main checkout.
- Leaves set `req.Args` including `--no-interceptor` and `--exec …` (or error forms).

```go
import (
	"encoding/json"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	req.RepoDir = repoDir
	req.MainRepo = repoDir
	ensureExecCreateHelpersUsed()
	return nil
}

type execInterceptorConfigFile struct {
	Version int `json:"version"`
	Create  *struct {
		Interceptor *execInterceptorSpec `json:"interceptor,omitempty"`
	} `json:"create,omitempty"`
}

type execInterceptorSpec struct {
	Enabled bool     `json:"enabled"`
	Argv    []string `json:"argv"`
}

// writeEnabledExecInterceptor writes a minimal enabled create.interceptor block.
// Used by interceptor-blocked; argv is a no-op binary so config is valid if checked.
func writeEnabledExecInterceptor(t *testing.T, wrkHome string) {
	t.Helper()
	cfg := execInterceptorConfigFile{Version: 1}
	cfg.Create = &struct {
		Interceptor *execInterceptorSpec `json:"interceptor,omitempty"`
	}{
		Interceptor: &execInterceptorSpec{
			Enabled: true,
			Argv:    []string{"true"},
		},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal interceptor config: %v", err)
	}
	data = append(data, '\n')
	writeFile(t, filepath.Join(wrkHome, "config.json"), string(data))
}

func ensureExecCreateHelpersUsed() {
	_ = writeEnabledExecInterceptor
}
```
