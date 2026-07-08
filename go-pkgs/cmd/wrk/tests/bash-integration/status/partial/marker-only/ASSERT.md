## Expected

- Exit code 1.
- Stdout reports `partial` with script absent and both profile markers present.
- Profile content unchanged.

## Side Effects

- Read-only inspection.

## Exit Code

- 1

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
	if resp.ExitCode != 1 {
		t.Fatalf("expected exit 1, got %d; stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assert.Output(t, resp.Stdout, `---
version: 2
__SCRIPT__: type=string, example=/tmp/.wrk/integration/bash.sh, bash.sh path
__BASH_PROFILE__: type=string, example=/tmp/home/.bash_profile, bash_profile path
__BASHRC__: type=string, example=/tmp/home/.bashrc, bashrc path
---
bash integration: partial
script: __SCRIPT__ (absent)
bash_profile: __BASH_PROFILE__ (marker present)
bashrc: __BASHRC__ (marker present)

`)

	if resp.BashProfileMarkerCount != 1 || resp.BashRCMarkerCount != 1 {
		t.Fatalf("expected marker-only state; profile=%d bashrc=%d",
			resp.BashProfileMarkerCount, resp.BashRCMarkerCount)
	}
	if _, statErr := os.Stat(resp.BashShPath); !os.IsNotExist(statErr) {
		t.Fatalf("status must not create bash.sh at %s", resp.BashShPath)
	}
	assertHomeIsolated(t, resp.BashProfilePath, resp.Home)
	assertHomeIsolated(t, resp.BashRCPath, resp.Home)
}
```