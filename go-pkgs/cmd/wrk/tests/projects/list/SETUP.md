# Scenario

**Feature**: wrk --projects lists recorded project paths

```
wrk --projects -> sorted absolute main-repo paths (one per line)
```

## Preconditions

- `wrk --projects` is a standalone mode.

## Steps

- Descendants pre-populate `projects.json` or leave it empty.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}
```