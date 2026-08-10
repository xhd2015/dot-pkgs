## Expected

- `IsClean` returns nil (`IsCleanNil=true`, empty `IsCleanErr`)
- `IsCleanWrk` is true
- Run itself returns no error

## Side Effects

- None

## Errors

- None expected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
	if !resp.IsCleanNil {
		t.Fatalf("IsClean = error %q, want nil on clean repo", resp.IsCleanErr)
	}
	if resp.IsCleanErr != "" {
		t.Fatalf("IsCleanErr = %q, want empty", resp.IsCleanErr)
	}
	if !resp.IsCleanWrk {
		t.Fatal("IsCleanWrk = false, want true on clean repo")
	}
}
```
