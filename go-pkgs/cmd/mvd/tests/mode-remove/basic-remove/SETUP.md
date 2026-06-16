# Scenario

Remove a tracked entry with no movement history.

mvd --add tracked → [(tracked)]
mvd --rm tracked → []  (removed)

## Steps
- Create a directory `tracked` and add it to mvd's tracking via `--add`.
- Remove it with `--rm`. Since it has no movement history (only one location), the removal succeeds without requiring `--force`.

```go
func Setup(t *testing.T, req *Request) error {
    dir := filepath.Join(req.WorkRoot, "tracked")
    mkdirAll(t, dir)
    
    req.Args = []string{"--add", dir}
    resp, err := runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }
    
    req.Args = []string{"--rm", dir}
    return nil
}
```
