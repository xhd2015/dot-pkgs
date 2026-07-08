## Expected Output

Preview would-write script and would-append marker blocks to both profiles.

## Expected

- Exit code 0.
- Stdout previews writing `integration/bash.sh` and appending marker blocks to both profiles.
- Stdout includes the wrk marker block body.
- No bash.sh or profile files created.

## Side Effects

- No profile modifications.
- No bash.sh write.
- No `events.jsonl`.

## Exit Code

- 0

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}

	assert.Output(t, resp.Stdout, `---
version: 2
---
dry-run: would write integration/bash.sh
dry-run: would append marker block to ~/.bash_profile
dry-run: would append marker block to ~/.bashrc

# === wrk integration begin ===
_wrk_home="${WRK_HOME:-$HOME/.wrk}"
[[ -f "$_wrk_home/integration/bash.sh" ]] && source "$_wrk_home/integration/bash.sh"
# === wrk integration end ===

`)

	if _, statErr := os.Stat(resp.BashShPath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create bash.sh at %s", resp.BashShPath)
	}
	if _, statErr := os.Stat(resp.BashProfilePath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create .bash_profile at %s", resp.BashProfilePath)
	}
	if _, statErr := os.Stat(resp.BashRCPath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create .bashrc at %s", resp.BashRCPath)
	}
	assertNoEventsJSONL(t, resp)
}
```