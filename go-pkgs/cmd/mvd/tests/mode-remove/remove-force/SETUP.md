## Steps
- Create `src` and `d1` directories, then move `src` → `d1` to create a movement history.
- Run `mvd --rm -f src` — the `--force` flag clears the history even though the entry has multiple locations.

```go
func Setup(t *testing.T, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    d1 := filepath.Join(req.WorkRoot, "d1")
    mkdirAll(t, src)
    mkdirAll(t, d1)
    
    req.Args = []string{src, d1}
    resp, err := runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }
    
    req.Args = []string{"--rm", "-f", src}
    return nil
}
```
