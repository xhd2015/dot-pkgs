## Expected
- `--list` shows both aliases ("mp", "myproj-alias") for the project.
- `aliases.json` does NOT exist.
- `history.json` contains the project with both aliases in its `aliases` array.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	assertContains(t, resp.Output, "mp")
	assertContains(t, resp.Output, "myproj-alias")

	aliasesFile := filepath.Join(req.ConfigHome, "aliases.json")
	assertFileNotExists(t, aliasesFile)

	h := readHistoryFile(t, req.ConfigHome)
	if h == nil {
		t.Fatalf("expected history.json")
	}
	projDir := filepath.Join(req.WorkRoot, "projects", "myproj")
	proj, ok := h.Projects[projDir]
	if !ok {
		t.Fatalf("expected project %s in history", projDir)
	}
	if len(proj.Aliases) < 2 {
		t.Fatalf("expected at least 2 aliases, got %d: %v", len(proj.Aliases), proj.Aliases)
	}
	hasMp := false
	hasMyprojAlias := false
	for _, a := range proj.Aliases {
		if a == "mp" {
			hasMp = true
		}
		if a == "myproj-alias" {
			hasMyprojAlias = true
		}
	}
	if !hasMp {
		t.Fatalf("expected alias 'mp' in project aliases, got %v", proj.Aliases)
	}
	if !hasMyprojAlias {
		t.Fatalf("expected alias 'myproj-alias' in project aliases, got %v", proj.Aliases)
	}
}
```
