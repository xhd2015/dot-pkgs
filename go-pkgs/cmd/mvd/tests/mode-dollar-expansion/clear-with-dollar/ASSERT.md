## Expected
- Exit code 0.
- Output contains "cleared history".
- History is nil (no projects remain).

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "cleared history")
	assertHistoryNil(t, req.ConfigHome)
}
```
