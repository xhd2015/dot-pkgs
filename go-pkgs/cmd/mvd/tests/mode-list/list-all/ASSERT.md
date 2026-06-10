## Expected
- Exit code 0.
- Output lists both proj1 and proj2.

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "proj1")
	assertContains(t, resp.Output, "proj2")
}
```
