## Expected

- `DefaultInstallDir` equals `filepath.Join(os.UserHomeDir(), "installed")`.

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("DefaultInstallDir failed: %v", resp.Err)
	}
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		t.Fatalf("UserHomeDir: %v", homeErr)
	}
	want := filepath.Join(home, "installed")
	if resp.InstallDir != want {
		t.Fatalf("DefaultInstallDir() = %q, want %q", resp.InstallDir, want)
	}
}
```
