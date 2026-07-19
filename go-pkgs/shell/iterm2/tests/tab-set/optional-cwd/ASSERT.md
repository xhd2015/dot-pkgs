## Expected

- Script embeds the cwd path `/tmp/iterm2-tabset-cwd` (or its escaped form).
- Script contains a `cd` write-text for that path (e.g. `write text ("cd "` or
  a command string containing `cd` and the path).
- Both tab commands `echo-with-cwd` and `echo-no-cwd` appear.
- Unlike Open dir flow, there is no mandatory shared `targetDir` project cd for
  every tab: either `quoted form of targetDir` is absent, or there is at most
  one cd involving the cwd path (only the cwd tab).

## Exit Code

- N/A (build-tab-set-script phase)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	const cwd = `/tmp/iterm2-tabset-cwd`
	if !strings.Contains(s, cwd) {
		t.Fatalf("missing cwd path %q; script:\n%s", cwd, s)
	}
	// A cd related to the cwd path must appear (AppleScript form flexible).
	hasCdForm := strings.Contains(s, `write text ("cd "`) ||
		strings.Contains(s, `write text "cd `) ||
		(strings.Contains(s, "cd ") && strings.Contains(s, cwd))
	if !hasCdForm {
		t.Fatalf("expected a cd write for non-empty Cwd; script:\n%s", s)
	}
	if !strings.Contains(s, "echo-with-cwd") {
		t.Fatalf("missing command echo-with-cwd; script:\n%s", s)
	}
	if !strings.Contains(s, "echo-no-cwd") {
		t.Fatalf("missing command echo-no-cwd; script:\n%s", s)
	}
	// Empty-cwd tab must not require Open-style project targetDir cd for all sessions.
	// Allow at most one Open-style shared cd line; prefer none for pure tab-set.
	openStyleCD := `write text ("cd " & quoted form of targetDir)`
	if n := countOccurrences(s, openStyleCD); n > 1 {
		t.Fatalf("Open-style project cd appears %d times; empty Cwd should not force project cd for every tab; script:\n%s", n, s)
	}
}
```
