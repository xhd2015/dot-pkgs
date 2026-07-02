## Expected

- Exit code 0.
- Stdout contains `wrk example.com/dep1 at ./external/mydep1-main-2026-06-30` and `wrk example.com/dep2 at ./external/mydep2-main-2026-06-30`.
- Stdout ends with `wrk 2 deps`.
- Consumer sub-module `go-pkgs/go.mod` has replaces for both deps.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	stdout := strings.TrimSpace(resp.Stdout)

	assertContains(t, stdout, "wrk example.com/dep1 at ./external/mydep1-main-"+wrkDate)
	assertContains(t, stdout, "wrk example.com/dep2 at ./external/mydep2-main-"+wrkDate)
	assertContains(t, stdout, "wrk 2 deps")

	// Both replaces must be in go-pkgs/go.mod (not a root go.mod).
	modDir := filepath.Join(req.ConsumerTop, "go-pkgs")
	mod, err := allDepsReadGoMod(modDir)
	if err != nil {
		t.Fatalf("read go-pkgs/go.mod: %v", err)
	}

	wantPath1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantPath2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantPath1) {
		t.Fatalf("go-pkgs/go.mod missing replace example.com/dep1 => %s: %+v", wantPath1, mod.Replace)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantPath2) {
		t.Fatalf("go-pkgs/go.mod missing replace example.com/dep2 => %s: %+v", wantPath2, mod.Replace)
	}

	ok, err := allDepsGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```