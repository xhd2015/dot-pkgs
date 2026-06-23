## Expected

- `resp.Display` is `.codex/hooks/agent-sessions-stop.sh` (platform-native separators).
- Display must not equal the absolute input path.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(".codex", "hooks", "agent-sessions-stop.sh")
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (cwd=%q path=%q)", want, resp.Display, resp.Cwd, req.Path)
	}
	if resp.Display == req.Path {
		t.Fatalf("display must shorten missing leaf path, got absolute %q", resp.Display)
	}
}
```