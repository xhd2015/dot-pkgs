## Expected
- Exit code 0.
- Output contains the basename "myproject".
- Output shows the original location marker.

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "myproject")
	assertContains(t, resp.Output, "(original)")
}
```
