## Preconditions

- The repository origin is `github.com`.
- Root `go.mod` is a comment-only marker file (no `go` directive).
- `.github/workflows/test.yml` does not exist.

## Steps

1. Overwrite root `go.mod` with content that has no `go` directive.
2. Run the command in check mode.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "check invalid gomod missing go directive"
	if err := writeFile(req.RepoDir, "go.mod", "// Not a Go module.\n// Template only.\n"); err != nil {
		return err
	}
	return nil
}
```
