## Expected

- The scanner returns exactly one local replace issue.
- The issue points at `../`.
- The issue is classified as intra-repo even when `GIT_DIR` is inherited from a Git hook environment.

## Exit Code

- No error.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("CheckLocalReplaces returned error: %v", resp.Err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("issues len = %d, want 1: %+v", len(resp.Issues), resp.Issues)
	}
	issue := resp.Issues[0]
	if issue.NewPath != "../" {
		t.Fatalf("issue NewPath = %q, want ../: %+v", issue.NewPath, issue)
	}
	if !issue.IsIntraRepo {
		t.Fatalf("issue IsIntraRepo = false, want true under inherited GIT_DIR hook env: %+v", issue)
	}
}

```
