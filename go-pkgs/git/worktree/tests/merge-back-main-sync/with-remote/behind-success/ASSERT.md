# Assert

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Err != "" {
		t.Fatalf("unexpected error: %s", resp.Err)
	}
	// After sync, main has remote-only while feature branched from pre-sync
	// tip → diverged → rebased-and-merged (not a plain ahead merge).
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("Action=%q want rebased-and-merged", resp.Action)
	}
	for _, name := range []string{"remote-only.txt", "feature.txt"} {
		p := filepath.Join(req.MainRepo, name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("%s missing on main: %v", name, err)
		}
	}
	remoteTip := runGitOutput(t, req.MainRepo, "rev-parse", "origin/master")
	cmd := exec.Command("git", "-C", req.MainRepo, "merge-base", "--is-ancestor", remoteTip, "HEAD")
	if err := cmd.Run(); err != nil {
		t.Fatalf("main HEAD does not contain synced origin/master tip %s", remoteTip)
	}
}
```
