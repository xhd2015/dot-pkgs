## Expected

- Exit code 0 after second install.
- Exactly one marker block remains in each profile.
- Pre-seeded `bash.sh` still present; content not duplicated/overwritten with a second copy.

## Side Effects

- No duplicate markers appended.
- Existing bash.sh not overwritten.

## Exit Code

- 0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if resp.BashProfileMarkerCount != 1 {
		t.Fatalf("idempotent install must not duplicate .bash_profile marker; count=%d:\n%s",
			resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 1 {
		t.Fatalf("idempotent install must not duplicate .bashrc marker; count=%d:\n%s",
			resp.BashRCMarkerCount, resp.BashRCContent)
	}
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("bash.sh missing after idempotent install: %v", statErr)
	}
	if !strings.Contains(resp.BashShContent, "pre-seeded wrk bash integration") {
		t.Fatalf("idempotent install must not overwrite existing bash.sh:\n%s", resp.BashShContent)
	}
	assertNoEventsJSONL(t, resp)
}
```