## Expected
- The output contains "has movement history", indicating the removal was rejected because the entry has multiple locations and `--force` was not provided.
- The history still contains the `src` project (it was not removed).

## Exit Code
- 1 (error)

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp == nil {
        t.Fatalf("expected response, got error: %v", err)
    }
    if resp.ExitCode == 0 {
        t.Fatalf("expected non-zero exit code, got 0, output:\n%s", resp.Output)
    }
    assertContains(t, resp.Output, "has movement history")
    assertHistoryLen(t, req.ConfigHome, 1)
}
```
