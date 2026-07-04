## Expected

- The command exits with an error.
- The output contains the matched file path `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md`.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for subdirectory REQUIREMENT-*.md match, got 0\n%s", resp.Output)
	}
	want := "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md"
	if !strings.Contains(resp.Output, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, resp.Output)
	}
}
```