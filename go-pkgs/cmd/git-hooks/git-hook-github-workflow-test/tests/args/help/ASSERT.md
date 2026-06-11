## Expected

- The command exits successfully.
- The output includes usage text.
- The output documents `--fix` and `--origin-domain`.

## Exit Code

- Exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	for _, want := range []string{"Usage: git-hook-github-workflow-test", "--fix", "--origin-domain"} {
		if !strings.Contains(resp.Output, want) {
			t.Fatalf("expected help to contain %q, got:\n%s", want, resp.Output)
		}
	}
}
```
