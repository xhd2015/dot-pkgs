## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` mentions `list` and documents `--owner` and `--json`.
- `resp.Stdout` ends with trailing `\n`.
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation.

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
	assertHelpStdout(t, resp, "list", "--owner", "--json")
}
```
