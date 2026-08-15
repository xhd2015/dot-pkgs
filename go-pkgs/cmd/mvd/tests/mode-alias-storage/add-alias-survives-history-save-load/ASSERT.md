## Expected
- `--which mp` resolves the alias to the project path.
- `aliases.json` does NOT exist.
- The alias "mp" is present in history.json's project entry after the save/load cycle.

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
	found := false
	for _, a := range proj.Aliases {
		if a == "mp" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("alias 'mp' lost after history save/load, got: %v", proj.Aliases)
	}
}
```
