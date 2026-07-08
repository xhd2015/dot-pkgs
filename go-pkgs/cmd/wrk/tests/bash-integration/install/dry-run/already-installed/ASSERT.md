## Expected

- Exit code 0.
- Stdout reports already installed with script and both profile marker states.
- Pre-seeded bash.sh and profile content unchanged.

## Side Effects

- No profile modifications.
- No bash.sh overwrite.
- No `events.jsonl`.

## Exit Code

- 0

```go
import (
	"os"
	"strings"
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
__SCRIPT__: type=string, example=/tmp/.wrk/integration/bash.sh, bash.sh path
__BASH_PROFILE__: type=string, example=/tmp/home/.bash_profile, bash_profile path
__BASHRC__: type=string, example=/tmp/home/.bashrc, bashrc path
---
wrk bash integration: already installed
script: __SCRIPT__ (exists)
bash_profile: __BASH_PROFILE__ (marker present)
bashrc: __BASHRC__ (marker present)
no changes needed

`)

	if resp.BashProfileMarkerCount != 1 {
		t.Fatalf("dry-run must not change .bash_profile marker count; got %d:\n%s",
			resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 1 {
		t.Fatalf("dry-run must not change .bashrc marker count; got %d:\n%s",
			resp.BashRCMarkerCount, resp.BashRCContent)
	}
	if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
		t.Fatalf("dry-run must preserve unrelated .bash_profile content:\n%s", resp.BashProfileContent)
	}
	if !strings.Contains(resp.BashShContent, "pre-seeded wrk bash integration") {
		t.Fatalf("dry-run must not overwrite bash.sh:\n%s", resp.BashShContent)
	}
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("bash.sh must still exist: %v", statErr)
	}
	assertNoEventsJSONL(t, resp)
}
```