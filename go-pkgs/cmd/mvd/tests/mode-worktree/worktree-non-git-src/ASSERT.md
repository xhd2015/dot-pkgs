## Expected
- Non-zero exit code.
- Output contains "not a git repository".
- No history recorded.

## Exit Code
- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\noutput: %s", resp.Output)
	}
	assertContains(t, resp.Output, "not a git repository")
	assertHistoryNil(t, req.ConfigHome)
}
```
