## Expected
- mvd detects the duplicate and reports it with "already recorded".
- History still contains exactly one project.

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "already recorded")
	assertHistoryLen(t, req.ConfigHome, 1)
}
```
