## Expected
- Exit code is 0.
- Output contains `dry-run: would clear`.
- History still has the entry (not cleared).

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "dry-run: would clear history")
	assertHistoryLen(t, req.ConfigHome, 1)
}
```
