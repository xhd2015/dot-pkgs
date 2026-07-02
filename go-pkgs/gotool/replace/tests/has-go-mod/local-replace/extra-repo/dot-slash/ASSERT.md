## Expected

- No error from the scan.
- Exactly 1 issue is returned.
- The issue has `IsIntraRepo == false`.
- The issue has `NewPath == "./nonexistent"`.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("expected no error, got: %v", resp.Err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %+v", len(resp.Issues), resp.Issues)
	}
	issue := resp.Issues[0]
	if issue.IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=false, got true. issue: %+v", issue)
	}
	if issue.NewPath != "./nonexistent" {
		t.Fatalf("expected NewPath=./nonexistent, got %q", issue.NewPath)
	}
}
```