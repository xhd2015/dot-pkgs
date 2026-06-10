## Steps
- Create a project whose root basename is `opencode-latest` (NOT `opencode`).
- Register an alias `opencode` pointing to this project via `--add-alias opencode opencode-latest`.
- Move the project by its alias `opencode` into `scratch`.
- Run `mvd --back opencode` from a separate working directory — this must resolve the alias to find the project in history.

```go
func Setup(t *testing.T, req *Request) error {
    projectRoot := filepath.Join(req.WorkRoot, "projects")
    src := filepath.Join(projectRoot, "opencode-latest")
    dst := filepath.Join(req.WorkRoot, "scratch")
    mkdirAll(t, src)
    mkdirAll(t, dst)
    
    req.Args = []string{"--add", src}
    resp, err := runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }
    
    req.Args = []string{"--add-alias", "opencode", "opencode-latest"}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add-alias: %s", resp.Output) }
    
    req.Args = []string{"opencode", dst}
    resp, err = runMvd(t, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move by alias: %s", resp.Output) }
    
    cwd := filepath.Join(req.WorkRoot, "cwd")
    mkdirAll(t, cwd)
    if err := os.Chdir(cwd); err != nil { return err }
    
    req.Args = []string{"--back", "opencode"}
    return nil
}
```
