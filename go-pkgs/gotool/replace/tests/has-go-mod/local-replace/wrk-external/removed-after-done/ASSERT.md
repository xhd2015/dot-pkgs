## Expected

- No error from the scan.
- Zero issues returned (resolved path prefix is still inside the scanning worktree).

## Exit Code

- No error.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("CheckLocalReplaces returned error: %v", resp.Err)
	}
	if len(resp.Issues) != 0 {
		t.Fatalf("expected 0 issues: removed external path still under worktree root, got %d: %+v", len(resp.Issues), resp.Issues)
	}
}

```