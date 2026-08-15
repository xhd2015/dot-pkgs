## Expected

- One repo with two remotes: `origin` and `upstream`.
- `origin`: Host `github.com`, Owner `xhd2015`, Repo `lifelog`.
- `upstream`: Host `github.com`, Owner `golang`, Repo `go`.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	if len(resp.Repos[0].Remotes) != 2 {
		t.Fatalf("expected 2 remotes, got %v", resp.Repos[0].Remotes)
	}
	byName := map[string]struct{ Host, Owner, Repo string }{}
	for _, r := range resp.Repos[0].Remotes {
		byName[r.Name] = struct{ Host, Owner, Repo string }{r.Host, r.Owner, r.Repo}
	}
	origin, ok := byName["origin"]
	if !ok {
		t.Fatal("missing origin remote")
	}
	if origin.Host != "github.com" || origin.Owner != "xhd2015" || origin.Repo != "lifelog" {
		t.Fatalf("origin = %+v, want github.com/xhd2015/lifelog", origin)
	}
	upstream, ok := byName["upstream"]
	if !ok {
		t.Fatal("missing upstream remote")
	}
	if upstream.Host != "github.com" || upstream.Owner != "golang" || upstream.Repo != "go" {
		t.Fatalf("upstream = %+v, want github.com/golang/go", upstream)
	}
}
```