# Scenario

--dry-run with --vscode: prints intent, no code launched.

mvd --add tracked → [(tracked)]
mvd --dry-run --vscode tracked → prints 'would open VSCode'

## Steps
- Add a directory to history, then dry-run `--vscode`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	// First add
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run vscode
	req.Args = []string{"--dry-run", "--vscode", dir}
	return nil
}
```
