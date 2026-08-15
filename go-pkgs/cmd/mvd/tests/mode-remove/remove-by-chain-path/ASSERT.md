## Expected
- `--rm` succeeds without `-f`, removing only the specified path from the chain.
- The output contains "removed:".
- The history file still has the root entry with a single location.

## Exit Code
- 0 (success)

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		t.Fatalf("expected response, got error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "removed:")

	repo := filepath.Join(req.WorkRoot, "repo")
	h := assertHistoryLen(t, req.ConfigHome, 1)
	proj := h.Projects[repo]
	if len(proj.Locations) != 1 {
		t.Fatalf("expected 1 location (root only), got %d: %#v", len(proj.Locations), proj.Locations)
	}
	if proj.Locations[0].Path != repo {
		t.Fatalf("expected root %s, got %s", repo, proj.Locations[0].Path)
	}
}
```
