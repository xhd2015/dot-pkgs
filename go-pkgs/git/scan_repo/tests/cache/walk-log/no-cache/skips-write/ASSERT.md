## Expected

- Scan succeeds and discovers `my-repo` as a main.
- `resp.WalkLogOK` is false; `home/walk.jsonl` does not exist.
- `resp.CursorOK` is false; `home/walk.cursor.json` does not exist.
- `resp.WalkEvents` is empty / nil.

## Errors

- `err` is nil.

## Side Effects

- No walk log durable files under `<CacheRoot>/home/` for this NoCache Scan.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	rootAbs := absPath(t, req.Roots[0])
	repo := absPath(t, filepath.Join(rootAbs, "my-repo"))
	found := false
	for _, r := range resp.Repos {
		if absPath(t, r.Path) == repo && r.RepoType == scan_repo.RepoTypeMain {
			found = true
		}
	}
	if !found {
		t.Fatalf("Result missing main my-repo %q; repos=%v", repo, resp.Repos)
	}

	wantLog := homeWalkLogPath(req.CacheRoot)
	wantCur := homeWalkCursorPath(req.CacheRoot)

	if resp.WalkLogOK {
		t.Fatal("WalkLogOK=true with NoCache; want no walk.jsonl")
	}
	if resp.CursorOK {
		t.Fatal("CursorOK=true with NoCache; want no walk.cursor.json")
	}
	if len(resp.WalkEvents) != 0 {
		t.Fatalf("WalkEvents = %v, want empty with NoCache", resp.WalkEvents)
	}
	if _, statErr := os.Stat(wantLog); !os.IsNotExist(statErr) {
		t.Fatalf("home/walk.jsonl should be absent with NoCache; stat err=%v", statErr)
	}
	if _, statErr := os.Stat(wantCur); !os.IsNotExist(statErr) {
		t.Fatalf("home/walk.cursor.json should be absent with NoCache; stat err=%v", statErr)
	}
}
```
