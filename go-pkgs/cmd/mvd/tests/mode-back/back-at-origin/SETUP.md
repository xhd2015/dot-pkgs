## Steps
- Create `src` and `dst` directories.
- Move `src` into `dst` (creates a history chain of two locations).
- Run `mvd --back <moved-path>` to return the project to its original location `src`.
- Run `mvd --back src` again — since the project is already at its origin, this should be a no-op that reports "nothing to move back".

```go
func Setup(t *testing.T, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    dst := filepath.Join(req.WorkRoot, "dst")
    mkdirAll(t, src)
    mkdirAll(t, dst)
    
    req.Args = []string{src, dst}
    resp, err := runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }
    
    p := filepath.Join(dst, "src")
    req.Args = []string{"--back", p}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("first back: %s", resp.Output) }
    
    req.Args = []string{"--back", src}
    return nil
}
```
