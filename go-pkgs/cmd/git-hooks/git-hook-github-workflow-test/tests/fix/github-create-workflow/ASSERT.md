## Expected

- The command exits successfully.
- `.github/workflows/test.yml` is created.
- The workflow uses container image `golang:1.22`.
- The workflow runs `go test -v ./...`.
- The workflow runs `doctest test -v --label-all ./...`.
- The Go test step appears before the doctest step.

## Exit Code

- Exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if !resp.WorkflowExists {
		t.Fatalf("expected workflow to be created at %s", resp.WorkflowPath)
	}
	content := resp.WorkflowContent
	for _, want := range []string{"container:", "golang:1.22", "go test -v ./...", "doctest test -v --label-all ./..."} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, content)
		}
	}
	goTest := strings.Index(content, "go test -v ./...")
	doctest := strings.Index(content, "doctest test -v --label-all ./...")
	if goTest < 0 || doctest < 0 || goTest > doctest {
		t.Fatalf("expected go test before doctest, got:\n%s", content)
	}
}
```
