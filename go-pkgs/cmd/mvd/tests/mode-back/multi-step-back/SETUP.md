# Scenario

Multiple --back calls unwind the full chain to the origin.

mvd src d1 → [(src), (d1/src)]
mvd d1/src d2 → [(src), (d1/src), (d2/src)]
mvd --back d2/src → [(src), (d1/src)]
mvd --back d1/src → [(src)]

## Steps
- Create three directories: `src`, `d1`, and `d2`.
- Move `src` → `d1` (first hop), then `d1/src` → `d2` (second hop). This leaves the project at `d2/src` with history chain `src → d1/src → d2/src`.
- Run `mvd --back d2/src` to move back one step (to `d1/src`).
- Run `mvd --back d1/src` to move back another step (to `src`, the origin).
- Run `mvd --back src` — the project is already at origin, so this should be a no-op.

```go
func Setup(t *testing.T, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    d1 := filepath.Join(req.WorkRoot, "d1")
    d2 := filepath.Join(req.WorkRoot, "d2")
    mkdirAll(t, src)
    mkdirAll(t, d1)
    mkdirAll(t, d2)
    
    req.Args = []string{src, d1}
    resp, err := runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move 1: %s", resp.Output) }
    p1 := filepath.Join(d1, "src")
    
    req.Args = []string{p1, d2}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move 2: %s", resp.Output) }
    p2 := filepath.Join(d2, "src")
    
    req.Args = []string{"--back", p2}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("back p2: %s", resp.Output) }
    
    req.Args = []string{"--back", p1}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("back p1: %s", resp.Output) }
    
    req.Args = []string{"--back", src}
    return nil
}
```
