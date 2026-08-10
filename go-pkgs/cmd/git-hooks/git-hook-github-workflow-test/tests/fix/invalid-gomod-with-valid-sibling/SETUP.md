## Preconditions

- The repository origin is `github.com`.
- Root `go.mod` declares Go version `1.22`.
- Nested `kool-template/go.mod` is a comment-only marker (no `go` directive).
- `.github/workflows/test.yml` does not exist.

## Steps

1. Keep the default root `go.mod`.
2. Write an invalid nested `go.mod` without a `go` directive.
3. Run the command with `--fix`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "fix invalid gomod with valid sibling"
	if err := writeFile(req.RepoDir, "kool-template/go.mod", "// Not a Go module.\n// Template only.\n"); err != nil {
		return err
	}
	return nil
}
```
