## Expected
- Non-zero exit code.
- Output contains "does not exist".

## Exit Code
- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0: %s", resp.Output)
	}
	assertContains(t, resp.Output, "does not exist")
}
```
