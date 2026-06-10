## Expected
- Exit code 0.
- Output marks the original location with "(original)".
- Output marks the current location with "*".

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "(original)")
	assertContains(t, resp.Output, "*")
}
```
