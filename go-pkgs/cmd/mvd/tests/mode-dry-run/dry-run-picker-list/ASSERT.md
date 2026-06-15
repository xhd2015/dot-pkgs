## Expected
- Exit code is 0.
- Output contains normal picker-list output (`->` separator between display and full path).
- Output does NOT contain `dry-run:`.

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "->")
	assertNotContains(t, resp.Output, "dry-run:")
}
```
