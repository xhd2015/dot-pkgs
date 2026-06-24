## Expected

- One repo with one remote named `origin`.
- `Host` is `github.com`, `Owner` is `xhd2015`, `Repo` is `lifelog`.
- `URL` contains the configured remote URL.

## Errors

- `err` is nil.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	if len(resp.Repos[0].Remotes) != 1 {
		t.Fatalf("expected 1 remote, got %v", resp.Repos[0].Remotes)
	}
	rem := resp.Repos[0].Remotes[0]
	if rem.Name != "origin" {
		t.Fatalf("remote name = %q, want origin", rem.Name)
	}
	if rem.Host != "github.com" {
		t.Fatalf("Host = %q, want github.com", rem.Host)
	}
	if rem.Owner != "xhd2015" {
		t.Fatalf("Owner = %q, want xhd2015", rem.Owner)
	}
	if rem.Repo != "lifelog" {
		t.Fatalf("Repo = %q, want lifelog", rem.Repo)
	}
	if !strings.Contains(rem.URL, "lifelog") {
		t.Fatalf("URL = %q, expected to reference lifelog", rem.URL)
	}
}
```