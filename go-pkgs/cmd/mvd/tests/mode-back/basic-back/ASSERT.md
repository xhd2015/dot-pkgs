## Expected
- The project is moved back to its original location at `src`.
- The file `src/f.txt` that was originally in `src` exists again at the original path.
- The moved path `dst/src` no longer exists.
- The history chain is reduced to a single entry at `src` (the project is back at its origin).

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
    assertContains(t, resp.Output, "moved back")
    
    src := filepath.Join(req.WorkRoot, "src")
    dstSrc := filepath.Join(req.WorkRoot, "dst", "src")
    assertFileExists(t, filepath.Join(src, "f.txt"))
    assertFileNotExists(t, dstSrc)
    
    assertHistoryChain(t, req.ConfigHome, src, src)
}
```
