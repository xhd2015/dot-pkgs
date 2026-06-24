## Expected

- `resp` is nil.
- Error message contains `gh auth login`.
- No repo results returned.

## Side Effects

- Mock `gh` invoked for `api user` only; `repo list` and search subcommands never run.

## Errors

- `err` is non-nil.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected authentication error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "gh auth login") {
		t.Fatalf("expected gh auth login hint, got %v", err)
	}
	if resp != nil && len(resp.Results) > 0 {
		t.Fatalf("expected no results on auth failure, got %+v", resp.Results)
	}
	argv := readGhArgv(t, req.GhBin)
	if strings.Contains(argv, "repo list") {
		t.Fatalf("repo list should not run before auth, argv=%q", argv)
	}
}
```