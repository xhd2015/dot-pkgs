## Expected

- Exit code is 0
- Stdout contains the usage text with all flag descriptions: `--listen-port`, `--upstream-proxy`, `--no-fallback-direct`, `--help`

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", resp.ExitCode, resp.Output)
	}
	output := resp.Output
	for _, want := range []string{"--listen-port", "--upstream-proxy", "--no-fallback-direct", "--help", "Usage"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}
```
