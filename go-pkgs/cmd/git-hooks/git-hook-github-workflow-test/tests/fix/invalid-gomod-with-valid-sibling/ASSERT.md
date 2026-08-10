## Expected

- The command exits successfully.
- The output warns that nested `kool-template/go.mod` is missing a `go` directive.
- `.github/workflows/test.yml` is created using the valid root module version.
- The workflow uses container image `golang:1.22`.
- The workflow runs root `go test -v ./...` and does not add a step for `kool-template`.

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
	if !strings.Contains(resp.Output, "warning:") || !strings.Contains(resp.Output, "missing go directive") {
		t.Fatalf("expected missing go directive warning, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "kool-template") {
		t.Fatalf("expected warning to mention kool-template, got:\n%s", resp.Output)
	}
	if !resp.WorkflowExists {
		t.Fatalf("expected workflow to be created at %s", resp.WorkflowPath)
	}
	content := resp.WorkflowContent
	for _, want := range []string{"container:", "golang:1.22", "go test -v ./..."} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, content)
		}
	}
	if strings.Contains(content, "kool-template") {
		t.Fatalf("workflow must not include skipped invalid module, got:\n%s", content)
	}
}
```
