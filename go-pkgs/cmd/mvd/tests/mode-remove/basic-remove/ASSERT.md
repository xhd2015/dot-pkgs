## Expected
- The output contains "removed", confirming the tracked entry was deleted from history.
- The history file is empty (no remaining tracked projects).

## Exit Code
- 0 (success)

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp == nil {
        t.Fatalf("expected response, got error: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
    }
    assertContains(t, resp.Output, "removed")
    assertHistoryNil(t, req.ConfigHome)
}
```
