## Expected

- Exit code 0.
- Stdout lists both main repo paths, one per line, sorted lexicographically.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertProjectsSortedStdout(t, resp.Stdout, []string{req.MainRepo, req.SecondRepo})
	assertProjectsCount(t, req.WrkHome, 2)
}
```