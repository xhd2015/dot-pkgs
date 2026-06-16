# Scenario

Error when removing an entry with history without --force.

mvd --add tracked → [(tracked)]
mvd tracked dst → [(tracked), (dst/tracked)]
mvd --rm tracked → error → has move history

## Steps
- Create `src` and `d1` directories, then move `src` → `d1` to create a movement history (more than one location).
- Run `mvd --rm src` without the `--force` flag — this should fail because the entry has movement history that would be lost.

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
    
    req.Args = []string{"--rm", src}
    return nil
}
```
