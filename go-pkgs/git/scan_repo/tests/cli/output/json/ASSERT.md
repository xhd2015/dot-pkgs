## Expected

- Exit code 0.
- Stdout is valid JSON array with two elements.
- Elements are path-sorted (`alpha` before `zebra`).
- Each `RepoType` is the string `"main"` (not a numeric enum).

## Exit Code

- `0`.

```go
import (
	"encoding/json"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

type jsonRepo struct {
	Path     string `json:"Path"`
	Name     string `json:"Name"`
	RepoType string `json:"RepoType"`
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	var repos []jsonRepo
	if err := json.Unmarshal([]byte(resp.Stdout), &repos); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout:\n%s", err, resp.Stdout)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos in JSON, got %d: %s", len(repos), resp.Stdout)
	}
	roots := rootsFromArgs(req.Args)
	wantAlpha := absPath(t, filepath.Join(roots[0], "alpha"))
	wantZebra := absPath(t, filepath.Join(roots[0], "zebra"))
	if repos[0].Path != wantAlpha {
		t.Fatalf("repos[0].Path = %q, want %q", repos[0].Path, wantAlpha)
	}
	if repos[1].Path != wantZebra {
		t.Fatalf("repos[1].Path = %q, want %q", repos[1].Path, wantZebra)
	}
	for i, r := range repos {
		if r.RepoType != "main" {
			t.Fatalf("repos[%d].RepoType = %q, want main", i, r.RepoType)
		}
	}
}
```