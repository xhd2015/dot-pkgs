## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` is non-empty, ends with `\n`, and mentions `repo`.
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation (`req.GhBin` unset).

## Errors

- `err` from harness is nil.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHelpStdout(t, resp, "repo")
}
```
