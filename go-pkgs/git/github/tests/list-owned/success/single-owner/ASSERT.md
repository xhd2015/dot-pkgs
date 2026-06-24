## Expected

- `len(resp.Repos)` is 2.
- Repos sorted by `FullName` ascending: `alice/alpha` before `alice/beta`.
- `alice/alpha`: Owner `alice`, Name `alpha`, Description `first repo`, IsFork true, URL `https://github.com/alice/alpha`.
- `alice/beta`: Owner `alice`, Name `beta`, Description `second repo`, IsFork false, URL `https://github.com/alice/beta` (SSH input normalized).

## Side Effects

- Mock `gh` invoked once for owner `alice`.

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %+v", len(resp.Repos), resp.Repos)
	}
	names := []string{resp.Repos[0].FullName, resp.Repos[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Repos[0].FullName != "alice/alpha" {
		t.Fatalf("expected first repo alice/alpha, got %q", resp.Repos[0].FullName)
	}
	if resp.Repos[1].FullName != "alice/beta" {
		t.Fatalf("expected second repo alice/beta, got %q", resp.Repos[1].FullName)
	}

	alpha := resp.Repos[0]
	if alpha.Owner != "alice" || alpha.Name != "alpha" || alpha.Description != "first repo" || !alpha.IsFork {
		t.Fatalf("unexpected alpha mapping: %+v", alpha)
	}
	if alpha.URL != "https://github.com/alice/alpha" {
		t.Fatalf("expected normalized alpha URL, got %q", alpha.URL)
	}

	beta := resp.Repos[1]
	if beta.Owner != "alice" || beta.Name != "beta" || beta.Description != "second repo" || beta.IsFork {
		t.Fatalf("unexpected beta mapping: %+v", beta)
	}
	if beta.URL != "https://github.com/alice/beta" {
		t.Fatalf("expected normalized beta URL from SSH, got %q", beta.URL)
	}
}```