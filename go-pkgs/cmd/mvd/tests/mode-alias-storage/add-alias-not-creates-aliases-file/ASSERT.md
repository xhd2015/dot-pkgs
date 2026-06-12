## Expected
- `--which mp` resolves the alias successfully.
- `aliases.json` does NOT exist in the config directory.
- `history.json` contains the project with alias "mp" in its `aliases` array.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	aliasesFile := filepath.Join(req.ConfigHome, "aliases.json")
	assertFileNotExists(t, aliasesFile)

	h := readHistoryFile(t, req.ConfigHome)
	if h == nil || len(h.Projects) == 0 {
		t.Fatalf("expected history.json with projects")
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
		t.Fatalf("expected alias 'mp' in project aliases, got %v", proj.Aliases)
	}
}
```
