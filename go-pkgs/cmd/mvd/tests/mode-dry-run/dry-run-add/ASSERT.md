## Expected
- Exit code is 0.
- Output contains `dry-run: would add`.
- No history entry was created.

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "dry-run: would add")
	assertHistoryNil(t, req.ConfigHome)
}
```
