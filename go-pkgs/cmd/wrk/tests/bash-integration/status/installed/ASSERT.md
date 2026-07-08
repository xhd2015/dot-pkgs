## Expected

- Exit code 0.
- Stdout reports `installed` with script present and both profile markers present.
- No filesystem changes from status itself.

## Side Effects

- Read-only inspection.

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
__SCRIPT__: type=string, example=/tmp/.wrk/integration/bash.sh, bash.sh path
__BASH_PROFILE__: type=string, example=/tmp/home/.bash_profile, bash_profile path
__BASHRC__: type=string, example=/tmp/home/.bashrc, bashrc path
---
bash integration: installed
script: __SCRIPT__ (present)
bash_profile: __BASH_PROFILE__ (marker present)
bashrc: __BASHRC__ (marker present)

`)

	if resp.BashProfileMarkerCount != 1 || resp.BashRCMarkerCount != 1 {
		t.Fatalf("expected installed markers; profile=%d bashrc=%d",
			resp.BashProfileMarkerCount, resp.BashRCMarkerCount)
	}
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("expected bash.sh present: %v", statErr)
	}
	assertHomeIsolated(t, resp.BashProfilePath, resp.Home)
	assertHomeIsolated(t, resp.BashRCPath, resp.Home)
	assertWrkHomeIsolated(t, resp.BashShPath, resp.WrkHome)
}
```