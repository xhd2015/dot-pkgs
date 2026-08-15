## Expected

- Exit code 0; stderr empty.
- Stdout has exactly two path-sorted lines:
  - `{abs brand-new-repo}\tmain`
  - `{abs known-repo}\tmain`
- Proves `--refresh` forced a cold full walk (warm would omit brand-new).

## Side Effects

- Force-refresh path discovers post-seed plants that warm soft incompleteness would miss.

## Exit Code

- `0`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}

	roots := rootsFromArgs(req.Args)
	if len(roots) != 1 {
		t.Fatalf("expected 1 --root, got %v", roots)
	}
	brandNew := absPath(t, filepath.Join(roots[0], "brand-new-repo"))
	known := absPath(t, filepath.Join(roots[0], "known-repo"))
	// Path-sorted: brand-new-repo before known-repo
	want := brandNew + "\tmain\n" + known + "\tmain\n"
	if resp.Stdout != want {
		t.Fatalf("stdout = %q, want %q", resp.Stdout, want)
	}

	// Extra guard: brand-new must appear (warm without --refresh would omit it).
	if !strings.Contains(resp.Stdout, brandNew) {
		t.Fatalf("--refresh must list brand-new-repo %q; stdout:\n%s", brandNew, resp.Stdout)
	}
}
```
